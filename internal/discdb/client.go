package discdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client interface {
	LookupDisc(ctx context.Context, discHash string) (*DiscRecord, error)
}

type CachedClient struct {
	// cache  *Cache
	client *RemoteClient
}

func NewCachedClient() *CachedClient {
	return &CachedClient{
		// cache:  nil,
		client: newRemoteClient(),
	}
}

func (c *CachedClient) LookupDisc(ctx context.Context, discHash string) (*DiscRecord, error) {
	return c.client.LookupDisc(ctx, discHash)
}

type RemoteClient struct {
	httpClient *http.Client
	endpoint   string
}

func newRemoteClient() *RemoteClient {
	return &RemoteClient{
		endpoint: "https://thediscdb.com/graphql",
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type graphqlPayload struct {
	OperationName string         `json:"operationName"`
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables"`
}

type graphqlResponse[T any] struct {
	Data   T              `json:"data"`
	Errors []graphqlError `json:"errors"`
}

// There are additional fields that I don't think I'll need.
type graphqlError struct {
	Message string   `json:"message"`
	Path    []string `json:"path"`
}

type discDbGraphqlResponse struct {
	MediaItems struct {
		Nodes []MediaItemResponse `json:"nodes"`
	} `json:"mediaItems"`
}

type MediaItemResponse struct {
	Title    string            `json:"title"`
	Year     int               `json:"year"`
	Type     string            `json:"type"`
	Releases []ReleaseResponse `json:"releases"`
}

type ReleaseResponse struct {
	Slug   string         `json:"slug"`
	Locale string         `json:"locale"`
	Year   int            `json:"year"`
	Title  string         `json:"title"`
	Discs  []DiscResponse `json:"discs"`
}

type DiscResponse struct {
	ContentHash string          `json:"contentHash"`
	Index       int             `json:"index"`
	Name        string          `json:"name"`
	Format      string          `json:"format"`
	Slug        string          `json:"slug"`
	Titles      []TitleResponse `json:"titles"`
}

type TitleResponse struct {
	Index       int    `json:"index"`
	Duration    string `json:"duration"`
	DisplaySize string `json:"displaySize"`
	SourceFile  string `json:"sourceFile"`
	Size        uint64 `json:"size"`
	SegmentMap  string `json:"segmentMap"`
	// Item is known to be nullable. Making it a pointer so that it can be checked for nil
	Item *ItemResponse `json:"item"`
}

type ItemResponse struct {
	Title   string `json:"title"`
	Season  string `json:"season"`
	Episode string `json:"episode"`
	Type    string `json:"type"`
}

func (r *RemoteClient) LookupDisc(ctx context.Context, discHash string) (*DiscRecord, error) {
	payload := graphqlPayload{
		OperationName: "GetDiscByContentHash",
		Query:         getDiscByContentHash,
		Variables: map[string]any{
			"hash": discHash,
		},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal graphql payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("build graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute graphql request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read graphql response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s returned %s:\n%s", res.Request.URL, res.Status, body)
	}

	var graphqlResponse graphqlResponse[discDbGraphqlResponse]
	if err := json.Unmarshal(body, &graphqlResponse); err != nil {
		return nil, fmt.Errorf("decode graphql response: %w", err)
	}

	// TODO: Determine how to return these errors
	if len(graphqlResponse.Errors) > 0 {
		return nil, fmt.Errorf("graphql returned errors")
	}

	mediaItemNodes := graphqlResponse.Data.MediaItems.Nodes
	switch len(mediaItemNodes) {
	case 0:
		return nil, fmt.Errorf("no disc found for hash %s", discHash)
	case 1:
		return mediaItemResponseToDiscRecord(&mediaItemNodes[0], discHash)
	default:
		return nil, fmt.Errorf("multiple discs found for hash %s", discHash)
	}
}

type DiscRecord struct {
	Media   Media
	Release Release
	Disc    Disc
}

type Media struct {
	Title string
	Year  int
	Type  string
}

type Release struct {
	Slug   string
	Locale string
	Year   int
	Title  string
}

type Disc struct {
	ContentHash string
	Name        string
	Format      string
	Slug        string
	Titles      []Title
}

type Title struct {
	Duration    string
	DisplaySize string
	SourceFile  string
	Size        uint64
	SegmentMap  string
	item        *Item
}

func (t *Title) Item() (Item, bool) {
	if t.item == nil {
		return Item{}, false
	}

	return *t.item, true
}

type Item struct {
	Title   string
	Season  string
	Episode string
	Type    ItemType
}

type ItemType string

const (
	ItemTypeMovie        ItemType = "MainMovie"
	ItemTypeExtra        ItemType = "Extra"
	ItemTypeEpisode      ItemType = "Episode"
	ItemTypeDeletedScene ItemType = "DeletedScene"
	ItemTypeTrailer      ItemType = "Trailer"
)

func mediaItemResponseToDiscRecord(mediaItemResponse *MediaItemResponse, discHash string) (*DiscRecord, error) {
	var matchedRelease *ReleaseResponse
	var matchedDisc *DiscResponse

	for i := range mediaItemResponse.Releases {
		release := &mediaItemResponse.Releases[i]

		for j := range release.Discs {
			disc := &release.Discs[j]

			if disc.ContentHash != discHash {
				continue
			}

			// TODO: Consider allowing multiple matches in the future.
			// Will this require allowing the user to select from the matches?
			// Might be able to compare the segment maps to those reported by makemkv to find a unique match
			if matchedDisc != nil || matchedRelease != nil {
				return nil, fmt.Errorf("multiple matching discs found for hash %s", discHash)
			}

			matchedDisc = disc
			matchedRelease = release
		}
	}

	if matchedDisc == nil || matchedRelease == nil {
		return nil, fmt.Errorf("no matching disc found for hash %s", discHash)
	}

	return &DiscRecord{
		Media: Media{
			Title: mediaItemResponse.Title,
			Year:  mediaItemResponse.Year,
			Type:  mediaItemResponse.Type,
		},
		Release: Release{
			Slug:   matchedRelease.Slug,
			Locale: matchedRelease.Locale,
			Year:   matchedRelease.Year,
			Title:  matchedRelease.Title,
		},
		Disc: Disc{
			ContentHash: matchedDisc.ContentHash,
			Name:        matchedDisc.Name,
			Format:      matchedDisc.Format,
			Slug:        matchedDisc.Slug,
			Titles:      titleResponsesToTitles(matchedDisc.Titles),
		},
	}, nil
}

func titleResponsesToTitles(titleResponses []TitleResponse) []Title {
	titles := make([]Title, 0, len(titleResponses))

	for _, titleResponse := range titleResponses {
		titles = append(titles, Title{
			Duration:    titleResponse.Duration,
			DisplaySize: titleResponse.DisplaySize,
			SourceFile:  titleResponse.SourceFile,
			Size:        titleResponse.Size,
			SegmentMap:  titleResponse.SegmentMap,
			item:        itemResponseToItem(titleResponse.Item),
		})
	}

	return titles
}

func itemResponseToItem(itemResponse *ItemResponse) *Item {
	if itemResponse == nil {
		return nil
	}

	return &Item{
		Title:   itemResponse.Title,
		Season:  itemResponse.Season,
		Episode: itemResponse.Episode,
		Type:    ItemType(itemResponse.Type),
	}
}

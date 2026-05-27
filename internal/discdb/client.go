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

const discDbGraphqlEndpoint = "https://thediscdb.com/graphql"

type Client interface {
	LookupDisc(ctx context.Context, discHash string) (DiscRecord, error)
}

type CachedClient struct {
	cache  Cache
	client Client
}

func NewCachedClient(cache Cache, client Client) (*CachedClient, error) {
	return &CachedClient{
		cache:  cache,
		client: client,
	}, nil
}

func (c *CachedClient) LookupDisc(ctx context.Context, discHash string) (DiscRecord, error) {
	if record, ok, err := c.cache.GetDiscRecord(ctx, discHash); err != nil {
		return DiscRecord{}, fmt.Errorf("disc cache read: %w", err)
	} else if ok {
		return record, nil
	}
	
	record, err := c.client.LookupDisc(ctx, discHash)
	if err != nil {
		return DiscRecord{}, fmt.Errorf("disc lookup: %w", err) 
	}

	err = c.cache.PutDiscRecord(ctx, discHash, record)
	if err != nil {
		return DiscRecord{}, fmt.Errorf("disc cache write: %w", err)
	}
	
	return record, nil
}

type RemoteClient struct {
	httpClient *http.Client
	endpoint   string
}

func NewRemoteClient() *RemoteClient {
	return &RemoteClient{
		endpoint: discDbGraphqlEndpoint,
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

func (r *RemoteClient) LookupDisc(ctx context.Context, discHash string) (DiscRecord, error) {
	payload := graphqlPayload{
		OperationName: "GetDiscByContentHash",
		Query:         getDiscByContentHash,
		Variables: map[string]any{
			"hash": discHash,
		},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return DiscRecord{}, fmt.Errorf("marshal graphql payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return DiscRecord{}, fmt.Errorf("build graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := r.httpClient.Do(req)
	if err != nil {
		return DiscRecord{}, fmt.Errorf("execute graphql request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return DiscRecord{}, fmt.Errorf("read graphql response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return DiscRecord{}, fmt.Errorf("request to %s returned %s:\n%s", res.Request.URL, res.Status, body)
	}

	var graphqlResponse graphqlResponse[discDbGraphqlResponse]
	if err := json.Unmarshal(body, &graphqlResponse); err != nil {
		return DiscRecord{}, fmt.Errorf("decode graphql response: %w", err)
	}

	// TODO: Determine how to return these errors
	if len(graphqlResponse.Errors) > 0 {
		return DiscRecord{}, fmt.Errorf("graphql returned errors")
	}

	mediaItemNodes := graphqlResponse.Data.MediaItems.Nodes
	switch len(mediaItemNodes) {
	case 0:
		return DiscRecord{}, fmt.Errorf("no disc found for hash %s", discHash)
	case 1:
		return mediaItemResponseToDiscRecord(&mediaItemNodes[0], discHash)
	default:
		return DiscRecord{}, fmt.Errorf("multiple discs found for hash %s", discHash)
	}
}


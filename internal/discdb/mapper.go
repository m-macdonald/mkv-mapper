package discdb

import (
	"encoding/json"
	"fmt"

	"m-macdonald/mkv-mapper/internal/signature"
)

type DiscRecord struct {
	Media   Media   `json:"media"`
	Release Release `json:"release"`
	Disc    Disc    `json:"disc"`
}

type Media struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	Type  string `json:"type"`
}

type Release struct {
	Slug   string `json:"slug"`
	Locale string `json:"locale"`
	Year   int    `json:"year"`
	Title  string `json:"title"`
}

type Disc struct {
	ContentHash string  `json:"contentHash"`
	Name        string  `json:"name"`
	Format      string  `json:"format"`
	Slug        string  `json:"slug"`
	Titles      []Title `json:"titles"`
}

type Title struct {
	Duration    string                     `json:"duration"`
	DisplaySize string                     `json:"displaySize"`
	SourceFile  string                     `json:"sourceFile"`
	Size        uint64                     `json:"size"`
	SegmentMap  string                     `json:"segmentMap"`
	Signature   signature.SegmentSignature `json:"signature"`
	Item        *Item                      `json:"item,omitempty"`
}

func (t *Title) ItemValue() (Item, bool) {
	if t.Item == nil {
		return Item{}, false
	}

	return *t.Item, true
}

func (t *Title) UnmarshalJSON(data []byte) error {
	// Creates an alias type that does not inherit Title's methods
	// thus preventing infinite recursion on unmarshal
	type Alias Title
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

type Item struct {
	Title   string   `json:"title"`
	Season  string   `json:"season"`
	Episode string   `json:"episode"`
	Type    ItemType `json:"type"`
}

type ItemType string

const (
	ItemTypeMovie        ItemType = "MainMovie"
	ItemTypeExtra        ItemType = "Extra"
	ItemTypeEpisode      ItemType = "Episode"
	ItemTypeDeletedScene ItemType = "DeletedScene"
	ItemTypeTrailer      ItemType = "Trailer"
)

func mediaItemResponseToDiscRecord(mediaItemResponse *MediaItemResponse, discHash string) (DiscRecord, error) {
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
				return DiscRecord{}, fmt.Errorf("multiple matching discs found for hash %s", discHash)
			}

			matchedDisc = disc
			matchedRelease = release
		}
	}

	if matchedDisc == nil || matchedRelease == nil {
		return DiscRecord{}, fmt.Errorf("no matching disc found for hash %s", discHash)
	}

	titles, err := titleResponsesToTitles(matchedDisc.Titles)
	if err != nil {
		return DiscRecord{}, fmt.Errorf("mapping title response to title: %w", err)
	}

	return DiscRecord{
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
			Titles:      titles,
		},
	}, nil
}

func titleResponsesToTitles(titleResponses []TitleResponse) ([]Title, error) {
	titles := make([]Title, 0, len(titleResponses))

	for _, titleResponse := range titleResponses {
		signature, err := signature.NormalizeSegments(titleResponse.SegmentMap)
		if err != nil {
			return nil, err
		}
		titles = append(titles, Title{
			Duration:    titleResponse.Duration,
			DisplaySize: titleResponse.DisplaySize,
			SourceFile:  titleResponse.SourceFile,
			Size:        titleResponse.Size,
			SegmentMap:  titleResponse.SegmentMap,
			Signature:   signature,
			Item:        itemResponseToItem(titleResponse.Item),
		})
	}

	return titles, nil
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

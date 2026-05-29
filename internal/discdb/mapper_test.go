package discdb

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMediaItemResponseToDiscRecord(t *testing.T) {
	validMediaItem := MediaItemResponse{
		Title: "Test Movie",
		Year:  2024,
		Type:  "Movie",
		Releases: []ReleaseResponse{
			{
				Slug:   "test-release",
				Locale: "en-us",
				Year:   2024,
				Title:  "Test Release",
				Discs: []DiscResponse{
					{
						ContentHash: "ABC123",
						Name:        "Disc 1",
						Format:      "Blu-ray",
						Slug:        "disc-1",
						Titles: []TitleResponse{
							{
								SegmentMap:  "1,2,3",
								Duration:    "2:00:00",
								DisplaySize: "25.0 GB",
								Size:        25000000000,
								Item: &ItemResponse{
									Title: "Test Movie",
									Type:  "MainMovie",
								},
							},
							{
								SegmentMap: "4,5,6",
								Size:       1000000,
								Item:       nil,
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name       string
		mediaItem  MediaItemResponse
		discHash   string
		wantRecord DiscRecord
		wantErr    bool
	}{
		{
			name:      "successful mapping",
			mediaItem: validMediaItem,
			discHash:  "ABC123",
			wantRecord: DiscRecord{
				Media: Media{
					Title: "Test Movie",
					Year:  2024,
					Type:  "Movie",
				},
				Release: Release{
					Slug:   "test-release",
					Locale: "en-us",
					Year:   2024,
					Title:  "Test Release",
				},
				Disc: Disc{
					ContentHash: "ABC123",
					Name:        "Disc 1",
					Format:      "Blu-ray",
					Slug:        "disc-1",
					Titles: []Title{
						{
							SegmentMap:  "1,2,3",
							Duration:    "2:00:00",
							DisplaySize: "25.0 GB",
							Size:        25000000000,
							Item: &Item{
								Title: "Test Movie",
								Type:  ItemTypeMovie,
							},
						},
						{
							SegmentMap: "4,5,6",
							Size:       1000000,
							Item:       nil,
						},
					},
				},
			},
		},
		{
			name:      "no matching disc",
			mediaItem: validMediaItem,
			discHash:  "NONEXISTENT",
			wantErr:   true,
		},
		{
			name:     "multiple matching discs",
			discHash: "ABC123",
			mediaItem: MediaItemResponse{
				Title: "Test Movie",
				Year:  2024,
				Type:  "Movie",
				Releases: []ReleaseResponse{
					{
						Discs: []DiscResponse{
							{ContentHash: "ABC123"},
							{ContentHash: "ABC123"},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mediaItemResponseToDiscRecord(&tt.mediaItem, tt.discHash)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.wantRecord, got); diff != "" {
				t.Errorf("record mismatch: %s", diff)
			}
		})
	}
}

package discdb

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSQLiteCache(t *testing.T) {
	newTestCache := func(t *testing.T) *SQLiteCache {
		t.Helper()
		path := filepath.Join(t.TempDir(), "test.db")
		cache, err := NewSQLiteCache(path)
		if err != nil {
			t.Fatalf("failed to create test cache: %v", err)
		}
		t.Cleanup(func() { cache.Close() })
		return cache
	}

	testRecord := DiscRecord{
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
					Signature:   "1,2,3",
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
					Signature:  "4,5,6",
					Size:       1000000,
					Item:       nil,
				},
			},
		},
	}

	tests := []struct {
		name  string
		setup func(t *testing.T, cache *SQLiteCache)
		run   func(t *testing.T, cache *SQLiteCache)
	}{
		{
			name:  "cache miss",
			setup: func(t *testing.T, cache *SQLiteCache) {},
			run: func(t *testing.T, cache *SQLiteCache) {
				_, ok, err := cache.GetDiscRecord(context.Background(), "nonexistent")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if ok {
					t.Error("expected cache miss, got hit")
				}
			},
		},
		{
			name: "put and get",
			setup: func(t *testing.T, cache *SQLiteCache) {
				if err := cache.PutDiscRecord(context.Background(), "ABC123", testRecord); err != nil {
					t.Fatalf("unexpected error on put: %v", err)
				}
			},
			run: func(t *testing.T, cache *SQLiteCache) {
				got, ok, err := cache.GetDiscRecord(context.Background(), "ABC123")
				if err != nil {
					t.Fatalf("unexpected error on get: %v", err)
				}
				if !ok {
					t.Fatal("expected cache hit, got miss")
				}
				if diff := cmp.Diff(testRecord, got); diff != "" {
					t.Errorf("record mismatch: %s", diff)
				}
			},
		},
		{
			name: "nullable item preserved",
			setup: func(t *testing.T, cache *SQLiteCache) {
				if err := cache.PutDiscRecord(context.Background(), "ABC123", testRecord); err != nil {
					t.Fatalf("unexpected error on put: %v", err)
				}
			},
			run: func(t *testing.T, cache *SQLiteCache) {
				got, _, err := cache.GetDiscRecord(context.Background(), "ABC123")
				if err != nil {
					t.Fatalf("unexpected error on get: %v", err)
				}
				if got.Disc.Titles[0].Item == nil {
					t.Error("expected non-nil Item on first title")
				}
				if got.Disc.Titles[1].Item != nil {
					t.Error("expected nil Item on second title")
				}
			},
		},
		{
			name: "put overwrites existing record",
			setup: func(t *testing.T, cache *SQLiteCache) {
				if err := cache.PutDiscRecord(context.Background(), "ABC123", testRecord); err != nil {
					t.Fatalf("unexpected error on first put: %v", err)
				}
			},
			run: func(t *testing.T, cache *SQLiteCache) {
				updated := testRecord
				updated.Media.Title = "Updated Title"
				if err := cache.PutDiscRecord(context.Background(), "ABC123", updated); err != nil {
					t.Fatalf("unexpected error on second put: %v", err)
				}
				got, _, err := cache.GetDiscRecord(context.Background(), "ABC123")
				if err != nil {
					t.Fatalf("unexpected error on get: %v", err)
				}
				if got.Media.Title != "Updated Title" {
					t.Errorf("expected updated title, got %q", got.Media.Title)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := newTestCache(t)
			tt.setup(t, cache)
			tt.run(t, cache)
		})
	}
}

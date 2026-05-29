package discdb

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type fakeCache struct {
	getRecord DiscRecord
	getOk     bool
	getErr    error
	putErr    error
	putCalled bool
}

func (f *fakeCache) GetDiscRecord(_ context.Context, _ string) (DiscRecord, bool, error) {
	return f.getRecord, f.getOk, f.getErr
}

func (f *fakeCache) PutDiscRecord(_ context.Context, _ string, _ DiscRecord) error {
	f.putCalled = true
	return f.putErr
}

type fakeClient struct {
	record DiscRecord
	err    error
	called bool
}

func (f *fakeClient) LookupDisc(_ context.Context, _ string) (DiscRecord, error) {
	f.called = true
	return f.record, f.err
}

func TestCachedClient(t *testing.T) {
	sampleRecord := DiscRecord{
		Media: Media{
			Title: "Test Movie",
			Year:  2024,
			Type:  "Movie",
		},
	}

	tests := []struct {
		name          string
		cache         *fakeCache
		client        *fakeClient
		wantRecord    DiscRecord
		wantErr       bool
		wantPutCalled bool
		wantClientCalled bool
	}{
		{
			name: "cache hit",
			cache: &fakeCache{
				getRecord: sampleRecord,
				getOk:     true,
			},
			client:           &fakeClient{},
			wantRecord:       sampleRecord,
			wantClientCalled: false,
			wantPutCalled:    false,
		},
		{
			name:  "cache miss calls client and caches result",
			cache: &fakeCache{},
			client: &fakeClient{
				record: sampleRecord,
			},
			wantRecord:       sampleRecord,
			wantClientCalled: true,
			wantPutCalled:    true,
		},
		{
			name: "cache read error",
			cache: &fakeCache{
				getErr: errors.New("cache read error"),
			},
			client:  &fakeClient{},
			wantErr: true,
		},
		{
			name:  "client error on miss",
			cache: &fakeCache{},
			client: &fakeClient{
				err: errors.New("client error"),
			},
			wantErr:          true,
			wantClientCalled: true,
			wantPutCalled:    false,
		},
		{
			name: "cache write error",
			cache: &fakeCache{
				putErr: errors.New("cache write error"),
			},
			client: &fakeClient{
				record: sampleRecord,
			},
			wantErr:          true,
			wantClientCalled: true,
			wantPutCalled:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := NewCachedClient(test.cache, test.client)
			if err != nil {
				t.Fatalf("unexpected error creating client: %v", err)
			}

			got, err := c.LookupDisc(context.Background(), "ABC123")
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.wantErr {
				return
			}
			if diff := cmp.Diff(test.wantRecord, got); diff != "" {
				t.Errorf("record mismatch: %s", diff)
			}
			if test.client.called != test.wantClientCalled {
				t.Errorf("client called: expected %v, got %v", test.wantClientCalled, test.client.called)
			}
			if test.cache.putCalled != test.wantPutCalled {
				t.Errorf("cache put called: expected %v, got %v", test.wantPutCalled, test.cache.putCalled)
			}
		})
	}
}

func TestRemoteClientLookupDisc(t *testing.T) {
    validResponse := `{
        "data": {
            "mediaItems": {
                "nodes": [{
                    "title": "Test Movie",
                    "year": 2024,
                    "type": "Movie",
                    "releases": [{
                        "slug": "test-release",
                        "locale": "en-us",
                        "year": 2024,
                        "title": "Test Release",
                        "discs": [{
                            "contentHash": "ABC123",
                            "name": "Disc 1",
                            "format": "Blu-ray",
                            "slug": "disc-1",
                            "titles": [{
                                "segmentMap": "1,2,3",
                                "duration": "2:00:00",
                                "displaySize": "25.0 GB",
                                "size": 25000000000,
                                "item": {
                                    "title": "Test Movie",
                                    "type": "MainMovie"
                                }
                            }]
                        }]
                    }]
                }]
            }
        }
    }`

    tests := []struct {
        name       string
        handler    http.HandlerFunc
        discHash   string
        wantRecord DiscRecord
        wantErr    bool
    }{
        {
            name:     "successful lookup",
            discHash: "ABC123",
            handler: func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
                fmt.Fprint(w, validResponse)
            },
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
                    },
                },
            },
        },
        {
            name:     "non-200 status",
            discHash: "ABC123",
            handler: func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusInternalServerError)
                fmt.Fprint(w, "internal server error")
            },
            wantErr: true,
        },
        {
            name:     "graphql errors",
            discHash: "ABC123",
            handler: func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
                fmt.Fprint(w, `{"errors": [{"message": "something went wrong"}]}`)
            },
            wantErr: true,
        },
        {
            name:     "no results",
            discHash: "ABC123",
            handler: func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
                fmt.Fprint(w, `{"data": {"mediaItems": {"nodes": []}}}`)
            },
            wantErr: true,
        },
        {
            name:     "multiple results",
            discHash: "ABC123",
            handler: func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
                fmt.Fprint(w, `{"data": {"mediaItems": {"nodes": [{}, {}]}}}`)
            },
            wantErr: true,
        },
        {
            name:     "invalid json response",
            discHash: "ABC123",
            handler: func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
                fmt.Fprint(w, `not json`)
            },
            wantErr: true,
        },
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            server := httptest.NewServer(test.handler)
            defer server.Close()

            client := &RemoteClient{
                endpoint:   server.URL,
                httpClient: server.Client(),
            }

            got, err := client.LookupDisc(context.Background(), test.discHash)
            if (err != nil) != test.wantErr {
                t.Fatalf("unexpected error: %v", err)
            }
            if test.wantErr {
                return
            }
            if diff := cmp.Diff(test.wantRecord, got); diff != "" {
                t.Errorf("record mismatch: %s", diff)
            }
        })
    }
}

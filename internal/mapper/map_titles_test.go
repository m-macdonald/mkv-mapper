package mapper

import (
	"testing"

	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/signature"

	"github.com/google/go-cmp/cmp"
)

func TestMapTitles(t *testing.T) {
	tests := []struct {
		name    string
		titles  []makemkv.Title
		discRecord discdb.DiscRecord
		want    []TitleMapping
		wantErr bool
	}{
		{
			name: "successful mapping",
			discRecord: discdb.DiscRecord{
				Disc: discdb.Disc{
					Titles: []discdb.Title{
						{
							SegmentMap: "05, 07",
						},
						{
							SegmentMap: "01,05,06",
						},
					},
				},
			},
			titles: []makemkv.Title{
				{
					Segments: "01, 05, 06",
				},
				{
					Segments: "05, 07",
				},
				{
					Segments: "06, 010",
				},
			},
			want: []TitleMapping{
				{
					MakeMkvTitle: makemkv.Title{
						Segments: "05, 07",
					},
					DiscDbTitle: discdb.Title{
						SegmentMap: "05, 07",
					},
				},
				{
					MakeMkvTitle: makemkv.Title{
						Segments: "01, 05, 06",
					},
					DiscDbTitle: discdb.Title{
						SegmentMap: "01,05,06",
					},
				},
			},
		},
		{
			name: "error grouping",
			discRecord: discdb.DiscRecord{
				Disc: discdb.Disc{
					Titles: []discdb.Title{
						{
							SegmentMap: "05, 07",
						},
						{
							SegmentMap: "01,05,06",
						},
					},
				},
			},
			titles: []makemkv.Title{
				{
					Segments: "01, 05, 06",
				},
				{
					Segments: "kaboom",
				},
				{
					Segments: "06, 010",
				},
			},
			wantErr: true,
		},
		{
			name: "error normalizing discdb.Title.SegmentMap",
			discRecord: discdb.DiscRecord{
				Disc: discdb.Disc{
					Titles: []discdb.Title{
						{
							SegmentMap: "kaboom",
						},
						{
							SegmentMap: "01,05,06",
						},
					},
				},
			},
			titles: []makemkv.Title{
				{
					Segments: "01, 05, 06",
				},
				{
					Segments: "07, 010",
				},
				{
					Segments: "06, 010",
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			got, err := MapTitles(test.discRecord, test.titles)

			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected err: %v", err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("mapping did not match (-want +got)\n%s", diff)
			}
		})
	}
}

func TestGroupBySegmentSignature(t *testing.T) {
	tests := []struct {
		name    string
		titles  []makemkv.Title
		want    map[signature.SegmentSignature]makemkv.Title
		wantErr bool
	}{
		{
			name: "successful grouping",
			titles: []makemkv.Title{
				{
					SourceFilename: "file1.mkv",
					OutputFilename: "file1-out.mkv",
					Segments:       "01,02",
				},
				{
					SourceFilename: "file2.mkv",
					OutputFilename: "file2-out.mkv",
					Segments:       "02,05",
				},
			},
			want: map[signature.SegmentSignature]makemkv.Title{
				signature.SegmentSignature("00001|00002"): {
					SourceFilename: "file1.mkv",
					OutputFilename: "file1-out.mkv",
					Segments:       "01,02",
				},
				signature.SegmentSignature("00002|00005"): {
					SourceFilename: "file2.mkv",
					OutputFilename: "file2-out.mkv",
					Segments:       "02,05",
				},
			},
		},
		{
			name: "error",
			titles: []makemkv.Title{
				{
					SourceFilename: "file1.mkv",
					OutputFilename: "file1-out.mkv",
					Segments:       "kaboom",
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := groupBySegmentSignature(test.titles)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected err = %v", err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("title grouping mismatch (-want +got): \n%s", diff)
			}
		})
	}
}

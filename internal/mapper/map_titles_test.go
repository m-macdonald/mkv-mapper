package mapper

import (
	"testing"

	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	th "m-macdonald/mkv-mapper/internal/mkvmappertest"
	"m-macdonald/mkv-mapper/internal/signature"

	"github.com/google/go-cmp/cmp"
)

func TestMapTitles(t *testing.T) {
	tests := []struct {
		name       string
		titles     []makemkv.Title
		discRecord discdb.DiscRecord
		want       []TitleMapping
		wantErr    bool
	}{
		{
			name: "successful mapping",
			discRecord: th.NewDiscRecord(
				th.WithTitles(
					th.NewDiscTitle(th.WithSegmentMap("05, 07")),
					th.NewDiscTitle(th.WithSegmentMap("01, 05, 06")))),
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSegments("01, 05, 06")),
				th.NewMakeMkvTitle(th.WithSegments("05, 07")),
				th.NewMakeMkvTitle(th.WithSegments("06, 010")),
			},
			want: []TitleMapping{
				newTitleMapping("05, 07"),
				newTitleMapping("01, 05, 06"),
			},
		},
		{
			name: "error grouping",
			discRecord: th.NewDiscRecord(
				th.WithTitles(
					th.NewDiscTitle(th.WithSegmentMap("05, 07")),
					th.NewDiscTitle(th.WithSegmentMap("01, 05, 06")))),
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSegments("01, 05, 06")),
				th.NewMakeMkvTitle(th.WithSegments("  ")),
				th.NewMakeMkvTitle(th.WithSegments("06, 010")),
			},
			wantErr: true,
		},
		{
			name: "error normalizing discdb.Title.SegmentMap",
			discRecord: th.NewDiscRecord(
				th.WithTitles(
					th.NewDiscTitle(th.WithSegmentMap("  ")),
					th.NewDiscTitle(th.WithSegmentMap("01, 05, 06")))),
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSegments("01, 05, 06")),
				th.NewMakeMkvTitle(th.WithSegments("07, 010")),
				th.NewMakeMkvTitle(th.WithSegments("06, 010")),
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
				th.NewMakeMkvTitle(th.WithSegments("01,02")),
				th.NewMakeMkvTitle(th.WithSegments("02,05")),
			},
			want: map[signature.SegmentSignature]makemkv.Title{
				signature.SegmentSignature("01,02"): th.NewMakeMkvTitle(th.WithSegments("01,02")),
				signature.SegmentSignature("02,05"): th.NewMakeMkvTitle(th.WithSegments("02,05")),
			},
		},
		{
			name: "error",
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSegments("  ")),
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

func newTitleMapping(segmentMap string) TitleMapping {
	return TitleMapping{
		MakeMkvTitle: th.NewMakeMkvTitle(th.WithSegments(segmentMap)),
		DiscDbTitle:  th.NewDiscTitle(th.WithSegmentMap(segmentMap)),
	}
}

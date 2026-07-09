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
	}{
		{
			name: "successful mapping",
			discRecord: th.NewDiscRecord(
				th.WithTitles(
					th.NewDiscTitle(th.WithSegmentMap("05,07")),
					th.NewDiscTitle(th.WithSegmentMap("01,05,06")))),
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSignature("01,05,06")),
				th.NewMakeMkvTitle(th.WithSignature("05,07")),
				th.NewMakeMkvTitle(th.WithSignature("06,010")),
			},
			want: []TitleMapping{
				newTitleMapping("01,05,06", true),
				newTitleMapping("05,07", true),
				newTitleMapping("06,010", false),
			},
		},
		{
			name: "unmatched makemkv title",
			discRecord: th.NewDiscRecord(
				th.WithTitles(
					th.NewDiscTitle(th.WithSegmentMap("05,07")),
				),
			),
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSignature("06,010")),
			},
			want: []TitleMapping{
				newTitleMapping("06,010", false),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := MapTitles(test.discRecord, test.titles)

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("mapping did not match (-want +got)\n%s", diff)
			}
		})
	}
}

func TestGroupMakeMkvBySignature(t *testing.T) {
	tests := []struct {
		name   string
		titles []makemkv.Title
		want   []makemkv.Title
	}{
		{
			name: "successful grouping",
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSignature("01,02")),
				th.NewMakeMkvTitle(th.WithSignature("02,05")),
			},
			want: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSignature("01,02")),
				th.NewMakeMkvTitle(th.WithSignature("02,05")),
			},
		},
		{
			name: "ignores previously identified segment maps",
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSignature("01,02"), th.WithTitleId(1)),
				th.NewMakeMkvTitle(th.WithSignature("01,02"), th.WithTitleId(2)),
			},
			want: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSignature("01,02"), th.WithTitleId(1)),
			},
		},
		{
			name:   "empty slice",
			titles: []makemkv.Title{},
			want:   []makemkv.Title{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := groupMakeMkvBySignature(test.titles)

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("title grouping mismatch (-want +got): \n%s", diff)
			}
		})
	}
}

func TestGroupDiscDbBySignature(t *testing.T) {
	tests := []struct {
		name   string
		titles []discdb.Title
		want   map[signature.SegmentSignature]discdb.Title
	}{
		{
			name: "successful grouping",
			titles: []discdb.Title{
				th.NewDiscTitle(th.WithSegmentMap("01,02")),
				th.NewDiscTitle(th.WithSegmentMap("02,05")),
			},
			want: map[signature.SegmentSignature]discdb.Title{
				"01,02": th.NewDiscTitle(th.WithSegmentMap("01,02")),
				"02,05": th.NewDiscTitle(th.WithSegmentMap("02,05")),
			},
		},
		{
			name: "ignores previously identified segment maps",
			titles: []discdb.Title{
				th.NewDiscTitle(th.WithSegmentMap("01,02"), th.WithItem(&discdb.Item{Season: "1"})),
				th.NewDiscTitle(th.WithSegmentMap("01,02"), th.WithItem(&discdb.Item{Season: "2"})),
			},
			want: map[signature.SegmentSignature]discdb.Title{
				"01,02": th.NewDiscTitle(th.WithSegmentMap("01,02"), th.WithItem(&discdb.Item{Season: "1"})),
			},
		},
		{
			name:   "empty map",
			titles: []discdb.Title{},
			want:   map[signature.SegmentSignature]discdb.Title{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := groupDiscDbBySignature(test.titles)

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("title grouping mismatch (-want +got): \n%s", diff)
			}
		})
	}
}

func newTitleMapping(sig string, includeTitle bool) TitleMapping {
	var title discdb.Title
	if includeTitle {
		title = th.NewDiscTitle(th.WithSegmentMap(sig))
	}
	return TitleMapping{
		MakeMkvTitle: th.NewMakeMkvTitle(th.WithSignature(signature.SegmentSignature(sig))),
		DiscDbTitle:  title,
	}
}

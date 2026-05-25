package mapper

import (
	"testing"

	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/mkvmappertest"
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
			discRecord: mkvmappertest.NewDiscRecord(
				mkvmappertest.NewDiscTitle("05, 07"), 
				mkvmappertest.NewDiscTitle("01, 05, 06")),
			titles: []makemkv.Title{
				mkvmappertest.NewMakeMkvTitle("01, 05, 06"),
				mkvmappertest.NewMakeMkvTitle("05, 07"),
				mkvmappertest.NewMakeMkvTitle("06, 010"),
			},
			want: []TitleMapping{
				newTitleMapping("05, 07"),
				newTitleMapping("01, 05, 06"),
			},
		},
		{
			name: "error grouping",
			discRecord: mkvmappertest.NewDiscRecord(
				mkvmappertest.NewDiscTitle("05, 07"),
				mkvmappertest.NewDiscTitle("01, 05, 06")),
			titles: []makemkv.Title{
				mkvmappertest.NewMakeMkvTitle("01, 05, 06"),
				mkvmappertest.NewMakeMkvTitle("kaboom"),
				mkvmappertest.NewMakeMkvTitle("06, 010"),
			},
			wantErr: true,
		},
		{
			name: "error normalizing discdb.Title.SegmentMap",
			discRecord: mkvmappertest.NewDiscRecord(
				mkvmappertest.NewDiscTitle("kaboom"),
				mkvmappertest.NewDiscTitle("01, 05, 06")),
			titles: []makemkv.Title{
				mkvmappertest.NewMakeMkvTitle("01, 05, 06"),
				mkvmappertest.NewMakeMkvTitle("07, 010"),
				mkvmappertest.NewMakeMkvTitle("06, 010"),
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
				mkvmappertest.NewMakeMkvTitle("01,02"),
				mkvmappertest.NewMakeMkvTitle("02,05"),
			},
			want: map[signature.SegmentSignature]makemkv.Title{
				signature.SegmentSignature("00001|00002"): mkvmappertest.NewMakeMkvTitle("01,02"),
				signature.SegmentSignature("00002|00005"): mkvmappertest.NewMakeMkvTitle("02,05"),
			},
		},
		{
			name: "error",
			titles: []makemkv.Title{
				mkvmappertest.NewMakeMkvTitle("kaboom"),
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
		MakeMkvTitle: mkvmappertest.NewMakeMkvTitle(segmentMap),
		DiscDbTitle: mkvmappertest.NewDiscTitle(segmentMap),
	}
}

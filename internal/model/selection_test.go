package model

import (
	"fmt"
	"testing"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"

	"github.com/google/go-cmp/cmp"
)

func TestFullSelection(t *testing.T) {
	tests := []struct {
		name string
		plan Plan
		want Selection
	}{
		{
			name: "selects all titles",
			plan: Plan{
				PlanBase: PlanBase{
					Titles: []TitlePlan{
						{
							IsMatched: true,
							TitleId: 1,
						},
						{
							IsMatched: false,
							TitleId: 2,
						},
						{
							IsMatched: true,
							TitleId: 3,
						},
					},
				},
			},
			want: Selection{
				Mode: config.ModeFullAuto,
				SelectedIds: []lines.TitleId{1, 2, 3},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := FullSelection(test.plan)

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("selection did not match (-want +got)\n%s", diff)
			}
		})
	}
}

func TestTrimmedSelection(t *testing.T) {
	tests := []struct {
		name string
		plan Plan
		want Selection
	}{
		{
			name: "selects only matched titles",
			plan: Plan{
				PlanBase: PlanBase{
					Titles: []TitlePlan{
						{
							IsMatched: true,
							TitleId: 1,
						},
						{
							IsMatched: false,
							TitleId: 2,
						},
						{
							IsMatched: true,
							TitleId: 3,
						},
					},
				},
			},
			want: Selection{
				Mode: config.ModeTrimmedAuto,
				SelectedIds: []lines.TitleId{1, 3},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := TrimmedSelection(test.plan)

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("selection did not match (-want +got)\n%s", diff)
			}
		})
	}
}

func testErr() error {
	return fmt.Errorf("test error")
}

package engine

import (
	"fmt"
	"testing"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/event"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	th "m-macdonald/mkv-mapper/internal/mkvmappertest"
	"m-macdonald/mkv-mapper/internal/model"

	"github.com/google/go-cmp/cmp"
)

func TestParsedLineToEvent(t *testing.T) {
	tests := []struct {
		name      string
		line      lines.ParsedLine
		wantEvent event.Event
		wantOk    bool
	}{
		{
			name: "progress value",
			line: lines.ProgressValue{
				Current: 23,
				Total:   60,
				Max:     100,
			},
			wantEvent: event.ProgressPercentEvent{
				TotalPercent:   60,
				CurrentPercent: 23,
			},
			wantOk: true,
		},
		{
			name:      "progress current",
			line:      lines.ProgressCurrent{Name: "Ripping title 1"},
			wantEvent: event.ProgressCurrentEvent{Message: "Ripping title 1"},
			wantOk:    true,
		},
		{
			name:      "progress title",
			line:      lines.ProgressTitle{Name: "Overall progress"},
			wantEvent: event.ProgressTotalEvent{Message: "Overall progress"},
			wantOk:    true,
		},
		{
			name:      "message",
			line:      lines.Message{Message: "Some message"},
			wantEvent: event.MessageEvent{Message: "Some message"},
			wantOk:    true,
		},
		{
			name:   "unrecognized line",
			line:   lines.StreamInfo{},
			wantOk: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := parsedLineToEvent(test.line)
			if ok != test.wantOk {
				t.Fatalf("expected ok %v, got %v", test.wantOk, ok)
			}
			if !test.wantOk {
				return
			}
			if diff := cmp.Diff(test.wantEvent, got); diff != "" {
				t.Errorf("event mismatch: %s", diff)
			}
		})
	}
}

func TestSelectPlanFullAuto(t *testing.T) {
	plan := th.TestPlan(
		th.TestTitlePlan(1, "Movie.mkv", 5*th.GB, true),
		th.TestTitlePlan(2, "Extra.mkv", 1*th.GB, false),
		th.TestTitlePlan(3, "Trailer.mkv", 500*th.MB, true),
	)

	eng := Engine{selector: th.NewSelector(model.Selection{})}
	selectedPlan, err := eng.SelectPlan(config.ModeFullAuto, plan)
	if err != nil {
		t.Fatalf("SelectPlan failed: %v", err)
	}
	if len(selectedPlan.Titles) != 3 {
		t.Errorf("full-auto: expected 3 titles, got %d", len(selectedPlan.Titles))
	}
	if !selectedPlan.IsAllTitles {
		t.Error("full-auto: IsAllTitles should be true")
	}
}

func TestSelectPlanTrimmedAuto(t *testing.T) {
	plan := th.TestPlan(
		th.TestTitlePlan(1, "Movie.mkv", 5*th.GB, true),
		th.TestTitlePlan(2, "Extra.mkv", 1*th.GB, false),
		th.TestTitlePlan(3, "Trailer.mkv", 500*th.MB, true),
	)

	eng := Engine{selector: th.NewSelector(model.Selection{})}
	selectedPlan, err := eng.SelectPlan(config.ModeTrimmedAuto, plan)
	if err != nil {
		t.Fatalf("SelectPlan failed: %v", err)
	}
	if len(selectedPlan.Titles) != 2 {
		t.Errorf("trimmed-auto: expected 2 matched titles, got %d", len(selectedPlan.Titles))
	}
	if selectedPlan.IsAllTitles {
		t.Error("trimmed-auto: IsAllTitles should be false")
	}
}

func TestSelectPlanManualSubset(t *testing.T) {
	plan := th.TestPlan(
		th.TestTitlePlan(1, "Movie.mkv", 5*th.GB, true),
		th.TestTitlePlan(2, "Extra.mkv", 1*th.GB, false),
		th.TestTitlePlan(3, "Trailer.mkv", 500*th.MB, true),
	)

	userSelection := th.Selection(config.ModeManual, 1, 3)
	eng := Engine{selector: th.NewSelector(userSelection)}
	selectedPlan, err := eng.SelectPlan(config.ModeManual, plan)
	if err != nil {
		t.Fatalf("SelectPlan failed: %v", err)
	}
	if len(selectedPlan.Titles) != 2 {
		t.Errorf("manual subset: expected 2 titles, got %d", len(selectedPlan.Titles))
	}
	if selectedPlan.IsAllTitles {
		t.Error("manual subset: IsAllTitles should be false")
	}
}

func TestSelectPlanManualAllTitles(t *testing.T) {
	plan := th.TestPlan(
		th.TestTitlePlan(1, "Movie.mkv", 5*th.GB, true),
		th.TestTitlePlan(2, "Extra.mkv", 1*th.GB, false),
	)

	userSelection := th.Selection(config.ModeManual, 1, 2)
	eng := Engine{selector: th.NewSelector(userSelection)}
	selectedPlan, err := eng.SelectPlan(config.ModeManual, plan)
	if err != nil {
		t.Fatalf("SelectPlan failed: %v", err)
	}
	if !selectedPlan.IsAllTitles {
		t.Error("manual all: IsAllTitles should be true")
	}
}

func TestSelectPlanSelectorError(t *testing.T) {
	plan := th.TestPlan(th.TestTitlePlan(1, "Movie.mkv", 5*th.GB, true))
	testErr := fmt.Errorf("user cancelled")
	eng := Engine{selector: th.NewSelectorWithError(testErr)}
	_, err := eng.SelectPlan(config.ModeManual, plan)

	if err == nil {
		t.Error("SelectPlan should propagate selector error")
	}
}

func TestValidatePlanConstruction(t *testing.T) {
	selectedPlan := model.SelectedPlan{
		PlanBase: model.PlanBase{
			OutputDir: "/test/output",
			DiscRoot:  "/test/disc",
			Titles: []model.TitlePlan{
				{TitleId: 1, FinalName: "Movie.mkv", EstimatedSize: 5 * th.GB},
			},
		},
		BuildReport: model.BuildReport{Warnings: make([]model.PlanWarning, 0)},
	}

	eng := Engine{selector: th.NewSelector(model.Selection{})}
	validatedPlan := eng.ValidatePlan(selectedPlan)

	if validatedPlan.OutputDir != "/test/output" {
		t.Error("ValidatePlan: OutputDir not carried forward")
	}
	if len(validatedPlan.Titles) != 1 {
		t.Error("ValidatePlan: Titles not carried forward")
	}
}

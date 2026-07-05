package planner

import (
	"fmt"
	"testing"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/engine"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/mkvmappertest"
)

func TestSelectPlanFullAuto(t *testing.T) {
	plan := testPlan(
		testTitle(1, "Movie.mkv", 5*GB, true),
		testTitle(2, "Extra.mkv", 1*GB, false),
		testTitle(3, "Trailer.mkv", 500*MB, true),
	)

	eng := testEngine(mkvmappertest.NewSelector(planner.Selection{}))
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
	plan := testPlan(
		testTitle(1, "Movie.mkv", 5*GB, true),
		testTitle(2, "Extra.mkv", 1*GB, false),
		testTitle(3, "Trailer.mkv", 500*MB, true),
	)

	eng := testEngine(mkvmappertest.NewSelector(planner.Selection{}))
	selectedPlan, err := eng.SelectPlan(config.ModeTrimmedAuto, plan)
	if err != nil {
		t.Fatalf("SelectPlan failed: %v", err)
	}
	if len(selectedPlan.Titles) != 2 {
		t.Errorf("trimmed-auto: expected 2 matched titles, got %d", len(selectedPlan.Titles))
	}
	// Check that only matched titles remain
	for _, title := range selectedPlan.Titles {
		if !title.IsMatched {
			t.Errorf("trimmed-auto: unmatched title %d in results", title.TitleId)
		}
	}
	if selectedPlan.IsAllTitles {
		t.Error("trimmed-auto: IsAllTitles should be false (one title excluded)")
	}
}

func TestSelectPlanManualSubset(t *testing.T) {
	plan := testPlan(
		testTitle(1, "Movie.mkv", 5*GB, true),
		testTitle(2, "Extra.mkv", 1*GB, false),
		testTitle(3, "Trailer.mkv", 500*MB, true),
	)

	// User selects titles 1 and 3, skipping 2
	userSelection := mkvmappertest.Selection(config.ModeManual, 1, 3)
	eng := testEngine(mkvmappertest.NewSelector(userSelection))
	selectedPlan, err := eng.SelectPlan(config.ModeManual, plan)
	if err != nil {
		t.Fatalf("SelectPlan failed: %v", err)
	}
	if len(selectedPlan.Titles) != 2 {
		t.Errorf("manual: expected 2 titles, got %d", len(selectedPlan.Titles))
	}

	// Verify correct titles were selected
	selectedIds := make(map[lines.TitleId]bool)
	for _, title := range selectedPlan.Titles {
		selectedIds[title.TitleId] = true
	}
	if !selectedIds[1] || !selectedIds[3] {
		t.Error("manual: wrong titles selected")
	}
	if selectedIds[2] {
		t.Error("manual: title 2 should not be selected")
	}
	if selectedPlan.IsAllTitles {
		t.Error("manual: IsAllTitles should be false (one title excluded)")
	}
}

func TestSelectPlanManualAllTitles(t *testing.T) {
	plan := testPlan(
		testTitle(1, "Movie.mkv", 5*GB, true),
		testTitle(2, "Extra.mkv", 1*GB, false),
	)

	// User selects all titles
	userSelection := mkvmappertest.Selection(config.ModeManual, 1, 2)
	eng := testEngine(mkvmappertest.NewSelector(userSelection))
	selectedPlan, err := eng.SelectPlan(config.ModeManual, plan)
	if err != nil {
		t.Fatalf("SelectPlan failed: %v", err)
	}
	if len(selectedPlan.Titles) != 2 {
		t.Errorf("manual all: expected 2 titles, got %d", len(selectedPlan.Titles))
	}
	if !selectedPlan.IsAllTitles {
		t.Error("manual all: IsAllTitles should be true")
	}
}

func TestSelectPlanUnknownTitleId(t *testing.T) {
	plan := testPlan(
		testTitle(1, "Movie.mkv", 5*GB, true),
		testTitle(2, "Extra.mkv", 1*GB, false),
	)

	// User somehow selects a title ID that doesn't exist
	userSelection := mkvmappertest.Selection(config.ModeManual, 1, 999)
	eng := testEngine(mkvmappertest.NewSelector(userSelection))
	_, err := eng.SelectPlan(config.ModeManual, plan)

	if err == nil {
		t.Error("SelectPlan should return error for unknown title ID")
	}
}

func TestSelectPlanSelectorError(t *testing.T) {
	plan := testPlan(testTitle(1, "Movie.mkv", 5*GB, true))

	// Selector returns an error (e.g., user cancelled)
	eng := testEngine(mkvmappertest.NewSelectorWithError(testErr))
	_, err := eng.SelectPlan(config.ModeManual, plan)

	if err == nil {
		t.Error("SelectPlan should propagate selector error")
	}
}

func TestSelectPlanCarriesBuildReport(t *testing.T) {
	plan := testPlan(testTitle(1, "Movie.mkv", 5*GB, true))
	plan.BuildReport.Warnings = append(plan.BuildReport.Warnings, planner.PlanWarning{
		TitleId: 1,
		Code:    planner.WarningNoMetadata,
		Message: "test warning",
	})

	eng := testEngine(mkvmappertest.NewSelector(mkvmappertest.Selection(config.ModeManual, 1)))
	selectedPlan, err := eng.SelectPlan(config.ModeManual, plan)
	if err != nil {
		t.Fatalf("SelectPlan failed: %v", err)
	}
	if len(selectedPlan.BuildReport.Warnings) != 1 {
		t.Errorf("SelectPlan: expected 1 warning, got %d", len(selectedPlan.BuildReport.Warnings))
	}
}

func TestSelectPlanCarriesPlanBase(t *testing.T) {
	plan := testPlan(
		testTitle(1, "Movie.mkv", 5*GB, true),
		testTitle(2, "Extra.mkv", 1*GB, false),
	)

	eng := testEngine(mkvmappertest.NewSelector(mkvmappertest.Selection(config.ModeManual, 1)))
	selectedPlan, err := eng.SelectPlan(config.ModeManual, plan)
	if err != nil {
		t.Fatalf("SelectPlan failed: %v", err)
	}
	if selectedPlan.OutputDir != plan.OutputDir {
		t.Error("SelectPlan: OutputDir not carried forward")
	}
	if selectedPlan.DiscRoot != plan.DiscRoot {
		t.Error("SelectPlan: DiscRoot not carried forward")
	}
	if selectedPlan.MediaInfo != plan.MediaInfo {
		t.Error("SelectPlan: MediaInfo not carried forward")
	}
}

// helpers
var (
	testErr = func() error { return fmt.Errorf("test error") }()
)

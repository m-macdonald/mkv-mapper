package engine

import (
	"m-macdonald/mkv-mapper/internal/model"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateOutputDir(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantStatus model.ValidationStatus
		wantCode   model.ValidationCode
	}{
		{
			name: "valid output directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantStatus: model.ValidationStatusPass,
		},
		{
			name: "directory does not exist",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantStatus: model.ValidationStatusFail,
			wantCode:   model.ValidationOutputDirInvalid,
		},
		{
			name: "output path is a file not a directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "file.txt")
				if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantStatus: model.ValidationStatusFail,
			wantCode:   model.ValidationOutputDirInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir := test.setup(t)
			plan := model.SelectedPlan{
				PlanBase: model.PlanBase{
					OutputDir: outputDir,
				},
			}
			report := &model.ValidationReport{}
			validateOutputDir(plan, report)

			if len(report.Results) == 0 {
				t.Fatal("expected at least one result")
			}
			result := report.Results[0]
			if result.Status != test.wantStatus {
				t.Errorf("expected status %q, got %q", test.wantStatus, result.Status)
			}
			if test.wantCode != "" && result.Code != test.wantCode {
				t.Errorf("expected code %q, got %q", test.wantCode, result.Code)
			}
		})
	}
}

func TestValidateDiskSpace(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		titles     []model.TitlePlan
		wantStatus model.ValidationStatus
		wantCode   model.ValidationCode
	}{
		{
			name: "sufficient disk space",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			titles:     []model.TitlePlan{{EstimatedSize: 1000}},
			wantStatus: model.ValidationStatusPass,
		},
		{
			name: "insufficient disk space",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			titles:     []model.TitlePlan{{EstimatedSize: math.MaxUint64}},
			wantStatus: model.ValidationStatusFail,
			wantCode:   model.ValidationInsufficientSpace,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir := test.setup(t)
			plan := model.SelectedPlan{
				PlanBase: model.PlanBase{
					OutputDir: outputDir,
					Titles:    test.titles,
				},
			}
			report := &model.ValidationReport{}
			validateDiskSpace(plan, report)

			if len(report.Results) == 0 {
				t.Fatal("expected at least one result")
			}
			result := report.Results[0]
			if result.Status != test.wantStatus {
				t.Errorf("expected status %q, got %q", test.wantStatus, result.Status)
			}
			if test.wantCode != "" && result.Code != test.wantCode {
				t.Errorf("expected code %q, got %q", test.wantCode, result.Code)
			}
		})
	}
}

func TestValidateExistingFiles(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) (string, []model.TitlePlan)
		wantStatus model.ValidationStatus
		wantCode   model.ValidationCode
	}{
		{
			name: "no existing files",
			setup: func(t *testing.T) (string, []model.TitlePlan) {
				dir := t.TempDir()
				return dir, []model.TitlePlan{
					{FinalName: "movie.mkv", TitleId: 1},
				}
			},
			wantStatus: model.ValidationStatusPass,
		},
		{
			name: "existing file conflict",
			setup: func(t *testing.T) (string, []model.TitlePlan) {
				dir := t.TempDir()
				path := filepath.Join(dir, "movie.mkv")
				if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
					t.Fatal(err)
				}
				return dir, []model.TitlePlan{
					{FinalName: "movie.mkv", TitleId: 1},
				}
			},
			wantStatus: model.ValidationStatusFail,
			wantCode:   model.ValidationOutputExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir, titles := test.setup(t)
			plan := model.SelectedPlan{
				PlanBase: model.PlanBase{
					OutputDir: outputDir,
					Titles:    titles,
				},
			}
			report := &model.ValidationReport{}
			validateExistingFiles(plan, report)

			if len(report.Results) == 0 {
				t.Fatal("expected at least one result")
			}
			result := report.Results[0]
			if result.Status != test.wantStatus {
				t.Errorf("expected status %q, got %q", test.wantStatus, result.Status)
			}
			if test.wantCode != "" && result.Code != test.wantCode {
				t.Errorf("expected code %q, got %q", test.wantCode, result.Code)
			}
		})
	}
}

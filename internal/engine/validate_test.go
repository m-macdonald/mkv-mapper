package engine

import (
	"m-macdonald/mkv-mapper/internal/planner"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateOutputDir(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantStatus ValidationStatus
		wantCode   ValidationCode
	}{
		{
			name: "valid output directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantStatus: ValidationStatusPass,
		},
		{
			name: "directory does not exist",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantStatus: ValidationStatusFail,
			wantCode:   ValidationOutputDirInvalid,
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
			wantStatus: ValidationStatusFail,
			wantCode:   ValidationOutputDirInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir := test.setup(t)
			plan := planner.DiscPlan{OutputDir: outputDir}
			report := &ValidationReport{}
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
		titles     []planner.TitlePlan
		wantStatus ValidationStatus
		wantCode   ValidationCode
	}{
		{
			name: "sufficient disk space",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			titles:     []planner.TitlePlan{{EstimatedSize: 1000}},
			wantStatus: ValidationStatusPass,
		},
		{
			name: "insufficient disk space",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			// max uint64 value will always exceed available space
			titles:     []planner.TitlePlan{{EstimatedSize: ^uint64(0)}},
			wantStatus: ValidationStatusFail,
			wantCode:   ValidationInsufficientSpace,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir := test.setup(t)
			plan := planner.DiscPlan{
				OutputDir: outputDir,
				Titles:    test.titles,
			}
			report := &ValidationReport{}
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
		setup      func(t *testing.T) (string, []planner.TitlePlan)
		wantStatus ValidationStatus
		wantCode   ValidationCode
	}{
		{
			name: "no existing files",
			setup: func(t *testing.T) (string, []planner.TitlePlan) {
				dir := t.TempDir()
				return dir, []planner.TitlePlan{
					{FinalName: "movie.mkv", TitleId: 1},
				}
			},
			wantStatus: ValidationStatusPass,
		},
		{
			name: "existing file conflict",
			setup: func(t *testing.T) (string, []planner.TitlePlan) {
				dir := t.TempDir()
				path := filepath.Join(dir, "movie.mkv")
				if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
					t.Fatal(err)
				}
				return dir, []planner.TitlePlan{
					{FinalName: "movie.mkv", TitleId: 1},
				}
			},
			wantStatus: ValidationStatusFail,
			wantCode:   ValidationOutputExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir, titles := test.setup(t)
			plan := planner.DiscPlan{
				OutputDir: outputDir,
				Titles:    titles,
			}
			report := &ValidationReport{}
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

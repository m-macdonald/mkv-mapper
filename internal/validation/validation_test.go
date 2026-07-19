package validation

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateOutputDir(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantStatus Status
		wantCode   Code
	}{
		{
			name: "valid output directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantStatus: StatusPass,
		},
		{
			name: "directory does not exist",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantStatus: StatusFail,
			wantCode:   OutputDirInvalid,
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
			wantStatus: StatusFail,
			wantCode:   OutputDirInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir := test.setup(t)
			checkFunc := OutputDirValid(outputDir)
			got := checkFunc(context.Background())

			if len(got) == 0 {
				t.Fatal("expected at least one result")
			}
			result := got[0]
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
		setup      func(t *testing.T) (string, uint64)
		wantStatus Status
		wantCode   Code
	}{
		{
			name: "sufficient disk space",
			setup: func(t *testing.T) (string, uint64) {
				return t.TempDir(), 1024
			},
			wantStatus: StatusPass,
		},
		{
			name: "insufficient disk space",
			setup: func(t *testing.T) (string, uint64) {
				return t.TempDir(), math.MaxUint64
			},
			wantStatus: StatusFail,
			wantCode:   InsufficientSpace,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir, needed := test.setup(t)
			checkFunc := FreeSpace(outputDir, needed)
			got := checkFunc(context.Background())

			if len(got) == 0 {
				t.Fatal("expected at least one result")
			}
			result := got[0]
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
		setup      func(t *testing.T) (string, []FilenameTarget)
		wantStatus Status
		wantCode   Code
	}{
		{
			name: "no existing files",
			setup: func(t *testing.T) (string, []FilenameTarget) {
				dir := t.TempDir()
				return dir, []FilenameTarget{
					{ID: "1", FileName: "movie.mkv"},
				}
			},
			wantStatus: StatusPass,
		},
		{
			name: "existing file conflict",
			setup: func(t *testing.T) (string, []FilenameTarget) {
				dir := t.TempDir()
				path := filepath.Join(dir, "movie.mkv")
				if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
					t.Fatal(err)
				}
				return dir, []FilenameTarget{
					{ID: "1", FileName: "movie.mkv"},
				}
			},
			wantStatus: StatusFail,
			wantCode:   OutputExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir, fileNames := test.setup(t)

			checkFunc := NoConflicts(outputDir, fileNames)
			got := checkFunc(context.Background())

			if len(got) == 0 {
				t.Fatal("expected at least one result")
			}
			result := got[0]
			if result.Status != test.wantStatus {
				t.Errorf("expected status %q, got %q", test.wantStatus, result.Status)
			}
			if test.wantCode != "" && result.Code != test.wantCode {
				t.Errorf("expected code %q, got %q", test.wantCode, result.Code)
			}
		})
	}
}

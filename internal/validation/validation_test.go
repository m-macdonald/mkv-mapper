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

func TestFreeSpaceShared(t *testing.T) {
	// Both requirements live under the same temp dir, so they share a filesystem.
	root := t.TempDir()
	ripDir := filepath.Join(root, "rip")
	backupDir := filepath.Join(root, "backup")
	for _, dir := range []string{ripDir, backupDir} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	half := uint64(math.MaxUint64 / 2)

	tests := []struct {
		name       string
		own        SpaceRequirement
		others     []SpaceRequirement
		wantStatus Status
		wantCode   Code
	}{
		{
			name:       "no concurrent requirements",
			own:        SpaceRequirement{Label: RipLabel, Dir: ripDir, Bytes: 1},
			wantStatus: StatusPass,
		},
		{
			name:       "small shared requirements fit",
			own:        SpaceRequirement{Label: RipLabel, Dir: ripDir, Bytes: 1},
			others:     []SpaceRequirement{{Label: BackupLabel, Dir: backupDir, Bytes: 1}},
			wantStatus: StatusPass,
		},
		{
			// Neither half fits on its own, so this also fails without sharing,
			// but it exercises summation across the shared filesystem.
			name:       "shared requirements sum past free space",
			own:        SpaceRequirement{Label: RipLabel, Dir: ripDir, Bytes: half},
			others:     []SpaceRequirement{{Label: BackupLabel, Dir: backupDir, Bytes: half}},
			wantStatus: StatusFail,
			wantCode:   InsufficientSpace,
		},
		{
			name:       "unresolvable concurrent dir is conservatively counted",
			own:        SpaceRequirement{Label: RipLabel, Dir: ripDir, Bytes: 1},
			others:     []SpaceRequirement{{Label: BackupLabel, Dir: filepath.Join(root, "missing"), Bytes: math.MaxUint64 - 1}},
			wantStatus: StatusFail,
			wantCode:   InsufficientSpace,
		},
		{
			name:       "unresolvable own dir fails",
			own:        SpaceRequirement{Label: RipLabel, Dir: filepath.Join(root, "missing"), Bytes: 1},
			others:     []SpaceRequirement{{Label: BackupLabel, Dir: backupDir, Bytes: 1}},
			wantStatus: StatusFail,
			wantCode:   OutputDirInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := FreeSpaceShared(test.own, test.others...)(context.Background())
			if len(got) != 1 {
				t.Fatalf("expected exactly one result, got %d", len(got))
			}
			if got[0].Status != test.wantStatus {
				t.Errorf("expected status %q, got %q (%s)", test.wantStatus, got[0].Status, got[0].Message)
			}
			if test.wantCode != "" && got[0].Code != test.wantCode {
				t.Errorf("expected code %q, got %q", test.wantCode, got[0].Code)
			}
		})
	}
}

func TestSharedSpaceNeeded(t *testing.T) {
	root := t.TempDir()
	own := SpaceRequirement{Label: RipLabel, Dir: root, Bytes: 10}

	tests := []struct {
		name          string
		others        []SpaceRequirement
		wantNeeded    uint64
		wantBreakdown bool
	}{
		{
			name:          "no concurrent requirements",
			wantNeeded:    10,
			wantBreakdown: false,
		},
		{
			name:          "requirement on same filesystem is summed",
			others:        []SpaceRequirement{{Label: BackupLabel, Dir: root, Bytes: 25}},
			wantNeeded:    35,
			wantBreakdown: true,
		},
		{
			name:          "unresolvable requirement is conservatively summed",
			others:        []SpaceRequirement{{Label: BackupLabel, Dir: filepath.Join(root, "missing"), Bytes: 25}},
			wantNeeded:    35,
			wantBreakdown: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			needed, breakdown, err := sharedSpaceNeeded(own, test.others)
			if err != nil {
				t.Fatal(err)
			}
			if needed != test.wantNeeded {
				t.Errorf("expected %d bytes needed, got %d", test.wantNeeded, needed)
			}
			if gotBreakdown := breakdown != ""; gotBreakdown != test.wantBreakdown {
				t.Errorf("expected breakdown present=%t, got %q", test.wantBreakdown, breakdown)
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

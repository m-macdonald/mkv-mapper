package mapper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenameTitles(t *testing.T) {
	tests := []struct {
		name     string
		mappings map[string]string
		wantErr  bool
	}{
		{
			name:     "renames files successfully",
			mappings: map[string]string{"title_01.mkv": "MovieName.mkv"},
		},
		{
			name:     "error oon missing source file",
			mappings: map[string]string{"nonexistent.mkv": "MovieName.mkv"},
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srcDir := t.TempDir()
			destDir := t.TempDir()

			if !test.wantErr {
				for src := range test.mappings {
					path := filepath.Join(srcDir, src)
					if err := os.WriteFile(path, []byte{}, 0644); err != nil {
						t.Fatal(err)
					}
				}
			}

			gotErrs := RenameTitles(srcDir, destDir, test.mappings)

			if test.wantErr && len(gotErrs) == 0 {
				t.Error("expected errors, got none")
			}
			if !test.wantErr && len(gotErrs) > 0 {
				t.Errorf("unexpected errors: %v", gotErrs)
			}
		})
	}
}

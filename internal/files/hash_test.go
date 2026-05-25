package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type mockFileInfo struct {
	name string
	size int64
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() fs.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() any           { return nil }

func TestNewDiscHasher(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		want    discHasher
		wantErr bool
	}{
		{
			name: "bluray",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				os.MkdirAll(filepath.Join(root, "BDMV", "STREAM"), 0o755)
				return root
			},
			want: bluRayHasher{},
		},
		{
			name: "dvd",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				os.MkdirAll(filepath.Join(root, "VIDEO_TS"), 0o755)
				return root
			},
			want: dvdHasher{},
		},
		{
			name: "iso",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				f, _ := os.Create(filepath.Join(root, "disc.iso"))
				f.Close()
				return filepath.Join(root, "disc.iso")
			},
			want: isoHasher{},
		},
		{
			name: "unrecognized",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
		},
		{
			name: "nonexistent",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent", "path")
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := test.setup(t)
			got, err := newDiscHasher(path)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.wantErr {
				return
			}

			switch test.want.(type) {
			case bluRayHasher:
				if _, ok := got.(bluRayHasher); !ok {
					t.Fatalf("expected bluRayHasher, got %T", got)
				}
			case dvdHasher:
				if _, ok := got.(dvdHasher); !ok {
					t.Fatalf("expected dvdHasher, got %T", got)
				}
			case isoHasher:
				if _, ok := got.(isoHasher); !ok {
					t.Fatalf("expected isoHasher, got %T", got)
				}
			}
		})
	}
}

func TestBluRayHasher(t *testing.T) {
	tests := []struct {
		name string
		setup func(t *testing.T) string
		want string
	}{
		{
			name: "single m2ts file",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				dir := filepath.Join(root, "BDMV", "STREAM")
				os.MkdirAll(dir, 0755)
				createFileOfSize(t, filepath.Join(dir, "1.m2ts"), 1000)
				return root
			},
			want: "6A6874884400EACF8E6F8C87507971FD",
		},
		{
            name: "non-m2ts files excluded",
            setup: func(t *testing.T) string {
                root := t.TempDir()
                dir := filepath.Join(root, "BDMV", "STREAM")
                os.MkdirAll(dir, 0755)
                createFileOfSize(t, filepath.Join(dir, "00001.m2ts"), 1000)
                createFileOfSize(t, filepath.Join(dir, "00001.clpi"), 500)
                return root
            },
            want: "6A6874884400EACF8E6F8C87507971FD",
        },
	}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            root := tt.setup(t)
            got, err := bluRayHasher{root: root}.Hash()
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tt.want {
                t.Errorf("want %q, got %q", tt.want, got)
            }
        })
    }
}

func TestDvdHasher(t *testing.T) {
	tests := []struct {
		name string
		setup func(t *testing.T) string
		want string
	}{
		{
			name: "single VOB file",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				dir := filepath.Join(root, "VIDEO_TS")
				os.MkdirAll(dir, 0755)
				createFileOfSize(t, filepath.Join(dir, "1.VOB"), 1000)
				return root
			},
			want: "6A6874884400EACF8E6F8C87507971FD",
		},
	}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            root := tt.setup(t)
            got, err := dvdHasher{root: root}.Hash()
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tt.want {
                t.Errorf("want %q, got %q", tt.want, got)
            }
        })
    }
}

// Only testing the failure paths, as this will require a valid iso to test.
// May add that iso to this repo in the future
func TestISOHasher(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(t *testing.T) string
        wantErr bool
    }{
        {
            name: "nonexistent file",
            setup: func(t *testing.T) string {
                return "/nonexistent/path.iso"
            },
            wantErr: true,
        },
        {
            name: "invalid iso",
            setup: func(t *testing.T) string {
                root := t.TempDir()
                path := filepath.Join(root, "disc.iso")
                os.WriteFile(path, []byte("not an iso"), 0644)
                return path
            },
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := isoHasher{root: tt.setup(t)}.Hash()
            if (err != nil) != tt.wantErr {
                t.Fatalf("unexpected error: %v", err)
            }
            if !tt.wantErr && got == "" {
                t.Error("expected non-empty hash")
            }
        })
    }
}

func createFileOfSize(t *testing.T, path string, size int64) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := f.Truncate(size); err != nil {
		t.Fatal(err)
	}
}

func TestHashSizes(t *testing.T) {
	tests := []struct {
		name  string
		files []os.FileInfo
		want  string
	}{
		{
			name: "single file",
			files: []os.FileInfo{
				mockFileInfo{
					name: "1.m2ts",
					size: 1000,
				},
			},
			want: "6A6874884400EACF8E6F8C87507971FD",
		},
		{
			name: "multiple files",
			files: []os.FileInfo{
				mockFileInfo{
					name: "1.m2ts",
					size: 1000,
				},
				mockFileInfo{
					name: "2.m2ts",
					size: 2000,
				},
			},
			want: "F9E8DE64204B5BD8BD971856043E3D0E",
		},
		{
			name:  "no files",
			files: []os.FileInfo{},
			want:  "D41D8CD98F00B204E9800998ECF8427E",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := hashSizes(test.files)
			if err != nil {
				t.Fatalf("got unexpected error: %v", err)
			}

			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

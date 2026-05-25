package files

import (
	"io/fs"
	"os"
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

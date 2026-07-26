package files

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestResolveDiscRoot(t *testing.T) {
	tests := []struct {
		name    string
		cliRoot string
		want    []string
		wantErr bool
	}{
		{
			name:    "returns cliRoot when given",
			cliRoot: "/path/to/disc/",
			want:    []string{"/path/to/disc/"},
		},
	}

	for _, test := range tests {
		resolver := Resolver{}
		got, err := resolver.ResolveDiscRoot(test.cliRoot)

		if (err != nil) != test.wantErr {
			t.Fatalf("unexpected error: %v", err)
		}

		if !slices.Equal(got, test.want) {
			t.Fatalf("got %q, want %q", got, test.want)
		}
	}
}

func TestFindMountedDisc(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		want    func(base string) []string
		wantErr bool
	}{
		{
			name: "finds valid mount",
			setup: func(t *testing.T) string {
				base := t.TempDir()
				os.MkdirAll(filepath.Join(base, "DISC1", "BDMV", "STREAM"), 0o755)

				return base
			},
			want: func(base string) []string {
				return []string{filepath.Join(base, "DISC1")}
			},
		},
		{
			name: "returns error when dir does not exist",
			setup: func(t *testing.T) string {
				return "DOESNOTEXIST"
			},
			wantErr: true,
		},
		{
			name: "returns empty slice when no mount found",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base := test.setup(t)
			got, err := findMountedDisc(base)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected err: %v", err)
			}

			var want []string
			if test.want != nil {
				want = test.want(base)
			}
			if !slices.Equal(want, got) {
				t.Fatalf("want %q, got %q", want, got)
			}
		})
	}
}

package naming

import (
	"fmt"
	"testing"

	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/signature"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type fakeFilenameGenerator struct {
	name string
	err  error
}

func (f fakeFilenameGenerator) Generate(titleCtx TitleContext) (string, error) {
	return f.name, f.err
}

func TestResolveFilename(t *testing.T) {
	titleContext := TitleContext{
		DiscDbMedia: discdb.Media{
			Title: "Test Media",
			Year:  2026,
			Type:  "Episode",
		},
		DiscDbTitle: discdb.Title{
			Duration:    "Duration",
			DisplaySize: "DisplaySize",
			SourceFile:  "sourceFile.mpls",
			SegmentMap:  "05,7",
			Item: &discdb.Item{
				Title:   "",
				Season:  "",
				Episode: "",
				Type:    "",
			},
		},
		DiscDbDisc: discdb.Disc{},
		MakeMkvTitle: makemkv.Title{
			SourceFilename:   "sourceFilename.mpls",
			OutputFilename:   "outputFilename.mkv",
			SegmentSignature: signature.SegmentSignature(""),
			OutputFileSize:   123456789,
			TitleId:          1,
		},
	}

	tests := []struct {
		name           string
		used           map[string]struct{}
		resolver       filenameGenerator
		wantResolution FilenameResolution
		wantErr        bool
	}{
		{
			name: "successful, no events",
			used: map[string]struct{}{},
			resolver: fakeFilenameGenerator{
				name: titleContext.DiscDbMedia.Title,
				err:  nil,
			},
			wantResolution: FilenameResolution{
				FinalName: fmt.Sprintf("%s.mkv", titleContext.DiscDbMedia.Title),
				Events:    nil,
			},
			wantErr: false,
		},
		{
			name: "fallback to makemkv filename",
			used: map[string]struct{}{},
			resolver: fakeFilenameGenerator{
				name: "",
				err:  fmt.Errorf(""),
			},
			wantResolution: FilenameResolution{
				FinalName: titleContext.MakeMkvTitle.OutputFilename,
				Events: []FilenameEvent{
					{
						Code: WarningNamingFallback,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolution, err := ResolveFilename(test.resolver, titleContext, test.used)
			if (err != nil) != test.wantErr {
				t.Fatalf("err = %v, want %v", err, test.wantErr)
			}

			if diff := cmp.Diff(test.wantResolution, resolution, cmpopts.IgnoreFields(FilenameEvent{}, "Cause", "Message")); diff != "" {
				t.Fatalf("FilnameResolution mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEnsureUniqueFilename(t *testing.T) {
	tests := []struct {
		name                  string
		base                  string
		ext                   string
		titleId               int
		used                  map[string]struct{}
		wantFilename          string
		wantCollisionResolved bool
		wantErr               bool
	}{
		{
			name:                  "no collisions",
			base:                  "movie",
			ext:                   ".mkv",
			titleId:               1,
			used:                  map[string]struct{}{},
			wantFilename:          "movie.mkv",
			wantCollisionResolved: false,
			wantErr:               false,
		},
		{
			name:    "collision",
			base:    "movie",
			ext:     ".mkv",
			titleId: 1,
			used: map[string]struct{}{
				"movie.mkv": {},
			},
			wantFilename:          "movie_t1.mkv",
			wantCollisionResolved: true,
			wantErr:               false,
		},
		{
			name:    "multiple collisions",
			base:    "movie",
			ext:     ".mkv",
			titleId: 1,
			used: map[string]struct{}{
				"movie.mkv":    {},
				"movie_t1.mkv": {},
			},
			wantFilename:          "movie_t1_1.mkv",
			wantCollisionResolved: true,
			wantErr:               false,
		},
		{
			name:    "exhaust maxUniqueFilenameAttempts",
			base:    "movie",
			ext:     ".mkv",
			titleId: 1,
			used: func() map[string]struct{} {
				m := map[string]struct{}{
					"movie.mkv":    {},
					"movie_t1.mkv": {},
				}
				for i := 1; i <= maxUniqueFilenameAttempts; i++ {
					m[fmt.Sprintf("movie_t1_%d.mkv", i)] = struct{}{}
				}
				return m
			}(),
			wantFilename:          "",
			wantCollisionResolved: false,
			wantErr:               true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, collisionResolved, err := ensureUniqueFilename(test.base, test.ext, test.titleId, test.used)
			if (err != nil) != test.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.wantFilename {
				t.Fatalf("got filename %q, want %q", got, test.wantFilename)
			}
			if collisionResolved != test.wantCollisionResolved {
				t.Fatalf("got collision resolved %v, want %v", collisionResolved, test.wantCollisionResolved)
			}
		})
	}
}

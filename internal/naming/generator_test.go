package naming

import (
	"testing"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/signature"

	"github.com/google/go-cmp/cmp"
)

func TestFilenameGeneratorGenerate(t *testing.T) {
	tests := []struct {
		name        string
		titleCtx    TitleContext
		templateCfg config.TemplateConfig
		want        string
		wantErr     bool
	}{
		{
			name:     "uses unknown when title has no item",
			titleCtx: TitleContext{
				DiscDbTitle: discdb.Title{},
			},
			templateCfg: config.TemplateConfig{
				Unknown: "unknown",	
			},
			want:    "unknown",
			wantErr:  false,
		},
		{
			name: "uses override when it exists",
			titleCtx: TitleContext{
				
			},
			templateCfg: config.TemplateConfig{
				Override: "override",
			},
			want: "override",
			wantErr: false,
		},
		{
			name: "uses mapped template when item exists",
			templateCfg: config.TemplateConfig{
				Episode: "episode",
				Unknown: "unknown",
			},
			titleCtx: TitleContext{
				DiscDbTitle: discdb.Title{
					Item: &discdb.Item{
						Type: "Episode",
					},
				},
			},
			want: "episode",
			wantErr: false,
		},
		{
			name: "returns error when selected template fails",
			templateCfg: config.TemplateConfig{
				Unknown: "{{ .DoesNotExist }}",
			},
			titleCtx: TitleContext{
				DiscDbTitle: discdb.Title{},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filenameGen, err := NewFilenameGenerator(test.templateCfg)
			if err != nil {
				t.Fatalf("constructing FilenameGenerator: %v", err)
			}
			got, err := filenameGen.Generate(test.titleCtx)
			if (err != nil) != test.wantErr {
				t.Fatalf("err = %v, want %v", err, test.wantErr)
			}

			if test.want != got {
				t.Fatalf("want %q, got %q", test.want, got)
			}
		}) 
	}
}

func TestBuildTemplateVars(t *testing.T) {
	titleCtx := titleContext()
	want := TemplateVars{
		Media: TemplateMedia{
			Title: titleCtx.DiscDbMedia.Title,
			Year:  titleCtx.DiscDbMedia.Year,
			Type:  titleCtx.DiscDbMedia.Type,
		},
		Disc: TemplateDisc{
			ContentHash: titleCtx.DiscDbDisc.ContentHash,
			Format:      titleCtx.DiscDbDisc.Format,
			Name:        titleCtx.DiscDbDisc.Name,
			Slug:        titleCtx.DiscDbDisc.Slug,
		},
		Title: TemplateTitle{
			DisplaySize: titleCtx.DiscDbTitle.DisplaySize,
			Duration:    titleCtx.DiscDbTitle.Duration,
			SegmentMap:  titleCtx.DiscDbTitle.SegmentMap,
			Size:        titleCtx.DiscDbTitle.Size,
			SourceFile:  titleCtx.DiscDbTitle.SourceFile,
		},
		MakeMkv: TemplateMakeMkvTitle{
			TitleId:          titleCtx.MakeMkvTitle.TitleId,
			OutputFilename:   titleCtx.MakeMkvTitle.OutputFilename,
			SourceFilename:   titleCtx.MakeMkvTitle.SourceFilename,
			SegmentSignature: string(titleCtx.MakeMkvTitle.SegmentSignature),
			OutputFileSize:   titleCtx.MakeMkvTitle.OutputFileSize,
		},
		Season:       titleCtx.DiscDbTitle.Item.Season,
		Episode:      titleCtx.DiscDbTitle.Item.Episode,
		EpisodeTitle: titleCtx.DiscDbTitle.Item.Title,
		MovieTitle:   titleCtx.DiscDbTitle.Item.Title,
	}

	got := buildTemplateVars(titleCtx)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("template variables not equal (-want +got)\n%s", diff)
	}
}

func TestPad(t *testing.T) {
	tests := []struct {
		name   string
		padCnt int
		val    string
		want   string
	}{
		{
			name:   "pads string",
			padCnt: 6,
			val:    "test",
			want:   "00test",
		},
		{
			name:   "does not pad",
			padCnt: 3,
			val:    "5",
			want:   "005",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := pad(test.padCnt, test.val)

			if got != test.want {
				t.Fatalf("want %q, got %q", test.want, got)
			}
		})
	}
}

func TestDflt(t *testing.T) {
	tests := []struct {
		name  string
		deflt string
		val   string
		want  string
	}{
		{
			name:  "returns value",
			deflt: "default",
			val:   " value ",
			want:  " value ",
		},
		{
			name:  "returns default",
			deflt: "default",
			val:   "",
			want:  "default",
		},
		{
			name:  "value is empty tring, returns default",
			deflt: "default",
			val:   "   ",
			want:  "default",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := dflt(test.deflt, test.val)

			if got != test.want {
				t.Fatalf("want %q, got %q", test.want, got)
			}
		})
	}
}

func titleContext() TitleContext {
	return TitleContext{
		DiscDbMedia: discdb.Media{
			Title: "MediaTitle",
			Year:  0,
			Type:  "MediaType",
		},
		DiscDbDisc: discdb.Disc{
			ContentHash: "DiscContentHash",
			Format:      "DiscFormat",
			Name:        "DiscName",
			Slug:        "DiscSlug",
		},
		DiscDbTitle: discdb.Title{
			DisplaySize: "TitleDisplaySize",
			Duration:    "TitleDuration",
			SegmentMap:  "TitleSegmentMap",
			Size:        1,
			SourceFile:  "TitleSourceFile",
			Item: &discdb.Item{
				Title:   "ItemTitle",
				Season:  "ItemSeason",
				Episode: "ItemEpisode",
				Type:    "ItemType",
			},
		},
		MakeMkvTitle: makemkv.Title{
			TitleId:          2,
			OutputFilename:   "MakeMkvOutputFilename",
			SourceFilename:   "MakeMkvSourceFilename",
			SegmentSignature: signature.SegmentSignature("MakeMkvSegmentSignature"),
			OutputFileSize:   3,
		},
	}
}

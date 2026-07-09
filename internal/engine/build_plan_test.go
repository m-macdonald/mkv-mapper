package engine

import (
	"testing"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/mapper"
	th "m-macdonald/mkv-mapper/internal/mkvmappertest"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/naming"
)

func TestBuildPlan(t *testing.T) {
	validConfig := config.TemplateConfig{
		Movie:   "{{.Media.Title}}",
		Episode: "{{.Media.Title}}",
		Extra:   "{{.Media.Title}}",
		Unknown: "{{.Media.Title}}",
	}
	tests := []struct {
		name       string
		config     config.TemplateConfig
		discRecord discdb.DiscRecord
		titles     []makemkv.Title
		wantTitles int
		wantErr    bool
	}{
		{
			name:   "success",
			config: validConfig,
			discRecord: th.NewDiscRecord(
				th.WithMedia(discdb.Media{Title: "Test", Year: 2024, Type: "Movie"}),
				th.WithTitles(
					th.NewDiscTitle(th.WithSegmentMap("1,2,3"), th.WithItem(th.NewDiscItem())),
				),
			),
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSignature("1,2,3")),
			},
			wantTitles: 1,
		},
		{
			name:   "no mappings",
			config: validConfig,
			discRecord: th.NewDiscRecord(
				th.WithMedia(discdb.Media{Title: "Test", Year: 2024, Type: "Movie"}),
			),
			titles:     []makemkv.Title{},
			wantTitles: 0,
		},
		{
			name: "invalid template fails at construction",
			config: config.TemplateConfig{
				Movie:   "{{.Missing",
				Episode: "{{.Missing",
				Extra:   "{{.Missing",
				Unknown: "{{.Missing",
			},
			discRecord: th.NewDiscRecord(
				th.WithTitles(
					th.NewDiscTitle(th.WithSegmentMap("1,2,3"), th.WithItem(th.NewDiscItem())),
				),
			),
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSignature("1,2,3")),
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := buildPlan("/disc", "/output", test.config, test.discRecord, test.titles)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.wantErr {
				return
			}
			if len(plan.Titles) != test.wantTitles {
				t.Errorf("expected %d titles, got %d", test.wantTitles, len(plan.Titles))
			}
			if plan.MediaInfo.Title != test.discRecord.Media.Title {
				t.Errorf("expected MediaTitle %q, got %q", test.discRecord.Media.Title, plan.MediaInfo.Title)
			}
			if plan.MediaInfo.Year != test.discRecord.Media.Year {
				t.Errorf("expected MediaYear %d, got %d", test.discRecord.Media.Year, plan.MediaInfo.Year)
			}
			if plan.DiscRoot != "/disc" {
				t.Errorf("expected DiscRoot %q, got %q", "/disc", plan.DiscRoot)
			}
			if plan.OutputDir != "/output" {
				t.Errorf("expected OutputDir %q, got %q", "/output", plan.OutputDir)
			}
			for _, title := range plan.Titles {
				if title.FinalName == "" {
					t.Error("expected non-empty FinalName")
				}
			}
		})
	}
}

func TestResolveFilenames(t *testing.T) {
	validConfig := config.TemplateConfig{
		Movie:   "{{.Media.Title}}",
		Episode: "{{.Media.Title}}",
		Extra:   "{{.Media.Title}}",
		Unknown: "{{.Media.Title}}",
	}
	invalidConfig := config.TemplateConfig{
		Movie:   "{{.Missing}}",
		Episode: "{{.Missing}}",
		Extra:   "{{.Missing}}",
		Unknown: "{{.Missing}}",
	}
	discRecord := th.NewDiscRecord(
		th.WithMedia(discdb.Media{Title: "Test"}),
		th.WithTitles(
			th.NewDiscTitle(th.WithSegmentMap("1,2,3"), th.WithItem(th.NewDiscItem())),
			th.NewDiscTitle(th.WithSegmentMap("4,5,6"), th.WithItem(th.NewDiscItem())),
			th.NewDiscTitle(th.WithSegmentMap("7,8,9")),
		),
	)
	tests := []struct {
		name           string
		config         config.TemplateConfig
		mappings       []mapper.TitleMapping
		wantTitles     int
		wantFinalNames []string
		wantWarnings   []model.WarningCode
		wantErr        bool
	}{
		{
			name:   "success",
			config: validConfig,
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(th.WithSignature("1,2,3")),
					DiscDbTitle: th.NewDiscTitle(
						th.WithSegmentMap("1,2,3"),
						th.WithItem(th.NewDiscItem()),
					),
				},
			},
			wantTitles:     1,
			wantFinalNames: []string{"Test.mkv"},
			wantWarnings:   []model.WarningCode{},
		},
		{
			name:   "no metadata warning",
			config: validConfig,
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(th.WithSignature("1,2,3")),
					DiscDbTitle:  th.NewDiscTitle(th.WithSegmentMap("1,2,3")),
				},
			},
			wantTitles:   1,
			wantWarnings: []model.WarningCode{model.WarningNoMetadata},
		},
		{
			name:   "fallback warning",
			config: invalidConfig,
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(th.WithSignature("1,2,3"), th.WithOutputFilename("Fallback.mkv")),
					DiscDbTitle:  th.NewDiscTitle(th.WithSegmentMap("1,2,3"), th.WithItem(th.NewDiscItem())),
				},
			},
			wantTitles:     1,
			wantFinalNames: []string{"Fallback.mkv"},
			wantWarnings:   []model.WarningCode{model.WarningCode(naming.WarningNamingFallback)},
		},
		{
			name:   "no metadata and fallback",
			config: invalidConfig,
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(
						th.WithSignature("1,2,3"),
						th.WithOutputFilename("Fallback.mkv")),
					DiscDbTitle: th.NewDiscTitle(th.WithSegmentMap("1,2,3")),
				},
			},
			wantTitles:     1,
			wantFinalNames: []string{"Fallback.mkv"},
			wantWarnings: []model.WarningCode{
				model.WarningNoMetadata,
				model.WarningCode(naming.WarningNamingFallback),
			},
		},
		{
			name:   "collision warning",
			config: validConfig,
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(th.WithSignature("1,2,3")),
					DiscDbTitle: th.NewDiscTitle(
						th.WithSegmentMap("1,2,3"),
						th.WithItem(th.NewDiscItem())),
				},
				{
					MakeMkvTitle: th.NewMakeMkvTitle(
						th.WithSignature("4,5,6"),
						th.WithTitleId(1)),
					DiscDbTitle: th.NewDiscTitle(
						th.WithSegmentMap("4,5,6"),
						th.WithItem(th.NewDiscItem())),
				},
			},
			wantTitles: 2,
			wantFinalNames: []string{
				"Test.mkv",
				"Test_t1.mkv",
			},
			wantWarnings: []model.WarningCode{
				model.WarningCode(naming.WarningFilenameSuffixed),
			},
		},
		{
			name:         "empty mappings",
			config:       validConfig,
			mappings:     []mapper.TitleMapping{},
			wantTitles:   0,
			wantWarnings: []model.WarningCode{},
		},
		{
			name: "invalid template fails at construction",
			config: config.TemplateConfig{
				Movie:   "{{.Missing",
				Episode: "{{.Missing",
				Extra:   "{{.Missing",
				Unknown: "{{.Missing",
			},
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(th.WithSignature("1,2,3")),
					DiscDbTitle:  th.NewDiscTitle(th.WithSegmentMap("1,2,3"), th.WithItem(th.NewDiscItem())),
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := &model.Plan{
				PlanBase: model.PlanBase{
					Titles: make([]model.TitlePlan, 0),
				},
			}
			err := resolveFilenames(test.config, test.mappings, discRecord, plan)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.wantErr {
				return
			}
			if len(plan.Titles) != test.wantTitles {
				t.Errorf("expected %d titles, got %d", test.wantTitles, len(plan.Titles))
			}
			for _, title := range plan.Titles {
				if title.FinalName == "" {
					t.Error("expected non-empty FinalName")
				}
			}
			if test.wantFinalNames != nil {
				if len(test.wantFinalNames) != len(plan.Titles) {
					t.Errorf("expected %d final names, got %d", len(test.wantFinalNames), len(plan.Titles))
				} else {
					for i, wantName := range test.wantFinalNames {
						if plan.Titles[i].FinalName != wantName {
							t.Errorf("title %d: expected FinalName %q, got %q", i, wantName, plan.Titles[i].FinalName)
						}
					}
				}
			}
			if len(plan.BuildReport.Warnings) != len(test.wantWarnings) {
				t.Errorf("expected %d warnings, got %d", len(test.wantWarnings), len(plan.BuildReport.Warnings))
				return
			}
			for i, w := range plan.BuildReport.Warnings {
				if w.Code != test.wantWarnings[i] {
					t.Errorf("warning %d: expected code %q, got %q", i, test.wantWarnings[i], w.Code)
				}
			}
		})
	}
}

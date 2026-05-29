package planner

import (
	"testing"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/mapper"
	th "m-macdonald/mkv-mapper/internal/mkvmappertest"
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
				th.NewMakeMkvTitle(th.WithSegments("1,2,3")),
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
				th.NewMakeMkvTitle(th.WithSegments("1,2,3")),
			},
			wantErr: true,
		},
		{
			name:   "maptitles failure",
			config: validConfig,
			discRecord: th.NewDiscRecord(
				th.WithTitles(
					th.NewDiscTitle(th.WithSegmentMap("   "), th.WithItem(th.NewDiscItem())),
				),
			),
			titles: []makemkv.Title{
				th.NewMakeMkvTitle(th.WithSegments("1,2,3")),
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan, err := BuildPlan("/disc", "/output", test.config, test.discRecord, test.titles)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.wantErr {
				return
			}
			if len(plan.DiscPlan.Titles) != test.wantTitles {
				t.Errorf("expected %d titles, got %d", test.wantTitles, len(plan.DiscPlan.Titles))
			}
			if plan.DiscPlan.MediaTitle != test.discRecord.Media.Title {
				t.Errorf("expected MediaTitle %q, got %q", test.discRecord.Media.Title, plan.DiscPlan.MediaTitle)
			}
			if plan.DiscPlan.MediaYear != test.discRecord.Media.Year {
				t.Errorf("expected MediaYear %d, got %d", test.discRecord.Media.Year, plan.DiscPlan.MediaYear)
			}
			if plan.DiscPlan.DiscRoot != "/disc" {
				t.Errorf("expected DiscRoot %q, got %q", "/disc", plan.DiscPlan.DiscRoot)
			}
			if plan.DiscPlan.OutputDir != "/output" {
				t.Errorf("expected OutputDir %q, got %q", "/output", plan.DiscPlan.OutputDir)
			}
			for _, title := range plan.DiscPlan.Titles {
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
		wantWarnings   []WarningCode
		wantErr        bool
	}{
		{
			name:   "success",
			config: validConfig,
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(th.WithSegments("1,2,3")),
					DiscDbTitle: th.NewDiscTitle(
						th.WithSegmentMap("1,2,3"),
						th.WithItem(th.NewDiscItem()),
					),
				},
			},
			wantTitles:     1,
			wantFinalNames: []string{"Test.mkv"},
			wantWarnings:   []WarningCode{},
		},
		{
			name:   "no metadata warning",
			config: validConfig,
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(th.WithSegments("1,2,3")),
					DiscDbTitle:  th.NewDiscTitle(th.WithSegmentMap("1,2,3")),
				},
			},
			wantTitles:   1,
			wantWarnings: []WarningCode{WarningNoMetadata},
		},
		{
			name:   "fallback warning",
			config: invalidConfig,
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(th.WithSegments("1,2,3"), th.WithOutputFilename("Fallback.mkv")),
					DiscDbTitle:  th.NewDiscTitle(th.WithSegmentMap("1,2,3"), th.WithItem(th.NewDiscItem())),
				},
			},
			wantTitles:     1,
			wantFinalNames: []string{"Fallback.mkv"},
			wantWarnings:   []WarningCode{WarningCode(naming.WarningNamingFallback)},
		},
		{
			name:   "no metadata and fallback",
			config: invalidConfig,
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(
						th.WithSegments("1,2,3"),
						th.WithOutputFilename("Fallback.mkv")),
					DiscDbTitle: th.NewDiscTitle(th.WithSegmentMap("1,2,3")),
				},
			},
			wantTitles:     1,
			wantFinalNames: []string{"Fallback.mkv"},
			wantWarnings: []WarningCode{
				WarningNoMetadata,
				WarningCode(naming.WarningNamingFallback),
			},
		},
		{
			name:   "collision warning",
			config: validConfig,
			mappings: []mapper.TitleMapping{
				{
					MakeMkvTitle: th.NewMakeMkvTitle(th.WithSegments("1,2,3")),
					DiscDbTitle: th.NewDiscTitle(
						th.WithSegmentMap("1,2,3"),
						th.WithItem(th.NewDiscItem())),
				},
				{
					MakeMkvTitle: th.NewMakeMkvTitle(
						th.WithSegments("4,5,6"),
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
			wantWarnings: []WarningCode{
				WarningCode(naming.WarningFilenameSuffixed),
			},
		},
		{
			name:         "empty mappings",
			config:       validConfig,
			mappings:     []mapper.TitleMapping{},
			wantTitles:   0,
			wantWarnings: []WarningCode{},
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
					MakeMkvTitle: th.NewMakeMkvTitle(th.WithSegments("1,2,3")),
					DiscDbTitle:  th.NewDiscTitle(th.WithSegmentMap("1,2,3"), th.WithItem(th.NewDiscItem())),
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := &DiscPlan{
				Titles: make([]TitlePlan, 0),
			}
			report := &BuildReport{
				Warnings: make([]PlanWarning, 0),
			}
			err := resolveFilenames(test.config, test.mappings, discRecord, plan, report)
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
			if len(report.Warnings) != len(test.wantWarnings) {
				t.Errorf("expected %d warnings, got %d", len(test.wantWarnings), len(report.Warnings))
				return
			}
			for i, w := range report.Warnings {
				if w.Code != test.wantWarnings[i] {
					t.Errorf("warning %d: expected code %q, got %q", i, test.wantWarnings[i], w.Code)
				}
			}
		})
	}
}

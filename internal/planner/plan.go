package planner

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/naming"
	"m-macdonald/mkv-mapper/internal/signature"
)

type DiscPlan struct {
	DiscHash  string
	DiscRoot  string
	OutputDir string
	Titles    []TitlePlan
}

type TitlePlan struct {
	TitleId           int
	SourcePlaylist    string
	SegmentSignature  signature.SegmentSignature
	MakeMkvOutputFile string
	FinalName         string
	EstimatedSize     uint
}

func BuildPlan(
	discRoot string,
	outputDir string,
	templateConfig config.TemplateConfig,
	disc *discdb.Disc,
	titles map[signature.SegmentSignature]makemkv.Title,
) (*DiscPlan, *BuildReport, error) {
	mappings, err := mapper.MapTitles(disc, titles)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to map MakeMkv titles to DiscDB titles %w", err)
	}

	filenameGen, err := naming.NewGenerator(templateConfig)
	if err != nil {
		return nil, nil, err
	}

	plan := &DiscPlan{
		DiscRoot:  discRoot,
		OutputDir: outputDir,
		Titles:    make([]TitlePlan, 0, len(mappings)),
	}
	report := &BuildReport{
		Warnings: make([]PlanWarning, 0),
	}

	usedNames := make(map[string]struct{}, len(mappings))
	for _, mapping := range mappings {
		filenameResolution, err := resolveFilename(filenameGen, disc, mapping, usedNames)
		if err != nil {
			return nil, report, fmt.Errorf(
				"failed to resolve filename for makemkv title %d (%s): %w",
				mapping.MakeMkvTitle.TitleId,
				mapping.MakeMkvTitle.OutputFileName,
				err)
		}

		fmt.Printf("Resolved filename: %s\n", filenameResolution.FinalName)
		for _, event := range filenameResolution.Events {
			report.Warnings = append(report.Warnings, PlanWarning{
				TitleId: mapping.MakeMkvTitle.TitleId,
				Code:    event.Code,
				Message: event.Message,
				Cause:   event.Cause,
			})
		}

		plan.Titles = append(plan.Titles, TitlePlan{
			SourcePlaylist:    mapping.MakeMkvTitle.SourceFileName,
			MakeMkvOutputFile: mapping.MakeMkvTitle.OutputFileName,
			FinalName:         filenameResolution.FinalName,
			EstimatedSize:     mapping.MakeMkvTitle.OutputFileSize,
		})
	}

	return plan, report, nil
}

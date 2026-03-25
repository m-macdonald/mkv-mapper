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
	EstimatedSize     uint64
}

func BuildPlan(
	discRoot string,
	outputDir string,
	templateConfig config.TemplateConfig,
	discRecord *discdb.DiscRecord,
	titles map[signature.SegmentSignature]makemkv.Title,
) (*DiscPlan, *BuildReport, error) {
	mappings, err := mapper.MapTitles(discRecord, titles)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to map MakeMkv titles to DiscDB titles %w", err)
	}

	filenameGen, err := naming.NewFilenameGenerator(templateConfig)
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

	// Track used filenames so that we can resolve conflicts
	usedNames := make(map[string]struct{}, len(mappings))
	for _, mapping := range mappings {
		titleContext := naming.TitleContext{
			DiscDbMedia: discRecord.Media,
			DiscDbTitle: mapping.DiscDbTitle,
			DiscDbDisc: discRecord.Disc,
			MakeMkvTitle: mapping.MakeMkvTitle,
		}
		filenameResolution, err := naming.ResolveFilename(filenameGen, titleContext, usedNames)
		if err != nil {
			return nil, report, fmt.Errorf(
				"failed to resolve filename for makemkv title %d (%s): %w",
				mapping.MakeMkvTitle.TitleId,
				mapping.MakeMkvTitle.OutputFilename,
				err)
		}

		for _, event := range filenameResolution.Events {
			report.Warnings = append(report.Warnings, PlanWarning{
				TitleId: mapping.MakeMkvTitle.TitleId,
				// TODO: Translate this better
				Code:    WarningCode(event.Code),
				Message: event.Message,
				Cause:   event.Cause,
			})
		}

		plan.Titles = append(plan.Titles, TitlePlan{
			SourcePlaylist:    mapping.MakeMkvTitle.SourceFilename,
			MakeMkvOutputFile: mapping.MakeMkvTitle.OutputFilename,
			FinalName:         filenameResolution.FinalName,
			EstimatedSize:     mapping.MakeMkvTitle.OutputFileSize,
		})
	}

	return plan, report, nil
}

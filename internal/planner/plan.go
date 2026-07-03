package planner

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/naming"
	"m-macdonald/mkv-mapper/internal/signature"
)

type Plan struct {
	PlanBase
	BuildReport BuildReport
}

type TitlePlan struct {
	TitleId           lines.TitleId
	SourcePlaylist    string
	SegmentSignature  signature.SegmentSignature
	MakeMkvOutputFile string
	FinalName         string
	EstimatedSize     uint64
	IsMatched         bool // Indicates if this title had a matching Item definition in TheDiscDb
}

func BuildPlan(
	discRoot string,
	outputDir string,
	templateConfig config.TemplateConfig,
	discRecord discdb.DiscRecord,
	titles []makemkv.Title,
) (Plan, error) {
	mappings, err := mapper.MapTitles(discRecord, titles)
	if err != nil {
		return Plan{}, fmt.Errorf("failed to map MakeMkv titles to DiscDB titles %w", err)
	}

	plan := Plan{
		PlanBase: PlanBase{
			MediaInfo: MediaInfo{
				Title: discRecord.Media.Title,
				Year:  discRecord.Media.Year,
			},
			Disc: Disc{
				Format: discRecord.Disc.Format,
				Hash:   discRecord.Disc.ContentHash,
			},
			DiscRoot:  discRoot,
			OutputDir: outputDir,
			Titles:    make([]TitlePlan, 0, len(mappings)),
		},
		BuildReport: BuildReport{
			Warnings: make([]PlanWarning, 0),
		},
	}

	err = resolveFilenames(templateConfig, mappings, discRecord, &plan)
	if err != nil {
		return Plan{}, err
	}

	return plan, nil
}

func resolveFilenames(
	templateConfig config.TemplateConfig,
	mappings []mapper.TitleMapping,
	discRecord discdb.DiscRecord,
	plan *Plan,
) error {
	filenameGen, err := naming.NewFilenameGenerator(templateConfig)
	if err != nil {
		return err
	}
	// Track used filenames so that we can resolve conflicts
	usedNames := make(map[string]struct{}, len(mappings))
	for _, mapping := range mappings {
		if mapping.DiscDbTitle.Item == nil {
			plan.BuildReport.Warnings = append(plan.BuildReport.Warnings, PlanWarning{
				TitleId: mapping.MakeMkvTitle.TitleId,
				Code:    WarningNoMetadata,
				Message: "Title has no DiscDB metadata",
			})
		}
		titleContext := naming.TitleContext{
			DiscDbMedia:  discRecord.Media,
			DiscDbTitle:  mapping.DiscDbTitle,
			DiscDbDisc:   discRecord.Disc,
			MakeMkvTitle: mapping.MakeMkvTitle,
		}
		filenameResolution, err := naming.ResolveFilename(filenameGen, titleContext, usedNames)
		if err != nil {
			return fmt.Errorf(
				"failed to resolve filename for makemkv title %d (%s): %w",
				mapping.MakeMkvTitle.TitleId,
				mapping.MakeMkvTitle.OutputFilename,
				err)
		}

		for _, event := range filenameResolution.Events {
			plan.BuildReport.Warnings = append(plan.BuildReport.Warnings, PlanWarning{
				TitleId: mapping.MakeMkvTitle.TitleId,
				// TODO: Translate this better
				Code:    WarningCode(event.Code),
				Message: event.Message,
				Cause:   event.Cause,
			})
		}

		plan.Titles = append(plan.Titles, TitlePlan{
			TitleId:           mapping.MakeMkvTitle.TitleId,
			SourcePlaylist:    mapping.MakeMkvTitle.SourceFilename,
			MakeMkvOutputFile: mapping.MakeMkvTitle.OutputFilename,
			FinalName:         filenameResolution.FinalName,
			EstimatedSize:     mapping.MakeMkvTitle.OutputFileSize,
			IsMatched:         mapping.DiscDbTitle.Item != nil,
		})
	}

	return nil
}

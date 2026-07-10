package engine

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/naming"
)

func buildPlan(
	discRoot string,
	cfg BuildPlanConfig,
	discRecord discdb.DiscRecord,
	discInfo makemkv.DiscInfo,
) (model.Plan, error) {
	mappings := mapper.MapTitles(discRecord, discInfo.Titles)

	plan := model.Plan{
		PlanBase: model.PlanBase{
			MediaInfo: model.MediaInfo{
				Title: discRecord.Media.Title,
				Year:  discRecord.Media.Year,
			},
			Disc: model.Disc{
				Label:  discInfo.Label,
				Format: discRecord.Disc.Format,
				Hash:   discRecord.Disc.ContentHash,
			},
			DiscRoot:   discRoot,
			OutputDir:  cfg.OutputDir,
			Backup:     cfg.Rip.Backup,
			BackupDir:  cfg.Rip.BackupDir,
			KeepBackup: cfg.Rip.KeepBackup,
		},
	}

	err := resolveFilenames(cfg.Templates, mappings, discRecord, &plan)
	if err != nil {
		return model.Plan{}, err
	}

	return plan, nil
}

func resolveFilenames(
	templateConfig config.TemplateConfig,
	mappings []mapper.TitleMapping,
	discRecord discdb.DiscRecord,
	plan *model.Plan,
) error {
	filenameGen, err := naming.NewFilenameGenerator(templateConfig)
	if err != nil {
		return err
	}
	// Track used filenames so that we can resolve conflicts
	usedNames := make(map[string]struct{}, len(mappings))
	for _, mapping := range mappings {
		if mapping.DiscDbTitle.Item == nil {
			plan.BuildReport.Warnings = append(plan.BuildReport.Warnings, model.PlanWarning{
				TitleId: mapping.MakeMkvTitle.TitleId,
				Code:    model.WarningNoMetadata,
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
			plan.BuildReport.Warnings = append(plan.BuildReport.Warnings, model.PlanWarning{
				TitleId: mapping.MakeMkvTitle.TitleId,
				// TODO: Translate this better
				Code:    model.WarningCode(event.Code),
				Message: event.Message,
				Cause:   event.Cause,
			})
		}

		plan.Titles = append(plan.Titles, model.TitlePlan{
			TitleId:           mapping.MakeMkvTitle.TitleId,
			SourcePlaylist:    mapping.MakeMkvTitle.SourceFilename,
			MakeMkvOutputFile: mapping.MakeMkvTitle.OutputFilename,
			FinalName:         filenameResolution.FinalName,
			EstimatedSize:     mapping.MakeMkvTitle.OutputFileSize,
			Duration:          mapping.DiscDbTitle.Duration,
			IsMatched:         mapping.DiscDbTitle.Item != nil,
		})
	}

	return nil
}

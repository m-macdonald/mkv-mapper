package engine

import (
	"context"
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/naming"
)

type BuildPlanConfig struct {
	DiscRoot  string
	OutputDir string
	Templates config.TemplateConfig
	Rip       config.RipConfig
	Backup    config.BackupConfig
}

func (e *Engine) BuildPlan(
	ctx context.Context,
	cfg BuildPlanConfig,
) (model.Plan, error) {
	discRoot, err := e.discResolver.ResolveDiscRoot(cfg.DiscRoot)
	if err != nil {
		return model.Plan{}, err
	}

	hash, err := files.Hash(discRoot)
	if err != nil {
		return model.Plan{}, fmt.Errorf("unable to hash disc %w", err)
	}

	disc, err := e.discdb.LookupDisc(ctx, hash)
	if err != nil {
		return model.Plan{}, fmt.Errorf("failed to retrieve disc definitions from TheDiscDB %w", err)
	}

	discInfo, err := e.makemkv.ReadDisc(ctx, discRoot)
	if err != nil {
		return model.Plan{}, fmt.Errorf("unable to read disc titles using MakeMkv %w", err)
	}

	mappings := mapper.MapTitles(disc, discInfo.Titles)

	plan := model.Plan{
		PlanBase: model.PlanBase{
			MediaInfo: model.MediaInfo{
				Title: disc.Media.Title,
				Year:  disc.Media.Year,
			},
			Disc: model.Disc{
				Label:  discInfo.Label,
				Format: disc.Disc.Format,
				Hash:   disc.Disc.ContentHash,
			},
			DiscRoot:   discRoot,
			OutputDir:  cfg.OutputDir,
		},
	}

	err = resolveFilenames(cfg.Templates, mappings, disc, &plan)
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
			SegmentSignature:  mapping.DiscDbTitle.Signature,
		})
	}

	return nil
}

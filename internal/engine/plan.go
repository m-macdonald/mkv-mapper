package engine

import (
	"context"
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv"
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

type planContent struct {
	titles      []model.TitlePlan
	buildReport model.BuildReport
}

func (e *Engine) BuildPlan(ctx context.Context, cfg BuildPlanConfig) (model.Plan, error) {
	identity, discInfo, err := e.ScanDisc(ctx, cfg.DiscRoot)
	if err != nil {
		return model.Plan{}, err
	}
	return e.CompletePlan(ctx, identity, discInfo, cfg)
}

func (e *Engine) BuildBackupPlan(ctx context.Context, cfg BuildBackupPlanConfig) (model.BackupPlan, error) {
	identity, discInfo, err := e.ScanDisc(ctx, cfg.DiscRoot)
	if err != nil {
		return model.BackupPlan{}, err
	}
	return e.CompleteBackupPlan(identity, discInfo, cfg), nil
}

func (e *Engine) CompletePlan(
	ctx context.Context,
	identity model.DiscIdentity,
	discInfo makemkv.DiscInfo,
	cfg BuildPlanConfig,
) (model.Plan, error) {
	hash, err := files.Hash(identity.DiscRoot)
	if err != nil {
		return model.Plan{}, fmt.Errorf("unable to hash disc: %w", err)
	}

	disc, err := e.discdb.LookupDisc(ctx, hash)
	if err != nil {
		return model.Plan{}, fmt.Errorf("failed to retrieve disc definitions from TheDiscDB: %w", err)
	}

	mappings := mapper.MapTitles(disc, discInfo.Titles)

	planContent, err := resolveFilenames(cfg.Templates, mappings, disc)
	if err != nil {
		return model.Plan{}, err
	}

	plan := model.Plan{
		PlanBase: model.PlanBase{
			DiscIdentity: identity,
			MediaInfo: model.MediaInfo{
				Title: disc.Media.Title,
				Year:  disc.Media.Year,
			},
			DiscInfo: model.DiscInfo{
				Format: disc.Disc.Format,
				Hash:   hash,
			},
			OutputDir: cfg.OutputDir,
			Titles:    planContent.titles,
		},
		BuildReport: planContent.buildReport,
	}

	return plan, nil
}

type BuildBackupPlanConfig struct {
	DiscRoot   string
	OutputDir  string
	KeepBackup bool
}

func (e *Engine) CompleteBackupPlan(
	identity model.DiscIdentity,
	discInfo makemkv.DiscInfo,
	cfg BuildBackupPlanConfig,
) model.BackupPlan {
	titles := make([]model.BackupTitle, 0, len(discInfo.Titles))
	for _, title := range discInfo.Titles {
		titles = append(titles, model.BackupTitle{
			TitleId:       title.TitleId,
			EstimatedSize: title.OutputFileSize,
		})
	}

	return model.BackupPlan{
		DiscIdentity: identity,
		OutputDir:    cfg.OutputDir,
		KeepBackup:   cfg.KeepBackup,
		Titles:       titles,
	}
}

func resolveFilenames(
	templateConfig config.TemplateConfig,
	mappings []mapper.TitleMapping,
	discRecord discdb.DiscRecord,
) (planContent, error) {
	filenameGen, err := naming.NewFilenameGenerator(templateConfig)
	if err != nil {
		return planContent{}, err
	}

	var content planContent
	// Track used filenames so that we can resolve conflicts
	usedNames := make(map[string]struct{}, len(mappings))
	for _, mapping := range mappings {
		if mapping.DiscDbTitle.Item == nil {
			content.buildReport.Warnings = append(content.buildReport.Warnings, model.PlanWarning{
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
			return planContent{}, fmt.Errorf(
				"failed to resolve filename for makemkv title %d (%s): %w",
				mapping.MakeMkvTitle.TitleId,
				mapping.MakeMkvTitle.OutputFilename,
				err)
		}

		for _, event := range filenameResolution.Events {
			content.buildReport.Warnings = append(content.buildReport.Warnings, model.PlanWarning{
				TitleId: mapping.MakeMkvTitle.TitleId,
				// TODO: Translate this better
				Code:    model.WarningCode(event.Code),
				Message: event.Message,
				Cause:   event.Cause,
			})
		}

		content.titles = append(content.titles, model.TitlePlan{
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

	return content, nil
}

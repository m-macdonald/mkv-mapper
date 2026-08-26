package engine

import (
	"context"
	"fmt"
	"strconv"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/naming"
	"m-macdonald/mkv-mapper/internal/validation"
)

type BuildRipPlanConfig struct {
	DiscRoot          string
	Templates         config.TemplateConfig
	Rip               config.RipConfig
	Backup            config.BackupConfig
}

type planContent struct {
	titles      []model.TitleRipPlan
	buildReport model.BuildReport
}

func (e *Engine) BuildRipPlan(
	ctx context.Context,
	cfg BuildRipPlanConfig,
) (model.RipPlan, error) {
	identity, discInfo, err := e.ScanDisc(ctx, cfg.DiscRoot)
	if err != nil {
		return model.RipPlan{}, err
	}
	return e.CompleteRipPlan(ctx, identity, discInfo, cfg)
}

func (e *Engine) CompleteRipPlan(
	ctx context.Context,
	identity model.DiscIdentity,
	discInfo makemkv.DiscInfo,
	cfg BuildRipPlanConfig,
) (model.RipPlan, error) {
	hash, err := files.Hash(identity.DiscRoot)
	if err != nil {
		return model.RipPlan{}, fmt.Errorf("unable to hash disc: %w", err)
	}

	disc, err := e.discdb.LookupDisc(ctx, hash)
	if err != nil {
		return model.RipPlan{}, fmt.Errorf("failed to retrieve disc definitions from TheDiscDB: %w", err)
	}

	mappings := mapper.MapTitles(disc, discInfo.Titles)
	planContent, err := resolveFilenames(cfg.Templates, mappings, disc)
	if err != nil {
		return model.RipPlan{}, err
	}

	outputDir, err := resolveRipOutputDirectory(cfg.Rip.OutputDirTemplate, disc)
	if err != nil {
		return model.RipPlan{}, err
	}

	plan := model.RipPlan{
		RipPlanBase: model.RipPlanBase{
			DiscIdentity: identity,
			MediaInfo: model.MediaInfo{
				Title: disc.Media.Title,
				Year:  disc.Media.Year,
			},
			DiscInfo: model.DiscInfo{
				Format: disc.Disc.Format,
				Hash:   hash,
			},
			OutputDir: outputDir,
			Titles:    planContent.titles,
		},
		BuildReport: planContent.buildReport,
	}

	return plan, nil
}

func resolveRipOutputDirectory(template string, discRecord discdb.DiscRecord) (string, error) {
	directoryGen, err := naming.NewRipOutputDirGenerator(template)
	if err != nil {
		return "", fmt.Errorf("resolving output directory: %w", err)
	}
	outputDir, err := directoryGen.Generate(naming.DirectoryContext{
		Media: discRecord.Media,
		Disc:  discRecord.Disc,
	})
	if err != nil {
		return "", fmt.Errorf("resolving output directory: %w", err)
	}
	if err := files.EnsureDir(outputDir); err != nil {
		return "", err
	}
	return outputDir, nil
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

		content.titles = append(content.titles, model.TitleRipPlan{
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

func (e *Engine) RunRipPlan(
	ctx context.Context,
	plan model.ValidatedRipPlan,
	onEvent EngineEventSink,
) error {
	if err := plan.Err(); err != nil {
		return err
	}

	if plan.IsAllTitles {
		if err := e.ripAll(ctx, plan.DiscRoot, plan.OutputDir, onEvent); err != nil {
			return err
		}
	} else {
		if err := e.ripSelected(ctx, plan.DiscRoot, plan.OutputDir, plan.Titles, onEvent); err != nil {
			return err
		}
	}

	mappings := make(map[string]string)
	for _, titlePlan := range plan.Titles {
		mappings[titlePlan.MakeMkvOutputFile] = titlePlan.FinalName
	}
	errs := mapper.RenameTitles(plan.OutputDir, plan.OutputDir, mappings)
	if len(errs) != 0 {
		e.logger.Errorf("%v", errs)
	}

	return nil
}

func (e *Engine) ripAll(
	ctx context.Context,
	ripSource string,
	outputDir string,
	onEvent EngineEventSink,
) error {
	return e.makemkv.RipDisc(
		ctx,
		ripSource,
		outputDir,
		func(pl lines.ParsedLine) {
			if event, ok := parsedLineToEvent(pl); ok {
				onEvent(event)
			}
		})
}

func (e *Engine) ripSelected(
	ctx context.Context,
	ripSource string,
	outputDir string,
	titles []model.TitleRipPlan,
	onEvent EngineEventSink,
) error {
	for _, title := range titles {
		err := e.makemkv.RipTitle(
			ctx,
			ripSource,
			outputDir,
			title.TitleId,
			func(pl lines.ParsedLine) {
				if event, ok := parsedLineToEvent(pl); ok {
					onEvent(event)
				}
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func RipChecks(plan model.RipPlan, estimatedBytes uint64) validation.CheckGroup {
	targets := make([]validation.FilenameTarget, 0, len(plan.Titles))
	for _, title := range plan.Titles {
		targets = append(targets, validation.FilenameTarget{
			ID:       strconv.Itoa(int(title.TitleId)),
			FileName: title.FinalName,
		})
	}
	return validation.CheckGroup{
		Label: "Rip",
		Checks: []validation.Check{
			validation.OutputDirValid(plan.OutputDir),
			validation.FreeSpace(plan.OutputDir, estimatedBytes),
			validation.NoConflicts(plan.OutputDir, targets),
		},
	}
}

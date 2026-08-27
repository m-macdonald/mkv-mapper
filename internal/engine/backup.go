package engine

import (
	"context"
	"fmt"

	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/naming"
	"m-macdonald/mkv-mapper/internal/validation"
)

type BuildBackupPlanConfig struct {
	OutputDirTemplate string
	KeepBackup        bool
}

func (e *Engine) CompleteBackupPlan(
	identity model.DiscIdentity,
	discInfo makemkv.DiscInfo,
	cfg BuildBackupPlanConfig,
) (model.BackupPlan, error) {
	titles := make([]model.BackupTitle, 0, len(discInfo.Titles))
	for _, title := range discInfo.Titles {
		titles = append(titles, model.BackupTitle{
			TitleId:       title.TitleId,
			EstimatedSize: title.OutputFileSize,
		})
	}

	outputDir, err := resolveBackupOutputDirectory(cfg.OutputDirTemplate, identity)
	if err != nil {
		return model.BackupPlan{}, err
	}

	return model.BackupPlan{
		DiscIdentity: identity,
		OutputDir:    outputDir,
		KeepBackup:   cfg.KeepBackup,
		Titles:       titles,
	}, nil
}

func resolveBackupOutputDirectory(template string, discIdentity model.DiscIdentity) (string, error) {
	directoryGen, err := naming.NewBackupOutputDirGenerator(template)
	if err != nil {
		return "", fmt.Errorf("resolving output directory: %w", err)
	}
	outputDir, err := directoryGen.Generate(naming.BackupDirectoryContext{
		Label: discIdentity.Label,
	})
	if err != nil {
		return "", fmt.Errorf("resolving output directory: %w", err)
	}
	if err := files.EnsureDir(outputDir); err != nil {
		return "", err
	}
	return outputDir, nil
}

func (e *Engine) RunBackupPlan(
	ctx context.Context,
	plan model.ValidatedBackupPlan,
	onEvent EngineEventSink,
) error {
	if err := plan.Err(); err != nil {
		return err
	}
	source, err := e.discResolver.ResolveByLabel(ctx, plan.DiscIdentity.Label)
	if err != nil {
		return fmt.Errorf("resolving backup source: %w", err)
	}

	if !source.IsOptical() {
		return fmt.Errorf("backup requires an optical source (disc:/dev:), got %q", source)
	}

	e.logger.Info("Beginning disc backup...")

	return e.makemkv.BackupDisc(
		ctx,
		string(source),
		plan.OutputDir,
		func(pl lines.ParsedLine) {
			if event, ok := parsedLineToEvent(pl); ok {
				onEvent(event)
			}
		},
	)
}

func BackupChecks(dir string, estimatedBytes uint64) validation.CheckGroup {
	return validation.CheckGroup{
		Label: "Backup",
		Checks: []validation.Check{
			validation.FreeSpace(dir, estimatedBytes),
		},
	}
}

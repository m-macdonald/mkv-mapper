package engine

import (
	"context"
	"fmt"

	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/validation"
)

func (e *Engine) BackupPlanDisc(ctx context.Context, plan model.ValidatedPlan, outputDir string, onEvent EngineEventSink) error {
	if err := plan.Err(); err != nil {
		return err
	}
	source, err := e.discResolver.ResolveByLabel(ctx, plan.Disc.Label)
	if err != nil {
		return fmt.Errorf("resolving backup source: %w", err)
	}
	return e.Backup(ctx, source, outputDir, onEvent)
}

func (e *Engine) Backup(ctx context.Context, specified files.DiscSource, outputDir string, onEvent EngineEventSink) error {
	source, err := e.discResolver.Resolve(ctx, specified)
	if err != nil {
		return err
	}
	if !source.IsOptical() {
		return fmt.Errorf("backup requires an optical source (disc:/dev:), got %q", source)
	}

	e.logger.Info("Beginning disc backup...")
	return e.makemkv.BackupDisc(
		ctx,
		string(source),
		outputDir,
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

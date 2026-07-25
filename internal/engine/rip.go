package engine

import (
	"context"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/validation"
)

func (e *Engine) RunPlan(
	ctx context.Context,
	plan model.ValidatedPlan,
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
	titles []model.TitlePlan,
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

func RipChecks(plan model.SelectedPlan, estimatedBytes uint64) validation.CheckGroup {
	targets := make([]validation.FilenameTarget, 0, len(plan.Titles))
	for _, title := range plan.Titles {
		targets = append(targets, validation.FilenameTarget{
			ID:       string(title.TitleId),
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

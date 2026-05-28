package engine

import (
	"context"
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/event"
	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/planner"

	"go.uber.org/zap"
)

type Engine struct {
	makemkv *makemkv.Client
	discdb  *discdb.CachedClient
	logger  *zap.SugaredLogger
}

type ValidatedPlan struct {
	planner.Plan
	ValidationReport ValidationReport
}

type EngineEventSink func(event.Event)

func New(
	makemkv *makemkv.Client,
	discdb *discdb.CachedClient,
	logger *zap.SugaredLogger,
) *Engine {
	return &Engine{
		makemkv: makemkv,
		discdb:  discdb,
		logger:  logger,
	}
}

func (e *Engine) BuildPlan(
	ctx context.Context,
	discRoot string,
	outputDir string,
	templateConfig config.TemplateConfig,
) (planner.Plan, error) {
	root, err := files.ResolveDiscRoot(discRoot)
	if err != nil {
		return planner.Plan{}, fmt.Errorf("unable to find disc root %w", err)
	}
	hash, err := files.Hash(root)
	if err != nil {
		return planner.Plan{}, fmt.Errorf("unable to hash disc %w", err)
	}

	disc, err := e.discdb.LookupDisc(ctx, hash)
	if err != nil {
		return planner.Plan{}, fmt.Errorf("failed to retrieve disc definitions from TheDiscDB %w", err)
	}

	titles, err := e.makemkv.ReadTitles(ctx, root)
	if err != nil {
		return planner.Plan{}, fmt.Errorf("unable to read disc titles using MakeMkv %w", err)
	}

	return planner.BuildPlan(root, outputDir, templateConfig, disc, titles)
}

func (e *Engine) ValidatePlan(plan planner.Plan) *ValidatedPlan {
	return &ValidatedPlan{
		Plan:             plan,
		ValidationReport: validatePlan(plan.DiscPlan),
	}
}

func (e *Engine) RunPlan(
	ctx context.Context,
	plan ValidatedPlan,
	onEvent EngineEventSink,
) error {
	err := e.makemkv.RipDisc(
		ctx,
		plan.DiscPlan.DiscRoot,
		plan.DiscPlan.OutputDir,
		func(pl lines.ParsedLine) {
			if event, ok := parsedLineToEvent(pl); ok {
				onEvent(event)
			}
		})
	if err != nil {
		return err
	}

	mappings := make(map[string]string)
	for _, titlePlan := range plan.DiscPlan.Titles {
		mappings[titlePlan.MakeMkvOutputFile] = titlePlan.FinalName
	}
	errs := mapper.RenameTitles(plan.DiscPlan.OutputDir, plan.DiscPlan.OutputDir, mappings)
	if len(errs) != 0 {
		e.logger.Errorf("%v", errs)
	}

	return nil
}

func parsedLineToEvent(line lines.ParsedLine) (event.Event, bool) {
	switch l := line.(type) {
	case lines.ProgressValue:
		return event.ProgressPercentEvent{
			TotalPercent:   l.TotalPercent(),
			CurrentPercent: l.CurrentPercent(),
		}, true
	case lines.ProgressCurrent:
		return event.ProgressCurrentEvent{
			Message: l.Name,
		}, true
	case lines.ProgressTitle:
		return event.ProgressTotalEvent{
			Message: l.Name,
		}, true
	case lines.Message:
		return event.MessageEvent{
			Message: l.Message,
		}, true
	}

	return nil, false
}

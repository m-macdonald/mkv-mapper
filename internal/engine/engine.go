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
) (*planner.DiscPlan, *planner.BuildReport, error) {
	root, err := files.ResolveDiscRoot(discRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find disc root %w", err)
	}
	hash, err := files.Hash(root)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to hash disc %w", err)
	}

	disc, err := e.discdb.LookupDisc(ctx, hash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve disc definitions from TheDiscDB %w", err)
	}

	titles, err := e.makemkv.ReadTitles(ctx, root)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read disc titles using MakeMkv %w", err)
	}

	return planner.BuildPlan(root, outputDir, templateConfig, disc, titles)
}

// TODO: Move validation into a separate package?
func (e *Engine) ValidatePlan(plan *planner.DiscPlan) *ValidationReport {
	return ValidatePlan(plan)
}

func (e *Engine) RunPlan(
	ctx context.Context,
	plan *planner.DiscPlan,
	onEvent EngineEventSink,
) error {
	err := e.makemkv.RipDisc(
		ctx,
		plan.DiscRoot,
		plan.OutputDir,
		func(pl lines.ParsedLine) {
			if event, ok := event.ParsedLineToEvent(pl); ok {
				onEvent(event)
			}
		})
	if err != nil {
		return err
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

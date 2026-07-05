package engine

import (
	"context"
	"fmt"
	"strings"

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
	makemkv  *makemkv.Client
	discdb   *discdb.CachedClient
	selector planner.Selector
	logger   *zap.SugaredLogger
}

type EngineEventSink func(event.Event)

func New(
	makemkv *makemkv.Client,
	discdb *discdb.CachedClient,
	logger *zap.SugaredLogger,
	selector planner.Selector,
) *Engine {
	return &Engine{
		makemkv:  makemkv,
		discdb:   discdb,
		logger:   logger,
		selector: selector,
	}
}

func (e *Engine) BuildPlan(
	ctx context.Context,
	discRoot string,
	outputDir string,
	templateConfig config.TemplateConfig,
) (planner.Plan, error) {
	discMounts, err := files.ResolveDiscRoot(discRoot)
	switch {
	case err != nil:
		return planner.Plan{}, fmt.Errorf("failure resolving disc root: %w", err)
	case len(discMounts) < 1:
		return planner.Plan{}, fmt.Errorf("no disc found")
	case len(discMounts) > 1:
		return planner.Plan{}, fmt.Errorf("multiple discs found: %s\nUse --disc-root to specify which disc to rip", strings.Join(discMounts, ", "))
	}

	// The above switch makes sure that we can safely get the first (and only) element
	root := discMounts[0]

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

func (e *Engine) SelectPlan(mode config.SelectionMode, plan planner.Plan) (planner.SelectedPlan, error) {
	var selection planner.Selection
	var err error
	switch mode {
	case config.ModeFullAuto:
		selection = planner.FullSelection(plan)
	case config.ModeTrimmedAuto:
		selection = planner.TrimmedSelection(plan)
	case config.ModeManual:
		selection, err = e.selector.Select(plan)
	default:
		return planner.SelectedPlan{}, fmt.Errorf("unknown selection mode: %v", mode)
	}
	if err != nil {
		return planner.SelectedPlan{}, err
	}

	return planner.NewSelectedPlan(plan, selection)
}

func (e *Engine) ValidatePlan(plan planner.SelectedPlan) planner.ValidatedPlan {
	return planner.ValidatePlan(plan)
}

func (e *Engine) RunPlan(
	ctx context.Context,
	plan planner.ValidatedPlan,
	onEvent EngineEventSink,
) error {
	if plan.ValidationReport.HasErrors() {
		return fmt.Errorf("plan has validation errors, aborting rip")
	}

	if plan.IsAllTitles {
		if err := e.ripAll(ctx, plan, onEvent); err != nil {
			return err
		}
	} else {
		if err := e.ripSelected(ctx, plan, onEvent); err != nil {
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

func (e *Engine) ripAll(ctx context.Context, plan planner.ValidatedPlan, onEvent EngineEventSink) error {
	return e.makemkv.RipDisc(
		ctx,
		plan.DiscRoot,
		plan.OutputDir,
		func(pl lines.ParsedLine) {
			if event, ok := parsedLineToEvent(pl); ok {
				onEvent(event)
			}
		})
}

func (e *Engine) ripSelected(ctx context.Context, plan planner.ValidatedPlan, onEvent EngineEventSink) error {
	for _, title := range plan.Titles {
		err := e.makemkv.RipTitle(
			ctx,
			plan.DiscRoot,
			plan.OutputDir,
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

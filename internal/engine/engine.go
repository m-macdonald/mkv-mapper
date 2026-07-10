package engine

import (
	"context"
	"fmt"
	"os"
	"strings"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/event"
	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/model"

	"go.uber.org/zap"
)

type Engine struct {
	makemkv  *makemkv.Client
	discdb   *discdb.CachedClient
	selector model.Selector
	logger   *zap.SugaredLogger
}

type BuildPlanConfig struct {
	DiscRoot  string
	OutputDir string
	Templates config.TemplateConfig
	Rip       config.RipConfig
}

type EngineEventSink func(event.Event)

func New(
	makemkv *makemkv.Client,
	discdb *discdb.CachedClient,
	logger *zap.SugaredLogger,
	selector model.Selector,
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
	cfg BuildPlanConfig,
) (model.Plan, error) {
	discRoot, err := files.ResolveDiscRoot(cfg.DiscRoot)
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
	
	return buildPlan(discRoot, cfg, disc, discInfo)
}

func (e *Engine) SelectPlan(mode config.SelectionMode, plan model.Plan) (model.SelectedPlan, error) {
	var selection model.Selection
	var err error
	switch mode {
	case config.ModeFullAuto:
		selection = model.FullSelection(plan)
	case config.ModeTrimmedAuto:
		selection = model.TrimmedSelection(plan)
	case config.ModeManual:
		selection, err = e.selector.Select(plan)
	default:
		return model.SelectedPlan{}, fmt.Errorf("unknown selection mode: %v", mode)
	}
	if err != nil {
		return model.SelectedPlan{}, err
	}

	return model.NewSelectedPlan(plan, selection)
}

func (e *Engine) ValidatePlan(plan model.SelectedPlan) model.ValidatedPlan {
	report := buildValidationReport(plan)
	return model.ValidatedPlan{
		PlanBase:         plan.PlanBase,
		BuildReport:      plan.BuildReport,
		ValidationReport: report,
		IsAllTitles:      plan.IsAllTitles,
	}
}

func (e *Engine) RunPlan(
	ctx context.Context,
	plan model.ValidatedPlan,
	onEvent EngineEventSink,
) error {
	if plan.ValidationReport.HasErrors() {
		return fmt.Errorf("plan has validation errors, aborting rip")
	}

	if plan.Backup {
		backupSource, err := e.resolveBackupSource(ctx, plan.Disc.Label)
		if err != nil {
			return err
		}
		e.logger.Info("Beginning disc backup...")
		e.makemkv.BackupDisc(
			ctx,
			backupSource,
			plan.BackupDir,
			func(pl lines.ParsedLine) {
				if event, ok := parsedLineToEvent(pl); ok {
					onEvent(event)
				}
			})
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

	// TODO: Move this direct os use elsewhere. files package?
	if plan.Backup && !plan.KeepBackup {
		err := os.Remove(plan.BackupDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) resolveBackupSource(ctx context.Context, knownLabel string) (string, error) {
	drives, err := e.makemkv.ScanDrives(ctx)
	if err != nil {
		return "", fmt.Errorf("scanning drives: %w", err)
	}

	var matches []lines.DriveScan
	for _, d := range drives {
		if d.DiscName == knownLabel {
			matches = append(matches, d)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no drive found with disc label %q — was the disc ejected or replaced?", knownLabel)
	case 1:
		return fmt.Sprintf("disc:%d", matches[0].Index), nil
	default:
		var devices []string
		for _, d := range matches {
			devices = append(devices, d.Device)
		}
		return "", fmt.Errorf("multiple drives have a disc labeled %q (%s) — remove the duplicate before backing up", knownLabel, strings.Join(devices, ", "))
	}
}

func (e *Engine) ripAll(ctx context.Context, plan model.ValidatedPlan, onEvent EngineEventSink) error {
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

func (e *Engine) ripSelected(ctx context.Context, plan model.ValidatedPlan, onEvent EngineEventSink) error {
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

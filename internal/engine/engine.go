package engine 

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/planner"

	"go.uber.org/zap"
)

type Engine struct {
	makemkv *makemkv.Client
	discdb  *discdb.Client
	logger  *zap.SugaredLogger
}

type EventSink func(RipEvent)

type RipEvent struct {
	Type         RipEventType
	TitlePercent float64
	DiscPercent  float64
	TitlePlan    *planner.TitlePlan
	Message      string
}

type RipEventType string

const (
	EventTitleStarted  RipEventType = "title_started"
	EventTitleProgress RipEventType = "title_progress"
	EventTitleFinished RipEventType = "title_finished"
)

func New(
	makemkv *makemkv.Client,
	discdb *discdb.Client,
	logger *zap.SugaredLogger,
) *Engine {
	return &Engine{
		makemkv: makemkv,
		discdb:  discdb,
		logger:  logger,
	}
}

func (e *Engine) BuildPlan(
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

	disc, err := e.discdb.GetDisc(hash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve disc definitions from TheDiscDB %w", err)
	}

	titles, err := e.makemkv.ReadTitles(root)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read disc titles using MakeMkv %w", err)
	}

	return planner.BuildPlan(root, outputDir, templateConfig, disc, titles)
}

// TODO: Move validation into a separate package?
func (e *Engine) ValidatePlan(plan *planner.DiscPlan) *ValidationReport {
	return ValidatePlan(plan)
}

func (e *Engine) RunPlan(plan *planner.DiscPlan, sink EventSink) error {
	err := e.makemkv.RipDisc(plan.DiscRoot, plan.OutputDir)
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

package app

import (
	"context"
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/engine"
	"m-macdonald/mkv-mapper/internal/planner"

	"go.uber.org/zap"
)

type Ripper struct {
	engine *engine.Engine
	logger *zap.SugaredLogger
}

type RipPreview struct {
	Plan             *planner.DiscPlan
	BuildReport      *planner.BuildReport
	ValidationReport *engine.ValidationReport
}

type ExecutionReport struct{}

func NewRipper(engine *engine.Engine, logger *zap.SugaredLogger) *Ripper {
	return &Ripper{
		engine: engine,
		logger: logger,
	}
}

func (r *Ripper) PreviewRip(
	ctx context.Context,
	discRoot string,
	outputDir string,
	templates config.TemplateConfig,
) (*RipPreview, error) {
	plan, buildReport, err := r.engine.BuildPlan(ctx, discRoot, outputDir, templates)
	if err != nil {
		return nil, fmt.Errorf("build plan: %w", err)
	}
	validationReport := r.engine.ValidatePlan(plan)

	return &RipPreview{
		Plan:             plan,
		BuildReport:      buildReport,
		ValidationReport: validationReport,
	}, nil
}

func (r *Ripper) ExecuteRip(
	ctx context.Context,
	plan *planner.DiscPlan,
	onEvent engine.EngineEventSink,
) error {
	return r.engine.RunPlan(ctx, plan, onEvent)
}

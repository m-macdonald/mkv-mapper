package app

import (
	"context"
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/engine"
	"m-macdonald/mkv-mapper/internal/makemkv"
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
	onLine makemkv.LineSink,
) error {
	return r.engine.RunPlan(ctx, plan, onLine)
}

// func (r *Ripper) RipDisc(
// 	discRoot string,
// 	outputDir string,
// 	templates config.TemplateConfig,
// ) (*RipResult, error) {
// 	ripResult := &RipResult{
// 		Preview: &RipPreview{},
// 	}
// 	ripResult.Preview.BuildReport = buildReport
//
// 	// TODO: Log plan and build report
// 	ripResult.Preview.ValidationReport = validationReport
//
//
// 	return ripResult, nil
// }

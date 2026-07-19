package engine

import (
	"context"

	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/validation"
)

func (e *Engine) ValidatePlan(ctx context.Context, plan model.SelectedPlan, checkGroups []validation.CheckGroup) model.ValidatedPlan {
	report := validation.Run(ctx, checkGroups)
	return model.ValidatedPlan{
		PlanBase:         plan.PlanBase,
		BuildReport:      plan.BuildReport,
		ValidationReport: report,
		IsAllTitles:      plan.IsAllTitles,
	}
}

package engine

import (
	"context"

	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/validation"
)

func (e *Engine) ValidateRipPlan(
	ctx context.Context,
	plan model.RipPlan,
	checkGroups []validation.CheckGroup,
) model.ValidatedRipPlan {
	report := validation.Run(ctx, checkGroups)
	return model.ValidatedRipPlan{
		RipPlanBase:      plan.RipPlanBase,
		BuildReport:      plan.BuildReport,
		ValidationReport: report,
		IsAllTitles:      plan.IsAllTitles,
	}
}

func (e *Engine) ValidateBackupPlan(
	ctx context.Context,
	plan model.BackupPlan,
	checkGroups []validation.CheckGroup,
) model.ValidatedBackupPlan {
	report := validation.Run(ctx, checkGroups)
	return model.ValidatedBackupPlan{
		BackupPlan: plan,
		Report:     report,
	}
}

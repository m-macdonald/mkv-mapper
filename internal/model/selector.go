package model

import "m-macdonald/mkv-mapper/internal/makemkv/lines"

type Selector interface {
	Select(plan Plan) ([]lines.TitleId, error)
}

type SelectedPlan struct {
	PlanBase
	// Indicates if the selected titles include every title from the plan
	IsAllTitles bool
	BuildReport BuildReport
}

func NewSelectedPlan(plan Plan, selection Selection) (SelectedPlan, error) {
	updatedBase := plan.PlanBase
	updatedBase.Titles = selection.Selected
	return SelectedPlan{
		PlanBase:    updatedBase,
		IsAllTitles: len(plan.Titles) == len(selection.Selected),
		BuildReport: plan.BuildReport,
	}, nil
}

package model

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"
)

type Selector interface {
	Select(plan Plan) (Selection, error)
}

func NewSelectedPlan(plan Plan, selection Selection) (SelectedPlan, error) {
	byId := make(map[lines.TitleId]TitlePlan, len(plan.Titles))
	for _, title := range plan.Titles {
		byId[title.TitleId] = title
	}

	titles := make([]TitlePlan, 0, len(selection.SelectedIds))
	for _, id := range selection.SelectedIds {
		title, ok := byId[id]
		if !ok {
			return SelectedPlan{}, fmt.Errorf("selection references unknown title id %v", id)
		}
		titles = append(titles, title)
	}
	updatedBase := plan.PlanBase
	updatedBase.Titles = titles
	return SelectedPlan{
		PlanBase:    updatedBase,
		IsAllTitles: len(plan.Titles) == len(selection.SelectedIds),
		BuildReport: plan.BuildReport,
	}, nil
}

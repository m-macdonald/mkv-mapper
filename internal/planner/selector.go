package planner

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
)

type Selection struct {
	Mode        config.SelectionMode
	SelectedIds []lines.TitleId
}

func (s Selection) IsSelected(id lines.TitleId) bool {
	for _, sel := range s.SelectedIds {
		if sel == id {
			return true
		}
	}
	return false
}

func FullSelection(plan Plan) Selection {
	ids := make([]lines.TitleId, 0, len(plan.Titles))
	for _, title := range plan.Titles {
		ids = append(ids, title.TitleId)
	}
	return Selection{Mode: config.ModeFullAuto, SelectedIds: ids}
}

func TrimmedSelection(plan Plan) Selection {
	var ids []lines.TitleId
	for _, title := range plan.Titles {
		if title.IsMatched {
			ids = append(ids, title.TitleId)
		}
	}
	return Selection{Mode: config.ModeTrimmedAuto, SelectedIds: ids}
}

type SelectedPlan struct {
	PlanBase
	BuildReport BuildReport
}

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
	return SelectedPlan{PlanBase: updatedBase, BuildReport: plan.BuildReport}, nil
}

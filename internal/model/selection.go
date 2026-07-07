package model 

import (
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
	// Indicates if the selected titles include every title from the plan
	IsAllTitles bool
	BuildReport BuildReport
}


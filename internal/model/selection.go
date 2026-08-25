package model

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
)

type Selection struct {
	Mode     config.SelectionMode
	Selected []TitleRipPlan
}

func (s Selection) IsSelected(id lines.TitleId) bool {
	for _, sel := range s.Selected {
		if sel.TitleId == id {
			return true
		}
	}
	return false
}

func FullSelection(plan RipPlan) Selection {
	titles := filterTitles(plan.Titles, func(tp TitleRipPlan) bool { return true })
	return Selection{Mode: config.ModeFullAuto, Selected: titles}
}

func TrimmedSelection(plan RipPlan) Selection {
	titles := filterTitles(plan.Titles, func(tp TitleRipPlan) bool { return tp.IsMatched })
	return Selection{Mode: config.ModeTrimmedAuto, Selected: titles}
}

func SelectionFromIds(plan RipPlan, mode config.SelectionMode, ids []lines.TitleId) (Selection, error) {
	titlesById := make(map[lines.TitleId]TitleRipPlan, len(plan.Titles))
	for _, title := range plan.Titles {
		titlesById[title.TitleId] = title
	}
	titles := make([]TitleRipPlan, 0, len(ids))
	for _, id := range ids {
		title, ok := titlesById[id]
		if !ok {
			return Selection{}, fmt.Errorf("selection references unknown title id %v", id)
		}
		titles = append(titles, title)
	}
	return Selection{Mode: mode, Selected: titles}, nil
}

func filterTitles(titles []TitleRipPlan, keep func(TitleRipPlan) bool) []TitleRipPlan {
	var kept []TitleRipPlan
	for _, title := range titles {
		if keep(title) {
			kept = append(kept, title)
		}
	}
	return kept
}

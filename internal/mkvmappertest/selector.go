package mkvmappertest

import (
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/planner"
)

type Selector struct {
	selection planner.Selection
	err       error
}

func NewSelector(selection planner.Selection) *Selector {
	return &Selector{selection: selection}
}

func NewSelectorWithError(err error) *Selector {
	return &Selector{err: err}
}

func (s *Selector) Select(plan planner.Plan) (planner.Selection, error) {
	if s.err != nil {
		return planner.Selection{}, s.err
	}
	return s.selection, nil
}

func Selection(mode config.SelectionMode, ids ...lines.TitleId) planner.Selection {
	return planner.Selection{
		Mode:        mode,
		SelectedIds: ids,
	}
}

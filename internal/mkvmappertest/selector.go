package mkvmappertest

import (
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/model"
)

type Selector struct {
	selection model.Selection
	err       error
}

func NewSelector(selection model.Selection) *Selector {
	return &Selector{selection: selection}
}

func NewSelectorWithError(err error) *Selector {
	return &Selector{err: err}
}

func (s *Selector) Select(plan model.Plan) (model.Selection, error) {
	if s.err != nil {
		return model.Selection{}, s.err
	}
	return s.selection, nil
}

func Selection(mode config.SelectionMode, ids ...lines.TitleId) model.Selection {
	return model.Selection{
		Mode:        mode,
		SelectedIds: ids,
	}
}

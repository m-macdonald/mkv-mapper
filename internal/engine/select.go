package engine

import (
	"fmt"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/model"
)

func (e *Engine) SelectPlan(mode config.SelectionMode, plan model.Plan) (model.SelectedPlan, error) {
	var selection model.Selection
	var err error
	switch mode {
	case config.ModeFullAuto:
		selection = model.FullSelection(plan)
	case config.ModeTrimmedAuto:
		selection = model.TrimmedSelection(plan)
	case config.ModeManual:
		selection, err = e.selector.Select(plan)
	default:
		return model.SelectedPlan{}, fmt.Errorf("unknown selection mode: %v", mode)
	}
	if err != nil {
		return model.SelectedPlan{}, err
	}

	return model.NewSelectedPlan(plan, selection)
}

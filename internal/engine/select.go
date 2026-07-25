package engine

import (
	"fmt"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/model"
)

func (e *Engine) SelectPlan(mode config.SelectionMode, plan model.Plan) (model.Plan, error) {
	var selection model.Selection
	var err error
	switch mode {
	case config.ModeFullAuto:
		selection = model.FullSelection(plan)
	case config.ModeTrimmedAuto:
		selection = model.TrimmedSelection(plan)
	case config.ModeManual:
		ids, err := e.selector.Select(plan)
		if err != nil {
			return model.Plan{}, err
		}
		selection, err = model.SelectionFromIds(plan, mode, ids)
		if err != nil {
			return model.Plan{}, err
		}
	default:
		return model.Plan{}, fmt.Errorf("unknown selection mode: %v", mode)
	}
	if err != nil {
		return model.Plan{}, err
	}

	return model.NewSelectedPlan(plan, selection)
}

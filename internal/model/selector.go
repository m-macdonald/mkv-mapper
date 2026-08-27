package model

import "m-macdonald/mkv-mapper/internal/makemkv/lines"

type Selector interface {
	Select(plan RipPlan) ([]lines.TitleId, error)
}

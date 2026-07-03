package planner

import "m-macdonald/mkv-mapper/internal/makemkv/lines"

type BuildReport struct {
	Warnings []PlanWarning
}

type PlanWarning struct {
	TitleId lines.TitleId
	Code    WarningCode
	Message string
	Cause   error
}

type WarningCode string

const (
	WarningNoMetadata WarningCode = "no_metadata"
)

package model

import (
	"fmt"
	"m-macdonald/mkv-mapper/internal/validation"
)

type ValidatedRipPlan struct {
	RipPlanBase
	BuildReport      BuildReport
	IsAllTitles      bool
	ValidationReport validation.Report
}

func (v ValidatedRipPlan) Err() error {
	if v.ValidationReport.HasErrors() {
		return fmt.Errorf("plan has validation errors")
	}
	return nil
}

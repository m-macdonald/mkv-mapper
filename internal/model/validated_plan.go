package model

import (
	"fmt"
	"m-macdonald/mkv-mapper/internal/validation"
)

type ValidatedPlan struct {
	PlanBase
	BuildReport      BuildReport
	IsAllTitles      bool
	ValidationReport validation.Report
}

func (v ValidatedPlan) Err() error {
	if v.ValidationReport.HasErrors() {
		return fmt.Errorf("plan has validation errors")
	}
	return nil
}

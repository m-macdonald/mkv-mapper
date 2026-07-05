package planner

import "m-macdonald/mkv-mapper/internal/makemkv/lines"

type ValidatedPlan struct {
	PlanBase
	BuildReport      BuildReport
	IsAllTitles      bool
	ValidationReport ValidationReport
}

func NewValidatedPlan(selectedPlan SelectedPlan, report ValidationReport) ValidatedPlan {
	return ValidatedPlan{
		PlanBase:         selectedPlan.PlanBase,
		BuildReport:      selectedPlan.BuildReport,
		ValidationReport: report,
	}
}

type ValidationReport struct {
	Results []ValidationResult
}

func (v *ValidationReport) AddResult(result ValidationResult) {
	v.Results = append(v.Results, result)
}

func (v *ValidationReport) Passes() []ValidationResult {
	return v.filter(ValidationStatusPass)
}

func (v *ValidationReport) Warnings() []ValidationResult {
	return v.filter(ValidationStatusWarn)
}

func (v *ValidationReport) Errors() []ValidationResult {
	return v.filter(ValidationStatusFail)
}

func (v *ValidationReport) HasErrors() bool {
	return len(v.Errors()) > 0
}

func (v *ValidationReport) filter(status ValidationStatus) []ValidationResult {
	var results []ValidationResult
	for _, r := range v.Results {
		if r.Status == status {
			results = append(results, r)
		}
	}
	return results
}

type ValidationResult struct {
	Status  ValidationStatus
	Code    ValidationCode
	Message string
	Cause   error
	TitleId *lines.TitleId
}

type ValidationStatus string

const (
	ValidationStatusPass ValidationStatus = "pass"
	ValidationStatusWarn ValidationStatus = "warn"
	ValidationStatusFail ValidationStatus = "fail"
)

type ValidationCode string

const (
	ValidationInsufficientSpace ValidationCode = "insufficient_space"
	ValidationOutputExists      ValidationCode = "output_exists"
	ValidationOutputDirInvalid  ValidationCode = "output_dir_invalid"
)

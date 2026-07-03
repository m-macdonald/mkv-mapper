package planner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/util"
)

type ValidatedPlan struct {
	PlanBase
	BuildReport      BuildReport
	ValidationReport ValidationReport
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

func ValidatePlan(plan SelectedPlan) ValidatedPlan {
	report := &ValidationReport{}

	validateOutputDir(plan, report)
	validateDiskSpace(plan, report)
	validateExistingFiles(plan, report)

	return ValidatedPlan{
		PlanBase: plan.PlanBase,
		BuildReport: plan.BuildReport,
		ValidationReport: *report,
	}
}

func validateOutputDir(plan SelectedPlan, report *ValidationReport) {
	info, err := os.Stat(plan.OutputDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Making this an error for now. I might just auto-create the outputdir in the future
			report.AddResult(ValidationResult{
				Status:  ValidationStatusFail,
				Code:    ValidationOutputDirInvalid,
				Message: fmt.Sprintf("output directory does not exist: %s", plan.OutputDir),
				Cause:   err,
			})

			return
		}
		report.AddResult(ValidationResult{
			Status:  ValidationStatusFail,
			Code:    ValidationOutputDirInvalid,
			Message: fmt.Sprintf("could not stat output directory: %s", plan.OutputDir),
			Cause:   err,
		})

		return
	}

	if !info.IsDir() {
		report.AddResult(ValidationResult{
			Status:  ValidationStatusFail,
			Code:    ValidationOutputDirInvalid,
			Message: fmt.Sprintf("output path is not a directory: %s", plan.OutputDir),
		})

		return
	}

	report.AddResult(ValidationResult{
		Status:  ValidationStatusPass,
		Message: fmt.Sprintf("output directory valid: %s", plan.OutputDir),
	})
}

func validateDiskSpace(plan SelectedPlan, report *ValidationReport) {
	free, err := files.GetFreeDiskSpace(plan.OutputDir)
	if err != nil {
		report.AddResult(ValidationResult{
			Status:  ValidationStatusFail,
			Code:    ValidationOutputDirInvalid,
			Message: fmt.Sprintf("could not determine free space for output directory: %s", plan.OutputDir),
			Cause:   err,
		})

		return
	}

	var required uint64
	for _, title := range plan.Titles {
		required += uint64(title.EstimatedSize)
	}

	if free < required {
		report.AddResult(ValidationResult{
			Status: ValidationStatusFail,
			Code:   ValidationInsufficientSpace,
			Message: fmt.Sprintf(
				"insufficient disk space %s: need %s, have %s",
				plan.OutputDir,
				util.FormatSize(required),
				util.FormatSize(free)),
		})

		return
	}

	report.AddResult(ValidationResult{
		Status: ValidationStatusPass,
		Message: fmt.Sprintf(
			"sufficient disk space %s: need %s, have %s",
			plan.OutputDir,
			util.FormatSize(required),
			util.FormatSize(free)),
	})
}

func validateExistingFiles(plan SelectedPlan, report *ValidationReport) {
	hasIssue := false
	for _, title := range plan.Titles {
		outPath := filepath.Join(plan.OutputDir, title.FinalName)

		_, err := os.Stat(outPath)
		if err == nil {
			titleId := title.TitleId
			report.AddResult(ValidationResult{
				Status:  ValidationStatusFail,
				Code:    ValidationOutputExists,
				Message: fmt.Sprintf("output file already exists: %s", outPath),
				TitleId: &titleId,
			})
			hasIssue = true

			continue
		}

		if !errors.Is(err, fs.ErrNotExist) {
			titleId := title.TitleId
			report.AddResult(ValidationResult{
				Status:  ValidationStatusFail,
				Code:    ValidationOutputDirInvalid,
				Message: fmt.Sprintf("could not stat output file path: %s", outPath),
				Cause:   err,
				TitleId: &titleId,
			})
			hasIssue = true

			continue
		}
	}

	if !hasIssue {
		report.AddResult(ValidationResult{
			Status:  ValidationStatusPass,
			Message: "No existing file conflicts",
		})
	}
}

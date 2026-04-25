package engine

import (
	"fmt"
	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/planner"
	"os"
	"path/filepath"
)

type ValidationReport struct {
	Errors   []ValidationIssue
	Warnings []ValidationIssue
}

func (r *ValidationReport) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *ValidationReport) AddError(issue ValidationIssue) {
	r.Errors = append(r.Errors, issue)
}

func (r *ValidationReport) AddWarning(issue ValidationIssue) {
	r.Warnings = append(r.Warnings, issue)
}

type ValidationIssue struct {
	Code    ValidationCode
	Message string
	Cause   error
	TitleId *int
}

type ValidationCode string

const (
	ValidationInsufficientSpace ValidationCode = "insufficient_space"
	ValidationOutputExists ValidationCode = "output_exists"
	ValidationOutputDirInvalid ValidationCode = "output_dir_invalid"
)

func ValidatePlan(plan planner.DiscPlan) ValidationReport {
	report := &ValidationReport{
		Errors:   make([]ValidationIssue, 0),
		Warnings: make([]ValidationIssue, 0),
	}

	validateOutputDir(plan, report)
	validateDiskSpace(plan, report)
	validateExistingFiles(plan, report)

	return *report
}

func validateOutputDir(plan planner.DiscPlan, report *ValidationReport) {
	info, err := os.Stat(plan.OutputDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Making this an error for now. I might just auto-create the outputdir in the future
			report.AddError(ValidationIssue{
				Code: ValidationOutputDirInvalid,
				Message: fmt.Sprintf("output directory does not exist: %s", plan.OutputDir),
				Cause: err,
			})

			return
		}
		report.AddError(ValidationIssue{
			Code: ValidationOutputDirInvalid,
			Message: fmt.Sprintf("could not stat output directory: %s", plan.OutputDir),
			Cause: err,
		})

		return
	}

	if !info.IsDir() {
		report.AddError(ValidationIssue{
			Code: ValidationOutputDirInvalid,
			Message: fmt.Sprintf("output path is not a directory: %s", plan.OutputDir),
		})
	}
}

func validateDiskSpace(plan planner.DiscPlan, report *ValidationReport) {
	free, err := files.GetFreeDiskSpace(plan.OutputDir)
	if err != nil {
		report.AddError(ValidationIssue{
			Code: ValidationOutputDirInvalid,
			Message: fmt.Sprintf("could not determine free space for output directory: %s", plan.OutputDir),
			Cause: err,
		})
	}

	var required uint64
	for _, title := range plan.Titles {
		required += uint64(title.EstimatedSize)
	}

	if free < required {
		report.AddError(ValidationIssue{
			Code: ValidationInsufficientSpace,
			Message: fmt.Sprintf(
				"not enough free space in %s: need %d bytes, have %d bytes",
				plan.OutputDir,
				required,
				free),
		})
	}
}

func validateExistingFiles(plan planner.DiscPlan, report *ValidationReport) {
	for _, title := range plan.Titles {
		outPath := filepath.Join(plan.OutputDir, title.FinalName)

		_, err := os.Stat(outPath)
		if err == nil {
			titleId := title.TitleId
			report.AddError(ValidationIssue{
				Code: ValidationOutputExists,
				Message: fmt.Sprintf("output file already exists: %s", outPath),
				TitleId: &titleId,
			})

			continue
		}

		if !os.IsNotExist(err) {
			titleId := title.TitleId
			report.AddError(ValidationIssue{
				Code: ValidationOutputDirInvalid,
				Message: fmt.Sprintf("could not stat output file path: %s", outPath),
				Cause: err,
				TitleId: &titleId,
			})
		}
	}
}

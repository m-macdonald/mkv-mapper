package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/planner"
	"m-macdonald/mkv-mapper/internal/util"
)

func buildValidationReport(plan planner.SelectedPlan) planner.ValidationReport {
	report := &planner.ValidationReport{}

	validateOutputDir(plan, report)
	validateDiskSpace(plan, report)
	validateExistingFiles(plan, report)

	return *report
}

func validateOutputDir(plan planner.SelectedPlan, report *planner.ValidationReport) {
	info, err := os.Stat(plan.OutputDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Making this an error for now. I might just auto-create the outputdir in the future
			report.AddResult(planner.ValidationResult{
				Status:  planner.ValidationStatusFail,
				Code:    planner.ValidationOutputDirInvalid,
				Message: fmt.Sprintf("output directory does not exist: %s", plan.OutputDir),
				Cause:   err,
			})

			return
		}
		report.AddResult(planner.ValidationResult{
			Status:  planner.ValidationStatusFail,
			Code:    planner.ValidationOutputDirInvalid,
			Message: fmt.Sprintf("could not stat output directory: %s", plan.OutputDir),
			Cause:   err,
		})

		return
	}

	if !info.IsDir() {
		report.AddResult(planner.ValidationResult{
			Status:  planner.ValidationStatusFail,
			Code:    planner.ValidationOutputDirInvalid,
			Message: fmt.Sprintf("output path is not a directory: %s", plan.OutputDir),
		})

		return
	}

	report.AddResult(planner.ValidationResult{
		Status:  planner.ValidationStatusPass,
		Message: fmt.Sprintf("output directory valid: %s", plan.OutputDir),
	})
}

func validateDiskSpace(plan planner.SelectedPlan, report *planner.ValidationReport) {
	free, err := files.GetFreeDiskSpace(plan.OutputDir)
	if err != nil {
		report.AddResult(planner.ValidationResult{
			Status:  planner.ValidationStatusFail,
			Code:    planner.ValidationOutputDirInvalid,
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
		report.AddResult(planner.ValidationResult{
			Status: planner.ValidationStatusFail,
			Code:   planner.ValidationInsufficientSpace,
			Message: fmt.Sprintf(
				"insufficient disk space %s: need %s, have %s",
				plan.OutputDir,
				util.FormatSize(required),
				util.FormatSize(free)),
		})

		return
	}

	report.AddResult(planner.ValidationResult{
		Status: planner.ValidationStatusPass,
		Message: fmt.Sprintf(
			"sufficient disk space %s: need %s, have %s",
			plan.OutputDir,
			util.FormatSize(required),
			util.FormatSize(free)),
	})
}

func validateExistingFiles(plan planner.SelectedPlan, report *planner.ValidationReport) {
	hasIssue := false
	for _, title := range plan.Titles {
		outPath := filepath.Join(plan.OutputDir, title.FinalName)

		_, err := os.Stat(outPath)
		if err == nil {
			titleId := title.TitleId
			report.AddResult(planner.ValidationResult{
				Status:  planner.ValidationStatusFail,
				Code:    planner.ValidationOutputExists,
				Message: fmt.Sprintf("output file already exists: %s", outPath),
				TitleId: &titleId,
			})
			hasIssue = true

			continue
		}

		if !errors.Is(err, fs.ErrNotExist) {
			titleId := title.TitleId
			report.AddResult(planner.ValidationResult{
				Status:  planner.ValidationStatusFail,
				Code:    planner.ValidationOutputDirInvalid,
				Message: fmt.Sprintf("could not stat output file path: %s", outPath),
				Cause:   err,
				TitleId: &titleId,
			})
			hasIssue = true

			continue
		}
	}

	if !hasIssue {
		report.AddResult(planner.ValidationResult{
			Status:  planner.ValidationStatusPass,
			Message: "No existing file conflicts",
		})
	}
}

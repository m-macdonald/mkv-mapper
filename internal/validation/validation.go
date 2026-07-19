package validation

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/util"
)

type CheckGroupLabel string

const (
	RipLabel    CheckGroupLabel = "Rip"
	BackupLabel CheckGroupLabel = "Backup"
)

type Check func(ctx context.Context) []Result

type CheckGroup struct {
	Label  CheckGroupLabel
	Checks []Check
}

type Report struct {
	ResultsByGroup map[CheckGroupLabel][]Result
}

func (v *Report) Passes() []Result {
	return v.filter(StatusPass)
}

func (v *Report) Warnings() []Result {
	return v.filter(StatusWarn)
}

func (v *Report) Errors() []Result {
	return v.filter(StatusFail)
}

func (v *Report) HasErrors() bool {
	return len(v.Errors()) > 0
}

func (v *Report) filter(status Status) []Result {
	var results []Result
	for _, results := range v.ResultsByGroup {
		for _, result := range results {
			if result.Status == status {
				results = append(results, result)
			}
		}
	}
	return results
}

type Result struct {
	Status  Status
	Code    Code
	Message string
	Cause   error
	RefID   string
}

type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

type Code string

const (
	InsufficientSpace Code = "insufficient_space"
	OutputExists      Code = "output_exists"
	OutputDirInvalid  Code = "output_dir_invalid"
)

func Run(ctx context.Context, checkGroups []CheckGroup) Report {
	report := Report{
		ResultsByGroup: make(map[CheckGroupLabel][]Result, len(checkGroups)),
	}
	for _, checkGroup := range checkGroups {
		var results []Result
		for _, check := range checkGroup.Checks {
			results = append(results, check(ctx)...)
		}
		report.ResultsByGroup[checkGroup.Label] = append(
			report.ResultsByGroup[checkGroup.Label],
			results...,
		)
	}
	return report
}

func OutputDirValid(dir string) Check {
	return func(ctx context.Context) []Result {
		var results []Result
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				// Making this an error for now. I might just auto-create the outputdir in the future
				results = append(results, Result{
					Status:  StatusFail,
					Code:    OutputDirInvalid,
					Message: fmt.Sprintf("output directory does not exist: %s", dir),
					Cause:   err,
				})
			} else {
				results = append(results, Result{
					Status:  StatusFail,
					Code:    OutputDirInvalid,
					Message: fmt.Sprintf("could not stat output directory: %s", dir),
					Cause:   err,
				})
			}
			return results
		}

		if !info.IsDir() {
			return append(results, Result{
				Status:  StatusFail,
				Code:    OutputDirInvalid,
				Message: fmt.Sprintf("output path is not a directory: %s", dir),
			})
		}

		return append(results, Result{
			Status:  StatusPass,
			Message: fmt.Sprintf("output directory valid: %s", dir),
		})
	}
}

func FreeSpace(dir string, needed uint64) Check {
	return func(ctx context.Context) []Result {
		free, err := files.GetFreeDiskSpace(dir)
		if err != nil {
			return []Result{
				{
					Status:  StatusFail,
					Code:    OutputDirInvalid,
					Message: fmt.Sprintf("could not determine free space for output directory: %s", dir),
					Cause:   err,
				},
			}
		}

		if free < needed {
			return []Result{
				{
					Status: StatusFail,
					Code:   InsufficientSpace,
					Message: fmt.Sprintf(
						"insufficient disk space %s: need %s, have %s",
						dir,
						util.FormatSize(needed),
						util.FormatSize(free)),
				},
			}
		}

		return []Result{
			{
				Status: StatusPass,
				Message: fmt.Sprintf(
					"sufficient disk space %s: need %s, have %s",
					dir,
					util.FormatSize(needed),
					util.FormatSize(free)),
			},
		}
	}
}

type FilenameTarget struct {
	ID       string
	FileName string
}

func NoConflicts(dir string, targets []FilenameTarget) Check {
	return func(ctx context.Context) []Result {
		var results []Result
		for _, target := range targets {
			outPath := filepath.Join(dir, target.FileName)

			_, err := os.Stat(outPath)
			if err == nil {
				results = append(results, Result{
					Status:  StatusFail,
					Code:    OutputExists,
					Message: fmt.Sprintf("output file already exists: %s", outPath),
					RefID:   target.ID,
				})

				continue
			}

			if !errors.Is(err, fs.ErrNotExist) {
				results = append(results, Result{
					Status:  StatusFail,
					Code:    OutputDirInvalid,
					Message: fmt.Sprintf("could not stat output file path: %s", outPath),
					Cause:   err,
					RefID:   target.ID,
				})

				continue
			}
		}

		if len(results) == 0 {
			results = append(results, Result{
				Status:  StatusPass,
				Message: "No existing file conflicts",
			})
		}

		return results
	}
}

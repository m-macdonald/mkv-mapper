package terminal

import (
	"fmt"
	"io"

	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/util"
	"m-macdonald/mkv-mapper/internal/validation"
)

type BackupSummaryRenderer struct {
	out io.Writer
}

func NewBackupSummaryRenderer(out io.Writer) BackupSummaryRenderer {
	return BackupSummaryRenderer{
		out: out,
	}
}

func (b *BackupSummaryRenderer) Render(plan model.ValidatedBackupPlan) error {
	_, err := fmt.Fprintf(b.out, "Backup:\n %s → %s (~%s)\n\n",
		plan.Label,
		plan.OutputDir,
		util.FormatSize(plan.SumTitleSizes()),
	)
	if err != nil {
		return err
	}

	results, ok := plan.Report.ResultsByGroup[validation.BackupLabel]
	if !ok {
		return nil
	}
	return renderCheckResults(b.out, validation.BackupLabel, results)
}

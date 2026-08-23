package terminal

import (
	"fmt"
	"io"

	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/preview"

	"github.com/pterm/pterm"
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
	view := preview.BuildBackupPlanView(plan)

	if err := b.renderHeader(view); err != nil {
		return err
	}

	if err := b.renderValidation(view); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(b.out); err != nil {
		return err
	}
	return nil
}

func (b *BackupSummaryRenderer) renderHeader(view preview.BackupPlanView) error {
	line := pterm.NewStyle(pterm.Bold).Sprintf("Backup: %s → %s (~%s)", view.Label, view.OutputDir, view.Size)
	_, err := fmt.Fprintf(b.out, "%s\n\n", line)
	return err
}

func (b *BackupSummaryRenderer) renderValidation(view preview.BackupPlanView) error {
	for _, group := range view.CheckGroups {
		if err := renderCheckResults(b.out, group.Label, group.Results); err != nil {
			return err
		}
	}

	return nil
}

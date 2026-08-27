package terminal

import (
	"fmt"
	"io"

	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/preview"

	"github.com/pterm/pterm"
)

type BackupPreviewRenderer struct {
	out io.Writer
}

func NewBackupPreviewRenderer(out io.Writer) BackupPreviewRenderer {
	return BackupPreviewRenderer{
		out: out,
	}
}

func (b *BackupPreviewRenderer) Render(plan model.ValidatedBackupPlan) error {
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

func (b *BackupPreviewRenderer) renderHeader(view preview.BackupPlanView) error {
	line := pterm.NewStyle(pterm.Bold).Sprintf("Backup: %s → %s (~%s)", view.Label, view.OutputDir, view.Size)
	_, err := fmt.Fprintf(b.out, "%s\n\n", line)
	return err
}

func (b *BackupPreviewRenderer) renderValidation(view preview.BackupPlanView) error {
	for _, group := range view.CheckGroups {
		if err := renderCheckResults(b.out, group.Label, group.Results); err != nil {
			return err
		}
	}

	return nil
}

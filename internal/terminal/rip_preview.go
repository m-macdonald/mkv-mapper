package terminal

import (
	"fmt"
	"io"

	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/preview"
	"m-macdonald/mkv-mapper/internal/util"

	"github.com/pterm/pterm"
)

type RipPreviewRenderer struct {
	out io.Writer
}

func NewRipPreviewRenderer(out io.Writer) RipPreviewRenderer {
	return RipPreviewRenderer{
		out: out,
	}
}

func (r *RipPreviewRenderer) Render(plan model.ValidatedRipPlan) error {
	view := preview.BuildRipPlanView(plan)

	if err := r.renderHeader(view); err != nil {
		return err
	}
	if err := r.renderTitles(view); err != nil {
		return err
	}
	if err := r.renderValidation(view); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(r.out); err != nil {
		return err
	}
	return nil
}

func (r *RipPreviewRenderer) renderHeader(view preview.RipPlanView) error {
	title := pterm.NewStyle(pterm.Bold).Sprintf("%s (%d) - %s", view.DiscName, view.Year, view.Format)
	hash := pterm.NewStyle(pterm.FgGray).Sprintf("Hash: %s", view.Hash)

	_, err := fmt.Fprintf(r.out, "%s\n%s\n\n", title, hash)
	return err
}

func (r *RipPreviewRenderer) renderTitles(view preview.RipPlanView) error {
	if _, err := fmt.Fprintf(r.out, "Titles:\n"); err != nil {
		return err
	}

	for _, t := range view.Matched {
		if err := r.renderTitle(t); err != nil {
			return err
		}
	}

	return r.renderUnmatched(view)
}

func (r *RipPreviewRenderer) renderTitle(titleView preview.TitleRipPlanView) error {
	_, err := fmt.Fprintf(
		r.out,
		"  %s → %s (%s)\n",
		titleView.Source,
		titleView.Target,
		util.FormatSize(titleView.Size),
	)
	if err != nil {
		return err
	}

	for _, note := range titleView.Notes {
		symbol := getValidationSymbol(note.Status)
		if _, err := fmt.Fprintf(r.out, "	%s %s\n", symbol, note.Message); err != nil {
			return err
		}
	}
	return nil
}

func (r *RipPreviewRenderer) renderValidation(view preview.RipPlanView) error {
	for _, group := range view.CheckGroups {
		if err := renderCheckResults(r.out, group.Label, group.Results); err != nil {
			return err
		}
	}

	return nil
}

func (r *RipPreviewRenderer) renderUnmatched(view preview.RipPlanView) error {
	if len(view.Unmatched) == 0 {
		return nil
	}

	label := pterm.NewStyle(pterm.FgGray, pterm.Italic).Sprintf(
		" %d titles unmatched (no TheDiscDb match, %s each)\n",
		len(view.Unmatched),
		view.UnmatchedRange,
	)

	_, err := fmt.Fprintf(r.out, label)
	return err
}

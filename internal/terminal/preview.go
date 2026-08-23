package terminal

import (
	"fmt"
	"io"

	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/preview"
	"m-macdonald/mkv-mapper/internal/util"

	"github.com/pterm/pterm"
)

type PreviewRenderer struct {
	out io.Writer
}

func NewPreviewRenderer(out io.Writer) PreviewRenderer {
	return PreviewRenderer{
		out: out,
	}
}

func (p *PreviewRenderer) Render(plan model.ValidatedPlan) error {
	view := preview.BuildPlanView(plan)

	if err := p.renderHeader(view); err != nil {
		return err
	}
	if err := p.renderTitles(view); err != nil {
		return err
	}
	if err := p.renderValidation(view); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(p.out); err != nil {
		return err
	}
	return nil
}

func (p *PreviewRenderer) renderHeader(view preview.PlanView) error {
	title := pterm.NewStyle(pterm.Bold).Sprintf("%s (%d) - %s", view.DiscName, view.Year, view.Format)
	hash := pterm.NewStyle(pterm.FgGray).Sprintf("Hash: %s", view.Hash)

	_, err := fmt.Fprintf(p.out, "%s\n%s\n\n", title, hash)
	return err
}

func (p *PreviewRenderer) renderTitles(view preview.PlanView) error {
	if _, err := fmt.Fprintf(p.out, "Titles:\n"); err != nil {
		return err
	}

	for _, t := range view.Matched {
		if err := p.renderTitle(t); err != nil {
			return err
		}
	}

	return p.renderUnmatched(view)
}

func (p *PreviewRenderer) renderTitle(titleView preview.TitleView) error {
	_, err := fmt.Fprintf(
		p.out,
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
		if _, err := fmt.Fprintf(p.out, "	%s %s\n", symbol, note.Message); err != nil {
			return err
		}
	}
	return nil
}

func (p *PreviewRenderer) renderValidation(view preview.PlanView) error {
	for _, group := range view.CheckGroups {
		if err := renderCheckResults(p.out, group.Label, group.Results); err != nil {
			return err
		}
	}

	return nil
}

func (p *PreviewRenderer) renderUnmatched(view preview.PlanView) error {
	if len(view.Unmatched) == 0 {
		return nil
	}

	label := pterm.NewStyle(pterm.FgGray, pterm.Italic).Sprintf(
		" %d titles unmatched (no TheDiscDb match, %s each)\n",
		len(view.Unmatched),
		view.UnmatchedRange,
	)

	_, err := fmt.Fprintf(p.out, label)
	return err
}

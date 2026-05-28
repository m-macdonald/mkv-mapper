package display

import (
	"fmt"
	"io"

	"m-macdonald/mkv-mapper/internal/engine"
	"m-macdonald/mkv-mapper/internal/format"
	"m-macdonald/mkv-mapper/internal/planner"
)

type PreviewRenderer struct {
	out io.Writer
}

func NewPreviewRenderer(out io.Writer) PreviewRenderer {
	return PreviewRenderer{
		out: out,
	}
}

func (p *PreviewRenderer) Render(plan *engine.ValidatedPlan) error {
	if err := p.renderHeader(plan); err != nil {
		return err
	}
	if err := p.renderTitles(plan); err != nil {
		return err
	}
	if err := p.renderValidation(plan); err != nil {
		return err
	}
	return nil
}

func (p *PreviewRenderer) renderHeader(plan *engine.ValidatedPlan) error {
	_, err := fmt.Fprintf(p.out, "%s (%d) — %s\nHash: %s\n\n",
		plan.DiscPlan.MediaTitle,
		plan.DiscPlan.MediaYear,
		plan.DiscPlan.DiscFormat,
		plan.DiscPlan.DiscHash,
	)
	return err
}

func (p *PreviewRenderer) renderTitles(plan *engine.ValidatedPlan) error {
	warningsByTitle := indexByTitleId(
		plan.BuildReport.Warnings,
		func(w planner.PlanWarning) *int { return &w.TitleId })
	validationsByTitle := indexByTitleId(
		plan.ValidationReport.Results,
		func(v engine.ValidationResult) *int { return v.TitleId })
	if _, err := fmt.Fprintln(p.out, "Titles:"); err != nil {
		return err
	}
	for _, title := range plan.DiscPlan.Titles {
		_, err := fmt.Fprintf(p.out, "  %s → %s (%s)\n",
			title.MakeMkvOutputFile,
			title.FinalName,
			format.Size(title.EstimatedSize),
		)
		if err != nil {
			return err
		}

		for _, warning := range warningsByTitle[title.TitleId] {
			if _, err = fmt.Fprintf(p.out, "    ⚠ %s\n", warning.Message); err != nil {
				return err
			}
		}
		for _, validation := range validationsByTitle[title.TitleId] {
			symbol := getValidationSymbol(validation.Status)
			if _, err = fmt.Fprintf(p.out, "    %s %s\n", symbol, validation.Message); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *PreviewRenderer) renderValidation(plan *engine.ValidatedPlan) error {
	var discValidations []engine.ValidationResult
	for _, validation := range plan.ValidationReport.Results {
		// Validations that are not associated with a title id are considered "disc-level"
		if validation.TitleId == nil {
			discValidations = append(discValidations, validation)
		}
	}
	if len(discValidations) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(r.out, "\nValidation:"); err != nil {
		return err
	}
	for _, result := range discValidations {
		symbol := getValidationSymbol(result.Status)
		_, err := fmt.Fprintf(r.out, "  %s %s\n", symbol, result.Message)
		if err != nil {
			return err
		}
	}
	return nil
}

func indexByTitleId[T any](items []T, getId func(T) *int) map[int][]T {
	index := map[int][]T{}
	for _, item := range items {
		id := getId(item)
		if id != nil {
			index[*id] = append(index[*id], item)
		}
	}
	return index
}

func getValidationSymbol(status engine.ValidationStatus) string {
	switch status {
	case engine.ValidationStatusPass:
		return "✓"
	case engine.ValidationStatusWarn:
		return "⚠"
	case engine.ValidationStatusFail:
		return "✗"
	}
	return ""
}

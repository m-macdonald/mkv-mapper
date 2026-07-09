package terminal

import (
	"fmt"
	"io"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/util"
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

func (p *PreviewRenderer) renderHeader(plan model.ValidatedPlan) error {
	_, err := fmt.Fprintf(p.out, "%s (%d) — %s\nHash: %s\n\n",
		plan.MediaInfo.Title,
		plan.MediaInfo.Year,
		plan.Disc.Format,
		plan.Disc.Hash,
	)
	return err
}

func (p *PreviewRenderer) renderTitles(plan model.ValidatedPlan) error {
	warningsByTitle := indexByTitleId(
		plan.BuildReport.Warnings,
		func(w model.PlanWarning) *lines.TitleId { return &w.TitleId })
	validationsByTitle := indexByTitleId(
		plan.ValidationReport.Results,
		func(v model.ValidationResult) *lines.TitleId { return v.TitleId })
	if _, err := fmt.Fprintln(p.out, "Titles:"); err != nil {
		return err
	}
	for _, title := range plan.Titles {
		_, err := fmt.Fprintf(p.out, "  %s → %s (%s)\n",
			title.MakeMkvOutputFile,
			title.FinalName,
			util.FormatSize(title.EstimatedSize),
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

func (r *PreviewRenderer) renderValidation(plan model.ValidatedPlan) error {
	var discValidations []model.ValidationResult
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

func indexByTitleId[T any](items []T, getId func(T) *lines.TitleId) map[lines.TitleId][]T {
	index := map[lines.TitleId][]T{}
	for _, item := range items {
		id := getId(item)
		if id != nil {
			index[*id] = append(index[*id], item)
		}
	}
	return index
}

func getValidationSymbol(status model.ValidationStatus) string {
	switch status {
	case model.ValidationStatusPass:
		return "✓"
	case model.ValidationStatusWarn:
		return "⚠"
	case model.ValidationStatusFail:
		return "✗"
	}
	return ""
}

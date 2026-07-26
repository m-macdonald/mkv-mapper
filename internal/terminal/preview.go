package terminal

import (
	"fmt"
	"io"
	"strconv"

	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/util"
	"m-macdonald/mkv-mapper/internal/validation"
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
		func(w model.PlanWarning) *string { s := strconv.Itoa(int(w.TitleId)); return &s })
	validationsByTitle := indexByTitleId(
		plan.ValidationReport.ResultsByGroup[validation.RipLabel],
		func(v validation.Result) *string { return &v.RefID })
	if _, err := fmt.Fprintln(p.out, "Titles:"); err != nil {
		return err
	}
	for _, title := range plan.Titles {
		_, err := fmt.Fprintf(p.out, "  %s → %s (%s)\n",
			title.SourcePlaylist,
			title.FinalName,
			util.FormatSize(title.EstimatedSize),
		)
		if err != nil {
			return err
		}

		titleId := strconv.Itoa(int(title.TitleId))
		for _, warning := range warningsByTitle[titleId] {
			if _, err = fmt.Fprintf(p.out, "    ⚠ %s\n", warning.Message); err != nil {
				return err
			}
		}
		for _, validation := range validationsByTitle[titleId] {
			symbol := getValidationSymbol(validation.Status)
			if _, err = fmt.Fprintf(p.out, "    %s %s\n", symbol, validation.Message); err != nil {
				return err
			}
		}
	}
	return nil
}

var groupOrder = []validation.CheckGroupLabel{validation.RipLabel, validation.BackupLabel}

func (r *PreviewRenderer) renderValidation(plan model.ValidatedPlan) error {
	for _, label := range groupOrder {
		results, ok := plan.ValidationReport.ResultsByGroup[label]
		if !ok {
			continue
		}

		var discValidations []validation.Result
		for _, result := range results {
			if result.RefID == "" {
				discValidations = append(discValidations, result)
			}
		}
		if len(discValidations) == 0 {
			continue
		}

		if _, err := fmt.Fprintf(r.out, "\n%s:\n", label); err != nil {
			return err
		}
		for _, result := range discValidations {
			symbol := getValidationSymbol(result.Status)
			if _, err := fmt.Fprintf(r.out, " %s %s\n", symbol, result.Message); err != nil {
				return err
			}
		}
	}
	return nil
}

func indexByTitleId[T any](items []T, getId func(T) *string) map[string][]T {
	index := map[string][]T{}
	for _, item := range items {
		id := getId(item)
		if id != nil {
			index[*id] = append(index[*id], item)
		}
	}
	return index
}

func getValidationSymbol(status validation.Status) string {
	switch status {
	case validation.StatusPass:
		return "✓"
	case validation.StatusWarn:
		return "⚠"
	case validation.StatusFail:
		return "✗"
	}
	return ""
}

package terminal

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/util"

	"github.com/pterm/pterm"
)

type Selector struct{}

func NewSelector() Selector {
	return Selector{}
}

func (Selector) Select(plan model.Plan) ([]lines.TitleId, error) {
	if len(plan.Titles) == 0 {
		return nil, fmt.Errorf("no titles available to select from")
	}

	options := newTitleOptions(plan.Titles)

	selected, err := pterm.DefaultInteractiveMultiselect.
		WithOptions(options.labels()).
		WithDefaultText(fmt.Sprintf("Select titles to rip - %d total", len(plan.Titles))).
		WithMaxHeight(15).
		Show()
	if err != nil {
		return nil, fmt.Errorf("title selection: %w", err)
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("at least one title must be selected")
	}

	ids := make([]lines.TitleId, 0, len(selected))
	for _, label := range selected {
		if id, ok := options.idFor(label); ok {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

func formatTitleLabel(title model.TitlePlan) string {
	matched := ""
	if title.IsMatched {
		matched = " [matched]"
	}
	// This label is safe to key off of only because SourcePlaylist and FinalName are both unique. Bear this in mind when making any future changes
	return fmt.Sprintf("%s: %s (%s / %s)%s", title.SourcePlaylist, title.FinalName, title.Duration, util.FormatSize(title.EstimatedSize), matched)
}

type titleOption struct {
	id    lines.TitleId
	label string
}

type titleOptions []titleOption

func newTitleOptions(titles []model.TitlePlan) titleOptions {
	opts := make(titleOptions, 0, len(titles))
	for _, title := range titles {
		opts = append(opts, titleOption{id: title.TitleId, label: formatTitleLabel(title)})
	}
	return opts
}

func (opts titleOptions) labels() []string {
	labels := make([]string, len(opts))
	for i, opt := range opts {
		labels[i] = opt.label
	}
	return labels
}

func (opts titleOptions) idFor(label string) (lines.TitleId, bool) {
	for _, opt := range opts {
		if opt.label == label {
			return opt.id, true
		}
	}
	return 0, false
}

package model

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"
)

type DiscIdentity struct {
	DiscRoot string
	Label    string
}

type RipPlanBase struct {
	DiscIdentity
	MediaInfo MediaInfo
	DiscInfo  DiscInfo
	Format    string
	OutputDir string
	Titles    []TitleRipPlan
}

func (p RipPlanBase) SumTitleSizes() uint64 {
	var total uint64
	for _, title := range p.Titles {
		total += title.EstimatedSize
	}
	return total
}

func (p RipPlanBase) Intents() []TitleIntent {
	intents := make([]TitleIntent, 0, len(p.Titles))
	for _, title := range p.Titles {
		intents = append(intents, TitleIntent{
			Identity:  title.Identity,
			FinalName: title.FinalName,
		})
	}
	return intents
}

type MediaInfo struct {
	Title string
	Year  int
}

type DiscInfo struct {
	Hash   string
	Format string
}

type BuildReport struct {
	Warnings []PlanWarning
}

type PlanWarning struct {
	TitleId lines.TitleId
	Code    WarningCode
	Message string
	Cause   error
}

type WarningCode string

const (
	WarningNoMetadata WarningCode = "no_metadata"
)

type RipPlan struct {
	RipPlanBase
	BuildReport BuildReport
	// Indicates if the Plan contains all of the titles that makemkv discovered
	IsAllTitles bool
}

func NewRipPlan(base RipPlanBase, report BuildReport) RipPlan {
	return RipPlan{
		RipPlanBase: base,
		BuildReport: report,
		// Newly constructed plans always have every title
		IsAllTitles: true,
	}
}

func (p RipPlan) ApplySelection(selection Selection) RipPlan {
	updatedBase := p.RipPlanBase
	updatedBase.Titles = selection.Selected

	selected := NewRipPlan(updatedBase, p.BuildReport)
	selected.IsAllTitles = len(p.Titles) == len(selection.Selected)

	return selected
}

type TitleRipPlan struct {
	TitleId           lines.TitleId
	Identity          TitleIdentity
	MakeMkvOutputFile string
	FinalName         string
	EstimatedSize     uint64
	Duration          string
	IsMatched         bool // Indicates if this title had a matching Item definition in TheDiscDb
}

type TitleIntent struct {
	Identity  TitleIdentity
	FinalName string
}

func (p RipPlan) MergeIntents(intents []TitleIntent) (RipPlan, error) {
	byIdentityName := groupTitleIntentByIdentityName(intents)

	titles := make([]TitleRipPlan, 0, len(intents))
	for _, tp := range p.Titles {
		intent, ok := findMatchingIntent(tp.Identity, byIdentityName)
		if !ok {
			continue
		}
		tp.FinalName = intent.FinalName
		titles = append(titles, tp)
	}

	if len(titles) != len(intents) {
		return RipPlan{}, fmt.Errorf("only matched %d of %d selected titles when re-scanning", len(titles), len(intents))
	}

	return p.ApplySelection(Selection{Selected: titles}), nil
}

// Indexes each intent under every name in its identity
// so that a match can be found by looking it up under any name that it's known by.
func groupTitleIntentByIdentityName(intents []TitleIntent) map[SourceFilename]TitleIntent {
	index := make(map[SourceFilename]TitleIntent)
	for _, intent := range intents {
		for _, name := range intent.Identity.Names() {
			index[name] = intent
		}
	}
	return index
}

func findMatchingIntent(id TitleIdentity, index map[SourceFilename]TitleIntent) (TitleIntent, bool) {
	for _, name := range id.Names() {
		if intent, ok := index[name]; ok {
			return intent, true
		}
	}
	return TitleIntent{}, false
}

package model

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/signature"
)

type DiscIdentity struct {
	DiscRoot string
	Label    string
}

type PlanBase struct {
	DiscIdentity
	MediaInfo MediaInfo
	DiscInfo  DiscInfo
	Format    string
	OutputDir string
	Titles    []TitlePlan
}

func (p PlanBase) SumTitleSizes() uint64 {
	var total uint64
	for _, title := range p.Titles {
		total += title.EstimatedSize
	}
	return total
}

func (p PlanBase) Intents() []TitleIntent {
	intents := make([]TitleIntent, 0, len(p.Titles))
	for _, title := range p.Titles {
		intents = append(intents, TitleIntent{
			Signature: title.SegmentSignature,
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

type Plan struct {
	PlanBase
	BuildReport BuildReport
	// Indicates if the Plan contains all of the titles that makemkv discovered
	IsAllTitles bool
}

func NewPlan(base PlanBase, report BuildReport) Plan {
	return Plan{
		PlanBase:    base,
		BuildReport: report,
		// Newly constructed plans always have every title
		IsAllTitles: true,
	}
}

func (p Plan) ApplySelection(selection Selection) Plan {
	updatedBase := p.PlanBase
	updatedBase.Titles = selection.Selected

	selected := NewPlan(updatedBase, p.BuildReport)
	selected.IsAllTitles = len(p.Titles) == len(selection.Selected)

	return selected
}

type TitlePlan struct {
	TitleId           lines.TitleId
	SourcePlaylist    string
	SegmentSignature  signature.SegmentSignature
	MakeMkvOutputFile string
	FinalName         string
	EstimatedSize     uint64
	Duration          string
	IsMatched         bool // Indicates if this title had a matching Item definition in TheDiscDb
}

type TitleIntent struct {
	Signature signature.SegmentSignature
	FinalName string
}

func (p Plan) MergeIntents(intents []TitleIntent) (Plan, error) {
	bySignature := groupTitleIntentBySignature(intents)
	matched := filterTitles(p.Titles, func(tp TitlePlan) bool {
		_, ok := bySignature[tp.SegmentSignature]
		return ok
	})
	if len(matched) != len(intents) {
		return Plan{}, fmt.Errorf("only matched %d of %d selected titles when re-scanning", len(matched), len(intents))
	}
	titles := make([]TitlePlan, 0, len(matched))
	for _, t := range matched {
		t.FinalName = bySignature[t.SegmentSignature].FinalName
		titles = append(titles, t)
	}
	return p.ApplySelection(Selection{Selected: titles}), nil
}

func groupTitleIntentBySignature(intents []TitleIntent) map[signature.SegmentSignature]TitleIntent {
	bySignature := make(map[signature.SegmentSignature]TitleIntent, len(intents))
	for _, intent := range intents {
		bySignature[intent.Signature] = intent
	}
	return bySignature
}

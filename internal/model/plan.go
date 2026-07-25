package model

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/signature"
)

type PlanBase struct {
	Disc      Disc
	DiscRoot  string
	MediaInfo MediaInfo
	OutputDir string
	Titles    []TitlePlan

	Backup     bool
	BackupDir  string
	KeepBackup bool
}

func (p PlanBase) SumTitleSizes() uint64 {
	var total uint64
	for _, title := range p.Titles {
		total += title.EstimatedSize
	}
	return total
}

type Disc struct {
	Label  string
	Format string
	Hash   string
}

type MediaInfo struct {
	Title string
	Year  int
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

func (sp SelectedPlan) Intents() []TitleIntent {
	intents := make([]TitleIntent, 0, len(sp.Titles))
	for _, title := range sp.Titles {
		intents = append(intents, TitleIntent{
			Signature: title.SegmentSignature,
			FinalName: title.FinalName,
		})
	}
	return intents
}

func (p Plan) MergeIntents(intents []TitleIntent) (SelectedPlan, error) {
	bySignature := groupTitleIntentBySignature(intents)
	matched := filterTitles(p.Titles, func(tp TitlePlan) bool {
		_, ok := bySignature[tp.SegmentSignature]
		return ok
	})
	if len(matched) != len(intents) {
		return SelectedPlan{}, fmt.Errorf("only matched %d of %d selected titles when re-scanning", len(matched), len(intents))
	}
	titles := make([]TitlePlan, len(matched))
	for _, t := range matched {
		t.FinalName = bySignature[t.SegmentSignature].FinalName
		titles = append(titles, t)
	}
	return NewSelectedPlan(p, Selection{Selected: titles})
}

func groupTitleIntentBySignature(intents []TitleIntent) map[signature.SegmentSignature]TitleIntent {
	bySignature := make(map[signature.SegmentSignature]TitleIntent, len(intents))
	for _, intent := range intents {
		bySignature[intent.Signature] = intent
	}
	return bySignature
}

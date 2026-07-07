package model 

import (
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/signature"
)

type PlanBase struct {
	Disc      Disc
	DiscRoot  string
	MediaInfo MediaInfo
	OutputDir string
	Titles    []TitlePlan
}

type Disc struct {
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

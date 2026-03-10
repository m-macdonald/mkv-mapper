package planner

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/signature"
)

type DiscPlan struct {
	DiscHash  string
	DiscRoot  string
	OutputDir string
	Titles    []TitlePlan
}

type TitlePlan struct {
	TitleId           int
	SourcePlaylist    string
	SegmentSignature  signature.SegmentSignature
	MakeMkvOutputFile string
	FinalName         string
}

func BuildPlan(discRoot string, outputDir string, disc *discdb.Disc, titles map[signature.SegmentSignature]string) (*DiscPlan, error) {
	mappings, err := mapper.MapTitles(disc, titles)
	if err != nil {
		return nil, fmt.Errorf("failed to map MakeMkv titles to DiscDB titles %w", err)
	}

	plan := &DiscPlan{
		DiscRoot:  discRoot,
		OutputDir: outputDir,
	}

	for mkvFile, title := range mappings {
		plan.Titles = append(plan.Titles, TitlePlan{
			SourcePlaylist:    title.SourceFile,
			MakeMkvOutputFile: mkvFile,
			TitleId:           title.Index,
			// TODO: Allow the final name to be built with a template?
			FinalName:         title.Item.Title,
		})
	}

	return plan, nil
}

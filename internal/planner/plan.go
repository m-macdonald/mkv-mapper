package planner

import (
	"fmt"
	"path/filepath"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/naming"
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
	EstimatedSize     uint
}

func BuildPlan(
	discRoot string,
	outputDir string,
	templateConfig config.TemplateConfig,
	disc *discdb.Disc,
	titles map[signature.SegmentSignature]makemkv.Title,
) (*DiscPlan, error) {
	mappings, err := mapper.MapTitles(disc, titles)
	if err != nil {
		return nil, fmt.Errorf("failed to map MakeMkv titles to DiscDB titles %w", err)
	}

	filenmGen, err := naming.NewGenerator(templateConfig)
	if err != nil {
		return nil, err
	}

	plan := &DiscPlan{
		DiscRoot:  discRoot,
		OutputDir: outputDir,
	}

	for _, mapping := range mappings {
		filenm, err := renderFileNm(filenmGen, *disc, mapping)
		if err != nil {
			// TODO: Probably preferable to not kill the whole process here. Just report that this specific title could not be renamed
			return nil, err
		}
		plan.Titles = append(plan.Titles, TitlePlan{
			SourcePlaylist:    mapping.MakeMkvTitle.SourceFileName,
			MakeMkvOutputFile: mapping.MakeMkvTitle.OutputFileName,
			FinalName:         filenm,
			EstimatedSize:     mapping.MakeMkvTitle.OutputFileSize,
		})
	}

	return plan, nil
}

func renderFileNm(filenmGen *naming.Generator, disc discdb.Disc, mapping mapper.TitleMapping) (string, error) {
	titleContext := naming.TitleContext{
		DiscDbDisc:   disc,
		DiscDbTitle:  mapping.DiscDbTitle,
		MakeMkvTitle: mapping.MakeMkvTitle,
	}
	filenm, err := filenmGen.Render(titleContext)
	if err != nil {
		return "", err
	}

	// Pretty sure this should always be ".mkv", but just in case...
	fileExt := filepath.Ext(mapping.MakeMkvTitle.OutputFileName)

	return filenm + fileExt, nil
}

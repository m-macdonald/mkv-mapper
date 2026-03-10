package app

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/planner"

	"go.uber.org/zap"
)

type Pipeline struct {
	makemkv *makemkv.Client
	discdb  *discdb.Client
	logger  *zap.SugaredLogger
}

func New(
	makemkv *makemkv.Client,
	discdb *discdb.Client,
	logger *zap.SugaredLogger,
) *Pipeline {
	return &Pipeline{
		makemkv: makemkv,
		discdb: discdb,
		logger: logger,
	}
}

func (p *Pipeline) BuildPlan(discRoot string, outputDir string) (*planner.DiscPlan, error) {
	root, err := files.ResolveDiscRoot(discRoot)
	if err != nil {
		return nil, fmt.Errorf("unable to find disc root %w", err)
	}
	// if isValid := fs.ValidPath(outputDir); !isValid {
	// 	return nil, fmt.Errorf("the given outputDir is invalid: %s", outputDir)	
	// }
	hash, err := files.Hash(root)
	if err != nil {
		return nil, fmt.Errorf("unable to hash disc %w", err)
	}

	disc, err := p.discdb.GetDisc(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve disc definitions from TheDiscDB %w", err)
	}

	titles, err := p.makemkv.ReadTitles(root)
	if err != nil {
		return nil, fmt.Errorf("unable to read disc titles using MakeMkv %w", err)
	}

	plan, err := planner.BuildPlan(root, outputDir, disc, titles)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a plan for ripping the disc %w", err)	
	}

	return plan, nil
}

func (p *Pipeline) RunPlan(plan *planner.DiscPlan) error {
	err := p.makemkv.RipDisc(plan.DiscRoot, plan.OutputDir)
	if err != nil {
		return err
	}
	
	mappings := make(map[string]string)
	for _, titlePlan := range plan.Titles {
		mappings[titlePlan.MakeMkvOutputFile] = titlePlan.FinalName 
	}
	errs := mapper.RenameTitles(plan.OutputDir, plan.OutputDir, mappings)
	if len(errs) != 0 {
		p.logger.Errorf("%v", errs)
	}

	return nil
}

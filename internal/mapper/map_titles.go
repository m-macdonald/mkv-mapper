package mapper

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/signature"
)

type TitleMapping struct {
	MakeMkvTitle makemkv.Title
	DiscdbTitle  discdb.Title
}

func MapTitles(discDbDisc *discdb.Disc, makeMkvTitlesBySegmentSignature map[signature.SegmentSignature]makemkv.Title) ([]TitleMapping, error) {
	var mappings []TitleMapping
	for _, discDbTitle := range discDbDisc.Titles {
		segmentSignature, err := signature.NormalizeSegments(discDbTitle.SegmentMap)
		if err != nil {
			return nil, fmt.Errorf("unable to create segment signature for %d %s %w", discDbTitle.Index, discDbTitle.SegmentMap, err)
		}
		if makeMkvTitle, ok := makeMkvTitlesBySegmentSignature[segmentSignature]; ok {
			mappings = append(mappings, TitleMapping{
				MakeMkvTitle: 	makeMkvTitle,
				DiscdbTitle:    discDbTitle,
			})
		}
	}

	return mappings, nil
}

package mapper

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/signature"
)

type TitleMapping struct {
	MakeMkvTitle makemkv.Title
	DiscDbTitle  discdb.Title
}

func MapTitles(
	discRecord *discdb.DiscRecord,
	makeMkvTitlesBySegmentSignature map[signature.SegmentSignature]makemkv.Title,
) ([]TitleMapping, error) {
	var mappings []TitleMapping
	for _, discDbTitle := range discRecord.Disc.Titles {
		segmentSignature, err := signature.NormalizeSegments(discDbTitle.SegmentMap)
		if err != nil {
			return nil, fmt.Errorf("unable to create segment signature for discdb segment map %s: %w", discDbTitle.SegmentMap, err)
		}
		if makeMkvTitle, ok := makeMkvTitlesBySegmentSignature[segmentSignature]; ok {
			mappings = append(mappings, TitleMapping{
				MakeMkvTitle: makeMkvTitle,
				DiscDbTitle:  discDbTitle,
			})
		}
	}

	return mappings, nil
}

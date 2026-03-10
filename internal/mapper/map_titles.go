package mapper

import (
	"fmt"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/signature"
)

func MapTitles(discDbDisc *discdb.Disc, makeMkvOutputFilesBySegmentSignature map[signature.SegmentSignature]string) (map[string]discdb.Title, error) {
    mappings := make(map[string]discdb.Title)
	for _, discDbTitle := range discDbDisc.Titles {
		segmentSignature, err := signature.NormalizeSegments(discDbTitle.SegmentMap)
		if err != nil {
			return nil, fmt.Errorf("unable to create segment signature for %d %s %w", discDbTitle.Index, discDbTitle.SegmentMap, err)
		}
		if makeMkvOutputFileName, ok := makeMkvOutputFilesBySegmentSignature[segmentSignature]; ok {
			mappings[makeMkvOutputFileName] = discDbTitle
		}
	}

	return mappings, nil
}

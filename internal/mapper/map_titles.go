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
	discRecord discdb.DiscRecord,
	makemkvTitles []makemkv.Title,
) ([]TitleMapping, error) {
	grouped, err := groupBySegmentSignature(makemkvTitles)
	if err != nil {
		return nil, err
	}

	var mappings []TitleMapping
	for _, discDbTitle := range discRecord.Disc.Titles {
		segmentSignature, err := signature.NormalizeSegments(discDbTitle.SegmentMap)
		if err != nil {
			return nil, fmt.Errorf("unable to create segment signature for discdb segment map %s: %w", discDbTitle.SegmentMap, err)
		}
		if makeMkvTitle, ok := grouped[segmentSignature]; ok {
			mappings = append(mappings, TitleMapping{
				MakeMkvTitle: makeMkvTitle,
				DiscDbTitle:  discDbTitle,
			})
		}
	}

	return mappings, nil
}

func groupBySegmentSignature(titles []makemkv.Title) (map[signature.SegmentSignature]makemkv.Title, error) {
	grouped := make(map[signature.SegmentSignature]makemkv.Title, len(titles))
	for _, title := range titles {
		sig, err := signature.NormalizeSegments(title.Segments)
		if err != nil {
			return nil, err
		}
		grouped[sig] = title
	}

	return grouped, nil
}

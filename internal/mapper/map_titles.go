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
	groupedDiscDb, err := groupDiscDbBySignature(discRecord.Disc.Titles)
	if err != nil {
		return nil, err
	}
	groupedMakeMkv, err := groupMakeMkvBySignature(makemkvTitles)
	if err != nil {
		return nil, err
	}
	
	mappings := make([]TitleMapping, 0, len(makemkvTitles))
	for _, makeMkvTitle := range groupedMakeMkv { 
		signature, err := signature.NormalizeSegments(makeMkvTitle.Segments)
		if err != nil {
			return nil, fmt.Errorf("unable to create segment signature for discdb segment map %s: %w", makeMkvTitle.Segments, err)
		}
		mappings = append(mappings, TitleMapping{
			MakeMkvTitle: makeMkvTitle,
			// Worth keeping in mind that this will result in a zero-valued DiscDbTitle if there is no match.
			DiscDbTitle: groupedDiscDb[signature],
		})
	}
	return mappings, nil
}

func groupMakeMkvBySignature(titles []makemkv.Title) ([]makemkv.Title, error) {
	seen := make(map[signature.SegmentSignature]bool, len(titles))
	deduped := make([]makemkv.Title, 0, len(titles))
	for _, t := range titles {
		sig, err := signature.NormalizeSegments(t.Segments)
		if err != nil {
			return nil, fmt.Errorf("unable to create segment signature for makemkv segments %s: %w", t.Segments, err)
		}
		if seen[sig] {
			continue // duplicate content, already represented by an earlier title
		}
		seen[sig] = true
		deduped = append(deduped, t)
	}
	return deduped, nil
}

func groupDiscDbBySignature(titles []discdb.Title) (map[signature.SegmentSignature]discdb.Title, error) {
	grouped := make(map[signature.SegmentSignature]discdb.Title, len(titles))
	for _, title := range titles {
		signature, err := signature.NormalizeSegments(title.SegmentMap)
		if err != nil {
			return nil, fmt.Errorf("unable to create segment signature for discdb segment map %s: %w", title.SegmentMap, err)
		}
		if existing, ok := grouped[signature]; ok && existing.Item != nil {
			continue
		}
		grouped[signature] = title
	}
	return grouped, nil
}

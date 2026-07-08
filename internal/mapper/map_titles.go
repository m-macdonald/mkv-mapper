package mapper

import (
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
) []TitleMapping {
	groupedDiscDb := groupDiscDbBySignature(discRecord.Disc.Titles)
	groupedMakeMkv := groupMakeMkvBySignature(makemkvTitles)

	mappings := make([]TitleMapping, 0, len(makemkvTitles))
	for _, makeMkvTitle := range groupedMakeMkv {
		mappings = append(mappings, TitleMapping{
			MakeMkvTitle: makeMkvTitle,
			// Worth keeping in mind that this will result in a zero-valued DiscDbTitle if there is no match.
			DiscDbTitle: groupedDiscDb[makeMkvTitle.Signature],
		})
	}
	return mappings
}

func groupMakeMkvBySignature(titles []makemkv.Title) []makemkv.Title {
	seen := make(map[signature.SegmentSignature]bool, len(titles))
	deduped := make([]makemkv.Title, 0, len(titles))
	for _, title := range titles {
		if seen[title.Signature] {
			continue // duplicate content, already represented by an earlier title
		}
		seen[title.Signature] = true
		deduped = append(deduped, title)
	}
	return deduped
}

func groupDiscDbBySignature(titles []discdb.Title) map[signature.SegmentSignature]discdb.Title {
	grouped := make(map[signature.SegmentSignature]discdb.Title, len(titles))
	for _, title := range titles {
		if existing, ok := grouped[title.Signature]; ok && existing.Item != nil {
			continue
		}
		grouped[title.Signature] = title
	}
	return grouped
}

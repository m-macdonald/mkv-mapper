package mapper

import (
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
)

type TitleMapping struct {
	MakeMkvTitle makemkv.Title
	DiscDbTitle  discdb.Title
}

func MapTitles(
	discRecord discdb.DiscRecord,
	makemkvTitles []makemkv.Title,
) []TitleMapping {
	mappings := make([]TitleMapping, 0, len(makemkvTitles))
	for _, makeMkvTitle := range makemkvTitles {
		mappings = append(mappings, TitleMapping{
			MakeMkvTitle: makeMkvTitle,
			// Worth keeping in mind that this will result in a zero-valued DiscDbTitle if there is no match.
			DiscDbTitle: matchDiscDbTitle(makeMkvTitle, discRecord.Disc.Titles),
		})
	}
	return mappings
}

func matchDiscDbTitle(
	makemkvTitle makemkv.Title,
	discDbTitles []discdb.Title,
) discdb.Title {
	var matches []discdb.Title
	for _, t := range discDbTitles {
		if makemkvTitle.Identity.Matches(t.Identity()) {
			matches = append(matches, t)
		}
	}

	switch len(matches) {
	case 0:
		return discdb.Title{}
	case 1:
		return matches[0]
	default:
		// TODO: Add logging about ambiguous matching
		return discdb.Title{}
	}
}

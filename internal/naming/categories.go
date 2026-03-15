package naming

import "m-macdonald/mkv-mapper/internal/discdb"

type templateType string

const (
	templateTypeEpisode  templateType = "episode"
	templateTypeExtra    templateType = "extra"
	templateTypeMovie    templateType = "movie"
	templateTypeOverride templateType = "override"
	templateTypeUnknown  templateType = "fallback"
)

func templateTypeFromItemType(t discdb.ItemType) templateType {
	switch t {
	case discdb.ItemTypeMovie:
		return templateTypeMovie
	case discdb.ItemTypeEpisode:
		return templateTypeEpisode
	case discdb.ItemTypeDeletedScene,
		discdb.ItemTypeExtra,
		discdb.ItemTypeTrailer:
		return templateTypeExtra
	default:
		return templateTypeUnknown
	}
}

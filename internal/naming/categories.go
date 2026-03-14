package naming

import "m-macdonald/mkv-mapper/internal/discdb"

type templateType string

const (
	templateTypeMovie    templateType = "movie"
	templateTypeEpisode  templateType = "episode"
	templateTypeExtra    templateType = "extra"
	templateTypeOverride templateType = "override"
	templateTypeFallback templateType = "fallback"
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
		return templateTypeFallback
	}
}

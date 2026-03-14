package naming

import "m-macdonald/mkv-mapper/internal/config"

var defaultTemplates = config.TemplateConfig{
	Movie:    "{{.Disc.Title}} ({{.Disc.Year}})",
	Episode:  "{{.Disc.SeriesTitle}}/Season {{.Item.Season}}/{{.Disc.SeriesTitle}} - S{{pad 2 .Item.Season}}E{{.Item.Episode}} - {{.Item.Title}}",
	Extra:    "Extras/{{.Item.Title}}",
	Fallback: "{{.MakeMkv.OutputFileName}}",
}

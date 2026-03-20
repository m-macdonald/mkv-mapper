package naming

import "m-macdonald/mkv-mapper/internal/config"

var defaultTemplates = config.TemplateConfig{
	Movie:   "{{.Media.Title}} ({{.Disc.Year}})",
	Episode: "{{.Media.Title}}/Season {{.Item.Season}}/{{.Disc.SeriesTitle}} - S{{pad 2 .Item.Season}}E{{.Item.Episode}} - {{.Item.Title}}",
	Extra:   "Extras/{{.Item.Title}}",
	Unknown: "{{.MakeMkv.OutputFileName}}",
}

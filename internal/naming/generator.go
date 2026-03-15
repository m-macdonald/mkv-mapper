package naming

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
)

type Generator struct {
	templates *template.Template
}

type TitleContext struct {
	DiscDbDisc   discdb.Disc
	DiscDbTitle  discdb.Title
	MakeMkvTitle makemkv.Title
}

func NewGenerator(userTemplates config.TemplateConfig) (*Generator, error) {
	merged := mergeTemplates(userTemplates)

	rootTemplate := template.New("root").
		Funcs(templateFuncs()).
		Option("missingkey=error")

	templates := map[templateType]string{
		templateTypeMovie:   merged.Movie,
		templateTypeEpisode: merged.Episode,
		templateTypeExtra:   merged.Extra,
		templateTypeUnknown: merged.Unknown,
	}

	if merged.Override != "" {
		templates[templateTypeOverride] = merged.Override
	}

	for name, template := range templates {
		if _, err := rootTemplate.New(string(name)).Parse(template); err != nil {
			return nil, fmt.Errorf("parsing %s template: %w", name, err)
		}
	}

	return &Generator{templates: rootTemplate}, nil
}

func (g *Generator) Render(titleCtx TitleContext) (string, error) {
	templateType := templateTypeFromItemType(titleCtx.DiscDbTitle.Item.Type)
	vars := buildTemplateVars(titleCtx)

	if g.templates.Lookup(string(templateTypeOverride)) != nil {
		templateType = templateTypeOverride
	}

	var buf bytes.Buffer
	err := g.templates.ExecuteTemplate(&buf, string(templateType), vars)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func mergeTemplates(userTemplates config.TemplateConfig) config.TemplateConfig {
	result := config.TemplateConfig{}

	if userTemplates.Movie != "" {
		result.Movie = userTemplates.Movie
	} else {
		result.Movie = defaultTemplates.Movie
	}

	if userTemplates.Episode != "" {
		result.Episode = userTemplates.Episode
	} else {
		result.Episode = defaultTemplates.Episode
	}

	if userTemplates.Extra != "" {
		result.Extra = userTemplates.Extra
	} else {
		result.Extra = defaultTemplates.Extra
	}

	if userTemplates.Unknown != "" {
		result.Unknown = userTemplates.Unknown
	} else {
		result.Unknown = defaultTemplates.Unknown
	}

	if userTemplates.Override != "" {
		result.Override = userTemplates.Override
	}

	return result
}

func buildTemplateVars(titleCtx TitleContext) map[string]any {
	return map[string]any{
		"Disc":    titleCtx.DiscDbDisc,
		"Title":   titleCtx.DiscDbTitle,
		"MakeMkv": titleCtx.MakeMkvTitle,

		"Season":       titleCtx.DiscDbTitle.Item.Season,
		"Episode":      titleCtx.DiscDbTitle.Item.Episode,
		"EpisodeTitle": titleCtx.DiscDbTitle.Item.Title,
	}
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"pad": func(padCnt uint, val string) string {
			return fmt.Sprintf("%0*s", padCnt, val)
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"dflt": func(dflt string, val string) string {
			if strings.TrimSpace(val) == "" {
				return dflt
			}

			return val
		},
	}
}

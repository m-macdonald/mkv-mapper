package naming

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
)

type Generator struct {
	templates *template.Template
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

func (g *Generator) Render(disc discdb.Disc, title discdb.Title) (string, error) {
	vars := buildTemplateVars(disc, title)

	var buf bytes.Buffer
	if template := g.templates.Lookup(string(templateTypeOverride)); template != nil {
		err := template.Execute(&buf, vars)
		if err != nil {
			return "", fmt.Errorf("override template failed to execute %w", err)
		}

		return buf.String(), nil
	}

	templateType := templateTypeFromItemType(title.Item.Type)
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

	if userTemplates.Episode != "" {
		result.Extra = userTemplates.Extra
	} else {
		result.Extra = defaultTemplates.Extra
	}

	if userTemplates.Override != "" {
		result.Override = userTemplates.Override
	}

	return result
}

func buildTemplateVars(disc discdb.Disc, title discdb.Title) map[string]any {
	return map[string]any{
		"Disc":  disc,
		"Title": title,

		"Season":       title.Item.Season,
		"Episode":      title.Item.Episode,
		"EpisodeTitle": title.Item.Title,
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

package naming

import (
	"bytes"
	"fmt"
	"m-macdonald/mkv-mapper/internal/discdb"
	"strings"
	"text/template"
)

type Generator struct {
	tmpl *template.Template
}

func NewGenerator(templateStr string) (*Generator, error) {
	tmpl, err := template.
		New("filename").
		Funcs(templateFuncs()).
		Parse(templateStr)

	if err != nil {
		return nil, err
	}

	return &Generator{tmpl: tmpl}, nil
}

func (g *Generator) Render(disc discdb.Disc, title discdb.Title) (string, error) {
	vars := buildTemplateVars(disc, title)

	var buf bytes.Buffer

	err := g.tmpl.Execute(&buf, vars)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func buildTemplateVars(disc discdb.Disc, title discdb.Title) map[string]any {
	return map[string]any{
		"Disc": disc,
		"Title": title,

		"Season": title.Item.Season,
		"Episode": title.Item.Episode,
		"EpisodeTitle": title.Item.Title,
	}
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"pad": func(padCnt uint, val int) string {
			padTmpl := fmt.Sprintf("%0d", padCnt)

			return fmt.Sprintf(padTmpl, val)
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

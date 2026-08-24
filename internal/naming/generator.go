package naming

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type generator[C any, V any] struct {
	templates      *template.Template
	buildVars      func(C) V
	selectTemplate func(C) string
}

func newGenerator[C any, V any](
	templates map[string]string,
	buildVars func(C) V,
	selectTemplate func(C) string,
) (*generator[C, V], error) {
	root := template.New("root").
		Funcs(templateFuncs()).
		Option("missingkey=error")

	for name, tmplText := range templates {
		if _, err := root.New(name).Parse(tmplText); err != nil {
			return nil, fmt.Errorf("parsing %s to template: %w", name, err)
		}
	}

	return &generator[C, V]{templates: root, buildVars: buildVars, selectTemplate: selectTemplate}, nil
}

func newSingleTemplateGenerator[C any, V any](
	name string,
	tmplText string,
	buildVars func(C) V,
) (*generator[C, V], error) {
	return newGenerator(
		map[string]string{name: tmplText},
		buildVars,
		func(C) string { return name },
	)
}

func (g *generator[C, V]) Generate(ctx C) (string, error) {
	name := g.selectTemplate(ctx)
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, name, g.buildVars(ctx)); err != nil {
		return "", fmt.Errorf("rendering %s template: %w", name, err)
	}
	return buf.String(), nil
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"pad":   pad,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"dflt":  dflt,
	}
}

func pad(padCnt int, val string) string {
	return fmt.Sprintf("%0*s", padCnt, val)
}

func dflt(dflt string, val string) string {
	if strings.TrimSpace(val) == "" {
		return dflt
	}

	return val
}

package naming

import (
	"bytes"
	"fmt"
	"html/template"

	"m-macdonald/mkv-mapper/internal/discdb"
)

type DirectoryContext struct {
	Media discdb.Media
	Disc  discdb.Disc
}

type BackupDirectoryContext struct {
	Label string
}

type outputDirVars struct {
	Media TemplateMedia
	Disc  TemplateDisc
}

type backupOutputDirVars struct {
	Disc templateDiscIdentity
}

type templateDiscIdentity struct {
	Label string
}

type DirectoryGenerator struct {
	template *template.Template
}

func NewDirectoryGenerator(dirTemplate string) (*DirectoryGenerator, error) {
	root, err := template.New("root").
		Funcs(templateFuncs()).
		Option("missingkey=error").
		Parse(dirTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing directory template: %w", err)
	}
	return &DirectoryGenerator{template: root}, nil
}

func (g *DirectoryGenerator) Generate(ctx DirectoryContext) (string, error) {
	vars := outputDirVars{
		Media: TemplateMedia{
			Title: sanitizeSegment(ctx.Media.Title),
			Year:  ctx.Media.Year,
		},
		Disc: TemplateDisc{
			ContentHash: ctx.Disc.ContentHash,
			Format:      ctx.Disc.Format,
			Name:        sanitizeSegment(ctx.Disc.Name),
			Slug:        ctx.Disc.Slug,
		},
	}

	return g.execute(vars)
}

func (g *DirectoryGenerator) GenerateBackup(ctx BackupDirectoryContext) (string, error) {
	vars := backupOutputDirVars{
		Disc: templateDiscIdentity{Label: ctx.Label},
	}
	return g.execute(vars)
}

func (g *DirectoryGenerator) execute(vars any) (string, error) {
	var buf bytes.Buffer
	if err := g.template.Execute(&buf, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

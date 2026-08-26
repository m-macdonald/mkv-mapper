package naming

import (
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

type OutputDirGenerator = generator[DirectoryContext, outputDirVars]

func NewRipOutputDirGenerator(tmplText string) (*OutputDirGenerator, error) {
	return newSingleTemplateGenerator("ripOutputDir", tmplText, func(ctx DirectoryContext) outputDirVars {
		return outputDirVars{
			Media: TemplateMedia{
				Title: sanitizeSegment(ctx.Media.Title),
				Year:  ctx.Media.Year,
				Type:  ctx.Media.Type,
			},
			Disc: TemplateDisc{
				ContentHash: ctx.Disc.ContentHash,
				Format:      ctx.Disc.Format,
				Name:        sanitizeSegment(ctx.Disc.Name),
				Slug:        ctx.Disc.Slug,
			},
		}
	})
}

type BackupOutputDirGenerator = generator[BackupDirectoryContext, backupOutputDirVars]

func NewBackupOutputDirGenerator(tmplText string) (*BackupOutputDirGenerator, error) {
	return newSingleTemplateGenerator("backupOutputDir", tmplText, func(ctx BackupDirectoryContext) backupOutputDirVars {
		return backupOutputDirVars{
			Disc: templateDiscIdentity{Label: sanitizeSegment(ctx.Label)},
		}
	})
}

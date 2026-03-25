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

type filenameGenerator interface {
	Generate(titleCtx TitleContext) (string, error)
}

type FilenameGenerator struct {
	templates *template.Template
}

type TitleContext struct {
	DiscDbMedia  discdb.Media
	DiscDbTitle  discdb.Title
	DiscDbDisc   discdb.Disc
	MakeMkvTitle makemkv.Title
}

func NewFilenameGenerator(templateConfig config.TemplateConfig) (*FilenameGenerator, error) {
	rootTemplate := template.New("root").
		Funcs(templateFuncs()).
		Option("missingkey=error")

	templates := map[templateType]string{
		templateTypeMovie:   templateConfig.Movie,
		templateTypeEpisode: templateConfig.Episode,
		templateTypeExtra:   templateConfig.Extra,
		templateTypeUnknown: templateConfig.Unknown,
	}

	if templateConfig.Override != "" {
		templates[templateTypeOverride] = templateConfig.Override
	}

	for name, template := range templates {
		if _, err := rootTemplate.New(string(name)).Parse(template); err != nil {
			return nil, fmt.Errorf("parsing %s template: %w", name, err)
		}
	}

	return &FilenameGenerator{templates: rootTemplate}, nil
}

func (f *FilenameGenerator) Generate(titleCtx TitleContext) (string, error) {
	templateType := templateTypeUnknown
	if item, ok := titleCtx.DiscDbTitle.ItemValue(); ok {
		templateType = templateTypeFromItemType(item.Type)
	}

	vars := buildTemplateVars(titleCtx)

	if f.templates.Lookup(string(templateTypeOverride)) != nil {
		templateType = templateTypeOverride
	}

	var buf bytes.Buffer
	err := f.templates.ExecuteTemplate(&buf, string(templateType), vars)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type TemplateVars struct {
	Media   TemplateMedia
	Disc    TemplateDisc
	Title   TemplateTitle
	MakeMkv TemplateMakeMkvTitle

	Season       string
	Episode      string
	EpisodeTitle string
	MovieTitle   string
}

type TemplateMedia struct {
	Title string
	Year  int
	Type  string
}

type TemplateDisc struct {
	ContentHash string
	Format      string
	Name        string
	Slug        string
}

type TemplateTitle struct {
	DisplaySize string
	Duration    string
	SegmentMap  string
	Size        uint64
	SourceFile  string
}

type TemplateMakeMkvTitle struct {
	TitleId          int
	OutputFilename   string
	SourceFilename   string
	SegmentSignature string
	OutputFileSize   uint64
}

func buildTemplateVars(titleCtx TitleContext) TemplateVars {
	item, _ := titleCtx.DiscDbTitle.ItemValue()

	return TemplateVars{
		Media: TemplateMedia{
			Title: titleCtx.DiscDbMedia.Title,
			Year:  titleCtx.DiscDbMedia.Year,
			Type:  titleCtx.DiscDbMedia.Type,
		},
		Disc: TemplateDisc{
			ContentHash: titleCtx.DiscDbDisc.ContentHash,
			Format:      titleCtx.DiscDbDisc.Format,
			Name:        titleCtx.DiscDbDisc.Name,
			Slug:        titleCtx.DiscDbDisc.Slug,
		},
		Title: TemplateTitle{
			DisplaySize: titleCtx.DiscDbTitle.DisplaySize,
			Duration:    titleCtx.DiscDbTitle.Duration,
			SegmentMap:  titleCtx.DiscDbTitle.SegmentMap,
			Size:        titleCtx.DiscDbTitle.Size,
			SourceFile:  titleCtx.DiscDbTitle.SourceFile,
		},
		MakeMkv: TemplateMakeMkvTitle{
			TitleId:          titleCtx.MakeMkvTitle.TitleId,
			OutputFilename:   titleCtx.MakeMkvTitle.OutputFilename,
			SourceFilename:   titleCtx.MakeMkvTitle.SourceFilename,
			SegmentSignature: string(titleCtx.MakeMkvTitle.SegmentSignature),
			OutputFileSize:   titleCtx.MakeMkvTitle.OutputFileSize,
		},

		Season:       item.Season,
		Episode:      item.Episode,
		EpisodeTitle: item.Title,
		MovieTitle:   item.Title,
	}
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"pad": pad,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"dflt": dflt,
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

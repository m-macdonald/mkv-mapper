package naming

import (
	"fmt"
	"path/filepath"
	"strings"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
)

const maxUniqueFilenameAttempts = 1000

type FilenameResolution struct {
	FinalName string
	Events    []FilenameEvent
}

type FilenameEvent struct {
	Code    WarningCode
	Message string
	Cause   error
}

type WarningCode string

const (
	WarningNamingFallback   WarningCode = "naming_fallback"
	WarningFilenameSuffixed WarningCode = "filename_suffixed"
)

func ResolveFilename(
	filenameGen filenameGenerator,
	titleContext TitleContext,
	used map[string]struct{},
) (FilenameResolution, error) {
	var events []FilenameEvent

	ext := filepath.Ext(titleContext.MakeMkvTitle.OutputFilename)
	baseName, err := filenameGen.Generate(titleContext)
	if err != nil {
		baseName = strings.TrimSuffix(titleContext.MakeMkvTitle.OutputFilename, ext)
		events = append(events, FilenameEvent{
			Code:    WarningNamingFallback,
			Message: "failed to resolve configured filename; using MakeMKV filename",
			Cause:   err,
		})
	}

	sanitizedName := sanitizeSegment(baseName)

	finalName, collisionResolved, err := ensureUniqueFilename(
		sanitizedName,
		ext,
		titleContext.MakeMkvTitle.TitleId,
		used)
	if err != nil {
		return FilenameResolution{}, err
	}
	if collisionResolved {
		events = append(events, FilenameEvent{
			Code:    WarningFilenameSuffixed,
			Message: "generated filename was not unique; appended title suffix",
		})
	}

	return FilenameResolution{
		FinalName: finalName,
		Events:    events,
	}, nil
}

func ensureUniqueFilename(
	baseName string,
	ext string,
	titleId lines.TitleId,
	used map[string]struct{},
) (string, bool, error) {
	filename := baseName + ext
	if _, exists := used[filename]; !exists {
		used[filename] = struct{}{}

		return filename, false, nil
	}

	filename = fmt.Sprintf("%s_t%d%s", baseName, titleId, ext)
	if _, exists := used[filename]; !exists {
		used[filename] = struct{}{}

		return filename, true, nil
	}

	// This should realistically never be needed.
	// It's even less likely that we exhaust the maxAttempts
	// If we do, something has gone quite wrong and we should exit
	for n := 1; n <= maxUniqueFilenameAttempts; n++ {
		filename = fmt.Sprintf("%s_t%d_%d%s", baseName, titleId, n, ext)
		if _, exists := used[filename]; !exists {
			used[filename] = struct{}{}

			return filename, true, nil
		}
	}

	return "", false, fmt.Errorf(
		"could not resolve unique filename for title %s after %d attempts",
		baseName,
		maxUniqueFilenameAttempts)
}

type filenameGenerator interface {
	Generate(titleCtx TitleContext) (string, error)
}

type FilenameGenerator = generator[TitleContext, TemplateVars] 

type TitleContext struct {
	DiscDbMedia  discdb.Media
	DiscDbTitle  discdb.Title
	DiscDbDisc   discdb.Disc
	MakeMkvTitle makemkv.Title
}

func NewFilenameGenerator(templateConfig config.TemplateConfig) (*FilenameGenerator, error) {
	templates := map[string]string{
		string(templateTypeMovie):   templateConfig.Movie,
		string(templateTypeEpisode): templateConfig.Episode,
		string(templateTypeExtra):   templateConfig.Extra,
		string(templateTypeUnknown): templateConfig.Unknown,
	}

	hasOverride := templateConfig.Override != ""
	if hasOverride {
		templates[string(templateTypeOverride)] = templateConfig.Override
	}
	
	selectTemplate := func(titleCtx TitleContext) string {
		if hasOverride {
			return string(templateTypeOverride)
		}
		templateType := templateTypeUnknown
		if item, ok := titleCtx.DiscDbTitle.ItemValue(); ok {
			templateType = templateTypeFromItemType(item.Type)
		}
		return string(templateType)
	}

	return newGenerator(templates, buildTemplateVars, selectTemplate)
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
	TitleId        lines.TitleId
	OutputFilename string
	SourceFilename string
	Segments       string
	OutputFileSize uint64
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
			TitleId:        titleCtx.MakeMkvTitle.TitleId,
			OutputFilename: titleCtx.MakeMkvTitle.OutputFilename,
			SourceFilename: titleCtx.MakeMkvTitle.SourceFilename,
			Segments:       string(titleCtx.MakeMkvTitle.Signature),
			OutputFileSize: titleCtx.MakeMkvTitle.OutputFileSize,
		},

		Season:       item.Season,
		Episode:      item.Episode,
		EpisodeTitle: item.Title,
		MovieTitle:   item.Title,
	}
}

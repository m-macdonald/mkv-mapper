package planner

import (
	"fmt"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/mapper"
	"m-macdonald/mkv-mapper/internal/naming"
	"path/filepath"
	"strings"
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

func resolveFilename(
	filenameGen *naming.Generator,
	disc *discdb.Disc,
	mapping mapper.TitleMapping,
	used map[string]struct{},
) (FilenameResolution, error) {
	var events []FilenameEvent

	ext := filepath.Ext(mapping.MakeMkvTitle.OutputFileName)
	baseName, err := resolveBaseFilename(filenameGen, *disc, mapping)
	if err != nil {
		baseName = strings.TrimSuffix(mapping.MakeMkvTitle.OutputFileName, ext)
		events = append(events, FilenameEvent{
			Code:    WarningNamingFallback,
			Message: "failed to resolve configured filename; using MakeMKV filename",
			Cause:   err,
		})
	}

	finalName, collisionResolved, err := ensureUniqueFilename(
		baseName,
		ext,
		mapping.MakeMkvTitle.TitleId,
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

func resolveBaseFilename(
	filenmGen *naming.Generator,
	disc discdb.Disc,
	mapping mapper.TitleMapping,
) (string, error) {
	titleContext := naming.TitleContext{
		DiscDbDisc:   disc,
		DiscDbTitle:  mapping.DiscDbTitle,
		MakeMkvTitle: mapping.MakeMkvTitle,
	}

	return filenmGen.Render(titleContext)
}

func ensureUniqueFilename(
	baseName string,
	ext string,
	titleId uint,
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

	return "", true, fmt.Errorf(
		"could not resolve unique filename for title %s after %d attempts",
		baseName,
		maxUniqueFilenameAttempts)
}

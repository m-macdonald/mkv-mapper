package naming

import (
	"fmt"
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

	finalName, collisionResolved, err := ensureUniqueFilename(
		baseName,
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
	titleId int,
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

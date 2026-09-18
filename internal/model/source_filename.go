package model

import "strings"

// SourceFilename is the name of the underlying file backing a title.
// It is used in title equality comparisons across the codebase
type SourceFilename string

func NewSourceFilename(s string) SourceFilename {
	return SourceFilename(strings.TrimSpace(s))
}

package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IsAncestorOrEqual(dir1 string, dir2 string) (bool, error) {
	dir1, err := filepath.Abs(dir1)
	if err != nil {
		return false, fmt.Errorf("resolving directory: %w", err)
	}
	dir2, err = filepath.Abs(dir2)
	if err != nil {
		return false, fmt.Errorf("resolving directory: %w", err)
	}

	rel, err := filepath.Rel(dir1, dir2)
	if err != nil {
		return false, fmt.Errorf("comparing directories: %w", err)
	}

	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))), nil
}

func EnsureDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %q: %w", dir, err)
	}
	return nil
}

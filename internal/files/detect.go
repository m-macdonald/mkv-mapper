package files

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func ResolveDiscRoot(cliRoot string) (string, error) {
	if cliRoot != "" {
		return cliRoot, nil
	}
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	base := filepath.Join("/run/media", u.Username)

	return findMountedDisc(base)
}

func findMountedDisc(base string) (string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", err
	}

	var discMounts []string
	for _, e := range entries {
		name := e.Name()
		candidate := filepath.Join(base, name)
		if e.IsDir() {
			if isDir(filepath.Join(candidate, "BDMV", "STREAM")) ||
				isDir(filepath.Join(candidate, "VIDEO_TS")) {
				discMounts = append(discMounts, candidate)
			}
		} else if strings.EqualFold(filepath.Ext(name), ".iso") {
			discMounts = append(discMounts, candidate)
		}
	}

	switch len(discMounts) {
	case 0:
		return "", fmt.Errorf("failure resolving disc root: %w", err)
	case 1:
		return discMounts[0], nil
	default:
		return "", fmt.Errorf("multiple discs found: %s\nUse --disc-root to specify which disc to rip", strings.Join(discMounts, ", "))
	}
}

func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	return stat.IsDir()
}

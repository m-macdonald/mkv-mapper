package files

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func ResolveDiscRoot(cliRoot string) ([]string, error) {
	if cliRoot != "" {
		return []string{cliRoot}, nil
	}
	u, err := user.Current()
	if err != nil {
		return []string{}, err
	}
	base := filepath.Join("/run/media", u.Username)

	return findMountedDisc(base)
}

func findMountedDisc(base string) ([]string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
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
	
	return discMounts, nil
}


func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	return stat.IsDir()
}

package files

import (
	"errors"
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

// Stops on the first mounted disc that it finds.
// TODO: Extend this to allow for multiple discs to be returned
func findMountedDisc(base string) (string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", err
	}

	for _, e := range entries {
		name := e.Name()
		candidate := filepath.Join(base, name)
		if e.IsDir() {
			if isDir(filepath.Join(candidate, "BDMV", "STREAM")) ||
				isDir(filepath.Join(candidate, "VIDEO_TS")) {
				return candidate, nil
			}
		} else if strings.EqualFold(filepath.Ext(name), ".iso") {
			return candidate, nil
		}
	}
	
	return "", errors.New("no disc found")
}


func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	return stat.IsDir()
}

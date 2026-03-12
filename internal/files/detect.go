package files

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
)

func ResolveDiscRoot(cliRoot string) (string, error) {
	if cliRoot != "" {
		return cliRoot, nil
	}

	return findBluRayMount()
}

func findBluRayMount() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	base := filepath.Join("/run/media", u.Username)

	entries, err := os.ReadDir(base)
	if err != nil {
		return "", err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		candidate := filepath.Join(base, e.Name())
		streamDir := filepath.Join(candidate, "BDMV", "STREAM")

		info, err := os.Stat(streamDir)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	
	return "", errors.New("no Blu-ray mount found")
}

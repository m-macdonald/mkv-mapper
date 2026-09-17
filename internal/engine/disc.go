package engine

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/model"
)

func (e *Engine) resolveDiscSourceByLabel(ctx context.Context, label string) (model.DiscSource, error) {
	drives, err := e.makemkv.ScanDrives(ctx)
	if err != nil {
		return "", fmt.Errorf("scanning drives: %w", err)
	}

	var matches []lines.DriveScan
	for _, d := range drives {
		if d.DiscName == label {
			matches = append(matches, d)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no drive found with disc label %q — was the disc ejected or replaced?", label)
	case 1:
		return model.DiscSourceFromIndex(matches[0].Index), nil
	default:
		var devices []string
		for _, d := range matches {
			devices = append(devices, d.Device)
		}
		return "", fmt.Errorf("multiple drives have a disc labeled %q (%s) — remove the duplicate before backing up", label, strings.Join(devices, ", "))
	}
}

func (e *Engine) resolveDiscRoot(cliRoot string) (string, error) {
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
			if files.IsDir(filepath.Join(candidate, "BDMV", "STREAM")) ||
				files.IsDir(filepath.Join(candidate, "VIDEO_TS")) {
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

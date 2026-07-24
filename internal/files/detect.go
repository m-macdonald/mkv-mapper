package files

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
)

type DiscSource string

func DiscIndex(n int) DiscSource { return DiscSource(fmt.Sprintf("disc:%d", n)) }

func (s DiscSource) IsOptical() bool {
	return strings.HasPrefix(string(s), "disc:") || strings.HasPrefix(string(s), "dev:")
}

type Resolver struct {
	makemkv *makemkv.Client
}

func NewResolver(makemkvClient *makemkv.Client) *Resolver {
	return &Resolver{
		makemkv: makemkvClient,
	}
}

func (r *Resolver) Resolve(ctx context.Context, specified DiscSource) (DiscSource, error) {
	if specified != "" {
		return specified, nil
	}

	drives, err := r.makemkv.ScanDrives(ctx)
	if err != nil {
		return "", fmt.Errorf("scanning drives: %w", err)
	}

	var loaded []lines.DriveScan
	for _, drive := range drives {
		if drive.DiscName != "" {
			loaded = append(loaded, drive)
		}
	}

	switch len(loaded) {
	case 0:
		return "", fmt.Errorf("no disc found in any drive")
	case 1:
		return DiscIndex(loaded[0].Index), nil
	default:
		var names []string
		for _, drive := range loaded {
			names = append(names, fmt.Sprintf("disc:%d (%s)", drive.Index, drive.DriveName))
		}
		return "", fmt.Errorf("multiple discs found, specify one with --disc: %s", strings.Join(names, ", "))
	}
}

func (r *Resolver) ResolveByLabel(ctx context.Context, label string) (DiscSource, error) {
	drives, err := r.makemkv.ScanDrives(ctx)
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
		return DiscIndex(matches[0].Index), nil
	default:
		var devices []string
		for _, d := range matches {
			devices = append(devices, d.Device)
		}
		return "", fmt.Errorf("multiple drives have a disc labeled %q (%s) — remove the duplicate before backing up", label, strings.Join(devices, ", "))
	}
}

func (r *Resolver) ResolveDiscRoot(cliRoot string) (string, error) {
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

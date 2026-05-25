package files

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/diskfs/go-diskfs/backend/file"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
)

func Hash(root string) (string, error) {
	hasher, err := newDiscHasher(root)
	if err != nil {
		return "", err
	}

	return hasher.Hash()
}

func newDiscHasher(root string) (discHasher, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return isoHasher{root: root}, nil
	}
	if isDir(filepath.Join(root, "BDMV", "STREAM")) {
		return bluRayHasher{root: root}, nil
	}
	if isDir(filepath.Join(root, "VIDEO_TS")) {
		return dvdHasher{root: root}, nil
	}

	return nil, errors.New("unrecognized disc structure")
}

type discHasher interface {
	Hash() (string, error)
}

type isoHasher struct {
	root string
}

func (i isoHasher) Hash() (string, error) {
	readOnly := true
	backend, err := file.OpenFromPath(i.root, readOnly)
	if err != nil {
		return "", err
	}
	defer backend.Close()

	stat, err := backend.Stat()
	if err != nil {
		return "", err
	}

	fs, err := iso9660.Read(backend, stat.Size(), 0, 0)
	if err != nil {
		return "", err
	}
	defer fs.Close()

	dirEntries, err := fs.ReadDir("VIDEO_TS")
	if err != nil {
		return "", err
	}

	// All files within the VIDEO_TS dir are used in hash calculation
	filter := func(entry os.DirEntry) bool {
		return !entry.IsDir()
	}

	fileInfos, err := getFiles(dirEntries, filter)
	if err != nil {
		return "", err
	}

	return hashSizes(fileInfos)
}

type dvdHasher struct {
	root string
}

func (d dvdHasher) Hash() (string, error) {
	videoDir := filepath.Join(d.root, "VIDEO_TS")

	dirEntries, err := os.ReadDir(videoDir)
	if err != nil {
		return "", err
	}

	// All files within the VIDEO_TS dir are used in hash calculation
	filter := func(entry os.DirEntry) bool {
		return !entry.IsDir()
	}

	fileInfos, err := getFiles(dirEntries, filter)
	if err != nil {
		return "", err
	}

	return hashSizes(fileInfos)
}

type bluRayHasher struct {
	root string
}

func (b bluRayHasher) Hash() (string, error) {
	streamDir := filepath.Join(b.root, "BDMV", "STREAM")
	dirEntries, err := os.ReadDir(streamDir)
	if err != nil {
		return "", err
	}

	filter := func(entry os.DirEntry) bool {
		if entry.IsDir() {
			return false
		}

		name := entry.Name()

		return strings.EqualFold(filepath.Ext(name), ".m2ts")
	}

	fileInfos, err := getFiles(dirEntries, filter)
	if err != nil {
		return "", err
	}

	return hashSizes(fileInfos)
}

func getFiles(dirEntries []os.DirEntry, filter func(entry os.DirEntry) bool) ([]os.FileInfo, error) {
	var files []os.FileInfo
	for _, e := range dirEntries {
		if !filter(e) {
			continue
		}

		info, err := e.Info()
		if err != nil {
			return nil, err
		}

		files = append(files, info)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	return files, nil
}

func hashSizes(files []os.FileInfo) (string, error) {
	h := md5.New()

	for _, f := range files {
		size := uint64(f.Size())

		if err := binary.Write(h, binary.LittleEndian, size); err != nil {
			return "", err
		}
	}

	return strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil))), nil
}

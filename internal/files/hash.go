package files 

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DiscHasher interface {
	Hash() (string, error)
}

func Hash(root string) (string, error) {
	// TODO: Accommodate disc types beyond Blu-Ray
	streamDir := filepath.Join(root, "BDMV", "STREAM")

	files, err := getStreamFiles(streamDir)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", err
	}

	hash, err := hashSizes(files)
	if err != nil {
		return "", err
	}

	return strings.ToUpper(fmt.Sprintf("%x", hash)), nil
}


func getStreamFiles(streamDir string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(streamDir)
	if err != nil {
		return nil, err
	}

	var files []os.FileInfo

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if !strings.EqualFold(filepath.Ext(name), ".m2ts") {
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

func hashSizes(files []os.FileInfo) ([]byte, error) {
	h := md5.New()

	for _, f := range files {
		size := uint64(f.Size())

		if err := binary.Write(h, binary.LittleEndian, size); err != nil {
			return nil, err	
		}
	}

	return h.Sum(nil), nil
}

// +build windows

package files

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func GetFreeDiskSpace(dir string) (uint64, error) {
	var freeBytesAvailable uint64
	err := windows.GetDiskFreeSpaceEx(
		windows.StringToUTF16Ptr(dir),
		&freeBytesAvailable,
		nil,
		nil)
	if err != nil {
		return 0, err
	}

	return freeBytesAvailable, nil
}

func FilesystemID(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	p, err := windows.UTF16PtrFromString(abs)
	if err != nil {
		return "", err
	}
	buf := make([]uint16, windows.MAX_PATH+1)
	if err := windows.GetVolumePathName(p, &buf[0], uint32(len(buf))); err != nil {
		return "", err
	}
	return strings.ToLower(windows.UTF16ToString(buf)), nil
}

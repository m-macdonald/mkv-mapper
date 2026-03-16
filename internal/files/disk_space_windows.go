// +build windows

package files

import "golang.org/x/sys/windows"

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

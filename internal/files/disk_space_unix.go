// +build unix

package files

import (
	"strconv"

	"golang.org/x/sys/unix"
)

func GetFreeDiskSpace(dir string) (uint64, error) {
	var stat unix.Statfs_t

	err := unix.Statfs(dir, &stat)
	if err != nil {
		return 0, err
	}

	return stat.Bavail * uint64(stat.Bsize), nil
}

// Some setups (e.g. btrfs subvolumes, bind mounts) may report distinct IDs for paths sharing free space.
func FilesystemID(path string) (string, error) {
	var stat unix.Stat_t
	if err := unix.Stat(path, &stat); err != nil {
		return "", err
	}
	return strconv.FormatUint(uint64(stat.Dev), 10), nil
}

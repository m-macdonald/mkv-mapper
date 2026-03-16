// +build unix

package files

import "golang.org/x/sys/unix"

func GetFreeDiskSpace(dir string) (uint64, error) {
	var stat unix.Statfs_t

	err := unix.Statfs(dir, &stat)
	if err != nil {
		return 0, err
	}

	return stat.Bavail * uint64(stat.Bsize), nil
}

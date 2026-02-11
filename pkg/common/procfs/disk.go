package procfs

import (
	"syscall"
)

// ReadDiskUsage returns used and total bytes for the given path (statfs)
func ReadDiskUsage(path string) (used uint64, total uint64, err error) {
	var st syscall.Statfs_t
	if err = syscall.Statfs(path, &st); err != nil {
		return 0, 0, err
	}
	total = uint64(st.Blocks) * uint64(st.Bsize)
	free := uint64(st.Bfree) * uint64(st.Bsize)
	used = total - free
	return used, total, nil
}

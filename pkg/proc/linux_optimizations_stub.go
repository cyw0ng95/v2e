//go:build !linux

package proc

// LinuxOptimizations provides platform-specific performance optimizations
// On non-Linux platforms, these are no-ops
type LinuxOptimizations struct {
	UseZeroCopy bool
	UseSplice   bool
}

// SetSocketOptions is a no-op on non-Linux platforms
func SetSocketOptions(fd int) error {
	return nil
}

// Splice is not available on non-Linux platforms
func Splice(rfd int, wfd int, len int, flags int) (int64, error) {
	return 0, nil
}

// Sendfile is not available on non-Linux platforms
func Sendfile(outfd int, infd int, offset *int64, count int) (int, error) {
	return 0, nil
}

// Madvise operations are no-ops on non-Linux platforms
func MadviseSequential(data []byte) error {
	return nil
}

func MadviseWillNeed(data []byte) error {
	return nil
}

func MadviseDontNeed(data []byte) error {
	return nil
}

// SetThreadAffinity is a no-op on non-Linux platforms
func SetThreadAffinity(cpuSet []int) error {
	return nil
}

// GetCPUCount returns 1 on non-Linux platforms
func GetCPUCount() int {
	return 1
}

// Memcpy performs standard memory copy
func Memcpy(dst, src []byte) {
	copy(dst, src)
}

// PrefetchRead is a no-op on non-Linux platforms
func PrefetchRead(data []byte) {
}

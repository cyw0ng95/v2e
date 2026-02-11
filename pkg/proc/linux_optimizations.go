//go:build linux

package proc

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// LinuxOptimizations provides platform-specific performance optimizations
type LinuxOptimizations struct {
	// UseZeroCopy enables zero-copy operations where possible
	UseZeroCopy bool
	// UseSplice enables splice() syscall for pipe-to-socket transfers
	UseSplice bool
}

// SetSocketOptions configures Linux-specific socket optimizations
func SetSocketOptions(fd int) error {
	// Enable TCP_NODELAY to disable Nagle's algorithm for low latency
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1); err != nil {
		return fmt.Errorf("failed to set TCP_NODELAY: %w", err)
	}

	// Enable TCP_QUICKACK for immediate ACK responses
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, unix.TCP_QUICKACK, 1); err != nil {
		// TCP_QUICKACK might not be supported on older kernels or all configurations
		// Log the error for debugging but don't fail socket setup
		// This option is a performance optimization, not a requirement
		fmt.Printf("warning: TCP_QUICKACK not available on this system: %v\n", err)
	}

	// Set SO_SNDBUF and SO_RCVBUF for optimal buffer sizes
	bufSize := 256 * 1024 // 256KB
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, bufSize); err != nil {
		return err
	}
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, bufSize); err != nil {
		return err
	}

	return nil
}

// Splice performs zero-copy data transfer between file descriptors
// This is significantly faster than read+write for large transfers
func Splice(rfd int, wfd int, len int, flags int) (int64, error) {
	n, err := syscall.Splice(rfd, nil, wfd, nil, len, flags)
	return int64(n), err
}

// Sendfile performs zero-copy file transmission
func Sendfile(outfd int, infd int, offset *int64, count int) (int, error) {
	return syscall.Sendfile(outfd, infd, offset, count)
}

// Madvise provides memory access pattern hints to the kernel
func MadviseSequential(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return unix.Madvise(data, unix.MADV_SEQUENTIAL)
}

// MadviseWillNeed hints that data will be accessed soon
func MadviseWillNeed(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return unix.Madvise(data, unix.MADV_WILLNEED)
}

// MadviseDontNeed hints that data won't be accessed soon
func MadviseDontNeed(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return unix.Madvise(data, unix.MADV_DONTNEED)
}

// SetThreadAffinity pins the current thread to specific CPU cores
func SetThreadAffinity(cpuSet []int) error {
	var mask unix.CPUSet
	for _, cpu := range cpuSet {
		mask.Set(cpu)
	}

	// Pin current thread to specified CPUs
	return unix.SchedSetaffinity(0, &mask)
}

// GetCPUCount returns the number of online CPUs
func GetCPUCount() int {
	var mask unix.CPUSet
	if err := unix.SchedGetaffinity(0, &mask); err != nil {
		return 1
	}
	return mask.Count()
}

// Memcpy performs optimized memory copy using memmove.
// Returns an error if the destination and source lengths don't match.
func Memcpy(dst, src []byte) error {
	if len(dst) != len(src) {
		return fmt.Errorf("memcpy: length mismatch: dst=%d, src=%d", len(dst), len(src))
	}
	if len(dst) == 0 {
		return nil
	}

	// Use unsafe pointer conversion for direct memory copy
	// This bypasses Go's bounds checking for maximum performance
	memmove(unsafe.Pointer(&dst[0]), unsafe.Pointer(&src[0]), uintptr(len(dst)))
	return nil
}

//go:linkname memmove runtime.memmove
func memmove(to, from unsafe.Pointer, n uintptr)

// PrefetchRead hints to prefetch data into CPU cache
func PrefetchRead(data []byte) {
	if len(data) == 0 {
		return
	}
	// Prefetch first cache line
	_ = data[0]
}

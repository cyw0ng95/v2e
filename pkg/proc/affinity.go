package proc

import (
	"fmt"
	"runtime"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

// SetCPUAffinity sets the CPU affinity for the current process based on a hexadecimal mask
// mask: hexadecimal string (e.g., "0x03" for cores 0-1, "0x0F" for cores 0-3)
// Returns an error if the mask is invalid or setting affinity fails
func SetCPUAffinity(mask string) error {
	if mask == "" {
		// No affinity specified, use all cores
		return nil
	}

	// Parse hexadecimal mask
	var affinityMask uint64
	var err error
	
	// Handle "0x" prefix
	if len(mask) > 2 && mask[:2] == "0x" {
		affinityMask, err = strconv.ParseUint(mask[2:], 16, 64)
	} else {
		affinityMask, err = strconv.ParseUint(mask, 16, 64)
	}
	
	if err != nil {
		return fmt.Errorf("invalid CPU affinity mask %q: %w", mask, err)
	}

	if affinityMask == 0 {
		return fmt.Errorf("CPU affinity mask cannot be zero")
	}

	// Lock the goroutine to the current OS thread
	// This ensures affinity is set for the main thread
	runtime.LockOSThread()

	// Build CPU set from mask
	var cpuSet unix.CPUSet
	cpuSet.Zero()

	numCPU := runtime.NumCPU()
	setCPUs := 0
	
	for i := 0; i < numCPU && i < 64; i++ {
		if affinityMask&(1<<uint(i)) != 0 {
			cpuSet.Set(i)
			setCPUs++
		}
	}

	if setCPUs == 0 {
		return fmt.Errorf("no valid CPUs specified in mask %q (system has %d CPUs)", mask, numCPU)
	}

	// Set affinity for current process
	pid := syscall.Getpid()
	if err := unix.SchedSetaffinity(pid, &cpuSet); err != nil {
		return fmt.Errorf("failed to set CPU affinity: %w", err)
	}

	return nil
}

// GetCPUAffinity retrieves the current CPU affinity mask as a hexadecimal string
func GetCPUAffinity() (string, error) {
	var cpuSet unix.CPUSet
	
	pid := syscall.Getpid()
	if err := unix.SchedGetaffinity(pid, &cpuSet); err != nil {
		return "", fmt.Errorf("failed to get CPU affinity: %w", err)
	}

	// Convert CPU set to bitmask
	var mask uint64
	numCPU := runtime.NumCPU()
	
	for i := 0; i < numCPU && i < 64; i++ {
		if cpuSet.IsSet(i) {
			mask |= 1 << uint(i)
		}
	}

	return fmt.Sprintf("0x%x", mask), nil
}

// ParseCPUAffinityMask parses a CPU affinity mask and returns the list of CPU cores
func ParseCPUAffinityMask(mask string) ([]int, error) {
	if mask == "" {
		return nil, nil
	}

	var affinityMask uint64
	var err error
	
	if len(mask) > 2 && mask[:2] == "0x" {
		affinityMask, err = strconv.ParseUint(mask[2:], 16, 64)
	} else {
		affinityMask, err = strconv.ParseUint(mask, 16, 64)
	}
	
	if err != nil {
		return nil, fmt.Errorf("invalid CPU affinity mask %q: %w", mask, err)
	}

	var cpus []int
	for i := 0; i < 64; i++ {
		if affinityMask&(1<<uint(i)) != 0 {
			cpus = append(cpus, i)
		}
	}

	return cpus, nil
}

package subprocess

import (
	"strconv"
)

// These variables are injected at build time via ldflags
var (
	buildRPCInputFD   = "3"    // Default RPC input file descriptor, can be overridden with -ldflags "-X subprocess.buildRPCInputFD=3"
	buildRPCOutputFD  = "4"    // Default RPC output file descriptor, can be overridden with -ldflags "-X subprocess.buildRPCOutputFD=4"
	buildProcCommType = "uds"  // Default communication type: "uds" or "fd"; override with -X subprocess.buildProcCommType=fd
	buildProcAutoExit = "true" // Default auto-exit behavior; override with -X subprocess.buildProcAutoExit=false
)

// DefaultBuildRPCInputFD returns the default RPC input file descriptor based on build configuration
func DefaultBuildRPCInputFD() int {
	fd := 3 // default value
	if buildRPCInputFD != "" {
		if parsed, err := strconv.Atoi(buildRPCInputFD); err == nil {
			fd = parsed
		}
	}
	return fd
}

// DefaultBuildRPCOutputFD returns the default RPC output file descriptor based on build configuration
func DefaultBuildRPCOutputFD() int {
	fd := 4 // default value
	if buildRPCOutputFD != "" {
		if parsed, err := strconv.Atoi(buildRPCOutputFD); err == nil {
			fd = parsed
		}
	}
	return fd
}

// DefaultProcCommType returns the default communication type set at build time
func DefaultProcCommType() string {
	if buildProcCommType == "" {
		return "uds"
	}
	return buildProcCommType
}

// DefaultProcAutoExit returns whether subprocesses should auto-exit when broker exits
func DefaultProcAutoExit() bool {
	return buildProcAutoExit == "true"
}

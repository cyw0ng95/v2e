package subprocess

// These variables are injected at build time via ldflags
var (
	buildProcAutoExit    = "true" // Default auto-exit behavior
	buildProcUDSBasePath = ""     // Default UDS base path
)

// DefaultProcAutoExit returns whether subprocesses should auto-exit when broker exits
func DefaultProcAutoExit() bool {
	return buildProcAutoExit == "true"
}

// DefaultProcUDSBasePath returns the base path used to construct UDS socket paths
func DefaultProcUDSBasePath() string {
	if buildProcUDSBasePath == "" {
		return "/tmp/v2e_uds"
	}
	return buildProcUDSBasePath
}

// DefaultProcCommType returns the default communication type set at build time
// Always returns "uds" since FD pipe support has been removed
func DefaultProcCommType() string {
	return "uds"
}

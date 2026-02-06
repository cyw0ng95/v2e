package transport

// Build-time configuration variable for UDS base path
// Set via: -X 'github.com/cyw0ng95/v2e/cmd/v2broker/transport/manager.buildUDSBasePath=/tmp/v2e_uds'
var buildUDSBasePath string

// buildUDSBasePathValue returns the UDS base path from build-time configuration.
// If the build-time variable is not set (empty string), returns the default value.
func buildUDSBasePathValue() string {
	if buildUDSBasePath == "" {
		return "/tmp/v2e_uds" // default
	}
	return buildUDSBasePath
}

package subprocess

import (
	"github.com/cyw0ng95/v2e/pkg/common"
)

// These variables are injected at build time via ldflags
var (
	buildLogLevel   = "INFO"   // Default log level, can be overridden with -ldflags "-X subprocess.buildLogLevel=DEBUG"
	buildLogDir     = "./logs" // Default log directory, can be overridden with -ldflags "-X subprocess.buildLogDir=/custom/logs"
	buildLogRefresh = "true"   // Default log refresh behavior, can be overridden with -ldflags "-X subprocess.buildLogRefresh=true"
)

// DefaultBuildLogLevel returns the default log level based on build configuration
func DefaultBuildLogLevel() common.LogLevel {
	switch buildLogLevel {
	case "DEBUG":
		common.Info("Using DEBUG log level")
		return common.DebugLevel
	case "INFO":
		common.Info("Using INFO log level")
		return common.InfoLevel
	case "WARN":
		common.Info("Using WARN log level")
		return common.WarnLevel
	case "ERROR":
		common.Info("Using ERROR log level")
		return common.ErrorLevel
	default:
		common.Info("Using default INFO log level")
		return common.InfoLevel // fallback to INFO if invalid value
	}
}

// DefaultBuildLogDir returns the default log directory based on build configuration
func DefaultBuildLogDir() string {
	return buildLogDir
}

// DefaultBuildLogRefresh returns the default log refresh behavior based on build configuration
func DefaultBuildLogRefresh() bool {
	return buildLogRefresh == "true"
}
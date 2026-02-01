package core

import (
	"strings"
)

// These variables are injected at build time via ldflags
var (
	buildBootBins = "access,remote,local,meta,sysmon" // Default boot bins list, can be overridden with -ldflags "-X core.buildBootBins=access,remote"
)

// DefaultBuildBootBins returns the default boot bins list based on build configuration
func DefaultBuildBootBins() []string {
	if buildBootBins != "" {
		bins := strings.Split(buildBootBins, ",")
		// Trim whitespace from each bin
		for i, bin := range bins {
			bins[i] = strings.TrimSpace(bin)
		}
		return bins
	}
	return []string{"access", "remote", "local", "meta", "sysmon"}
}

package common

import (
	"bytes"
	"os/exec"
	"strings"
	"sync"
)

const defaultVersion = "0.1.0"

var (
	versionOnce   sync.Once
	versionCached string
)

// Version returns the project version. It attempts to read the latest git tag
// from the repository (using `git describe --tags --abbrev=0` or falling back to
// the newest tag list). If git is not available or the repository has no tags,
// it returns the compiled-in default version.
func Version() string {
	versionOnce.Do(func() {
		// Try the simple describe command first
		if out, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output(); err == nil {
			versionCached = strings.TrimSpace(string(bytes.TrimSpace(out)))
			return
		}

		// Fallback: list tags sorted by version (requires git >= 2.0)
		if out, err := exec.Command("git", "tag", "--sort=-v:refname").Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			if len(lines) > 0 && lines[0] != "" {
				versionCached = strings.TrimSpace(lines[0])
				return
			}
		}

		// Final fallback: default constant
		versionCached = defaultVersion
	})
	return versionCached
}

package main

// These variables are injected at build time via ldflags
var (
	buildServerAddr = "0.0.0.0:8080" // Default server address, can be overridden with -ldflags "-X main.buildServerAddr=0.0.0.0:9090"
	buildStaticDir  = "website"      // Default static directory, can be overridden with -ldflags "-X main.buildStaticDir=dist"
)

// DefaultServerAddr returns the default server address based on build configuration
func DefaultServerAddr() string {
	return buildServerAddr
}

// DefaultStaticDir returns the default static directory based on build configuration
func DefaultStaticDir() string {
	return buildStaticDir
}
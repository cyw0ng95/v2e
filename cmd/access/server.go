package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	// DefaultRPCTimeout is the default timeout for RPC requests used by tests and main
	DefaultRPCTimeout = 30 * time.Second
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 10 * time.Second
)

// setupRouter creates the Gin router, registers middleware and handlers
func setupRouter(rpcClient *RPCClient, rpcTimeoutSec int, staticDir string) *gin.Engine {
	// Set Gin mode to release (minimal logging)
	gin.SetMode(gin.ReleaseMode)

	// Disable Gin's default logger to prevent stdout pollution
	gin.DefaultWriter = os.Stderr
	gin.DefaultErrorWriter = os.Stderr

	// Create Gin router without default middleware
	router := gin.New()
	// Add recovery middleware but log to stderr
	router.Use(gin.RecoveryWithWriter(os.Stderr))

	// Add CORS middleware
	router.Use(cors.Default())

	// Create RESTful API group
	restful := router.Group("/restful")
	registerHandlers(restful, rpcClient, int(rpcTimeoutSec))

	// Serve static files from configured staticDir if present (Next.js static export)
	outDir := staticDir
	if _, err := os.Stat(outDir); err == nil {
		absOut, _ := filepath.Abs(outDir)
		common.Info("[ACCESS] Serving static files from %s", absOut)

		// Use NoRoute to serve files and fallback to index.html for SPA routes.
		// Avoid registering a catch-all route which conflicts with existing API prefixes.
		router.NoRoute(func(c *gin.Context) {
			// Do not handle API routes here
			if strings.HasPrefix(c.Request.URL.Path, "/restful") {
				c.JSON(http.StatusNotFound, gin.H{"retcode": 404, "message": "not found", "payload": nil})
				return
			}

			// Clean requested path and map to filesystem
			reqPath := path.Clean(c.Request.URL.Path)
			if reqPath == "." || reqPath == "/" {
				c.File(filepath.Join(outDir, "index.html"))
				return
			}

			relPath := strings.TrimPrefix(reqPath, "/")
			fullPath := filepath.Join(outDir, relPath)
			if fi, err := os.Stat(fullPath); err == nil && !fi.IsDir() {
				c.File(fullPath)
				return
			}

			// Fallback to index.html for SPA routing
			c.File(filepath.Join(outDir, "index.html"))
		})
	} else {
		common.Info("[ACCESS] Static dir %s not found, skipping static serving", outDir)
	}

	return router
}

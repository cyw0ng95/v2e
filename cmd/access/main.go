package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	// Set up logger
	logLevel := common.InfoLevel
	if *debug {
		logLevel = common.DebugLevel
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	common.SetLevel(logLevel)

	// Create Gin router
	router := gin.Default()

	// Health check endpoint (minimal endpoint for verification)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	common.Info("Starting access server on %s", addr)

	if err := router.Run(addr); err != nil {
		common.Error("Failed to start server: %v", err)
		os.Exit(1)
	}
}

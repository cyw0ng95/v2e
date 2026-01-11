package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/gin-gonic/gin"
)

// AccessServer holds the access server components
type AccessServer struct {
	broker *proc.Broker
	router *gin.Engine
}

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

	// Initialize broker and spawn backend services
	broker, err := initializeServices()
	if err != nil {
		common.Error("Failed to initialize services: %v", err)
		os.Exit(1)
	}
	defer broker.Shutdown()

	// Create access server
	server := &AccessServer{
		broker: broker,
		router: gin.Default(),
	}

	// Set up routes
	server.setupRoutes()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	common.Info("Starting access server on %s", addr)

	errChan := make(chan error, 1)
	go func() {
		if err := server.router.Run(addr); err != nil {
			errChan <- err
		}
	}()

	// Wait for either error or signal
	select {
	case err := <-errChan:
		common.Error("Server error: %v", err)
		os.Exit(1)
	case <-sigChan:
		common.Info("Received shutdown signal, cleaning up...")
	}
}

// initializeServices spawns broker and backend services
func initializeServices() (*proc.Broker, error) {
	// Create broker
	broker := proc.NewBroker()

	// Get the path to the built binaries
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	baseDir := filepath.Dir(exePath)

	// Find module root for go run paths
	moduleRoot, err := findModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find module root: %w", err)
	}

	// Set database path
	dbPath := os.Getenv("CVE_DB_PATH")
	if dbPath == "" {
		dbPath = "cve.db"
	}
	os.Setenv("CVE_DB_PATH", dbPath)

	// Spawn CVE meta service (which will spawn cve-local and cve-remote)
	cveMetaPath := filepath.Join(baseDir, "cve-meta")
	if _, err := os.Stat(cveMetaPath); os.IsNotExist(err) {
		cveMetaPath = "go"
		cveMetaSource := filepath.Join(moduleRoot, "cmd", "cve-meta")
		_, err = broker.SpawnRPC("cve-meta", cveMetaPath, "run", cveMetaSource)
	} else {
		_, err = broker.SpawnRPC("cve-meta", cveMetaPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to spawn cve-meta: %w", err)
	}

	// Give cve-meta time to initialize and spawn its services
	// We drain ready events to avoid blocking the message channel
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Wait for at least one ready event (from cve-meta)
	// cve-meta will also receive ready events from its subprocesses
	readyReceived := false
	timeout := time.After(5 * time.Second)
	
	for !readyReceived {
		select {
		case <-timeout:
			// We've waited long enough, proceed anyway
			common.Warn("Timeout waiting for ready event, proceeding anyway")
			return broker, nil
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while waiting for services")
		default:
			msgCtx, msgCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			msg, err := broker.ReceiveMessage(msgCtx)
			msgCancel()

			if err == nil && msg.Type == proc.MessageTypeEvent && msg.ID == "subprocess_ready" {
				readyReceived = true
				common.Debug("Received ready event from subprocess")
			}
		}
	}

	// Give a bit more time for full initialization
	time.Sleep(500 * time.Millisecond)

	return broker, nil
}

// findModuleRoot finds the Go module root directory by looking for go.mod
func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			return "", fmt.Errorf("could not find go.mod in directory tree")
		}
		dir = parent
	}
}

// setupRoutes configures all HTTP routes
func (s *AccessServer) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.handleHealth)

	// RESTful API routes
	restful := s.router.Group("/restful")
	{
		// CVE operations
		cve := restful.Group("/cve")
		{
			cve.GET("/count", s.handleGetCVECount)
			cve.GET("/:id", s.handleGetCVE)
			cve.POST("/batch", s.handleBatchFetchCVEs)
		}

		// Process operations (for debugging/monitoring)
		processes := restful.Group("/processes")
		{
			processes.GET("", s.handleListProcesses)
			processes.GET("/:id", s.handleGetProcess)
		}
	}
}

// handleHealth handles the health check endpoint
func (s *AccessServer) handleHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

// handleGetCVECount gets the total CVE count from NVD
func (s *AccessServer) handleGetCVECount(c *gin.Context) {
	// Send RPC request to cve-meta
	req, err := proc.NewRequestMessage("RPCGetRemoteCVECount", map[string]interface{}{})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create request: %v", err)})
		return
	}

	if err := s.broker.SendToProcess("cve-meta", req); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to send request: %v", err)})
		return
	}

	// Wait for response
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	for {
		msg, err := s.broker.ReceiveMessage(ctx)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("timeout waiting for response: %v", err)})
			return
		}

		// Skip event messages
		if msg.Type == proc.MessageTypeEvent {
			continue
		}

		if msg.Type == proc.MessageTypeResponse && msg.ID == "RPCGetRemoteCVECount" {
			var response map[string]interface{}
			if err := msg.UnmarshalPayload(&response); err != nil {
				c.JSON(500, gin.H{"error": fmt.Sprintf("failed to parse response: %v", err)})
				return
			}
			c.JSON(200, response)
			return
		}

		if msg.Type == proc.MessageTypeError {
			c.JSON(500, gin.H{"error": msg.Error})
			return
		}
	}
}

// handleGetCVE fetches a CVE by ID
func (s *AccessServer) handleGetCVE(c *gin.Context) {
	cveID := c.Param("id")
	if cveID == "" {
		c.JSON(400, gin.H{"error": "CVE ID is required"})
		return
	}

	// Send RPC request to cve-meta
	req, err := proc.NewRequestMessage("RPCFetchAndStoreCVE", map[string]string{
		"cve_id": cveID,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create request: %v", err)})
		return
	}

	if err := s.broker.SendToProcess("cve-meta", req); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to send request: %v", err)})
		return
	}

	// Wait for response
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	for {
		msg, err := s.broker.ReceiveMessage(ctx)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("timeout waiting for response: %v", err)})
			return
		}

		// Skip event messages
		if msg.Type == proc.MessageTypeEvent {
			continue
		}

		if msg.Type == proc.MessageTypeResponse && msg.ID == "RPCFetchAndStoreCVE" {
			var response map[string]interface{}
			if err := msg.UnmarshalPayload(&response); err != nil {
				c.JSON(500, gin.H{"error": fmt.Sprintf("failed to parse response: %v", err)})
				return
			}
			c.JSON(200, response)
			return
		}

		if msg.Type == proc.MessageTypeError {
			c.JSON(500, gin.H{"error": msg.Error})
			return
		}
	}
}

// handleBatchFetchCVEs fetches multiple CVEs in batch
func (s *AccessServer) handleBatchFetchCVEs(c *gin.Context) {
	var requestBody struct {
		CVEIDs []string `json:"cve_ids"`
	}

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid request body: %v", err)})
		return
	}

	if len(requestBody.CVEIDs) == 0 {
		c.JSON(400, gin.H{"error": "cve_ids is required and must not be empty"})
		return
	}

	// Send RPC request to cve-meta
	req, err := proc.NewRequestMessage("RPCBatchFetchCVEs", map[string]interface{}{
		"cve_ids": requestBody.CVEIDs,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create request: %v", err)})
		return
	}

	if err := s.broker.SendToProcess("cve-meta", req); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to send request: %v", err)})
		return
	}

	// Wait for response
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	for {
		msg, err := s.broker.ReceiveMessage(ctx)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("timeout waiting for response: %v", err)})
			return
		}

		// Skip event messages
		if msg.Type == proc.MessageTypeEvent {
			continue
		}

		if msg.Type == proc.MessageTypeResponse && msg.ID == "RPCBatchFetchCVEs" {
			var response map[string]interface{}
			if err := msg.UnmarshalPayload(&response); err != nil {
				c.JSON(500, gin.H{"error": fmt.Sprintf("failed to parse response: %v", err)})
				return
			}
			c.JSON(200, response)
			return
		}

		if msg.Type == proc.MessageTypeError {
			c.JSON(500, gin.H{"error": msg.Error})
			return
		}
	}
}

// handleListProcesses lists all processes managed by the broker
func (s *AccessServer) handleListProcesses(c *gin.Context) {
	processes := s.broker.ListProcesses()
	c.JSON(200, gin.H{
		"processes": processes,
		"count":     len(processes),
	})
}

// handleGetProcess gets information about a specific process
func (s *AccessServer) handleGetProcess(c *gin.Context) {
	processID := c.Param("id")
	if processID == "" {
		c.JSON(400, gin.H{"error": "Process ID is required"})
		return
	}

	process, err := s.broker.GetProcess(processID)
	if err != nil {
		c.JSON(404, gin.H{"error": fmt.Sprintf("process not found: %v", err)})
		return
	}

	c.JSON(200, process)
}

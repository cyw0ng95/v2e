package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/gin-gonic/gin"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 10 * time.Second
)

func main() {
	// Load configuration
	config, err := common.LoadConfig("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Set default address if not configured
	address := "0.0.0.0:8080"
	if config.Server.Address != "" {
		address = config.Server.Address
	}

	// Set up logger with dual output (stdout + file) if log file is configured
	var logOutput io.Writer
	if config.Broker.LogFile != "" {
		logFile, err := os.OpenFile(config.Broker.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
			os.Exit(1)
		}
		defer logFile.Close()
		logOutput = io.MultiWriter(os.Stdout, logFile)
	} else {
		logOutput = os.Stdout
	}

	// Set default logger output
	common.SetOutput(logOutput)
	common.SetLevel(common.InfoLevel)

	// Create broker instance - access service acts as the central entry point
	broker := NewBroker()
	defer broker.Shutdown()

	// Set up broker logger
	brokerLogger := common.NewLogger(logOutput, "[BROKER] ", common.InfoLevel)
	broker.SetLogger(brokerLogger)

	// Load processes from configuration
	if err := broker.LoadProcessesFromConfig(config); err != nil {
		common.Error("Error loading processes from config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.Default()

	// Create RESTful API group
	restful := router.Group("/restful")
	{
		// Health check endpoint
		restful.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})

		// Process management endpoints
		restful.POST("/processes", handleSpawnProcess(broker))
		restful.GET("/processes", handleListProcesses(broker))
		restful.GET("/processes/:id", handleGetProcess(broker))
		restful.DELETE("/processes/:id", handleKillProcess(broker))

		// Statistics endpoint
		restful.GET("/stats", handleGetStats(broker))
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		common.Info("Starting access service on %s", address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Error("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	common.Info("Shutting down access service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		common.Error("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	common.Info("Access service stopped")
}

// handleSpawnProcess handles POST /processes - spawn a new process
func handleSpawnProcess(broker *Broker) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID      string   `json:"id" binding:"required"`
			Command string   `json:"command" binding:"required"`
			Args    []string `json:"args"`
			RPC     bool     `json:"rpc"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var info *ProcessInfo
		var err error

		if req.RPC {
			info, err = broker.SpawnRPC(req.ID, req.Command, req.Args...)
		} else {
			info, err = broker.Spawn(req.ID, req.Command, req.Args...)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      info.ID,
			"pid":     info.PID,
			"command": info.Command,
			"args":    info.Args,
			"status":  info.Status,
		})
	}
}

// handleListProcesses handles GET /processes - list all processes
func handleListProcesses(broker *Broker) gin.HandlerFunc {
	return func(c *gin.Context) {
		processes := broker.ListProcesses()

		processData := make([]gin.H, len(processes))
		for i, p := range processes {
			processData[i] = gin.H{
				"id":      p.ID,
				"pid":     p.PID,
				"command": p.Command,
				"args":    p.Args,
				"status":  p.Status,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"processes": processData,
			"count":     len(processes),
		})
	}
}

// handleGetProcess handles GET /processes/:id - get process info
func handleGetProcess(broker *Broker) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		info, err := broker.GetProcess(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      info.ID,
			"pid":     info.PID,
			"command": info.Command,
			"args":    info.Args,
			"status":  info.Status,
		})
	}
}

// handleKillProcess handles DELETE /processes/:id - kill a process
func handleKillProcess(broker *Broker) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		err := broker.Kill(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"id":      id,
		})
	}
}

// handleGetStats handles GET /stats - get broker statistics
func handleGetStats(broker *Broker) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := broker.GetMessageStats()
		c.JSON(http.StatusOK, gin.H{
			"total_sent":      stats.TotalSent,
			"total_received":  stats.TotalReceived,
			"request_count":   stats.RequestCount,
			"response_count":  stats.ResponseCount,
			"event_count":     stats.EventCount,
			"error_count":     stats.ErrorCount,
			"first_message":   stats.FirstMessageTime,
			"last_message":    stats.LastMessageTime,
		})
	}
}

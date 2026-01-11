package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/gin-gonic/gin"
)

const (
	// DefaultRPCTimeout is the default timeout for RPC requests
	DefaultRPCTimeout = 30 * time.Second
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

	// Set up logger
	common.SetLevel(common.InfoLevel)

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create broker instance for backend communication
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Load processes from configuration
	if err := broker.LoadProcessesFromConfig(config); err != nil {
		common.Error("Error loading processes from config: %v", err)
	}

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

		// List all processes
		restful.GET("/processes", func(c *gin.Context) {
			processes := broker.ListProcesses()
			result := make([]map[string]interface{}, 0, len(processes))
			for _, p := range processes {
				result = append(result, map[string]interface{}{
					"id":        p.ID,
					"pid":       p.PID,
					"command":   p.Command,
					"status":    p.Status,
					"exit_code": p.ExitCode,
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"processes": result,
				"count":     len(result),
			})
		})

		// Get process details
		restful.GET("/processes/:id", func(c *gin.Context) {
			id := c.Param("id")
			info, err := broker.GetProcess(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": fmt.Sprintf("Process not found: %s", id),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"id":        info.ID,
				"pid":       info.PID,
				"command":   info.Command,
				"status":    info.Status,
				"exit_code": info.ExitCode,
			})
		})

		// Spawn a new process
		restful.POST("/processes", func(c *gin.Context) {
			var req struct {
				ID      string   `json:"id" binding:"required"`
				Command string   `json:"command" binding:"required"`
				Args    []string `json:"args"`
				RPC     bool     `json:"rpc"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Invalid request: %v", err),
				})
				return
			}

			var info *proc.ProcessInfo
			var spawnErr error
			if req.RPC {
				info, spawnErr = broker.SpawnRPC(req.ID, req.Command, req.Args...)
			} else {
				info, spawnErr = broker.Spawn(req.ID, req.Command, req.Args...)
			}

			if spawnErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to spawn process: %v", spawnErr),
				})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"id":      info.ID,
				"pid":     info.PID,
				"command": info.Command,
				"status":  info.Status,
			})
		})

		// Kill a process
		restful.DELETE("/processes/:id", func(c *gin.Context) {
			id := c.Param("id")
			if err := broker.Kill(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to kill process: %v", err),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"id":      id,
			})
		})

		// Get broker message statistics
		restful.GET("/stats", func(c *gin.Context) {
			stats := broker.GetMessageStats()
			c.JSON(http.StatusOK, gin.H{
				"total_sent":         stats.TotalSent,
				"total_received":     stats.TotalReceived,
				"request_count":      stats.RequestCount,
				"response_count":     stats.ResponseCount,
				"event_count":        stats.EventCount,
				"error_count":        stats.ErrorCount,
				"first_message_time": stats.FirstMessageTime,
				"last_message_time":  stats.LastMessageTime,
			})
		})

		// Forward RPC requests to processes
		restful.POST("/rpc/:process_id/:endpoint", func(c *gin.Context) {
			processID := c.Param("process_id")
			endpoint := c.Param("endpoint")

			// Read request body
			var payload interface{}
			if err := c.ShouldBindJSON(&payload); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Invalid request body: %v", err),
				})
				return
			}

			// Create RPC request message
			msg, err := proc.NewRequestMessage(endpoint, payload)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to create RPC message: %v", err),
				})
				return
			}

			// Send message to the process
			if err := broker.SendToProcess(processID, msg); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to send RPC message: %v", err),
				})
				return
			}

			// Wait for response with timeout
			// Note: This is a simplified implementation. In production, you would want
			// to use a request-response correlation mechanism to match responses to requests
			ctx, cancel := context.WithTimeout(context.Background(), DefaultRPCTimeout)
			defer cancel()

			respMsg, err := broker.ReceiveMessage(ctx)
			if err != nil {
				c.JSON(http.StatusGatewayTimeout, gin.H{
					"error": fmt.Sprintf("Timeout waiting for response: %v", err),
				})
				return
			}

			// Handle response or error
			if respMsg.Type == proc.MessageTypeError {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": respMsg.Error,
				})
				return
			}

			// Parse and return the response payload
			var result interface{}
			if err := sonic.Unmarshal(respMsg.Payload, &result); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to parse response: %v", err),
				})
				return
			}

			c.JSON(http.StatusOK, result)
		})
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		common.Info("Starting access server on %s", address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Error("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	common.Info("Shutting down access server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		common.Error("Server shutdown error: %v", err)
	}

	common.Info("Access server stopped")
}

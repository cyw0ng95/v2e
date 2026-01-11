package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/gin-gonic/gin"
)

// Global broker instance for RPC communication
var globalBroker *proc.Broker

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	debug := flag.Bool("debug", false, "Enable debug mode")
	cveMetaPath := flag.String("cve-meta", "", "Path to cve-meta binary (required for CVE endpoints)")
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

	// Initialize broker if path provided
	if *cveMetaPath != "" {
		common.Info("Initializing CVE meta service...")
		broker, err := initializeBroker(*cveMetaPath)
		if err != nil {
			common.Error("Failed to initialize cve-meta: %v", err)
			common.Warn("CVE endpoints will not be available")
		} else {
			globalBroker = broker
			defer globalBroker.Shutdown()
			common.Info("CVE meta service initialized")
		}
	} else {
		common.Warn("No cve-meta path provided, CVE endpoints will not be available")
	}

	// Create Gin router
	router := gin.Default()

	// Health check endpoint (minimal endpoint for verification)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// CVE endpoints (only available if broker is connected)
	if globalBroker != nil {
		router.GET("/cve/:id", getCVEByID)
		router.GET("/cve/count", getCVECount)
	}

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	common.Info("Starting access server on %s", addr)

	if err := router.Run(addr); err != nil {
		common.Error("Failed to start server: %v", err)
		os.Exit(1)
	}
}

// initializeBroker initializes broker and spawns cve-meta service
func initializeBroker(cveMetaPath string) (*proc.Broker, error) {
	broker := proc.NewBroker()

	// Spawn cve-meta service via broker
	_, err := broker.SpawnRPC("cve-meta", cveMetaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to spawn cve-meta: %w", err)
	}

	// Wait for cve-meta to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	readyCount := 0
	for readyCount < 1 {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for cve-meta to be ready")
		default:
			msgCtx, msgCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			msg, err := broker.ReceiveMessage(msgCtx)
			msgCancel()

			if err == nil && msg.Type == proc.MessageTypeEvent && msg.ID == "subprocess_ready" {
				readyCount++
			}
		}
	}

	return broker, nil
}

// getCVEByID handles GET /cve/:id
func getCVEByID(c *gin.Context) {
	if globalBroker == nil {
		c.JSON(503, gin.H{
			"error": "CVE service not available",
		})
		return
	}

	cveID := c.Param("id")
	if cveID == "" {
		c.JSON(400, gin.H{
			"error": "CVE ID is required",
		})
		return
	}

	common.Info("Fetching CVE: %s", cveID)

	// Send RPC request to cve-meta
	reqMsg, _ := proc.NewRequestMessage("RPCFetchAndStoreCVE", map[string]string{
		"cve_id": cveID,
	})
	if err := globalBroker.SendToProcess("cve-meta", reqMsg); err != nil {
		common.Error("Failed to send request to cve-meta: %v", err)
		c.JSON(500, gin.H{
			"error": "Failed to communicate with CVE service",
		})
		return
	}

	// Wait for response
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	for {
		respMsg, err := globalBroker.ReceiveMessage(ctx)
		if err != nil {
			common.Error("Failed to receive response from cve-meta: %v", err)
			c.JSON(500, gin.H{
				"error": "CVE service timeout",
			})
			return
		}

		// Skip event messages
		if respMsg.Type == proc.MessageTypeEvent {
			continue
		}

		if respMsg.Type == proc.MessageTypeResponse && respMsg.ID == "RPCFetchAndStoreCVE" {
			var result map[string]interface{}
			if err := respMsg.UnmarshalPayload(&result); err != nil {
				common.Error("Failed to unmarshal response: %v", err)
				c.JSON(500, gin.H{
					"error": "Failed to parse CVE data",
				})
				return
			}

			c.JSON(200, result)
			return
		}

		if respMsg.Type == proc.MessageTypeError {
			common.Error("Error from cve-meta: %s", respMsg.Error)
			c.JSON(500, gin.H{
				"error": respMsg.Error,
			})
			return
		}
	}
}

// getCVECount handles GET /cve/count
func getCVECount(c *gin.Context) {
	if globalBroker == nil {
		c.JSON(503, gin.H{
			"error": "CVE service not available",
		})
		return
	}

	common.Info("Fetching CVE count")

	// Send RPC request to cve-meta
	reqMsg, _ := proc.NewRequestMessage("RPCGetRemoteCVECount", map[string]interface{}{})
	if err := globalBroker.SendToProcess("cve-meta", reqMsg); err != nil {
		common.Error("Failed to send request to cve-meta: %v", err)
		c.JSON(500, gin.H{
			"error": "Failed to communicate with CVE service",
		})
		return
	}

	// Wait for response
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		respMsg, err := globalBroker.ReceiveMessage(ctx)
		if err != nil {
			common.Error("Failed to receive response from cve-meta: %v", err)
			c.JSON(500, gin.H{
				"error": "CVE service timeout",
			})
			return
		}

		// Skip event messages
		if respMsg.Type == proc.MessageTypeEvent {
			continue
		}

		if respMsg.Type == proc.MessageTypeResponse && respMsg.ID == "RPCGetRemoteCVECount" {
			var result map[string]interface{}
			if err := respMsg.UnmarshalPayload(&result); err != nil {
				common.Error("Failed to unmarshal response: %v", err)
				c.JSON(500, gin.H{
					"error": "Failed to parse CVE count",
				})
				return
			}

			c.JSON(200, result)
			return
		}

		if respMsg.Type == proc.MessageTypeError {
			common.Error("Error from cve-meta: %s", respMsg.Error)
			c.JSON(500, gin.H{
				"error": respMsg.Error,
			})
			return
		}
	}
}

package main

import (
	"fmt"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/gin-gonic/gin"
)

// registerHandlers registers the REST endpoints on the provided router group
func registerHandlers(restful *gin.RouterGroup, rpcClient *RPCClient, rpcTimeoutSec int) {
	// Health check endpoint
	restful.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Generic RPC forwarding endpoint
	restful.POST("/rpc", func(c *gin.Context) {
		// Parse request body
		var request struct {
			Method string                 `json:"method" binding:"required"`
			Params map[string]interface{} `json:"params"`
			Target string                 `json:"target"` // Optional target process (defaults to "broker")
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"retcode": 400,
				"message": fmt.Sprintf("Invalid request: %v", err),
				"payload": nil,
			})
			return
		}

		// Default target to broker if not specified
		target := request.Target
		if target == "" {
			target = "broker"
		}

		// Forward RPC request to target process (use configured rpc timeout)
		response, err := rpcClient.InvokeRPCWithTarget(c.Request.Context(), target, request.Method, request.Params)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"retcode": 500,
				"message": fmt.Sprintf("RPC error: %v", err),
				"payload": nil,
			})
			return
		}

		// Check response type
		if response.Type == subprocess.MessageTypeError {
			c.JSON(http.StatusOK, gin.H{
				"retcode": 500,
				"message": response.Error,
				"payload": nil,
			})
			return
		}

		// Parse payload
		var payload interface{}
		if response.Payload != nil {
			if err := sonic.Unmarshal(response.Payload, &payload); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"retcode": 500,
					"message": fmt.Sprintf("Failed to parse response: %v", err),
					"payload": nil,
				})
				return
			}
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{
			"retcode": 0,
			"message": "success",
			"payload": payload,
		})
	})
}

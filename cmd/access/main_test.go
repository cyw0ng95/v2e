package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/gin-gonic/gin"
)

func TestHealthEndpoint(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create broker instance
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Create router
	router := gin.Default()
	restful := router.Group("/restful")
	{
		restful.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})
	}

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/restful/health", nil)
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestListProcesses(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create broker instance
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Spawn a test process
	_, err := broker.Spawn("test-echo", "echo", "hello")
	if err != nil {
		t.Fatalf("Failed to spawn test process: %v", err)
	}

	// Create router
	router := gin.Default()
	restful := router.Group("/restful")
	{
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
	}

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/restful/processes", nil)
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	count, ok := response["count"].(float64)
	if !ok {
		t.Errorf("Expected count to be a number")
	}

	if count < 1 {
		t.Errorf("Expected at least 1 process, got %v", count)
	}
}

func TestSpawnProcess(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	common.SetLevel(common.ErrorLevel) // Suppress logs during testing

	// Create broker instance
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Create router
	router := gin.Default()
	restful := router.Group("/restful")
	{
		restful.POST("/processes", func(c *gin.Context) {
			var req struct {
				ID      string   `json:"id" binding:"required"`
				Command string   `json:"command" binding:"required"`
				Args    []string `json:"args"`
				RPC     bool     `json:"rpc"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
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
					"error": spawnErr.Error(),
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
	}

	// Create request
	requestBody := map[string]interface{}{
		"id":      "test-process",
		"command": "echo",
		"args":    []string{"test"},
	}
	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/restful/processes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["id"] != "test-process" {
		t.Errorf("Expected process ID 'test-process', got '%v'", response["id"])
	}

	if response["command"] != "echo" {
		t.Errorf("Expected command 'echo', got '%v'", response["command"])
	}
}

func TestGetProcess(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	common.SetLevel(common.ErrorLevel) // Suppress logs during testing

	// Create broker instance
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Spawn a test process
	_, err := broker.Spawn("test-get", "echo", "hello")
	if err != nil {
		t.Fatalf("Failed to spawn test process: %v", err)
	}

	// Create router
	router := gin.Default()
	restful := router.Group("/restful")
	{
		restful.GET("/processes/:id", func(c *gin.Context) {
			id := c.Param("id")
			info, err := broker.GetProcess(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": err.Error(),
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
	}

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/restful/processes/test-get", nil)
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["id"] != "test-get" {
		t.Errorf("Expected process ID 'test-get', got '%v'", response["id"])
	}
}

func TestGetProcessNotFound(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	common.SetLevel(common.ErrorLevel) // Suppress logs during testing

	// Create broker instance
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Create router
	router := gin.Default()
	restful := router.Group("/restful")
	{
		restful.GET("/processes/:id", func(c *gin.Context) {
			id := c.Param("id")
			info, err := broker.GetProcess(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": err.Error(),
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
	}

	// Create request for non-existent process
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/restful/processes/non-existent", nil)
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

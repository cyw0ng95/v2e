package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterMiddleware_BasicLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &RateLimiterConfig{
		MaxTokens:       3,                   // Only allow 3 requests
		RefillInterval:  time.Second * 10,     // Slow refill
		CleanupInterval: time.Minute,
		TrustedProxies:  []string{},
		ExcludedPaths:   []string{},
	}

	router := gin.New()
	router.Use(RateLimiterMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make 3 successful requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("request %d should succeed, got status %d", i, w.Code)
		}
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("4th request should be rate limited, got status %d", w.Code)
	}
}

func TestRateLimiterMiddleware_DifferentClients(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &RateLimiterConfig{
		MaxTokens:       2,
		RefillInterval:  time.Second * 10,
		CleanupInterval: time.Minute,
		TrustedProxies:  []string{},
		ExcludedPaths:   []string{},
	}

	router := gin.New()
	router.Use(RateLimiterMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Client 1 makes 2 requests
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("client1 request %d should succeed", i)
		}
	}

	// Client 1 should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatal("client1 should be rate limited")
	}

	// Client 2 should still work
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("client2 should not be rate limited")
	}
}

func TestRateLimiterMiddleware_ExcludedPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &RateLimiterConfig{
		MaxTokens:       1,
		RefillInterval:  time.Second * 10,
		CleanupInterval: time.Minute,
		TrustedProxies:  []string{},
		ExcludedPaths:   []string{"/health", "/metrics"},
	}

	router := gin.New()
	router.Use(RateLimiterMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Exhaust rate limit on /test
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("first request should succeed")
	}

	// /test should now be rate limited
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatal("/test should be rate limited")
	}

	// /health should NOT be rate limited
	req = httptest.NewRequest("GET", "/health", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("/health should not be rate limited")
	}
}

func TestRateLimiterMiddleware_TrustedProxies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &RateLimiterConfig{
		MaxTokens:       1,
		RefillInterval:  time.Second * 10,
		CleanupInterval: time.Minute,
		TrustedProxies:  []string{"127.0.0.1", "::1"},
		ExcludedPaths:   []string{},
	}

	router := gin.New()
	router.Use(RateLimiterMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Trusted proxy should bypass rate limiting
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("trusted proxy request %d should succeed", i)
		}
	}
}

func TestRateLimiterMiddleware_Concurrent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &RateLimiterConfig{
		MaxTokens:       100,
		RefillInterval:  time.Millisecond * 100,
		CleanupInterval: time.Minute,
		TrustedProxies:  []string{},
		ExcludedPaths:   []string{},
	}

	router := gin.New()
	router.Use(RateLimiterMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	var wg sync.WaitGroup
	successCount := make(chan int, 200)

	// Launch 5 goroutines making 50 requests each
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			clientIP := "192.168.1." + string(rune(clientID))
			for j := 0; j < 50; j++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = clientIP
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				if w.Code == http.StatusOK {
					successCount <- 1
				}
			}
		}(i)
	}

	wg.Wait()
	close(successCount)

	count := 0
	for range successCount {
		count++
	}

	// Each client should be limited to 100 requests
	if count > 500 {
		t.Fatalf("expected max 500 successful requests, got %d", count)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name         string
		xForwarded   string
		xReal        string
		remoteAddr   string
		expectedIP   string
	}{
		{
			name:       "use remote addr",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "use x-real-ip",
			xReal:      "10.0.0.1",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "10.0.0.1",
		},
		{
			name:       "use x-forwarded-for first IP",
			xForwarded: "10.0.0.2, 10.0.0.3",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "10.0.0.2",
		},
		{
			name:       "x-forwarded-for takes precedence over x-real-ip",
			xForwarded: "10.0.0.4",
			xReal:      "10.0.0.5",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "10.0.0.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			var capturedIP string
			router.GET("/test", func(c *gin.Context) {
				capturedIP = getClientIP(c)
				c.JSON(http.StatusOK, gin.H{})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.xForwarded != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwarded)
			}
			if tt.xReal != "" {
				req.Header.Set("X-Real-IP", tt.xReal)
			}
			req.RemoteAddr = tt.remoteAddr

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if capturedIP != tt.expectedIP {
				t.Errorf("expected IP %s, got %s", tt.expectedIP, capturedIP)
			}
		})
	}
}

func TestIsTrustedProxy(t *testing.T) {
	trusted := []string{"127.0.0.1", "::1", "10.0.0.1"}

	tests := []struct {
		ip       string
		trusted  bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"10.0.0.1", true},
		{"192.168.1.1", false},
		{"8.8.8.8", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := isTrustedProxy(tt.ip, trusted)
			if result != tt.trusted {
				t.Errorf("isTrustedProxy(%s) = %v, want %v", tt.ip, result, tt.trusted)
			}
		})
	}
}

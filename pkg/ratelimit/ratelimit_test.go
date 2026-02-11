package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestTokenBucket_Allow_Basic(t *testing.T) {
	tb := NewTokenBucket(5, time.Millisecond*100)

	// Should allow 5 requests immediately
	for i := 0; i < 5; i++ {
		if !tb.Allow() {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	// 6th request should be denied
	if tb.Allow() {
		t.Fatal("6th request should be denied")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	tb := NewTokenBucket(1, time.Millisecond*50)

	// First request should be allowed
	if !tb.Allow() {
		t.Fatal("first request should be allowed")
	}

	// Second request should be denied
	if tb.Allow() {
		t.Fatal("second request should be denied immediately")
	}

	// Wait for refill
	time.Sleep(time.Millisecond * 60)

	// Now should be allowed again
	if !tb.Allow() {
		t.Fatal("request should be allowed after refill")
	}
}

func TestTokenBucket_MaxCapacity(t *testing.T) {
	tb := NewTokenBucket(3, time.Millisecond*100)

	// Use all tokens
	for i := 0; i < 3; i++ {
		if !tb.Allow() {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	// Wait for multiple refills
	time.Sleep(time.Millisecond * 350)

	// Should have refilled to max, not exceeded it
	allowedCount := 0
	for i := 0; i < 5; i++ {
		if tb.Allow() {
			allowedCount++
		}
	}

	if allowedCount != 3 {
		t.Fatalf("expected 3 requests allowed after refill, got %d", allowedCount)
	}
}

func TestClientLimiter_Allow(t *testing.T) {
	cl := NewClientLimiter(2, time.Second)

	// Same client should be limited
	if !cl.Allow("client1") {
		t.Fatal("first request from client1 should be allowed")
	}
	if !cl.Allow("client1") {
		t.Fatal("second request from client1 should be allowed")
	}
	if cl.Allow("client1") {
		t.Fatal("third request from client1 should be denied")
	}

	// Different client should have separate limit
	if !cl.Allow("client2") {
		t.Fatal("first request from client2 should be allowed")
	}
}

func TestClientLimiter_Concurrent(t *testing.T) {
	cl := NewClientLimiter(100, time.Millisecond)

	var wg sync.WaitGroup
	allowed := make(chan bool, 1000)

	// Launch 10 goroutines making 100 requests each
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			clientKey := string(rune(clientID))
			for j := 0; j < 100; j++ {
				allowed <- cl.Allow(clientKey)
			}
		}(i)
	}

	wg.Wait()
	close(allowed)

	count := 0
	for range allowed {
		count++
	}

	// All requests should complete without race
	if count != 1000 {
		t.Fatalf("expected 1000 total requests, got %d", count)
	}
}

func TestClientLimiter_Cleanup(t *testing.T) {
	cl := NewClientLimiter(1, time.Second)

	// Add many clients
	for i := 0; i < 100; i++ {
		cl.Allow(string(rune(i)))
	}

	// Cleanup should not panic
	cl.Cleanup(time.Hour)

	// Limiter should still work
	if !cl.Allow("new_client") {
		t.Fatal("new client should be allowed after cleanup")
	}
}

func TestClientLimiter_CleanupTrigger(t *testing.T) {
	cl := NewClientLimiter(1, time.Second)

	// Add more than 10000 clients to trigger cleanup
	for i := 0; i < 10001; i++ {
		cl.Allow(string(rune(i % 1000))) // Reuse keys to stay within valid range
	}

	// This should trigger cleanup
	cl.Cleanup(time.Hour)

	// New client should still work
	if !cl.Allow("new_client") {
		t.Fatal("new client should be allowed after cleanup")
	}
}

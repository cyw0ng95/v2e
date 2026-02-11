package taskflow

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

func TestLeakDetector_Track_Release(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLeakDetector_Track_Release", nil, func(t *testing.T, tx *gorm.DB) {
		ld := NewLeakDetectorWithDefaults()

		ref := ld.Track("obj-1", PoolSmall, 1000)
		if ref == nil {
			t.Fatal("expected non-nil reference")
		}
		if ref.ID != "obj-1" {
			t.Errorf("expected ID obj-1, got %s", ref.ID)
		}

		ld.Release("obj-1")

		stats := ld.GetStats()
		if stats["active_count"].(int) != 0 {
			t.Errorf("expected 0 active, got %d", stats["active_count"])
		}
	})
}

func TestLeakDetector_LeakDetection(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLeakDetector_LeakDetection", nil, func(t *testing.T, tx *gorm.DB) {
		config := DefaultLeakDetectorConfig()
		config.MaxLifetime = 100 * time.Millisecond
		config.CheckInterval = 50 * time.Millisecond
		config.Threshold = 1

		ld := NewLeakDetector(config)
		var leakDetected atomic.Bool
		ld.SetLeakCallback(func(leaked []*ObjectRef) {
			leakDetected.Store(true)
		})

		ld.Track("leak-obj", PoolSmall, 1000)

		ld.Start()
		defer ld.Stop()

		// Wait for leak detection
		time.Sleep(200 * time.Millisecond)

		if !leakDetected.Load() {
			t.Error("leak was not detected")
		}
	})
}

func TestLeakDetector_GetStats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLeakDetector_GetStats", nil, func(t *testing.T, tx *gorm.DB) {
		ld := NewLeakDetectorWithDefaults()

		// Track some objects
		ld.Track("obj-1", PoolSmall, 1000)
		ld.Track("obj-2", PoolMedium, 5000)
		ld.Track("obj-3", PoolLarge, 15000)

		// Release one
		ld.Release("obj-1")

		stats := ld.GetStats()

		if stats["active_count"].(int) != 2 {
			t.Errorf("expected 2 active, got %d", stats["active_count"])
		}
		if stats["tracked_total"].(int) != 3 {
			t.Errorf("expected 3 tracked, got %d", stats["tracked_total"])
		}
	})
}

func TestLeakDetector_StackTrace(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLeakDetector_StackTrace", nil, func(t *testing.T, tx *gorm.DB) {
		ld := NewLeakDetectorWithDefaults()

		ref := ld.Track("stack-obj", PoolTiny, 100)

		stack := ld.GetStackTrace(ref)
		if stack == "" {
			t.Error("expected non-empty stack trace")
		}
	})
}

func TestLeakDetector_Disabled(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLeakDetector_Disabled", nil, func(t *testing.T, tx *gorm.DB) {
		config := DefaultLeakDetectorConfig()
		config.Enabled = false

		ld := NewLeakDetector(config)

		ref := ld.Track("obj-1", PoolSmall, 1000)
		if ref != nil {
			t.Error("expected nil reference when disabled")
		}

		stats := ld.GetStats()
		if stats["active_count"].(int) != 0 {
			t.Errorf("expected 0 active when disabled, got %d", stats["active_count"])
		}
	})
}

func TestLeakDetector_ConcurrentTracking(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLeakDetector_ConcurrentTracking", nil, func(t *testing.T, tx *gorm.DB) {
		ld := NewLeakDetectorWithDefaults()

		// Concurrent track and release
		for i := 0; i < 100; i++ {
			id := "obj-" + string(rune(i))
			ld.Track(id, PoolSmall, 1000)
			ld.Release(id)
		}

		stats := ld.GetStats()
		if stats["active_count"].(int) != 0 {
			t.Errorf("expected 0 active after release, got %d", stats["active_count"])
		}
	})
}

func TestLeakDetector_StartStop(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLeakDetector_StartStop", nil, func(t *testing.T, tx *gorm.DB) {
		ld := NewLeakDetectorWithDefaults()

		ld.Start()
		if !ld.started {
			t.Error("expected started to be true")
		}

		ld.Stop()
		if ld.started {
			t.Error("expected started to be false after stop")
		}
	})
}

func TestLeakDetector_CustomConfig(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLeakDetector_CustomConfig", nil, func(t *testing.T, tx *gorm.DB) {
		config := LeakDetectorConfig{
			Enabled:       true,
			CheckInterval: 10 * time.Second,
			MaxLifetime:   1 * time.Minute,
			Threshold:     5,
		}

		ld := NewLeakDetector(config)

		if ld.config.CheckInterval != 10*time.Second {
			t.Errorf("expected check interval 10s, got %v", ld.config.CheckInterval)
		}
		if ld.config.Threshold != 5 {
			t.Errorf("expected threshold 5, got %d", ld.config.Threshold)
		}
	})
}

func TestLeakDetector_NilRef(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLeakDetector_NilRef", nil, func(t *testing.T, tx *gorm.DB) {
		ld := NewLeakDetectorWithDefaults()

		stack := ld.GetStackTrace(nil)
		if stack != "no stack trace available" {
			t.Errorf("expected default message, got %s", stack)
		}
	})
}

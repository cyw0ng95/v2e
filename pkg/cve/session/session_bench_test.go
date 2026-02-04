package session

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"os"
	"path/filepath"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
)

func TestMain(m *testing.M) {
	// Set up any necessary global state or resources here.
	// For example, you might initialize a logger or a database connection.

	// Run the tests
	code := m.Run()

	// Clean up any global state or resources here.

	// Exit with the code from running the tests
	os.Exit(code)
}

// BenchmarkCreateSession benchmarks session creation
func BenchmarkCreateSession(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "bench_session.db")

	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	manager, err := NewManager(dbPath, logger)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Delete previous session
		manager.DeleteSession()

		// Create new session
		_, err := manager.CreateSession("bench-session", 0, 100)
		if err != nil {
			b.Fatalf("Failed to create session: %v", err)
		}
	}
}

// BenchmarkGetSession benchmarks session retrieval
func BenchmarkGetSession(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "bench_get_session.db")

	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	manager, err := NewManager(dbPath, logger)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create a session
	_, err = manager.CreateSession("bench-session", 0, 100)
	if err != nil {
		b.Fatalf("Failed to create session: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := manager.GetSession()
		if err != nil {
			b.Fatalf("Failed to get session: %v", err)
		}
	}
}

// BenchmarkUpdateState benchmarks state updates
func BenchmarkUpdateState(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "bench_update_state.db")

	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	manager, err := NewManager(dbPath, logger)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	_, err = manager.CreateSession("bench-session", 0, 100)
	if err != nil {
		b.Fatalf("Failed to create session: %v", err)
	}

	states := []SessionState{StateRunning, StatePaused, StateRunning, StateStopped}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state := states[i%len(states)]
		err := manager.UpdateState(state)
		if err != nil {
			b.Fatalf("Failed to update state: %v", err)
		}
	}
}

// BenchmarkUpdateProgress benchmarks progress updates
func BenchmarkUpdateProgress(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "bench_update_progress.db")

	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	manager, err := NewManager(dbPath, logger)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	_, err = manager.CreateSession("bench-session", 0, 100)
	if err != nil {
		b.Fatalf("Failed to create session: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := manager.UpdateProgress(1, 1, 0)
		if err != nil {
			b.Fatalf("Failed to update progress: %v", err)
		}
	}
}

// BenchmarkDeleteSession benchmarks session deletion
func BenchmarkDeleteSession(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "bench_delete_session.db")

	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	manager, err := NewManager(dbPath, logger)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create a session before each delete
		_, err := manager.CreateSession("bench-session", 0, 100)
		if err != nil {
			b.Fatalf("Failed to create session: %v", err)
		}
		b.StartTimer()

		err = manager.DeleteSession()
		if err != nil {
			b.Fatalf("Failed to delete session: %v", err)
		}
	}
}

// BenchmarkConcurrentGetSession benchmarks concurrent session retrieval
func BenchmarkConcurrentGetSession(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "bench_concurrent_get.db")

	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	manager, err := NewManager(dbPath, logger)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	_, err = manager.CreateSession("bench-session", 0, 100)
	if err != nil {
		b.Fatalf("Failed to create session: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := manager.GetSession()
			if err != nil {
				b.Fatalf("Failed to get session: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentUpdateProgress benchmarks concurrent progress updates
func BenchmarkConcurrentUpdateProgress(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "bench_concurrent_progress.db")

	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	manager, err := NewManager(dbPath, logger)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	_, err = manager.CreateSession("bench-session", 0, 100)
	if err != nil {
		b.Fatalf("Failed to create session: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := manager.UpdateProgress(1, 1, 0)
			if err != nil {
				b.Fatalf("Failed to update progress: %v", err)
			}
		}
	})
}

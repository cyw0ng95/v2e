package taskflow

import (
	"path/filepath"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// NewTestLogger returns a logger suitable for tests
func NewTestLogger(t testing.TB) *common.Logger {
	return common.NewLogger(testingWriter{t}, "[test]", common.DebugLevel)
}

type testingWriter struct {
	t testing.TB
}

func (w testingWriter) Write(p []byte) (int, error) {
	w.t.Log(string(p))
	return len(p), nil
}

// NewTempRunStore creates a RunStore backed by a BoltDB file in t.TempDir()
func NewTempRunStore(t testing.TB) *RunStore {
	dbPath := filepath.Join(t.TempDir(), "runs.db")
	logger := NewTestLogger(t)
	rs, err := NewRunStore(dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create RunStore: %v", err)
	}
	t.Cleanup(func() {
		rs.Close()
	})
	return rs
}

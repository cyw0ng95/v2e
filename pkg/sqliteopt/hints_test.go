package sqliteopt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestApplyFileHints(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.db")
	
	// Create the file
	file, err := os.Create(testFile)
	require.NoError(t, err)
	file.WriteString("test data")
	file.Close()
	
	// Apply hints
	err = ApplyFileHints(testFile)
	assert.NoError(t, err)
}

func TestApplyFileHintsNonExistent(t *testing.T) {
	err := ApplyFileHints("/nonexistent/file.db")
	assert.Error(t, err)
}

func TestApplyRandomAccessHints(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.db")
	
	file, err := os.Create(testFile)
	require.NoError(t, err)
	file.WriteString("test data")
	file.Close()
	
	err = ApplyRandomAccessHints(testFile)
	assert.NoError(t, err)
}

func TestDropCachesForFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.db")
	
	file, err := os.Create(testFile)
	require.NoError(t, err)
	file.WriteString("test data")
	file.Close()
	
	err = DropCachesForFile(testFile)
	assert.NoError(t, err)
}

func TestConfigureOptimalSQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	// Create a GORM database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)
	
	// Configure with optimal settings
	err = ConfigureOptimalSQLite(db, dbPath)
	assert.NoError(t, err)
	
	// Verify pragmas were set
	var result string
	db.Raw("PRAGMA journal_mode").Scan(&result)
	assert.Equal(t, "wal", result)
	
	db.Raw("PRAGMA synchronous").Scan(&result)
	assert.Contains(t, []string{"1", "NORMAL"}, result)
}

func TestConfigureOptimalSQLiteNilDB(t *testing.T) {
	err := ConfigureOptimalSQLite(nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestApplyKernelHintsNilDB(t *testing.T) {
	err := ApplyKernelHints(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestApplyPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.db")
	
	file, err := os.Create(testFile)
	require.NoError(t, err)
	file.WriteString("test data")
	file.Close()
	
	tests := []struct {
		name    string
		pattern FileAccessPattern
		wantErr bool
	}{
		{"Sequential", PatternSequential, false},
		{"Random", PatternRandom, false},
		{"WillNeed", PatternWillNeed, false},
		{"DontNeed", PatternDontNeed, false},
		{"Unknown", FileAccessPattern(999), true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyPattern(testFile, tt.pattern)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplyPatternNonExistent(t *testing.T) {
	err := ApplyPattern("/nonexistent/file.db", PatternSequential)
	assert.Error(t, err)
}

func BenchmarkApplyFileHints(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench.db")
	
	file, _ := os.Create(testFile)
	file.WriteString("benchmark data")
	file.Close()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = ApplyFileHints(testFile)
	}
}

func BenchmarkConfigureOptimalSQLite(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		db, _ := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		_ = ConfigureOptimalSQLite(db, dbPath)
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}
}

package sqliteopt

import (
	"database/sql"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
	"gorm.io/gorm"
)

// ApplyKernelHints applies POSIX file access hints to improve SQLite performance
func ApplyKernelHints(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB: %w", err)
	}
	
	return ApplyKernelHintsToSQLDB(sqlDB)
}

// ApplyKernelHintsToSQLDB applies kernel hints to a sql.DB instance
func ApplyKernelHintsToSQLDB(sqlDB *sql.DB) error {
	if sqlDB == nil {
		return fmt.Errorf("SQL DB is nil")
	}
	
	// Get database file path from connection
	// Note: This is a simplified implementation. In practice, you'd need to
	// extract the file path from the connection string or use a different method
	
	// For now, return nil since we can't easily get the file descriptor
	// from sql.DB without deeper integration
	return nil
}

// ApplyFileHints applies POSIX_FADV_SEQUENTIAL and other hints to a file
func ApplyFileHints(filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()
	
	// Get file descriptor
	fd := int(file.Fd())
	
	// Apply POSIX_FADV_SEQUENTIAL hint
	// This tells the kernel that we'll be reading the file sequentially
	err = unix.Fadvise(fd, 0, 0, unix.FADV_SEQUENTIAL)
	if err != nil {
		return fmt.Errorf("failed to apply FADV_SEQUENTIAL: %w", err)
	}
	
	// Apply POSIX_FADV_WILLNEED hint
	// This tells the kernel we'll need this data soon, triggering readahead
	err = unix.Fadvise(fd, 0, 0, unix.FADV_WILLNEED)
	if err != nil {
		return fmt.Errorf("failed to apply FADV_WILLNEED: %w", err)
	}
	
	return nil
}

// ApplyRandomAccessHints applies hints for random access patterns
func ApplyRandomAccessHints(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()
	
	fd := int(file.Fd())
	
	// Apply POSIX_FADV_RANDOM hint
	// This tells the kernel that we'll be accessing the file randomly
	err = unix.Fadvise(fd, 0, 0, unix.FADV_RANDOM)
	if err != nil {
		return fmt.Errorf("failed to apply FADV_RANDOM: %w", err)
	}
	
	return nil
}

// DropCachesForFile drops page cache for a specific file
func DropCachesForFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()
	
	fd := int(file.Fd())
	
	// Apply POSIX_FADV_DONTNEED hint
	// This tells the kernel we're done with this data and it can be evicted
	err = unix.Fadvise(fd, 0, 0, unix.FADV_DONTNEED)
	if err != nil {
		return fmt.Errorf("failed to apply FADV_DONTNEED: %w", err)
	}
	
	return nil
}

// ConfigureOptimalSQLite configures SQLite with optimal settings
func ConfigureOptimalSQLite(db *gorm.DB, dbPath string) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// Apply kernel hints to the database file
	if dbPath != "" {
		// Apply sequential access hints (good for table scans)
		if err := ApplyFileHints(dbPath); err != nil {
			// Log warning but don't fail - hints are optional optimizations
			// In production, you'd use proper logging here
			fmt.Printf("Warning: failed to apply file hints: %v\n", err)
		}
	}
	
	// Configure SQLite pragmas for optimal performance
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB: %w", err)
	}
	
	// Enable WAL mode for better concurrent access
	if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}
	
	// Set synchronous mode to NORMAL (faster than FULL, still safe with WAL)
	if err := db.Exec("PRAGMA synchronous=NORMAL").Error; err != nil {
		return fmt.Errorf("failed to set synchronous mode: %w", err)
	}
	
	// Increase cache size (40MB)
	if err := db.Exec("PRAGMA cache_size=-40000").Error; err != nil {
		return fmt.Errorf("failed to set cache size: %w", err)
	}
	
	// Set temp store to memory
	if err := db.Exec("PRAGMA temp_store=MEMORY").Error; err != nil {
		return fmt.Errorf("failed to set temp store: %w", err)
	}
	
	// Set mmap size (256MB)
	if err := db.Exec("PRAGMA mmap_size=268435456").Error; err != nil {
		return fmt.Errorf("failed to set mmap size: %w", err)
	}
	
	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	
	return nil
}

// FileAccessPattern represents different access patterns for files
type FileAccessPattern int

const (
	// PatternSequential indicates sequential file access
	PatternSequential FileAccessPattern = iota
	// PatternRandom indicates random file access
	PatternRandom
	// PatternWillNeed indicates data will be needed soon
	PatternWillNeed
	// PatternDontNeed indicates data won't be needed anymore
	PatternDontNeed
)

// ApplyPattern applies the specified access pattern hint to a file
func ApplyPattern(filePath string, pattern FileAccessPattern) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()
	
	fd := int(file.Fd())
	
	var advice int
	switch pattern {
	case PatternSequential:
		advice = unix.FADV_SEQUENTIAL
	case PatternRandom:
		advice = unix.FADV_RANDOM
	case PatternWillNeed:
		advice = unix.FADV_WILLNEED
	case PatternDontNeed:
		advice = unix.FADV_DONTNEED
	default:
		return fmt.Errorf("unknown access pattern: %d", pattern)
	}
	
	err = unix.Fadvise(fd, 0, 0, advice)
	if err != nil {
		return fmt.Errorf("failed to apply pattern %d: %w", pattern, err)
	}
	
	return nil
}

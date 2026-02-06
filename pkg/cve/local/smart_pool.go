package local

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// SmartConnectionPool provides intelligent connection pooling with query pattern awareness
type SmartConnectionPool struct {
	readPool  *ConnectionPool
	writePool *ConnectionPool
	analyzer  *QueryPatternAnalyzer
	cache     *PreparedStatementCache
	metrics   *PoolMetrics
}

// ConnectionPool represents a basic connection pool
type ConnectionPool struct {
	db          *gorm.DB
	maxConns    int
	activeConns int32
	mu          sync.Mutex
	healthCheck time.Duration
}

// PoolMetrics tracks connection pool efficiency
type PoolMetrics struct {
	TotalQueries       int64
	CacheHits          int64
	CacheMisses        int64
	AvgQueryTime       int64
	ActiveConnections  int64
	HealthChecks       int64
	FailedHealthChecks int64
}

// PreparedStatementCache caches prepared statements
type PreparedStatementCache struct {
	cache   map[string]*gorm.Stmt
	maxSize int
	mu      sync.RWMutex
	metrics *PoolMetrics
}

// NewSmartConnectionPool creates a new smart connection pool
func NewSmartConnectionPool(db *gorm.DB, maxReadConns, maxWriteConns int) (*SmartConnectionPool, error) {
	readPool := &ConnectionPool{
		db:          db,
		maxConns:    maxReadConns,
		healthCheck: 30 * time.Second,
	}

	writePool := &ConnectionPool{
		db:          db,
		maxConns:    maxWriteConns,
		healthCheck: 30 * time.Second,
	}

	analyzer := NewQueryPatternAnalyzer(100)
	cache := NewPreparedStatementCache(500)
	metrics := &PoolMetrics{}

	scp := &SmartConnectionPool{
		readPool:  readPool,
		writePool: writePool,
		analyzer:  analyzer,
		cache:     cache,
		metrics:   metrics,
	}

	return scp, nil
}

// Query executes a read query using the read pool
func (scp *SmartConnectionPool) Query(ctx context.Context, sql string, args ...interface{}) *gorm.DB {
	atomic.AddInt64(&scp.metrics.TotalQueries, 1)
	startTime := time.Now()

	scp.analyzer.RecordQuery(sql, time.Since(startTime))

	return scp.readPool.db.WithContext(ctx).Raw(sql, args...)
}

// Exec executes a write query using the write pool
func (scp *SmartConnectionPool) Exec(ctx context.Context, sql string, args ...interface{}) *gorm.DB {
	atomic.AddInt64(&scp.metrics.TotalQueries, 1)
	return scp.writePool.db.WithContext(ctx).Exec(sql, args...)
}

// GetReadDB returns the read database connection
func (scp *SmartConnectionPool) GetReadDB() *gorm.DB {
	return scp.readPool.db
}

// GetWriteDB returns the write database connection
func (scp *SmartConnectionPool) GetWriteDB() *gorm.DB {
	return scp.writePool.db
}

// PerformHealthCheck performs health checks on both pools
func (scp *SmartConnectionPool) PerformHealthCheck() error {
	atomic.AddInt64(&scp.metrics.HealthChecks, 1)

	sqlDB, err := scp.readPool.db.DB()
	if err != nil {
		atomic.AddInt64(&scp.metrics.FailedHealthChecks, 1)
		return fmt.Errorf("read pool DB error: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		atomic.AddInt64(&scp.metrics.FailedHealthChecks, 1)
		return fmt.Errorf("read pool ping failed: %w", err)
	}

	sqlDB, err = scp.writePool.db.DB()
	if err != nil {
		atomic.AddInt64(&scp.metrics.FailedHealthChecks, 1)
		return fmt.Errorf("write pool DB error: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		atomic.AddInt64(&scp.metrics.FailedHealthChecks, 1)
		return fmt.Errorf("write pool ping failed: %w", err)
	}

	return nil
}

// GetMetrics returns current pool metrics
func (scp *SmartConnectionPool) GetMetrics() PoolMetrics {
	scp.metrics.ActiveConnections = int64(atomic.LoadInt32(&scp.readPool.activeConns) + atomic.LoadInt32(&scp.writePool.activeConns))
	return *scp.metrics
}

// Close closes both connection pools
func (scp *SmartConnectionPool) Close() error {
	sqlDB, err := scp.readPool.db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Close(); err != nil {
		return err
	}

	sqlDB, err = scp.writePool.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AcquireConnection acquires a connection from the appropriate pool
func (cp *ConnectionPool) AcquireConnection() (*gorm.DB, error) {
	atomic.AddInt32(&cp.activeConns, 1)
	return cp.db, nil
}

// ReleaseConnection releases a connection back to the pool
func (cp *ConnectionPool) ReleaseConnection() {
	atomic.AddInt32(&cp.activeConns, -1)
}

// NewPreparedStatementCache creates a new prepared statement cache
func NewPreparedStatementCache(maxSize int) *PreparedStatementCache {
	return &PreparedStatementCache{
		cache:   make(map[string]interface{}),
		maxSize: maxSize,
		metrics: &PoolMetrics{},
	}
}

// Get retrieves a cached statement
func (psc *PreparedStatementCache) Get(key string) interface{} {
	psc.mu.RLock()
	defer psc.mu.RUnlock()

	stmt, exists := psc.cache[key]
	if exists {
		atomic.AddInt64(&psc.metrics.CacheHits, 1)
		return stmt
	}
	atomic.AddInt64(&psc.metrics.CacheMisses, 1)
	return nil
}

// Set stores a statement in the cache
func (psc *PreparedStatementCache) Set(key string, stmt interface{}) {
	psc.mu.Lock()
	defer psc.mu.Unlock()

	if len(psc.cache) >= psc.maxSize {
		psc.evictLRU()
	}

	psc.cache[key] = stmt
}

// evictLRU evicts the least recently used statement
func (psc *PreparedStatementCache) evictLRU() {
	// Simple LRU: remove first entry
	for key := range psc.cache {
		delete(psc.cache, key)
		return
	}
}

// Clear clears the cache
func (psc *PreparedStatementCache) Clear() {
	psc.mu.Lock()
	defer psc.mu.Unlock()

	psc.cache = make(map[string]*gorm.Stmt)
}

package local

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
	"time"
)

// QueryPatternAnalyzer analyzes query execution patterns
type QueryPatternAnalyzer struct {
	patterns    map[string]*QueryPattern
	maxPatterns int
	mu          sync.RWMutex
}

// QueryPattern represents a query execution pattern
type QueryPattern struct {
	SQL          string
	ExecuteCount int64
	AvgDuration  int64
	LastExecuted time.Time
	ParamsHash   string
}

// NewQueryPatternAnalyzer creates a new query pattern analyzer
func NewQueryPatternAnalyzer(maxPatterns int) *QueryPatternAnalyzer {
	return &QueryPatternAnalyzer{
		patterns:    make(map[string]*QueryPattern),
		maxPatterns: maxPatterns,
	}
}

// RecordQuery records a query execution for pattern analysis
func (qpa *QueryPatternAnalyzer) RecordQuery(sql string, duration time.Duration) {
	qpa.mu.Lock()
	defer qpa.mu.Unlock()

	key := qpa.normalizeSQL(sql)
	pattern, exists := qpa.patterns[key]

	if !exists {
		pattern = &QueryPattern{
			SQL:          sql,
			ExecuteCount: 0,
			AvgDuration:  0,
			LastExecuted: time.Now(),
			ParamsHash:   "",
		}

		qpa.patterns[key] = pattern

		if len(qpa.patterns) > qpa.maxPatterns {
			qpa.evictLeastUsed()
		}
	}

	pattern.ExecuteCount++
	pattern.LastExecuted = time.Now()

	durationNanos := int64(duration)
	if pattern.ExecuteCount == 1 {
		pattern.AvgDuration = durationNanos
	} else {
		pattern.AvgDuration = (pattern.AvgDuration*(pattern.ExecuteCount-1) + durationNanos) / pattern.ExecuteCount
	}
}

// normalizeSQL normalizes SQL for pattern matching
func (qpa *QueryPatternAnalyzer) normalizeSQL(sql string) string {
	hash := md5.Sum([]byte(sql))
	return hex.EncodeToString(hash[:])
}

// GetFrequentQueries returns frequently executed queries
func (qpa *QueryPatternAnalyzer) GetFrequentQueries(limit int) []*QueryPattern {
	qpa.mu.RLock()
	defer qpa.mu.RUnlock()

	patterns := make([]*QueryPattern, 0, len(qpa.patterns))
	for _, pattern := range qpa.patterns {
		patterns = append(patterns, pattern)
	}

	for i := 0; i < len(patterns)-1; i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[i].ExecuteCount < patterns[j].ExecuteCount {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	if len(patterns) > limit {
		patterns = patterns[:limit]
	}

	return patterns
}

// GetPattern returns pattern for a specific SQL
func (qpa *QueryPatternAnalyzer) GetPattern(sql string) *QueryPattern {
	qpa.mu.RLock()
	defer qpa.mu.RUnlock()

	key := qpa.normalizeSQL(sql)
	return qpa.patterns[key]
}

// evictLeastUsed evicts the least frequently used pattern
func (qpa *QueryPatternAnalyzer) evictLeastUsed() {
	var leastUsedKey string
	var leastUsedCount int64 = -1

	for key, pattern := range qpa.patterns {
		if leastUsedCount == -1 || pattern.ExecuteCount < leastUsedCount {
			leastUsedKey = key
			leastUsedCount = pattern.ExecuteCount
		}
	}

	if leastUsedKey != "" {
		delete(qpa.patterns, leastUsedKey)
	}
}

// Clear clears all patterns
func (qpa *QueryPatternAnalyzer) Clear() {
	qpa.mu.Lock()
	defer qpa.mu.Unlock()

	qpa.patterns = make(map[string]*QueryPattern)
}

// GetStats returns pattern statistics
func (qpa *QueryPatternAnalyzer) GetStats() map[string]interface{} {
	qpa.mu.RLock()
	defer qpa.mu.RUnlock()

	totalExecutions := int64(0)
	for _, pattern := range qpa.patterns {
		totalExecutions += pattern.ExecuteCount
	}

	return map[string]interface{}{
		"total_patterns":   len(qpa.patterns),
		"total_executions": totalExecutions,
	}
}

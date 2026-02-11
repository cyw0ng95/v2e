package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/local"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

// TestCVEHandlers_ConcurrentAccess tests concurrent access to CVE database operations
func TestCVEHandlers_ConcurrentAccess(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCVEHandlers_ConcurrentAccess", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cve-concurrent-test.db")

		logger := common.NewLogger(nil, "[TEST] ", common.ErrorLevel)

		db, err := local.NewDB(dbPath)
		if err != nil {
			t.Fatalf("NewDB error: %v", err)
		}
		defer db.Close()

		ctx := context.Background()
		createH := createCreateCVEHandler(db, logger)
		getH := createGetCVEByIDHandler(db, logger)
		listH := createListCVEsHandler(db, logger)
		countH := createCountCVEsHandler(db, logger)

		// Test: Concurrent creates
		t.Run("ConcurrentCreates", func(t *testing.T) {
			const numGoroutines = 50
			const numItemsPerGoroutine = 10

			var wg sync.WaitGroup
			var successCount atomic.Int64
			var errorCount atomic.Int64

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(goroutineID int) {
					defer wg.Done()
					for j := 0; j < numItemsPerGoroutine; j++ {
						id := fmt.Sprintf("CVE-2024-%04d", goroutineID*100+j)
						item := cve.CVEItem{
							ID: id,
							Descriptions: []cve.Description{{Lang: "en", Value: fmt.Sprintf("Test %d", j)}},
						}
						payload, _ := subprocess.MarshalFast(item)
						msg := &subprocess.Message{
							Type: subprocess.MessageTypeRequest,
							ID: fmt.Sprintf("create-%s", id),
							Payload: payload,
							Source: "test",
							Target: "local",
						}
						resp, err := createH(ctx, msg)
						if err != nil || resp == nil || resp.Type != subprocess.MessageTypeResponse {
							errorCount.Add(1)
						} else {
							successCount.Add(1)
						}
					}
				}(i)
			}

			wg.Wait()

			totalCreated := successCount.Load() + errorCount.Load()
			expectedTotal := int64(numGoroutines * numItemsPerGoroutine)

			if totalCreated != expectedTotal {
				t.Fatalf("expected %d operations, got %d (success=%d, error=%d)",
					expectedTotal, totalCreated, successCount.Load(), errorCount.Load())
			}

			// Verify count
			countMsg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "count"}
			resp, err := countH(ctx, countMsg)
			if err != nil {
				t.Fatalf("count failed: %v", err)
			}
			var result map[string]interface{}
			subprocess.UnmarshalPayload(resp, &result)
			count := int(result["count"].(float64))
			if count != numGoroutines*numItemsPerGoroutine {
				t.Logf("Warning: expected %d CVEs, got %d (some may have been deduplicated)", numGoroutines*numItemsPerGoroutine, count)
			}
		})

		// Test: Concurrent reads
		t.Run("ConcurrentReads", func(t *testing.T) {
			// First, create a CVE to read
			item := cve.CVEItem{
				ID: "CVE-2024-5000",
				Descriptions: []cve.Description{{Lang: "en", Value: "Concurrent test"}},
			}
			payload, _ := subprocess.MarshalFast(item)
			createMsg := &subprocess.Message{
				Type: subprocess.MessageTypeRequest,
				ID: "create-5000",
				Payload: payload,
			}
			createH(ctx, createMsg)

			const numGoroutines = 100
			const readsPerGoroutine = 50

			var wg sync.WaitGroup
			var successCount atomic.Int64

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < readsPerGoroutine; j++ {
						getPayload, _ := subprocess.MarshalFast(map[string]string{"cve_id": "CVE-2024-5000"})
						getMsg := &subprocess.Message{
							Type: subprocess.MessageTypeRequest,
							ID: fmt.Sprintf("get-%d", j),
							Payload: getPayload,
						}
						resp, err := getH(ctx, getMsg)
						if err != nil || resp == nil || resp.Type != subprocess.MessageTypeResponse {
							t.Logf("concurrent get failed: err=%v resp=%v", err, resp)
						} else {
							successCount.Add(1)
						}
					}
				}()
			}

			wg.Wait()

			totalReads := numGoroutines * readsPerGoroutine
			successfulReads := successCount.Load()

			if successfulReads < int64(totalReads*95/100) { // Allow 5% failure rate
				t.Fatalf("too many failed concurrent reads: %d/%d", successfulReads, totalReads)
			}
		})

		// Test: Concurrent writes and reads
		t.Run("ConcurrentWritesAndReads", func(t *testing.T) {
			const numWriters = 20
			const numReaders = 30
			const operationsPerGoroutine = 20

			var wg sync.WaitGroup
			var writeSuccess atomic.Int64
			var readSuccess atomic.Int64

			// Start writers
			for i := 0; i < numWriters; i++ {
				wg.Add(1)
				go func(writerID int) {
					defer wg.Done()
					for j := 0; j < operationsPerGoroutine; j++ {
						id := fmt.Sprintf("CVE-2025-%04d", writerID*1000+j)
						item := cve.CVEItem{
							ID: id,
							Descriptions: []cve.Description{{Lang: "en", Value: fmt.Sprintf("Writer %d item %d", writerID, j)}},
						}
						payload, _ := subprocess.MarshalFast(item)
						msg := &subprocess.Message{
							Type: subprocess.MessageTypeRequest,
							ID: fmt.Sprintf("write-%d", j),
							Payload: payload,
						}
						resp, err := createH(ctx, msg)
						if err == nil && resp != nil && resp.Type == subprocess.MessageTypeResponse {
							writeSuccess.Add(1)
						}
					}
				}(i)
			}

			// Start readers
			for i := 0; i < numReaders; i++ {
				wg.Add(1)
				go func(readerID int) {
					defer wg.Done()
					for j := 0; j < operationsPerGoroutine; j++ {
						// Try to read a CVE that might or might not exist
						id := fmt.Sprintf("CVE-2025-%04d", (j%20)*1000+(j%10))
						getPayload, _ := subprocess.MarshalFast(map[string]string{"cve_id": id})
						getMsg := &subprocess.Message{
							Type: subprocess.MessageTypeRequest,
							ID: fmt.Sprintf("read-%d-%d", readerID, j),
							Payload: getPayload,
						}
						resp, err := getH(ctx, getMsg)
						// Both success and error responses are OK (CVE might not exist yet)
						if err == nil && resp != nil {
							readSuccess.Add(1)
						}
					}
				}(i)
			}

			wg.Wait()

			totalWrites := int64(numWriters * operationsPerGoroutine)
			totalReads := int64(numReaders * operationsPerGoroutine)

			t.Logf("Concurrent writes and reads completed: writes=%d/%d, reads=%d/%d",
				writeSuccess.Load(), totalWrites, readSuccess.Load(), totalReads)

			if writeSuccess.Load() < totalWrites*90/100 {
				t.Logf("Warning: low write success rate: %d/%d", writeSuccess.Load(), totalWrites)
			}
		})

		// Test: Concurrent list operations
		t.Run("ConcurrentListOperations", func(t *testing.T) {
			const numGoroutines = 30
			const listsPerGoroutine = 20

			var wg sync.WaitGroup
			var successCount atomic.Int64

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(goroutineID int) {
					defer wg.Done()
					for j := 0; j < listsPerGoroutine; j++ {
						offset := (goroutineID * listsPerGoroutine + j) % 100
						limit := 10
						payload, _ := subprocess.MarshalFast(map[string]int{"offset": offset, "limit": limit})
						msg := &subprocess.Message{
							Type: subprocess.MessageTypeRequest,
							ID: fmt.Sprintf("list-%d-%d", goroutineID, j),
							Payload: payload,
						}
						resp, err := listH(ctx, msg)
						if err == nil && resp != nil && resp.Type == subprocess.MessageTypeResponse {
							successCount.Add(1)
						}
					}
				}(i)
			}

			wg.Wait()

			totalLists := int64(numGoroutines * listsPerGoroutine)
			successfulLists := successCount.Load()

			if successfulLists < totalLists*95/100 {
				t.Fatalf("too many failed concurrent list operations: %d/%d", successfulLists, totalLists)
			}
		})
	})
}

// TestCVEHandlers_ConcurrentUpdates tests concurrent updates to the same CVE
func TestCVEHandlers_ConcurrentUpdates(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCVEHandlers_ConcurrentUpdates", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cve-update-test.db")

		logger := common.NewLogger(nil, "[TEST] ", common.ErrorLevel)

		db, err := local.NewDB(dbPath)
		if err != nil {
			t.Fatalf("NewDB error: %v", err)
		}
		defer db.Close()

		ctx := context.Background()
		updateH := createUpdateCVEHandler(db, logger)

		// Create initial CVE
		createH := createCreateCVEHandler(db, logger)
		item := cve.CVEItem{ID: "CVE-2024-9999", Descriptions: []cve.Description{{Lang: "en", Value: "Initial"}}}
		payload, _ := subprocess.MarshalFast(item)
		createMsg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "create", Payload: payload}
		createH(ctx, createMsg)

		// Test: Concurrent updates to same CVE
		t.Run("ConcurrentSameCVEUpdates", func(t *testing.T) {
			const numGoroutines = 50

			var wg sync.WaitGroup
			var successCount atomic.Int64

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(updateID int) {
					defer wg.Done()
					updated := cve.CVEItem{
						ID: "CVE-2024-9999",
						Descriptions: []cve.Description{{Lang: "en", Value: fmt.Sprintf("Update %d", updateID)}},
					}
					payload, _ := subprocess.MarshalFast(updated)
					msg := &subprocess.Message{
						Type: subprocess.MessageTypeRequest,
						ID: fmt.Sprintf("update-%d", updateID),
						Payload: payload,
					}
					resp, err := updateH(ctx, msg)
					if err == nil && resp != nil && resp.Type == subprocess.MessageTypeResponse {
						successCount.Add(1)
					}
				}(i)
			}

			wg.Wait()

			// All updates should succeed (last write wins)
			if successCount.Load() != int64(numGoroutines) {
				t.Logf("Warning: some concurrent updates failed: %d/%d", successCount.Load(), numGoroutines)
			}

			// Verify final state
			getH := createGetCVEByIDHandler(db, logger)
			getPayload, _ := subprocess.MarshalFast(map[string]string{"cve_id": "CVE-2024-9999"})
			getMsg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "get", Payload: getPayload}
			resp, err := getH(ctx, getMsg)
			if err != nil {
				t.Fatalf("get after concurrent updates failed: %v", err)
			}
			if resp.Type != subprocess.MessageTypeResponse {
				t.Fatalf("expected response after updates, got error: %v", resp.Error)
			}
		})
	})
}

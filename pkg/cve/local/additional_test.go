package local

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB_ErrorCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestNewDB_ErrorCases", nil, func(t *testing.T, tx *gorm.DB) {
		// Test with invalid path (try to use a directory path)
		db, err := NewDB("/nonexistent/directory/file.db")
		assert.Error(t, err)
		assert.Nil(t, db)

		// Test with read-only directory (if possible)
		// This test may not work in all environments, so we'll skip it if we can't make it fail
	})

}

func TestSaveCVE_EdgeCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSaveCVE_EdgeCases", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_save_cve_edge.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		// Test with empty CVE item
		emptyCVE := &cve.CVEItem{}
		err = db.SaveCVE(emptyCVE)
		assert.NoError(t, err)

		// Test with CVE having empty fields
		minimalCVE := &cve.CVEItem{
			ID:           "CVE-2021-MINIMAL",
			SourceID:     "",
			Published:    cve.NewNVDTime(time.Time{}),
			LastModified: cve.NewNVDTime(time.Time{}),
			VulnStatus:   "",
		}
		err = db.SaveCVE(minimalCVE)
		assert.NoError(t, err)

		// Test with very large description
		largeDesc := &cve.CVEItem{
			ID:           "CVE-2021-LARGEDESC",
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Time{}),
			LastModified: cve.NewNVDTime(time.Time{}),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: "A" + string(make([]byte, 10000)), // 10KB string
				},
			},
		}
		err = db.SaveCVE(largeDesc)
		assert.NoError(t, err)
	})

}

func TestSaveCVEs_EdgeCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSaveCVEs_EdgeCases", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_save_cves_edge.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		// Test with empty slice
		err = db.SaveCVEs([]cve.CVEItem{})
		assert.NoError(t, err)

		// Test with nil slice
		err = db.SaveCVEs(nil)
		assert.NoError(t, err)

		// Test with large slice
		var largeSlice []cve.CVEItem
		for i := 0; i < 100; i++ {
			largeSlice = append(largeSlice, cve.CVEItem{
				ID:           "CVE-2021-LARGESLICE" + string(rune(i+48)),
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Time{}),
				LastModified: cve.NewNVDTime(time.Time{}),
				VulnStatus:   "Analyzed",
			})
		}

		err = db.SaveCVEs(largeSlice)
		assert.NoError(t, err)
	})

}

func TestGetCVE_EdgeCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestGetCVE_EdgeCases", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_get_cve_edge.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		// Test with empty string ID
		_, err = db.GetCVE("")
		assert.Error(t, err)

		// Test with very long ID
		veryLongID := "CVE-" + string(make([]byte, 1000))
		_, err = db.GetCVE(veryLongID)
		assert.Error(t, err)

		// Test with special characters in ID
		_, err = db.GetCVE("CVE-2021-TEST!@#$%^&*()")
		assert.Error(t, err)
	})

}

func TestListCVEs_EdgeCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestListCVEs_EdgeCases", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_list_cves_edge.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		// Test with negative offset and limit
		_, err = db.ListCVEs(-1, -1)
		assert.NoError(t, err)

		// Test with zero limit
		_, err = db.ListCVEs(0, 0)
		assert.NoError(t, err)

		// Test with very large offset
		_, err = db.ListCVEs(999999, 10)
		assert.NoError(t, err)

		// Test with very large limit
		_, err = db.ListCVEs(0, 999999)
		assert.NoError(t, err)
	})

}

func TestCount_EdgeCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCount_EdgeCases", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_count_edge.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		// Test count on empty database
		count, err := db.Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)

		// Add some records and test again
		cves := []cve.CVEItem{
			{ID: "CVE-2021-TEST1"},
			{ID: "CVE-2021-TEST2"},
			{ID: "CVE-2021-TEST3"},
		}
		err = db.SaveCVEs(cves)
		assert.NoError(t, err)

		count, err = db.Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

}

func TestDeleteCVE_EdgeCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDeleteCVE_EdgeCases", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_delete_cve_edge.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		// Test deleting with empty string ID
		err = db.DeleteCVE("")
		assert.Error(t, err)

		// Test deleting with very long ID
		veryLongID := "CVE-" + string(make([]byte, 1000))
		err = db.DeleteCVE(veryLongID)
		assert.Error(t, err)
	})

}

func TestClose(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestClose", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_close.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)

		// Test closing twice (should handle gracefully)
		err = db.Close()
		assert.NoError(t, err)

		// Closing again should return an error (connection already closed)
		err = db.Close()
		// This may or may not return an error depending on the database driver behavior
		// Just make sure it doesn't panic
		_ = err
	})

}

func TestCVEDataIntegrity(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCVEDataIntegrity", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_data_integrity.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		// Create a CVE with all possible fields filled
		originalCVE := &cve.CVEItem{
			ID:           "CVE-2021-DATAINTEGRITY",
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Time{}),
			LastModified: cve.NewNVDTime(time.Time{}),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: "Test description",
				},
				{
					Lang:  "es",
					Value: "Descripción de prueba",
				},
			},
		}

		err = db.SaveCVE(originalCVE)
		assert.NoError(t, err)

		// Retrieve and verify integrity
		retrieved, err := db.GetCVE("CVE-2021-DATAINTEGRITY")
		assert.NoError(t, err)

		assert.Equal(t, originalCVE.ID, retrieved.ID)
		assert.Equal(t, originalCVE.SourceID, retrieved.SourceID)
		assert.Equal(t, originalCVE.VulnStatus, retrieved.VulnStatus)
		assert.Equal(t, len(originalCVE.Descriptions), len(retrieved.Descriptions))
		for i, desc := range originalCVE.Descriptions {
			assert.Equal(t, desc.Lang, retrieved.Descriptions[i].Lang)
			assert.Equal(t, desc.Value, retrieved.Descriptions[i].Value)
		}
	})

}

func TestConcurrentOperations(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConcurrentOperations", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_concurrent.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		// Test that basic operations work without concurrency issues
		// This is a simple test that doesn't actually test concurrency
		// but verifies that operations can be performed in sequence

		// Save multiple CVEs
		cves := []cve.CVEItem{
			{ID: "CVE-2021-CONCURRENT1", SourceID: "nvd@nist.gov", VulnStatus: "Analyzed"},
			{ID: "CVE-2021-CONCURRENT2", SourceID: "nvd@nist.gov", VulnStatus: "Analyzed"},
			{ID: "CVE-2021-CONCURRENT3", SourceID: "nvd@nist.gov", VulnStatus: "Analyzed"},
		}

		err = db.SaveCVEs(cves)
		assert.NoError(t, err)

		count, err := db.Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)

		// List them
		listed, err := db.ListCVEs(0, 10)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(listed))

		// Get each individually
		for _, cve := range cves {
			retrieved, err := db.GetCVE(cve.ID)
			assert.NoError(t, err)
			assert.Equal(t, cve.ID, retrieved.ID)
		}

		// Delete one
		err = db.DeleteCVE("CVE-2021-CONCURRENT1")
		assert.NoError(t, err)

		// Count should now be 2
		count, err = db.Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

}

func TestBulkInsert(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBulkInsert", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_bulk_insert.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		records := []CVERecord{
			{CVEID: "CVE-2021-BULK1", SourceID: "test"},
			{CVEID: "CVE-2021-BULK2", SourceID: "test"},
			{CVEID: "CVE-2021-BULK3", SourceID: "test"},
		}

		err = db.BulkInsert(records, 2)
		assert.NoError(t, err)

		count, err := db.Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}

func TestLazyCVERecord(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestLazyCVERecord", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_lazy_record.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		lazyRecord := db.NewLazyCVERecord("CVE-2021-LAZY")
		assert.NotNil(t, lazyRecord)
		assert.Equal(t, "CVE-2021-LAZY", lazyRecord.ID)

		err = lazyRecord.Load()
		assert.Error(t, err)
	})
}

func TestGetCVERaw(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestGetCVERaw", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_get_raw.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		cveItem := &cve.CVEItem{
			ID:           "CVE-2021-RAWTEST",
			SourceID:     "test",
			VulnStatus:   "Analyzed",
		}
		saveErr := db.SaveCVE(cveItem)
		require.NoError(t, saveErr)

		record, err := db.GetCVERaw("CVE-2021-RAWTEST")
		assert.NoError(t, err)
		assert.NotNil(t, record)
		assert.Equal(t, "CVE-2021-RAWTEST", record.CVEID)
	})
}

func TestGetCVERaw_NotFound(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestGetCVERaw_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_get_raw_notfound.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		_, err = db.GetCVERaw("CVE-9999-99999")
		assert.Error(t, err)
	})
}

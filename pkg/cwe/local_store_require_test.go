package cwe

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/stretchr/testify/require"
)

func TestLocalCWEStore_InMemoryCRUDAndPagination(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestLocalCWEStore_InMemoryCRUDAndPagination", nil, func(t *testing.T, tx *gorm.DB) {
		store, err := NewLocalCWEStore(":memory:")
		require.NoError(t, err)

		ctx := context.Background()

		// Initially empty
		items, total, err := store.ListCWEsPaginated(ctx, 0, 10)
		require.NoError(t, err)
		require.Equal(t, int64(0), total)
		require.Len(t, items, 0)

		// Insert a CWEItemModel directly via GORM
		m := CWEItemModel{
			ID:          "CWE-TEST-1",
			Name:        "Test CWE",
			Description: "desc",
		}
		require.NoError(t, store.db.Create(&m).Error)

		// Insert a related weakness for completeness
		rw := RelatedWeaknessModel{CWEID: m.ID, Nature: "example", CweID: "CWE-2", ViewID: "v1", Ordinal: "1"}
		require.NoError(t, store.db.Create(&rw).Error)

		// GetByID
		item, err := store.GetByID(ctx, "CWE-TEST-1")
		require.NoError(t, err)
		require.Equal(t, "CWE-TEST-1", item.ID)
		require.Equal(t, "Test CWE", item.Name)
		require.Len(t, item.RelatedWeaknesses, 1)

		// List paginated
		items, total, err = store.ListCWEsPaginated(ctx, 0, 10)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, items, 1)

		// Offset beyond range returns empty
		items, total, err = store.ListCWEsPaginated(ctx, 10, 10)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, items, 0)
	})

}

func TestLocalCWEStore_GetByID_NotFound(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestLocalCWEStore_GetByID_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		store, err := NewLocalCWEStore(":memory:")
		require.NoError(t, err)

		ctx := context.Background()
		item, err := store.GetByID(ctx, "DOES-NOT-EXIST")
		require.Error(t, err)
		require.Nil(t, item)
	})

}

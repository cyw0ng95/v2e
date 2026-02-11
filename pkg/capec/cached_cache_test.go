//go:build CONFIG_USE_LIBXML2

package capec

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestSetAndGetCachedCAPEC(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetAndGetCachedCAPEC", nil, func(t *testing.T, tx *gorm.DB) {
		s := &CachedLocalCAPECStore{
			cache: make(map[int]*capecCacheItem),
			ttl:   10 * time.Minute,
		}

		item := &CAPECItemModel{CAPECID: 123, Name: "test"}
		s.setCachedCAPEC(123, item)

		got, ok := s.getCachedCAPEC(123)
		if !ok {
			t.Fatalf("expected cached item present")
		}
		if got.CAPECID != 123 || got.Name != "test" {
			t.Fatalf("unexpected cached item: %+v", got)
		}
	})

}

func TestInvalidateCachedCAPECAndInvalidateAllCache(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestInvalidateCachedCAPECAndInvalidateAllCache", nil, func(t *testing.T, tx *gorm.DB) {
		s := &CachedLocalCAPECStore{
			cache: make(map[int]*capecCacheItem),
			ttl:   10 * time.Minute,
		}

		s.setCachedCAPEC(1, &CAPECItemModel{CAPECID: 1})
		s.setCachedCAPEC(2, &CAPECItemModel{CAPECID: 2})

		s.invalidateCachedCAPEC(1)
		if _, ok := s.getCachedCAPEC(1); ok {
			t.Fatalf("expected entry 1 invalidated")
		}

		s.invalidateAllCache()
		if _, ok := s.getCachedCAPEC(2); ok {
			t.Fatalf("expected cache to be empty after invalidateAllCache")
		}
	})

}

func TestGetCachedCAPEC_TTLExpiryNoSleep(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestGetCachedCAPEC_TTLExpiryNoSleep", nil, func(t *testing.T, tx *gorm.DB) {
		s := &CachedLocalCAPECStore{
			cache: make(map[int]*capecCacheItem),
			ttl:   5 * time.Minute,
		}
		// insert an item with an old timestamp
		s.cache[42] = &capecCacheItem{data: &CAPECItemModel{CAPECID: 42}, timestamp: time.Now().Add(-time.Hour)}

		if _, ok := s.getCachedCAPEC(42); ok {
			t.Fatalf("expected cached item to be considered expired without sleeping")
		}
	})

}

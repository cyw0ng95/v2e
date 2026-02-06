package permits

import (
	"os"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
)

func TestNewPermitManager(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "[TEST] ", common.DebugLevel)

	tests := []struct {
		name          string
		totalPermits  int
		wantTotal     int
		wantAvailable int
	}{
		{
			name:          "valid total permits",
			totalPermits:  10,
			wantTotal:     10,
			wantAvailable: 10,
		},
		{
			name:          "zero permits defaults to 10",
			totalPermits:  0,
			wantTotal:     10,
			wantAvailable: 10,
		},
		{
			name:          "negative permits defaults to 10",
			totalPermits:  -5,
			wantTotal:     10,
			wantAvailable: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPermitManager(tt.totalPermits, logger)
			if pm.totalPermits != tt.wantTotal {
				t.Errorf("totalPermits = %d, want %d", pm.totalPermits, tt.wantTotal)
			}
			if pm.availablePermits != tt.wantAvailable {
				t.Errorf("availablePermits = %d, want %d", pm.availablePermits, tt.wantAvailable)
			}
		})
	}
}

func TestRequestPermits(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "[TEST] ", common.DebugLevel)

	tests := []struct {
		name          string
		totalPermits  int
		request       *PermitRequest
		wantGranted   int
		wantAvailable int
		wantErr       bool
	}{
		{
			name:         "valid request within available",
			totalPermits: 10,
			request: &PermitRequest{
				ProviderID:  "provider-1",
				PermitCount: 5,
			},
			wantGranted:   5,
			wantAvailable: 5,
			wantErr:       false,
		},
		{
			name:         "request exceeds available",
			totalPermits: 10,
			request: &PermitRequest{
				ProviderID:  "provider-1",
				PermitCount: 15,
			},
			wantGranted:   10,
			wantAvailable: 0,
			wantErr:       false,
		},
		{
			name:         "empty provider ID",
			totalPermits: 10,
			request: &PermitRequest{
				ProviderID:  "",
				PermitCount: 5,
			},
			wantErr: true,
		},
		{
			name:         "zero permit count",
			totalPermits: 10,
			request: &PermitRequest{
				ProviderID:  "provider-1",
				PermitCount: 0,
			},
			wantErr: true,
		},
		{
			name:         "negative permit count",
			totalPermits: 10,
			request: &PermitRequest{
				ProviderID:  "provider-1",
				PermitCount: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPermitManager(tt.totalPermits, logger)
			resp, err := pm.RequestPermits(tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Granted != tt.wantGranted {
				t.Errorf("granted = %d, want %d", resp.Granted, tt.wantGranted)
			}
			if resp.Available != tt.wantAvailable {
				t.Errorf("available = %d, want %d", resp.Available, tt.wantAvailable)
			}
			if pm.availablePermits != tt.wantAvailable {
				t.Errorf("pm.availablePermits = %d, want %d", pm.availablePermits, tt.wantAvailable)
			}
		})
	}
}

func TestReleasePermits(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "[TEST] ", common.DebugLevel)

	t.Run("release valid allocation", func(t *testing.T) {
		pm := NewPermitManager(10, logger)

		// Allocate 5 permits
		_, err := pm.RequestPermits(&PermitRequest{
			ProviderID:  "provider-1",
			PermitCount: 5,
		})
		if err != nil {
			t.Fatalf("failed to request permits: %v", err)
		}

		// Release 3 permits
		resp, err := pm.ReleasePermits("provider-1", 3)
		if err != nil {
			t.Fatalf("failed to release permits: %v", err)
		}

		if resp.Granted != 3 {
			t.Errorf("released = %d, want 3", resp.Granted)
		}
		if resp.Available != 8 {
			t.Errorf("available = %d, want 8", resp.Available)
		}
		if pm.allocations["provider-1"] != 2 {
			t.Errorf("remaining allocation = %d, want 2", pm.allocations["provider-1"])
		}
	})

	t.Run("release all permits removes allocation", func(t *testing.T) {
		pm := NewPermitManager(10, logger)

		// Allocate 5 permits
		pm.RequestPermits(&PermitRequest{
			ProviderID:  "provider-1",
			PermitCount: 5,
		})

		// Release all 5 permits
		_, err := pm.ReleasePermits("provider-1", 5)
		if err != nil {
			t.Fatalf("failed to release permits: %v", err)
		}

		if _, exists := pm.allocations["provider-1"]; exists {
			t.Error("allocation should be removed after releasing all permits")
		}
		if pm.availablePermits != 10 {
			t.Errorf("availablePermits = %d, want 10", pm.availablePermits)
		}
	})

	t.Run("release more than allocated", func(t *testing.T) {
		pm := NewPermitManager(10, logger)

		pm.RequestPermits(&PermitRequest{
			ProviderID:  "provider-1",
			PermitCount: 5,
		})

		// Try to release more than allocated
		resp, err := pm.ReleasePermits("provider-1", 10)
		if err != nil {
			t.Fatalf("failed to release permits: %v", err)
		}

		// Should only release what was allocated
		if resp.Granted != 5 {
			t.Errorf("released = %d, want 5", resp.Granted)
		}
		if pm.availablePermits != 10 {
			t.Errorf("availablePermits = %d, want 10", pm.availablePermits)
		}
	})

	t.Run("release from non-existent provider", func(t *testing.T) {
		pm := NewPermitManager(10, logger)

		_, err := pm.ReleasePermits("non-existent", 5)
		if err == nil {
			t.Error("expected error for non-existent provider, got nil")
		}
	})
}

func TestRevokePermits(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "[TEST] ", common.DebugLevel)

	t.Run("revoke from single provider", func(t *testing.T) {
		pm := NewPermitManager(10, logger)

		pm.RequestPermits(&PermitRequest{
			ProviderID:  "provider-1",
			PermitCount: 8,
		})

		revocations := pm.RevokePermits(3)

		if len(revocations) != 1 {
			t.Errorf("revocations count = %d, want 1", len(revocations))
		}
		if revocations["provider-1"] != 3 {
			t.Errorf("revoked from provider-1 = %d, want 3", revocations["provider-1"])
		}
		if pm.availablePermits != 5 {
			t.Errorf("availablePermits = %d, want 5", pm.availablePermits)
		}
		if pm.allocations["provider-1"] != 5 {
			t.Errorf("remaining allocation = %d, want 5", pm.allocations["provider-1"])
		}
	})

	t.Run("revoke from multiple providers", func(t *testing.T) {
		pm := NewPermitManager(10, logger)

		pm.RequestPermits(&PermitRequest{ProviderID: "provider-1", PermitCount: 4})
		pm.RequestPermits(&PermitRequest{ProviderID: "provider-2", PermitCount: 4})

		revocations := pm.RevokePermits(4)

		// Should revoke proportionally (2 from each)
		totalRevoked := 0
		for _, count := range revocations {
			totalRevoked += count
		}
		if totalRevoked != 4 {
			t.Errorf("total revoked = %d, want 4", totalRevoked)
		}

		// Check total available increased
		if pm.availablePermits != 6 {
			t.Errorf("availablePermits = %d, want 6", pm.availablePermits)
		}
	})

	t.Run("revoke with no allocations", func(t *testing.T) {
		pm := NewPermitManager(10, logger)

		revocations := pm.RevokePermits(5)

		if revocations != nil {
			t.Error("expected nil revocations when no allocations exist")
		}
	})
}

func TestGetStats(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "[TEST] ", common.DebugLevel)
	pm := NewPermitManager(10, logger)

	// Make some allocations
	pm.RequestPermits(&PermitRequest{ProviderID: "provider-1", PermitCount: 5})
	pm.RequestPermits(&PermitRequest{ProviderID: "provider-2", PermitCount: 2})
	pm.ReleasePermits("provider-1", 1)
	pm.RevokePermits(1)

	stats := pm.GetStats()

	if stats["total_permits"].(int) != 10 {
		t.Errorf("total_permits = %v, want 10", stats["total_permits"])
	}
	if stats["total_requests"].(int64) != 2 {
		t.Errorf("total_requests = %v, want 2", stats["total_requests"])
	}
	if stats["total_releases"].(int64) != 1 {
		t.Errorf("total_releases = %v, want 1", stats["total_releases"])
	}
}

func TestSetTotalPermits(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "[TEST] ", common.DebugLevel)

	t.Run("increase total permits", func(t *testing.T) {
		pm := NewPermitManager(10, logger)

		err := pm.SetTotalPermits(20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pm.totalPermits != 20 {
			t.Errorf("totalPermits = %d, want 20", pm.totalPermits)
		}
		if pm.availablePermits != 20 {
			t.Errorf("availablePermits = %d, want 20", pm.availablePermits)
		}
	})

	t.Run("cannot reduce below allocated", func(t *testing.T) {
		pm := NewPermitManager(10, logger)

		pm.RequestPermits(&PermitRequest{ProviderID: "provider-1", PermitCount: 8})

		err := pm.SetTotalPermits(5)
		if err == nil {
			t.Error("expected error when reducing below allocated, got nil")
		}
	})

	t.Run("invalid total permits", func(t *testing.T) {
		pm := NewPermitManager(10, logger)

		err := pm.SetTotalPermits(0)
		if err == nil {
			t.Error("expected error for zero total, got nil")
		}

		err = pm.SetTotalPermits(-5)
		if err == nil {
			t.Error("expected error for negative total, got nil")
		}
	})
}

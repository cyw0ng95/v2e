package proc

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCPUAffinityMask(t *testing.T) {
	tests := []struct {
		name        string
		mask        string
		expectedCPUs []int
		wantErr     bool
	}{
		{
			name:        "Empty mask",
			mask:        "",
			expectedCPUs: nil,
			wantErr:     false,
		},
		{
			name:        "Single CPU (core 0)",
			mask:        "0x01",
			expectedCPUs: []int{0},
			wantErr:     false,
		},
		{
			name:        "Two CPUs (cores 0-1)",
			mask:        "0x03",
			expectedCPUs: []int{0, 1},
			wantErr:     false,
		},
		{
			name:        "Four CPUs (cores 0-3)",
			mask:        "0x0F",
			expectedCPUs: []int{0, 1, 2, 3},
			wantErr:     false,
		},
		{
			name:        "Non-contiguous CPUs (cores 0, 2, 4)",
			mask:        "0x15",
			expectedCPUs: []int{0, 2, 4},
			wantErr:     false,
		},
		{
			name:        "Mask without 0x prefix",
			mask:        "03",
			expectedCPUs: []int{0, 1},
			wantErr:     false,
		},
		{
			name:        "Invalid mask",
			mask:        "invalid",
			expectedCPUs: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpus, err := ParseCPUAffinityMask(tt.mask)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCPUs, cpus)
		})
	}
}

func TestSetCPUAffinity(t *testing.T) {
	// Skip if not enough CPUs for testing
	if runtime.NumCPU() < 2 {
		t.Skip("Skipping CPU affinity test: need at least 2 CPUs")
	}

	tests := []struct {
		name    string
		mask    string
		wantErr bool
	}{
		{
			name:    "Empty mask (no affinity)",
			mask:    "",
			wantErr: false,
		},
		{
			name:    "Valid single CPU",
			mask:    "0x01",
			wantErr: false,
		},
		{
			name:    "Valid multiple CPUs",
			mask:    "0x03",
			wantErr: false,
		},
		{
			name:    "Zero mask",
			mask:    "0x00",
			wantErr: true,
		},
		{
			name:    "Invalid mask",
			mask:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetCPUAffinity(tt.mask)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// If we set an affinity, verify it was applied
			if tt.mask != "" {
				currentMask, err := GetCPUAffinity()
				require.NoError(t, err)
				t.Logf("Set affinity to %s, current affinity: %s", tt.mask, currentMask)
				
				// Parse both masks to compare
				expectedCPUs, err := ParseCPUAffinityMask(tt.mask)
				require.NoError(t, err)
				
				currentCPUs, err := ParseCPUAffinityMask(currentMask)
				require.NoError(t, err)
				
				// Current affinity should include all expected CPUs
				for _, cpu := range expectedCPUs {
					assert.Contains(t, currentCPUs, cpu, "CPU %d should be in affinity", cpu)
				}
			}
		})
	}
}

func TestGetCPUAffinity(t *testing.T) {
	mask, err := GetCPUAffinity()
	require.NoError(t, err)
	assert.NotEmpty(t, mask)
	
	t.Logf("Current CPU affinity: %s", mask)

	// Verify we can parse the returned mask
	cpus, err := ParseCPUAffinityMask(mask)
	require.NoError(t, err)
	assert.NotEmpty(t, cpus, "Should have at least one CPU in affinity")
}

func TestSetAndGetCPUAffinity(t *testing.T) {
	if runtime.NumCPU() < 2 {
		t.Skip("Skipping: need at least 2 CPUs")
	}

	// Set affinity to first CPU
	err := SetCPUAffinity("0x01")
	require.NoError(t, err)

	// Get and verify
	mask, err := GetCPUAffinity()
	require.NoError(t, err)

	cpus, err := ParseCPUAffinityMask(mask)
	require.NoError(t, err)
	
	// Should contain CPU 0
	assert.Contains(t, cpus, 0, "Should be pinned to CPU 0")
	t.Logf("Affinity set to CPU(s): %v", cpus)
}

package cve

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNVDTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkVal bool
		expected time.Time
	}{
		{
			name:     "valid NVD format",
			input:    `"2021-12-10T10:15:09.143"`,
			wantErr:  false,
			checkVal: true,
			expected: time.Date(2021, 12, 10, 10, 15, 9, 143000000, time.UTC),
		},
		{
			name:     "valid RFC3339 format",
			input:    `"2021-12-10T10:15:09Z"`,
			wantErr:  false,
			checkVal: true,
			expected: time.Date(2021, 12, 10, 10, 15, 9, 0, time.UTC),
		},
		{
			name:     "null value",
			input:    `null`,
			wantErr:  false,
			checkVal: true,
			expected: time.Time{},
		},
		{
			name:     "empty string",
			input:    `""`,
			wantErr:  false,
			checkVal: true,
			expected: time.Time{},
		},
		{
			name:    "invalid format",
			input:   `"invalid-date"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nvdTime NVDTime
			err := json.Unmarshal([]byte(tt.input), &nvdTime)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.checkVal && !tt.wantErr {
				if !nvdTime.Time.Equal(tt.expected) {
					t.Errorf("UnmarshalJSON() got = %v, want %v", nvdTime.Time, tt.expected)
				}
			}
		})
	}
}

func TestNVDTime_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "valid time",
			time:     time.Date(2021, 12, 10, 10, 15, 9, 143000000, time.UTC),
			expected: `"2021-12-10T10:15:09.143"`,
		},
		{
			name:     "zero time",
			time:     time.Time{},
			expected: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nvdTime := NVDTime{Time: tt.time}
			data, err := json.Marshal(nvdTime)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			
			if string(data) != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", string(data), tt.expected)
			}
		})
	}
}

func TestNewNVDTime(t *testing.T) {
	now := time.Now()
	nvdTime := NewNVDTime(now)
	
	if !nvdTime.Time.Equal(now) {
		t.Errorf("NewNVDTime() = %v, want %v", nvdTime.Time, now)
	}
}

func TestNVDTime_RoundTrip(t *testing.T) {
	original := NVDTime{Time: time.Date(2021, 12, 10, 10, 15, 9, 143000000, time.UTC)}
	
	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	
	// Unmarshal
	var decoded NVDTime
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	
	// Compare (truncate to millisecond precision)
	originalTrunc := original.Time.Truncate(time.Millisecond)
	decodedTrunc := decoded.Time.Truncate(time.Millisecond)
	
	if !originalTrunc.Equal(decodedTrunc) {
		t.Errorf("Round trip failed: got %v, want %v", decoded.Time, original.Time)
	}
}

package proc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodingType_String(t *testing.T) {
	tests := []struct {
		name     string
		encoding EncodingType
		expected string
	}{
		{"None", EncodingNone, "none"},
		{"JSON", EncodingJSON, "json"},
		{"Binary", EncodingBinary, "binary"},
		{"ZSTD", EncodingZSTD, "zstd"},
		{"LZ4", EncodingLZ4, "lz4"},
		{"Unknown", EncodingType(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.encoding.String())
		})
	}
}

func TestCompressDecompressZSTD(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte{}},
		{"Small", []byte("hello world")},
		{"Medium", bytes.Repeat([]byte("test data "), 100)},
		{"Large", bytes.Repeat([]byte("large test data with repetition "), 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compress
			compressed, err := CompressZSTD(tt.data)
			require.NoError(t, err)

			// Verify compression occurred for larger data
			if len(tt.data) > 100 {
				assert.Less(t, len(compressed), len(tt.data), "Compression should reduce size")
			}

			// Decompress
			decompressed, err := DecompressZSTD(compressed)
			require.NoError(t, err)

			// Verify data matches (handle nil vs empty slice)
			if len(tt.data) == 0 {
				assert.Empty(t, decompressed)
			} else {
				assert.Equal(t, tt.data, decompressed)
			}
		})
	}
}

func TestCompressDecompressLZ4(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte{}},
		{"Small", []byte("hello world")},
		{"Medium", bytes.Repeat([]byte("test data "), 100)},
		{"Large", bytes.Repeat([]byte("large test data with repetition "), 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compress
			compressed, err := CompressLZ4(tt.data)
			require.NoError(t, err)

			// Decompress
			decompressed, err := DecompressLZ4(compressed)
			require.NoError(t, err)

			// Verify data matches (handle nil vs empty slice)
			if len(tt.data) == 0 {
				assert.Empty(t, decompressed)
			} else {
				assert.Equal(t, tt.data, decompressed)
			}
		})
	}
}

func TestCompressWithEncoding(t *testing.T) {
	testData := []byte("test data for compression")

	tests := []struct {
		name     string
		encoding EncodingType
		wantErr  bool
	}{
		{"None", EncodingNone, false},
		{"JSON", EncodingJSON, false},
		{"Binary", EncodingBinary, false},
		{"ZSTD", EncodingZSTD, false},
		{"LZ4", EncodingLZ4, false},
		{"Invalid", EncodingType(99), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := CompressWithEncoding(testData, tt.encoding)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, compressed)

			// For no-compression types, data should be unchanged
			if tt.encoding == EncodingNone || tt.encoding == EncodingJSON || tt.encoding == EncodingBinary {
				assert.Equal(t, testData, compressed)
			}
		})
	}
}

func TestDecompressWithEncoding(t *testing.T) {
	testData := []byte("test data for decompression")

	tests := []struct {
		name     string
		encoding EncodingType
		setup    func() []byte
		wantErr  bool
	}{
		{
			"None",
			EncodingNone,
			func() []byte { return testData },
			false,
		},
		{
			"JSON",
			EncodingJSON,
			func() []byte { return testData },
			false,
		},
		{
			"Binary",
			EncodingBinary,
			func() []byte { return testData },
			false,
		},
		{
			"ZSTD",
			EncodingZSTD,
			func() []byte {
				compressed, _ := CompressZSTD(testData)
				return compressed
			},
			false,
		},
		{
			"LZ4",
			EncodingLZ4,
			func() []byte {
				compressed, _ := CompressLZ4(testData)
				return compressed
			},
			false,
		},
		{
			"Invalid",
			EncodingType(99),
			func() []byte { return testData },
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.setup()
			decompressed, err := DecompressWithEncoding(data, tt.encoding)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testData, decompressed)
		})
	}
}

func TestCompressionRatio(t *testing.T) {
	// Create highly compressible data
	repeatData := bytes.Repeat([]byte("this is highly repetitive data "), 1000)

	t.Run("ZSTD", func(t *testing.T) {
		compressed, err := CompressZSTD(repeatData)
		require.NoError(t, err)
		
		ratio := float64(len(compressed)) / float64(len(repeatData))
		t.Logf("ZSTD compression ratio: %.2f%% (original: %d, compressed: %d)", 
			ratio*100, len(repeatData), len(compressed))
		
		// Should achieve significant compression
		assert.Less(t, ratio, 0.1, "ZSTD should compress repetitive data to <10%")
	})

	t.Run("LZ4", func(t *testing.T) {
		compressed, err := CompressLZ4(repeatData)
		require.NoError(t, err)
		
		ratio := float64(len(compressed)) / float64(len(repeatData))
		t.Logf("LZ4 compression ratio: %.2f%% (original: %d, compressed: %d)", 
			ratio*100, len(repeatData), len(compressed))
		
		// LZ4 should still achieve good compression
		assert.Less(t, ratio, 0.2, "LZ4 should compress repetitive data to <20%")
	})
}

func TestErrorCases(t *testing.T) {
	t.Run("Invalid ZSTD data", func(t *testing.T) {
		invalidData := []byte("not compressed data")
		_, err := DecompressZSTD(invalidData)
		assert.Error(t, err)
	})

	t.Run("Invalid LZ4 data", func(t *testing.T) {
		invalidData := []byte("not compressed data")
		_, err := DecompressLZ4(invalidData)
		assert.Error(t, err)
	})
}

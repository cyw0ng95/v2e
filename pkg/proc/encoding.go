package proc

import (
	"bytes"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
)

// EncodingType represents the encoding type for message payloads
type EncodingType byte

const (
	// EncodingNone represents no encoding/compression
	EncodingNone EncodingType = 0
	// EncodingJSON represents JSON encoding (legacy/default)
	EncodingJSON EncodingType = 1
	// EncodingBinary represents binary encoding
	EncodingBinary EncodingType = 2
	// EncodingZSTD represents Zstandard compression
	EncodingZSTD EncodingType = 3
	// EncodingLZ4 represents LZ4 compression
	EncodingLZ4 EncodingType = 4
)

// String returns the string representation of the encoding type
func (e EncodingType) String() string {
	switch e {
	case EncodingNone:
		return "none"
	case EncodingJSON:
		return "json"
	case EncodingBinary:
		return "binary"
	case EncodingZSTD:
		return "zstd"
	case EncodingLZ4:
		return "lz4"
	default:
		return fmt.Sprintf("unknown(%d)", e)
	}
}

// CompressZSTD compresses data using Zstandard compression
func CompressZSTD(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	encoder, err := zstd.NewWriter(&buf, zstd.WithEncoderLevel(zstd.SpeedFastest))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
	}
	defer encoder.Close()

	if _, err := encoder.Write(data); err != nil {
		return nil, fmt.Errorf("failed to compress with zstd: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zstd encoder: %w", err)
	}

	return buf.Bytes(), nil
}

// DecompressZSTD decompresses data using Zstandard compression
func DecompressZSTD(data []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
	}
	defer decoder.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, decoder); err != nil {
		return nil, fmt.Errorf("failed to decompress with zstd: %w", err)
	}

	return buf.Bytes(), nil
}

// CompressLZ4 compresses data using LZ4 compression
func CompressLZ4(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := lz4.NewWriter(&buf)
	
	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed to compress with lz4: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close lz4 writer: %w", err)
	}

	return buf.Bytes(), nil
}

// DecompressLZ4 decompresses data using LZ4 compression
func DecompressLZ4(data []byte) ([]byte, error) {
	reader := lz4.NewReader(bytes.NewReader(data))
	
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, fmt.Errorf("failed to decompress with lz4: %w", err)
	}

	return buf.Bytes(), nil
}

// CompressWithEncoding compresses data using the specified encoding type
func CompressWithEncoding(data []byte, encoding EncodingType) ([]byte, error) {
	switch encoding {
	case EncodingNone, EncodingJSON, EncodingBinary:
		// No compression
		return data, nil
	case EncodingZSTD:
		return CompressZSTD(data)
	case EncodingLZ4:
		return CompressLZ4(data)
	default:
		return nil, fmt.Errorf("unsupported encoding type: %d", encoding)
	}
}

// DecompressWithEncoding decompresses data using the specified encoding type
func DecompressWithEncoding(data []byte, encoding EncodingType) ([]byte, error) {
	switch encoding {
	case EncodingNone, EncodingJSON, EncodingBinary:
		// No decompression
		return data, nil
	case EncodingZSTD:
		return DecompressZSTD(data)
	case EncodingLZ4:
		return DecompressLZ4(data)
	default:
		return nil, fmt.Errorf("unsupported encoding type: %d", encoding)
	}
}

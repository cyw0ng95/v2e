package proc

import (
	"fmt"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestBinaryHeader_NewBinaryHeader(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBinaryHeader_NewBinaryHeader", nil, func(t *testing.T, tx *gorm.DB) {
		header := NewBinaryHeader()

		if header.Magic[0] != MagicByte1 || header.Magic[1] != MagicByte2 {
			t.Errorf("Expected magic bytes [%02x %02x], got [%02x %02x]",
				MagicByte1, MagicByte2, header.Magic[0], header.Magic[1])
		}

		if header.Version != ProtocolVersion {
			t.Errorf("Expected version %d, got %d", ProtocolVersion, header.Version)
		}
	})
}

func TestBinaryHeader_EncodeDecodeRoundTrip(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBinaryHeader_EncodeDecodeRoundTrip", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a header with test data
		original := NewBinaryHeader()
		original.Encoding = EncodingJSON
		original.MsgType = BinaryMessageTypeRequest
		original.PayloadLen = 1234
		original.SetMessageID("test-msg-id")
		original.SetSourceID("source-process")
		original.SetTargetID("target-process")
		original.SetCorrelationID("corr-123")

		// Encode
		buf := make([]byte, HeaderSize)
		if err := original.EncodeHeader(buf); err != nil {
			t.Fatalf("EncodeHeader failed: %v", err)
		}

		// Decode
		decoded, err := DecodeHeader(buf)
		if err != nil {
			t.Fatalf("DecodeHeader failed: %v", err)
		}

		// Verify fields
		if decoded.Magic[0] != original.Magic[0] || decoded.Magic[1] != original.Magic[1] {
			t.Errorf("Magic mismatch: expected [%02x %02x], got [%02x %02x]",
				original.Magic[0], original.Magic[1], decoded.Magic[0], decoded.Magic[1])
		}

		if decoded.Version != original.Version {
			t.Errorf("Version mismatch: expected %d, got %d", original.Version, decoded.Version)
		}

		if decoded.Encoding != original.Encoding {
			t.Errorf("Encoding mismatch: expected %d, got %d", original.Encoding, decoded.Encoding)
		}

		if decoded.MsgType != original.MsgType {
			t.Errorf("MsgType mismatch: expected %d, got %d", original.MsgType, decoded.MsgType)
		}

		if decoded.PayloadLen != original.PayloadLen {
			t.Errorf("PayloadLen mismatch: expected %d, got %d", original.PayloadLen, decoded.PayloadLen)
		}

		if decoded.GetMessageID() != original.GetMessageID() {
			t.Errorf("MessageID mismatch: expected %s, got %s", original.GetMessageID(), decoded.GetMessageID())
		}

		if decoded.GetSourceID() != original.GetSourceID() {
			t.Errorf("SourceID mismatch: expected %s, got %s", original.GetSourceID(), decoded.GetSourceID())
		}

		if decoded.GetTargetID() != original.GetTargetID() {
			t.Errorf("TargetID mismatch: expected %s, got %s", original.GetTargetID(), decoded.GetTargetID())
		}

		if decoded.GetCorrelationID() != original.GetCorrelationID() {
			t.Errorf("CorrelationID mismatch: expected %s, got %s", original.GetCorrelationID(), decoded.GetCorrelationID())
		}
	})
}

func TestBinaryHeader_InvalidMagic(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBinaryHeader_InvalidMagic", nil, func(t *testing.T, tx *gorm.DB) {
		buf := make([]byte, HeaderSize)
		buf[0] = 0xFF // Invalid magic byte
		buf[1] = 0xFF

		_, err := DecodeHeader(buf)
		if err == nil {
			t.Error("Expected error for invalid magic bytes, got nil")
		}
	})
}

func TestBinaryHeader_BufferTooSmall(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBinaryHeader_BufferTooSmall", nil, func(t *testing.T, tx *gorm.DB) {
		header := NewBinaryHeader()
		buf := make([]byte, 64) // Too small

		err := header.EncodeHeader(buf)
		if err == nil {
			t.Error("Expected error for buffer too small, got nil")
		}

		_, err = DecodeHeader(buf)
		if err == nil {
			t.Error("Expected error for buffer too small on decode, got nil")
		}
	})
}

func TestBinaryMessage_MarshalUnmarshalJSON(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBinaryMessage_MarshalUnmarshalJSON", nil, func(t *testing.T, tx *gorm.DB) {
		type TestPayload struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}

		payload := TestPayload{
			Command: "echo",
			Args:    []string{"hello", "world"},
		}

		// Create original message
		original, err := NewRequestMessage("req-1", payload)
		if err != nil {
			t.Fatalf("NewRequestMessage failed: %v", err)
		}
		original.Source = "source-proc"
		original.Target = "target-proc"
		original.CorrelationID = "corr-123"

		// Marshal to binary
		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary failed: %v", err)
		}

		// Verify it's a binary message
		if !IsBinaryMessage(data) {
			t.Error("IsBinaryMessage returned false for binary message")
		}

		// Unmarshal from binary
		decoded, err := UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary failed: %v", err)
		}

		// Verify fields
		if decoded.Type != original.Type {
			t.Errorf("Type mismatch: expected %s, got %s", original.Type, decoded.Type)
		}

		if decoded.ID != original.ID {
			t.Errorf("ID mismatch: expected %s, got %s", original.ID, decoded.ID)
		}

		if decoded.Source != original.Source {
			t.Errorf("Source mismatch: expected %s, got %s", original.Source, decoded.Source)
		}

		if decoded.Target != original.Target {
			t.Errorf("Target mismatch: expected %s, got %s", original.Target, decoded.Target)
		}

		if decoded.CorrelationID != original.CorrelationID {
			t.Errorf("CorrelationID mismatch: expected %s, got %s", original.CorrelationID, decoded.CorrelationID)
		}

		// Unmarshal and verify payload
		var decodedPayload TestPayload
		if err := decoded.UnmarshalPayload(&decodedPayload); err != nil {
			t.Fatalf("UnmarshalPayload failed: %v", err)
		}

		if decodedPayload.Command != payload.Command {
			t.Errorf("Command mismatch: expected %s, got %s", payload.Command, decodedPayload.Command)
		}

		if len(decodedPayload.Args) != len(payload.Args) {
			t.Errorf("Args length mismatch: expected %d, got %d", len(payload.Args), len(decodedPayload.Args))
		}
	})
}

func TestBinaryMessage_MarshalUnmarshalGOB(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBinaryMessage_MarshalUnmarshalGOB", nil, func(t *testing.T, tx *gorm.DB) {
		type TestPayload struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}

		payload := TestPayload{
			Command: "test",
			Args:    []string{"arg1", "arg2"},
		}

		// Create original message
		original, err := NewRequestMessage("req-1", payload)
		if err != nil {
			t.Fatalf("NewRequestMessage failed: %v", err)
		}
		original.Source = "source"
		original.Target = "target"

		// Marshal to binary with GOB encoding
		data, err := MarshalBinaryWithEncoding(original, EncodingGOB)
		if err != nil {
			t.Fatalf("MarshalBinaryWithEncoding(GOB) failed: %v", err)
		}

		// Unmarshal from binary
		decoded, err := UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary failed: %v", err)
		}

		// Verify basic fields
		if decoded.Type != original.Type {
			t.Errorf("Type mismatch: expected %s, got %s", original.Type, decoded.Type)
		}

		if decoded.ID != original.ID {
			t.Errorf("ID mismatch: expected %s, got %s", original.ID, decoded.ID)
		}

		// Unmarshal payload (should be converted back to JSON internally)
		var decodedPayload TestPayload
		if err := decoded.UnmarshalPayload(&decodedPayload); err != nil {
			t.Fatalf("UnmarshalPayload failed: %v", err)
		}

		if decodedPayload.Command != payload.Command {
			t.Errorf("Command mismatch: expected %s, got %s", payload.Command, decodedPayload.Command)
		}
	})
}

func TestBinaryMessage_ErrorMessage(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBinaryMessage_ErrorMessage", nil, func(t *testing.T, tx *gorm.DB) {
		original := NewErrorMessage("err-1", fmt.Errorf("test error"))
		original.Source = "source"
		original.Target = "target"

		// Marshal to binary
		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary failed: %v", err)
		}

		// Unmarshal from binary
		decoded, err := UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary failed: %v", err)
		}

		// Verify error message
		if decoded.Type != MessageTypeError {
			t.Errorf("Expected type %s, got %s", MessageTypeError, decoded.Type)
		}

		if decoded.Error != "test error" {
			t.Errorf("Error mismatch: expected 'test error', got '%s'", decoded.Error)
		}
	})
}

func TestBinaryMessage_PlainEncoding(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBinaryMessage_PlainEncoding", nil, func(t *testing.T, tx *gorm.DB) {
		// Create message with plain text payload
		original := GetMessage()
		original.Type = MessageTypeEvent
		original.ID = "evt-1"
		original.Payload = []byte("plain text message")
		original.Source = "source"

		// Marshal with PLAIN encoding
		data, err := MarshalBinaryWithEncoding(original, EncodingPLAIN)
		if err != nil {
			t.Fatalf("MarshalBinaryWithEncoding(PLAIN) failed: %v", err)
		}

		// Unmarshal
		decoded, err := UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary failed: %v", err)
		}

		// Verify payload
		if string(decoded.Payload) != "plain text message" {
			t.Errorf("Payload mismatch: expected 'plain text message', got '%s'", string(decoded.Payload))
		}
	})
}

func TestBinaryMessage_IsBinaryMessage(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBinaryMessage_IsBinaryMessage", nil, func(t *testing.T, tx *gorm.DB) {
		// Test with binary message
		msg, _ := NewRequestMessage("req-1", map[string]string{"key": "value"})
		binaryData, _ := msg.MarshalBinary()

		if !IsBinaryMessage(binaryData) {
			t.Error("IsBinaryMessage returned false for binary message")
		}

		// Test with JSON message
		jsonData, _ := msg.Marshal()
		if IsBinaryMessage(jsonData) {
			t.Error("IsBinaryMessage returned true for JSON message")
		}

		// Test with short buffer
		shortBuf := []byte{0x56}
		if IsBinaryMessage(shortBuf) {
			t.Error("IsBinaryMessage returned true for short buffer")
		}

		// Test with empty buffer
		if IsBinaryMessage([]byte{}) {
			t.Error("IsBinaryMessage returned true for empty buffer")
		}
	})
}

func TestBinaryMessage_LongStrings(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBinaryMessage_LongStrings", nil, func(t *testing.T, tx *gorm.DB) {
		// Test with strings longer than fixed fields
		longID := "this-is-a-very-long-message-id-that-exceeds-32-bytes-and-should-be-truncated"
		longSource := "this-is-a-very-long-source-id-that-exceeds-32-bytes"
		longTarget := "this-is-a-very-long-target-id-that-exceeds-32-bytes"
		longCorrelation := "this-is-a-very-long-correlation-id-exceeding-20-bytes"

		msg := GetMessage()
		msg.Type = MessageTypeRequest
		msg.ID = longID
		msg.Source = longSource
		msg.Target = longTarget
		msg.CorrelationID = longCorrelation

		// Marshal
		data, err := msg.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary failed: %v", err)
		}

		// Unmarshal
		decoded, err := UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("UnmarshalBinary failed: %v", err)
		}

		// Verify truncation (should be truncated to field size)
		if len(decoded.ID) > 32 {
			t.Errorf("ID not truncated: length %d", len(decoded.ID))
		}

		if len(decoded.Source) > 32 {
			t.Errorf("Source not truncated: length %d", len(decoded.Source))
		}

		if len(decoded.Target) > 32 {
			t.Errorf("Target not truncated: length %d", len(decoded.Target))
		}

		if len(decoded.CorrelationID) > 20 {
			t.Errorf("CorrelationID not truncated: length %d", len(decoded.CorrelationID))
		}

		// Verify values are truncated versions
		if decoded.ID != longID[:32] {
			t.Errorf("ID mismatch after truncation")
		}
	})
}

func TestMemcpy_LengthMismatch(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMemcpy_LengthMismatch", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			name    string
			dst     []byte
			src     []byte
			wantErr bool
		}{
			{
				name:    "equal lengths",
				dst:     make([]byte, 10),
				src:     []byte("0123456789"),
				wantErr: false,
			},
			{
				name:    "dst shorter than src",
				dst:     make([]byte, 5),
				src:     []byte("0123456789"),
				wantErr: true,
			},
			{
				name:    "dst longer than src",
				dst:     make([]byte, 15),
				src:     []byte("0123456789"),
				wantErr: true,
			},
			{
				name:    "both empty",
				dst:     []byte{},
				src:     []byte{},
				wantErr: false,
			},
			{
				name:    "empty dst non-empty src",
				dst:     []byte{},
				src:     []byte("data"),
				wantErr: true,
			},
			{
				name:    "non-empty dst empty src",
				dst:     make([]byte, 4),
				src:     []byte{},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := Memcpy(tt.dst, tt.src)
				if (err != nil) != tt.wantErr {
					t.Errorf("Memcpy() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.wantErr && err != nil {
					// Verify error message contains useful information
					if !containsSubstring(err.Error(), "length mismatch") {
						t.Errorf("Error message should contain 'length mismatch', got: %v", err)
					}
				}
			})
		}
	})
}

func TestMemcpy_SuccessfulCopy(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMemcpy_SuccessfulCopy", nil, func(t *testing.T, tx *gorm.DB) {
		src := []byte("hello world")
		dst := make([]byte, len(src))

		err := Memcpy(dst, src)
		if err != nil {
			t.Fatalf("Memcpy() unexpected error: %v", err)
		}

		if string(dst) != string(src) {
			t.Errorf("Memcpy() dst = %q, want %q", dst, src)
		}
	})
}

func TestMemcpy_LargeBuffer(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMemcpy_LargeBuffer", nil, func(t *testing.T, tx *gorm.DB) {
		size := 64 * 1024 // 64KB
		src := make([]byte, size)
		for i := range src {
			src[i] = byte(i % 256)
		}
		dst := make([]byte, size)

		err := Memcpy(dst, src)
		if err != nil {
			t.Fatalf("Memcpy() unexpected error: %v", err)
		}

		// Verify all bytes were copied correctly
		for i := 0; i < size; i++ {
			if dst[i] != src[i] {
				t.Errorf("Mismatch at index %d: got %d, want %d", i, dst[i], src[i])
				break
			}
		}
	})
}

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

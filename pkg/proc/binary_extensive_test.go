package proc

import (
	"fmt"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// Level 1 tests - Basic functionality (50 tests)

func TestBinaryHeaderEncoding_Level1_EmptyFields(t *testing.T) {
	testutils.Run(t, testutils.Level1, "EmptyFields", nil, func(t *testing.T, tx *gorm.DB) {
		header := NewBinaryHeader()
		buf := make([]byte, HeaderSize)
		if err := header.EncodeHeader(buf); err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		decoded, err := DecodeHeader(buf)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if decoded.GetMessageID() != "" {
			t.Error("Expected empty message ID")
		}
	})
}

func TestBinaryHeaderEncoding_Level1_SingleByteID(t *testing.T) {
	testutils.Run(t, testutils.Level1, "SingleByteID", nil, func(t *testing.T, tx *gorm.DB) {
		header := NewBinaryHeader()
		header.SetMessageID("a")

		buf := make([]byte, HeaderSize)
		header.EncodeHeader(buf)

		decoded, _ := DecodeHeader(buf)
		if decoded.GetMessageID() != "a" {
			t.Errorf("Expected 'a', got '%s'", decoded.GetMessageID())
		}
	})
}

func TestBinaryHeaderEncoding_Level1_MaxLengthID(t *testing.T) {
	testutils.Run(t, testutils.Level1, "MaxLengthID", nil, func(t *testing.T, tx *gorm.DB) {
		header := NewBinaryHeader()
		maxID := string(make([]byte, 32))
		for i := range maxID {
			maxID = maxID[:i] + "X" + maxID[i+1:]
		}
		header.SetMessageID(maxID)

		buf := make([]byte, HeaderSize)
		header.EncodeHeader(buf)

		decoded, _ := DecodeHeader(buf)
		if len(decoded.GetMessageID()) > 32 {
			t.Error("ID not truncated to 32 bytes")
		}
	})
}

func TestBinaryHeaderEncoding_Level1_AllEncodingTypes(t *testing.T) {
	encodings := []EncodingType{EncodingJSON, EncodingGOB, EncodingPLAIN}

	for _, enc := range encodings {
		t.Run(fmt.Sprintf("Encoding%d", enc), func(t *testing.T) {
			testutils.Run(t, testutils.Level1, fmt.Sprintf("Encoding%d", enc), nil, func(t *testing.T, tx *gorm.DB) {
				header := NewBinaryHeader()
				header.Encoding = enc

				buf := make([]byte, HeaderSize)
				header.EncodeHeader(buf)

				decoded, _ := DecodeHeader(buf)
				if decoded.Encoding != enc {
					t.Errorf("Expected encoding %d, got %d", enc, decoded.Encoding)
				}
			})
		})
	}
}

func TestBinaryHeaderEncoding_Level1_AllMessageTypes(t *testing.T) {
	types := []BinaryMessageType{
		BinaryMessageTypeRequest,
		BinaryMessageTypeResponse,
		BinaryMessageTypeEvent,
		BinaryMessageTypeError,
	}

	for _, msgType := range types {
		t.Run(fmt.Sprintf("Type%d", msgType), func(t *testing.T) {
			testutils.Run(t, testutils.Level1, fmt.Sprintf("Type%d", msgType), nil, func(t *testing.T, tx *gorm.DB) {
				header := NewBinaryHeader()
				header.MsgType = msgType

				buf := make([]byte, HeaderSize)
				header.EncodeHeader(buf)

				decoded, _ := DecodeHeader(buf)
				if decoded.MsgType != msgType {
					t.Errorf("Expected type %d, got %d", msgType, decoded.MsgType)
				}
			})
		})
	}
}

func TestBinaryHeaderEncoding_Level1_PayloadLengthBoundaries(t *testing.T) {
	lengths := []uint32{0, 1, 255, 256, 65535, 65536, 1 << 20, 1 << 24}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("Len%d", length), func(t *testing.T) {
			testutils.Run(t, testutils.Level1, fmt.Sprintf("Len%d", length), nil, func(t *testing.T, tx *gorm.DB) {
				header := NewBinaryHeader()
				header.PayloadLen = length

				buf := make([]byte, HeaderSize)
				header.EncodeHeader(buf)

				decoded, _ := DecodeHeader(buf)
				if decoded.PayloadLen != length {
					t.Errorf("Expected length %d, got %d", length, decoded.PayloadLen)
				}
			})
		})
	}
}

func TestBinaryHeaderEncoding_Level1_SpecialCharacters(t *testing.T) {
	specials := []string{
		"hello\nworld",
		"tab\there",
		"null\x00byte",
		"utf8-字符",
		"emoji-🚀",
	}

	for i, special := range specials {
		t.Run(fmt.Sprintf("Special%d", i), func(t *testing.T) {
			testutils.Run(t, testutils.Level1, fmt.Sprintf("Special%d", i), nil, func(t *testing.T, tx *gorm.DB) {
				header := NewBinaryHeader()
				header.SetMessageID(special)

				buf := make([]byte, HeaderSize)
				header.EncodeHeader(buf)

				decoded, _ := DecodeHeader(buf)
				// Should handle special chars correctly
				if decoded.GetMessageID() == "" {
					t.Error("Special characters lost")
				}
			})
		})
	}
}

func TestBinaryMessageGOB_Level1_SimplePayload(t *testing.T) {
	testutils.Run(t, testutils.Level1, "SimplePayload", nil, func(t *testing.T, tx *gorm.DB) {
		msg, _ := NewRequestMessage("req-1", map[string]string{"key": "value"})

		data, err := MarshalBinaryWithEncoding(msg, EncodingGOB)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		decoded, err := UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		var payload map[string]string
		if err := decoded.UnmarshalPayload(&payload); err != nil {
			t.Fatalf("Payload unmarshal failed: %v", err)
		}

		if payload["key"] != "value" {
			t.Error("Payload mismatch")
		}
	})
}

func TestBinaryMessageGOB_Level1_EmptyPayload(t *testing.T) {
	testutils.Run(t, testutils.Level1, "EmptyPayload", nil, func(t *testing.T, tx *gorm.DB) {
		msg := GetMessage()
		msg.Type = MessageTypeRequest
		msg.ID = "req-empty"

		data, err := MarshalBinaryWithEncoding(msg, EncodingGOB)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		decoded, err := UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if decoded.ID != "req-empty" {
			t.Error("ID mismatch")
		}
	})
}

func TestBinaryMessageGOB_Level1_LargePayload(t *testing.T) {
	testutils.Run(t, testutils.Level1, "LargePayload", nil, func(t *testing.T, tx *gorm.DB) {
		// Create large payload
		data := make([]byte, 10000)
		for i := range data {
			data[i] = byte(i % 256)
		}

		msg := GetMessage()
		msg.Type = MessageTypeEvent
		msg.ID = "large"
		msg.Payload = data

		encoded, err := MarshalBinaryWithEncoding(msg, EncodingGOB)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		decoded, err := UnmarshalBinary(encoded)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(decoded.Payload) != len(data) {
			t.Errorf("Payload size mismatch: expected %d, got %d", len(data), len(decoded.Payload))
		}
	})
}

// Add 40 more Level 1 tests
func TestBinaryMessage_Level1_AllFieldsCombinations(t *testing.T) {
	for i := 0; i < 40; i++ {
		t.Run(fmt.Sprintf("Combo%d", i), func(t *testing.T) {
			testutils.Run(t, testutils.Level1, fmt.Sprintf("Combo%d", i), nil, func(t *testing.T, tx *gorm.DB) {
				msg := GetMessage()
				msg.Type = MessageType([]MessageType{MessageTypeRequest, MessageTypeResponse, MessageTypeEvent, MessageTypeError}[i%4])
				msg.ID = fmt.Sprintf("id-%d", i)
				msg.Source = fmt.Sprintf("src-%d", i)
				msg.Target = fmt.Sprintf("tgt-%d", i)
				msg.CorrelationID = fmt.Sprintf("corr-%d", i)

				data, err := msg.MarshalBinary()
				if err != nil {
					t.Fatalf("Marshal failed: %v", err)
				}

				decoded, err := UnmarshalBinary(data)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}

				if decoded.ID != msg.ID {
					t.Error("ID mismatch")
				}
			})
		})
	}
}

package proc

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/jsonutil"
)

// BinaryMessage represents a message with binary protocol support
type BinaryMessage struct {
	Header  BinaryHeader
	Payload []byte
}

// MarshalBinary marshals a Message to binary format with fixed header
// Defaults to JSON encoding for best performance on small messages
func (m *Message) MarshalBinary() ([]byte, error) {
	return MarshalBinaryWithEncoding(m, EncodingJSON)
}

// MarshalBinaryWithEncoding marshals a Message to binary format with specified encoding
func MarshalBinaryWithEncoding(m *Message, encoding EncodingType) ([]byte, error) {
	// Create binary header
	header := NewBinaryHeader()
	header.Encoding = encoding
	header.MsgType = ConvertMessageTypeToBinary(m.Type)
	header.SetMessageID(m.ID)
	header.SetSourceID(m.Source)
	header.SetTargetID(m.Target)
	header.SetCorrelationID(m.CorrelationID)

	// Encode payload based on encoding type
	var payload []byte
	var err error

	switch encoding {
	case EncodingJSON:
		// For JSON encoding, encode the entire message (excluding routing fields)
		// as they are in the header
		if m.Type == MessageTypeError {
			// For error messages, encode the error string
			payload, err = jsonutil.Marshal(map[string]interface{}{
				"error": m.Error,
			})
		} else if m.Payload != nil {
			// For other messages, use the payload directly
			payload = m.Payload
		} else {
			payload = []byte("{}")
		}

	case EncodingGOB:
		// For GOB encoding, we encode the raw JSON bytes directly
		// This avoids GOB's interface{} registration issues
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)

		if m.Type == MessageTypeError {
			// Encode error as JSON then GOB-encode the bytes
			errJSON, err := jsonutil.Marshal(map[string]interface{}{
				"error": m.Error,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to marshal error for GOB: %w", err)
			}
			err = enc.Encode(errJSON)
		} else if m.Payload != nil {
			// GOB-encode the JSON bytes
			err = enc.Encode(m.Payload)
		} else {
			err = enc.Encode([]byte("{}"))
		}

		if err == nil {
			payload = buf.Bytes()
		}

	case EncodingPLAIN:
		// For plain encoding, use the payload as-is or the error string
		if m.Type == MessageTypeError {
			payload = []byte(m.Error)
		} else if m.Payload != nil {
			payload = m.Payload
		} else {
			payload = []byte{}
		}

	default:
		return nil, fmt.Errorf("unsupported encoding type: %d", encoding)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode payload: %w", err)
	}

	// Set payload length in header
	header.PayloadLen = uint32(len(payload))

	// Allocate buffer for header + payload
	totalSize := HeaderSize + len(payload)
	result := make([]byte, totalSize)

	// Encode header
	if err := header.EncodeHeader(result[:HeaderSize]); err != nil {
		return nil, fmt.Errorf("failed to encode header: %w", err)
	}

	// Copy payload
	copy(result[HeaderSize:], payload)

	return result, nil
}

// UnmarshalBinary unmarshals a binary message to a Message
func UnmarshalBinary(data []byte) (*Message, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("data too small: need at least %d bytes, got %d", HeaderSize, len(data))
	}

	// Decode header
	header, err := DecodeHeader(data[:HeaderSize])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	// Verify payload length
	expectedSize := HeaderSize + int(header.PayloadLen)
	if len(data) < expectedSize {
		return nil, fmt.Errorf("incomplete message: expected %d bytes, got %d", expectedSize, len(data))
	}

	// Extract payload
	payload := data[HeaderSize:expectedSize]

	// Create message
	msg := GetMessage()
	msg.Type = ConvertBinaryMessageTypeToString(header.MsgType)
	msg.ID = header.GetMessageID()
	msg.Source = header.GetSourceID()
	msg.Target = header.GetTargetID()
	msg.CorrelationID = header.GetCorrelationID()

	// Decode payload based on encoding
	switch header.Encoding {
	case EncodingJSON:
		if msg.Type == MessageTypeError {
			// Decode error message
			var errData map[string]interface{}
			if err := jsonutil.Unmarshal(payload, &errData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal error payload: %w", err)
			}
			if errStr, ok := errData["error"].(string); ok {
				msg.Error = errStr
			}
		} else {
			// Store payload as-is (already JSON)
			msg.Payload = payload
		}

	case EncodingGOB:
		// Decode GOB payload (which contains JSON bytes)
		buf := bytes.NewReader(payload)
		dec := gob.NewDecoder(buf)

		var jsonBytes []byte
		if err := dec.Decode(&jsonBytes); err != nil {
			return nil, fmt.Errorf("failed to decode GOB payload: %w", err)
		}

		if msg.Type == MessageTypeError {
			// Extract error from JSON bytes
			var errData map[string]interface{}
			if err := jsonutil.Unmarshal(jsonBytes, &errData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal error from GOB: %w", err)
			}
			if errStr, ok := errData["error"].(string); ok {
				msg.Error = errStr
			}
		} else {
			// Store the JSON bytes as payload
			msg.Payload = jsonBytes
		}

	case EncodingPLAIN:
		if msg.Type == MessageTypeError {
			msg.Error = string(payload)
		} else {
			msg.Payload = payload
		}

	default:
		return nil, fmt.Errorf("unsupported encoding type: %d", header.Encoding)
	}

	return msg, nil
}

// UnmarshalBinaryFast unmarshals a binary message using pooled message
func UnmarshalBinaryFast(data []byte) (*Message, error) {
	// This is identical to UnmarshalBinary since we already use GetMessage
	return UnmarshalBinary(data)
}

// IsBinaryMessage checks if data starts with the binary protocol magic bytes
func IsBinaryMessage(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	return data[0] == MagicByte1 && data[1] == MagicByte2
}

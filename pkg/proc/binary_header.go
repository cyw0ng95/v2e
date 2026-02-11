package proc

import (
	"encoding/binary"
	"fmt"
)

// Binary Protocol Constants
const (
	// HeaderSize is the fixed size of the binary message header (128 bytes)
	HeaderSize = 128

	// MagicBytes is the protocol identifier
	MagicByte1 byte = 0x56 // 'V'
	MagicByte2 byte = 0x32 // '2'

	// ProtocolVersion is the current protocol version
	ProtocolVersion byte = 0x01
)

// EncodingType represents the payload encoding
type EncodingType byte

const (
	// EncodingJSON uses JSON encoding for the payload
	EncodingJSON EncodingType = 0
	// EncodingGOB uses GOB encoding for the payload
	EncodingGOB EncodingType = 1
	// EncodingPLAIN uses plain text encoding for the payload
	EncodingPLAIN EncodingType = 2
)

// BinaryMessageType represents message type as a byte
type BinaryMessageType byte

const (
	// BinaryMessageTypeRequest represents a request message
	BinaryMessageTypeRequest BinaryMessageType = 0
	// BinaryMessageTypeResponse represents a response message
	BinaryMessageTypeResponse BinaryMessageType = 1
	// BinaryMessageTypeEvent represents an event message
	BinaryMessageTypeEvent BinaryMessageType = 2
	// BinaryMessageTypeError represents an error message
	BinaryMessageTypeError BinaryMessageType = 3
)

// BinaryHeader represents the fixed-size binary message header
// Layout (128 bytes total):
//   - Magic (2 bytes): Protocol identifier
//   - Version (1 byte): Protocol version
//   - Encoding (1 byte): Payload encoding type
//   - MsgType (1 byte): Message type
//   - Reserved1 (3 bytes): Reserved for future use
//   - PayloadLen (4 bytes): Length of the payload (uint32)
//   - MessageID (32 bytes): Message ID
//   - SourceID (32 bytes): Sender process ID
//   - TargetID (32 bytes): Recipient process ID
//   - CorrelationID (20 bytes): Correlation ID for request-response matching
type BinaryHeader struct {
	Magic         [2]byte           // 0-1: Protocol identifier (V2)
	Version       byte              // 2: Protocol version
	Encoding      EncodingType      // 3: Payload encoding
	MsgType       BinaryMessageType // 4: Message type
	Reserved1     [3]byte           // 5-7: Reserved for future use
	PayloadLen    uint32            // 8-11: Payload length
	MessageID     [32]byte          // 12-43: Message ID
	SourceID      [32]byte          // 44-75: Source process ID
	TargetID      [32]byte          // 76-107: Target process ID
	CorrelationID [20]byte          // 108-127: Correlation ID
}

// NewBinaryHeader creates a new binary header with default values
func NewBinaryHeader() *BinaryHeader {
	return &BinaryHeader{
		Magic:   [2]byte{MagicByte1, MagicByte2},
		Version: ProtocolVersion,
	}
}

// EncodeHeader encodes the binary header to a byte slice
func (h *BinaryHeader) EncodeHeader(buf []byte) error {
	if len(buf) < HeaderSize {
		return fmt.Errorf("buffer too small: need %d bytes, got %d", HeaderSize, len(buf))
	}

	// Hint to kernel that we'll access this buffer sequentially
	_ = MadviseSequential(buf)

	// Magic (2 bytes)
	buf[0] = h.Magic[0]
	buf[1] = h.Magic[1]

	// Version (1 byte)
	buf[2] = h.Version

	// Encoding (1 byte)
	buf[3] = byte(h.Encoding)

	// MsgType (1 byte)
	buf[4] = byte(h.MsgType)

	// Reserved (3 bytes)
	buf[5] = h.Reserved1[0]
	buf[6] = h.Reserved1[1]
	buf[7] = h.Reserved1[2]

	// PayloadLen (4 bytes, big-endian)
	binary.BigEndian.PutUint32(buf[8:12], h.PayloadLen)

	// MessageID (32 bytes) - use optimized copy
	srcMsgID := h.MessageID[:]
	dstMsgID := buf[12:44]
	if err := Memcpy(dstMsgID, srcMsgID); err != nil {
		return fmt.Errorf("failed to copy MessageID: %w", err)
	}

	// SourceID (32 bytes) - use optimized copy
	srcSourceID := h.SourceID[:]
	dstSourceID := buf[44:76]
	if err := Memcpy(dstSourceID, srcSourceID); err != nil {
		return fmt.Errorf("failed to copy SourceID: %w", err)
	}

	// TargetID (32 bytes) - use optimized copy
	srcTargetID := h.TargetID[:]
	dstTargetID := buf[76:108]
	if err := Memcpy(dstTargetID, srcTargetID); err != nil {
		return fmt.Errorf("failed to copy TargetID: %w", err)
	}

	// CorrelationID (20 bytes) - use optimized copy
	srcCorrID := h.CorrelationID[:]
	dstCorrID := buf[108:128]
	if err := Memcpy(dstCorrID, srcCorrID); err != nil {
		return fmt.Errorf("failed to copy CorrelationID: %w", err)
	}

	return nil
}

// DecodeHeader decodes a binary header from a byte slice
func DecodeHeader(buf []byte) (*BinaryHeader, error) {
	if len(buf) < HeaderSize {
		return nil, fmt.Errorf("buffer too small: need %d bytes, got %d", HeaderSize, len(buf))
	}

	h := &BinaryHeader{}

	// Magic (2 bytes)
	h.Magic[0] = buf[0]
	h.Magic[1] = buf[1]

	// Verify magic bytes
	if h.Magic[0] != MagicByte1 || h.Magic[1] != MagicByte2 {
		return nil, fmt.Errorf("invalid magic bytes: expected [%02x %02x], got [%02x %02x]",
			MagicByte1, MagicByte2, h.Magic[0], h.Magic[1])
	}

	// Version (1 byte)
	h.Version = buf[2]

	// Encoding (1 byte)
	h.Encoding = EncodingType(buf[3])

	// MsgType (1 byte)
	h.MsgType = BinaryMessageType(buf[4])

	// Reserved (3 bytes)
	h.Reserved1[0] = buf[5]
	h.Reserved1[1] = buf[6]
	h.Reserved1[2] = buf[7]

	// PayloadLen (4 bytes, big-endian)
	h.PayloadLen = binary.BigEndian.Uint32(buf[8:12])

	// MessageID (32 bytes)
	copy(h.MessageID[:], buf[12:44])

	// SourceID (32 bytes)
	copy(h.SourceID[:], buf[44:76])

	// TargetID (32 bytes)
	copy(h.TargetID[:], buf[76:108])

	// CorrelationID (20 bytes)
	copy(h.CorrelationID[:], buf[108:128])

	return h, nil
}

// GetTotalSize returns the total message size (header + payload)
func (h *BinaryHeader) GetTotalSize() int {
	return HeaderSize + int(h.PayloadLen)
}

// stringToFixedBytes converts a string to a fixed-size byte array
func stringToFixedBytes(s string, size int) []byte {
	result := make([]byte, size)
	copy(result, []byte(s))
	return result
}

// fixedBytesToString converts a fixed-size byte array to a string.
// It handles the edge case where all bytes are zero (uninitialized or empty string)
// by consistently returning an empty string, matching the behavior of stringToFixedBytes("").
func fixedBytesToString(b []byte) string {
	// Handle empty slice edge case
	if len(b) == 0 {
		return ""
	}

	// Find the first null byte and return the string up to that point.
	// For all-zero byte arrays (uninitialized fields), this returns "" at i=0,
	// which is the correct and consistent behavior.
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// SetMessageID sets the message ID in the header
func (h *BinaryHeader) SetMessageID(id string) {
	copy(h.MessageID[:], stringToFixedBytes(id, 32))
}

// GetMessageID gets the message ID from the header
func (h *BinaryHeader) GetMessageID() string {
	return fixedBytesToString(h.MessageID[:])
}

// SetSourceID sets the source process ID in the header
func (h *BinaryHeader) SetSourceID(id string) {
	copy(h.SourceID[:], stringToFixedBytes(id, 32))
}

// GetSourceID gets the source process ID from the header
func (h *BinaryHeader) GetSourceID() string {
	return fixedBytesToString(h.SourceID[:])
}

// SetTargetID sets the target process ID in the header
func (h *BinaryHeader) SetTargetID(id string) {
	copy(h.TargetID[:], stringToFixedBytes(id, 32))
}

// GetTargetID gets the target process ID from the header
func (h *BinaryHeader) GetTargetID() string {
	return fixedBytesToString(h.TargetID[:])
}

// SetCorrelationID sets the correlation ID in the header
func (h *BinaryHeader) SetCorrelationID(id string) {
	copy(h.CorrelationID[:], stringToFixedBytes(id, 20))
}

// GetCorrelationID gets the correlation ID from the header
func (h *BinaryHeader) GetCorrelationID() string {
	return fixedBytesToString(h.CorrelationID[:])
}

// ConvertMessageTypeToBinary converts MessageType to BinaryMessageType
func ConvertMessageTypeToBinary(t MessageType) BinaryMessageType {
	switch t {
	case MessageTypeRequest:
		return BinaryMessageTypeRequest
	case MessageTypeResponse:
		return BinaryMessageTypeResponse
	case MessageTypeEvent:
		return BinaryMessageTypeEvent
	case MessageTypeError:
		return BinaryMessageTypeError
	default:
		return BinaryMessageTypeRequest
	}
}

// ConvertBinaryMessageTypeToString converts BinaryMessageType to MessageType
func ConvertBinaryMessageTypeToString(t BinaryMessageType) MessageType {
	switch t {
	case BinaryMessageTypeRequest:
		return MessageTypeRequest
	case BinaryMessageTypeResponse:
		return MessageTypeResponse
	case BinaryMessageTypeEvent:
		return MessageTypeEvent
	case BinaryMessageTypeError:
		return MessageTypeError
	default:
		return MessageTypeRequest
	}
}

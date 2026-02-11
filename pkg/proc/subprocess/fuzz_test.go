package subprocess

import (
	"testing"

	"github.com/bytedance/sonic"
)

// FuzzMessageUnmarshal tests unmarshaling of arbitrary byte sequences
func FuzzMessageUnmarshal(f *testing.F) {
	// Seed corpus with valid messages
	f.Add([]byte(`{"type":"request","id":"test"}`))
	f.Add([]byte(`{"type":"response","id":"resp-1","payload":{"status":"ok"}}`))
	f.Add([]byte(`{"type":"event","id":"evt-1"}`))
	f.Add([]byte(`{"type":"error","id":"err-1","error":"test error"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"type":"request","id":"test","target":"broker","correlation_id":"abc123"}`))

	// Fuzz test
	f.Fuzz(func(t *testing.T, data []byte) {
		// Attempt to unmarshal - should not panic
		var msg Message
		_ = sonic.Unmarshal(data, &msg)

		// If unmarshal succeeded, marshal it back - should not panic
		if msg.Type != "" {
			_, _ = sonic.Marshal(&msg)
		}
	})
}

// FuzzMessageMarshal tests marshaling of messages with arbitrary field values
func FuzzMessageMarshal(f *testing.F) {
	// Seed corpus
	f.Add("request", "test-id", "", "", "")
	f.Add("response", "resp-1", "target-1", "source-1", "corr-1")
	f.Add("event", "", "", "", "")

	// Fuzz test
	f.Fuzz(func(t *testing.T, msgType, id, target, source, corrID string) {
		// Create message with fuzzed fields
		msg := &Message{
			Type:          MessageType(msgType),
			ID:            id,
			Target:        target,
			Source:        source,
			CorrelationID: corrID,
		}

		// Marshal - should not panic
		data, err := sonic.Marshal(msg)
		if err != nil {
			return
		}

		// Unmarshal back - should not panic
		var msg2 Message
		_ = sonic.Unmarshal(data, &msg2)
	})
}

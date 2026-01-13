package proc

import (
	"testing"
)

// FuzzMessageMarshal tests marshaling messages with arbitrary content
func FuzzMessageMarshal(f *testing.F) {
	// Seed corpus
	f.Add("request", "test-id", "broker", "access", "corr-123", "")
	f.Add("response", "resp-1", "access", "broker", "corr-456", "")
	f.Add("error", "err-1", "", "", "", "error message")
	f.Add("event", "", "", "", "", "")
	
	// Fuzz test
	f.Fuzz(func(t *testing.T, msgType, id, target, source, corrID, errMsg string) {
		// Create message with fuzzed fields
		msg := &Message{
			Type:          MessageType(msgType),
			ID:            id,
			Target:        target,
			Source:        source,
			CorrelationID: corrID,
			Error:         errMsg,
		}
		
		// Marshal - should not panic
		data, err := msg.Marshal()
		if err != nil {
			return
		}
		
		// Unmarshal back - should not panic
		_, _ = Unmarshal(data)
	})
}

// FuzzMessageUnmarshal tests unmarshaling arbitrary byte sequences
func FuzzMessageUnmarshal(f *testing.F) {
	// Seed corpus with valid messages
	f.Add([]byte(`{"type":"request","id":"test"}`))
	f.Add([]byte(`{"type":"response","id":"resp-1","payload":{"status":"ok"}}`))
	f.Add([]byte(`{"type":"event","id":"evt-1"}`))
	f.Add([]byte(`{"type":"error","id":"err-1","error":"test error"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"type":"request","id":"test","target":"broker","correlation_id":"abc123"}`))
	f.Add([]byte(`{"invalid":"json","missing":"type"}`))
	f.Add([]byte(`null`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`"string"`))
	
	// Fuzz test
	f.Fuzz(func(t *testing.T, data []byte) {
		// Attempt to unmarshal - should not panic
		_, _ = Unmarshal(data)
	})
}

// FuzzNewRequestMessage tests creating request messages with arbitrary payloads
func FuzzNewRequestMessage(f *testing.F) {
	// Seed corpus
	f.Add("RPCGetCVE", `{"cve_id":"CVE-2021-44228"}`)
	f.Add("RPCListCVEs", `{"limit":10,"offset":0}`)
	f.Add("RPCInvalid", `invalid json`)
	f.Add("", `{}`)
	f.Add("RPCTest", `null`)
	
	// Fuzz test
	f.Fuzz(func(t *testing.T, method, payloadStr string) {
		// Create request message - should not panic
		_, _ = NewRequestMessage(method, []byte(payloadStr))
	})
}

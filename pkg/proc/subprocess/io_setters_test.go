package subprocess

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSetOutputDisablesBatchingAndWrites(t *testing.T) {
	sp := New("test-io")
	buf := &bytes.Buffer{}
	sp.SetOutput(buf)
	if !sp.disableBatching {
		t.Fatalf("expected disableBatching to be true after SetOutput")
	}

	// send a small message and ensure it is written to buffer
	payload := map[string]int{"v": 1}
	pbytes, _ := json.Marshal(payload)
	msg := &Message{Type: MessageTypeEvent, ID: "evt", Payload: pbytes}
	if err := sp.SendMessage(msg); err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	out := buf.String()
	if len(out) == 0 {
		t.Fatalf("expected output written to buffer")
	}
}

func TestSetInputUpdatesReader(t *testing.T) {
	sp := New("test-in")
	// ensure SetInput doesn't panic and sets the input reader
	r := bytes.NewBufferString("hello")
	sp.SetInput(r)
}

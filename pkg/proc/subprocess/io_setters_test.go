package subprocess

import (
	"bytes"
	"encoding/json"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestSetOutputDisablesBatchingAndWrites(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetOutputDisablesBatchingAndWrites", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

func TestSetInputUpdatesReader(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetInputUpdatesReader", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test-in")
		// ensure SetInput doesn't panic and sets the input reader
		r := bytes.NewBufferString("hello")
		sp.SetInput(r)
	})

}

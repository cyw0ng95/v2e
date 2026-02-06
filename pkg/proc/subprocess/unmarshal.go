package subprocess

import (
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/jsonutil"
)

// UnmarshalPayload is a helper to unmarshal message payload
func UnmarshalPayload(msg *Message, v interface{}) error {
	if msg.Payload == nil {
		return fmt.Errorf("no payload to unmarshal")
	}
	return jsonutil.Unmarshal(msg.Payload, v)
}

package subprocess

import (
	"fmt"

	"github.com/bytedance/sonic"
)

// UnmarshalPayload is a helper to unmarshal message payload
func UnmarshalPayload(msg *Message, v interface{}) error {
	if msg.Payload == nil {
		return fmt.Errorf("no payload to unmarshal")
	}
	return sonic.Unmarshal(msg.Payload, v)
}

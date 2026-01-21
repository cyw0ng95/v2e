package proc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRequestMessage_MarshalFailure(t *testing.T) {
	// payload containing a channel should fail to marshal
	payload := struct{ C chan int }{C: make(chan int)}
	msg, err := NewRequestMessage("test-id", payload)
	require.Error(t, err)
	require.Nil(t, msg)
}

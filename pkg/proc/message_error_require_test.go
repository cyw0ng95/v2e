package proc

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/stretchr/testify/require"
)

func TestNewRequestMessage_MarshalFailure(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestNewRequestMessage_MarshalFailure", nil, func(t *testing.T, tx *gorm.DB) {
		// payload containing a channel should fail to marshal
		payload := struct{ C chan int }{C: make(chan int)}
		msg, err := NewRequestMessage("test-id", payload)
		require.Error(t, err)
		require.Nil(t, msg)
	})

}

package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestInvokeRPCWithUnmarshalableParams(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestInvokeRPCWithUnmarshalableParams", nil, func(t *testing.T, tx *gorm.DB) {
		sp := subprocess.New("test-client")
		logger := common.NewLogger(os.Stderr, "[ACCESS] ", common.InfoLevel)
		client := NewRPCClientWithSubprocess(sp, logger, time.Second)

		// channels are not marshalable to JSON; expect marshal error
		_, err := client.InvokeRPCWithTarget(context.Background(), "broker", "m", make(chan int))
		if err == nil {
			t.Fatalf("expected marshal error for non-marshable params")
		}
		if !strings.Contains(err.Error(), "marshal") {
			t.Fatalf("expected marshal-related error, got: %v", err)
		}
	})

}

func TestInvokeRPCWithTimeout(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestInvokeRPCWithTimeout", nil, func(t *testing.T, tx *gorm.DB) {
		// Removed: this test relied on a very short (10ms) timeout which is
		// unreliable on CI. Timeouts and pending-request cleanup are covered by
		// other tests in the common RPC client package.
	})

}

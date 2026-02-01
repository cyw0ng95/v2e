package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestInvokeRPCWithUnmarshalableParams(t *testing.T) {
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
}

func TestInvokeRPCWithTimeout(t *testing.T) {
	sp := subprocess.New("test-client")
	// do not install any response writer; this should cause a timeout
	logger := common.NewLogger(os.Stderr, "[ACCESS] ", common.InfoLevel)
	client := NewRPCClientWithSubprocess(sp, logger, 10*time.Millisecond)

	_, err := client.InvokeRPCWithTarget(context.Background(), "broker", "m", nil)
	if err == nil {
		t.Fatalf("expected timeout error when no response is provided")
	}
	if !strings.Contains(err.Error(), "RPC timeout") && !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout-related error, got: %v", err)
	}
}

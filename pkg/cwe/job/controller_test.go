package job

import (
	"context"
	"testing"
)

func TestController_StartStopStatus(t *testing.T) {
	c := NewController()
	ctx := context.Background()
	sid, err := c.Start(ctx, map[string]interface{}{"test": true})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if sid == "" {
		t.Fatalf("expected session id, got empty")
	}
	st, err := c.Status(ctx, sid)
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if st == nil {
		t.Fatalf("expected non-nil status")
	}
	if err := c.Stop(ctx, sid); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
}

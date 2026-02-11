package job

import (
	"context"
	"io"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/bytedance/sonic"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// mockInvoker returns an empty views response so the job loop exits quickly.
type mockInvoker struct{}

func (m *mockInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	// return a subprocess.Message with an empty "views" array
	payload, _ := sonic.Marshal(map[string]interface{}{"views": []interface{}{}})
	return &subprocess.Message{
		Type:    subprocess.MessageTypeResponse,
		Payload: payload,
	}, nil
}

func TestController_StartStopStatus(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestController_StartStopStatus", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(io.Discard, "test", common.InfoLevel)
		mock := &mockInvoker{}
		c := NewController(mock, logger)

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
	})

}

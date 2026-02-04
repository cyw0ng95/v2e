package mq

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

func TestBusSendReceiveStats(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBusSendReceiveStats", nil, func(t *testing.T, tx *gorm.DB) {
		ctx := context.Background()
		bus := NewBus(ctx, 4)

		msg := &proc.Message{Type: proc.MessageTypeRequest, ID: "req1", Source: "source", Target: "target"}
		if err := bus.Send(ctx, msg); err != nil {
			t.Fatalf("Send returned error: %v", err)
		}

		got, err := bus.Receive(ctx)
		if err != nil {
			t.Fatalf("Receive returned error: %v", err)
		}
		if got != msg {
			t.Fatalf("Expected to receive original message pointer")
		}

		bus.Record(&proc.Message{Type: proc.MessageTypeResponse, Target: "t2"}, true)
		bus.Record(&proc.Message{Type: proc.MessageTypeEvent, Source: "s3"}, false)
		bus.Record(&proc.Message{Type: proc.MessageTypeError, Source: "s4"}, false)

		stats := bus.GetMessageStats()
		if stats.TotalSent != 2 || stats.TotalReceived != 3 {
			t.Fatalf("Totals mismatch: sent=%d received=%d", stats.TotalSent, stats.TotalReceived)
		}
		if stats.RequestCount != 2 || stats.ResponseCount != 1 || stats.EventCount != 1 || stats.ErrorCount != 1 {
			t.Fatalf("Type counts mismatch: %+v", stats)
		}

		per := bus.GetPerProcessStats()
		if per["target"].TotalSent != 1 || per["target"].RequestCount != 1 {
			t.Fatalf("Per-process stats for target missing send counts: %+v", per["target"])
		}
		if per["source"].TotalReceived != 1 || per["source"].RequestCount != 1 {
			t.Fatalf("Per-process stats for source missing receive counts: %+v", per["source"])
		}
		if per["t2"].ResponseCount != 1 || per["t2"].TotalSent != 1 {
			t.Fatalf("Per-process stats for t2 missing response counts: %+v", per["t2"])
		}
		if per["s3"].EventCount != 1 || per["s3"].TotalReceived != 1 {
			t.Fatalf("Per-process stats for s3 missing event counts: %+v", per["s3"])
		}
		if per["s4"].ErrorCount != 1 || per["s4"].TotalReceived != 1 {
			t.Fatalf("Per-process stats for s4 missing error counts: %+v", per["s4"])
		}

		if bus.GetMessageCount() != stats.TotalSent+stats.TotalReceived {
			t.Fatalf("Message count mismatch")
		}
	})

}

func TestBusSendCanceled(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBusSendCanceled", nil, func(t *testing.T, tx *gorm.DB) {
		bus := NewBus(context.Background(), 0)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := bus.Send(ctx, &proc.Message{})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Expected context.Canceled, got %v", err)
		}
	})

}

func TestBusSendOnClosedChannel(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBusSendOnClosedChannel", nil, func(t *testing.T, tx *gorm.DB) {
		bus := NewBus(context.Background(), 1)
		bus.Close()

		if err := bus.Send(context.Background(), &proc.Message{}); err == nil || err.Error() != "message channel is closed" {
			t.Fatalf("Expected closed channel error, got %v", err)
		}

		if _, err := bus.Receive(context.Background()); !errors.Is(err, context.Canceled) {
			t.Fatalf("Expected context.Canceled on closed channel receive, got %v", err)
		}
	})

}

func TestBusReceiveCanceled(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBusReceiveCanceled", nil, func(t *testing.T, tx *gorm.DB) {
		bus := NewBus(context.Background(), 1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := bus.Receive(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("Expected context.Canceled receive, got %v", err)
		}
	})

}

func TestSendRespectsBusContext(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSendRespectsBusContext", nil, func(t *testing.T, tx *gorm.DB) {
		busCtx, cancelBus := context.WithCancel(context.Background())
		bus := NewBus(busCtx, 0)
		cancelBus()

		err := bus.Send(context.Background(), &proc.Message{})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Expected bus context canceled error, got %v", err)
		}
	})

}

func TestFirstAndLastMessageTimesSet(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFirstAndLastMessageTimesSet", nil, func(t *testing.T, tx *gorm.DB) {
		bus := NewBus(context.Background(), 1)
		msg := &proc.Message{Type: proc.MessageTypeEvent, Source: "s", Target: "t"}

		if err := bus.Send(context.Background(), msg); err != nil {
			t.Fatalf("Send returned error: %v", err)
		}
		if _, err := bus.Receive(context.Background()); err != nil {
			t.Fatalf("Receive returned error: %v", err)
		}

		stats := bus.GetMessageStats()
		if stats.FirstMessageTime.IsZero() || stats.LastMessageTime.IsZero() {
			t.Fatalf("Expected timestamps to be set: %+v", stats)
		}
		if stats.LastMessageTime.Before(stats.FirstMessageTime.Add(-time.Millisecond)) {
			t.Fatalf("LastMessageTime should be >= FirstMessageTime")
		}
	})

}

package perf

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// simpleRouter records routed messages for assertions
type simpleRouter struct {
	mu   sync.Mutex
	msgs []*proc.Message
}

func (r *simpleRouter) Route(msg *proc.Message, sourceProcess string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs = append(r.msgs, msg)
	return nil
}

func (r *simpleRouter) ProcessBrokerMessage(msg *proc.Message) error {
	return r.Route(msg, "broker")
}

func TestOfferDropOldest(t *testing.T) {
	router := &simpleRouter{}
	opt := NewWithParams(router, 2, 1, 100*time.Millisecond, "drop_oldest", 0, 1, 10*time.Millisecond)
	defer opt.Stop()

	m1 := &proc.Message{ID: "m1"}
	m2 := &proc.Message{ID: "m2"}
	m3 := &proc.Message{ID: "m3"}

	if !opt.Offer(m1) {
		t.Fatal("expected m1 accepted")
	}
	if !opt.Offer(m2) {
		t.Fatal("expected m2 accepted")
	}
	// buffer full; drop_oldest should remove m1 and accept m3
	if !opt.Offer(m3) {
		t.Fatal("expected m3 accepted under drop_oldest policy")
	}

	// Wait for worker to process messages using a deadline instead of fixed sleep
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		router.mu.Lock()
		count := len(router.msgs)
		router.mu.Unlock()
		if count >= 2 {
			break
		}
		time.Sleep(5 * time.Millisecond) // Small poll interval
	}

	router.mu.Lock()
	ids := make(map[string]bool)
	for _, m := range router.msgs {
		ids[m.ID] = true
	}
	router.mu.Unlock()

	if ids["m1"] {
		t.Fatal("m1 should have been dropped by drop_oldest policy")
	}
	if !ids["m2"] || !ids["m3"] {
		t.Fatalf("expected m2 and m3 to be processed, got: %v", ids)
	}
}

func TestBatching(t *testing.T) {
	router := &simpleRouter{}
	// batch size 3, small flush interval
	opt := NewWithParams(router, 10, 1, 100*time.Millisecond, "drop", 0, 3, 20*time.Millisecond)
	defer opt.Stop()

	for i := 0; i < 5; i++ {
		m := &proc.Message{ID: fmt.Sprintf("%c", 'a'+i)}
		ok := opt.Offer(m)
		t.Logf("offered %s ok=%v", m.ID, ok)
		if !ok {
			t.Fatalf("offer failed for %v", m.ID)
		}
	}

	// Poll for all messages to be processed instead of fixed sleep
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		router.mu.Lock()
		count := len(router.msgs)
		router.mu.Unlock()
		if count >= 5 {
			break
		}
		time.Sleep(10 * time.Millisecond) // Small poll interval
	}

	router.mu.Lock()
	count := len(router.msgs)
	ids := make([]string, 0, len(router.msgs))
	for _, m := range router.msgs {
		ids = append(ids, m.ID)
	}
	router.mu.Unlock()

	t.Logf("processed ids: %v", ids)
	t.Logf("queue len=%d dropped=%d", len(opt.optimizedMessages), atomic.LoadInt64(&opt.droppedMessages))

	if count != 5 {
		t.Fatalf("expected 5 messages processed, got %d", count)
	}
}

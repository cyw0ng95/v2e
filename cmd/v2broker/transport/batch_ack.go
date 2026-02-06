package transport

import (
	"sync"
	"sync/atomic"
	"time"
)

type AckType uint8

const (
	AckImmediate AckType = iota
	AckBatch
	AckDeferred
)

type AckMessage struct {
	SeqNum  uint64
	Success bool
	Error   error
}

type BatchAckConfig struct {
	MaxBatchSize  int
	FlushInterval time.Duration
	AckType       AckType
}

type BatchAck struct {
	mu           sync.Mutex
	config       BatchAckConfig
	buffer       []AckMessage
	pendingCount atomic.Int32
	lastFlush    time.Time
	flushTimer   *time.Timer
	flushChan    chan struct{}
	closed       bool
	onFlush      func([]AckMessage)
	onAckError   func(AckMessage, error)
}

func NewBatchAck(config BatchAckConfig) *BatchAck {
	if config.MaxBatchSize <= 0 {
		config.MaxBatchSize = 32
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 5 * time.Millisecond
	}

	ba := &BatchAck{
		config:    config,
		buffer:    make([]AckMessage, 0, config.MaxBatchSize),
		flushChan: make(chan struct{}, 1),
	}

	if config.AckType == AckBatch {
		ba.startFlushTimer()
	}

	return ba
}

func (ba *BatchAck) AddAck(msg AckMessage) {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	if ba.closed {
		ba.handleError(msg, nil)
		return
	}

	if ba.config.AckType == AckImmediate {
		if ba.onFlush != nil {
			ba.onFlush([]AckMessage{msg})
		}
		return
	}

	ba.buffer = append(ba.buffer, msg)
	ba.pendingCount.Add(1)

	if len(ba.buffer) >= ba.config.MaxBatchSize {
		ba.flush()
	}
}

func (ba *BatchAck) AddAckBatch(msgs []AckMessage) {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	if ba.closed {
		for _, msg := range msgs {
			ba.handleError(msg, nil)
		}
		return
	}

	if ba.config.AckType == AckImmediate {
		if ba.onFlush != nil {
			ba.onFlush(msgs)
		}
		return
	}

	for _, msg := range msgs {
		ba.buffer = append(ba.buffer, msg)
		ba.pendingCount.Add(1)
	}

	if len(ba.buffer) >= ba.config.MaxBatchSize {
		ba.flush()
	}
}

func (ba *BatchAck) Flush() error {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	if ba.closed {
		return nil
	}

	ba.flush()
	return nil
}

func (ba *BatchAck) flush() {
	if len(ba.buffer) == 0 {
		return
	}

	if ba.onFlush != nil {
		ba.onFlush(ba.buffer)
	}

	ba.pendingCount.Add(-int32(len(ba.buffer)))
	ba.buffer = ba.buffer[:0]
	ba.lastFlush = time.Now()
}

func (ba *BatchAck) startFlushTimer() {
	ba.flushTimer = time.AfterFunc(ba.config.FlushInterval, func() {
		if err := ba.Flush(); err != nil {
			return
		}
		ba.startFlushTimer()
	})
}

func (ba *BatchAck) SetOnFlush(fn func([]AckMessage)) {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	ba.onFlush = fn
}

func (ba *BatchAck) SetOnAckError(fn func(AckMessage, error)) {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	ba.onAckError = fn
}

func (ba *BatchAck) PendingCount() int32 {
	return ba.pendingCount.Load()
}

func (ba *BatchAck) LastFlush() time.Time {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	return ba.lastFlush
}

func (ba *BatchAck) handleError(msg AckMessage, err error) {
	if ba.onAckError != nil {
		ba.onAckError(msg, err)
	}
}

func (ba *BatchAck) Close() error {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	if ba.closed {
		return nil
	}

	ba.closed = true

	if ba.flushTimer != nil {
		ba.flushTimer.Stop()
	}

	ba.flush()

	close(ba.flushChan)
	return nil
}

type AckManager struct {
	batches map[uint64]*BatchAck
	mu      sync.RWMutex
	config  BatchAckConfig
}

func NewAckManager(config BatchAckConfig) *AckManager {
	return &AckManager{
		batches: make(map[uint64]*BatchAck),
		config:  config,
	}
}

func (am *AckManager) GetBatch(connID uint64) *BatchAck {
	am.mu.RLock()
	batch, exists := am.batches[connID]
	am.mu.RUnlock()

	if exists {
		return batch
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	if batch, exists := am.batches[connID]; exists {
		return batch
	}

	batch = NewBatchAck(am.config)
	am.batches[connID] = batch
	return batch
}

func (am *AckManager) RemoveBatch(connID uint64) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	batch, exists := am.batches[connID]
	if !exists {
		return nil
	}

	if err := batch.Close(); err != nil {
		return err
	}

	delete(am.batches, connID)
	return nil
}

func (am *AckManager) FlushAll() error {
	am.mu.RLock()
	batches := make([]*BatchAck, 0, len(am.batches))
	for _, batch := range am.batches {
		batches = append(batches, batch)
	}
	am.mu.RUnlock()

	for _, batch := range batches {
		if err := batch.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func (am *AckManager) Close() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for connID, batch := range am.batches {
		_ = batch.Close()
		delete(am.batches, connID)
	}

	return nil
}

func (am *AckManager) PendingTotal() int32 {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var total int32
	for _, batch := range am.batches {
		total += batch.PendingCount()
	}

	return total
}

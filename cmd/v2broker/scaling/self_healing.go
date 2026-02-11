package scaling

import (
	"fmt"
	"sync"
	"time"
)

// SelfHealingManager manages automatic fault recovery
type SelfHealingManager struct {
	mu                sync.Mutex
	faults            []FaultRecord
	healingStrategies map[string]HealingStrategy
	maxFaultHistory   int
	enabled           bool
	alertCallback     func(HealingAlert)
}

// FaultRecord represents a recorded fault
type FaultRecord struct {
	ID          string
	Timestamp   time.Time
	Type        string
	Severity    string
	Description string
	Resolved    bool
	ResolvedAt  time.Time
	Attempts    int
}

// HealingStrategy defines how to handle a specific fault type
type HealingStrategy interface {
	CanHandle(fault FaultRecord) bool
	Heal(fault FaultRecord) error
	GetRecoveryTime() time.Duration
}

// HealingAlert represents a healing event
type HealingAlert struct {
	Timestamp time.Time
	FaultID   string
	Action    string
	Success   bool
	Message   string
	Duration  time.Duration
}

// AutoRestartStrategy automatically restarts a service
type AutoRestartStrategy struct {
	attempts     int
	maxAttempts  int
	restartDelay time.Duration
	restartFunc  func() error
}

// RetryWithBackoffStrategy retries an operation with exponential backoff
type RetryWithBackoffStrategy struct {
	maxRetries    int
	initialDelay  time.Duration
	maxDelay      time.Duration
	backoffFactor float64
	retryFunc     func(attempt int) error
}

// CircuitBreakerStrategy implements circuit breaker pattern
type CircuitBreakerStrategy struct {
	failureThreshold int
	recoveryTimeout  time.Duration
	failures         int
	lastFailureTime  time.Time
	state            string // "closed", "open", "half-open"
	executeFunc      func() error
}

const (
	CircuitClosed   = "closed"
	CircuitOpen     = "open"
	CircuitHalfOpen = "half-open"
)

// NewSelfHealingManager creates a new self-healing manager
func NewSelfHealingManager(maxFaultHistory int) *SelfHealingManager {
	return &SelfHealingManager{
		faults:            make([]FaultRecord, 0),
		healingStrategies: make(map[string]HealingStrategy),
		maxFaultHistory:   maxFaultHistory,
		enabled:           true,
	}
}

// RegisterStrategy registers a healing strategy for a fault type
func (shm *SelfHealingManager) RegisterStrategy(faultType string, strategy HealingStrategy) {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	shm.healingStrategies[faultType] = strategy
}

// SetAlertCallback sets the alert callback
func (shm *SelfHealingManager) SetAlertCallback(callback func(HealingAlert)) {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	shm.alertCallback = callback
}

// ReportFault reports a fault for healing
func (shm *SelfHealingManager) ReportFault(faultType, severity, description string) error {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	if !shm.enabled {
		return fmt.Errorf("self-healing is disabled")
	}

	fault := FaultRecord{
		ID:          generateFaultID(),
		Timestamp:   time.Now(),
		Type:        faultType,
		Severity:    severity,
		Description: description,
		Resolved:    false,
		Attempts:    0,
	}

	shm.faults = append(shm.faults, fault)

	if len(shm.faults) > shm.maxFaultHistory {
		shm.faults = shm.faults[1:]
	}

	go shm.attemptHealing(fault)

	return nil
}

// attemptHealing attempts to heal a fault using registered strategies
func (shm *SelfHealingManager) attemptHealing(fault FaultRecord) {
	shm.mu.Lock()
	strategy, exists := shm.healingStrategies[fault.Type]
	shm.mu.Unlock()

	if !exists {
		shm.sendAlert(HealingAlert{
			Timestamp: time.Now(),
			FaultID:   fault.ID,
			Action:    "no_strategy",
			Success:   false,
			Message:   fmt.Sprintf("No healing strategy for fault type: %s", fault.Type),
			Duration:  0,
		})
		return
	}

	if !strategy.CanHandle(fault) {
		shm.sendAlert(HealingAlert{
			Timestamp: time.Now(),
			FaultID:   fault.ID,
			Action:    "cannot_handle",
			Success:   false,
			Message:   fmt.Sprintf("Strategy cannot handle fault: %s", fault.ID),
			Duration:  0,
		})
		return
	}

	startTime := time.Now()
	err := strategy.Heal(fault)
	duration := time.Since(startTime)

	if err == nil {
		shm.markResolved(fault.ID)
		shm.sendAlert(HealingAlert{
			Timestamp: time.Now(),
			FaultID:   fault.ID,
			Action:    "healed",
			Success:   true,
			Message:   fmt.Sprintf("Successfully healed fault: %s", fault.Description),
			Duration:  duration,
		})
	} else {
		shm.sendAlert(HealingAlert{
			Timestamp: time.Now(),
			FaultID:   fault.ID,
			Action:    "heal_failed",
			Success:   false,
			Message:   fmt.Sprintf("Failed to heal fault: %v", err),
			Duration:  duration,
		})
	}
}

// markResolved marks a fault as resolved
func (shm *SelfHealingManager) markResolved(faultID string) {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	for i := range shm.faults {
		if shm.faults[i].ID == faultID {
			shm.faults[i].Resolved = true
			shm.faults[i].ResolvedAt = time.Now()
			break
		}
	}
}

// sendAlert sends a healing alert
func (shm *SelfHealingManager) sendAlert(alert HealingAlert) {
	shm.mu.Lock()
	callback := shm.alertCallback
	shm.mu.Unlock()

	if callback != nil {
		callback(alert)
	}
}

// Enable enables self-healing
func (shm *SelfHealingManager) Enable() {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	shm.enabled = true
}

// Disable disables self-healing
func (shm *SelfHealingManager) Disable() {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	shm.enabled = false
}

// GetFaults returns the fault history
func (shm *SelfHealingManager) GetFaults() []FaultRecord {
	shm.mu.Lock()
	defer shm.mu.Unlock()
	return append([]FaultRecord{}, shm.faults...)
}

// GetActiveFaults returns unresolved faults
func (shm *SelfHealingManager) GetActiveFaults() []FaultRecord {
	shm.mu.Lock()
	defer shm.mu.Unlock()

	active := make([]FaultRecord, 0)
	for _, fault := range shm.faults {
		if !fault.Resolved {
			active = append(active, fault)
		}
	}
	return active
}

// AutoRestartStrategy implementation
func (ars *AutoRestartStrategy) CanHandle(fault FaultRecord) bool {
	return fault.Attempts < ars.maxAttempts
}

func (ars *AutoRestartStrategy) Heal(fault FaultRecord) error {
	time.Sleep(ars.restartDelay)
	err := ars.restartFunc()
	if err != nil {
		ars.attempts++
	}
	return err
}

func (ars *AutoRestartStrategy) GetRecoveryTime() time.Duration {
	return ars.restartDelay
}

// RetryWithBackoffStrategy implementation
func (rwbs *RetryWithBackoffStrategy) CanHandle(fault FaultRecord) bool {
	return fault.Attempts < rwbs.maxRetries
}

func (rwbs *RetryWithBackoffStrategy) Heal(fault FaultRecord) error {
	delay := rwbs.initialDelay
	for i := 0; i < rwbs.maxRetries; i++ {
		if i > 0 {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * rwbs.backoffFactor)
			if delay > rwbs.maxDelay {
				delay = rwbs.maxDelay
			}
		}
		err := rwbs.retryFunc(i)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("max retries exceeded")
}

func (rwbs *RetryWithBackoffStrategy) GetRecoveryTime() time.Duration {
	totalDelay := rwbs.initialDelay
	delay := rwbs.initialDelay
	for i := 1; i < rwbs.maxRetries; i++ {
		delay = time.Duration(float64(delay) * rwbs.backoffFactor)
		if delay > rwbs.maxDelay {
			delay = rwbs.maxDelay
		}
		totalDelay += delay
	}
	return totalDelay
}

// CircuitBreakerStrategy implementation
func (cbs *CircuitBreakerStrategy) CanHandle(fault FaultRecord) bool {
	return true
}

func (cbs *CircuitBreakerStrategy) Heal(fault FaultRecord) error {
	now := time.Now()

	if cbs.state == CircuitOpen {
		if now.Sub(cbs.lastFailureTime) > cbs.recoveryTimeout {
			cbs.state = CircuitHalfOpen
			cbs.failures = 0
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	err := cbs.executeFunc()
	if err != nil {
		cbs.failures++
		cbs.lastFailureTime = now
		if cbs.failures >= cbs.failureThreshold {
			cbs.state = CircuitOpen
		}
		return err
	}

	if cbs.state == CircuitHalfOpen {
		cbs.state = CircuitClosed
		cbs.failures = 0
	}

	return nil
}

func (cbs *CircuitBreakerStrategy) GetRecoveryTime() time.Duration {
	return cbs.recoveryTimeout
}

func generateFaultID() string {
	return fmt.Sprintf("fault-%d", time.Now().UnixNano())
}

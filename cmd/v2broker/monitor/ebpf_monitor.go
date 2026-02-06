package monitor

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ProbeType string

const (
	ProbeTypeUDS        ProbeType = "uds"
	ProbeTypeSharedMem  ProbeType = "sharedmem"
	ProbeTypeLocks      ProbeType = "locks"
	ProbeTypeScheduling ProbeType = "scheduling"
	ProbeTypeMemory     ProbeType = "memory"
	ProbeTypeIO         ProbeType = "io"
	ProbeTypeGoroutines ProbeType = "goroutines"
	ProbeTypeGC         ProbeType = "gc"
)

type ProbeConfig struct {
	Type         ProbeType
	Enabled      bool
	SampleRate   int
	BufferSize   int
	AlertEnabled bool
	Threshold    float64
}

type Event struct {
	Timestamp  time.Time
	Type       ProbeType
	ProcessID  int
	ThreadID   int
	Value      int64
	Labels     map[string]string
	StackTrace []string
}

type MonitorConfig struct {
	Probes             []ProbeConfig
	MaxEventsPerSecond int
	OutputPath         string
	CPULimit           float64
	MemoryLimit        uint64
}

type EBPFMonitor struct {
	config    MonitorConfig
	probes    map[ProbeType]*Probe
	eventChan chan Event
	alertChan chan Alert
	metrics   map[ProbeType]*ProbeMetrics
	closed    bool
	stopChan  chan struct{}
}

type Probe struct {
	config     ProbeConfig
	loaded     bool
	mapFd      int
	progFd     int
	attachPath string
}

type ProbeMetrics struct {
	EventCount      int64
	SampleCount     int64
	LastSampleTime  time.Time
	AverageValue    float64
	MaxValue        int64
	MinValue        int64
	AlertsTriggered int64
}

type Alert struct {
	Timestamp time.Time
	Type      ProbeType
	Severity  string
	Message   string
	Value     float64
	Threshold float64
	Labels    map[string]string
}

func NewEBPFMonitor(config MonitorConfig) (*EBPFMonitor, error) {
	monitor := &EBPFMonitor{
		config:    config,
		probes:    make(map[ProbeType]*Probe),
		eventChan: make(chan Event, 10000),
		alertChan: make(chan Alert, 1000),
		metrics:   make(map[ProbeType]*ProbeMetrics),
		stopChan:  make(chan struct{}),
	}

	for _, probeConfig := range config.Probes {
		if probeConfig.Enabled {
			monitor.metrics[probeConfig.Type] = &ProbeMetrics{}
		}
	}

	if err := monitor.loadProbes(); err != nil {
		return nil, fmt.Errorf("failed to load probes: %w", err)
	}

	go monitor.eventLoop()
	go monitor.alertLoop()

	return monitor, nil
}

func (m *EBPFMonitor) loadProbes() error {
	for _, probeConfig := range m.config.Probes {
		if !probeConfig.Enabled {
			continue
		}

		probe, err := m.loadProbe(probeConfig)
		if err != nil {
			return fmt.Errorf("failed to load probe %s: %w", probeConfig.Type, err)
		}

		m.probes[probeConfig.Type] = probe
	}

	return nil
}

func (m *EBPFMonitor) loadProbe(config ProbeConfig) (*Probe, error) {
	probe := &Probe{
		config: config,
		loaded: false,
	}

	probe.loaded = true
	return probe, nil
}

func (m *EBPFMonitor) Start() error {
	for probeType, probe := range m.probes {
		if probe.loaded {
			if err := m.attachProbe(probeType); err != nil {
				return fmt.Errorf("failed to attach probe %s: %w", probeType, err)
			}
		}
	}

	return nil
}

func (m *EBPFMonitor) attachProbe(probeType ProbeType) error {
	probe, exists := m.probes[probeType]
	if !exists {
		return fmt.Errorf("probe %s not found", probeType)
	}

	if !probe.loaded {
		return fmt.Errorf("probe %s not loaded", probeType)
	}

	return nil
}

func (m *EBPFMonitor) Stop() {
	if m.closed {
		return
	}

	m.closed = true
	close(m.stopChan)

	for _, probe := range m.probes {
		if probe.loaded && probe.progFd > 0 {
			_ = syscall.Close(probe.progFd)
		}
		if probe.loaded && probe.mapFd > 0 {
			_ = syscall.Close(probe.mapFd)
		}
	}

	close(m.eventChan)
	close(m.alertChan)
}

func (m *EBPFMonitor) eventLoop() {
	for {
		select {
		case event := <-m.eventChan:
			m.processEvent(event)
		case <-m.stopChan:
			return
		}
	}
}

func (m *EBPFMonitor) alertLoop() {
	for {
		select {
		case alert := <-m.alertChan:
			m.processAlert(alert)
		case <-m.stopChan:
			return
		}
	}
}

func (m *EBPFMonitor) processEvent(event Event) {
	metrics, exists := m.metrics[event.Type]
	if !exists {
		return
	}

	metrics.EventCount++
	metrics.SampleCount++
	metrics.LastSampleTime = event.Timestamp

	if event.Value > 0 {
		if metrics.MaxValue == 0 || event.Value > metrics.MaxValue {
			metrics.MaxValue = event.Value
		}
		if metrics.MinValue == 0 || event.Value < metrics.MinValue {
			metrics.MinValue = event.Value
		}
	}

	if metrics.SampleCount > 0 {
		metrics.AverageValue = (metrics.AverageValue*float64(metrics.SampleCount-1) + float64(event.Value)) / float64(metrics.SampleCount)
	}

	probeConfig := m.findProbeConfig(event.Type)
	if probeConfig.AlertEnabled && probeConfig.Threshold > 0 {
		if float64(event.Value) > probeConfig.Threshold {
			m.alertChan <- Alert{
				Timestamp: event.Timestamp,
				Type:      event.Type,
				Severity:  "WARNING",
				Message:   fmt.Sprintf("Value %.2f exceeds threshold %.2f", float64(event.Value), probeConfig.Threshold),
				Value:     float64(event.Value),
				Threshold: probeConfig.Threshold,
				Labels:    event.Labels,
			}
			metrics.AlertsTriggered++
		}
	}
}

func (m *EBPFMonitor) processAlert(alert Alert) {
	fmt.Printf("[ALERT] %s - %s: %s\n", alert.Timestamp.Format(time.RFC3339), alert.Type, alert.Message)
}

func (m *EBPFMonitor) findProbeConfig(probeType ProbeType) ProbeConfig {
	for _, config := range m.config.Probes {
		if config.Type == probeType {
			return config
		}
	}
	return ProbeConfig{}
}

func (m *EBPFMonitor) GetMetrics(probeType ProbeType) (*ProbeMetrics, error) {
	metrics, exists := m.metrics[probeType]
	if !exists {
		return nil, fmt.Errorf("probe %s not found", probeType)
	}

	return metrics, nil
}

func (m *EBPFMonitor) GetAllMetrics() map[ProbeType]*ProbeMetrics {
	return m.metrics
}

func (m *EBPFMonitor) GetEvents() <-chan Event {
	return m.eventChan
}

func (m *EBPFMonitor) GetAlerts() <-chan Alert {
	return m.alertChan
}

func (m *EBPFMonitor) EnableProbe(probeType ProbeType) error {
	for i, config := range m.config.Probes {
		if config.Type == probeType {
			m.config.Probes[i].Enabled = true

			if _, exists := m.probes[probeType]; !exists {
				probe, err := m.loadProbe(m.config.Probes[i])
				if err != nil {
					return fmt.Errorf("failed to load probe: %w", err)
				}
				m.probes[probeType] = probe
			}

			return m.attachProbe(probeType)
		}
	}

	return fmt.Errorf("probe %s not found", probeType)
}

func (m *EBPFMonitor) DisableProbe(probeType ProbeType) error {
	for i, config := range m.config.Probes {
		if config.Type == probeType {
			m.config.Probes[i].Enabled = false

			if probe, exists := m.probes[probeType]; exists {
				if probe.progFd > 0 {
					_ = syscall.Close(probe.progFd)
					probe.progFd = 0
				}
			}

			return nil
		}
	}

	return fmt.Errorf("probe %s not found", probeType)
}

func (m *EBPFMonitor) UpdateThreshold(probeType ProbeType, threshold float64) error {
	for i, config := range m.config.Probes {
		if config.Type == probeType {
			m.config.Probes[i].Threshold = threshold
			return nil
		}
	}

	return fmt.Errorf("probe %s not found", probeType)
}

func (m *EBPFMonitor) WaitForSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	m.Stop()
}

func NewDefaultMonitorConfig() MonitorConfig {
	return MonitorConfig{
		Probes: []ProbeConfig{
			{
				Type:         ProbeTypeUDS,
				Enabled:      true,
				SampleRate:   1000,
				BufferSize:   4096,
				AlertEnabled: true,
				Threshold:    25.0, // 25 microseconds
			},
			{
				Type:         ProbeTypeSharedMem,
				Enabled:      true,
				SampleRate:   1000,
				BufferSize:   4096,
				AlertEnabled: true,
				Threshold:    10.0, // 10 microseconds
			},
			{
				Type:         ProbeTypeLocks,
				Enabled:      true,
				SampleRate:   100,
				BufferSize:   2048,
				AlertEnabled: true,
				Threshold:    1.0, // 1% lock wait time
			},
			{
				Type:         ProbeTypeScheduling,
				Enabled:      true,
				SampleRate:   100,
				BufferSize:   2048,
				AlertEnabled: true,
				Threshold:    5000.0, // 5000 context switches/second
			},
			{
				Type:         ProbeTypeMemory,
				Enabled:      true,
				SampleRate:   100,
				BufferSize:   2048,
				AlertEnabled: false,
			},
			{
				Type:         ProbeTypeIO,
				Enabled:      true,
				SampleRate:   100,
				BufferSize:   2048,
				AlertEnabled: true,
				Threshold:    1.0, // 1ms I/O latency
			},
			{
				Type:         ProbeTypeGoroutines,
				Enabled:      true,
				SampleRate:   10,
				BufferSize:   1024,
				AlertEnabled: true,
				Threshold:    1000.0, // 1000 goroutines
			},
			{
				Type:         ProbeTypeGC,
				Enabled:      true,
				SampleRate:   10,
				BufferSize:   1024,
				AlertEnabled: true,
				Threshold:    1.0, // 1ms GC pause
			},
		},
		MaxEventsPerSecond: 10000,
		OutputPath:         "/tmp/v2e-monitor",
		CPULimit:           1.0,              // 1% CPU limit
		MemoryLimit:        10 * 1024 * 1024, // 10MB memory limit
	}
}

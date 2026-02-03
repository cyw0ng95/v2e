package main

import (
	core "github.com/cyw0ng95/v2e/cmd/broker/core"
)

// Re-export core types for compatibility within cmd/broker package.
type (
	Broker          = core.Broker
	Process         = core.Process
	ProcessInfo     = core.ProcessInfo
	ProcessStatus   = core.ProcessStatus
	RestartConfig   = core.RestartConfig
	MessageStats    = core.MessageStats
	PerProcessStats = core.PerProcessStats
	PendingRequest  = core.PendingRequest
)

const (
	ProcessStatusRunning = core.ProcessStatusRunning
	ProcessStatusExited  = core.ProcessStatusExited
	ProcessStatusFailed  = core.ProcessStatusFailed
)

var (
	NewBroker      = core.NewBroker
	NewTestProcess = core.NewTestProcess
)

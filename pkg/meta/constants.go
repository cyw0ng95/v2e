package meta

import "time"

const (
	// Meta Log Messages
	LogMsgMetaStoreCreated      = "Meta store created successfully"
	LogMsgMetaStoreClosed       = "Meta store closed"
	LogMsgDatabaseQueryExecuted = "Database query executed: table=meta, operation=%s"

	// Provider States
	ProviderStateIdle           = "IDLE"
	ProviderStateAcquiring      = "ACQUIRING"
	ProviderStateRunning        = "RUNNING"
	ProviderStateWaitingQuota   = "WAITING_QUOTA"
	ProviderStateWaitingBackoff = "WAITING_BACKOFF"
	ProviderStatePaused         = "PAUSED"
	ProviderStateTerminated     = "TERMINATED"

	// FSM States
	FSMStateBootstrapping = "BOOTSTRAPPING"
	FSMStateOrchestrating = "ORCHESTRATING"
	FSMStateStabilizing   = "STABILIZING"
	FSMStateDraining      = "DRAINING"

	// Permit Management
	MaxPermitsPerProvider = 10
	PermitAcquireTimeout  = 30 * time.Second

	// Memory Card
	MaxMemoryCards        = 100
	CardRetentionDuration = 7 * 24 * time.Hour // 7 days
)

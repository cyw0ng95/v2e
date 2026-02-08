package main

import (
	"context"
	"fmt"

	asvsprovider "github.com/cyw0ng95/v2e/pkg/asvs/provider"
	attackprovider "github.com/cyw0ng95/v2e/pkg/attack/provider"
	capecprovider "github.com/cyw0ng95/v2e/pkg/capec/provider"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve/provider"
	cweprovider "github.com/cyw0ng95/v2e/pkg/cwe/provider"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	ssgprovider "github.com/cyw0ng95/v2e/pkg/ssg/provider"
)

// Global FSM manager and providers
var (
	macroFSM  *fsm.MacroFSMManager
	providers []fsm.ProviderFSM
	storageDB *storage.Store
)

// initFSM initializes the FSM infrastructure
func initFSM(logger *common.Logger, dbPath string) error {
	// Initialize storage
	var err error
	storageDB, err = storage.NewStore(dbPath, logger)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Create macro FSM manager
	macroFSM, err = fsm.NewMacroFSMManager("uee-orchestrator", storageDB)
	if err != nil {
		return fmt.Errorf("failed to create macro FSM: %w", err)
	}

	// Initialize all data source providers
	if err := initFSMProviders(logger); err != nil {
		return fmt.Errorf("failed to initialize providers: %w", err)
	}

	logger.Info("FSM infrastructure initialized successfully")
	return nil
}

// initFSMProviders initializes all FSM-based providers
func initFSMProviders(logger *common.Logger) error {
	var providerList []fsm.ProviderFSM

	// CVE Provider
	cveProv, err := provider.NewCVEProvider("", storageDB)
	if err != nil {
		return fmt.Errorf("failed to create CVE provider: %w", err)
	}
	providerList = append(providerList, cveProv)

	// CWE Provider
	cweProv, err := cweprovider.NewCWEProvider("", storageDB)
	if err != nil {
		return fmt.Errorf("failed to create CWE provider: %w", err)
	}
	providerList = append(providerList, cweProv)

	// CAPEC Provider
	capecProv, err := capecprovider.NewCAPECProvider("", storageDB)
	if err != nil {
		return fmt.Errorf("failed to create CAPEC provider: %w", err)
	}
	providerList = append(providerList, capecProv)

	// ATT&CK Provider
	attackProv, err := attackprovider.NewATTACKProvider("", storageDB)
	if err != nil {
		return fmt.Errorf("failed to create ATT&CK provider: %w", err)
	}
	providerList = append(providerList, attackProv)

	// SSG Provider
	ssgProv, err := ssgprovider.NewSSGProvider("", storageDB)
	if err != nil {
		logger.Warn("Failed to create SSG provider (skipping): %v", err)
	} else {
		providerList = append(providerList, ssgProv)
	}

	// ASVS Provider
	asvsProv, err := asvsprovider.NewASVSProvider("", storageDB)
	if err != nil {
		logger.Warn("Failed to create ASVS provider (skipping): %v", err)
	} else {
		providerList = append(providerList, asvsProv)
	}

	// Initialize each provider
	for _, p := range providerList {
		if err := p.Initialize(context.Background()); err != nil {
			logger.Error("Failed to initialize provider %s: %v", p.GetID(), err)
			continue
		}
	}

	// Register providers with macro FSM
	for _, p := range providerList {
		if err := macroFSM.AddProvider(p); err != nil {
			logger.Error("Failed to add provider %s to macro FSM: %v", p.GetID(), err)
		}
	}

	providers = providerList
	logger.Info("All FSM providers initialized and registered")
	return nil
}

// recoverRunningProviders recovers state for providers that were running before shutdown
func recoverRunningProviders(logger *common.Logger) error {
	providerStates, err := storageDB.ListProviderStates()
	if err != nil {
		return fmt.Errorf("failed to list provider states: %w", err)
	}

	for _, state := range providerStates {
		provider := getProvider(state.ID)
		if provider == nil {
			logger.Warn("Provider %s not found in registry, skipping recovery", state.ID)
			continue
		}

		currentFSMState := fsm.ProviderState(state.State)

		// Recovery strategy per state
		switch currentFSMState {
		case fsm.ProviderRunning:
			// Resume execution via ACQUIRING state
			logger.Info("Resuming provider %s from RUNNING state", state.ID)
			if err := provider.Resume(); err != nil {
				logger.Error("Failed to resume provider %s: %v", state.ID, err)
			}

		case fsm.ProviderWaitingQuota:
			// Transition to ACQUIRING to retry quota acquisition
			logger.Info("Provider %s in WAITING_QUOTA, retrying acquisition", state.ID)
			if err := provider.Transition(fsm.ProviderAcquiring); err != nil {
				logger.Error("Failed to transition provider %s: %v", state.ID, err)
			}

		case fsm.ProviderPaused:
			// Keep paused - manual resume needed
			logger.Info("Provider %s is PAUSED, keeping paused", state.ID)

		case fsm.ProviderWaitingBackoff:
			// Maintain state - auto-retry timer will handle it
			logger.Info("Provider %s in WAITING_BACKOFF, maintaining state", state.ID)

		case fsm.ProviderTerminated, fsm.ProviderIdle:
			// Skip recovery
			logger.Info("Provider %s in %s state, skipping recovery", state.ID, state.State)

		default:
			logger.Warn("Provider %s in unknown state %s", state.ID, state.State)
		}
	}

	return nil
}

// getProvider returns a provider by ID
func getProvider(id string) fsm.ProviderFSM {
	for _, p := range providers {
		if p.GetID() == id {
			return p
		}
	}
	return nil
}

// RPCGetProviderList returns a list of all FSM providers with their states
func RPCGetProviderList() ([]ProviderStatusItem, error) {
	statusItems := make([]ProviderStatusItem, 0, len(providers))

	for _, p := range providers {
		stats := p.GetStats()
		statusItems = append(statusItems, ProviderStatusItem{
			ID:       stats["id"].(string),
			Type:     stats["provider_type"].(string),
			State:    stats["state"].(string),
			Progress: &ProviderProgress{},
		})
	}

	return statusItems, nil
}

// RPCStartProvider starts a specific provider by ID
func RPCStartProvider(providerID string) (bool, error) {
	provider := getProvider(providerID)
	if provider == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := provider.Start(); err != nil {
		return false, fmt.Errorf("failed to start provider: %w", err)
	}

	return true, nil
}

// RPCStopProvider stops a specific provider by ID
func RPCStopProvider(providerID string) (bool, error) {
	provider := getProvider(providerID)
	if provider == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := provider.Stop(); err != nil {
		return false, fmt.Errorf("failed to stop provider: %w", err)
	}

	return true, nil
}

// RPCPauseProvider pauses a specific provider by ID
func RPCPauseProvider(providerID string) (bool, error) {
	provider := getProvider(providerID)
	if provider == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := provider.Pause(); err != nil {
		return false, fmt.Errorf("failed to pause provider: %w", err)
	}

	return true, nil
}

// RPCResumeProvider resumes a specific provider by ID
func RPCResumeProvider(providerID string) (bool, error) {
	provider := getProvider(providerID)
	if provider == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := provider.Resume(); err != nil {
		return false, fmt.Errorf("failed to resume provider: %w", err)
	}

	return true, nil
}

// RPCGetEtlTree returns the hierarchical ETL tree (Macro FSM + all Provider FSMs)
func RPCGetEtlTree() (map[string]interface{}, error) {
	macroStats := macroFSM.GetStats()

	providerStats := make([]map[string]interface{}, 0, len(providers))
	for _, p := range providers {
		providerStats = append(providerStats, p.GetStats())
	}

	return map[string]interface{}{
		"macro_fsm": macroStats,
		"providers": providerStats,
	}, nil
}

// RPCGetProviderCheckpoints returns checkpoints for a specific provider
func RPCGetProviderCheckpoints(providerID string, limit int, successOnly bool) ([]map[string]interface{}, error) {
	checkpoints, err := storageDB.ListCheckpointsByProvider(providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoints: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(checkpoints))

	for _, cp := range checkpoints {
		if successOnly && !cp.Success {
			continue
		}

		if limit > 0 && len(result) >= limit {
			break
		}

		result = append(result, map[string]interface{}{
			"urn":           cp.URN,
			"provider_id":   cp.ProviderID,
			"processed_at":  cp.ProcessedAt.Format(time.RFC3339),
			"success":       cp.Success,
			"error_message": cp.ErrorMessage,
		})
	}

	return result, nil
}

// createFSMRPCHandlers creates FSM-related RPC handlers
func createFSMRPCHandlers() map[string]subprocess.Handler {
	return map[string]subprocess.Handler{
		"RPCGetProviderList":        handleGetProviderList,
		"RPCStartProvider":          handleStartProvider,
		"RPCStopProvider":           handleStopProvider,
		"RPCPauseProvider":          handlePauseProvider,
		"RPCResumeProvider":         handleResumeProvider,
		"RPCGetEtlTree":             handleGetEtlTree,
		"RPCGetProviderCheckpoints": handleGetProviderCheckpoints,
	}
}

// FSM RPC handler wrappers
func handleGetProviderList(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	providers, err := RPCGetProviderList()
	if err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}
	return subprocess.NewSuccessResponse(msg, map[string]interface{}{
		"providers": providers,
	})
}

func handleStartProvider(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := msg.UnmarshalParams(&params); err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}

	providerID, ok := params["provider_id"].(string)
	if !ok {
		return subprocess.NewErrorResponse(msg, "provider_id parameter required"), nil
	}

	success, err := RPCStartProvider(providerID)
	if err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}
	return subprocess.NewSuccessResponse(msg, map[string]interface{}{
		"success": success,
	})
}

func handleStopProvider(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := msg.UnmarshalParams(&params); err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}

	providerID, ok := params["provider_id"].(string)
	if !ok {
		return subprocess.NewErrorResponse(msg, "provider_id parameter required"), nil
	}

	success, err := RPCStopProvider(providerID)
	if err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}
	return subprocess.NewSuccessResponse(msg, map[string]interface{}{
		"success": success,
	})
}

func handlePauseProvider(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := msg.UnmarshalParams(&params); err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}

	providerID, ok := params["provider_id"].(string)
	if !ok {
		return subprocess.NewErrorResponse(msg, "provider_id parameter required"), nil
	}

	success, err := RPCPauseProvider(providerID)
	if err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}
	return subprocess.NewSuccessResponse(msg, map[string]interface{}{
		"success": success,
	})
}

func handleResumeProvider(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := msg.UnmarshalParams(&params); err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}

	providerID, ok := params["provider_id"].(string)
	if !ok {
		return subprocess.NewErrorResponse(msg, "provider_id parameter required"), nil
	}

	success, err := RPCResumeProvider(providerID)
	if err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}
	return subprocess.NewSuccessResponse(msg, map[string]interface{}{
		"success": success,
	})
}

func handleGetEtlTree(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	tree, err := RPCGetEtlTree()
	if err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}
	return subprocess.NewSuccessResponse(msg, tree)
}

func handleGetProviderCheckpoints(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	var params map[string]interface{}
	if err := msg.UnmarshalParams(&params); err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}

	providerID, ok := params["provider_id"].(string)
	if !ok {
		return subprocess.NewErrorResponse(msg, "provider_id parameter required"), nil
	}

	limit := 100
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	successOnly := false
	if s, ok := params["success_only"].(bool); ok {
		successOnly = s
	}

	checkpoints, err := RPCGetProviderCheckpoints(providerID, limit, successOnly)
	if err != nil {
		return subprocess.NewErrorResponse(msg, err.Error()), err
	}

	total := len(storageDB.ListCheckpointsByProvider(providerID))

	return subprocess.NewSuccessResponse(msg, map[string]interface{}{
		"checkpoints": checkpoints,
		"total":       total,
	})
}

package main

import (
	"context"
	"fmt"
	"time"

	attackprovider "github.com/cyw0ng95/v2e/pkg/attack/provider"
	capecprovider "github.com/cyw0ng95/v2e/pkg/capec/provider"
	cceprovider "github.com/cyw0ng95/v2e/pkg/cce/provider"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve/provider"
	cweprovider "github.com/cyw0ng95/v2e/pkg/cwe/provider"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// FSM infrastructure globals
var (
	macroFSM     *fsm.MacroFSMManager
	fsmProviders map[string]fsm.ProviderFSM
	storageDB    *storage.Store
	fsmSubprocess *subprocess.Subprocess
)

// initFSMInfrastructure initializes the UEE FSM infrastructure
func initFSMInfrastructure(logger *common.Logger, runDBPath string, sp *subprocess.Subprocess) error {
	// Initialize storage
	var err error
	storageDB, err = storage.NewStore(runDBPath, logger)
	if err != nil {
		return fmt.Errorf("failed to create FSM storage: %w", err)
	}

	// Store subprocess for provider initialization
	fsmSubprocess = sp

	// Create macro FSM manager
	macroFSM, err = fsm.NewMacroFSMManager("uee-orchestrator", storageDB)
	if err != nil {
		return fmt.Errorf("failed to create macro FSM: %w", err)
	}

	// Initialize provider map
	fsmProviders = make(map[string]fsm.ProviderFSM)

	// Initialize all FSM-based providers
	if err := initFSMProviders(logger); err != nil {
		return fmt.Errorf("failed to initialize FSM providers: %w", err)
	}

	// Recover running providers
	if err := recoverRunningFSMProviders(logger); err != nil {
		logger.Warn("Failed to recover providers: %v", err)
	}

	logger.Info("UEE FSM infrastructure initialized successfully")
	return nil
}

// initFSMProviders initializes all FSM-based providers and registers them with macro FSM
// Provider dependencies are configured to ensure correct startup order:
// - CAPEC depends on CWE (CAPEC data references CWE weaknesses)
// - ATT&CK depends on CAPEC (ATT&CK techniques may reference CAPEC patterns)
func initFSMProviders(logger *common.Logger) error {
	// CVE Provider (no dependencies)
	cveProv, err := provider.NewCVEProvider("", storageDB, fsmSubprocess)
	if err != nil {
		logger.Warn("Failed to create CVE provider (skipping): %v", err)
	} else {
		fsmProviders["cve"] = cveProv
		if err := macroFSM.AddProvider(cveProv); err != nil {
			logger.Error("Failed to add CVE provider to macro FSM: %v", err)
		}
	}

	// CWE Provider (no dependencies)
	cweProv, err := cweprovider.NewCWEProvider("", storageDB, fsmSubprocess)
	if err != nil {
		logger.Warn("Failed to create CWE provider (skipping): %v", err)
	} else {
		fsmProviders["cwe"] = cweProv
		if err := macroFSM.AddProvider(cweProv); err != nil {
			logger.Error("Failed to add CWE provider to macro FSM: %v", err)
		}
	}

	// CAPEC Provider (depends on CWE)
	capecDependencies := []string{"cwe"}
	capecProv, err := capecprovider.NewCAPECProvider("", storageDB, fsmSubprocess, capecDependencies)
	if err != nil {
		logger.Warn("Failed to create CAPEC provider (skipping): %v", err)
	} else {
		fsmProviders["capec"] = capecProv
		if err := macroFSM.AddProvider(capecProv); err != nil {
			logger.Error("Failed to add CAPEC provider to macro FSM: %v", err)
		}
	}

	// ATT&CK Provider (depends on CAPEC)
	attackDependencies := []string{"capec"}
	attackProv, err := attackprovider.NewATTACKProvider("", storageDB, attackDependencies)
	if err != nil {
		logger.Warn("Failed to create ATT&CK provider (skipping): %v", err)
	} else {
		fsmProviders["attack"] = attackProv
		if err := macroFSM.AddProvider(attackProv); err != nil {
			logger.Error("Failed to add ATT&CK provider to macro FSM: %v", err)
		}
	}

	// CCE Provider (no dependencies)
	cceProv, err := cceprovider.NewCCEProvider("", storageDB)
	if err != nil {
		logger.Warn("Failed to create CCE provider (skipping): %v", err)
	} else {
		fsmProviders["cce"] = cceProv
		if err := macroFSM.AddProvider(cceProv); err != nil {
			logger.Error("Failed to add CCE provider to macro FSM: %v", err)
		}
	}

	// Initialize each provider
	for id, provider := range fsmProviders {
		if err := provider.Initialize(context.Background()); err != nil {
			logger.Error("Failed to initialize provider %s: %v", id, err)
		}
	}

	logger.Info("All FSM providers initialized and registered with macro FSM")
	return nil
}

// recoverRunningFSMProviders recovers state for providers that were running before shutdown
func recoverRunningFSMProviders(logger *common.Logger) error {
	providerStates, err := storageDB.ListProviderStates()
	if err != nil {
		return fmt.Errorf("failed to list provider states: %w", err)
	}

	recoveredCount := 0
	for _, state := range providerStates {
		provider := fsmProviders[state.ID]
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
			} else {
				recoveredCount++
			}

		case fsm.ProviderWaitingQuota:
			// Transition to ACQUIRING to retry quota acquisition
			logger.Info("Provider %s in WAITING_QUOTA, retrying acquisition", state.ID)
			if err := provider.Transition(fsm.ProviderAcquiring); err != nil {
				logger.Error("Failed to transition provider %s: %v", state.ID, err)
			} else {
				recoveredCount++
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

	if recoveredCount > 0 {
		logger.Info("Recovered %d providers to RUNNING/ACQUIRING state", recoveredCount)
	}

	return nil
}

// GetFSMProvider returns a provider by ID
func GetFSMProvider(id string) fsm.ProviderFSM {
	return fsmProviders[id]
}

// GetFSMProviders returns all FSM providers
func GetFSMProviders() map[string]fsm.ProviderFSM {
	return fsmProviders
}

// GetMacroFSM returns the macro FSM manager
func GetMacroFSM() *fsm.MacroFSMManager {
	return macroFSM
}

// GetFSMStorageDB returns the FSM storage database
func GetFSMStorageDB() *storage.Store {
	return storageDB
}

// CreateFSMRPCHandlers returns handlers for FSM control operations
func CreateFSMRPCHandlers(logger *common.Logger) map[string]subprocess.Handler {
	return map[string]subprocess.Handler{
		"RPCFSMStartProvider":          createFSMStartProviderHandler(logger),
		"RPCFSMStopProvider":           createFSMStopProviderHandler(logger),
		"RPCFSMPauseProvider":          createFSMPauseProviderHandler(logger),
		"RPCFSMResumeProvider":         createFSMResumeProviderHandler(logger),
		"RPCFSMGetProviderList":        createFSMGetProviderListHandler(logger),
		"RPCFSMGetProviderCheckpoints": createFSMGetProviderCheckpointsHandler(logger),
		"RPCFSMGetEtlTree":             createFSMGetEtlTreeHandler(logger),
	}
}

// FSM RPC handlers
func createFSMStartProviderHandler(logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params map[string]interface{}
		if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
			return subprocess.NewErrorResponse(msg, err.Error()), nil
		}

		providerID, ok := params["provider_id"].(string)
		if !ok {
			return subprocess.NewErrorResponse(msg, "provider_id parameter required"), nil
		}

		provider := GetFSMProvider(providerID)
		if provider == nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("provider not found: %s", providerID)), nil
		}

		if err := provider.Start(); err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to start provider: %v", err)), nil
		}

		logger.Info("Started provider: %s", providerID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success":     true,
			"provider_id": providerID,
		})
	}
}

func createFSMStopProviderHandler(logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params map[string]interface{}
		if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
			return subprocess.NewErrorResponse(msg, err.Error()), nil
		}

		providerID, ok := params["provider_id"].(string)
		if !ok {
			return subprocess.NewErrorResponse(msg, "provider_id parameter required"), nil
		}

		provider := GetFSMProvider(providerID)
		if provider == nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("provider not found: %s", providerID)), nil
		}

		if err := provider.Stop(); err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to stop provider: %v", err)), nil
		}

		logger.Info("Stopped provider: %s", providerID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success":     true,
			"provider_id": providerID,
		})
	}
}

func createFSMPauseProviderHandler(logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params map[string]interface{}
		if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
			return subprocess.NewErrorResponse(msg, err.Error()), nil
		}

		providerID, ok := params["provider_id"].(string)
		if !ok {
			return subprocess.NewErrorResponse(msg, "provider_id parameter required"), nil
		}

		provider := GetFSMProvider(providerID)
		if provider == nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("provider not found: %s", providerID)), nil
		}

		if err := provider.Pause(); err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to pause provider: %v", err)), nil
		}

		logger.Info("Paused provider: %s", providerID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success":     true,
			"provider_id": providerID,
		})
	}
}

func createFSMResumeProviderHandler(logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params map[string]interface{}
		if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
			return subprocess.NewErrorResponse(msg, err.Error()), nil
		}

		providerID, ok := params["provider_id"].(string)
		if !ok {
			return subprocess.NewErrorResponse(msg, "provider_id parameter required"), nil
		}

		provider := GetFSMProvider(providerID)
		if provider == nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("provider not found: %s", providerID)), nil
		}

		if err := provider.Resume(); err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to resume provider: %v", err)), nil
		}

		logger.Info("Resumed provider: %s", providerID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success":     true,
			"provider_id": providerID,
		})
	}
}

func createFSMGetProviderListHandler(logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		providers := make([]map[string]interface{}, 0, len(fsmProviders))

		for id, provider := range fsmProviders {
			providers = append(providers, map[string]interface{}{
				"id":    id,
				"type":  provider.GetType(),
				"state": string(provider.GetState()),
			})
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"providers": providers,
			"count":     len(providers),
		})
	}
}

func createFSMGetProviderCheckpointsHandler(logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params map[string]interface{}
		if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
			return subprocess.NewErrorResponse(msg, err.Error()), nil
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

		checkpoints, err := storageDB.ListCheckpointsByProvider(providerID)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get checkpoints: %v", err)), nil
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

		total := len(checkpoints)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"checkpoints": result,
			"total":       total,
		})
	}
}

func createFSMGetEtlTreeHandler(logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		macroStats := macroFSM.GetStats()

		providerStats := make([]map[string]interface{}, 0, len(fsmProviders))
		for _, provider := range fsmProviders {
			stats := provider.GetStats()
			providerStats = append(providerStats, stats)
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"macro_fsm": macroStats,
			"providers": providerStats,
		})
	}
}

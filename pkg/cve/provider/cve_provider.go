package provider

import (
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve/remote"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// CVEProvider implements ProviderFSM for CVE data
type CVEProvider struct {
	*fsm.BaseProviderFSM
	fetcher *remote.Fetcher
	apiKey  string
	sp      *subprocess.Subprocess
}

// NewCVEProvider creates a new CVE provider with FSM support
// apiKey is optional NVD API key for higher rate limits
// sp is the subprocess for RPC communication (can be nil for testing)
func NewCVEProvider(apiKey string, store *storage.Store, sp *subprocess.Subprocess) (*CVEProvider, error) {
	fetcher, err := remote.NewFetcher(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create fetcher: %w", err)
	}

	provider := &CVEProvider{
		fetcher: fetcher,
		apiKey:  apiKey,
		sp:      sp,
	}

	// Create base FSM with custom executor
	base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           "cve",
		ProviderType: "cve",
		Storage:      store,
		Executor:     provider.execute,
		BatchSize:    2000,
	})
	if err != nil {
		return nil, err
	}

	provider.BaseProviderFSM = base
	return provider, nil
}

// execute performs CVE fetch and store operations
func (p *CVEProvider) execute() error {
	currentState := p.GetState()

	// Check if we should be running
	if currentState != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", currentState)
	}

	// Require subprocess for RPC calls
	if p.sp == nil {
		return fmt.Errorf("subprocess not configured, cannot make RPC calls")
	}

	batchSize := p.GetBatchSize()
	fetcher := p.fetcher

	startIndex := 0

	for {
		// Check for cancellation or state change
		if p.GetState() != fsm.ProviderRunning {
			break
		}

		// Fetch batch from NVD API
		resp, err := fetcher.FetchCVEs(startIndex, batchSize)
		if err != nil {
			// Handle rate limiting
			if err == remote.ErrRateLimited {
				if stateErr := p.OnRateLimited(30 * time.Second); stateErr != nil {
					return fmt.Errorf("rate limit handling failed: %w", stateErr)
				}
				// Backoff and retry
				time.Sleep(30 * time.Second)
				continue
			}

			// Handle other errors
			return fmt.Errorf("failed to fetch CVEs at index %d: %w", startIndex, err)
		}

		// Process each CVE in batch
		count := len(resp.Vulnerabilities)
		for _, vuln := range resp.Vulnerabilities {
			cveID := vuln.CVE.ID

			// Create URN for checkpointing
			itemURN := urn.MustParse(fmt.Sprintf("v2e::nvd::cve::%s", cveID))

			// Store CVE via RPC call to v2local service
			// Build RPC request message
			params := map[string]interface{}{
				"cve": vuln.CVE,
			}

			success := true
			errorMsg := ""

			// Send RPC request to broker to route to local service
			payload, err := subprocess.MarshalFast(params)
			if err != nil {
				success = false
				errorMsg = fmt.Sprintf("failed to marshal RPC request for %s: %v", cveID, err)
			} else {
				rpcMsg := &subprocess.Message{
					Type:    subprocess.MessageTypeRequest,
					ID:      "RPCSaveCVEByID",
					Payload: payload,
					Target:  "local",
					Source:  p.sp.ID,
				}

				if err := p.sp.SendMessage(rpcMsg); err != nil {
					success = false
					errorMsg = fmt.Sprintf("failed to send RPC request for %s: %v", cveID, err)
				}
				// For now, assume success - proper async response handling
				// would require waiting for response channel
			}

			// Save checkpoint
			if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
				return fmt.Errorf("failed to save checkpoint for %s: %w", cveID, err)
			}

			if !success {
				// Log error but continue processing
				fmt.Printf("[CVEProvider] Error saving %s: %s\n", cveID, errorMsg)
			}
		}

		// Check if we've fetched all CVEs
		if startIndex+count >= resp.TotalResults {
			break
		}

		startIndex += count
	}

	return nil
}

// GetAPIKey returns the NVD API key
func (p *CVEProvider) GetAPIKey() string {
	return p.apiKey
}

// GetFetcher returns the NVD fetcher
func (p *CVEProvider) GetFetcher() *remote.Fetcher {
	return p.fetcher
}

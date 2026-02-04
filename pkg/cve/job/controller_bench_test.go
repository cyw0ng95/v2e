package job

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/session"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// BenchmarkControllerNew benchmarks controller creation
func BenchmarkControllerNew(b *testing.B) {
	logger := common.NewLogger(os.Stderr, "bench", common.InfoLevel)
	dbPath := filepath.Join(b.TempDir(), "bench_controller.db")

	// Add logger setup for NewManager calls
	logger = common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// Update NewManager calls to include logger
	sessionManager, err := session.NewManager(dbPath, logger)
	if err != nil {
		b.Fatalf("Failed to create session manager: %v", err)
	}
	defer sessionManager.Close()

	mockRPC := &mockRPCInvoker{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewController(mockRPC, sessionManager, logger)
	}
}

// BenchmarkRPCInvoke benchmarks the overhead of RPC invocation in job loop
func BenchmarkRPCInvoke(b *testing.B) {
	emptyResponse := &cve.CVEResponse{Vulnerabilities: []struct {
		CVE cve.CVEItem `json:"cve"`
	}{}}
	emptyPayload, _ := sonic.Marshal(emptyResponse)
	emptyMsg := &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: emptyPayload}

	mockRPC := &mockRPCInvoker{fetchResponse: emptyMsg}

	ctx := context.Background()
	params := map[string]interface{}{
		"start_index":      0,
		"results_per_page": 100,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := mockRPC.InvokeRPC(ctx, "remote", "RPCFetchCVEs", params)
		if err != nil {
			b.Fatalf("RPC invocation failed: %v", err)
		}
	}
}

// BenchmarkMessageSerialization benchmarks CVE response serialization
func BenchmarkMessageSerialization(b *testing.B) {
	// Create a realistic CVE response
	cves := make([]struct {
		CVE cve.CVEItem `json:"cve"`
	}, 100)
	for i := 0; i < 100; i++ {
		cves[i] = struct {
			CVE cve.CVEItem `json:"cve"`
		}{
			CVE: cve.CVEItem{
				ID:         "CVE-2021-00001",
				SourceID:   "nvd",
				VulnStatus: "Analyzed",
			},
		}
	}

	response := &cve.CVEResponse{
		Vulnerabilities: cves,
		TotalResults:    100,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := sonic.Marshal(response)
		if err != nil {
			b.Fatalf("Serialization failed: %v", err)
		}
	}
}

// BenchmarkMessageDeserialization benchmarks CVE response deserialization
func BenchmarkMessageDeserialization(b *testing.B) {
	// Create a realistic CVE response
	cves := make([]struct {
		CVE cve.CVEItem `json:"cve"`
	}, 100)
	for i := 0; i < 100; i++ {
		cves[i] = struct {
			CVE cve.CVEItem `json:"cve"`
		}{
			CVE: cve.CVEItem{
				ID:         "CVE-2021-00001",
				SourceID:   "nvd",
				VulnStatus: "Analyzed",
			},
		}
	}

	response := &cve.CVEResponse{
		Vulnerabilities: cves,
		TotalResults:    100,
	}

	payload, _ := sonic.Marshal(response)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result cve.CVEResponse
		err := sonic.Unmarshal(payload, &result)
		if err != nil {
			b.Fatalf("Deserialization failed: %v", err)
		}
	}
}

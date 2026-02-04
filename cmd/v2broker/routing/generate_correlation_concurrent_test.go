package routing

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"sync"
	"testing"

	"github.com/cyw0ng95/v2e/cmd/v2broker/core"
)

func TestGenerateCorrelationID_Concurrent(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGenerateCorrelationID_Concurrent", nil, func(t *testing.T, tx *gorm.DB) {
		b := core.NewBroker()
		defer b.Shutdown()

		const N = 1000
		ids := make(chan string, N)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < N/10; j++ {
					ids <- b.GenerateCorrelationID()
				}
			}()
		}
		wg.Wait()
		close(ids)

		seen := make(map[string]struct{}, N)
		for id := range ids {
			if _, ok := seen[id]; ok {
				t.Fatalf("duplicate correlation id: %s", id)
			}
			seen[id] = struct{}{}
		}
	})

}

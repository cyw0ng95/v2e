package proc

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"fmt"
	"sync"
	"testing"

	"github.com/bytedance/sonic"
)

func TestUnmarshal_Concurrent(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestUnmarshal_Concurrent", nil, func(t *testing.T, tx *gorm.DB) {
		const goroutines = 50
		const perG = 50

		dataTemplate := `{"type":"request","id":"id-%d","payload":{"val":%d}}`

		errCh := make(chan error, goroutines)
		var wg sync.WaitGroup
		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				for i := 0; i < perG; i++ {
					d := []byte(fmt.Sprintf(dataTemplate, g*perG+i, i))
					// Use UnmarshalFast which uses sonic directly
					var raw Message
					if err := sonic.Unmarshal(d, &raw); err != nil {
						errCh <- fmt.Errorf("sonic unmarshal failed: %v", err)
						return
					}
					// basic sanity
					if raw.Type == "" || raw.ID == "" {
						errCh <- fmt.Errorf("unexpected empty fields: %+v", raw)
						return
					}
				}
			}(g)
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			if err != nil {
				t.Fatal(err)
			}
		}
	})

}

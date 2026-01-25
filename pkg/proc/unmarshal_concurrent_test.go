package proc

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bytedance/sonic"
)

func TestUnmarshal_Concurrent(t *testing.T) {
	const goroutines = 50
	const perG = 50

	dataTemplate := `{"type":"request","id":"id-%d","payload":{"val":%d}}`

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
					t.Fatalf("sonic unmarshal failed: %v", err)
				}
				// basic sanity
				if raw.Type == "" || raw.ID == "" {
					t.Fatalf("unexpected empty fields: %+v", raw)
				}
			}
		}(g)
	}
	wg.Wait()
}

package core

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"io"
	"os"
	"sync"
)

// InsertFakeProcess inserts a fake Process into the broker for testing.
// It does not start any OS-level process; it merely populates the broker map.
func InsertFakeProcess(b *Broker, id string, _ io.WriteCloser, _ io.ReadCloser, status ProcessStatus) *Process {
	p := NewTestProcess(id, status)
	b.InsertProcessForTest(p)
	return p
}

// ClosePipe closes both ends of an os.Pipe created by os.Pipe().
func ClosePipe(r *os.File, w *os.File) {
	var once sync.Once
	closeBoth := func() {
		if r != nil {
			r.Close()
		}
		if w != nil {
			w.Close()
		}
	}
	once.Do(closeBoth)
}

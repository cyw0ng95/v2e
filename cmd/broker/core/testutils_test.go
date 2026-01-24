package core

import (
	"io"
	"os"
	"sync"
)

// InsertFakeProcess inserts a fake Process into the broker for testing.
// It does not start any OS-level process; it merely populates the broker map.
func InsertFakeProcess(b *Broker, id string, stdin io.WriteCloser, stdout io.ReadCloser, status ProcessStatus) *Process {
	p := NewTestProcess(id, status, stdin, stdout)
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

package main

import (
    "io"
    "os"
    "sync"
)

// InsertFakeProcess inserts a fake Process into the broker for testing.
// It does not start any OS-level process; it merely populates the broker map.
func InsertFakeProcess(b *Broker, id string, stdin io.WriteCloser, stdout io.ReadCloser, status ProcessStatus) {
    p := &Process{
        info: &ProcessInfo{
            ID:     id,
            Status: status,
        },
        stdin:  stdin,
        stdout: stdout,
        done:   make(chan struct{}),
    }

    // Ensure the processes map is safe to write
    b.mu.Lock()
    b.processes[id] = p
    b.mu.Unlock()
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

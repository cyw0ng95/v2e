package main

import (
    "bufio"
    "io"
    "io/ioutil"
    "os"
    "testing"
    "time"

    "github.com/cyw0ng95/v2e/pkg/proc"
)

// BenchmarkSendMessage measures lightweight broker SendMessage path.
func BenchmarkSendMessage(b *testing.B) {
    br := NewBroker()
    b.ReportAllocs()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        m, _ := proc.NewEventMessage("evt", map[string]interface{}{"n": i})
        if err := br.SendMessage(m); err != nil {
            b.Fatalf("SendMessage error: %v", err)
        }
    }
}

// BenchmarkSendToProcess measures sending messages to a process stdin (pipe drained).
func BenchmarkSendToProcess(b *testing.B) {
    br := NewBroker()

    // Create pipe; broker writes to pw, we drain pr to avoid blocking.
    pr, pw, err := os.Pipe()
    if err != nil {
        b.Fatalf("pipe: %v", err)
    }
    defer pr.Close()
    defer pw.Close()

    InsertFakeProcess(br, "bench-proc", pw, nil, ProcessStatusRunning)

    // Drain reader
    done := make(chan struct{})
    go func() {
        io.Copy(ioutil.Discard, pr)
        close(done)
    }()

    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        m, _ := proc.NewEventMessage("evt", map[string]interface{}{"n": i})
        if err := br.SendToProcess("bench-proc", m); err != nil {
            b.Fatalf("SendToProcess error: %v", err)
        }
    }

    // Close writer to stop drain goroutine gracefully
    pw.Close()
    <-done
}

// BenchmarkInvokeRPC measures a full request/response roundtrip using pipes and a responder goroutine.
func BenchmarkInvokeRPC(b *testing.B) {
    br := NewBroker()

    // Create pipe: br writes to pw; test reads pr and responds
    pr, pw, err := os.Pipe()
    if err != nil {
        b.Fatalf("pipe: %v", err)
    }
    defer pr.Close()
    defer pw.Close()

    InsertFakeProcess(br, "rpc-proc", pw, nil, ProcessStatusRunning)

    // Responder goroutine: read lines, unmarshal and route response
    stop := make(chan struct{})
    go func() {
        rdr := bufio.NewReader(pr)
        for {
            line, err := rdr.ReadBytes('\n')
            if err != nil {
                select {
                case <-stop:
                    return
                default:
                    return
                }
            }
            // Try unmarshal; ignore errors in benchmark responder
            msg, err := proc.Unmarshal(line[:len(line)-1])
            if err != nil {
                continue
            }
            resp, _ := proc.NewResponseMessage(msg.ID, map[string]interface{}{"ok": true})
            resp.CorrelationID = msg.CorrelationID
            resp.Source = "rpc-proc"
            resp.Target = msg.Source
            _ = br.RouteMessage(resp, "rpc-proc")
        }
    }()

    b.ReportAllocs()
    b.ResetTimer()
    timeout := 2 * time.Second
    for i := 0; i < b.N; i++ {
        // small payload
        resp, err := br.InvokeRPC("bench-src", "rpc-proc", "Method", map[string]interface{}{"i": i}, timeout)
        if err != nil {
            b.Fatalf("InvokeRPC failed: %v", err)
        }
        if resp == nil {
            b.Fatalf("InvokeRPC returned nil response")
        }
    }

    // stop responder
    close(stop)
}

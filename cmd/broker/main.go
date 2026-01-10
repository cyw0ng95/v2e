package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

func main() {
	// Parse command line flags
	command := flag.String("cmd", "", "Command to execute")
	processID := flag.String("id", "", "Process ID")
	flag.Parse()

	// Set up logger
	common.SetLevel(common.InfoLevel)
	logger := common.NewLogger(os.Stdout, "", common.InfoLevel)

	// Create broker
	broker := proc.NewBroker()
	broker.SetLogger(logger)
	defer broker.Shutdown()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// If no command specified, run demo mode
	if *command == "" {
		runDemo(broker, logger, sigChan)
		return
	}

	// Parse command and args
	parts := strings.Fields(*command)
	if len(parts) == 0 {
		fmt.Println("Error: empty command")
		os.Exit(1)
	}

	cmd := parts[0]
	args := parts[1:]

	// Use provided ID or generate one
	id := *processID
	if id == "" {
		id = fmt.Sprintf("proc-%d", time.Now().Unix())
	}

	// Spawn the process
	common.Info("Spawning process: %s", *command)
	info, err := broker.Spawn(id, cmd, args...)
	if err != nil {
		common.Error("Failed to spawn process: %v", err)
		os.Exit(1)
	}

	common.Info("Process spawned: ID=%s PID=%d", info.ID, info.PID)

	// Monitor for process exit events
	go monitorMessages(broker, logger)

	// Wait for signal or process completion
	select {
	case <-sigChan:
		common.Info("Received shutdown signal")
	case <-time.After(30 * time.Second):
		common.Info("Timeout waiting for process")
	}
}

func runDemo(broker *proc.Broker, logger *common.Logger, sigChan chan os.Signal) {
	common.Info("Running broker demo...")
	common.Info("This demo will spawn multiple processes and monitor their lifecycle")

	// Spawn multiple demo processes
	processes := []struct {
		id   string
		cmd  string
		args []string
	}{
		{"echo-1", "echo", []string{"Hello from process 1"}},
		{"echo-2", "echo", []string{"Hello from process 2"}},
		{"sleep-1", "sleep", []string{"2"}},
	}

	for _, p := range processes {
		info, err := broker.Spawn(p.id, p.cmd, p.args...)
		if err != nil {
			common.Error("Failed to spawn %s: %v", p.id, err)
			continue
		}
		common.Info("Spawned process: ID=%s PID=%d Command=%s",
			info.ID, info.PID, info.Command)
	}

	// Monitor messages in background
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		for {
			msg, err := broker.ReceiveMessage(ctx)
			if err != nil {
				return
			}

			if msg.Type == proc.MessageTypeEvent {
				var event map[string]interface{}
				if err := msg.UnmarshalPayload(&event); err == nil {
					if event["event"] == "process_exited" {
						common.Info("Process exited: ID=%s PID=%.0f ExitCode=%.0f",
							event["id"], event["pid"], event["exit_code"])
					}
				}
			}
		}
	}()

	// Wait for all processes to complete
	time.Sleep(5 * time.Second)

	// List all processes
	common.Info("Process summary:")
	processes_list := broker.ListProcesses()
	for _, p := range processes_list {
		common.Info("  - ID=%s PID=%d Status=%s ExitCode=%d",
			p.ID, p.PID, p.Status, p.ExitCode)
	}

	common.Info("Demo complete")
}

func monitorMessages(broker *proc.Broker, logger *common.Logger) {
	ctx := context.Background()
	for {
		msg, err := broker.ReceiveMessage(ctx)
		if err != nil {
			return
		}

		switch msg.Type {
		case proc.MessageTypeEvent:
			var event map[string]interface{}
			if err := msg.UnmarshalPayload(&event); err == nil {
				if event["event"] == "process_exited" {
					common.Info("Process exited: ID=%s PID=%.0f ExitCode=%.0f",
						event["id"], event["pid"], event["exit_code"])
				}
			}
		case proc.MessageTypeError:
			common.Error("Error message: %s", msg.Error)
		}
	}
}

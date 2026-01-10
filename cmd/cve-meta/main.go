package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

func main() {
	// Parse command line flags
	cveID := flag.String("cve-id", "CVE-2021-44228", "CVE ID to fetch and store")
	dbPath := flag.String("db", "cve.db", "Path to CVE database")
	flag.Parse()

	// Set up logger
	common.SetLevel(common.InfoLevel)
	logger := common.NewLogger(os.Stdout, "[DEMO] ", common.InfoLevel)

	// Create broker
	broker := proc.NewBroker()
	broker.SetLogger(logger)
	defer broker.Shutdown()

	// Get the path to the built binaries
	exePath, err := os.Executable()
	if err != nil {
		common.Error("Failed to get executable path: %v", err)
		os.Exit(1)
	}
	baseDir := filepath.Dir(exePath)

	// Spawn CVE remote service
	common.Info("Starting CVE remote service...")
	cveRemotePath := filepath.Join(baseDir, "cve-remote")
	if _, err := os.Stat(cveRemotePath); os.IsNotExist(err) {
		cveRemotePath = "go"
		_, err = broker.SpawnRPC("cve-remote", cveRemotePath, "run", "./cmd/cve-remote")
	} else {
		_, err = broker.SpawnRPC("cve-remote", cveRemotePath)
	}
	if err != nil {
		common.Error("Failed to spawn cve-remote: %v", err)
		os.Exit(1)
	}

	// Spawn CVE local service
	common.Info("Starting CVE local service...")
	os.Setenv("CVE_DB_PATH", *dbPath)
	cveLocalPath := filepath.Join(baseDir, "cve-local")
	if _, err := os.Stat(cveLocalPath); os.IsNotExist(err) {
		cveLocalPath = "go"
		_, err = broker.SpawnRPC("cve-local", cveLocalPath, "run", "./cmd/cve-local")
	} else {
		_, err = broker.SpawnRPC("cve-local", cveLocalPath)
	}
	if err != nil {
		common.Error("Failed to spawn cve-local: %v", err)
		os.Exit(1)
	}

	// Wait for services to be ready
	time.Sleep(2 * time.Second)

	// Step 1: Check if CVE is already stored locally
	common.Info("Checking if %s is stored locally...", *cveID)
	checkMsg, _ := proc.NewRequestMessage("RPCIsCVEStoredByID", map[string]string{
		"cve_id": *cveID,
	})
	if err := broker.SendToProcess("cve-local", checkMsg); err != nil {
		common.Error("Failed to send check message: %v", err)
		os.Exit(1)
	}

	// Wait for response
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stored bool
	for {
		msg, err := broker.ReceiveMessage(ctx)
		if err != nil {
			common.Error("Failed to receive message: %v", err)
			os.Exit(1)
		}

		if msg.Type == proc.MessageTypeResponse && msg.ID == "RPCIsCVEStoredByID" {
			var resp map[string]interface{}
			if err := msg.UnmarshalPayload(&resp); err == nil {
				stored = resp["stored"].(bool)
				common.Info("CVE %s is stored: %v", *cveID, stored)
			}
			break
		}
	}

	// Step 2: If not stored, fetch from remote and save locally
	if !stored {
		common.Info("Fetching %s from NVD API...", *cveID)
		fetchMsg, _ := proc.NewRequestMessage("RPCGetCVEByID", map[string]string{
			"cve_id": *cveID,
		})
		if err := broker.SendToProcess("cve-remote", fetchMsg); err != nil {
			common.Error("Failed to send fetch message: %v", err)
			os.Exit(1)
		}

		// Wait for fetch response
		ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel2()

		for {
			msg, err := broker.ReceiveMessage(ctx2)
			if err != nil {
				common.Error("Failed to receive fetch response: %v", err)
				os.Exit(1)
			}

			if msg.Type == proc.MessageTypeResponse && msg.ID == "RPCGetCVEByID" {
				var resp map[string]interface{}
				if err := msg.UnmarshalPayload(&resp); err == nil {
					common.Info("Successfully fetched CVE data")

					// Extract the CVE item from the response
					vulns := resp["vulnerabilities"].([]interface{})
					if len(vulns) > 0 {
						vuln := vulns[0].(map[string]interface{})
						cveData := vuln["cve"]

						// Save to local database
						common.Info("Saving CVE to local database...")
						saveMsg, _ := proc.NewRequestMessage("RPCSaveCVEByID", map[string]interface{}{
							"cve": cveData,
						})
						if err := broker.SendToProcess("cve-local", saveMsg); err != nil {
							common.Error("Failed to send save message: %v", err)
							os.Exit(1)
						}

						// Wait for save response
						ctx3, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel3()

						for {
							msg, err := broker.ReceiveMessage(ctx3)
							if err != nil {
								common.Error("Failed to receive save response: %v", err)
								os.Exit(1)
							}

							if msg.Type == proc.MessageTypeResponse && msg.ID == "RPCSaveCVEByID" {
								var resp map[string]interface{}
								if err := msg.UnmarshalPayload(&resp); err == nil {
									if resp["success"].(bool) {
										common.Info("Successfully saved CVE %s to local database", resp["cve_id"])
									}
								}
								break
							}
						}
					}
				}
				break
			}
		}
	}

	// Step 3: Get CVE count from remote
	common.Info("Getting total CVE count from NVD...")
	cntMsg, _ := proc.NewRequestMessage("RPCGetCVECnt", map[string]interface{}{})
	if err := broker.SendToProcess("cve-remote", cntMsg); err != nil {
		common.Error("Failed to send count message: %v", err)
		os.Exit(1)
	}

	// Wait for count response
	ctx4, cancel4 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel4()

	for {
		msg, err := broker.ReceiveMessage(ctx4)
		if err != nil {
			common.Error("Failed to receive count response: %v", err)
			os.Exit(1)
		}

		if msg.Type == proc.MessageTypeResponse && msg.ID == "RPCGetCVECnt" {
			var resp map[string]interface{}
			if err := msg.UnmarshalPayload(&resp); err == nil {
				totalResults := int(resp["total_results"].(float64))
				common.Info("Total CVEs in NVD database: %d", totalResults)
			}
			break
		}

		// Also handle event messages
		if msg.Type == proc.MessageTypeEvent {
			var event map[string]interface{}
			if err := msg.UnmarshalPayload(&event); err == nil {
				eventType := event["event"]
				if eventType != nil {
					common.Debug("Event: %v", eventType)
				}
			}
		}
	}

	common.Info("Demo complete!")
}

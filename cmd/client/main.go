package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cyw0ng95/v2e/pkg/common"
)

func main() {
	fmt.Println("Starting client...")
	fmt.Printf("Version: %s\n", common.Version)

	// Load configuration
	config, err := common.LoadConfig("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Determine URL to connect to
	// Priority: 1. Command line arg, 2. Config file, 3. Default
	url := "http://localhost:8080"
	if config.Client.URL != "" {
		url = config.Client.URL
	}
	if len(os.Args) > 1 {
		url = os.Args[1]
	}

	fmt.Printf("Connecting to %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n", string(body))
}

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

	url := "http://localhost:8080"
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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/repo"
)

func main() {
	fmt.Println("Starting server...")
	fmt.Printf("Version: %s\n", common.Version)

	// Load configuration
	config, err := common.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use config or default values
	addr := ":8080"
	if config.Server.Address != "" {
		addr = config.Server.Address
	}

	// Initialize CVE fetcher
	cveFetcher := repo.NewCVEFetcher("")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from v2e server! Version: %s\n", common.Version)
	})

	// Add CVE endpoint
	http.HandleFunc("/cve/", func(w http.ResponseWriter, r *http.Request) {
		// Extract CVE ID from URL path
		cveID := r.URL.Path[len("/cve/"):]
		if cveID == "" {
			http.Error(w, "CVE ID is required", http.StatusBadRequest)
			return
		}

		// Fetch CVE data
		result, err := cveFetcher.FetchCVEByID(cveID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch CVE: %v", err), http.StatusInternalServerError)
			return
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	fmt.Printf("Server listening on %s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Println("  GET / - Server info")
	fmt.Println("  GET /cve/{cve-id} - Fetch CVE data from NVD")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

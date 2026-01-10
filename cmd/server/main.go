package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cyw0ng95/v2e/pkg/common"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from v2e server! Version: %s\n", common.Version)
	})

	fmt.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

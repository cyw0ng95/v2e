package main

import (
	"fmt"
	"net/http"

	"github.com/cyw0ng95/v2e/pkg/common"
)

func main() {
	common.Info("Starting server...")
	common.Info("Version: %s", common.Version)

	// Load configuration
	config, err := common.LoadConfig("")
	if err != nil {
		common.Fatal("Failed to load config: %v", err)
	}

	// Use config or default values
	addr := ":8080"
	if config.Server.Address != "" {
		addr = config.Server.Address
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		common.Info("Request received from %s", r.RemoteAddr)
		fmt.Fprintf(w, "Hello from v2e server! Version: %s\n", common.Version)
	})

	common.Info("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		common.Fatal("Server error: %v", err)
	}
}

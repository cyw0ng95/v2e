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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from v2e server! Version: %s\n", common.Version)
	})

	addr := ":8080"
	fmt.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

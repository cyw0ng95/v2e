package main

import (
	"log"

	"github.com/cyw0ng95/v2e/pkg/notes"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Notes Local Server...")

	// Initialize database
	db, err := gorm.Open(sqlite.Open("notes.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Migrate tables
	if err := notes.MigrateNotesTables(db); err != nil {
		log.Fatal("Failed to migrate notes tables:", err)
	}

	// Initialize service container
	container := notes.NewServiceContainer(db)

	// Initialize subprocess
	sp := subprocess.New("local")

	// Register RPC handlers
	_ = notes.NewRPCHandlers(container, sp)

	// Start subprocess - Run() blocks until the subprocess is stopped
	if err := sp.Run(); err != nil {
		log.Fatal("Failed to run subprocess:", err)
	}

	log.Println("Notes Local Server started successfully")
	log.Println("Waiting for RPC requests...")

	// Keep the server running
	select {}
}

package main

import (
	"log"
	"os"

	"github.com/Geipman/faktura/internal/db"
	"github.com/Geipman/faktura/internal/server"
)

func main() {
	// Simple config setup
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "faktura.db"
	}

	log.Printf("Starting Faktura application (DB: %s, Port: %s)", dbPath, port)

	// Connect to SQLite
	database, err := db.Connect(dbPath)
	if err != nil {
		log.Fatalf("Fatal error connecting to database: %v", err)
	}
	defer database.Close()

	// Perform basic schema initialization check
	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS system_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatalf("Fatal error initializing database schema: %v", err)
	}

	// Insert startup log
	_, err = database.Exec("INSERT INTO system_logs (message) VALUES ('System started successfully');")
	if err != nil {
		log.Printf("Warning: failed to write startup log: %v", err)
	}

	// Start server
	srv := server.NewServer(":"+port, database)
	if err := srv.Start(); err != nil {
		log.Fatalf("Fatal server error: %v", err)
	}
}

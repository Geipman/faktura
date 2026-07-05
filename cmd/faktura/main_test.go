package main

import (
	"testing"

	"github.com/Geipman/faktura/internal/db"
)

// TestDBConnection tests that SQLite connects correctly using the pure Go driver in-memory.
func TestDBConnection(t *testing.T) {
	// Connect to in-memory database
	database, err := db.Connect(":memory:")
	if err != nil {
		t.Fatalf("Expected no error connecting to memory DB, got: %v", err)
	}
	defer database.Close()

	// Verify we can execute a query
	var val int
	err = database.QueryRow("SELECT 1").Scan(&val)
	if err != nil {
		t.Fatalf("Expected query to execute, got error: %v", err)
	}

	if val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}
}

package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// Connect initializes the database connection and pings the SQLite file.
func Connect(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable WAL mode for better concurrency and foreign keys support
	_, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set database pragmas: %w", err)
	}

	return db, nil
}

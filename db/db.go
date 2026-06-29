// Package db manages the SQLite puzzle database for storing, deduplicating,
// and querying Sudoku puzzles by difficulty.
package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps a SQLite connection for puzzle storage.
type DB struct {
	conn *sql.DB
}

// Open opens (or creates) a SQLite database at the given path and runs
// schema migrations. Use ":memory:" for an in-memory database.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// migrate creates the puzzles table if it does not exist.
func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS puzzles (
			puzzle        TEXT PRIMARY KEY,
			difficulty    TEXT NOT NULL,
			score         INTEGER NOT NULL,
			max_technique TEXT NOT NULL,
			source        TEXT,
			created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

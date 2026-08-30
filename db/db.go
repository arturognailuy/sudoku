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
	if _, err := conn.Exec("PRAGMA busy_timeout=5000"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set busy timeout: %w", err)
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

// migrate creates the puzzles table and applies additive schema changes.
func (db *DB) migrate() error {
	if _, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS puzzles (
			puzzle         TEXT PRIMARY KEY,
			difficulty     TEXT NOT NULL,
			score          INTEGER NOT NULL,
			max_technique  TEXT NOT NULL,
			source         TEXT,
			play_count     INTEGER NOT NULL DEFAULT 0,
			last_played_at TIMESTAMP,
			completion_count INTEGER NOT NULL DEFAULT 0,
			last_completed_at TIMESTAMP,
			created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return err
	}

	columns, err := db.tableColumns("puzzles")
	if err != nil {
		return err
	}
	if !columns["play_count"] {
		if _, err := db.conn.Exec(`ALTER TABLE puzzles ADD COLUMN play_count INTEGER NOT NULL DEFAULT 0`); err != nil {
			return fmt.Errorf("add play_count: %w", err)
		}
	}
	if !columns["last_played_at"] {
		if _, err := db.conn.Exec(`ALTER TABLE puzzles ADD COLUMN last_played_at TIMESTAMP`); err != nil {
			return fmt.Errorf("add last_played_at: %w", err)
		}
	}
	if !columns["completion_count"] {
		if _, err := db.conn.Exec(`ALTER TABLE puzzles ADD COLUMN completion_count INTEGER NOT NULL DEFAULT 0`); err != nil {
			return fmt.Errorf("add completion_count: %w", err)
		}
	}
	if !columns["last_completed_at"] {
		if _, err := db.conn.Exec(`ALTER TABLE puzzles ADD COLUMN last_completed_at TIMESTAMP`); err != nil {
			return fmt.Errorf("add last_completed_at: %w", err)
		}
	}
	_, err = db.conn.Exec(`CREATE INDEX IF NOT EXISTS puzzles_acquisition_idx ON puzzles (difficulty, play_count, last_played_at)`)
	return err
}

func (db *DB) tableColumns(table string) (map[string]bool, error) {
	rows, err := db.conn.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return nil, fmt.Errorf("inspect %s columns: %w", table, err)
	}
	defer rows.Close()
	columns := make(map[string]bool)
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return nil, fmt.Errorf("scan %s columns: %w", table, err)
		}
		columns[name] = true
	}
	return columns, rows.Err()
}

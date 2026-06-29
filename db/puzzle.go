package db

import (
	"database/sql"
	"fmt"
)

// Puzzle represents a stored puzzle record.
type Puzzle struct {
	Puzzle       string // Normalized 81-char puzzle string.
	Difficulty   string // Difficulty level name (easy/medium/hard/expert/evil).
	Score        int    // Total difficulty score.
	MaxTechnique string // Highest-tier technique required.
	Source       string // Origin: "generated", "imported", or source name.
}

// InsertPuzzle stores a puzzle if it does not already exist.
// Returns true if the puzzle was inserted (new), false if it was a duplicate.
func (db *DB) InsertPuzzle(p Puzzle) (bool, error) {
	result, err := db.conn.Exec(
		`INSERT OR IGNORE INTO puzzles (puzzle, difficulty, score, max_technique, source)
		 VALUES (?, ?, ?, ?, ?)`,
		p.Puzzle, p.Difficulty, p.Score, p.MaxTechnique, p.Source,
	)
	if err != nil {
		return false, fmt.Errorf("insert puzzle: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}

	return rows > 0, nil
}

// GetRandom returns a random puzzle at the specified difficulty level,
// or nil if none exists.
func (db *DB) GetRandom(difficulty string) (*Puzzle, error) {
	row := db.conn.QueryRow(
		`SELECT puzzle, difficulty, score, max_technique, COALESCE(source, '')
		 FROM puzzles
		 WHERE difficulty = ?
		 ORDER BY RANDOM()
		 LIMIT 1`,
		difficulty,
	)

	var p Puzzle
	err := row.Scan(&p.Puzzle, &p.Difficulty, &p.Score, &p.MaxTechnique, &p.Source)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get random puzzle: %w", err)
	}

	return &p, nil
}

// Stats holds per-difficulty puzzle counts.
type Stats struct {
	Total      int
	ByLevel    map[string]int
}

// GetStats returns the total puzzle count and per-difficulty breakdown.
func (db *DB) GetStats() (*Stats, error) {
	rows, err := db.conn.Query(
		`SELECT difficulty, COUNT(*) FROM puzzles GROUP BY difficulty`,
	)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	defer rows.Close()

	stats := &Stats{ByLevel: make(map[string]int)}
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, fmt.Errorf("scan stats: %w", err)
		}
		stats.ByLevel[level] = count
		stats.Total += count
	}

	return stats, rows.Err()
}

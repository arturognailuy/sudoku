package db

import (
	"database/sql"
	"path/filepath"
	"sync"
	"testing"

	_ "modernc.org/sqlite"
)

func TestOpenClose(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := d.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestInsertAndGet(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	p := Puzzle{
		Puzzle:       "123456789456789123789123456214365897365897214897214365531642978642978531978531642",
		Difficulty:   "easy",
		Score:        100,
		MaxTechnique: "hidden-single",
		Source:       "generated",
	}

	// First insert should succeed.
	inserted, err := d.InsertPuzzle(p)
	if err != nil {
		t.Fatalf("InsertPuzzle: %v", err)
	}
	if !inserted {
		t.Fatal("expected puzzle to be inserted")
	}

	// Duplicate insert should return false.
	inserted, err = d.InsertPuzzle(p)
	if err != nil {
		t.Fatalf("InsertPuzzle duplicate: %v", err)
	}
	if inserted {
		t.Fatal("expected duplicate to not be inserted")
	}
}

func TestGetRandomEmpty(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	got, err := d.GetRandom("hard")
	if err != nil {
		t.Fatalf("GetRandom: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for empty database")
	}
}

func TestGetRandom(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	p := Puzzle{
		Puzzle:       "123456789456789123789123456214365897365897214897214365531642978642978531978531642",
		Difficulty:   "medium",
		Score:        200,
		MaxTechnique: "naked-pair",
		Source:       "test",
	}
	if _, err := d.InsertPuzzle(p); err != nil {
		t.Fatalf("InsertPuzzle: %v", err)
	}

	// Should find it at medium level.
	got, err := d.GetRandom("medium")
	if err != nil {
		t.Fatalf("GetRandom: %v", err)
	}
	if got == nil {
		t.Fatal("expected a puzzle")
	}
	if got.Puzzle != p.Puzzle {
		t.Fatalf("expected %s, got %s", p.Puzzle, got.Puzzle)
	}

	// Should not find at hard level.
	got, err = d.GetRandom("hard")
	if err != nil {
		t.Fatalf("GetRandom: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for non-matching difficulty")
	}
}

func TestGetStats(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	puzzles := []Puzzle{
		{Puzzle: "p1.......................................................................", Difficulty: "easy", Score: 50, MaxTechnique: "naked-single", Source: "test"},
		{Puzzle: "p2.......................................................................", Difficulty: "easy", Score: 60, MaxTechnique: "hidden-single", Source: "test"},
		{Puzzle: "p3.......................................................................", Difficulty: "hard", Score: 300, MaxTechnique: "x-wing", Source: "test"},
	}
	for _, p := range puzzles {
		if _, err := d.InsertPuzzle(p); err != nil {
			t.Fatalf("InsertPuzzle: %v", err)
		}
	}

	stats, err := d.GetStats()
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.Total != 3 {
		t.Fatalf("expected 3 total, got %d", stats.Total)
	}
	if stats.ByLevel["easy"] != 2 {
		t.Fatalf("expected 2 easy, got %d", stats.ByLevel["easy"])
	}
	if stats.ByLevel["hard"] != 1 {
		t.Fatalf("expected 1 hard, got %d", stats.ByLevel["hard"])
	}
}

func TestAcquireForPlayExhaustsPoolBeforeBalancedReuse(t *testing.T) {
	d, err := Open(filepath.Join(t.TempDir(), "acquire.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	for _, puzzle := range []string{"puzzle-a", "puzzle-b"} {
		if _, err := d.InsertPuzzle(Puzzle{Puzzle: puzzle, Difficulty: "easy", Score: 1, MaxTechnique: "naked-single", Source: "test"}); err != nil {
			t.Fatalf("InsertPuzzle: %v", err)
		}
	}
	first, err := d.AcquireForPlay("easy")
	if err != nil || first == nil || first.PlayCount != 1 || first.LastPlayedAt == "" {
		t.Fatalf("first acquisition = %+v, %v", first, err)
	}
	second, err := d.AcquireForPlay("easy")
	if err != nil || second == nil || second.Puzzle == first.Puzzle || second.PlayCount != 1 {
		t.Fatalf("second acquisition = %+v, %v; first = %+v", second, err, first)
	}
	third, err := d.AcquireForPlay("easy")
	if err != nil || third == nil || third.PlayCount != 2 {
		t.Fatalf("third acquisition = %+v, %v", third, err)
	}
	var minimum, maximum int
	if err := d.conn.QueryRow(`SELECT MIN(play_count), MAX(play_count) FROM puzzles WHERE difficulty = 'easy'`).Scan(&minimum, &maximum); err != nil {
		t.Fatalf("query counts: %v", err)
	}
	if minimum != 1 || maximum != 2 {
		t.Fatalf("play counts = %d..%d, want 1..2", minimum, maximum)
	}
	missing, err := d.AcquireForPlay("hard")
	if err != nil || missing != nil {
		t.Fatalf("missing acquisition = %+v, %v", missing, err)
	}
}

func TestOpenMigratesExistingPuzzleTable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	_, err = conn.Exec(`CREATE TABLE puzzles (
		puzzle TEXT PRIMARY KEY, difficulty TEXT NOT NULL, score INTEGER NOT NULL,
		max_technique TEXT NOT NULL, source TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err == nil {
		_, err = conn.Exec(`INSERT INTO puzzles (puzzle, difficulty, score, max_technique, source) VALUES ('legacy', 'easy', 1, 'naked-single', 'legacy')`)
	}
	conn.Close()
	if err != nil {
		t.Fatalf("prepare legacy database: %v", err)
	}

	d, err := Open(path)
	if err != nil {
		t.Fatalf("migrate legacy database: %v", err)
	}
	defer d.Close()
	got, err := d.AcquireForPlay("easy")
	if err != nil || got == nil || got.Puzzle != "legacy" || got.PlayCount != 1 {
		t.Fatalf("migrated acquisition = %+v, %v", got, err)
	}
}

func TestConcurrentAcquireForPlayReturnsDistinctUnplayedRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "concurrent.db")
	seed, err := Open(path)
	if err != nil {
		t.Fatalf("Open seed database: %v", err)
	}
	for _, puzzle := range []string{"puzzle-a", "puzzle-b"} {
		if _, err := seed.InsertPuzzle(Puzzle{Puzzle: puzzle, Difficulty: "easy", Score: 1, MaxTechnique: "naked-single", Source: "test"}); err != nil {
			t.Fatalf("InsertPuzzle: %v", err)
		}
	}
	seed.Close()

	databases := make([]*DB, 2)
	for index := range databases {
		databases[index], err = Open(path)
		if err != nil {
			t.Fatalf("Open database %d: %v", index, err)
		}
		defer databases[index].Close()
	}

	start := make(chan struct{})
	results := make(chan *Puzzle, 2)
	errors := make(chan error, 2)
	var workers sync.WaitGroup
	for _, database := range databases {
		workers.Add(1)
		go func(database *DB) {
			defer workers.Done()
			<-start
			puzzle, acquireErr := database.AcquireForPlay("easy")
			results <- puzzle
			errors <- acquireErr
		}(database)
	}
	close(start)
	workers.Wait()
	close(results)
	close(errors)

	for acquireErr := range errors {
		if acquireErr != nil {
			t.Fatalf("AcquireForPlay: %v", acquireErr)
		}
	}
	seen := make(map[string]bool)
	for puzzle := range results {
		if puzzle == nil {
			t.Fatal("AcquireForPlay returned nil")
		}
		seen[puzzle.Puzzle] = true
	}
	if len(seen) != 2 {
		t.Fatalf("concurrent acquisitions returned %v, want two distinct puzzles", seen)
	}
}

func TestMarkForPlayUpdatesOnlyTheSelectedPuzzle(t *testing.T) {
	d, err := Open(filepath.Join(t.TempDir(), "mark.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()
	for _, puzzle := range []string{"selected", "unselected"} {
		if _, err := d.InsertPuzzle(Puzzle{Puzzle: puzzle, Difficulty: "hard", Score: 1, MaxTechnique: "hidden-pair", Source: "test"}); err != nil {
			t.Fatalf("InsertPuzzle: %v", err)
		}
	}

	marked, err := d.MarkForPlay("selected")
	if err != nil || marked == nil || marked.Puzzle != "selected" || marked.PlayCount != 1 || marked.LastPlayedAt == "" {
		t.Fatalf("MarkForPlay = %+v, %v", marked, err)
	}
	missing, err := d.MarkForPlay("missing")
	if err != nil || missing != nil {
		t.Fatalf("missing MarkForPlay = %+v, %v", missing, err)
	}
	var count int
	if err := d.conn.QueryRow(`SELECT play_count FROM puzzles WHERE puzzle = 'unselected'`).Scan(&count); err != nil {
		t.Fatalf("query unselected puzzle: %v", err)
	}
	if count != 0 {
		t.Fatalf("unselected play count = %d, want 0", count)
	}
}

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

func TestCompletionStatisticsAndResetScopes(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "statistics.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	for _, puzzle := range []Puzzle{{Puzzle: "easy", Difficulty: "easy", Score: 1, MaxTechnique: "single"}, {Puzzle: "hard", Difficulty: "hard", Score: 2, MaxTechnique: "pair"}} {
		if _, err := database.InsertPuzzle(puzzle); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := database.MarkForPlay("easy"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.MarkForPlay("easy"); err != nil {
		t.Fatal(err)
	}
	if ok, err := database.RecordCompletion("easy"); err != nil || !ok {
		t.Fatalf("completion = %v, %v", ok, err)
	}
	if ok, err := database.RecordCompletion("missing"); err != nil || ok {
		t.Fatalf("missing completion = %v, %v", ok, err)
	}
	rows, err := database.PlayStatistics("")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 6 {
		t.Fatalf("statistics rows = %d, want five grades plus overall", len(rows))
	}
	if rows[0].Stored != 1 || rows[0].Acquisitions != 2 || rows[0].Completions != 1 || rows[5].Stored != 2 {
		t.Fatalf("statistics = %+v", rows)
	}
	preview, err := database.PreviewHistoryReset("easy")
	if err != nil || preview != (HistoryPreview{Rows: 1, Acquisitions: 2, Completions: 1}) {
		t.Fatalf("preview = %+v, %v", preview, err)
	}
	if err := database.ResetHistory("completion", "easy"); err != nil {
		t.Fatal(err)
	}
	preview, _ = database.PreviewHistoryReset("easy")
	if preview.Acquisitions != 2 || preview.Completions != 0 {
		t.Fatalf("completion reset changed wrong scope: %+v", preview)
	}
	if err := database.ResetHistory("all", "easy"); err != nil {
		t.Fatal(err)
	}
	preview, _ = database.PreviewHistoryReset("easy")
	if preview.Acquisitions != 0 || preview.Completions != 0 {
		t.Fatalf("all reset = %+v", preview)
	}
}

func TestCompletionMigrationPreservesAcquisitionHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pre-completion.db")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(`CREATE TABLE puzzles (puzzle TEXT PRIMARY KEY, difficulty TEXT NOT NULL, score INTEGER NOT NULL, max_technique TEXT NOT NULL, source TEXT, play_count INTEGER NOT NULL DEFAULT 0, last_played_at TIMESTAMP, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`)
	if err == nil {
		_, err = conn.Exec(`INSERT INTO puzzles (puzzle,difficulty,score,max_technique,play_count,last_played_at) VALUES ('legacy','easy',1,'single',3,CURRENT_TIMESTAMP)`)
	}
	conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	var plays, completions int
	var completed any
	if err := database.conn.QueryRow(`SELECT play_count, completion_count, last_completed_at FROM puzzles WHERE puzzle='legacy'`).Scan(&plays, &completions, &completed); err != nil {
		t.Fatal(err)
	}
	if plays != 3 || completions != 0 || completed != nil {
		t.Fatalf("migrated history = %d, %d, %v", plays, completions, completed)
	}
}

func TestConcurrentCompletionIncrements(t *testing.T) {
	path := filepath.Join(t.TempDir(), "completion.db")
	seed, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := seed.InsertPuzzle(Puzzle{Puzzle: "shared", Difficulty: "easy", Score: 1, MaxTechnique: "single"}); err != nil {
		t.Fatal(err)
	}
	seed.Close()
	const workers = 8
	start := make(chan struct{})
	errors := make(chan error, workers)
	var group sync.WaitGroup
	for i := 0; i < workers; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			database, err := Open(path)
			if err != nil {
				errors <- err
				return
			}
			defer database.Close()
			<-start
			_, err = database.RecordCompletion("shared")
			errors <- err
		}()
	}
	close(start)
	group.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	check, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer check.Close()
	var count int
	if err := check.conn.QueryRow(`SELECT completion_count FROM puzzles WHERE puzzle='shared'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != workers {
		t.Fatalf("completion_count = %d, want %d", count, workers)
	}
}

func TestResetHistoryRollsBackOnDatabaseError(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "rollback.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.InsertPuzzle(Puzzle{Puzzle: "protected", Difficulty: "easy", Score: 1, MaxTechnique: "single"}); err != nil {
		t.Fatal(err)
	}
	if _, err := database.conn.Exec(`UPDATE puzzles SET play_count=2, completion_count=3 WHERE puzzle='protected'`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.conn.Exec(`CREATE TRIGGER reject_history_reset BEFORE UPDATE ON puzzles BEGIN SELECT RAISE(ABORT, 'protected'); END`); err != nil {
		t.Fatal(err)
	}
	if err := database.ResetHistory("all", "easy"); err == nil {
		t.Fatal("expected reset failure")
	}
	var plays, completions int
	if err := database.conn.QueryRow(`SELECT play_count, completion_count FROM puzzles WHERE puzzle='protected'`).Scan(&plays, &completions); err != nil {
		t.Fatal(err)
	}
	if plays != 2 || completions != 3 {
		t.Fatalf("failed reset changed counters: %d %d", plays, completions)
	}
}

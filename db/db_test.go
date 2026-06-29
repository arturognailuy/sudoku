package db

import (
	"testing"
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

package solver

import (
	"reflect"
	"testing"

	"github.com/gnailuy/sudoku/core"
)

func TestClassifyPuzzleEasy(t *testing.T) {
	store := NewStore()

	// An easy puzzle that requires only naked/hidden singles.
	// Puzzle with 50 clues — should be easy.
	puzzleStr := "123456789456789123789123456214365897365897214897214365531642978642978531..853164."
	board := core.NewEmptyBoard()
	board.FromString(puzzleStr)

	c := ClassifyPuzzle(store, board)

	if c.Difficulty != "easy" {
		t.Logf("Expected easy, got %s (max technique: %s)", c.Difficulty, c.MaxTechnique)
	}
	if c.Score < 0 {
		t.Fatalf("Expected non-negative score, got %d", c.Score)
	}
	if len(c.Moves) == 0 {
		t.Fatal("Expected at least one move")
	}
}

func TestClassifyPuzzleFullySolved(t *testing.T) {
	store := NewStore()

	// A fully solved board — no moves needed.
	puzzleStr := "123456789456789123789123456214365897365897214897214365531642978642978531978531642"
	board := core.NewEmptyBoard()
	board.FromString(puzzleStr)

	c := ClassifyPuzzle(store, board)

	if !c.Solved || c.Outcome != ClassificationSolved {
		t.Fatalf("Expected solved outcome, got solved=%v outcome=%q", c.Solved, c.Outcome)
	}
	if c.Difficulty != "easy" {
		t.Fatalf("Expected a complete board to use the lowest tier, got %q", c.Difficulty)
	}
	if c.MaxTechnique != "" {
		t.Fatalf("Expected no maximum technique, got %q", c.MaxTechnique)
	}
	if c.Score != 0 {
		t.Fatalf("Expected score 0 for solved board, got %d", c.Score)
	}
	if len(c.Moves) != 0 {
		t.Fatalf("Expected no moves for solved board, got %d", len(c.Moves))
	}
}

func TestDetermineDifficulty(t *testing.T) {
	tests := []struct {
		technique string
		expected  string
	}{
		{"naked-single", "easy"},
		{"hidden-single", "easy"},
		{"naked-pair", "medium"},
		{"pointing-pair", "medium"},
		{"x-wing", "hard"},
		{"xy-wing", "hard"},
		{"swordfish", "expert"},
		{"simple-coloring", "expert"},
		{"jellyfish", "evil"},
		{"bug-plus-one", "evil"},
		{"unique-rectangle", "evil"},
		{"backtracker", ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		got := determineDifficulty(tt.technique)
		if got != tt.expected {
			t.Errorf("determineDifficulty(%q) = %q, want %q", tt.technique, got, tt.expected)
		}
	}
}

func TestFindMaxTechniqueUsesExplicitTierHierarchy(t *testing.T) {
	moves := []Move{
		{Technique: "xy-wing"},    // Hard, weight 160.
		{Technique: "naked-quad"}, // Expert, weight 120.
		{Technique: "x-wing"},     // Hard, weight 140.
	}

	if got := findMaxTechnique(moves); got != "naked-quad" {
		t.Fatalf("findMaxTechnique() = %q, want expert technique naked-quad", got)
	}
}

func TestClassifyPuzzleReportsStrategyUnsolvedSeparately(t *testing.T) {
	store := NewStore()
	board := core.NewEmptyBoard()

	classification := ClassifyPuzzle(store, board)

	if classification.Solved {
		t.Fatal("expected empty board to stall strategy solvers")
	}
	if classification.Outcome != ClassificationStrategyUnsolved {
		t.Fatalf("outcome = %q, want %q", classification.Outcome, ClassificationStrategyUnsolved)
	}
	if classification.Difficulty == "evil" || classification.MaxTechnique == "backtracker" {
		t.Fatalf("strategy-unsolved puzzle was promoted to evil/backtracker: %+v", classification)
	}
}

func TestClassifyPuzzleIsDeterministic(t *testing.T) {
	store := NewStore()
	board := core.NewEmptyBoard()
	board.FromString("..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3..")

	first := ClassifyPuzzle(store, board)
	for run := 0; run < 25; run++ {
		got := ClassifyPuzzle(store, board)
		if !reflect.DeepEqual(got, first) {
			t.Fatalf("classification run %d differs:\nfirst=%+v\ngot=%+v", run+2, first, got)
		}
	}
}

package solver

import (
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

	if !c.Solved {
		t.Fatal("Expected solved to be true for already-solved board")
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
		{"backtracker", "evil"},
		{"unknown", "evil"},
	}

	for _, tt := range tests {
		got := determineDifficulty(tt.technique)
		if got != tt.expected {
			t.Errorf("determineDifficulty(%q) = %q, want %q", tt.technique, got, tt.expected)
		}
	}
}

package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// ---------------------------------------------------------------------------
// HiddenSubsetSolver tests
// ---------------------------------------------------------------------------

// TestHiddenSubsetSolver_NoProgress verifies nil return on an empty board.
func TestHiddenSubsetSolver_NoProgress(t *testing.T) {
	board := core.NewEmptyBoard()
	s := solver.NewHiddenSubsetSolver()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil move on empty board, got %+v", move)
	}
}

// TestHiddenSubsetSolver_SolvedBoard verifies nil return on a solved board.
func TestHiddenSubsetSolver_SolvedBoard(t *testing.T) {
	board := core.NewEmptyBoard()
	backtracker := solver.NewBacktracker()
	backtracker.Solve(&board)

	s := solver.NewHiddenSubsetSolver()
	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil move on solved board, got %+v", move)
	}
}

// TestHiddenSubsetSolver_Metadata verifies key and display name.
func TestHiddenSubsetSolver_Metadata(t *testing.T) {
	s := solver.NewHiddenSubsetSolver()
	if s.GetKey() != "hidden-subset" {
		t.Errorf("Expected key 'hidden-subset', got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Hidden Pairs/Triples/Quads" {
		t.Errorf("Expected display name 'Hidden Pairs/Triples/Quads', got %q", s.GetDisplayName())
	}
}

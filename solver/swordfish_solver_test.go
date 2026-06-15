package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// ---------------------------------------------------------------------------
// SwordfishSolver tests
// ---------------------------------------------------------------------------

// TestSwordfishSolver_NoProgress verifies nil return on an empty board.
func TestSwordfishSolver_NoProgress(t *testing.T) {
	board := core.NewEmptyBoard()
	s := solver.NewSwordfishSolver()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil move on empty board, got %+v", move)
	}
}

// TestSwordfishSolver_SolvedBoard verifies nil return on a solved board.
func TestSwordfishSolver_SolvedBoard(t *testing.T) {
	board := core.NewEmptyBoard()
	backtracker := solver.NewBacktracker()
	backtracker.Solve(&board)

	s := solver.NewSwordfishSolver()
	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil move on solved board, got %+v", move)
	}
}

// TestSwordfishSolver_Metadata verifies key and display name.
func TestSwordfishSolver_Metadata(t *testing.T) {
	s := solver.NewSwordfishSolver()
	if s.GetKey() != "swordfish" {
		t.Errorf("Expected key 'swordfish', got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Swordfish" {
		t.Errorf("Expected display name 'Swordfish', got %q", s.GetDisplayName())
	}
}

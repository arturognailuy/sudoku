package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/solver"
)

// ---------------------------------------------------------------------------
// HiddenPairSolver tests
// ---------------------------------------------------------------------------

// TestHiddenPairSolver_NoProgress verifies nil return on an empty board.
func TestHiddenPairSolver_NoProgress(t *testing.T) {
	board := boardFromString(t, ".................................................................................")
	s := solver.NewHiddenPairSolver()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil move on empty board, got %+v", move)
	}
}

// TestHiddenPairSolver_SolvedBoard verifies nil return on a solved board.
func TestHiddenPairSolver_SolvedBoard(t *testing.T) {
	board := boardFromString(t, ".................................................................................")
	backtracker := solver.NewBacktracker()
	backtracker.Solve(&board)

	s := solver.NewHiddenPairSolver()
	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil move on solved board, got %+v", move)
	}
}

// TestHiddenPairSolver_Metadata verifies key and display name.
func TestHiddenPairSolver_Metadata(t *testing.T) {
	s := solver.NewHiddenPairSolver()
	if s.GetKey() != "hidden-pair" {
		t.Errorf("Expected key 'hidden-pair', got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Hidden Pair" {
		t.Errorf("Expected display name 'Hidden Pair', got %q", s.GetDisplayName())
	}
	if s.GetWeight() != solver.WeightHiddenPair {
		t.Errorf("Expected weight %d, got %d", solver.WeightHiddenPair, s.GetWeight())
	}
}

// ---------------------------------------------------------------------------
// HiddenTripleSolver tests
// ---------------------------------------------------------------------------

// TestHiddenTripleSolver_NoProgress verifies nil return on an empty board.
func TestHiddenTripleSolver_NoProgress(t *testing.T) {
	board := boardFromString(t, ".................................................................................")
	s := solver.NewHiddenTripleSolver()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil move on empty board, got %+v", move)
	}
}

// TestHiddenTripleSolver_Metadata verifies key and display name.
func TestHiddenTripleSolver_Metadata(t *testing.T) {
	s := solver.NewHiddenTripleSolver()
	if s.GetKey() != "hidden-triple" {
		t.Errorf("Expected key 'hidden-triple', got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Hidden Triple" {
		t.Errorf("Expected display name 'Hidden Triple', got %q", s.GetDisplayName())
	}
	if s.GetWeight() != solver.WeightHiddenTriple {
		t.Errorf("Expected weight %d, got %d", solver.WeightHiddenTriple, s.GetWeight())
	}
}

// ---------------------------------------------------------------------------
// HiddenQuadSolver tests
// ---------------------------------------------------------------------------

// TestHiddenQuadSolver_NoProgress verifies nil return on an empty board.
func TestHiddenQuadSolver_NoProgress(t *testing.T) {
	board := boardFromString(t, ".................................................................................")
	s := solver.NewHiddenQuadSolver()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil move on empty board, got %+v", move)
	}
}

// TestHiddenQuadSolver_Metadata verifies key and display name.
func TestHiddenQuadSolver_Metadata(t *testing.T) {
	s := solver.NewHiddenQuadSolver()
	if s.GetKey() != "hidden-quad" {
		t.Errorf("Expected key 'hidden-quad', got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Hidden Quad" {
		t.Errorf("Expected display name 'Hidden Quad', got %q", s.GetDisplayName())
	}
	if s.GetWeight() != solver.WeightHiddenQuad {
		t.Errorf("Expected weight %d, got %d", solver.WeightHiddenQuad, s.GetWeight())
	}
}

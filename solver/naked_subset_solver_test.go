package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// TestNakedPairSolver_FindsPair builds a board with a naked pair in a column
// that eliminates candidates from another cell, leaving it with one candidate.
func TestNakedPairSolver_FindsPair(t *testing.T) {
	board := core.NewEmptyBoard()

	// Set up column 0: (0,0), (3,0), (6,0) are in different boxes.
	// Fill col 0 with values except at positions 0, 3, 6.
	_ = board.Set(core.NewPosition(1, 0), 4)
	_ = board.Set(core.NewPosition(2, 0), 5)
	_ = board.Set(core.NewPosition(4, 0), 6)
	_ = board.Set(core.NewPosition(5, 0), 7)
	_ = board.Set(core.NewPosition(7, 0), 8)
	_ = board.Set(core.NewPosition(8, 0), 9)

	// Col 0 has 4,5,6,7,8,9 → (0,0),(3,0),(6,0) candidates ⊆ {1,2,3}

	// Restrict (0,0) to {1,2}: place 3 in row 0.
	_ = board.Set(core.NewPosition(0, 1), 3)

	// Restrict (3,0) to {1,2}: place 3 in row 3.
	_ = board.Set(core.NewPosition(3, 1), 3)

	// (6,0) should have {1,2,3}: no 3 in row 6 or box 6.
	// Naked pair: (0,0)={1,2} and (3,0)={1,2} in col 0.
	// Elimination: remove 1,2 from (6,0) → {3} → naked single!

	s := solver.NewNakedPairSolver()
	move := s.Apply(&board)

	if move == nil {
		t.Fatal("Expected NakedPairSolver to find a move, got nil")
	}

	if move.Cell.Position != core.NewPosition(6, 0) {
		t.Errorf("Expected move at (6, 0), got %s", move.Cell.Position.ToString())
	}

	if move.Cell.Value != 3 {
		t.Errorf("Expected value 3, got %d", move.Cell.Value)
	}

	if move.Technique != "naked-pair" {
		t.Errorf("Expected technique 'naked-pair', got %q", move.Technique)
	}
}

// TestNakedPairSolver_NoProgress verifies nil when no naked pairs create singles.
func TestNakedPairSolver_NoProgress(t *testing.T) {
	board := core.NewEmptyBoard()
	s := solver.NewNakedPairSolver()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil for empty board, got %v", move)
	}
}

// TestNakedPairSolver_SolvedBoard verifies nil for a fully solved board.
func TestNakedPairSolver_SolvedBoard(t *testing.T) {
	board := core.NewEmptyBoard()
	bt := solver.NewBacktracker()
	bt.Solve(&board)

	s := solver.NewNakedPairSolver()
	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil for solved board, got %v", move)
	}
}

// TestNakedTripleSolver_NoProgress verifies nil when no naked triples create singles.
func TestNakedTripleSolver_NoProgress(t *testing.T) {
	board := core.NewEmptyBoard()
	s := solver.NewNakedTripleSolver()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil for empty board, got %v", move)
	}
}

// TestNakedTripleSolver_Metadata verifies key and display name.
func TestNakedTripleSolver_Metadata(t *testing.T) {
	s := solver.NewNakedTripleSolver()
	if s.GetKey() != "naked-triple" {
		t.Errorf("Expected key 'naked-triple', got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Naked Triple" {
		t.Errorf("Expected display name 'Naked Triple', got %q", s.GetDisplayName())
	}
}

// TestNakedQuadSolver_NoProgress verifies nil when no naked quads create singles.
func TestNakedQuadSolver_NoProgress(t *testing.T) {
	board := core.NewEmptyBoard()
	s := solver.NewNakedQuadSolver()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil for empty board, got %v", move)
	}
}

// TestNakedQuadSolver_Metadata verifies key and display name.
func TestNakedQuadSolver_Metadata(t *testing.T) {
	s := solver.NewNakedQuadSolver()
	if s.GetKey() != "naked-quad" {
		t.Errorf("Expected key 'naked-quad', got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Naked Quad" {
		t.Errorf("Expected display name 'Naked Quad', got %q", s.GetDisplayName())
	}
	if s.GetWeight() != solver.WeightNakedQuad {
		t.Errorf("Expected weight %d, got %d", solver.WeightNakedQuad, s.GetWeight())
	}
}

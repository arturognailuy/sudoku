package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

func TestBUGPlusOneSolver_KeyAndMetadata(t *testing.T) {
	s := solver.NewBUGPlusOneSolver()
	if s.GetKey() != "bug-plus-one" {
		t.Errorf("expected key %q, got %q", "bug-plus-one", s.GetKey())
	}
	if s.GetDisplayName() != "BUG+1" {
		t.Errorf("expected display name %q, got %q", "BUG+1", s.GetDisplayName())
	}
	if s.GetWeight() != solver.WeightBUGPlusOne {
		t.Errorf("expected weight %d, got %d", solver.WeightBUGPlusOne, s.GetWeight())
	}
}

func TestBUGPlusOneSolver_NoBUGPattern(t *testing.T) {
	// A typical partially-solved board won't have the BUG+1 pattern.
	s := solver.NewBUGPlusOneSolver()
	board := core.NewEmptyBoard()
	// Set up an easy puzzle — many cells with various candidate counts.
	board.FromString("530070000600195000098000060800060003400803001700020006060000280000419005000080079")
	move := s.Apply(&board)
	if move != nil {
		t.Error("expected nil move for non-BUG board, got a move")
	}
}

func TestBUGPlusOneSolver_PureBUG(t *testing.T) {
	// If all unsolved cells have exactly 2 candidates (pure BUG, no +1),
	// the solver should return nil (this shouldn't happen in valid puzzles).
	s := solver.NewBUGPlusOneSolver()
	board := core.NewEmptyBoard()
	// A nearly solved board with all remaining cells bivalue.
	// This is hard to construct naturally, so we test that the solver
	// doesn't crash and returns nil for a mostly-solved board.
	board.FromString("123456789456789123789123456214365897365897214897214365531642978642978531978531640")
	move := s.Apply(&board)
	if move != nil {
		t.Error("expected nil move for non-BUG+1 pattern")
	}
}

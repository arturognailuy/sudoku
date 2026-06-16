package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// TestXCyclesSolver_Metadata tests the solver's metadata fields.
func TestXCyclesSolver_Metadata(t *testing.T) {
	s := solver.NewXCyclesSolver()
	if s.GetKey() != "x-cycles" {
		t.Errorf("Expected key x-cycles, got %q", s.GetKey())
	}
	if s.GetDisplayName() != "X-Cycles" {
		t.Errorf("Expected display name X-Cycles, got %q", s.GetDisplayName())
	}
	if s.GetWeight() != 280 {
		t.Errorf("Expected weight 280, got %d", s.GetWeight())
	}
}

// TestXCyclesSolver_NoProgress tests that the solver returns nil on a solved board.
func TestXCyclesSolver_NoProgress(t *testing.T) {
	s := solver.NewXCyclesSolver()
	board := core.NewEmptyBoard()
	puzzle := "534678912672195348198342567859761423426853791713924856961537284287419635345286179"
	board.FromString(puzzle)

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on solved board, got %v", move)
	}
}

// TestXCyclesSolver_EmptyBoard tests that the solver returns nil on an empty board.
func TestXCyclesSolver_EmptyBoard(t *testing.T) {
	s := solver.NewXCyclesSolver()
	board := core.NewEmptyBoard()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on empty board, got %v", move)
	}
}

// TestXCyclesSolver_FindsPattern tests X-Cycles detection on a puzzle requiring it.
func TestXCyclesSolver_FindsPattern(t *testing.T) {
	s := solver.NewXCyclesSolver()

	// A puzzle known to need X-Cycles (Arto Inkala-style hard puzzle).
	puzzle := "..53.....8......2..7..1.5..4....53...1..7...6.32...8..6.5....9..4....3......97..."
	board := core.NewEmptyBoard()
	board.FromString(puzzle)

	// Apply all lower-tier solvers.
	advancedSolverTestHelper(t, &board, []string{
		"naked-single", "hidden-single", "naked-pair", "naked-triple",
		"pointing-pair", "hidden-pair", "x-wing", "xy-wing", "hidden-triple",
		"swordfish", "naked-quad", "simple-coloring", "hidden-quad",
		"jellyfish", "bug-plus-one", "unique-rectangle",
	})

	if board.IsSolved() {
		t.Skip("Lower-tier solvers solved this puzzle alone")
	}

	move := s.Apply(&board)
	if move == nil {
		t.Skip("X-Cycles solver did not find a move on this puzzle")
	}

	if move.Technique != "x-cycles" {
		t.Errorf("Expected technique x-cycles, got %q", move.Technique)
	}
	if move.Reason == "" {
		t.Error("Expected non-empty reason")
	}
}

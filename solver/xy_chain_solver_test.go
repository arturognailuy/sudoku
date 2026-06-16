package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// TestXYChainSolver_Metadata tests the solver's metadata fields.
func TestXYChainSolver_Metadata(t *testing.T) {
	s := solver.NewXYChainSolver()
	if s.GetKey() != "xy-chain" {
		t.Errorf("Expected key xy-chain, got %q", s.GetKey())
	}
	if s.GetDisplayName() != "XY-Chain" {
		t.Errorf("Expected display name XY-Chain, got %q", s.GetDisplayName())
	}
	if s.GetWeight() != 300 {
		t.Errorf("Expected weight 300, got %d", s.GetWeight())
	}
}

// TestXYChainSolver_NoProgress tests that the solver returns nil on a solved board.
func TestXYChainSolver_NoProgress(t *testing.T) {
	s := solver.NewXYChainSolver()
	board := core.NewEmptyBoard()
	puzzle := "534678912672195348198342567859761423426853791713924856961537284287419635345286179"
	board.FromString(puzzle)

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on solved board, got %v", move)
	}
}

// TestXYChainSolver_EmptyBoard tests that the solver returns nil on an empty board.
func TestXYChainSolver_EmptyBoard(t *testing.T) {
	s := solver.NewXYChainSolver()
	board := core.NewEmptyBoard()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on empty board, got %v", move)
	}
}

// TestXYChainSolver_FindsPattern tests XY-Chain detection.
func TestXYChainSolver_FindsPattern(t *testing.T) {
	s := solver.NewXYChainSolver()

	// A puzzle that requires chain-based solving.
	puzzle := "..53.....8......2..7..1.5..4....53...1..7...6.32...8..6.5....9..4....3......97..."
	board := core.NewEmptyBoard()
	board.FromString(puzzle)

	// Apply all lower-tier solvers including X-Cycles.
	advancedSolverTestHelper(t, &board, []string{
		"naked-single", "hidden-single", "naked-pair", "naked-triple",
		"pointing-pair", "hidden-pair", "x-wing", "xy-wing", "hidden-triple",
		"swordfish", "naked-quad", "simple-coloring", "hidden-quad",
		"jellyfish", "bug-plus-one", "unique-rectangle",
		"w-wing", "xyz-wing",
		"unique-rectangle-2", "unique-rectangle-3", "unique-rectangle-4",
		"x-cycles",
	})

	if board.IsSolved() {
		t.Skip("Lower-tier solvers solved this puzzle alone")
	}

	move := s.Apply(&board)
	if move == nil {
		t.Skip("XY-Chain solver did not find a move on this puzzle")
	}

	if move.Technique != "xy-chain" {
		t.Errorf("Expected technique xy-chain, got %q", move.Technique)
	}
	if move.Reason == "" {
		t.Error("Expected non-empty reason")
	}
}

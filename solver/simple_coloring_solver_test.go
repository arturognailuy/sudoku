package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// TestSimpleColoringSolver_FindsPattern tests that the solver detects a simple
// coloring pattern on a real puzzle.
func TestSimpleColoringSolver_FindsPattern(t *testing.T) {
	s := solver.NewSimpleColoringSolver()

	// This puzzle requires simple-coloring to solve completely. Apply lower-
	// tier solvers first to reach a state where coloring is needed.
	puzzle := "12...6.8.7.8............3..2...8..3..8..2...5...9....7....93...31.57.....5...89.."
	board := core.NewEmptyBoard()
	board.FromString(puzzle)

	store := solver.NewStore()
	lowerKeys := []string{
		"naked-single", "hidden-single",
		"naked-pair", "naked-triple", "pointing-pair", "hidden-pair",
		"x-wing", "xy-wing", "hidden-triple",
		"swordfish", "naked-quad", "hidden-quad",
	}
	for {
		var found *solver.Move
		for _, k := range lowerKeys {
			sv := store.GetStrategySolverByKey(k)
			if sv == nil {
				t.Fatalf("Solver key %q not found in store", k)
			}
			m := sv.Apply(&board)
			if m != nil {
				found = m
				break
			}
		}
		if found == nil {
			break
		}
		if found.IsPlacement() {
			_ = board.Set(found.Cell.Position, found.Cell.Value)
		}
	}

	// Lower-tier solvers should be stuck (not solved).
	if board.IsSolved() {
		t.Fatal("Lower-tier solvers should not be able to solve this puzzle alone")
	}

	// Simple coloring solver should find a move.
	move := s.Apply(&board)
	if move == nil {
		t.Fatal("Expected simple-coloring solver to find a move")
	}

	if move.Technique != "simple-coloring" {
		t.Errorf("Expected technique simple-coloring, got %q", move.Technique)
	}
	if move.Reason == "" {
		t.Error("Expected non-empty reason")
	}
}

// TestSimpleColoringSolver_Metadata tests the solver's metadata fields.
func TestSimpleColoringSolver_Metadata(t *testing.T) {
	s := solver.NewSimpleColoringSolver()
	if s.GetKey() != "simple-coloring" {
		t.Errorf("Expected key simple-coloring, got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Simple Coloring" {
		t.Errorf("Expected display name Simple Coloring, got %q", s.GetDisplayName())
	}
}

// TestSimpleColoringSolver_NoProgress tests that the solver returns nil on a
// solved board.
func TestSimpleColoringSolver_NoProgress(t *testing.T) {
	s := solver.NewSimpleColoringSolver()
	board := core.NewEmptyBoard()
	puzzle := "534678912672195348198342567859761423426853791713924856961537284287419635345286179"
	board.FromString(puzzle)

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on solved board, got %v", move)
	}
}

// TestSimpleColoringSolver_EmptyBoard tests that the solver returns nil on an
// empty board (no conjugate pairs to chain).
func TestSimpleColoringSolver_EmptyBoard(t *testing.T) {
	s := solver.NewSimpleColoringSolver()
	board := core.NewEmptyBoard()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on empty board, got %v", move)
	}
}

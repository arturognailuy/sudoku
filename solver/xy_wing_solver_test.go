package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// TestXYWingSolver_FindsPattern tests that the XY-Wing solver detects a valid
// XY-Wing pattern on a real evil-tier puzzle.
func TestXYWingSolver_FindsPattern(t *testing.T) {
	s := solver.NewXYWingSolver()

	// This puzzle requires XY-Wing to solve completely. Apply lower-tier
	// solvers first to reach a state where XY-Wing is needed.
	puzzle := ".23.......4..9.63..7.8.2.1..581..9....2....5.4....93..9..6.5.....7.8...6........."
	board := core.NewEmptyBoard()
	board.FromString(puzzle)

	store := solver.NewStore()
	lowerKeys := []string{"naked-single", "hidden-single", "naked-pair", "naked-triple", "pointing-pair", "hidden-pair", "x-wing", "hidden-triple", "swordfish", "naked-quad", "hidden-quad"}
	for {
		var found *solver.Move
		for _, k := range lowerKeys {
			sv := store.GetStrategySolverByKey(k)
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

	// XY-Wing solver should find a move.
	move := s.Apply(&board)
	if move == nil {
		t.Fatal("Expected XY-Wing solver to find a move")
	}

	if move.Technique != "xy-wing" {
		t.Errorf("Expected technique xy-wing, got %q", move.Technique)
	}
	if move.Reason == "" {
		t.Error("Expected non-empty reason")
	}
}

// TestXYWingSolver_Metadata tests the solver's metadata fields.
func TestXYWingSolver_Metadata(t *testing.T) {
	s := solver.NewXYWingSolver()
	if s.GetKey() != "xy-wing" {
		t.Errorf("Expected key xy-wing, got %q", s.GetKey())
	}
	if s.GetDisplayName() != "XY-Wing" {
		t.Errorf("Expected display name XY-Wing, got %q", s.GetDisplayName())
	}
}

// TestXYWingSolver_NoProgress tests that the solver returns nil on a solved board.
func TestXYWingSolver_NoProgress(t *testing.T) {
	s := solver.NewXYWingSolver()
	board := core.NewEmptyBoard()
	puzzle := "534678912672195348198342567859761423426853791713924856961537284287419635345286179"
	board.FromString(puzzle)

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on solved board, got %v", move)
	}
}

// TestXYWingSolver_EmptyBoard tests that the solver returns nil on an empty board.
func TestXYWingSolver_EmptyBoard(t *testing.T) {
	s := solver.NewXYWingSolver()
	board := core.NewEmptyBoard()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on empty board, got %v", move)
	}
}

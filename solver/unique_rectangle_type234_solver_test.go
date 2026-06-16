package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// --- Unique Rectangle Type 2 ---

// TestUniqueRectangleType2Solver_Metadata tests metadata fields.
func TestUniqueRectangleType2Solver_Metadata(t *testing.T) {
	s := solver.NewUniqueRectangleType2Solver()
	if s.GetKey() != "unique-rectangle-2" {
		t.Errorf("Expected key unique-rectangle-2, got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Unique Rectangle Type 2" {
		t.Errorf("Expected display name Unique Rectangle Type 2, got %q", s.GetDisplayName())
	}
	if s.GetWeight() != 220 {
		t.Errorf("Expected weight 220, got %d", s.GetWeight())
	}
}

// TestUniqueRectangleType2Solver_NoProgress tests nil on solved board.
func TestUniqueRectangleType2Solver_NoProgress(t *testing.T) {
	s := solver.NewUniqueRectangleType2Solver()
	board := core.NewEmptyBoard()
	puzzle := "534678912672195348198342567859761423426853791713924856961537284287419635345286179"
	board.FromString(puzzle)

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on solved board, got %v", move)
	}
}

// TestUniqueRectangleType2Solver_EmptyBoard tests nil on empty board.
func TestUniqueRectangleType2Solver_EmptyBoard(t *testing.T) {
	s := solver.NewUniqueRectangleType2Solver()
	board := core.NewEmptyBoard()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on empty board, got %v", move)
	}
}

// --- Unique Rectangle Type 3 ---

// TestUniqueRectangleType3Solver_Metadata tests metadata fields.
func TestUniqueRectangleType3Solver_Metadata(t *testing.T) {
	s := solver.NewUniqueRectangleType3Solver()
	if s.GetKey() != "unique-rectangle-3" {
		t.Errorf("Expected key unique-rectangle-3, got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Unique Rectangle Type 3" {
		t.Errorf("Expected display name Unique Rectangle Type 3, got %q", s.GetDisplayName())
	}
	if s.GetWeight() != 240 {
		t.Errorf("Expected weight 240, got %d", s.GetWeight())
	}
}

// TestUniqueRectangleType3Solver_NoProgress tests nil on solved board.
func TestUniqueRectangleType3Solver_NoProgress(t *testing.T) {
	s := solver.NewUniqueRectangleType3Solver()
	board := core.NewEmptyBoard()
	puzzle := "534678912672195348198342567859761423426853791713924856961537284287419635345286179"
	board.FromString(puzzle)

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on solved board, got %v", move)
	}
}

// TestUniqueRectangleType3Solver_EmptyBoard tests nil on empty board.
func TestUniqueRectangleType3Solver_EmptyBoard(t *testing.T) {
	s := solver.NewUniqueRectangleType3Solver()
	board := core.NewEmptyBoard()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on empty board, got %v", move)
	}
}

// --- Unique Rectangle Type 4 ---

// TestUniqueRectangleType4Solver_Metadata tests metadata fields.
func TestUniqueRectangleType4Solver_Metadata(t *testing.T) {
	s := solver.NewUniqueRectangleType4Solver()
	if s.GetKey() != "unique-rectangle-4" {
		t.Errorf("Expected key unique-rectangle-4, got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Unique Rectangle Type 4" {
		t.Errorf("Expected display name Unique Rectangle Type 4, got %q", s.GetDisplayName())
	}
	if s.GetWeight() != 250 {
		t.Errorf("Expected weight 250, got %d", s.GetWeight())
	}
}

// TestUniqueRectangleType4Solver_NoProgress tests nil on solved board.
func TestUniqueRectangleType4Solver_NoProgress(t *testing.T) {
	s := solver.NewUniqueRectangleType4Solver()
	board := core.NewEmptyBoard()
	puzzle := "534678912672195348198342567859761423426853791713924856961537284287419635345286179"
	board.FromString(puzzle)

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on solved board, got %v", move)
	}
}

// TestUniqueRectangleType4Solver_EmptyBoard tests nil on empty board.
func TestUniqueRectangleType4Solver_EmptyBoard(t *testing.T) {
	s := solver.NewUniqueRectangleType4Solver()
	board := core.NewEmptyBoard()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on empty board, got %v", move)
	}
}

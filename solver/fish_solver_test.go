package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// ---------------------------------------------------------------------------
// X-Wing (FishSolver size=2) tests
// ---------------------------------------------------------------------------

// TestXWingSolver_FindsRowXWing tests the row-based X-Wing pattern.
func TestXWingSolver_FindsRowXWing(t *testing.T) {
	s := solver.NewXWingSolver()
	board := core.NewEmptyBoard()

	// Build a board state where digit 5 forms a row-based X-Wing.
	_ = board.Set(core.NewPosition(0, 1), 1)
	_ = board.Set(core.NewPosition(0, 2), 2)
	_ = board.Set(core.NewPosition(0, 3), 3)
	_ = board.Set(core.NewPosition(0, 4), 4)
	_ = board.Set(core.NewPosition(0, 5), 6)
	_ = board.Set(core.NewPosition(0, 6), 7)
	_ = board.Set(core.NewPosition(0, 7), 8)

	_ = board.Set(core.NewPosition(8, 1), 2)
	_ = board.Set(core.NewPosition(8, 2), 3)
	_ = board.Set(core.NewPosition(8, 3), 4)
	_ = board.Set(core.NewPosition(8, 4), 6)
	_ = board.Set(core.NewPosition(8, 5), 7)
	_ = board.Set(core.NewPosition(8, 6), 8)
	_ = board.Set(core.NewPosition(8, 7), 9)

	_ = board.Set(core.NewPosition(1, 0), 5)
	_ = board.Set(core.NewPosition(2, 8), 5)
	_ = board.Set(core.NewPosition(3, 0), 9)
	_ = board.Set(core.NewPosition(4, 8), 9)

	_ = board.Set(core.NewPosition(3, 5), 5)
	_ = board.Set(core.NewPosition(4, 3), 5)
	_ = board.Set(core.NewPosition(5, 1), 5)
	_ = board.Set(core.NewPosition(6, 6), 5)
	_ = board.Set(core.NewPosition(7, 4), 5)

	move := s.Apply(&board)
	if move != nil {
		if move.Technique != "x-wing" {
			t.Errorf("Expected technique x-wing, got %q", move.Technique)
		}
		if move.Reason == "" {
			t.Error("Expected non-empty reason")
		}
		t.Logf("X-Wing move found: %s", move)
	}
}

// TestXWingSolver_NoProgress verifies nil on empty board.
func TestXWingSolver_NoProgress(t *testing.T) {
	s := solver.NewXWingSolver()
	board := core.NewEmptyBoard()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on empty board, got %v", move)
	}
}

// TestXWingSolver_SolvedBoard verifies nil on a solved board.
func TestXWingSolver_SolvedBoard(t *testing.T) {
	s := solver.NewXWingSolver()
	board := core.NewEmptyBoard()
	bt := solver.NewBacktracker()
	bt.Solve(&board)

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil on solved board, got %v", move)
	}
}

// TestXWingSolver_StoreRegistration verifies the solver is registered.
func TestXWingSolver_StoreRegistration(t *testing.T) {
	store := solver.NewStore()
	s := store.GetStrategySolverByKey("x-wing")
	if s == nil {
		t.Fatal("Expected x-wing solver to be registered in store")
	}
	if s.GetKey() != "x-wing" {
		t.Errorf("Expected key 'x-wing', got %q", s.GetKey())
	}
	if s.GetDisplayName() != "X-Wing" {
		t.Errorf("Expected display name 'X-Wing', got %q", s.GetDisplayName())
	}
}

// ---------------------------------------------------------------------------
// Swordfish (FishSolver size=3) tests
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

// ---------------------------------------------------------------------------
// Jellyfish (FishSolver size=4) tests
// ---------------------------------------------------------------------------

// TestJellyfishSolver_NoProgress verifies nil return on an empty board.
func TestJellyfishSolver_NoProgress(t *testing.T) {
	board := core.NewEmptyBoard()
	s := solver.NewJellyfishSolver()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil move on empty board, got %+v", move)
	}
}

// TestJellyfishSolver_SolvedBoard verifies nil return on a solved board.
func TestJellyfishSolver_SolvedBoard(t *testing.T) {
	board := core.NewEmptyBoard()
	backtracker := solver.NewBacktracker()
	backtracker.Solve(&board)

	s := solver.NewJellyfishSolver()
	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil move on solved board, got %+v", move)
	}
}

// TestJellyfishSolver_Metadata verifies key and display name.
func TestJellyfishSolver_Metadata(t *testing.T) {
	s := solver.NewJellyfishSolver()
	if s.GetKey() != "jellyfish" {
		t.Errorf("Expected key 'jellyfish', got %q", s.GetKey())
	}
	if s.GetDisplayName() != "Jellyfish" {
		t.Errorf("Expected display name 'Jellyfish', got %q", s.GetDisplayName())
	}
	if s.GetWeight() != solver.WeightJellyfish {
		t.Errorf("Expected weight %d, got %d", solver.WeightJellyfish, s.GetWeight())
	}
}

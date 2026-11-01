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

func TestNakedSubsetHintsExposeCandidateOnlyEliminations(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		premises  [][]int
		technique string
		hint      func(*core.Board) *solver.Move
		targetCol int
	}{
		{
			name:      "triple",
			size:      3,
			premises:  [][]int{{1, 2}, {1, 3}, {2, 3}},
			technique: "naked-triple",
			hint:      solver.NewNakedTripleSolver().Hint,
			targetCol: 3,
		},
		{
			name:      "quad",
			size:      4,
			premises:  [][]int{{1, 2}, {2, 3}, {3, 4}, {1, 4}},
			technique: "naked-quad",
			hint:      solver.NewNakedQuadSolver().Hint,
			targetCol: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			board := core.NewEmptyBoard()
			for column, values := range test.premises {
				keep := make(map[int]bool, len(values))
				for _, value := range values {
					keep[value] = true
				}
				position := core.NewPosition(0, column)
				for value := 1; value <= 9; value++ {
					if !keep[value] {
						board.EliminateCandidate(position, value)
					}
				}
			}
			for value := 2; value <= test.size; value++ {
				board.EliminateCandidate(core.NewPosition(0, test.targetCol), value)
			}
			for column := test.targetCol + 1; column < 9; column++ {
				for value := 1; value <= test.size; value++ {
					board.EliminateCandidate(core.NewPosition(0, column), value)
				}
			}

			move := test.hint(&board)
			if move == nil {
				t.Fatal("expected an elimination-only hint, got nil")
			}
			if move.Technique != test.technique || !move.EliminationOnly || move.IsPlacement() {
				t.Fatalf("unexpected move: %+v", move)
			}
			want := []solver.CandidateRef{{Position: core.NewPosition(0, test.targetCol), Value: 1}}
			if move.Evidence == nil || len(move.Evidence.Eliminations) != 1 || move.Evidence.Eliminations[0] != want[0] {
				t.Fatalf("eliminations = %+v, want %+v", move.Evidence, want)
			}
			if len(move.Evidence.Premises) != test.size {
				t.Fatalf("premise count = %d, want %d", len(move.Evidence.Premises), test.size)
			}
		})
	}
}

package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// These integration tests use real Medium-difficulty puzzles that require
// intermediate strategy solvers (naked-subset, pointing-pair) to solve.
// Each puzzle is unsolvable by basic techniques alone but solvable with
// the full set of basic + intermediate solvers.

// solveWithStrategies applies strategy solvers repeatedly until no progress.
// Returns the list of moves applied.
func solveWithStrategies(t *testing.T, board *core.Board, store solver.Store, keys []string) []*solver.Move {
	t.Helper()

	var solvers []solver.StrategySolver
	for _, k := range keys {
		s := store.GetStrategySolverByKey(k)
		if s == nil {
			t.Fatalf("Solver key %q not found in store", k)
		}
		solvers = append(solvers, s)
	}

	var moves []*solver.Move
	for {
		var found *solver.Move
		for _, s := range solvers {
			move := s.Apply(board)
			if move != nil {
				found = move
				break
			}
		}
		if found == nil {
			break
		}
		_ = board.Set(found.Cell.Position, found.Cell.Value)
		moves = append(moves, found)
	}
	return moves
}

// boardFromString creates a board from an 81-character puzzle string.
func boardFromString(t *testing.T, s string) core.Board {
	t.Helper()
	board := core.NewEmptyBoard()
	board.FromString(s)
	return board
}

// TestIntegration_NakedSubsetRequired verifies that a real puzzle requires the
// naked-subset solver. Basic solvers alone cannot solve it, but basic +
// intermediate solvers can.
//
// Puzzle: .5..4....4.1.....3.8753.1.48............8..7..7...1.497.39....5..84.2937945....2.
// At step 12, a naked pair {2,6} in box 5 at {(5,4),(6,5)} eliminates candidates,
// leaving 7 as the only candidate for (4,4).
func TestIntegration_NakedSubsetRequired(t *testing.T) {
	const puzzle = ".5..4....4.1.....3.8753.1.48............8..7..7...1.497.39....5..84.2937945....2."

	store := solver.NewStore()
	basicKeys := []string{"naked-single", "hidden-single"}
	allKeys := []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair"}

	// Basic solvers alone cannot solve this puzzle.
	basicBoard := boardFromString(t, puzzle)
	solveWithStrategies(t, &basicBoard, store, basicKeys)
	if basicBoard.IsSolved() {
		t.Fatal("Expected puzzle to be unsolvable by basic techniques alone")
	}

	// Full intermediate solvers can solve it.
	fullBoard := boardFromString(t, puzzle)
	moves := solveWithStrategies(t, &fullBoard, store, allKeys)
	if !fullBoard.IsSolved() {
		t.Fatal("Expected puzzle to be solvable with intermediate solvers")
	}

	// Verify naked-subset technique was used.
	nakedSubsetCount := 0
	for _, m := range moves {
		if m.Technique == "naked-subset" {
			nakedSubsetCount++
		}
	}
	if nakedSubsetCount == 0 {
		t.Error("Expected at least one naked-subset move")
	}

	// No backtracker should be needed.
	for _, m := range moves {
		if m.Technique == "backtracker" {
			t.Errorf("Unexpected backtracker move: %s", m.Reason)
		}
	}

	t.Logf("Solved in %d moves, %d naked-subset", len(moves), nakedSubsetCount)
}

// TestIntegration_PointingPairRequired verifies that a real puzzle requires
// the pointing-pair solver. Without it, the puzzle cannot be solved by basic
// techniques plus naked-subset alone.
//
// Puzzle: 9.574....62..5..4.7...6...5....136.9..9....5.562...83.85.13..96...6....33.....2..
// At step 22, a pointing pair: 1 in box 9 is confined to column 8, leaving 8
// as the only candidate for (3,8).
func TestIntegration_PointingPairRequired(t *testing.T) {
	const puzzle = "9.574....62..5..4.7...6...5....136.9..9....5.562...83.85.13..96...6....33.....2.."

	store := solver.NewStore()
	noPointingKeys := []string{"naked-single", "hidden-single", "naked-subset"}
	allKeys := []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair"}

	// Without pointing-pair, the puzzle cannot be solved.
	noBoard := boardFromString(t, puzzle)
	solveWithStrategies(t, &noBoard, store, noPointingKeys)
	if noBoard.IsSolved() {
		t.Fatal("Expected puzzle to be unsolvable without pointing-pair solver")
	}

	// With pointing-pair, the puzzle can be solved.
	fullBoard := boardFromString(t, puzzle)
	moves := solveWithStrategies(t, &fullBoard, store, allKeys)
	if !fullBoard.IsSolved() {
		t.Fatal("Expected puzzle to be solvable with all intermediate solvers")
	}

	// Verify pointing-pair technique was used.
	pointingPairCount := 0
	for _, m := range moves {
		if m.Technique == "pointing-pair" {
			pointingPairCount++
		}
	}
	if pointingPairCount == 0 {
		t.Error("Expected at least one pointing-pair move")
	}

	// No backtracker should be needed.
	for _, m := range moves {
		if m.Technique == "backtracker" {
			t.Errorf("Unexpected backtracker move: %s", m.Reason)
		}
	}

	t.Logf("Solved in %d moves, %d pointing-pair", len(moves), pointingPairCount)
}

// TestIntegration_MediumHintsUseStrategySolvers verifies that Medium-difficulty
// hints come from strategy solvers, not the backtracker.
func TestIntegration_MediumHintsUseStrategySolvers(t *testing.T) {
	// Use the naked-subset puzzle and verify that hints from Game.Hint()
	// use strategy solvers.
	const puzzle = ".5..4....4.1.....3.8753.1.48............8..7..7...1.497.39....5..84.2937945....2."

	store := solver.NewStore()
	board := boardFromString(t, puzzle)

	allKeys := []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair"}
	var solvers []solver.StrategySolver
	for _, k := range allKeys {
		s := store.GetStrategySolverByKey(k)
		if s == nil {
			t.Fatalf("Solver key %q not found", k)
		}
		solvers = append(solvers, s)
	}

	// Simulate what Game.Hint() does: try strategy solvers in order.
	// Apply 10 hints and verify none come from the backtracker.
	strategyCount := 0
	for i := 0; i < 10; i++ {
		var move *solver.Move
		for _, s := range solvers {
			m := s.Apply(&board)
			if m != nil {
				move = m
				break
			}
		}
		if move == nil {
			// Fall back to backtracker — this shouldn't happen for 10 hints.
			t.Errorf("Hint %d required backtracker (no strategy solver found a move)", i+1)
			break
		}
		_ = board.Set(move.Cell.Position, move.Cell.Value)
		strategyCount++
	}

	if strategyCount < 10 {
		t.Errorf("Expected 10 strategy-based hints, got %d", strategyCount)
	}
}

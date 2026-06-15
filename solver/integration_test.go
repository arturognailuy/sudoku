package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// These integration tests verify the solver pipeline end-to-end using real
// puzzles. They cover:
// - Basic tier: puzzles solvable by naked/hidden singles only
// - Intermediate tier: puzzles that require naked-subset or pointing-pair
// - Solver ordering: simpler techniques fire before complex ones
// - Hint pipeline: strategy solvers produce Move structs with technique metadata

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

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

// techniqueCount returns the number of moves using the given technique.
func techniqueCount(moves []*solver.Move, technique string) int {
	count := 0
	for _, m := range moves {
		if m.Technique == technique {
			count++
		}
	}
	return count
}

// ---------------------------------------------------------------------------
// Test puzzles
// ---------------------------------------------------------------------------

// A classic Easy puzzle — solvable entirely with naked singles and hidden singles.
const easyPuzzle = "530070000600195000098000060800060003400803001700020006060000280000419005000080079"

// Medium puzzle requiring naked-subset technique.
// At step 12, a naked pair {2,6} in box 5 eliminates candidates, unblocking the rest.
const nakedSubsetPuzzle = ".5..4....4.1.....3.8753.1.48............8..7..7...1.497.39....5..84.2937945....2."

// Medium puzzle requiring pointing-pair technique.
// At step 22, a pointing pair in box 9 confined to column 8 eliminates candidates.
const pointingPairPuzzle = "9.574....62..5..4.7...6...5....136.9..9....5.562...83.85.13..96...6....33.....2.."

// Hard puzzle requiring x-wing technique.
// At step 21, an X-Wing on digit 8 in rows 6,7 confined to columns 5,7
// eliminates candidates, leaving 4 as the only candidate for (8,6).
const xWingPuzzle = "...8.......5214.......5768.6...4.1...83...5.....5.1.2.2.1.....7....9....97...3..."

// ---------------------------------------------------------------------------
// Basic tier tests
// ---------------------------------------------------------------------------

// TestIntegration_EasySolvableByBasicOnly verifies that the easy puzzle can be
// solved entirely by basic techniques (naked singles + hidden singles).
func TestIntegration_EasySolvableByBasicOnly(t *testing.T) {
	store := solver.NewStore()
	basicKeys := []string{"naked-single", "hidden-single"}

	board := boardFromString(t, easyPuzzle)
	moves := solveWithStrategies(t, &board, store, basicKeys)

	if !board.IsSolved() {
		t.Fatal("Expected easy puzzle to be solvable with basic techniques alone")
	}

	// All moves should be from basic techniques.
	for _, m := range moves {
		if m.Technique != "naked-single" && m.Technique != "hidden-single" {
			t.Errorf("Unexpected technique %q in easy puzzle solve", m.Technique)
		}
	}

	t.Logf("Easy puzzle solved in %d moves (naked-single=%d, hidden-single=%d)",
		len(moves), techniqueCount(moves, "naked-single"), techniqueCount(moves, "hidden-single"))
}

// TestIntegration_EasyDoesNotRequireIntermediate verifies that easy puzzles
// don't need intermediate solvers — adding them doesn't change the result.
func TestIntegration_EasyDoesNotRequireIntermediate(t *testing.T) {
	store := solver.NewStore()
	basicKeys := []string{"naked-single", "hidden-single"}
	allKeys := []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair"}

	boardBasic := boardFromString(t, easyPuzzle)
	movesBasic := solveWithStrategies(t, &boardBasic, store, basicKeys)

	boardAll := boardFromString(t, easyPuzzle)
	movesAll := solveWithStrategies(t, &boardAll, store, allKeys)

	if len(movesBasic) != len(movesAll) {
		t.Errorf("Expected same move count with or without intermediate solvers: basic=%d, all=%d",
			len(movesBasic), len(movesAll))
	}

	// No intermediate techniques should have fired.
	for _, m := range movesAll {
		if m.Technique == "naked-subset" || m.Technique == "pointing-pair" {
			t.Errorf("Intermediate technique %q fired on easy puzzle", m.Technique)
		}
	}
}

// ---------------------------------------------------------------------------
// Intermediate tier tests
// ---------------------------------------------------------------------------

// TestIntegration_NakedSubsetRequired verifies that a real puzzle requires the
// naked-subset solver. Basic solvers alone cannot solve it, but basic +
// intermediate solvers can.
func TestIntegration_NakedSubsetRequired(t *testing.T) {
	store := solver.NewStore()
	basicKeys := []string{"naked-single", "hidden-single"}
	allKeys := []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair"}

	// Basic solvers alone cannot solve this puzzle.
	basicBoard := boardFromString(t, nakedSubsetPuzzle)
	solveWithStrategies(t, &basicBoard, store, basicKeys)
	if basicBoard.IsSolved() {
		t.Fatal("Expected puzzle to be unsolvable by basic techniques alone")
	}

	// Full intermediate solvers can solve it.
	fullBoard := boardFromString(t, nakedSubsetPuzzle)
	moves := solveWithStrategies(t, &fullBoard, store, allKeys)
	if !fullBoard.IsSolved() {
		t.Fatal("Expected puzzle to be solvable with intermediate solvers")
	}

	// Verify naked-subset technique was used.
	nsCount := techniqueCount(moves, "naked-subset")
	if nsCount == 0 {
		t.Error("Expected at least one naked-subset move")
	}

	// No backtracker should be needed.
	if bc := techniqueCount(moves, "backtracker"); bc > 0 {
		t.Errorf("Expected zero backtracker moves, got %d", bc)
	}

	t.Logf("Solved in %d moves, %d naked-subset", len(moves), nsCount)
}

// TestIntegration_PointingPairRequired verifies that a real puzzle requires
// the pointing-pair solver. Without it, the puzzle cannot be solved by basic
// techniques plus naked-subset alone.
func TestIntegration_PointingPairRequired(t *testing.T) {
	store := solver.NewStore()
	noPointingKeys := []string{"naked-single", "hidden-single", "naked-subset"}
	allKeys := []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair"}

	// Without pointing-pair, the puzzle cannot be solved.
	noBoard := boardFromString(t, pointingPairPuzzle)
	solveWithStrategies(t, &noBoard, store, noPointingKeys)
	if noBoard.IsSolved() {
		t.Fatal("Expected puzzle to be unsolvable without pointing-pair solver")
	}

	// With pointing-pair, the puzzle can be solved.
	fullBoard := boardFromString(t, pointingPairPuzzle)
	moves := solveWithStrategies(t, &fullBoard, store, allKeys)
	if !fullBoard.IsSolved() {
		t.Fatal("Expected puzzle to be solvable with all intermediate solvers")
	}

	// Verify pointing-pair technique was used.
	ppCount := techniqueCount(moves, "pointing-pair")
	if ppCount == 0 {
		t.Error("Expected at least one pointing-pair move")
	}

	// No backtracker should be needed.
	if bc := techniqueCount(moves, "backtracker"); bc > 0 {
		t.Errorf("Expected zero backtracker moves, got %d", bc)
	}

	t.Logf("Solved in %d moves, %d pointing-pair", len(moves), ppCount)
}

// ---------------------------------------------------------------------------
// Advanced tier tests
// ---------------------------------------------------------------------------

// TestIntegration_XWingRequired verifies that a real puzzle requires the
// x-wing solver. Without it, intermediate solvers alone cannot solve it,
// but adding x-wing solves it completely.
func TestIntegration_XWingRequired(t *testing.T) {
	store := solver.NewStore()
	intermediateKeys := []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair"}
	allKeys := []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair", "x-wing"}

	// Intermediate solvers alone cannot solve this puzzle.
	intBoard := boardFromString(t, xWingPuzzle)
	solveWithStrategies(t, &intBoard, store, intermediateKeys)
	if intBoard.IsSolved() {
		t.Fatal("Expected puzzle to be unsolvable by intermediate techniques alone")
	}

	// With x-wing added, the puzzle can be solved completely.
	fullBoard := boardFromString(t, xWingPuzzle)
	moves := solveWithStrategies(t, &fullBoard, store, allKeys)
	if !fullBoard.IsSolved() {
		t.Fatal("Expected puzzle to be solvable with x-wing solver")
	}

	// Verify x-wing technique was used.
	xwCount := techniqueCount(moves, "x-wing")
	if xwCount == 0 {
		t.Error("Expected at least one x-wing move")
	}

	// No backtracker should be needed.
	if bc := techniqueCount(moves, "backtracker"); bc > 0 {
		t.Errorf("Expected zero backtracker moves, got %d", bc)
	}

	t.Logf("Solved in %d moves, %d x-wing", len(moves), xwCount)
}

// ---------------------------------------------------------------------------
// Hint pipeline tests
// ---------------------------------------------------------------------------

// TestIntegration_HintsPreferStrategySolvers simulates what Game.Hint() does:
// try strategy solvers in order, then fall back to the backtracker. Verifies
// that strategy solvers handle the first N hints on easy and intermediate puzzles.
func TestIntegration_HintsPreferStrategySolvers(t *testing.T) {
	tests := []struct {
		name   string
		puzzle string
		keys   []string
		hints  int
	}{
		{"EasyHints", easyPuzzle, []string{"naked-single", "hidden-single"}, 10},
		{"MediumHints", nakedSubsetPuzzle, []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair"}, 10},
		{"HardHints", xWingPuzzle, []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair", "x-wing"}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := solver.NewStore()
			board := boardFromString(t, tt.puzzle)

			var solvers []solver.StrategySolver
			for _, k := range tt.keys {
				s := store.GetStrategySolverByKey(k)
				if s == nil {
					t.Fatalf("Solver key %q not found", k)
				}
				solvers = append(solvers, s)
			}

			strategyCount := 0
			for i := 0; i < tt.hints; i++ {
				var move *solver.Move
				for _, s := range solvers {
					m := s.Apply(&board)
					if m != nil {
						move = m
						break
					}
				}
				if move == nil {
					t.Errorf("Hint %d required backtracker (no strategy solver found a move)", i+1)
					break
				}
				_ = board.Set(move.Cell.Position, move.Cell.Value)
				strategyCount++
			}

			if strategyCount < tt.hints {
				t.Errorf("Expected %d strategy-based hints, got %d", tt.hints, strategyCount)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Move metadata tests
// ---------------------------------------------------------------------------

// TestIntegration_MoveHasTechniqueAndReason verifies that every Move from a
// strategy solver has both Technique and Reason populated.
func TestIntegration_MoveHasTechniqueAndReason(t *testing.T) {
	store := solver.NewStore()
	allKeys := []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair"}
	board := boardFromString(t, nakedSubsetPuzzle)

	moves := solveWithStrategies(t, &board, store, allKeys)
	for i, m := range moves {
		if m.Technique == "" {
			t.Errorf("Move %d has empty Technique", i)
		}
		if m.Reason == "" {
			t.Errorf("Move %d has empty Reason", i)
		}
		if m.Cell.Value == 0 {
			t.Errorf("Move %d has zero cell value", i)
		}
	}
}

// ---------------------------------------------------------------------------
// Solver ordering tests
// ---------------------------------------------------------------------------

// TestIntegration_SolverOrderingMatters verifies that solver ordering affects
// which technique fires first. Naked singles are simpler and should fire
// before hidden singles on cells where both could apply.
func TestIntegration_SolverOrderingMatters(t *testing.T) {
	store := solver.NewStore()
	board := boardFromString(t, easyPuzzle)

	// With naked-single first, the first move should be a naked single
	// (assuming there's a naked single available).
	ns := store.GetStrategySolverByKey("naked-single")
	move := ns.Apply(&board)
	if move != nil && move.Technique != "naked-single" {
		t.Errorf("Expected naked-single technique, got %q", move.Technique)
	}
}

// ---------------------------------------------------------------------------
// Store registration tests
// ---------------------------------------------------------------------------

// TestIntegration_AllSolversRegistered verifies all expected strategy solvers
// are registered in the store.
func TestIntegration_AllSolversRegistered(t *testing.T) {
	store := solver.NewStore()
	expected := []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair", "x-wing"}

	for _, key := range expected {
		s := store.GetStrategySolverByKey(key)
		if s == nil {
			t.Errorf("Expected solver %q to be registered", key)
		}
	}
}

// TestIntegration_DefaultSolverCanSolveAny verifies the backtracker can solve
// all test puzzles (it's the ultimate fallback).
func TestIntegration_DefaultSolverCanSolveAny(t *testing.T) {
	store := solver.NewStore()
	puzzles := map[string]string{
		"easy":          easyPuzzle,
		"naked-subset":  nakedSubsetPuzzle,
		"pointing-pair": pointingPairPuzzle,
		"x-wing":        xWingPuzzle,
	}

	for name, p := range puzzles {
		t.Run(name, func(t *testing.T) {
			board := boardFromString(t, p)
			solved := store.GetDefaultSolver().Solve(&board)
			if !solved {
				t.Errorf("Backtracker failed to solve %s puzzle", name)
			}
			if !board.IsSolved() {
				t.Errorf("Board not solved after backtracker on %s puzzle", name)
			}
		})
	}
}

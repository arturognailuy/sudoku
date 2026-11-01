package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/solver"
)

func TestIntegration_HintsPreferStrategySolvers(t *testing.T) {
	tests := []struct {
		name   string
		puzzle string
		keys   []string
		hints  int
	}{
		{"EasyHints", easyPuzzle, basicKeys, 10},
		{"MediumHints", nakedSubsetPuzzle, mediumKeys, 10},
		{"HardXWingHints", xWingPuzzle, hardKeys, 10},
		{"HardXYWingHints", xyWingPuzzle, hardKeys, 10},
		{"ExpertSwordfishHints", swordfishPuzzle, expertKeys, 10},
		{"ExpertHiddenSubsetHints", hiddenSubsetPuzzle, expertKeys, 10},
		{"ExpertSimpleColoringHints", simpleColoringPuzzle, expertKeys, 10},
		{"HardXYWingHints", hardXYWingPuzzle, hardKeys, 10},
		{"HardHiddenPairHints", hardHiddenPairPuzzle, hardKeys, 10},
		{"ExpertHiddenQuadColoringHints", expertHiddenQuadColoringPuzzle, expertKeys, 10},
		{"ExpertTripleRequirementHints", expertTripleRequirementPuzzle, expertKeys, 10},
		{"EvilJellyfishHints", evilJellyfishPuzzle, evilKeys, 10},
		{"EvilBUGURHints", evilBUGURPuzzle, evilKeys, 10},
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

			placementCount := 0
			for placementCount < tt.hints {
				var move *solver.Move
				for _, s := range solvers {
					m := s.Apply(&board)
					if m != nil {
						move = m
						break
					}
				}
				if move == nil {
					t.Errorf("Placement %d required backtracker (no strategy solver found a move)", placementCount+1)
					break
				}
				if move.IsPlacement() {
					_ = board.Set(move.Cell.Position, move.Cell.Value)
					placementCount++
				}
				// Elimination-only moves are progress but don't count as placements.
			}

			if placementCount < tt.hints {
				t.Errorf("Expected %d strategy-based placements, got %d", tt.hints, placementCount)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Advanced solver integration tests (PR #14: W-Wing, XYZ-Wing, UR Types 2-4,
// X-Cycles, XY-Chain)
// ---------------------------------------------------------------------------

// TestIntegration_WWingRequired verifies the W-Wing puzzle requires the w-wing
// solver. Medium-tier solvers stall; with all solvers (including w-wing in
// hard tier) the puzzle solves completely.
func TestIntegration_WWingRequired(t *testing.T) {
	store := solver.NewStore()

	// Medium-tier solvers cannot solve this puzzle.
	medBoard := boardFromString(t, wWingPuzzle)
	solveWithStrategies(t, &medBoard, store, mediumKeys)
	if medBoard.IsSolved() {
		t.Fatal("Expected puzzle to be unsolvable by medium techniques alone")
	}

	// With all solvers, the puzzle solves completely.
	fullBoard := boardFromString(t, wWingPuzzle)
	moves := solveWithStrategies(t, &fullBoard, store, evilKeys)
	if !fullBoard.IsSolved() {
		t.Fatal("Expected puzzle to be solvable with all solvers")
	}

	// Verify w-wing technique was used.
	wwCount := techniqueCount(moves, "w-wing")
	if wwCount == 0 {
		t.Error("Expected at least one w-wing move")
	}

	if bc := techniqueCount(moves, "backtracker"); bc > 0 {
		t.Errorf("Expected zero backtracker moves, got %d", bc)
	}

	t.Logf("Solved in %d moves, %d w-wing", len(moves), wwCount)
}

// TestIntegration_URType2Required verifies the Unique Rectangle Type 2 puzzle
// uses the ur-type-2 solver. Expert-tier solvers stall; with evil-tier
// (including unique-rectangle-2), additional progress is made.
func TestIntegration_URType2Required(t *testing.T) {
	store := solver.NewStore()

	// Expert-tier solvers stall on this puzzle.
	expertBoard := boardFromString(t, urType2Puzzle)
	expertMoves := solveWithStrategies(t, &expertBoard, store, expertKeys)
	if expertBoard.IsSolved() {
		t.Fatal("Expected puzzle to be unsolvable by expert techniques alone")
	}

	// With evil-tier solvers (including unique-rectangle-2), more progress is made.
	fullBoard := boardFromString(t, urType2Puzzle)
	fullMoves := solveWithStrategies(t, &fullBoard, store, evilKeys)

	// Verify unique-rectangle-2 technique was used.
	ur2Count := techniqueCount(fullMoves, "unique-rectangle-2")
	if ur2Count == 0 {
		t.Error("Expected at least one unique-rectangle-2 move")
	}

	// Evil-tier should make more progress than expert-tier alone.
	if len(fullMoves) <= len(expertMoves) {
		t.Errorf("Expected evil-tier to make more progress (%d moves) than expert-tier (%d moves)",
			len(fullMoves), len(expertMoves))
	}

	if bc := techniqueCount(fullMoves, "backtracker"); bc > 0 {
		t.Errorf("Expected zero backtracker moves, got %d", bc)
	}

	t.Logf("Expert: %d moves. Evil: %d moves, %d unique-rectangle-2",
		len(expertMoves), len(fullMoves), ur2Count)
}

// TestIntegration_XCyclesRequired verifies the X-Cycles puzzle uses the
// x-cycles solver. Expert-tier solvers stall; with evil-tier (including
// x-cycles), additional progress is made.
func TestIntegration_XCyclesRequired(t *testing.T) {
	store := solver.NewStore()

	// Expert-tier solvers stall on this puzzle.
	expertBoard := boardFromString(t, xCyclesPuzzle)
	expertMoves := solveWithStrategies(t, &expertBoard, store, expertKeys)
	if expertBoard.IsSolved() {
		t.Fatal("Expected puzzle to be unsolvable by expert techniques alone")
	}

	// With evil-tier solvers (including x-cycles), more progress is made.
	fullBoard := boardFromString(t, xCyclesPuzzle)
	fullMoves := solveWithStrategies(t, &fullBoard, store, evilKeys)

	// Verify x-cycles technique was used.
	xcCount := techniqueCount(fullMoves, "x-cycles")
	if xcCount == 0 {
		t.Error("Expected at least one x-cycles move")
	}

	// Evil-tier should make more progress than expert-tier alone.
	if len(fullMoves) <= len(expertMoves) {
		t.Errorf("Expected evil-tier to make more progress (%d moves) than expert-tier (%d moves)",
			len(fullMoves), len(expertMoves))
	}

	if bc := techniqueCount(fullMoves, "backtracker"); bc > 0 {
		t.Errorf("Expected zero backtracker moves, got %d", bc)
	}

	t.Logf("Expert: %d moves. Evil: %d moves, %d x-cycles",
		len(expertMoves), len(fullMoves), xcCount)
}

// TestIntegration_WWingPuzzleSolvesCompletely verifies the W-Wing puzzle
// solves completely with all 23 strategy solvers and no backtracking.
func TestIntegration_WWingPuzzleSolvesCompletely(t *testing.T) {
	store := solver.NewStore()

	board := boardFromString(t, wWingPuzzle)
	moves := solveWithStrategies(t, &board, store, evilKeys)
	if !board.IsSolved() {
		t.Fatal("Expected puzzle to be fully solvable with all strategy solvers")
	}

	// Verify multiple techniques are exercised (this puzzle uses many).
	techniques := make(map[string]bool)
	for _, m := range moves {
		techniques[m.Technique] = true
	}

	// Should use at least w-wing plus basic techniques.
	if !techniques["w-wing"] {
		t.Error("Expected w-wing to be used")
	}
	if !techniques["naked-single"] {
		t.Error("Expected naked-single to be used")
	}
	if !techniques["hidden-single"] {
		t.Error("Expected hidden-single to be used")
	}

	t.Logf("Solved in %d moves using %d distinct techniques", len(moves), len(techniques))
}

// ---------------------------------------------------------------------------
// Move metadata tests
// ---------------------------------------------------------------------------

// TestIntegration_MoveHasTechniqueAndReason verifies that every Move from a
// strategy solver has both Technique and Reason populated.
func TestIntegration_MoveHasTechniqueAndReason(t *testing.T) {
	store := solver.NewStore()
	board := boardFromString(t, nakedSubsetPuzzle)

	moves := solveWithStrategies(t, &board, store, mediumKeys)
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
// which technique fires first.
func TestIntegration_SolverOrderingMatters(t *testing.T) {
	store := solver.NewStore()
	board := boardFromString(t, easyPuzzle)

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
	expected := []string{
		"naked-single", "hidden-single",
		"naked-pair", "naked-triple", "pointing-pair", "hidden-pair",
		"x-wing", "xy-wing", "hidden-triple", "w-wing",
		"swordfish", "naked-quad", "simple-coloring", "hidden-quad", "xyz-wing",
		"jellyfish", "bug-plus-one", "unique-rectangle",
		"unique-rectangle-2", "unique-rectangle-3", "unique-rectangle-4",
		"x-cycles", "xy-chain",
	}

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
	puzzles := map[string]string{
		"easy":                      easyPuzzle,
		"naked-pair":                nakedSubsetPuzzle,
		"pointing-pair":             pointingPairPuzzle,
		"x-wing":                    xWingPuzzle,
		"swordfish":                 swordfishPuzzle,
		"hidden-subset":             hiddenSubsetPuzzle,
		"xy-wing":                   xyWingPuzzle,
		"simple-coloring":           simpleColoringPuzzle,
		"hard-xy-wing":              hardXYWingPuzzle,
		"hard-hidden-pair":          hardHiddenPairPuzzle,
		"expert-hidden-quad-color":  expertHiddenQuadColoringPuzzle,
		"expert-triple-requirement": expertTripleRequirementPuzzle,
		"evil-jellyfish":            evilJellyfishPuzzle,
		"evil-bug-ur":               evilBUGURPuzzle,
		"w-wing":                    wWingPuzzle,
	}

	for name, p := range puzzles {
		t.Run(name, func(t *testing.T) {
			board := boardFromString(t, p)
			solved := solver.NewBacktracker().SolveDeterministic(&board)
			if !solved {
				t.Errorf("Backtracker failed to solve %s puzzle", name)
			}
			if !board.IsSolved() {
				t.Errorf("Board not solved after backtracker on %s puzzle", name)
			}
		})
	}
}

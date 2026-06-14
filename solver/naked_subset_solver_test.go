package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// TestNakedSubsetSolver_FindsPair builds a board with a naked pair in a row
// that eliminates candidates from another cell, leaving it with one candidate.
func TestNakedSubsetSolver_FindsPair(t *testing.T) {
	board := core.NewEmptyBoard()

	// Set up row 0 so that:
	// - Cells (0,0) and (0,1) form a naked pair with candidates {1, 2}
	// - Cell (0,2) has candidates {1, 2, 3} → after elimination becomes {3}
	//
	// Fill columns to restrict candidates:
	// Col 0: fill rows 1-8 with 3,4,5,6,7,8,9 → (0,0) candidates = {1,2} (minus row 0 values)
	// Col 1: fill rows 1-8 with 3,4,5,6,7,8,9 → (0,1) candidates = {1,2}
	// Col 2: fill rows 1-8 with 4,5,6,7,8,9 → (0,2) candidates = {1,2,3}
	// Fill rest of row 0 with 4,5,6,7,8,9 to lock out those candidates.

	// Row 0: fill (0,3)-(0,8) with 4,5,6,7,8,9
	rowVals := []int{4, 5, 6, 7, 8, 9}
	for i, v := range rowVals {
		_ = board.Set(core.NewPosition(0, 3+i), v)
	}

	// Col 0: fill rows 1-8 with values that exclude 1,2 from those rows.
	// We want (0,0) to have candidates {1,2}, so we need all of 3-9 present
	// in col 0's rows 1-8 (but we already placed 4-9 in row 0, cols 3-8).
	// For col 0 rows 1-7: place 3,4,5,6,7,8,9
	col0Vals := []int{3, 4, 5, 6, 7, 8, 9}
	for i, v := range col0Vals {
		_ = board.Set(core.NewPosition(1+i, 0), v)
	}

	// Col 1: fill rows 1-7 with 3,4,5,6,7,8,9
	col1Vals := []int{3, 4, 5, 6, 7, 8, 9}
	for i, v := range col1Vals {
		_ = board.Set(core.NewPosition(1+i, 1), v)
	}

	// Col 2: fill rows 1-7 with 4,5,6,7,8,9 (leaving 1,2,3 as candidates for (0,2))
	col2Vals := []int{4, 5, 6, 7, 8, 9}
	for i, v := range col2Vals {
		_ = board.Set(core.NewPosition(1+i, 2), v)
	}

	// Now also need to ensure box constraints don't interfere.
	// Box 0 (rows 0-2, cols 0-2): cells filled are (0,0)=empty, (0,1)=empty, (0,2)=empty
	// (1,0)=3, (1,1)=3 — wait, we can't have duplicate in row 1.
	// This simple approach may create invalid boards. Let me use a more controlled setup.

	// Start fresh with a carefully constructed board.
	board = core.NewEmptyBoard()

	// We'll set up a valid partial board where row 0 has a naked pair.
	// Leave (0,0), (0,1), (0,2) empty. Fill (0,3)-(0,8) with 4,5,6,7,8,9.
	for i, v := range []int{4, 5, 6, 7, 8, 9} {
		_ = board.Set(core.NewPosition(0, 3+i), v)
	}

	// Now we need (0,0) candidates = {1,2}, (0,1) candidates = {1,2}, (0,2) candidates = {1,2,3}.
	// To get (0,0) = {1,2}: need 3 present in col 0 or box 0.
	// Place 3 at (1,0).
	_ = board.Set(core.NewPosition(1, 0), 3)

	// To get (0,1) = {1,2}: need 3 present in col 1 or box 0.
	// 3 is already in box 0 from (1,0), so (0,1) loses 3 → candidates = {1,2}. ✓

	// To get (0,2) = {1,2,3}: 3 should NOT be eliminated for (0,2).
	// (0,2) is in box 0 (rows 0-2, cols 0-2). 3 is at (1,0) in box 0, so (0,2) also loses 3. ✗
	// Need to use a different approach: place (0,2) in a different box.
	// Hmm, (0,2) is in box 0 (cols 0-2). All three cells are in box 0.

	// Alternative: use columns 6,7,8 so they're in box 2 (cols 6-8).
	board = core.NewEmptyBoard()

	// Row 0: fill (0,0)-(0,5) with 4,5,6,7,8,9. Leave (0,6),(0,7),(0,8) empty.
	for i, v := range []int{4, 5, 6, 7, 8, 9} {
		_ = board.Set(core.NewPosition(0, i), v)
	}

	// (0,6),(0,7),(0,8) are in box 2 (rows 0-2, cols 6-8).
	// Row 0 already has 4-9, so candidates for all three are subsets of {1,2,3}.

	// To make (0,6) = {1,2}: need 3 eliminated from col 6 or box 2.
	_ = board.Set(core.NewPosition(1, 6), 3) // 3 in col 6 → (0,6) loses 3

	// To make (0,7) = {1,2}: need 3 eliminated from col 7 or box 2.
	// 3 is already in box 2 at (1,6), so (0,7) also loses 3.

	// To make (0,8) = {1,2,3}: need 3 NOT eliminated from col 8 and NOT from box 2.
	// But 3 is in box 2 at (1,6). So (0,8) loses 3 too. ✗

	// OK, let's try a different approach: use rows instead.
	// Place the pair in a column. Columns are simpler to control.
	board = core.NewEmptyBoard()

	// Fill col 0: (1,0)-(8,0) with 4,5,6,7,8,9. Leave (0,0) empty.
	// Actually, we need 3 cells empty in a column to test the pair.
	// Fill col 0 rows 3-8 with 4,5,6,7,8,9. Leave (0,0),(1,0),(2,0) empty.
	for i, v := range []int{4, 5, 6, 7, 8, 9} {
		_ = board.Set(core.NewPosition(3+i, 0), v)
	}

	// Now candidates for (0,0),(1,0),(2,0) in col 0 are subsets of {1,2,3}.
	// (0,0) is in box 0 (rows 0-2, cols 0-2).
	// (1,0) is in box 0.
	// (2,0) is in box 0.

	// To restrict (0,0) to {1,2}: place 3 in row 0 (not col 0).
	_ = board.Set(core.NewPosition(0, 1), 3) // 3 in row 0 → (0,0) loses 3

	// To restrict (1,0) to {1,2}: place 3 in row 1 (not col 0, and not already in box 0).
	// 3 is already in box 0 at (0,1), so (1,0) already loses 3.

	// (2,0) should keep {1,2,3}: 3 is in box 0 at (0,1), so (2,0) also loses 3. ✗

	// The problem is box 0 is always shared. Let me use cells from different boxes.
	// Column 0: (0,0) is box 0, (3,0) is box 3, (6,0) is box 6.
	board = core.NewEmptyBoard()

	// Fill col 0 with all except positions 0, 3, 6.
	// (1,0)=4, (2,0)=5, (4,0)=6, (5,0)=7, (7,0)=8, (8,0)=9
	_ = board.Set(core.NewPosition(1, 0), 4)
	_ = board.Set(core.NewPosition(2, 0), 5)
	_ = board.Set(core.NewPosition(4, 0), 6)
	_ = board.Set(core.NewPosition(5, 0), 7)
	_ = board.Set(core.NewPosition(7, 0), 8)
	_ = board.Set(core.NewPosition(8, 0), 9)

	// Now col 0 has 4,5,6,7,8,9 → (0,0),(3,0),(6,0) candidates ⊆ {1,2,3}

	// Restrict (0,0) to {1,2}: place 3 in row 0.
	_ = board.Set(core.NewPosition(0, 1), 3)

	// Restrict (3,0) to {1,2}: place 3 in row 3.
	_ = board.Set(core.NewPosition(3, 1), 3)

	// (6,0) should have {1,2,3}: no 3 in row 6 or box 6 (rows 6-8, cols 0-2).
	// Row 6: no 3. Box 6: (7,0)=8, (8,0)=9 — no 3. ✓
	// So (6,0) candidates = {1,2,3}

	// Naked pair: (0,0)={1,2} and (3,0)={1,2} in col 0.
	// Elimination: remove 1,2 from (6,0) → {3} → naked single!

	s := solver.NewNakedSubsetSolver()
	move := s.Apply(&board)

	if move == nil {
		t.Fatal("Expected NakedSubsetSolver to find a move, got nil")
	}

	if move.Cell.Position != core.NewPosition(6, 0) {
		t.Errorf("Expected move at (7, 1), got %s", move.Cell.Position.ToString())
	}

	if move.Cell.Value != 3 {
		t.Errorf("Expected value 3, got %d", move.Cell.Value)
	}

	if move.Technique != "naked-subset" {
		t.Errorf("Expected technique 'naked-subset', got %q", move.Technique)
	}
}

// TestNakedSubsetSolver_NoProgress verifies nil when no naked subsets create singles.
func TestNakedSubsetSolver_NoProgress(t *testing.T) {
	board := core.NewEmptyBoard()
	s := solver.NewNakedSubsetSolver()

	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil for empty board, got %v", move)
	}
}

// TestNakedSubsetSolver_SolvedBoard verifies nil for a fully solved board.
func TestNakedSubsetSolver_SolvedBoard(t *testing.T) {
	board := core.NewEmptyBoard()
	bt := solver.NewBacktracker()
	bt.Solve(&board)

	s := solver.NewNakedSubsetSolver()
	move := s.Apply(&board)
	if move != nil {
		t.Errorf("Expected nil for solved board, got %v", move)
	}
}

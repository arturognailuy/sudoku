package core

import "testing"

func TestCandidateSetBasicOperations(t *testing.T) {
	var cs CandidateSet

	// Empty set.
	if !cs.IsEmpty() {
		t.Error("new CandidateSet should be empty")
	}
	if cs.Count() != 0 {
		t.Errorf("expected count 0, got %d", cs.Count())
	}

	// Add and Has.
	cs.Add(3)
	cs.Add(7)
	if !cs.Has(3) {
		t.Error("expected Has(3) true")
	}
	if !cs.Has(7) {
		t.Error("expected Has(7) true")
	}
	if cs.Has(5) {
		t.Error("expected Has(5) false")
	}
	if cs.Count() != 2 {
		t.Errorf("expected count 2, got %d", cs.Count())
	}

	// Remove.
	cs.Remove(3)
	if cs.Has(3) {
		t.Error("expected Has(3) false after Remove")
	}
	if cs.Count() != 1 {
		t.Errorf("expected count 1, got %d", cs.Count())
	}
}

func TestCandidateSetSingle(t *testing.T) {
	var cs CandidateSet
	cs.Add(5)

	v, ok := cs.Single()
	if !ok || v != 5 {
		t.Errorf("expected Single() = (5, true), got (%d, %v)", v, ok)
	}

	cs.Add(9)
	_, ok = cs.Single()
	if ok {
		t.Error("expected Single() = (_, false) for two candidates")
	}
}

func TestCandidateSetValues(t *testing.T) {
	var cs CandidateSet
	cs.Add(1)
	cs.Add(4)
	cs.Add(9)

	vals := cs.Values()
	expected := []int{1, 4, 9}
	if len(vals) != len(expected) {
		t.Fatalf("expected %d values, got %d", len(expected), len(vals))
	}
	for i, v := range vals {
		if v != expected[i] {
			t.Errorf("Values()[%d] = %d, expected %d", i, v, expected[i])
		}
	}
}

func TestAllCandidates(t *testing.T) {
	cs := allCandidates
	if cs.Count() != 9 {
		t.Errorf("allCandidates should have 9 candidates, got %d", cs.Count())
	}
	for v := 1; v <= 9; v++ {
		if !cs.Has(v) {
			t.Errorf("allCandidates should have value %d", v)
		}
	}
	if cs.Has(0) {
		t.Error("allCandidates should not have value 0")
	}
}

func TestBoardCandidatesOnEmptyBoard(t *testing.T) {
	board := NewEmptyBoard()

	// Every cell should have all 9 candidates.
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			cs := board.Candidates(NewPosition(r, c))
			if cs.Count() != 9 {
				t.Errorf("empty board cell (%d,%d) should have 9 candidates, got %d", r, c, cs.Count())
			}
		}
	}
}

func TestBoardCandidatesAfterSet(t *testing.T) {
	board := NewEmptyBoard()

	// Place 5 at (0,0).
	_ = board.Set(NewPosition(0, 0), 5)

	// (0,0) should have no candidates (it's filled).
	if cs := board.Candidates(NewPosition(0, 0)); !cs.IsEmpty() {
		t.Errorf("filled cell (0,0) should have no candidates, got %d", cs.Count())
	}

	// Row peer (0,4) should not have 5 as candidate.
	if board.Candidates(NewPosition(0, 4)).Has(5) {
		t.Error("row peer (0,4) should not have candidate 5")
	}

	// Column peer (4,0) should not have 5 as candidate.
	if board.Candidates(NewPosition(4, 0)).Has(5) {
		t.Error("column peer (4,0) should not have candidate 5")
	}

	// Box peer (1,1) should not have 5 as candidate.
	if board.Candidates(NewPosition(1, 1)).Has(5) {
		t.Error("box peer (1,1) should not have candidate 5")
	}

	// Non-peer (4,4) should still have 5 as candidate.
	if !board.Candidates(NewPosition(4, 4)).Has(5) {
		t.Error("non-peer (4,4) should still have candidate 5")
	}
}

func TestBoardCandidatesAfterUnset(t *testing.T) {
	board := NewEmptyBoard()

	// Place 3 at (0,0), then unset it.
	_ = board.Set(NewPosition(0, 0), 3)
	board.Unset(NewPosition(0, 0))

	// (0,0) should have all 9 candidates again.
	if cs := board.Candidates(NewPosition(0, 0)); cs.Count() != 9 {
		t.Errorf("after unset, cell (0,0) should have 9 candidates, got %d", cs.Count())
	}

	// Row peer should have 3 as candidate again.
	if !board.Candidates(NewPosition(0, 4)).Has(3) {
		t.Error("after unset, row peer (0,4) should have candidate 3 again")
	}
}

func TestBoardCandidatesConstraintPropagation(t *testing.T) {
	board := NewEmptyBoard()

	// Fill the first row with values 1–8, leaving (0,8) empty.
	for c := 0; c < 8; c++ {
		_ = board.Set(NewPosition(0, c), c+1)
	}

	// (0,8) should have exactly one candidate: 9.
	cs := board.Candidates(NewPosition(0, 8))
	v, ok := cs.Single()
	if !ok || v != 9 {
		t.Errorf("(0,8) should have single candidate 9, got (%d, %v), count=%d", v, ok, cs.Count())
	}
}

func TestEmptyPositions(t *testing.T) {
	board := NewEmptyBoard()

	// All 81 cells should be empty.
	empty := board.EmptyPositions()
	if len(empty) != 81 {
		t.Errorf("empty board should have 81 empty positions, got %d", len(empty))
	}

	// Set one cell.
	_ = board.Set(NewPosition(0, 0), 1)
	empty = board.EmptyPositions()
	if len(empty) != 80 {
		t.Errorf("board with 1 filled cell should have 80 empty positions, got %d", len(empty))
	}
}

func TestForEachCell(t *testing.T) {
	board := NewEmptyBoard()
	_ = board.Set(NewPosition(0, 0), 5)
	_ = board.Set(NewPosition(8, 8), 3)

	count := 0
	board.ForEachCell(func(pos Position, value int) bool {
		count++
		return true
	})
	if count != 81 {
		t.Errorf("ForEachCell should visit 81 cells, visited %d", count)
	}

	// Test early exit.
	earlyCount := 0
	board.ForEachCell(func(pos Position, value int) bool {
		earlyCount++
		return earlyCount < 10
	})
	if earlyCount != 10 {
		t.Errorf("ForEachCell with early exit should visit 10 cells, visited %d", earlyCount)
	}
}

func TestCandidatesConsistencyWithIsValidInput(t *testing.T) {
	// Build a partial board and verify candidates match IsValidInput.
	board := NewEmptyBoard()
	_ = board.Set(NewPosition(0, 0), 1)
	_ = board.Set(NewPosition(0, 3), 4)
	_ = board.Set(NewPosition(1, 1), 5)
	_ = board.Set(NewPosition(3, 0), 2)
	_ = board.Set(NewPosition(4, 4), 9)

	// For every empty cell, verify each candidate matches IsValidInput.
	for _, pos := range board.EmptyPositions() {
		cs := board.Candidates(pos)
		for v := 1; v <= 9; v++ {
			hasCand := cs.Has(v)
			isValid := board.IsValidInput(pos, v)
			if hasCand != isValid {
				t.Errorf("candidate mismatch at %s for value %d: Candidates.Has=%v, IsValidInput=%v",
					pos.ToString(), v, hasCand, isValid)
			}
		}
	}
}

func TestCandidatesAfterFromString(t *testing.T) {
	board := NewEmptyBoard()
	// A known partial board string (17 clues minimum sudoku pattern).
	input := "1.......2.3...4...5.......6...7.8...........9.1...........2.3...........4.5......"
	board.FromString(input)

	// Verify candidates are consistent.
	for _, pos := range board.EmptyPositions() {
		cs := board.Candidates(pos)
		for v := 1; v <= 9; v++ {
			hasCand := cs.Has(v)
			isValid := board.IsValidInput(pos, v)
			if hasCand != isValid {
				t.Errorf("FromString: candidate mismatch at %s for value %d: Candidates.Has=%v, IsValidInput=%v",
					pos.ToString(), v, hasCand, isValid)
			}
		}
	}
}

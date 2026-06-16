package solver

import (
	"fmt"
	"strings"

	"github.com/gnailuy/sudoku/core"
)

// NakedSubsetSolver finds naked pairs and naked triples in rows, columns, and boxes.
//
// A naked pair occurs when two cells in a unit share the same two candidates —
// those candidates can be eliminated from all other cells in the unit.
//
// A naked triple occurs when three cells in a unit collectively have exactly
// three candidates — those candidates can be eliminated from all other cells
// in the unit.
//
// This solver does not place values directly. Instead, it finds eliminations
// and then checks if any elimination creates a naked single (a cell with one
// candidate remaining), returning that as the move.
type NakedSubsetSolver struct {
	Base
}

// nakedSubsetCell holds an empty cell's position and its candidate set.
type nakedSubsetCell struct {
	pos        core.Position
	candidates core.CandidateSet
}

// NewNakedSubsetSolver creates a NakedSubsetSolver and returns it.
func NewNakedSubsetSolver() *NakedSubsetSolver {
	return &NakedSubsetSolver{
		Base: Base{
			Key:         "naked-subset",
			DisplayName: "Naked Pairs/Triples",
			Description: "Finds two or three cells in a unit sharing the same candidates, enabling eliminations that reveal a single candidate.",
			Weight:      WeightNakedSubset,
		},
	}
}

// Apply scans all units for naked pairs and triples. When eliminations from a
// naked subset reduce a cell to a single candidate, that cell is returned as
// the move.
func (s *NakedSubsetSolver) Apply(board *core.Board) *Move {
	units := allUnits()

	for _, u := range units {
		if move := s.findNakedSubset(board, u.positions, u.name); move != nil {
			return move
		}
	}

	return nil
}

// findNakedSubset checks a single unit for naked pairs and triples.
func (s *NakedSubsetSolver) findNakedSubset(board *core.Board, positions []core.Position, unitName string) *Move {
	// Collect empty cells and their candidates.
	var emptyCells []nakedSubsetCell
	for _, pos := range positions {
		cands := board.Candidates(pos)
		if !cands.IsEmpty() {
			emptyCells = append(emptyCells, nakedSubsetCell{pos: pos, candidates: cands})
		}
	}

	// Try naked pairs (size 2).
	if move := s.trySubsetSize(board, emptyCells, positions, unitName, 2); move != nil {
		return move
	}

	// Try naked triples (size 3).
	if move := s.trySubsetSize(board, emptyCells, positions, unitName, 3); move != nil {
		return move
	}

	return nil
}

// trySubsetSize looks for naked subsets of the given size in the unit.
func (s *NakedSubsetSolver) trySubsetSize(board *core.Board, emptyCells []nakedSubsetCell, allPositions []core.Position, unitName string, size int) *Move {
	// Filter cells with at most 'size' candidates.
	var eligible []nakedSubsetCell
	for _, c := range emptyCells {
		if c.candidates.Count() <= size {
			eligible = append(eligible, c)
		}
	}

	if len(eligible) < size {
		return nil
	}

	// Generate combinations of 'size' cells from eligible.
	indices := make([]int, size)
	return s.combinationsSearch(eligible, allPositions, unitName, size, indices, 0, 0, board)
}

// combinationsSearch recursively generates combinations and checks for naked subsets.
func (s *NakedSubsetSolver) combinationsSearch(eligible []nakedSubsetCell, allPositions []core.Position, unitName string, size int, indices []int, start int, depth int, board *core.Board) *Move {
	if depth == size {
		// Compute the union of candidates in the selected cells.
		var union core.CandidateSet
		for i := 0; i < size; i++ {
			union |= eligible[indices[i]].candidates
		}

		// A naked subset: 'size' cells whose combined candidates have exactly 'size' values.
		if union.Count() != size {
			return nil
		}

		// Build a set of positions in the subset for quick lookup.
		subsetPositions := make(map[core.Position]bool, size)
		for i := 0; i < size; i++ {
			subsetPositions[eligible[indices[i]].pos] = true
		}

		// Check if any non-subset cell in the unit has candidates that can be eliminated.
		unionVals := union.Values()
		for _, pos := range allPositions {
			if subsetPositions[pos] || board.Get(pos) != 0 {
				continue
			}

			cands := board.Candidates(pos)
			// Check if this cell has any of the subset's candidates.
			hasOverlap := false
			for _, v := range unionVals {
				if cands.Has(v) {
					hasOverlap = true
					break
				}
			}
			if !hasOverlap {
				continue
			}

			// Elimination possible. Simulate the elimination to see if it
			// creates a naked single.
			reduced := cands
			for _, rv := range unionVals {
				reduced.Remove(rv)
			}

			if reduced.Count() == 1 {
				value := reduced.Values()[0]
				subsetName := s.subsetName(size)
				cellNames := s.cellNames(eligible, indices, size)
				valNames := s.valNames(unionVals)
				return &Move{
					Cell:      core.NewCell(pos, value),
					Technique: s.Key,
					Reason: fmt.Sprintf(
						"%s {%s} in %s at {%s} eliminates candidates, leaving %d as the only candidate for %s",
						subsetName, valNames, unitName, cellNames, value, pos.ToString(),
					),
				}
			}
		}

		return nil
	}

	for i := start; i < len(eligible); i++ {
		indices[depth] = i
		if move := s.combinationsSearch(eligible, allPositions, unitName, size, indices, i+1, depth+1, board); move != nil {
			return move
		}
	}

	return nil
}

// subsetName returns "Naked pair" or "Naked triple" based on size.
func (s *NakedSubsetSolver) subsetName(size int) string {
	switch size {
	case 2:
		return "Naked pair"
	case 3:
		return "Naked triple"
	default:
		return fmt.Sprintf("Naked subset (%d)", size)
	}
}

// cellNames formats the positions of the subset cells.
func (s *NakedSubsetSolver) cellNames(eligible []nakedSubsetCell, indices []int, size int) string {
	names := make([]string, size)
	for i := 0; i < size; i++ {
		names[i] = eligible[indices[i]].pos.ToString()
	}
	return strings.Join(names, ", ")
}

// valNames formats a slice of values as a comma-separated string.
func (s *NakedSubsetSolver) valNames(vals []int) string {
	names := make([]string, len(vals))
	for i, v := range vals {
		names[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(names, ", ")
}

package solver

import (
	"fmt"
	"strings"

	"github.com/gnailuy/sudoku/core"
)

// HiddenSubsetSolver finds hidden pairs and hidden triples in rows, columns,
// and boxes.
//
// A hidden pair occurs when two candidates appear in only two cells within a
// unit. All other candidates can be eliminated from those two cells. If any
// elimination leaves a cell with one candidate, that is the move.
//
// A hidden triple occurs when three candidates appear in only three cells
// within a unit. All other candidates can be eliminated from those three
// cells.
//
// Hidden subsets are the complement of naked subsets: naked subsets find cells
// whose candidates are restricted to a set, while hidden subsets find
// candidates whose positions are restricted to a set of cells.
type HiddenSubsetSolver struct {
	Base
}

// NewHiddenSubsetSolver creates a HiddenSubsetSolver and returns it.
func NewHiddenSubsetSolver() *HiddenSubsetSolver {
	return &HiddenSubsetSolver{
		Base: Base{
			Key:         "hidden-subset",
			DisplayName: "Hidden Pairs/Triples",
			Description: "Finds two or three candidates confined to the same cells in a unit, enabling eliminations that reveal a single candidate.",
		},
	}
}

// Apply scans all units for hidden pairs and triples.
func (s *HiddenSubsetSolver) Apply(board *core.Board) *Move {
	units := allUnits()

	for _, u := range units {
		if move := s.findHiddenSubset(board, u.positions, u.name); move != nil {
			return move
		}
	}

	return nil
}

// digitPositions maps a digit to the positions of empty cells that contain it
// as a candidate within a unit.
type digitPositions struct {
	digit     int
	positions []core.Position
}

// findHiddenSubset checks a single unit for hidden pairs and triples.
func (s *HiddenSubsetSolver) findHiddenSubset(board *core.Board, positions []core.Position, unitName string) *Move {
	// Build a map of digit → positions of empty cells containing that digit.
	var digits []digitPositions
	for d := 1; d <= 9; d++ {
		var cells []core.Position
		for _, pos := range positions {
			if board.Get(pos) == 0 && board.Candidates(pos).Has(d) {
				cells = append(cells, pos)
			}
		}
		// Only digits that appear in 2 or 3 cells are candidates for hidden subsets.
		if len(cells) >= 2 && len(cells) <= 3 {
			digits = append(digits, digitPositions{digit: d, positions: cells})
		}
	}

	// Try hidden pairs (size 2).
	if move := s.tryHiddenSize(board, digits, unitName, 2); move != nil {
		return move
	}

	// Try hidden triples (size 3).
	if move := s.tryHiddenSize(board, digits, unitName, 3); move != nil {
		return move
	}

	return nil
}

// tryHiddenSize looks for hidden subsets of the given size.
func (s *HiddenSubsetSolver) tryHiddenSize(board *core.Board, digits []digitPositions, unitName string, size int) *Move {
	if len(digits) < size {
		return nil
	}

	// Generate combinations of 'size' digits from the eligible digits.
	indices := make([]int, size)
	return s.combinationsSearch(board, digits, unitName, size, indices, 0, 0)
}

// combinationsSearch recursively generates combinations and checks for hidden subsets.
func (s *HiddenSubsetSolver) combinationsSearch(board *core.Board, digits []digitPositions, unitName string, size int, indices []int, start int, depth int) *Move {
	if depth == size {
		// Compute the union of positions for the selected digits.
		posSet := make(map[core.Position]bool)
		for i := 0; i < size; i++ {
			for _, pos := range digits[indices[i]].positions {
				posSet[pos] = true
			}
		}

		// A hidden subset: 'size' digits whose combined positions have exactly 'size' cells.
		if len(posSet) != size {
			return nil
		}

		// Build the set of digits in the hidden subset.
		hiddenDigits := make(map[int]bool, size)
		for i := 0; i < size; i++ {
			hiddenDigits[digits[indices[i]].digit] = true
		}

		// For each cell in the subset, check if removing non-hidden candidates
		// creates a naked single.
		for pos := range posSet {
			cands := board.Candidates(pos)
			// Count how many candidates are NOT in the hidden set.
			hasExtra := false
			for _, v := range cands.Values() {
				if !hiddenDigits[v] {
					hasExtra = true
					break
				}
			}
			if !hasExtra {
				continue // No elimination possible in this cell.
			}

			// Simulate the elimination: keep only the hidden subset digits.
			var reduced core.CandidateSet
			for d := range hiddenDigits {
				if cands.Has(d) {
					reduced.Add(d)
				}
			}

			if reduced.Count() == 1 {
				value := reduced.Values()[0]
				subsetName := s.subsetName(size)
				digitNames := s.digitNames(digits, indices, size)
				cellNames := s.cellNames(posSet)
				return &Move{
					Cell:      core.NewCell(pos, value),
					Technique: s.Key,
					Reason: fmt.Sprintf(
						"%s {%s} in %s confined to {%s}, leaving %d as the only candidate for %s",
						subsetName, digitNames, unitName, cellNames, value, pos.ToString(),
					),
				}
			}
		}

		return nil
	}

	for i := start; i < len(digits); i++ {
		indices[depth] = i
		if move := s.combinationsSearch(board, digits, unitName, size, indices, i+1, depth+1); move != nil {
			return move
		}
	}

	return nil
}

// subsetName returns "Hidden pair" or "Hidden triple" based on size.
func (s *HiddenSubsetSolver) subsetName(size int) string {
	switch size {
	case 2:
		return "Hidden pair"
	case 3:
		return "Hidden triple"
	default:
		return fmt.Sprintf("Hidden subset (%d)", size)
	}
}

// digitNames formats the digits of the hidden subset.
func (s *HiddenSubsetSolver) digitNames(digits []digitPositions, indices []int, size int) string {
	names := make([]string, size)
	for i := 0; i < size; i++ {
		names[i] = fmt.Sprintf("%d", digits[indices[i]].digit)
	}
	return strings.Join(names, ", ")
}

// cellNames formats the positions of the subset cells.
func (s *HiddenSubsetSolver) cellNames(posSet map[core.Position]bool) string {
	names := make([]string, 0, len(posSet))
	for pos := range posSet {
		names = append(names, pos.ToString())
	}
	return strings.Join(names, ", ")
}

package solver

import (
	"fmt"
	"strings"

	"github.com/gnailuy/sudoku/core"
)

// NakedPairSolver finds naked pairs in rows, columns, and boxes.
//
// A naked pair occurs when two cells in a unit share the same two candidates —
// those candidates can be eliminated from all other cells in the unit.
//
// This solver does not place values directly. Apply retains the solving
// contract of returning a move when an elimination creates a naked single;
// Hint can expose the elimination itself for interactive teaching.
type NakedPairSolver struct {
	Base
}

// NewNakedPairSolver creates a NakedPairSolver and returns it.
func NewNakedPairSolver() *NakedPairSolver {
	return &NakedPairSolver{
		Base: Base{
			Key:         "naked-pair",
			DisplayName: "Naked Pair",
			Description: "Finds two cells in a unit sharing the same two candidates, enabling eliminations that reveal a single candidate.",
			Weight:      WeightNakedPair,
		},
	}
}

// Apply scans all units for naked pairs.
func (s *NakedPairSolver) Apply(board *core.Board) *Move {
	units := allUnits()

	for _, u := range units {
		if move := findNakedSubsetOfSize(board, u.positions, u.name, 2, s.Key, "Naked pair"); move != nil {
			return move
		}
	}

	return nil
}

// NakedTripleSolver finds naked triples in rows, columns, and boxes.
//
// A naked triple occurs when three cells in a unit collectively have exactly
// three candidates — those candidates can be eliminated from all other cells
// in the unit.
type NakedTripleSolver struct {
	Base
}

// NewNakedTripleSolver creates a NakedTripleSolver and returns it.
func NewNakedTripleSolver() *NakedTripleSolver {
	return &NakedTripleSolver{
		Base: Base{
			Key:         "naked-triple",
			DisplayName: "Naked Triple",
			Description: "Finds three cells in a unit sharing the same three candidates, enabling eliminations that reveal a single candidate.",
			Weight:      WeightNakedTriple,
		},
	}
}

// Apply scans all units for naked triples.
func (s *NakedTripleSolver) Apply(board *core.Board) *Move {
	units := allUnits()

	for _, u := range units {
		if move := findNakedSubsetOfSize(board, u.positions, u.name, 3, s.Key, "Naked triple"); move != nil {
			return move
		}
	}

	return nil
}

// Hint returns the first useful naked-triple elimination even when it does not
// immediately create a placement.
func (s *NakedTripleSolver) Hint(board *core.Board) *Move {
	if move := s.Apply(board); move != nil {
		return move
	}
	return findNakedSubsetHint(board, 3, s.Key, "Naked triple")
}

// NakedQuadSolver finds naked quads in rows, columns, and boxes.
//
// A naked quad occurs when four cells in a unit collectively have exactly
// four candidates — those candidates can be eliminated from all other cells
// in the unit.
type NakedQuadSolver struct {
	Base
}

// NewNakedQuadSolver creates a NakedQuadSolver and returns it.
func NewNakedQuadSolver() *NakedQuadSolver {
	return &NakedQuadSolver{
		Base: Base{
			Key:         "naked-quad",
			DisplayName: "Naked Quad",
			Description: "Finds four cells in a unit sharing the same four candidates, enabling eliminations that reveal a single candidate.",
			Weight:      WeightNakedQuad,
		},
	}
}

// Apply scans all units for naked quads.
func (s *NakedQuadSolver) Apply(board *core.Board) *Move {
	units := allUnits()

	for _, u := range units {
		if move := findNakedSubsetOfSize(board, u.positions, u.name, 4, s.Key, "Naked quad"); move != nil {
			return move
		}
	}

	return nil
}

// Hint returns the first useful naked-quad elimination even when it does not
// immediately create a placement.
func (s *NakedQuadSolver) Hint(board *core.Board) *Move {
	if move := s.Apply(board); move != nil {
		return move
	}
	return findNakedSubsetHint(board, 4, s.Key, "Naked quad")
}

// nakedSubsetCell holds an empty cell's position and its candidate set.
type nakedSubsetCell struct {
	pos        core.Position
	candidates core.CandidateSet
}

// findNakedSubsetOfSize is the shared logic for naked pair/triple/quad solvers.
// It searches a single unit for naked subsets of the given size.
func findNakedSubsetOfSize(
	board *core.Board,
	positions []core.Position,
	unitName string,
	size int,
	technique string,
	displayName string,
) *Move {
	// Collect empty cells and their candidates.
	var emptyCells []nakedSubsetCell
	for _, pos := range positions {
		cands := board.Candidates(pos)
		if !cands.IsEmpty() {
			emptyCells = append(emptyCells, nakedSubsetCell{pos: pos, candidates: cands})
		}
	}

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
	return nakedSubsetCombinationsSearch(
		eligible,
		positions,
		unitName,
		size,
		indices,
		0,
		0,
		board,
		technique,
		displayName,
	)
}

// nakedSubsetCombinationsSearch recursively generates combinations and checks for naked subsets.
func nakedSubsetCombinationsSearch(
	eligible []nakedSubsetCell,
	allPositions []core.Position,
	unitName string,
	size int,
	indices []int,
	start int,
	depth int,
	board *core.Board,
	technique string,
	displayName string,
) *Move {
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
				cellNames := nakedSubsetCellNames(eligible, indices, size)
				valNames := nakedSubsetValNames(unionVals)
				premises := make([]CandidateGroup, size)
				for i := 0; i < size; i++ {
					selected := eligible[indices[i]]
					premises[i] = CandidateGroup{Position: selected.pos, Values: selected.candidates.Values()}
				}
				return &Move{
					Cell:      core.NewCell(pos, value),
					Technique: technique,
					Reason: fmt.Sprintf(
						"%s {%s} in %s at {%s} eliminates candidates, leaving %d as the only candidate for %s",
						displayName, valNames, unitName, cellNames, value, pos.ToString(),
					),
					Evidence: &Evidence{Unit: parseNakedSubsetUnit(unitName), Premises: premises},
				}
			}
		}

		return nil
	}

	for i := start; i < len(eligible); i++ {
		indices[depth] = i
		if move := nakedSubsetCombinationsSearch(eligible, allPositions, unitName, size, indices, i+1, depth+1, board,
			technique, displayName); move != nil {
			return move
		}
	}

	return nil
}

// nakedSubsetCellNames formats the positions of the subset cells.
func nakedSubsetCellNames(eligible []nakedSubsetCell, indices []int, size int) string {
	names := make([]string, size)
	for i := 0; i < size; i++ {
		names[i] = eligible[indices[i]].pos.ToString()
	}
	return strings.Join(names, ", ")
}

// nakedSubsetValNames formats a slice of values as a comma-separated string.
func nakedSubsetValNames(vals []int) string {
	names := make([]string, len(vals))
	for i, v := range vals {
		names[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(names, ", ")
}

// Hint returns the first naked-pair elimination as a teaching deduction without
// requiring that the elimination immediately creates a placement.
func (s *NakedPairSolver) Hint(board *core.Board) *Move {
	return findNakedSubsetHint(board, 2, s.Key, "Naked pair")
}

func findNakedSubsetHint(board *core.Board, size int, technique, displayName string) *Move {
	for _, unit := range allUnits() {
		if move := findNakedSubsetElimination(board, unit.positions, unit.name, size, technique, displayName); move != nil {
			return move
		}
	}
	return nil
}

func findNakedSubsetElimination(
	board *core.Board,
	positions []core.Position,
	unitName string,
	size int,
	technique, displayName string,
) *Move {
	var eligible []nakedSubsetCell
	for _, position := range positions {
		candidates := board.Candidates(position)
		if count := candidates.Count(); count > 0 && count <= size {
			eligible = append(eligible, nakedSubsetCell{pos: position, candidates: candidates})
		}
	}
	if len(eligible) < size {
		return nil
	}
	indices := make([]int, size)
	return nakedSubsetEliminationSearch(
		board,
		eligible,
		positions,
		unitName,
		size,
		technique,
		displayName,
		indices,
		0,
		0,
	)
}

func nakedSubsetEliminationSearch(
	board *core.Board,
	eligible []nakedSubsetCell,
	positions []core.Position,
	unitName string,
	size int,
	technique, displayName string,
	indices []int,
	start, depth int,
) *Move {
	if depth < size {
		for index := start; index < len(eligible); index++ {
			indices[depth] = index
			if move := nakedSubsetEliminationSearch(board, eligible, positions, unitName, size, technique, displayName,
				indices, index+1, depth+1); move != nil {
				return move
			}
		}
		return nil
	}

	var union core.CandidateSet
	selected := make(map[core.Position]bool, size)
	premises := make([]CandidateGroup, size)
	for index := 0; index < size; index++ {
		cell := eligible[indices[index]]
		union |= cell.candidates
		selected[cell.pos] = true
		premises[index] = CandidateGroup{Position: cell.pos, Values: cell.candidates.Values()}
	}
	if union.Count() != size {
		return nil
	}

	values := union.Values()
	var eliminations []CandidateRef
	for _, position := range positions {
		if selected[position] || board.Get(position) != 0 {
			continue
		}
		candidates := board.Candidates(position)
		for _, value := range values {
			if candidates.Has(value) {
				eliminations = append(eliminations, CandidateRef{Position: position, Value: value})
			}
		}
	}
	if len(eliminations) == 0 {
		return nil
	}
	return &Move{
		Technique: technique,
		Reason: fmt.Sprintf(
			"%s {%s} in %s at {%s} eliminates candidates from the other cells",
			displayName,
			nakedSubsetValNames(values),
			unitName,
			nakedSubsetCellNames(eligible, indices, size),
		),
		EliminationOnly: true,
		Evidence: &Evidence{
			Unit:         parseNakedSubsetUnit(unitName),
			Premises:     premises,
			Eliminations: eliminations,
		},
	}
}

func parseNakedSubsetUnit(name string) *UnitRef {
	var kind string
	var number int
	if _, err := fmt.Sscanf(name, "%s %d", &kind, &number); err != nil || number < 1 || number > 9 {
		return nil
	}
	unit := UnitRef{Kind: UnitKind(kind), Index: number - 1}
	return &unit
}

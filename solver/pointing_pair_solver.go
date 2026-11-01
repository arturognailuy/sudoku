package solver

import (
	"fmt"

	"github.com/gnailuy/sudoku/core"
)

// PointingPairSolver finds pointing pairs and box-line reductions.
//
// Pointing pair: When a candidate within a box is confined to a single row
// or column, that candidate can be eliminated from the rest of that row/column
// outside the box.
//
// Box-line reduction (the inverse): When a candidate within a row or column
// is confined to a single box, that candidate can be eliminated from the rest
// of that box.
//
// Apply retains the solving contract of returning a move only when an
// elimination creates a naked single. Hint can also return the elimination
// itself so an interactive player sees this simpler deduction before a more
// advanced strategy that reaches the same result.
type PointingPairSolver struct {
	Base
}

// NewPointingPairSolver creates a PointingPairSolver and returns it.
func NewPointingPairSolver() *PointingPairSolver {
	return &PointingPairSolver{
		Base: Base{
			Key:         "pointing-pair",
			DisplayName: "Pointing Pairs / Box-Line Reduction",
			Description: ("Finds candidates confined to a single row or column within a box (or vice versa), " +
				"enabling eliminations that reveal a single candidate."),
			Weight: WeightPointingPair,
		},
	}
}

// Apply checks all boxes for pointing pairs and all rows/columns for box-line reductions.
func (s *PointingPairSolver) Apply(board *core.Board) *Move {
	// Check pointing pairs: candidate in a box confined to one row or column.
	if move := s.findPointingPairs(board); move != nil {
		return move
	}

	// Check box-line reductions: candidate in a row/column confined to one box.
	if move := s.findBoxLineReductions(board); move != nil {
		return move
	}

	return nil
}

// Hint returns a useful pointing-pair or box-line elimination even when it
// does not immediately create a placement. Placement-producing moves retain
// the existing solving and teaching sequence.
func (s *PointingPairSolver) Hint(board *core.Board) *Move {
	if move := s.Apply(board); move != nil {
		return move
	}
	if move := s.findPointingPairEliminations(board); move != nil {
		return move
	}
	return s.findBoxLineEliminations(board)
}

func (s *PointingPairSolver) findPointingPairEliminations(board *core.Board) *Move {
	for boxRow := 0; boxRow < 3; boxRow++ {
		for boxCol := 0; boxCol < 3; boxCol++ {
			startRow, startCol := boxRow*3, boxCol*3
			box := UnitRef{Kind: UnitBox, Index: boxRow*3 + boxCol}

			for digit := 1; digit <= 9; digit++ {
				rowSet := make(map[int]bool)
				colSet := make(map[int]bool)
				var positions []core.Position
				for row := startRow; row < startRow+3; row++ {
					for column := startCol; column < startCol+3; column++ {
						position := core.NewPosition(row, column)
						if board.Get(position) == 0 && board.Candidates(position).Has(digit) {
							rowSet[row] = true
							colSet[column] = true
							positions = append(positions, position)
						}
					}
				}
				if len(positions) < 2 {
					continue
				}

				if len(rowSet) == 1 {
					row := positions[0].Row
					var eliminations []CandidateRef
					for column := 0; column < 9; column++ {
						if column >= startCol && column < startCol+3 {
							continue
						}
						position := core.NewPosition(row, column)
						if board.Get(position) == 0 && board.Candidates(position).Has(digit) {
							eliminations = append(eliminations, CandidateRef{Position: position, Value: digit})
						}
					}
					if move := pointingEliminationMove(s.Key, box, UnitRef{Kind: UnitRow, Index: row}, positions, eliminations, digit); move != nil {
						return move
					}
				}

				if len(colSet) == 1 {
					column := positions[0].Column
					var eliminations []CandidateRef
					for row := 0; row < 9; row++ {
						if row >= startRow && row < startRow+3 {
							continue
						}
						position := core.NewPosition(row, column)
						if board.Get(position) == 0 && board.Candidates(position).Has(digit) {
							eliminations = append(eliminations, CandidateRef{Position: position, Value: digit})
						}
					}
					if move := pointingEliminationMove(s.Key, box, UnitRef{Kind: UnitColumn, Index: column}, positions, eliminations, digit); move != nil {
						return move
					}
				}
			}
		}
	}
	return nil
}

func (s *PointingPairSolver) findBoxLineEliminations(board *core.Board) *Move {
	for row := 0; row < 9; row++ {
		for digit := 1; digit <= 9; digit++ {
			boxes := make(map[int]bool)
			var positions []core.Position
			for column := 0; column < 9; column++ {
				position := core.NewPosition(row, column)
				if board.Get(position) == 0 && board.Candidates(position).Has(digit) {
					boxes[column/3] = true
					positions = append(positions, position)
				}
			}
			if len(positions) < 2 || len(boxes) != 1 {
				continue
			}
			boxColumn := positions[0].Column / 3
			boxRow := row / 3
			var eliminations []CandidateRef
			for otherRow := boxRow * 3; otherRow < boxRow*3+3; otherRow++ {
				if otherRow == row {
					continue
				}
				for column := boxColumn * 3; column < boxColumn*3+3; column++ {
					position := core.NewPosition(otherRow, column)
					if board.Get(position) == 0 && board.Candidates(position).Has(digit) {
						eliminations = append(eliminations, CandidateRef{Position: position, Value: digit})
					}
				}
			}
			if move := pointingEliminationMove(s.Key, UnitRef{Kind: UnitRow, Index: row}, UnitRef{Kind: UnitBox,
				Index: boxRow*3 + boxColumn}, positions, eliminations, digit); move != nil {
				return move
			}
		}
	}

	for column := 0; column < 9; column++ {
		for digit := 1; digit <= 9; digit++ {
			boxes := make(map[int]bool)
			var positions []core.Position
			for row := 0; row < 9; row++ {
				position := core.NewPosition(row, column)
				if board.Get(position) == 0 && board.Candidates(position).Has(digit) {
					boxes[row/3] = true
					positions = append(positions, position)
				}
			}
			if len(positions) < 2 || len(boxes) != 1 {
				continue
			}
			boxRow := positions[0].Row / 3
			boxColumn := column / 3
			var eliminations []CandidateRef
			for row := boxRow * 3; row < boxRow*3+3; row++ {
				for otherColumn := boxColumn * 3; otherColumn < boxColumn*3+3; otherColumn++ {
					if otherColumn == column {
						continue
					}
					position := core.NewPosition(row, otherColumn)
					if board.Get(position) == 0 && board.Candidates(position).Has(digit) {
						eliminations = append(eliminations, CandidateRef{Position: position, Value: digit})
					}
				}
			}
			if move := pointingEliminationMove(s.Key, UnitRef{Kind: UnitColumn, Index: column}, UnitRef{Kind: UnitBox,
				Index: boxRow*3 + boxColumn}, positions, eliminations, digit); move != nil {
				return move
			}
		}
	}
	return nil
}

func pointingEliminationMove(
	technique string,
	source, intersecting UnitRef,
	positions []core.Position,
	eliminations []CandidateRef,
	digit int,
) *Move {
	if len(eliminations) == 0 {
		return nil
	}
	evidence := pointingEvidence(source, intersecting, positions, digit)
	evidence.Eliminations = eliminations
	return &Move{
		Technique: technique,
		Reason: fmt.Sprintf(
			"candidate %d is confined to the intersection of %s %d and %s %d",
			digit,
			source.Kind,
			source.Index+1,
			intersecting.Kind,
			intersecting.Index+1,
		),
		EliminationOnly: true,
		Evidence:        evidence,
	}
}

// findPointingPairs checks each box: if a candidate appears only in one row
// or one column within the box, eliminate it from that row/column outside the box.
func (s *PointingPairSolver) findPointingPairs(board *core.Board) *Move {
	for boxRow := 0; boxRow < 3; boxRow++ {
		for boxCol := 0; boxCol < 3; boxCol++ {
			startRow, startCol := boxRow*3, boxCol*3

			for digit := 1; digit <= 9; digit++ {
				// Track which rows and columns in this box contain the digit as a candidate.
				rowSet := make(map[int]bool)
				colSet := make(map[int]bool)
				var positions []core.Position

				for r := startRow; r < startRow+3; r++ {
					for c := startCol; c < startCol+3; c++ {
						pos := core.NewPosition(r, c)
						if board.Get(pos) == 0 && board.Candidates(pos).Has(digit) {
							rowSet[r] = true
							colSet[c] = true
							positions = append(positions, pos)
						}
					}
				}

				if len(positions) < 2 {
					continue
				}

				// Pointing pair in a row: all positions are in the same row.
				if len(rowSet) == 1 {
					row := positions[0].Row
					// Eliminate digit from rest of the row outside this box.
					for c := 0; c < 9; c++ {
						if c >= startCol && c < startCol+3 {
							continue // Skip cells inside the box.
						}
						pos := core.NewPosition(row, c)
						if board.Get(pos) != 0 {
							continue
						}

						cands := board.Candidates(pos)
						if cands.Has(digit) {
							reduced := cands
							reduced.Remove(digit)
							if reduced.Count() == 1 {
								value := reduced.Values()[0]
								return &Move{
									Cell:      core.NewCell(pos, value),
									Technique: s.Key,
									Reason: fmt.Sprintf(
										"Pointing pair: %d in box %d is confined to row %d, leaving %d as the only candidate for %s",
										digit,
										boxRow*3+boxCol+1,
										row+1,
										value,
										pos.ToString(),
									),
									Evidence: pointingEvidence(
										UnitRef{Kind: UnitBox, Index: boxRow*3 + boxCol},
										UnitRef{Kind: UnitRow, Index: row},
										positions,
										digit,
									),
								}
							}
						}
					}
				}

				// Pointing pair in a column: all positions are in the same column.
				if len(colSet) == 1 {
					col := positions[0].Column
					// Eliminate digit from rest of the column outside this box.
					for r := 0; r < 9; r++ {
						if r >= startRow && r < startRow+3 {
							continue // Skip cells inside the box.
						}
						pos := core.NewPosition(r, col)
						if board.Get(pos) != 0 {
							continue
						}

						cands := board.Candidates(pos)
						if cands.Has(digit) {
							reduced := cands
							reduced.Remove(digit)
							if reduced.Count() == 1 {
								value := reduced.Values()[0]
								return &Move{
									Cell:      core.NewCell(pos, value),
									Technique: s.Key,
									Reason: fmt.Sprintf(
										"Pointing pair: %d in box %d is confined to column %d, leaving %d as the only candidate for %s",
										digit,
										boxRow*3+boxCol+1,
										col+1,
										value,
										pos.ToString(),
									),
									Evidence: pointingEvidence(
										UnitRef{Kind: UnitBox, Index: boxRow*3 + boxCol},
										UnitRef{Kind: UnitColumn, Index: col},
										positions,
										digit,
									),
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// findBoxLineReductions checks each row and column: if a candidate appears
// only in one box within a row/column, eliminate it from the rest of that box.
func (s *PointingPairSolver) findBoxLineReductions(board *core.Board) *Move {
	// Check rows.
	for row := 0; row < 9; row++ {
		for digit := 1; digit <= 9; digit++ {
			boxColSet := make(map[int]bool)
			var positions []core.Position

			for col := 0; col < 9; col++ {
				pos := core.NewPosition(row, col)
				if board.Get(pos) == 0 && board.Candidates(pos).Has(digit) {
					boxColSet[col/3] = true
					positions = append(positions, pos)
				}
			}

			if len(positions) < 2 || len(boxColSet) != 1 {
				continue
			}

			// All occurrences of digit in this row are in one box.
			boxCol := positions[0].Column / 3
			boxRow := row / 3
			startRow, startCol := boxRow*3, boxCol*3

			// Eliminate digit from the rest of the box (cells not in this row).
			for r := startRow; r < startRow+3; r++ {
				if r == row {
					continue
				}
				for c := startCol; c < startCol+3; c++ {
					pos := core.NewPosition(r, c)
					if board.Get(pos) != 0 {
						continue
					}

					cands := board.Candidates(pos)
					if cands.Has(digit) {
						reduced := cands
						reduced.Remove(digit)
						if reduced.Count() == 1 {
							value := reduced.Values()[0]
							return &Move{
								Cell:      core.NewCell(pos, value),
								Technique: s.Key,
								Reason: fmt.Sprintf(
									"Box-line reduction: %d in row %d is confined to box %d, leaving %d as the only candidate for %s",
									digit,
									row+1,
									boxRow*3+boxCol+1,
									value,
									pos.ToString(),
								),
								Evidence: pointingEvidence(
									UnitRef{Kind: UnitRow, Index: row},
									UnitRef{Kind: UnitBox, Index: boxRow*3 + boxCol},
									positions,
									digit,
								),
							}
						}
					}
				}
			}
		}
	}

	// Check columns.
	for col := 0; col < 9; col++ {
		for digit := 1; digit <= 9; digit++ {
			boxRowSet := make(map[int]bool)
			var positions []core.Position

			for row := 0; row < 9; row++ {
				pos := core.NewPosition(row, col)
				if board.Get(pos) == 0 && board.Candidates(pos).Has(digit) {
					boxRowSet[row/3] = true
					positions = append(positions, pos)
				}
			}

			if len(positions) < 2 || len(boxRowSet) != 1 {
				continue
			}

			// All occurrences of digit in this column are in one box.
			boxRow := positions[0].Row / 3
			boxCol := col / 3
			startRow, startCol := boxRow*3, boxCol*3

			// Eliminate digit from the rest of the box (cells not in this column).
			for r := startRow; r < startRow+3; r++ {
				for c := startCol; c < startCol+3; c++ {
					if c == col {
						continue
					}
					pos := core.NewPosition(r, c)
					if board.Get(pos) != 0 {
						continue
					}

					cands := board.Candidates(pos)
					if cands.Has(digit) {
						reduced := cands
						reduced.Remove(digit)
						if reduced.Count() == 1 {
							value := reduced.Values()[0]
							return &Move{
								Cell:      core.NewCell(pos, value),
								Technique: s.Key,
								Reason: fmt.Sprintf(
									"Box-line reduction: %d in column %d is confined to box %d, leaving %d as the only candidate for %s",
									digit,
									col+1,
									boxRow*3+boxCol+1,
									value,
									pos.ToString(),
								),
								Evidence: pointingEvidence(
									UnitRef{Kind: UnitColumn, Index: col},
									UnitRef{Kind: UnitBox, Index: boxRow*3 + boxCol},
									positions,
									digit,
								),
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func pointingEvidence(source, intersecting UnitRef, positions []core.Position, digit int) *Evidence {
	premises := make([]CandidateGroup, len(positions))
	for i, position := range positions {
		premises[i] = CandidateGroup{Position: position, Values: []int{digit}}
	}
	return &Evidence{Unit: &source, Units: []UnitRef{source, intersecting}, Premises: premises}
}

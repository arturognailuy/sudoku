package solver

import (
	"fmt"

	"github.com/gnailuy/sudoku/core"
)

// XWingSolver finds X-Wing patterns.
//
// An X-Wing occurs when a candidate appears in exactly two cells in each of
// two rows, and those four cells share the same two columns (or the transpose
// with columns and rows swapped). The candidate can be eliminated from all
// other cells in those two columns (or rows), because one of the two
// row-pairs must contain the candidate — leaving no room for it elsewhere
// in those columns.
//
// Like the intermediate solvers, this solver returns a move only when an
// elimination creates a naked single.
type XWingSolver struct {
	Base
}

// NewXWingSolver creates an XWingSolver and returns it.
func NewXWingSolver() *XWingSolver {
	return &XWingSolver{
		Base: Base{
			Key:         "x-wing",
			DisplayName: "X-Wing",
			Description: "Finds a candidate confined to the same two columns in two rows (or same two rows in two columns), enabling eliminations that reveal a single candidate.",
		},
	}
}

// Apply checks for X-Wing patterns on all digits.
func (s *XWingSolver) Apply(board *core.Board) *Move {
	// Check row-based X-Wings: digit in exactly 2 cells in each of 2 rows,
	// sharing the same 2 columns → eliminate from those columns.
	if move := s.findRowXWing(board); move != nil {
		return move
	}

	// Check column-based X-Wings: digit in exactly 2 cells in each of 2
	// columns, sharing the same 2 rows → eliminate from those rows.
	if move := s.findColumnXWing(board); move != nil {
		return move
	}

	return nil
}

// findRowXWing searches for row-based X-Wing patterns.
func (s *XWingSolver) findRowXWing(board *core.Board) *Move {
	for digit := 1; digit <= 9; digit++ {
		// Collect rows where the digit appears as a candidate in exactly 2 cells.
		type rowPair struct {
			row  int
			cols [2]int
		}
		var pairs []rowPair

		for row := 0; row < 9; row++ {
			var cols []int
			for col := 0; col < 9; col++ {
				pos := core.NewPosition(row, col)
				if board.Get(pos) == 0 && board.Candidates(pos).Has(digit) {
					cols = append(cols, col)
				}
			}
			if len(cols) == 2 {
				pairs = append(pairs, rowPair{row: row, cols: [2]int{cols[0], cols[1]}})
			}
		}

		// Look for two rows sharing the same two columns.
		for i := 0; i < len(pairs); i++ {
			for j := i + 1; j < len(pairs); j++ {
				if pairs[i].cols != pairs[j].cols {
					continue
				}

				// Found an X-Wing. Eliminate digit from the two columns
				// (rows other than the X-Wing rows).
				col0, col1 := pairs[i].cols[0], pairs[i].cols[1]
				row0, row1 := pairs[i].row, pairs[j].row

				for _, col := range []int{col0, col1} {
					for row := 0; row < 9; row++ {
						if row == row0 || row == row1 {
							continue
						}
						pos := core.NewPosition(row, col)
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
										"X-Wing: %d in rows %d,%d is confined to columns %d,%d, leaving %d as the only candidate for %s",
										digit, row0+1, row1+1, col0+1, col1+1, value, pos.ToString(),
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

// findColumnXWing searches for column-based X-Wing patterns.
func (s *XWingSolver) findColumnXWing(board *core.Board) *Move {
	for digit := 1; digit <= 9; digit++ {
		// Collect columns where the digit appears as a candidate in exactly 2 cells.
		type colPair struct {
			col  int
			rows [2]int
		}
		var pairs []colPair

		for col := 0; col < 9; col++ {
			var rows []int
			for row := 0; row < 9; row++ {
				pos := core.NewPosition(row, col)
				if board.Get(pos) == 0 && board.Candidates(pos).Has(digit) {
					rows = append(rows, row)
				}
			}
			if len(rows) == 2 {
				pairs = append(pairs, colPair{col: col, rows: [2]int{rows[0], rows[1]}})
			}
		}

		// Look for two columns sharing the same two rows.
		for i := 0; i < len(pairs); i++ {
			for j := i + 1; j < len(pairs); j++ {
				if pairs[i].rows != pairs[j].rows {
					continue
				}

				// Found an X-Wing. Eliminate digit from the two rows
				// (columns other than the X-Wing columns).
				row0, row1 := pairs[i].rows[0], pairs[i].rows[1]
				col0, col1 := pairs[i].col, pairs[j].col

				for _, row := range []int{row0, row1} {
					for col := 0; col < 9; col++ {
						if col == col0 || col == col1 {
							continue
						}
						pos := core.NewPosition(row, col)
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
										"X-Wing: %d in columns %d,%d is confined to rows %d,%d, leaving %d as the only candidate for %s",
										digit, col0+1, col1+1, row0+1, row1+1, value, pos.ToString(),
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

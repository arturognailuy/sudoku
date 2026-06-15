package solver

import (
	"fmt"

	"github.com/gnailuy/sudoku/core"
)

// SwordfishSolver finds Swordfish patterns.
//
// A Swordfish is the 3×3 extension of X-Wing. A candidate appears in 2–3
// cells in each of three rows, and all those cells are confined to the same
// three columns (or the transpose with columns and rows swapped). The
// candidate can be eliminated from all other cells in those three columns
// (or rows), because one combination of the three row-sets must contain
// the candidate — leaving no room for it elsewhere in those columns.
//
// Like the other strategy solvers, this solver returns a move only when an
// elimination creates a naked single.
type SwordfishSolver struct {
	Base
}

// NewSwordfishSolver creates a SwordfishSolver and returns it.
func NewSwordfishSolver() *SwordfishSolver {
	return &SwordfishSolver{
		Base: Base{
			Key:         "swordfish",
			DisplayName: "Swordfish",
			Description: "Finds a candidate confined to the same three columns in three rows (or same three rows in three columns), enabling eliminations that reveal a single candidate.",
		},
	}
}

// Apply checks for Swordfish patterns on all digits.
func (s *SwordfishSolver) Apply(board *core.Board) *Move {
	// Check row-based Swordfish: digit in 2–3 cells in each of 3 rows,
	// confined to the same 3 columns → eliminate from those columns.
	if move := s.findRowSwordfish(board); move != nil {
		return move
	}

	// Check column-based Swordfish: digit in 2–3 cells in each of 3
	// columns, confined to the same 3 rows → eliminate from those rows.
	if move := s.findColumnSwordfish(board); move != nil {
		return move
	}

	return nil
}

// rowInfo holds a row index and the set of columns where a digit appears.
type rowInfo struct {
	row  int
	cols []int
}

// colInfo holds a column index and the set of rows where a digit appears.
type colInfo struct {
	col  int
	rows []int
}

// findRowSwordfish searches for row-based Swordfish patterns.
func (s *SwordfishSolver) findRowSwordfish(board *core.Board) *Move {
	for digit := 1; digit <= 9; digit++ {
		// Collect rows where the digit appears as a candidate in 2–3 cells.
		var eligible []rowInfo

		for row := 0; row < 9; row++ {
			var cols []int
			for col := 0; col < 9; col++ {
				pos := core.NewPosition(row, col)
				if board.Get(pos) == 0 && board.Candidates(pos).Has(digit) {
					cols = append(cols, col)
				}
			}
			if len(cols) >= 2 && len(cols) <= 3 {
				eligible = append(eligible, rowInfo{row: row, cols: cols})
			}
		}

		if len(eligible) < 3 {
			continue
		}

		// Try all combinations of 3 eligible rows.
		for i := 0; i < len(eligible); i++ {
			for j := i + 1; j < len(eligible); j++ {
				for k := j + 1; k < len(eligible); k++ {
					// Compute the union of columns.
					colSet := make(map[int]bool)
					for _, c := range eligible[i].cols {
						colSet[c] = true
					}
					for _, c := range eligible[j].cols {
						colSet[c] = true
					}
					for _, c := range eligible[k].cols {
						colSet[c] = true
					}

					// Swordfish: the union must be exactly 3 columns.
					if len(colSet) != 3 {
						continue
					}

					// Collect the 3 columns and 3 rows.
					cols := make([]int, 0, 3)
					for c := range colSet {
						cols = append(cols, c)
					}
					rows := []int{eligible[i].row, eligible[j].row, eligible[k].row}

					rowSet := make(map[int]bool, 3)
					for _, r := range rows {
						rowSet[r] = true
					}

					// Eliminate digit from other cells in those 3 columns.
					for _, col := range cols {
						for row := 0; row < 9; row++ {
							if rowSet[row] {
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
											"Swordfish: %d in rows %d,%d,%d is confined to columns %d,%d,%d, leaving %d as the only candidate for %s",
											digit, rows[0]+1, rows[1]+1, rows[2]+1, cols[0]+1, cols[1]+1, cols[2]+1, value, pos.ToString(),
										),
									}
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

// findColumnSwordfish searches for column-based Swordfish patterns.
func (s *SwordfishSolver) findColumnSwordfish(board *core.Board) *Move {
	for digit := 1; digit <= 9; digit++ {
		// Collect columns where the digit appears as a candidate in 2–3 cells.
		var eligible []colInfo

		for col := 0; col < 9; col++ {
			var rows []int
			for row := 0; row < 9; row++ {
				pos := core.NewPosition(row, col)
				if board.Get(pos) == 0 && board.Candidates(pos).Has(digit) {
					rows = append(rows, row)
				}
			}
			if len(rows) >= 2 && len(rows) <= 3 {
				eligible = append(eligible, colInfo{col: col, rows: rows})
			}
		}

		if len(eligible) < 3 {
			continue
		}

		// Try all combinations of 3 eligible columns.
		for i := 0; i < len(eligible); i++ {
			for j := i + 1; j < len(eligible); j++ {
				for k := j + 1; k < len(eligible); k++ {
					// Compute the union of rows.
					rowSet := make(map[int]bool)
					for _, r := range eligible[i].rows {
						rowSet[r] = true
					}
					for _, r := range eligible[j].rows {
						rowSet[r] = true
					}
					for _, r := range eligible[k].rows {
						rowSet[r] = true
					}

					// Swordfish: the union must be exactly 3 rows.
					if len(rowSet) != 3 {
						continue
					}

					// Collect the 3 rows and 3 columns.
					rows := make([]int, 0, 3)
					for r := range rowSet {
						rows = append(rows, r)
					}
					cols := []int{eligible[i].col, eligible[j].col, eligible[k].col}

					colSet := make(map[int]bool, 3)
					for _, c := range cols {
						colSet[c] = true
					}

					// Eliminate digit from other cells in those 3 rows.
					for _, row := range rows {
						for col := 0; col < 9; col++ {
							if colSet[col] {
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
											"Swordfish: %d in columns %d,%d,%d is confined to rows %d,%d,%d, leaving %d as the only candidate for %s",
											digit, cols[0]+1, cols[1]+1, cols[2]+1, rows[0]+1, rows[1]+1, rows[2]+1, value, pos.ToString(),
										),
									}
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

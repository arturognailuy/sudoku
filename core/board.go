package core

import (
	"errors"
	"fmt"

	"github.com/gnailuy/sudoku/util"
)

// Board represents a 9x9 Sudoku grid with automatic candidate tracking.
type Board struct {
	grid             [9][9]int
	candidates       [9][9]CandidateSet
	filledCellsCount int
}

// NewEmptyBoard creates an empty Sudoku board with all cells set to zero
// and all candidates (1–9) available in every cell.
func NewEmptyBoard() Board {
	var b Board
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			b.candidates[i][j] = allCandidates
		}
	}
	return b
}

// eliminatePeers removes value from the candidate sets of all peers
// (same row, column, and box) of the given position.
func (board *Board) eliminatePeers(position Position, value int) {
	// Row peers.
	for c := 0; c < 9; c++ {
		board.candidates[position.Row][c].Remove(value)
	}
	// Column peers.
	for r := 0; r < 9; r++ {
		board.candidates[r][position.Column].Remove(value)
	}
	// Box peers.
	startRow, startCol := position.Row/3*3, position.Column/3*3
	for r := startRow; r < startRow+3; r++ {
		for c := startCol; c < startCol+3; c++ {
			board.candidates[r][c].Remove(value)
		}
	}
}

// restorePeers recalculates candidates for all peers of the given position
// after a value is unset. It adds the value back as a candidate for any peer
// where the value does not conflict with existing filled cells.
func (board *Board) restorePeers(position Position, value int) {
	// Row peers.
	for c := 0; c < 9; c++ {
		if board.grid[position.Row][c] == 0 && board.isValueValidForCell(NewPosition(position.Row, c), value) {
			board.candidates[position.Row][c].Add(value)
		}
	}
	// Column peers.
	for r := 0; r < 9; r++ {
		if board.grid[r][position.Column] == 0 && board.isValueValidForCell(NewPosition(r, position.Column), value) {
			board.candidates[r][position.Column].Add(value)
		}
	}
	// Box peers.
	startRow, startCol := position.Row/3*3, position.Column/3*3
	for r := startRow; r < startRow+3; r++ {
		for c := startCol; c < startCol+3; c++ {
			if board.grid[r][c] == 0 && board.isValueValidForCell(NewPosition(r, c), value) {
				board.candidates[r][c].Add(value)
			}
		}
	}
}

// isValueValidForCell checks whether placing value at position would conflict
// with any existing filled cell in the same row, column, or box.
func (board *Board) isValueValidForCell(position Position, value int) bool {
	// Check row.
	for c := 0; c < 9; c++ {
		if c != position.Column && board.grid[position.Row][c] == value {
			return false
		}
	}
	// Check column.
	for r := 0; r < 9; r++ {
		if r != position.Row && board.grid[r][position.Column] == value {
			return false
		}
	}
	// Check box.
	startRow, startCol := position.Row/3*3, position.Column/3*3
	for r := startRow; r < startRow+3; r++ {
		for c := startCol; c < startCol+3; c++ {
			if (r != position.Row || c != position.Column) && board.grid[r][c] == value {
				return false
			}
		}
	}
	return true
}

// Function to set the value to a position.
func (board *Board) Set(position Position, value int) (err error) {
	if value < 1 || value > 9 {
		return errors.New("cannot set invalid number: " + fmt.Sprint(value))
	}

	oldValue := board.grid[position.Row][position.Column]
	if oldValue == 0 {
		board.filledCellsCount++
	} else if oldValue != value {
		// Restore old value's candidates in peers before overwriting.
		board.restorePeers(position, oldValue)
	}

	board.grid[position.Row][position.Column] = value
	// A filled cell has no candidates.
	board.candidates[position.Row][position.Column] = 0
	// Eliminate this value from all peers.
	if oldValue != value {
		board.eliminatePeers(position, value)
	}

	return nil
}

// Function to set the value of a cell.
func (board *Board) SetCell(cell Cell) (err error) {
	if !cell.IsValid() {
		return errors.New("cannot set invalid cell: " + cell.ToString())
	}

	_ = board.Set(cell.Position, cell.Value)

	return nil
}

// Function to unset the value of a position.
func (board *Board) Unset(position Position) {
	value := board.grid[position.Row][position.Column]
	if value > 0 {
		board.filledCellsCount--
		board.grid[position.Row][position.Column] = 0

		// Restore candidates for this cell.
		board.candidates[position.Row][position.Column] = board.computeCandidates(position)
		// Restore the removed value as candidate in peers.
		board.restorePeers(position, value)
	}
}

// computeCandidates returns the full candidate set for an empty cell at position,
// based on the current grid state.
func (board *Board) computeCandidates(position Position) CandidateSet {
	cs := allCandidates
	// Remove values in same row.
	for c := 0; c < 9; c++ {
		if v := board.grid[position.Row][c]; v != 0 {
			cs.Remove(v)
		}
	}
	// Remove values in same column.
	for r := 0; r < 9; r++ {
		if v := board.grid[r][position.Column]; v != 0 {
			cs.Remove(v)
		}
	}
	// Remove values in same box.
	startRow, startCol := position.Row/3*3, position.Column/3*3
	for r := startRow; r < startRow+3; r++ {
		for c := startCol; c < startCol+3; c++ {
			if v := board.grid[r][c]; v != 0 {
				cs.Remove(v)
			}
		}
	}
	return cs
}

// Function to get the value of a position.
func (board *Board) Get(position Position) int {
	return board.grid[position.Row][position.Column]
}

// Candidates returns the candidate set for the cell at position.
// For filled cells, the set is empty.
func (board *Board) Candidates(position Position) CandidateSet {
	return board.candidates[position.Row][position.Column]
}

// EmptyPositions returns all positions on the board that are empty (value 0).
func (board *Board) EmptyPositions() []Position {
	positions := make([]Position, 0, 81-board.filledCellsCount)
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board.grid[r][c] == 0 {
				positions = append(positions, NewPosition(r, c))
			}
		}
	}
	return positions
}

// ForEachCell calls fn for every cell on the board with its position and value.
// If fn returns false, iteration stops early.
func (board *Board) ForEachCell(fn func(position Position, value int) bool) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if !fn(NewPosition(r, c), board.grid[r][c]) {
				return
			}
		}
	}
}

// Function to get a random position satisfying the value validator.
func (board *Board) GetRandomPositionWith(validator func(int) bool) *Position {
	rowOrder := util.GenerateNumberArray(0, 9, true)
	columnOrder := util.GenerateNumberArray(0, 9, true)
	for _, row := range rowOrder {
		for _, column := range columnOrder {
			position := NewPosition(row, column)
			value := board.Get(position)
			if validator(value) {
				return &position
			}
		}
	}

	return nil
}

// Function to get the number of filled cells.
func (board *Board) GetFilledCellsCount() int {
	return board.filledCellsCount
}

// Function to return a copy of the board.
func (board *Board) Copy() Board {
	return Board{
		grid:             board.grid,
		candidates:       board.candidates,
		filledCellsCount: board.filledCellsCount,
	}
}

// Function to merge the board with another board.
func (board *Board) Merge(otherBoard Board) {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if board.grid[i][j] == 0 && otherBoard.grid[i][j] != 0 {
				_ = board.Set(NewPosition(i, j), otherBoard.grid[i][j])
			}
		}
	}
}

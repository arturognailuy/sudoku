package core

import (
	"errors"
	"fmt"

	"github.com/gnailuy/sudoku/util"
)

// Board represents a 9x9 Sudoku grid.
type Board struct {
	grid             [9][9]int
	filledCellsCount int
}

// NewEmptyBoard creates an empty Sudoku board with all cells set to zero.
func NewEmptyBoard() Board {
	return *new(Board)
}

// Function to set the value to a position.
func (board *Board) Set(position Position, value int) (err error) {
	if value < 1 || value > 9 {
		return errors.New("cannot set invalid number: " + fmt.Sprint(value))
	}

	if board.grid[position.Row][position.Column] == 0 {
		board.filledCellsCount++
	}
	board.grid[position.Row][position.Column] = value

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
	if board.grid[position.Row][position.Column] > 0 {
		board.filledCellsCount--
	}
	board.grid[position.Row][position.Column] = 0
}

// Function to get the value of a position.
func (board *Board) Get(position Position) int {
	return board.grid[position.Row][position.Column]
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
		filledCellsCount: board.filledCellsCount,
	}
}

// Function to merge the board with another board.
func (board *Board) Merge(otherBoard Board) {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if board.grid[i][j] == 0 && otherBoard.grid[i][j] != 0 {
				board.grid[i][j] = otherBoard.grid[i][j]
				board.filledCellsCount++
			}
		}
	}
}

package solver

import "github.com/gnailuy/sudoku/core"

// Solver defines the interface for all Sudoku solvers.
type Solver interface {
	// Return the key of the solver.
	GetKey() string

	// Return the display name of the solver.
	GetDisplayName() string

	// Return the description of the solver.
	GetDescription() string

	// Return if the solver is reliable.
	IsReliable() bool

	// Solve the Sudoku board, return false if the solver cannot fully solve the board.
	Solve(board *core.Board) bool

	// Give a hint for the next step of the board, return nil if the solver cannot give a hint.
	Hint(board *core.Board) *core.Cell

	// Count the number of solutions of the board. Return 0 if the solver cannot solve the board; return 1 if the board is already solved.
	CountSolutions(board *core.Board) int
}

// Base provides common fields for all solver implementations.
type Base struct {
	Key         string // The unique key of the solver.
	DisplayName string // The display name of the solver.
	Description string // The description of the solver.
	Reliable    bool   // If the solver is reliable, it will always solve a valid Sudoku board, otherwise, it may not be able to solve some boards.
}

// Function to get the key of the base solver.
func (solver Base) GetKey() string {
	return solver.Key
}

// Function to get the display name of the base solver.
func (solver Base) GetDisplayName() string {
	return solver.DisplayName
}

// Function to get the description of the base solver.
func (solver Base) GetDescription() string {
	return solver.Description
}

// Function to check if the base solver is reliable.
func (solver Base) IsReliable() bool {
	return solver.Reliable
}

// CountSolutions implements the default solution counting logic.
func (solver Base) CountSolutions(board *core.Board) int {
	// Reliable solvers should always override this function.
	if solver.Reliable {
		panic("Bug: Reliable solver should override the CountSolutions function")
	}

	// Unreliable solvers should return 0 as they may not be able to fully solve the board.
	return 0
}

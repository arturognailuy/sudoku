package generator

import "github.com/gnailuy/sudoku/solver"

// Options defines the parameters for puzzle generation.
type Options struct {
	// Public fields.
	MaximumSolutions  int
	MaximumIterations int
	Difficulty        Difficulty

	// Private fields.
	solverStore solver.Store
}

// NewProblemOptions creates default generator options.
func NewProblemOptions(solverStore solver.Store, difficulty Difficulty) Options {
	return Options{
		MaximumSolutions:  1,
		MaximumIterations: 1024,
		Difficulty:        difficulty,
		solverStore:       solverStore,
	}
}

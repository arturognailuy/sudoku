package game

import (
	"github.com/gnailuy/sudoku/solver"
)

// Options configures the game with a solver store and optional strategy solver keys.
type Options struct {
	// Public fields.
	StrategySolverKeys []string

	// Private fields.
	solverStore solver.Store
}

// NewDefaultOptions creates default game options with no strategy solvers.
func NewDefaultOptions(solverStore solver.Store) Options {
	return Options{
		StrategySolverKeys: []string{},
		solverStore:        solverStore,
	}
}

// Function to get the strategy solvers from the store.
func (options *Options) GetStrategySolvers() []solver.Solver {
	strategySolvers := []solver.Solver{}

	for _, key := range options.StrategySolverKeys {
		solver := options.solverStore.GetSolverByKey(key)
		if solver != nil {
			strategySolvers = append(strategySolvers, solver)
		} else {
			panic("Bug: Invalid solver key: " + key)
		}
	}

	return strategySolvers
}

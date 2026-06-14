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

// GetStrategySolvers returns the strategy solvers specified by the options.
func (options *Options) GetStrategySolvers() []solver.StrategySolver {
	strategySolvers := []solver.StrategySolver{}

	for _, key := range options.StrategySolverKeys {
		s := options.solverStore.GetStrategySolverByKey(key)
		if s != nil {
			strategySolvers = append(strategySolvers, s)
		} else {
			panic("Bug: Invalid solver key: " + key)
		}
	}

	return strategySolvers
}

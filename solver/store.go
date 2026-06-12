package solver

// Store maps solver keys to Solver implementations.
type Store map[string]Solver

// NewStore creates a Store and registers the default backtracking solver.
func NewStore() Store {
	store := make(Store)

	// Register the default solver.
	backtracker := NewBacktracker()
	store[backtracker.GetKey()] = backtracker

	return store
}

// GetSolverByKey returns the solver for the given key, or nil if not found.
func (store Store) GetSolverByKey(key string) Solver {
	if solver, ok := store[key]; ok {
		return solver
	}

	return nil
}

// GetDefaultSolver returns the default reliable solver, panicking if not found.
func (store Store) GetDefaultSolver() Solver {
	defaultSolver := store.GetSolverByKey("default")

	if defaultSolver == nil {
		panic("Bug: Default solver not found in the store")
	}

	if !defaultSolver.IsReliable() {
		panic("Bug: Default solver must be reliable")
	}

	return defaultSolver
}

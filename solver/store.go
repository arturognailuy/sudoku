package solver

// Store maps solver keys to Solver implementations and provides typed access
// for CompleteSolver and StrategySolver lookups.
type Store struct {
	complete  map[string]CompleteSolver
	strategy  map[string]StrategySolver
}

// NewStore creates a Store and registers the default backtracking solver
// and all built-in strategy solvers.
func NewStore() Store {
	store := Store{
		complete: make(map[string]CompleteSolver),
		strategy: make(map[string]StrategySolver),
	}

	// Register the default solver.
	backtracker := NewBacktracker()
	store.complete[backtracker.GetKey()] = backtracker

	// Register strategy solvers.
	store.RegisterStrategy(NewNakedSingleSolver())
	store.RegisterStrategy(NewHiddenSingleSolver())
	store.RegisterStrategy(NewNakedSubsetSolver())
	store.RegisterStrategy(NewPointingPairSolver())

	return store
}

// RegisterStrategy adds a StrategySolver to the store.
func (store Store) RegisterStrategy(s StrategySolver) {
	store.strategy[s.GetKey()] = s
}

// GetDefaultSolver returns the default reliable solver, panicking if not found.
func (store Store) GetDefaultSolver() CompleteSolver {
	defaultSolver, ok := store.complete["default"]
	if !ok {
		panic("Bug: Default solver not found in the store")
	}

	return defaultSolver
}

// GetStrategySolverByKey returns the strategy solver for the given key, or nil if not found.
func (store Store) GetStrategySolverByKey(key string) StrategySolver {
	if s, ok := store.strategy[key]; ok {
		return s
	}

	return nil
}

// GetAllStrategySolverKeys returns the keys of all registered strategy solvers.
func (store Store) GetAllStrategySolverKeys() []string {
	keys := make([]string, 0, len(store.strategy))
	for key := range store.strategy {
		keys = append(keys, key)
	}

	return keys
}

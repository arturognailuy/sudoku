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

// BestEffortOptions extends Options with time and round limits for
// best-effort puzzle generation.
type BestEffortOptions struct {
	Options

	// MaxRounds is the maximum number of full generate-from-scratch
	// attempts before giving up. Zero means unlimited (original behavior).
	MaxRounds int

	// MaxDurationMs is the wall-clock time limit in milliseconds.
	// Zero means unlimited (original behavior).
	MaxDurationMs int64
}

// NewBestEffortOptions creates best-effort generator options with
// sensible defaults.
func NewBestEffortOptions(solverStore solver.Store, difficulty Difficulty) BestEffortOptions {
	return BestEffortOptions{
		Options:       NewProblemOptions(solverStore, difficulty),
		MaxRounds:     10,
		MaxDurationMs: 5000, // 5 seconds default.
	}
}

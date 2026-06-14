package generator

// Difficulty defines the difficulty level of a Sudoku problem.
type Difficulty struct {
	MinimumClues       int      // Inclusive.
	MaximumClues       int      // Exclusive.
	StrategySolverKeys []string // Allowed strategies to solve the problem in this difficulty level. Empty means all strategies are allowed.
	RequiredSolverKeys []string // At least one solver from this list must be needed (basic-only must fail to solve). Empty means no requirement.
}

// NewEasyDifficulty creates the easy difficulty level.
// Easy puzzles are solvable using only naked singles and hidden singles.
func NewEasyDifficulty() Difficulty {
	return Difficulty{
		MinimumClues:       45,
		MaximumClues:       60,
		StrategySolverKeys: []string{"naked-single", "hidden-single"},
	}
}

// NewMediumDifficulty creates the medium difficulty level.
// Medium puzzles are solvable using basic techniques plus naked pairs/triples
// and pointing pairs / box-line reductions.
// RequiredSolverKeys ensures the puzzle actually requires at least one
// intermediate technique — basic techniques alone won't suffice.
func NewMediumDifficulty() Difficulty {
	return Difficulty{
		MinimumClues:       32,
		MaximumClues:       45,
		StrategySolverKeys: []string{"naked-single", "hidden-single", "naked-subset", "pointing-pair"},
		RequiredSolverKeys: []string{"naked-subset", "pointing-pair"},
	}
}

// NewHardDifficulty creates the hard difficulty level.
func NewHardDifficulty() Difficulty {
	return Difficulty{
		MinimumClues:       25,
		MaximumClues:       32,
		StrategySolverKeys: []string{},
	}
}

// NewExtremeDifficulty creates the extreme difficulty level.
func NewExtremeDifficulty() Difficulty {
	return Difficulty{
		MinimumClues:       20,
		MaximumClues:       25,
		StrategySolverKeys: []string{},
	}
}

// NewEvilDifficulty creates the evil difficulty level.
func NewEvilDifficulty() Difficulty {
	return Difficulty{
		MinimumClues:       17,
		MaximumClues:       20,
		StrategySolverKeys: []string{},
	}
}

// NewCustomDifficulty creates a custom difficulty level.
func NewCustomDifficulty(minimumClues int, maximumClues int, solverKeys []string) Difficulty {
	return Difficulty{
		MinimumClues:       minimumClues,
		MaximumClues:       maximumClues,
		StrategySolverKeys: solverKeys,
		RequiredSolverKeys: []string{},
	}
}

// Function to check if the number of clues is within the difficulty level.
func (difficulty Difficulty) IsWithinDifficultyLevel(numberOfClues int) bool {
	return numberOfClues >= difficulty.MinimumClues && numberOfClues < difficulty.MaximumClues
}

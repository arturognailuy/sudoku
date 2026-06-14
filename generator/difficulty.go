package generator

// Difficulty defines the difficulty level of a Sudoku problem.
//
// SolverKeys lists the strategy solvers introduced at this tier.
// The full set of allowed solvers for a level equals SolverKeys plus
// all solvers from lower tiers. For example:
//   - Easy:   SolverKeys = [naked-single, hidden-single]
//   - Medium: SolverKeys = [naked-subset, pointing-pair]
//     Allowed = Easy.SolverKeys + Medium.SolverKeys
//
// During generation, the generator uses AllowedSolverKeys (the
// cumulative set) to constrain cell removal, and verifies that
// lower-tier solvers alone cannot solve the puzzle — ensuring the
// puzzle genuinely requires at least one solver from this tier.
//
// LowerTierSolverKeys holds the cumulative allowed solvers from all
// lower tiers. It is set automatically by the New*Difficulty()
// constructors. Empty means this is the lowest tier (or no solver
// constraint).
type Difficulty struct {
	MinimumClues        int      // Inclusive.
	MaximumClues        int      // Exclusive.
	SolverKeys          []string // Solvers introduced at this tier. Empty means no technique constraint.
	LowerTierSolverKeys []string // Cumulative solvers from all lower tiers. Empty for the lowest tier.
}

// AllowedSolverKeys returns the full set of solvers the puzzle may use:
// this tier's SolverKeys plus all lower-tier solvers.
func (d Difficulty) AllowedSolverKeys() []string {
	all := make([]string, 0, len(d.LowerTierSolverKeys)+len(d.SolverKeys))
	all = append(all, d.LowerTierSolverKeys...)
	all = append(all, d.SolverKeys...)
	return all
}

// easySolverKeys is the solver set for the Easy tier.
var easySolverKeys = []string{"naked-single", "hidden-single"}

// NewEasyDifficulty creates the easy difficulty level.
// Easy puzzles are solvable using only naked singles and hidden singles.
func NewEasyDifficulty() Difficulty {
	return Difficulty{
		MinimumClues: 45,
		MaximumClues: 60,
		SolverKeys:   easySolverKeys,
	}
}

// NewMediumDifficulty creates the medium difficulty level.
// Medium puzzles require at least one intermediate technique (naked-subset
// or pointing-pair) — basic techniques alone won't suffice.
func NewMediumDifficulty() Difficulty {
	return Difficulty{
		MinimumClues:        32,
		MaximumClues:        45,
		SolverKeys:          []string{"naked-subset", "pointing-pair"},
		LowerTierSolverKeys: easySolverKeys,
	}
}

// NewHardDifficulty creates the hard difficulty level.
func NewHardDifficulty() Difficulty {
	return Difficulty{
		MinimumClues: 25,
		MaximumClues: 32,
		SolverKeys:   []string{},
	}
}

// NewExtremeDifficulty creates the extreme difficulty level.
func NewExtremeDifficulty() Difficulty {
	return Difficulty{
		MinimumClues: 20,
		MaximumClues: 25,
		SolverKeys:   []string{},
	}
}

// NewEvilDifficulty creates the evil difficulty level.
func NewEvilDifficulty() Difficulty {
	return Difficulty{
		MinimumClues: 17,
		MaximumClues: 20,
		SolverKeys:   []string{},
	}
}

// NewCustomDifficulty creates a custom difficulty level.
func NewCustomDifficulty(minimumClues int, maximumClues int, solverKeys []string) Difficulty {
	return Difficulty{
		MinimumClues: minimumClues,
		MaximumClues: maximumClues,
		SolverKeys:   solverKeys,
	}
}

// Function to check if the number of clues is within the difficulty level.
func (difficulty Difficulty) IsWithinDifficultyLevel(numberOfClues int) bool {
	return numberOfClues >= difficulty.MinimumClues && numberOfClues < difficulty.MaximumClues
}

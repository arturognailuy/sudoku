package generator

// Difficulty defines the difficulty level of a Sudoku problem.
//
// SolverKeys lists the strategy solvers introduced at this tier.
// The full set of allowed solvers for a level equals SolverKeys plus
// all solvers from lower tiers, derived from tierRegistry/tierOrder.
// For example:
//   - Easy:   SolverKeys = [naked-single, hidden-single]
//   - Medium: SolverKeys = [naked-subset, pointing-pair]
//     Allowed = Easy.SolverKeys + Medium.SolverKeys
//
// During generation, the generator uses AllowedSolverKeys (the
// cumulative set) to constrain cell removal, and verifies that
// lower-tier solvers alone cannot solve the puzzle — ensuring the
// puzzle genuinely requires at least one solver from this tier.
//
// Lower-tier solver keys are derived from tierRegistry/tierOrder —
// there is no separate field. The registry is the single source of
// truth for tier ordering.
type Difficulty struct {
	MinimumClues int      // Inclusive.
	MaximumClues int      // Exclusive.
	SolverKeys   []string // Solvers introduced at this tier. Empty means no technique constraint.
}

// tierOrder defines the sequence of difficulty tiers from lowest to
// highest. Used alongside tierRegistry to derive lower-tier solver keys.
var tierOrder = []string{"easy", "medium", "hard"}

// tierRegistry maps each difficulty level name to the strategy solvers
// introduced at that tier. This is the single source of truth for the
// tier hierarchy — lower-tier solver keys are derived from tierOrder
// and this map.
var tierRegistry = map[string][]string{
	"easy":   {"naked-single", "hidden-single"},
	"medium": {"naked-subset", "pointing-pair"},
	"hard":   {"x-wing"},
}

// lowerTierSolverKeys returns the cumulative solver keys from all tiers
// below the tier that owns the given SolverKeys. Returns nil if the
// tier is the lowest or is not found in the registry.
func lowerTierSolverKeys(solverKeys []string) []string {
	if len(solverKeys) == 0 {
		return nil
	}

	for i, name := range tierOrder {
		tier := tierRegistry[name]
		if matchesTier(tier, solverKeys) {
			if i == 0 {
				return nil // lowest tier
			}
			var lower []string
			for _, n := range tierOrder[:i] {
				lower = append(lower, tierRegistry[n]...)
			}
			return lower
		}
	}

	return nil // not in registry (custom/unconstrained)
}

// matchesTier checks whether solverKeys matches the given tier entry.
func matchesTier(tier, solverKeys []string) bool {
	if len(tier) != len(solverKeys) {
		return false
	}
	for i := range tier {
		if tier[i] != solverKeys[i] {
			return false
		}
	}
	return true
}

// AllowedSolverKeys returns the full set of solvers the puzzle may use:
// this tier's SolverKeys plus all lower-tier solvers (derived from
// tierRegistry/tierOrder).
func (d Difficulty) AllowedSolverKeys() []string {
	lower := lowerTierSolverKeys(d.SolverKeys)
	all := make([]string, 0, len(lower)+len(d.SolverKeys))
	all = append(all, lower...)
	all = append(all, d.SolverKeys...)
	return all
}

// LowerTierSolverKeys returns the cumulative solver keys from all tiers
// below this difficulty's tier. Derived from tierRegistry/tierOrder.
func (d Difficulty) LowerTierSolverKeys() []string {
	return lowerTierSolverKeys(d.SolverKeys)
}

// NewEasyDifficulty creates the easy difficulty level.
// Easy puzzles are solvable using only naked singles and hidden singles.
func NewEasyDifficulty() Difficulty {
	return Difficulty{
		MinimumClues: 45,
		MaximumClues: 60,
		SolverKeys:   tierRegistry["easy"],
	}
}

// NewMediumDifficulty creates the medium difficulty level.
// Medium puzzles require at least one intermediate technique (naked-subset
// or pointing-pair) — basic techniques alone won't suffice.
func NewMediumDifficulty() Difficulty {
	return Difficulty{
		MinimumClues: 32,
		MaximumClues: 45,
		SolverKeys:   tierRegistry["medium"],
	}
}

// NewHardDifficulty creates the hard difficulty level.
// Hard puzzles require at least one advanced technique (X-Wing) —
// basic and intermediate techniques alone won't suffice.
func NewHardDifficulty() Difficulty {
	return Difficulty{
		MinimumClues: 25,
		MaximumClues: 32,
		SolverKeys:   tierRegistry["hard"],
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

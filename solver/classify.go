package solver

import "github.com/gnailuy/sudoku/core"

// ClassificationOutcome distinguishes a completed strategy solve from a
// puzzle on which the registered strategy inventory stalled.
type ClassificationOutcome string

const (
	ClassificationSolved           ClassificationOutcome = "solved"
	ClassificationStrategyUnsolved ClassificationOutcome = "strategy-unsolved"
)

// Classification holds the solver-relative strategy grade and trace metrics.
// The legacy Difficulty field name carries the highest required strategy tier.
type Classification struct {
	Outcome      ClassificationOutcome // Whether the registered strategies completed the puzzle.
	Difficulty   string                // Highest strategy tier reached; empty if no technique applied.
	Score        int                   // Deterministic score for ordering within a strategy grade.
	MaxTechnique string                // Technique from the highest tier reached.
	Moves        []Move                // All moves used during solving.
	Solved       bool                  // Compatibility mirror of Outcome == ClassificationSolved.
}

// ClassifyPuzzle solves the puzzle using strategy solvers from the store,
// determines its strategy grade, within-grade score, and highest technique. Strategy
// order is the store's stable registration order. A stalled classification is
// reported separately instead of being promoted to evil/backtracking.
func ClassifyPuzzle(store Store, board core.Board) Classification {
	testBoard := board // Board is a value type — this is a copy.
	var moves []Move

	for {
		var found bool
		for _, key := range store.GetAllStrategySolverKeys() {
			s := store.GetStrategySolverByKey(key)
			if s == nil {
				continue
			}
			move := s.Apply(&testBoard)
			if move == nil {
				continue
			}
			moves = append(moves, *move)
			if move.IsPlacement() {
				_ = testBoard.Set(move.Cell.Position, move.Cell.Value)
			}
			found = true
			break
		}
		if !found {
			break
		}
	}

	solved := testBoard.IsSolved()
	outcome := ClassificationStrategyUnsolved
	if solved {
		outcome = ClassificationSolved
	}

	maxTechnique := findMaxTechnique(moves)
	difficulty := determineDifficulty(maxTechnique)
	// A complete board needs no strategy move and belongs to the lowest tier.
	if solved && maxTechnique == "" {
		difficulty = "easy"
	}

	return Classification{
		Outcome:      outcome,
		Difficulty:   difficulty,
		Score:        ScorePuzzle(store, moves),
		MaxTechnique: maxTechnique,
		Moves:        moves,
		Solved:       solved,
	}
}

// findMaxTechnique returns the first move from the highest explicit technique
// tier reached. Move order is deterministic because strategy order is stable.
func findMaxTechnique(moves []Move) string {
	maxKey := ""
	maxTier := -1
	for _, move := range moves {
		_, tier, ok := StrategyTierForTechnique(move.Technique)
		if ok && tier > maxTier {
			maxTier = tier
			maxKey = move.Technique
		}
	}
	return maxKey
}

func determineDifficulty(maxTechnique string) string {
	tier, _, ok := StrategyTierForTechnique(maxTechnique)
	if !ok {
		return ""
	}
	return tier
}

package solver

import "github.com/gnailuy/sudoku/core"

// Classification holds the result of classifying a puzzle's difficulty
// by solving it with strategy solvers.
type Classification struct {
	Difficulty   string // Difficulty level name (easy/medium/hard/expert/evil).
	Score        int    // Total difficulty score.
	MaxTechnique string // Highest-tier technique used.
	Moves        []Move // All moves used during solving.
	Solved       bool   // Whether strategy solvers alone could solve the puzzle.
}

// ClassifyPuzzle solves the puzzle using strategy solvers from the store,
// determines the difficulty tier, score, and highest technique used.
// If strategy solvers cannot fully solve it, Solved is false — the
// puzzle may require backtracking.
func ClassifyPuzzle(store Store, board core.Board) Classification {
	testBoard := board // Board is a value type — this is a copy.
	var moves []Move

	// Repeatedly apply strategy solvers until no more progress is made.
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

	score := ScorePuzzle(store, moves)
	maxTechnique := findMaxTechnique(store, moves)
	difficulty := determineDifficulty(maxTechnique)

	return Classification{
		Difficulty:   difficulty,
		Score:        score,
		MaxTechnique: maxTechnique,
		Moves:        moves,
		Solved:       testBoard.IsSolved(),
	}
}

// findMaxTechnique returns the key of the highest-tier technique used,
// based on technique weight. Returns "backtracker" if no strategy
// solver moves were recorded.
func findMaxTechnique(store Store, moves []Move) string {
	maxKey := "backtracker"
	maxWeight := 0

	for _, m := range moves {
		s := store.GetStrategySolverByKey(m.Technique)
		if s == nil {
			continue
		}
		if s.GetWeight() > maxWeight {
			maxWeight = s.GetWeight()
			maxKey = m.Technique
		}
	}

	return maxKey
}

// techniqueTierMap maps solver keys to difficulty level names.
var techniqueTierMap = map[string]string{
	// Easy
	"naked-single":  "easy",
	"hidden-single": "easy",
	// Medium
	"naked-pair":    "medium",
	"naked-triple":  "medium",
	"pointing-pair": "medium",
	"hidden-pair":   "medium",
	// Hard
	"x-wing":        "hard",
	"xy-wing":       "hard",
	"hidden-triple": "hard",
	"w-wing":        "hard",
	// Expert
	"swordfish":       "expert",
	"naked-quad":      "expert",
	"simple-coloring": "expert",
	"hidden-quad":     "expert",
	"xyz-wing":        "expert",
	// Evil
	"jellyfish":            "evil",
	"bug-plus-one":         "evil",
	"unique-rectangle":     "evil",
	"unique-rectangle-2":   "evil",
	"unique-rectangle-3":   "evil",
	"unique-rectangle-4":   "evil",
	"x-cycles":             "evil",
	"xy-chain":             "evil",
}

// determineDifficulty returns the difficulty level name for the given
// max technique key. Defaults to "evil" for unknown techniques
// (including backtracker, which means strategy solvers couldn't solve it).
func determineDifficulty(maxTechnique string) string {
	if level, ok := techniqueTierMap[maxTechnique]; ok {
		return level
	}
	return "evil"
}

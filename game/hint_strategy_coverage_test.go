package game

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// These are complete, valid games from the maintained calibration corpus and
// catalog audit. Each one is selected because the canonical solver trace
// reaches the named strategy; no candidate-only synthetic board is used.
var realGameForHintStrategy = map[string]string{
	"naked-single":       "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3..",
	"hidden-single":      "2...8.3...6..7..84.3.5..2.9...1.54.8.........4.27.6...3.1..7.4.72..4..6...4.1...3",
	"naked-pair":         "4.....8.5.3..........7......2.....6.....8.4......1.......6.3.7.5..2.....1.4......",
	"naked-triple":       "52...6.........7.13...........4..8..6......5...........418.........3..2...87.....",
	"pointing-pair":      ".......1....3.2..4..57.4.3..2...5..85...2.....83.9..7...2.....37......6.9..5.34..",
	"hidden-pair":        "4.....8.5.3..........7......2.....6.....8.4......1.......6.3.7.5..2.....1.4......",
	"x-wing":             ".....1..2..134...535.62..4...6.3......34.5..7.1.8.639..34....19.69...2..5......76",
	"xy-wing":            ".....1..2..1.3.45.36..4.........4..3.1.35..7.4.8....2..56.1.9...746...8.9..4..2..",
	"hidden-triple":      ".............12.34...5.67.8....2..9..69..1..2.2463.87..51......6........4871..5..",
	"w-wing":             ".....1..2....2345....45..61..4....7..581.7..997..4..3..9..1.....42..65.36.7.3..2.",
	"swordfish":          "..............1234..2...516..7.2...3.1..73...2...4.17.7....89..5...1..478.69.....",
	"naked-quad":         ".............12.34.15.36.27..........23.67.58.563.42.1....7..9..61.......98..3..6",
	"simple-coloring":    ".....1..2....3..4..152..3.....6..5...478..6..6...4..7..96.8..3..78.25.963....6..7",
	"hidden-quad":        ".............12.34..53..6.......7.8....94.1.3693.........58.7.2..2......18...9.5.",
	"xyz-wing":           "........1......23...4..5...........6.5..7.189.86.23....9..8.....62.1945.4.83.6...",
	"jellyfish":          "........1..2.1345..6174.32.......5..1.8..4.6754.2.1.93.....86....4.....27.6.32...",
	"bug-plus-one":       "........1..1234...256..7.......2..5.....8.49.593.4..1.9.....5..3.....68.7..4691..",
	"unique-rectangle":   "...............123.14.567.8..9..8....26.719.58..5...7..3..6.....9..4.6.27..2.....",
	"unique-rectangle-2": ".............12.34...3.567............8.3...225.6.174..........937..6.8.682.7...1",
	"unique-rectangle-3": ".......12...1234..124...5....6.15.7..72..4.5.5..7...2..8..92..5.61.5...99..3....6",
	"unique-rectangle-4": "........1..1.23.4.24.5.6.7..6.....8.8......6771......3.2..549...5.98..241........",
	"x-cycles":           "........1.12.34...5.67.8......3...1..7..6.9.418....3.....14..3..5....4..2..9...67",
	"xy-chain":           ".......12.......34.....56.7..7682....92.3....64.1..3....8451....26.7....15.9..7..",
}

func TestEveryRegisteredStrategyBuildsTypedPlanFromRealGame(t *testing.T) {
	store := solver.NewStore()
	keys := store.GetAllStrategySolverKeys()
	if len(keys) != 23 {
		t.Fatalf("registered strategy count = %d, want 23", len(keys))
	}

	for _, key := range keys {
		key := key
		t.Run(key, func(t *testing.T) {
			puzzle, ok := realGameForHintStrategy[key]
			if !ok {
				t.Fatalf("missing real-game fixture for %s", key)
			}
			board := boardBeforeStrategy(t, puzzle, key, store)
			options := NewDefaultOptions(store)
			options.StrategySolverKeys = []string{key}
			game := NewGame(board, options)

			first, second := game.Hint(), game.Hint()
			if first == nil || second == nil {
				t.Fatal("expected complete plan")
			}
			if first.Strategy.ID != key {
				t.Fatalf("strategy = %q, want %q", first.Strategy.ID, key)
			}
			if first.PlanID == "" || first.PlanID != second.PlanID {
				t.Fatalf("plan identity is not deterministic: %q / %q", first.PlanID, second.PlanID)
			}
			if len(first.Steps) < 2 {
				t.Fatalf("steps = %d, want an ordered teaching sequence", len(first.Steps))
			}
			if !planHasRole(*first, HintRolePremise) {
				t.Fatal("plan has no typed candidate premise")
			}
			if !planHasUnit(*first) {
				t.Fatal("plan has no typed affected unit")
			}
			last := first.Steps[len(first.Steps)-1]
			if last.Effect == nil {
				t.Fatal("final step has no exact effect")
			}
			if first.Conclusion.Placement == nil && len(first.Conclusion.Eliminations) == 0 {
				t.Fatal("plan has no placement or elimination conclusion")
			}
		})
	}
}

func boardBeforeStrategy(t *testing.T, puzzle, target string, store solver.Store) core.Board {
	t.Helper()
	board := core.NewEmptyBoard()
	board.FromString(puzzle)
	keys := store.GetAllStrategySolverKeys()
	for round := 0; round < 10000; round++ {
		progress := false
		for _, key := range keys {
			strategy := store.GetStrategySolverByKey(key)
			trial := board.Copy()
			move := strategy.Apply(&trial)
			if move == nil {
				continue
			}
			if key == target {
				return board
			}
			if move.IsPlacement() {
				if err := trial.SetCell(move.Cell); err != nil {
					t.Fatalf("apply %s placement: %v", key, err)
				}
			}
			board = trial
			progress = true
			break
		}
		if !progress {
			break
		}
	}
	t.Fatalf("real game never reached strategy %s", target)
	return core.Board{}
}

func planHasUnit(plan HintPlan) bool {
	for _, step := range plan.Steps {
		for _, mark := range step.Marks {
			switch mark.Target.Kind {
			case HintTargetRow, HintTargetColumn, HintTargetBox:
				return true
			}
		}
	}
	return false
}

func planHasRole(plan HintPlan, role HintRole) bool {
	for _, step := range plan.Steps {
		for _, mark := range step.Marks {
			if mark.Role == role {
				return true
			}
		}
	}
	return false
}

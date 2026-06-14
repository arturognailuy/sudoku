package generator_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/generator"
	"github.com/gnailuy/sudoku/solver"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// solveWithKeys attempts to solve the board using only the specified strategy
// solvers. Returns true if the board is fully solved.
func solveWithKeys(board *core.Board, store solver.Store, keys []string) bool {
	var solvers []solver.StrategySolver
	for _, k := range keys {
		s := store.GetStrategySolverByKey(k)
		if s != nil {
			solvers = append(solvers, s)
		}
	}

	for {
		var found bool
		for _, s := range solvers {
			move := s.Apply(board)
			if move != nil {
				_ = board.Set(move.Cell.Position, move.Cell.Value)
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	return board.IsSolved()
}

// ---------------------------------------------------------------------------
// Board generation tests
// ---------------------------------------------------------------------------

// TestGenerateNormalizedSolvedBoard verifies the generator produces a valid,
// fully solved, normalized board.
func TestGenerateNormalizedSolvedBoard(t *testing.T) {
	store := solver.NewStore()
	options := generator.NewProblemOptions(store, generator.NewEasyDifficulty())

	board := generator.GenerateNormalizedSolvedBoard(options)

	if !board.IsSolved() {
		t.Error("Generated solved board is not fully solved")
	}
	if !board.IsValid() {
		t.Error("Generated solved board is not valid")
	}

	// Normalized board should have first row as 1-9.
	for col := 0; col < 9; col++ {
		if board.Get(core.NewPosition(0, col)) != col+1 {
			t.Errorf("Normalized board row 0, col %d: expected %d, got %d",
				col, col+1, board.Get(core.NewPosition(0, col)))
		}
	}
}

// ---------------------------------------------------------------------------
// Easy difficulty tests
// ---------------------------------------------------------------------------

// TestGenerateEasyPuzzle verifies Easy puzzles meet clue count constraints
// and are solvable by basic techniques.
func TestGenerateEasyPuzzle(t *testing.T) {
	store := solver.NewStore()
	difficulty := generator.NewEasyDifficulty()
	options := generator.NewProblemOptions(store, difficulty)

	for i := 0; i < 3; i++ {
		puzzle := generator.GenerateSudokuProblem(options)

		// Verify valid board.
		if !puzzle.IsValid() {
			t.Errorf("Puzzle %d: generated Easy puzzle is not valid", i)
		}

		// Verify clue count within range.
		clues := puzzle.GetFilledCellsCount()
		if clues < 45 || clues >= 60 {
			t.Errorf("Puzzle %d: Easy clue count %d outside range [45, 60)", i, clues)
		}

		// Verify solvable by basic techniques only.
		testBoard := puzzle
		if !solveWithKeys(&testBoard, store, []string{"naked-single", "hidden-single"}) {
			t.Errorf("Puzzle %d: Easy puzzle not solvable by basic techniques", i)
		}

		// Verify unique solution.
		solutions := store.GetDefaultSolver().CountSolutions(&puzzle)
		if solutions != 1 {
			t.Errorf("Puzzle %d: Easy puzzle has %d solutions, expected 1", i, solutions)
		}

		t.Logf("Easy puzzle %d: %d clues, valid, basic-solvable, unique", i, clues)
	}
}

// ---------------------------------------------------------------------------
// Medium difficulty tests
// ---------------------------------------------------------------------------

// TestGenerateMediumPuzzle verifies Medium puzzles meet clue count constraints
// and require at least one intermediate technique (basic-only solvers get stuck).
func TestGenerateMediumPuzzle(t *testing.T) {
	store := solver.NewStore()
	difficulty := generator.NewMediumDifficulty()
	options := generator.NewProblemOptions(store, difficulty)

	for i := 0; i < 3; i++ {
		puzzle := generator.GenerateSudokuProblem(options)

		// Verify valid board.
		if !puzzle.IsValid() {
			t.Errorf("Puzzle %d: generated Medium puzzle is not valid", i)
		}

		// Verify clue count within range.
		clues := puzzle.GetFilledCellsCount()
		if clues < 32 || clues >= 45 {
			t.Errorf("Puzzle %d: Medium clue count %d outside range [32, 45)", i, clues)
		}

		// Verify NOT solvable by basic techniques alone.
		basicBoard := puzzle
		if solveWithKeys(&basicBoard, store, []string{"naked-single", "hidden-single"}) {
			t.Errorf("Puzzle %d: Medium puzzle should NOT be solvable by basic techniques alone", i)
		}

		// Verify unique solution (backtracker can solve it).
		solutions := store.GetDefaultSolver().CountSolutions(&puzzle)
		if solutions != 1 {
			t.Errorf("Puzzle %d: Medium puzzle has %d solutions, expected 1", i, solutions)
		}

		t.Logf("Medium puzzle %d: %d clues, valid, requires-intermediate, unique", i, clues)
	}
}

// ---------------------------------------------------------------------------
// Difficulty constraint tests
// ---------------------------------------------------------------------------

// TestDifficultyClueCountRanges verifies that each difficulty level's clue
// count constraints are non-overlapping and ordered.
func TestDifficultyClueCountRanges(t *testing.T) {
	difficulties := []struct {
		name string
		diff generator.Difficulty
		min  int
		max  int
	}{
		{"Easy", generator.NewEasyDifficulty(), 45, 60},
		{"Medium", generator.NewMediumDifficulty(), 32, 45},
		{"Hard", generator.NewHardDifficulty(), 25, 32},
		{"Extreme", generator.NewExtremeDifficulty(), 20, 25},
		{"Evil", generator.NewEvilDifficulty(), 17, 20},
	}

	for _, d := range difficulties {
		t.Run(d.name, func(t *testing.T) {
			// Verify boundary values.
			if !d.diff.IsWithinDifficultyLevel(d.min) {
				t.Errorf("%s: minimum clue count %d should be within range", d.name, d.min)
			}
			if d.diff.IsWithinDifficultyLevel(d.max) {
				t.Errorf("%s: maximum clue count %d should be outside range (exclusive)", d.name, d.max)
			}
			if d.diff.IsWithinDifficultyLevel(d.min - 1) {
				t.Errorf("%s: below minimum %d should be outside range", d.name, d.min-1)
			}
		})
	}

	// Verify no overlap: each difficulty's max should equal the next level's min.
	for i := 0; i < len(difficulties)-1; i++ {
		current := difficulties[i]
		next := difficulties[i+1]
		if current.min != next.max {
			t.Errorf("%s min (%d) != %s max (%d): gap or overlap",
				current.name, current.min, next.name, next.max)
		}
	}
}

// TestMediumRequiredSolverKeysSet verifies that Medium difficulty has
// RequiredSolverKeys configured.
func TestMediumRequiredSolverKeysSet(t *testing.T) {
	diff := generator.NewMediumDifficulty()
	if len(diff.RequiredSolverKeys) == 0 {
		t.Error("Medium difficulty should have RequiredSolverKeys set")
	}

	// RequiredSolverKeys should be a subset of StrategySolverKeys.
	allowed := make(map[string]bool)
	for _, k := range diff.StrategySolverKeys {
		allowed[k] = true
	}
	for _, k := range diff.RequiredSolverKeys {
		if !allowed[k] {
			t.Errorf("RequiredSolverKey %q is not in StrategySolverKeys", k)
		}
	}
}

// TestEasyHasNoRequiredSolverKeys verifies Easy difficulty doesn't require
// any specific solver (all basic techniques are sufficient by definition).
func TestEasyHasNoRequiredSolverKeys(t *testing.T) {
	diff := generator.NewEasyDifficulty()
	if len(diff.RequiredSolverKeys) != 0 {
		t.Errorf("Easy difficulty should not have RequiredSolverKeys, got %v", diff.RequiredSolverKeys)
	}
}

// ---------------------------------------------------------------------------
// Input parsing tests
// ---------------------------------------------------------------------------

// TestGenerateFromValidString verifies parsing a valid puzzle string.
func TestGenerateFromValidString(t *testing.T) {
	input := "530070000600195000098000060800060003400803001700020006060000280000419005000080079"
	board, err := generator.GenerateSudokuProblemFromString(input)
	if err != nil {
		t.Fatalf("GenerateSudokuProblemFromString returned error: %v", err)
	}
	if board == nil {
		t.Fatal("Expected non-nil board")
	}
	if !board.IsValid() {
		t.Error("Parsed board should be valid")
	}
}

// TestGenerateFromInvalidString verifies error on invalid puzzle strings.
func TestGenerateFromInvalidString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"TooShort", "12345"},
		{"InvalidChars", "53007000060019500009800006080006000340080300170002000606000028000041900500008007!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := generator.GenerateSudokuProblemFromString(tt.input)
			if err == nil {
				t.Error("Expected error for invalid input")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Custom difficulty tests
// ---------------------------------------------------------------------------

// TestCustomDifficulty verifies that custom difficulty creates the expected
// configuration.
func TestCustomDifficulty(t *testing.T) {
	diff := generator.NewCustomDifficulty(30, 40, []string{"naked-single"})
	if diff.MinimumClues != 30 {
		t.Errorf("Expected min 30, got %d", diff.MinimumClues)
	}
	if diff.MaximumClues != 40 {
		t.Errorf("Expected max 40, got %d", diff.MaximumClues)
	}
	if len(diff.RequiredSolverKeys) != 0 {
		t.Errorf("Custom difficulty should have empty RequiredSolverKeys, got %v", diff.RequiredSolverKeys)
	}
}

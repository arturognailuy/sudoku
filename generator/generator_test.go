package generator

import (
	"testing"

	"github.com/gnailuy/sudoku/solver"
)

func TestGenerateBestEffortEasy(t *testing.T) {
	store := solver.NewStore()
	opts := NewBestEffortOptions(store, NewEasyDifficulty())
	opts.MaxRounds = 1
	opts.MaxDurationMs = 30000 // Keep the single real generation round bounded.

	result := GenerateBestEffort(opts)

	if result.RoundsUsed < 1 {
		t.Fatal("Expected at least 1 round used")
	}
	if result.DurationMs < 0 {
		t.Fatal("Expected non-negative duration")
	}

	// Easy puzzles should almost always be generated successfully.
	if !result.Matched {
		t.Logf("Warning: easy puzzle generation did not match target difficulty "+
			"(got %s), but this is acceptable for best-effort",
			result.Classification.Difficulty)
	}

	// Verify the puzzle is valid and solvable.
	if !result.Puzzle.IsValid() {
		t.Fatal("Generated puzzle is not valid")
	}
	solutions := store.GetDefaultSolver().CountSolutions(&result.Puzzle)
	if solutions != 1 {
		t.Fatalf("Expected 1 solution, got %d", solutions)
	}
}

func TestGenerateBestEffortStopsAfterTimeBudgetBetweenRounds(t *testing.T) {
	store := solver.NewStore()
	opts := NewBestEffortOptions(store, NewEasyDifficulty())
	opts.MaxRounds = 2
	opts.MaxDurationMs = 1

	result := GenerateBestEffort(opts)

	// Generation cannot interrupt a round already in progress. Once that
	// round returns, the expired budget must prevent another randomized round.
	if result.RoundsUsed != 1 {
		t.Fatalf("Expected the expired budget to stop after one round, used %d", result.RoundsUsed)
	}
}

func TestGenerateBestEffortRoundLimited(t *testing.T) {
	store := solver.NewStore()
	opts := NewBestEffortOptions(store, NewHardDifficulty())
	opts.MaxRounds = 1
	opts.MaxDurationMs = 30000 // Long time — round limit should hit first.

	result := GenerateBestEffort(opts)

	if result.RoundsUsed > 1 {
		t.Fatalf("Expected at most 1 round, used %d", result.RoundsUsed)
	}
}

func TestDifficultyLevelName(t *testing.T) {
	tests := []struct {
		difficulty Difficulty
		expected   string
	}{
		{NewEasyDifficulty(), "easy"},
		{NewMediumDifficulty(), "medium"},
		{NewHardDifficulty(), "hard"},
		{NewExpertDifficulty(), "expert"},
		{NewEvilDifficulty(), "evil"},
	}

	for _, tt := range tests {
		got := difficultyLevelName(tt.difficulty)
		if got != tt.expected {
			t.Errorf("difficultyLevelName() = %q, want %q", got, tt.expected)
		}
	}
}

func TestIsBetterMatch(t *testing.T) {
	// Medium is closer to hard than easy.
	if !isBetterMatch("medium", "easy", "hard") {
		t.Error("medium should be better than easy for hard target")
	}
	// Expert is closer to hard than easy.
	if !isBetterMatch("expert", "easy", "hard") {
		t.Error("expert should be better than easy for hard target")
	}
	// Hard is the best match for hard.
	if !isBetterMatch("hard", "easy", "hard") {
		t.Error("hard should be better than easy for hard target")
	}
}

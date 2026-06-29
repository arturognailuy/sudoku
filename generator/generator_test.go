package generator

import (
	"testing"

	"github.com/gnailuy/sudoku/solver"
)

func TestGenerateBestEffortEasy(t *testing.T) {
	store := solver.NewStore()
	opts := NewBestEffortOptions(store, NewEasyDifficulty())
	opts.MaxRounds = 5
	opts.MaxDurationMs = 10000 // 10 seconds

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

func TestGenerateBestEffortTimeLimited(t *testing.T) {
	store := solver.NewStore()
	opts := NewBestEffortOptions(store, NewEasyDifficulty())
	opts.MaxRounds = 100
	opts.MaxDurationMs = 200 // Very short — should stop early.

	result := GenerateBestEffort(opts)

	// Should respect the time limit. Allow some overhead for the round
	// that was in progress when the limit was hit.
	if result.DurationMs > 5000 {
		t.Fatalf("Expected generation to stop reasonably quickly, took %dms", result.DurationMs)
	}
	// The key assertion: rounds should be limited by the time budget.
	if result.RoundsUsed >= 100 {
		t.Fatalf("Expected time limit to stop generation before 100 rounds, used %d", result.RoundsUsed)
	}
}

func TestGenerateBestEffortRoundLimited(t *testing.T) {
	store := solver.NewStore()
	opts := NewBestEffortOptions(store, NewHardDifficulty())
	opts.MaxRounds = 2
	opts.MaxDurationMs = 30000 // Long time — round limit should hit first.

	result := GenerateBestEffort(opts)

	if result.RoundsUsed > 2 {
		t.Fatalf("Expected at most 2 rounds, used %d", result.RoundsUsed)
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

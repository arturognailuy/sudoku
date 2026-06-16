package solver

import "testing"

func TestScorePuzzleEmpty(t *testing.T) {
	store := NewStore()
	score := ScorePuzzle(store, nil)
	if score != 0 {
		t.Errorf("ScorePuzzle(nil) = %d, want 0", score)
	}

	score = ScorePuzzle(store, []Move{})
	if score != 0 {
		t.Errorf("ScorePuzzle([]) = %d, want 0", score)
	}
}

func TestScorePuzzleSingleTechnique(t *testing.T) {
	store := NewStore()
	moves := []Move{
		{Technique: "naked-single"},
		{Technique: "naked-single"},
		{Technique: "naked-single"},
	}

	score := ScorePuzzle(store, moves)
	// 3 × naked-single weight = 12
	want := 3 * WeightNakedSingle
	if score != want {
		t.Errorf("ScorePuzzle = %d, want %d", score, want)
	}
}

func TestScorePuzzleMixedTechniques(t *testing.T) {
	store := NewStore()
	moves := []Move{
		{Technique: "naked-single"},
		{Technique: "hidden-single"},
		{Technique: "naked-single"},
		{Technique: "x-wing"},
		{Technique: "naked-subset"},
	}

	score := ScorePuzzle(store, moves)
	want := 2*WeightNakedSingle + WeightHiddenSingle + WeightXWing + WeightNakedSubset
	if score != want {
		t.Errorf("ScorePuzzle = %d, want %d", score, want)
	}
}

func TestScorePuzzleIgnoresUnknownTechniques(t *testing.T) {
	store := NewStore()
	moves := []Move{
		{Technique: "naked-single"},
		{Technique: "backtracker"},
		{Technique: "unknown-solver"},
	}

	score := ScorePuzzle(store, moves)
	// Only naked-single counts.
	want := WeightNakedSingle
	if score != want {
		t.Errorf("ScorePuzzle = %d, want %d", score, want)
	}
}

func TestScorePuzzleAllSolvers(t *testing.T) {
	store := NewStore()

	// One move from each strategy solver.
	moves := []Move{
		{Technique: "naked-single"},
		{Technique: "hidden-single"},
		{Technique: "naked-subset"},
		{Technique: "pointing-pair"},
		{Technique: "x-wing"},
		{Technique: "swordfish"},
		{Technique: "hidden-subset"},
		{Technique: "xy-wing"},
		{Technique: "simple-coloring"},
	}

	score := ScorePuzzle(store, moves)
	want := WeightNakedSingle + WeightHiddenSingle + WeightNakedSubset +
		WeightPointingPair + WeightXWing + WeightSwordfish +
		WeightHiddenSubset + WeightXYWing + WeightSimpleColoring
	if score != want {
		t.Errorf("ScorePuzzle = %d, want %d", score, want)
	}
}

func TestSolverWeights(t *testing.T) {
	// Verify each solver has the expected weight.
	tests := []struct {
		name   string
		solver StrategySolver
		want   int
	}{
		{"naked-single", NewNakedSingleSolver(), WeightNakedSingle},
		{"hidden-single", NewHiddenSingleSolver(), WeightHiddenSingle},
		{"naked-subset", NewNakedSubsetSolver(), WeightNakedSubset},
		{"pointing-pair", NewPointingPairSolver(), WeightPointingPair},
		{"x-wing", NewXWingSolver(), WeightXWing},
		{"swordfish", NewSwordfishSolver(), WeightSwordfish},
		{"hidden-subset", NewHiddenSubsetSolver(), WeightHiddenSubset},
		{"xy-wing", NewXYWingSolver(), WeightXYWing},
		{"simple-coloring", NewSimpleColoringSolver(), WeightSimpleColoring},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.solver.GetWeight(); got != tt.want {
				t.Errorf("%s.GetWeight() = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestBacktrackerWeightIsZero(t *testing.T) {
	b := NewBacktracker()
	if got := b.GetWeight(); got != 0 {
		t.Errorf("Backtracker.GetWeight() = %d, want 0", got)
	}
}

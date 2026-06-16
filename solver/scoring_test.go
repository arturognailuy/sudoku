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
	// 3 × naked-single weight (4) = 12
	want := 3 * 4
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
	// 2×4 + 1×14 + 1×140 + 1×70 = 8 + 14 + 140 + 70 = 232
	want := 2*4 + 14 + 140 + 70
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
	// Only naked-single counts: 1×4 = 4
	want := 4
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
	// 4 + 14 + 70 + 50 + 140 + 150 + 100 + 160 + 150 = 838
	want := 4 + 14 + 70 + 50 + 140 + 150 + 100 + 160 + 150
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
		{"naked-single", NewNakedSingleSolver(), 4},
		{"hidden-single", NewHiddenSingleSolver(), 14},
		{"naked-subset", NewNakedSubsetSolver(), 70},
		{"pointing-pair", NewPointingPairSolver(), 50},
		{"x-wing", NewXWingSolver(), 140},
		{"swordfish", NewSwordfishSolver(), 150},
		{"hidden-subset", NewHiddenSubsetSolver(), 100},
		{"xy-wing", NewXYWingSolver(), 160},
		{"simple-coloring", NewSimpleColoringSolver(), 150},
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

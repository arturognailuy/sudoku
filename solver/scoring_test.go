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
		{Technique: "naked-pair"},
	}

	score := ScorePuzzle(store, moves)
	want := 2*WeightNakedSingle + WeightHiddenSingle + WeightXWing + WeightNakedPair
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
		{Technique: "naked-pair"},
		{Technique: "naked-triple"},
		{Technique: "pointing-pair"},
		{Technique: "hidden-pair"},
		{Technique: "x-wing"},
		{Technique: "xy-wing"},
		{Technique: "hidden-triple"},
		{Technique: "swordfish"},
		{Technique: "naked-quad"},
		{Technique: "simple-coloring"},
		{Technique: "hidden-quad"},
		{Technique: "jellyfish"},
		{Technique: "bug-plus-one"},
		{Technique: "unique-rectangle"},
	}

	score := ScorePuzzle(store, moves)
	want := WeightNakedSingle + WeightHiddenSingle +
		WeightNakedPair + WeightNakedTriple + WeightPointingPair + WeightHiddenPair +
		WeightXWing + WeightXYWing + WeightHiddenTriple +
		WeightSwordfish + WeightNakedQuad + WeightSimpleColoring + WeightHiddenQuad +
		WeightJellyfish + WeightBUGPlusOne + WeightUniqueRectangle
	if score != want {
		t.Errorf("ScorePuzzle = %d, want %d", score, want)
	}
}

func TestSolverWeights(t *testing.T) {
	tests := []struct {
		name   string
		solver StrategySolver
		want   int
	}{
		{"naked-single", NewNakedSingleSolver(), WeightNakedSingle},
		{"hidden-single", NewHiddenSingleSolver(), WeightHiddenSingle},
		{"naked-pair", NewNakedPairSolver(), WeightNakedPair},
		{"naked-triple", NewNakedTripleSolver(), WeightNakedTriple},
		{"naked-quad", NewNakedQuadSolver(), WeightNakedQuad},
		{"pointing-pair", NewPointingPairSolver(), WeightPointingPair},
		{"hidden-pair", NewHiddenPairSolver(), WeightHiddenPair},
		{"hidden-triple", NewHiddenTripleSolver(), WeightHiddenTriple},
		{"hidden-quad", NewHiddenQuadSolver(), WeightHiddenQuad},
		{"x-wing", NewXWingSolver(), WeightXWing},
		{"swordfish", NewSwordfishSolver(), WeightSwordfish},
		{"jellyfish", NewJellyfishSolver(), WeightJellyfish},
		{"xy-wing", NewXYWingSolver(), WeightXYWing},
		{"simple-coloring", NewSimpleColoringSolver(), WeightSimpleColoring},
		{"bug-plus-one", NewBUGPlusOneSolver(), WeightBUGPlusOne},
		{"unique-rectangle", NewUniqueRectangleSolver(), WeightUniqueRectangle},
		{"w-wing", NewWWingSolver(), WeightWWing},
		{"xyz-wing", NewXYZWingSolver(), WeightXYZWing},
		{"unique-rectangle-2", NewUniqueRectangleType2Solver(), WeightUniqueRectangle2},
		{"unique-rectangle-3", NewUniqueRectangleType3Solver(), WeightUniqueRectangle3},
		{"unique-rectangle-4", NewUniqueRectangleType4Solver(), WeightUniqueRectangle4},
		{"x-cycles", NewXCyclesSolver(), WeightXCycles},
		{"xy-chain", NewXYChainSolver(), WeightXYChain},
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

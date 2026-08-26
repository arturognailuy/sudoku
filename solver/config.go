// Package solver centralizes the tunable strategy-rating configuration.
// Technique weights produce deterministic within-grade scores, while clue
// ranges guide generation without assigning or overriding a strategy grade.
// Policy changes require the evidence and review defined in
// .aidoc/designs/difficulty-calibration.md.
//
// Solver weights are based on HoDoKu's established values. Each weight
// represents the relative cost of one technique application.
//
// Clue-count ranges bound the generator's search space. MinimumClues is
// inclusive and MaximumClues is exclusive.
package solver

// Solver weights define relative cost per technique application.
const (
	WeightNakedSingle      = 4
	WeightHiddenSingle     = 14
	WeightPointingPair     = 50
	WeightNakedPair        = 60
	WeightNakedTriple      = 80
	WeightNakedQuad        = 120
	WeightHiddenPair       = 70
	WeightHiddenTriple     = 100
	WeightHiddenQuad       = 150
	WeightXWing            = 140
	WeightSwordfish        = 150
	WeightJellyfish        = 300
	WeightXYWing           = 160
	WeightSimpleColoring   = 150
	WeightBUGPlusOne       = 250
	WeightUniqueRectangle  = 200
	WeightWWing            = 150
	WeightXYZWing          = 180
	WeightUniqueRectangle2 = 220
	WeightUniqueRectangle3 = 240
	WeightUniqueRectangle4 = 250
	WeightXCycles          = 280
	WeightXYChain          = 300
)

// Clue-count guidance per requested generation tier.
// MinimumClues is inclusive; MaximumClues is exclusive.
const (
	EasyMinClues   = 45
	EasyMaxClues   = 60
	MediumMinClues = 32
	MediumMaxClues = 45
	HardMinClues   = 25
	HardMaxClues   = 32
	ExpertMinClues = 22
	ExpertMaxClues = 25
	EvilMinClues   = 17
	EvilMaxClues   = 22
)

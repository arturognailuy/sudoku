// Package solver — config.go is the single centralized file for all
// tunable parameters in the Sudoku program. Update this file when
// calibrating difficulty, adjusting scoring weights, or changing
// clue-count ranges.
//
// Solver Weights: based on HoDoKu's established values. Each weight
// represents the difficulty cost per application of a technique.
//
// Clue-Count Ranges: define the acceptable number of given clues for
// each difficulty level. MinimumClues is inclusive, MaximumClues is
// exclusive. These serve as secondary constraints alongside technique
// requirements — a puzzle must fall within both the clue range and
// the technique tier for its difficulty level.
package solver

// Solver weights — difficulty cost per application.
const (
	WeightNakedSingle    = 4
	WeightHiddenSingle   = 14
	WeightPointingPair   = 50
	WeightNakedPair      = 60
	WeightNakedTriple    = 80
	WeightNakedQuad      = 120
	WeightHiddenPair     = 70
	WeightHiddenTriple   = 100
	WeightHiddenQuad     = 150
	WeightXWing          = 140
	WeightSwordfish      = 150
	WeightJellyfish      = 300
	WeightXYWing         = 160
	WeightSimpleColoring = 150
)

// Clue-count ranges per difficulty level.
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

package solver

// ScorePuzzle calculates a deterministic within-grade ordering score
// from the techniques used and their weights.
//
// Each move's technique is looked up in the Store to find its weight,
// and the total score is the sum of all weights. Moves whose technique
// is not found in the Store (e.g., backtracker) contribute zero.
func ScorePuzzle(store Store, moves []Move) int {
	score := 0
	for _, move := range moves {
		if s := store.GetStrategySolverByKey(move.Technique); s != nil {
			score += s.GetWeight()
		}
	}

	return score
}

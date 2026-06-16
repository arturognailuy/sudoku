package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// advancedSolverTestHelper applies solvers by key until no more progress.
// Used by the advanced solver tests (w-wing, xyz-wing, ur types 2-4, x-cycles, xy-chain).
func advancedSolverTestHelper(t *testing.T, board *core.Board, keys []string) {
	t.Helper()
	store := solver.NewStore()
	for {
		var found *solver.Move
		for _, k := range keys {
			sv := store.GetStrategySolverByKey(k)
			if sv == nil {
				continue
			}
			m := sv.Apply(board)
			if m != nil {
				found = m
				break
			}
		}
		if found == nil {
			break
		}
		if found.IsPlacement() {
			_ = board.Set(found.Cell.Position, found.Cell.Value)
		}
	}
}

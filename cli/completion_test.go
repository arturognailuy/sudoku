package cli

import (
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/game"
	"github.com/gnailuy/sudoku/playrun"
	"github.com/gnailuy/sudoku/solver"
)

type completionRecorder struct{ calls int }

func (recorder *completionRecorder) RecordCompletion(string) (bool, error) {
	recorder.calls++
	return true, nil
}

func TestControllerRecordsCompletionThroughTracker(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString(".23456789456789123789123456214365897365897214897214365531642978642978531978531642")
	current := game.NewGame(board, game.NewDefaultOptions(solver.NewStore()))
	recorder := &completionRecorder{}
	controller := NewTrackedController(&current, playrun.New("normalized", recorder))
	if !controller.RunCommand("add 1 1 1") || recorder.calls != 1 {
		t.Fatalf("completion calls=%d", recorder.calls)
	}
}

package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

func TestModelRecordsCompletionThroughTracker(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString(".23456789456789123789123456214365897365897214897214365531642978642978531978531642")
	current := game.NewGame(board, game.NewDefaultOptions(solver.NewStore()))
	recorder := &completionRecorder{}
	model := NewTrackedModelWithRecovery(current, "", RecoveryOptions{}, playrun.New("normalized", recorder))
	model.Update(tea.WindowSizeMsg{Width: minWidth, Height: minHeight})
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if updated.(Model).snapshot.Status != game.StatusSolved || recorder.calls != 1 {
		t.Fatalf("status=%s calls=%d", updated.(Model).snapshot.Status, recorder.calls)
	}
}

package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/game"
	"github.com/gnailuy/sudoku/solver"
)

const testPuzzle = "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."

func testModel(t *testing.T) Model {
	t.Helper()
	board := core.NewEmptyBoard()
	board.FromString(testPuzzle)
	current := game.NewGame(board, game.NewDefaultOptions(solver.NewStore()))
	model := NewModel(current, "")
	model.width, model.height = 80, 40
	return model
}

func sendKey(t *testing.T, model Model, key tea.KeyMsg) Model {
	t.Helper()
	updated, _ := model.Update(key)
	return updated.(Model)
}

func TestNavigationAndValueNoteModes(t *testing.T) {
	model := testModel(t)
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRight})
	if model.column != 1 {
		t.Fatalf("column=%d", model.column)
	}
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	if model.snapshot.Values[0][1] != 5 || !model.dirty {
		t.Fatal("value was not applied")
	}
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyDown})
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	if !model.snapshot.Notes[1][1].Has(4) {
		t.Fatal("note was not applied")
	}
}

func TestUndoRedoAndQuitConfirmation(t *testing.T) {
	model := testModel(t)
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	if model.snapshot.Values[0][0] != 0 {
		t.Fatal("undo failed")
	}
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if model.snapshot.Values[0][0] != 1 {
		t.Fatal("redo failed")
	}
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if model.modal != quitModal {
		t.Fatal("dirty quit did not ask for confirmation")
	}
}

func TestResizeFallbackAndStableMarkers(t *testing.T) {
	model := testModel(t)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 30, Height: 10})
	model = updated.(Model)
	if !strings.Contains(model.View(), "Terminal too small") {
		t.Fatal("missing resize fallback")
	}
	model.width, model.height = 80, 40
	view := model.View()
	for _, marker := range []string{"<3>", "{ . }", "mode:VALUE", "? preview hint"} {
		if !strings.Contains(view, marker) {
			t.Errorf("view missing %q", marker)
		}
	}
}

func TestFocusStopsAtBoardEdges(t *testing.T) {
	model := testModel(t)
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyUp})
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyLeft})
	if model.row != 0 || model.column != 0 {
		t.Fatalf("focus escaped top-left: %d,%d", model.row, model.column)
	}
	model.row, model.column = 8, 8
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyDown})
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRight})
	if model.row != 8 || model.column != 8 {
		t.Fatalf("focus escaped bottom-right: %d,%d", model.row, model.column)
	}
}

func TestSaveUsesSerializedSessionAndClearsDirty(t *testing.T) {
	model := testModel(t)
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	model.modal, model.savePath = saveModal, "game.json"
	var path string
	model.writeSession = func(got string, data []byte) error {
		path = got
		if len(data) == 0 {
			t.Fatal("empty serialized session")
		}
		return nil
	}
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if path != "game.json" || model.dirty || model.modal != noModal {
		t.Fatalf("save state path=%q dirty=%v modal=%v", path, model.dirty, model.modal)
	}
}

func TestHintPreviewDoesNotMutateUntilEnter(t *testing.T) {
	model := testModel(t)
	before := model.snapshot
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if model.hint == nil || model.snapshot != before {
		t.Fatal("hint preview missing or mutated game")
	}
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if model.snapshot == before || !model.dirty {
		t.Fatal("hint was not applied")
	}
}

func TestFailedSaveKeepsSessionDirty(t *testing.T) {
	model := testModel(t)
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	model.modal, model.savePath = saveModal, "game.json"
	model.writeSession = func(string, []byte) error { return errors.New("disk full") }
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if !model.dirty || !strings.Contains(model.message, "disk full") {
		t.Fatalf("failed save state dirty=%v message=%q", model.dirty, model.message)
	}
}

func TestSmallTerminalIgnoresGameInput(t *testing.T) {
	model := testModel(t)
	model.width, model.height = 30, 10
	before := model.snapshot
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if model.snapshot != before || model.dirty {
		t.Fatal("small terminal accepted game input")
	}
}

package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/game"
	"github.com/gnailuy/sudoku/sessionfile"
	"github.com/gnailuy/sudoku/solver"
)

const controllerTestPuzzle = "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."

func newTestController(t *testing.T) *Controller {
	t.Helper()
	board := core.NewEmptyBoard()
	board.FromString(controllerTestPuzzle)
	store := solver.NewStore()
	newGame := game.NewGame(board, game.NewDefaultOptions(store))
	return NewController(&newGame)
}

func TestParseDigitArguments(t *testing.T) {
	for _, input := range []string{"1 2 3", "123", "  1   2  3 "} {
		values, err := parseDigitArguments(input, 3)
		if err != nil || values[0] != 1 || values[1] != 2 || values[2] != 3 {
			t.Fatalf("parseDigitArguments(%q) = %v, %v", input, values, err)
		}
	}
	for _, input := range []string{"", "1 2", "1 2 3 4", "10 2 3", "a 2 3"} {
		if _, err := parseDigitArguments(input, 3); err == nil {
			t.Errorf("parseDigitArguments(%q) unexpectedly succeeded", input)
		}
	}
}

func TestCommandsRejectExtraArguments(t *testing.T) {
	controller := newTestController(t)
	if controller.RunCommand("undo now") {
		t.Fatal("undo with extra arguments changed the board")
	}
	if controller.game.Snapshot().CanUndo {
		t.Fatal("undo with extra arguments was applied")
	}
}

func TestNoteCommandsAndRendering(t *testing.T) {
	controller := newTestController(t)
	if !controller.RunCommand("note 1 1 1") || !controller.RunCommand("n 1 1 9") {
		t.Fatal("note commands did not report a board change")
	}
	snapshot := controller.game.Snapshot()
	if !snapshot.Notes[0][0].Has(1) || !snapshot.Notes[0][0].Has(9) {
		t.Fatal("notes were not toggled")
	}
	rendered := renderBoard(snapshot)
	if !strings.Contains(rendered, "1  ") || !strings.Contains(rendered, "  9") {
		t.Fatalf("note board does not preserve candidate positions:\n%s", rendered)
	}
	if !controller.RunCommand("x 1 1") || !controller.game.Snapshot().Notes[0][0].IsEmpty() {
		t.Fatal("notes-clear did not clear the cell")
	}
	if strings.Contains(renderBoard(controller.game.Snapshot()), "+-----------+") {
		t.Fatal("board without notes did not return to compact rendering")
	}
}

func TestSaveCommandRoundTrip(t *testing.T) {
	controller := newTestController(t)
	if !controller.RunCommand("n 1 1 5") {
		t.Fatal("note command failed")
	}
	path := filepath.Join(t.TempDir(), "saved.json")
	if controller.RunCommand("save " + path) {
		t.Fatal("save should not report a board change")
	}
	data, err := sessionfile.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	store := solver.NewStore()
	restored, err := game.Restore(data, game.NewDefaultOptions(store))
	if err != nil {
		t.Fatal(err)
	}
	if !restored.Snapshot().Notes[0][0].Has(5) {
		t.Fatal("saved session did not preserve notes")
	}
}

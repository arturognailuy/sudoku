package tui

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
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
	model.width, model.height = 80, 42
	model.theme = darkTheme
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

func TestResizeFallbackAndCleanStableLayout(t *testing.T) {
	model := testModel(t)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 30, Height: 10})
	model = updated.(Model)
	if !strings.Contains(model.View(), "Terminal too small") {
		t.Fatal("missing resize fallback")
	}
	model.width, model.height = 80, 42
	view := ansi.Strip(model.View())
	for _, marker := range []string{"  3  ", "VALUE  •", "┏", "╋", "i hint", "? help"} {
		if !strings.Contains(view, marker) {
			t.Errorf("view missing %q", marker)
		}
	}
	for _, oldMarker := range []string{"<3>", "{", "}", "( . )", " . "} {
		if strings.Contains(view, oldMarker) {
			t.Errorf("view retained old cell marker %q", oldMarker)
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
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if model.hint == nil || model.snapshot != before {
		t.Fatal("hint preview missing or mutated game")
	}
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if model.snapshot == before || !model.dirty {
		t.Fatal("hint was not applied")
	}
}

func TestHelpOverlayOpensAndCloses(t *testing.T) {
	model := testModel(t)
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if model.modal != helpModal || !strings.Contains(ansi.Strip(model.View()), "KEYBOARD HELP") {
		t.Fatal("help overlay did not open")
	}
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyEscape})
	if model.modal != noModal {
		t.Fatal("help overlay did not close")
	}
}

func TestSemanticBackgroundsUseDeterministicThemeColors(t *testing.T) {
	model := testModel(t)
	view := model.View()
	for _, pattern := range []string{`\x1b\[[0-9;]*48;2;117;91;24m`, `\x1b\[[0-9;]*48;2;27;48;56m`} {
		if !regexp.MustCompile(pattern).MatchString(view) {
			t.Fatalf("dark view missing background pattern %q", pattern)
		}
	}
	model.theme = lightTheme
	view = model.View()
	for _, pattern := range []string{`\x1b\[[0-9;]*48;2;246;195;83m`, `\x1b\[[0-9;]*48;2;232;242;243m`} {
		if !regexp.MustCompile(pattern).MatchString(view) {
			t.Fatalf("light view missing background pattern %q", pattern)
		}
	}
}

func TestDeterministicThemeSelectionAndNoColorFallback(t *testing.T) {
	t.Setenv("SUDOKU_THEME", "light")
	if got := themeFromEnvironment(); got != lightTheme {
		t.Fatalf("light theme=%v", got)
	}
	t.Setenv("NO_COLOR", "1")
	if got := themeFromEnvironment(); got != noColorTheme {
		t.Fatalf("NO_COLOR theme=%v", got)
	}

	model := testModel(t)
	model.theme = noColorTheme
	model = sendKey(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	view := model.View()
	for _, colorCode := range []string{"\x1b[30m", "\x1b[31m", "\x1b[32m", "\x1b[33m", "\x1b[34m", "\x1b[35m", "\x1b[36m", "\x1b[37m", "\x1b[38;", "\x1b[40m", "\x1b[41m", "\x1b[42m", "\x1b[43m", "\x1b[44m", "\x1b[45m", "\x1b[46m", "\x1b[47m", "\x1b[48;"} {
		if strings.Contains(view, colorCode) {
			t.Fatalf("no-color view contains %q", colorCode)
		}
	}
	if !regexp.MustCompile(`\x1b\[[0-9;]*7[0-9;]*m`).MatchString(view) {
		t.Fatal("no-color focus is not distinguished with reverse video")
	}
	if !regexp.MustCompile(`\x1b\[[0-9;]*4[0-9;]*m`).MatchString(view) {
		t.Fatal("no-color invalid value is not underlined")
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

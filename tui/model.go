// Package tui implements the opt-in Bubble Tea terminal frontend.
package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/game"
	"github.com/gnailuy/sudoku/sessionfile"
	"github.com/gnailuy/sudoku/solver"
)

const (
	minWidth  = 54
	minHeight = 38
)

type inputMode uint8

const (
	valueMode inputMode = iota
	noteMode
)

type modalKind uint8

const (
	noModal modalKind = iota
	quitModal
	resetModal
	saveModal
)

// Model owns presentation state while delegating all puzzle transitions to game.Game.
type Model struct {
	game          *game.Game
	snapshot      game.Snapshot
	row, column   int
	width, height int
	mode          inputMode
	modal         modalKind
	message       string
	hint          *solver.Move
	dirty         bool
	savePath      string
	writeSession  func(string, []byte) error
}

// NewModel creates a TUI model. resumePath becomes the default save target.
func NewModel(current game.Game, resumePath string) Model {
	return Model{game: &current, snapshot: current.Snapshot(), savePath: resumePath, writeSession: sessionfile.Write}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = typed.Width, typed.Height
		return m, nil
	case tea.KeyMsg:
		if m.width > 0 && (m.width < minWidth || m.height < minHeight) {
			if typed.String() == "q" || typed.String() == "esc" || typed.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}
		if typed.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.modal != noModal {
			return m.updateModal(typed)
		}
		return m.updateBoard(typed)
	}
	return m, nil
}

func (m Model) updateBoard(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.message = ""
	switch key.String() {
	case "up", "k":
		if m.row > 0 {
			m.row--
		}
	case "down", "j":
		if m.row < 8 {
			m.row++
		}
	case "left", "h":
		if m.column > 0 {
			m.column--
		}
	case "right", "l":
		if m.column < 8 {
			m.column++
		}
	case "n":
		if m.mode == valueMode {
			m.mode = noteMode
		} else {
			m.mode = valueMode
		}
	case "u":
		m.apply(game.Undo{})
	case "r":
		m.apply(game.Redo{})
	case "?":
		m.hint = m.game.Hint()
		if m.hint == nil {
			m.message = "No hint is available."
		} else {
			m.message = "Hint preview: " + m.hint.String()
		}
	case "enter":
		if m.hint != nil {
			m.apply(game.ApplyHint{})
			m.hint = nil
		} else {
			m.message = "Press ? to preview a hint first."
		}
	case "c":
		m.snapshot = m.game.Snapshot()
		m.message = "Board is " + string(m.snapshot.Status) + "."
	case "R":
		m.modal = resetModal
	case "S":
		m.modal = saveModal
	case "q", "esc":
		if m.dirty {
			m.modal = quitModal
		} else {
			return m, tea.Quit
		}
	case "0", "backspace", "delete":
		position := core.NewPosition(m.row, m.column)
		if m.mode == noteMode {
			m.apply(game.ClearNotes{Position: position})
		} else {
			m.apply(game.ClearValue{Position: position})
		}
	default:
		if len(key.Runes) == 1 && key.Runes[0] >= '1' && key.Runes[0] <= '9' {
			value := int(key.Runes[0] - '0')
			position := core.NewPosition(m.row, m.column)
			if m.mode == noteMode {
				m.apply(game.ToggleNote{Position: position, Value: value})
			} else {
				m.apply(game.SetValue{Position: position, Value: value})
			}
		}
	}
	return m, nil
}

func (m Model) updateModal(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.modal == saveModal {
		switch key.String() {
		case "esc":
			m.modal = noModal
		case "enter":
			if strings.TrimSpace(m.savePath) == "" {
				m.message = "Save path cannot be empty."
				return m, nil
			}
			data, err := m.game.Serialize()
			if err == nil {
				err = m.writeSession(m.savePath, data)
			}
			if err != nil {
				m.message = "Save failed: " + err.Error()
			} else {
				m.dirty = false
				m.message = "Saved to " + m.savePath
			}
			m.modal = noModal
		case "backspace":
			if m.savePath != "" {
				_, size := utf8.DecodeLastRuneInString(m.savePath)
				m.savePath = m.savePath[:len(m.savePath)-size]
			}
		default:
			if len(key.Runes) > 0 {
				m.savePath += string(key.Runes)
			}
		}
		return m, nil
	}
	switch key.String() {
	case "y", "Y":
		if m.modal == quitModal {
			return m, tea.Quit
		}
		m.modal = noModal
		m.apply(game.Reset{})
	case "n", "N", "esc":
		m.modal = noModal
	}
	return m, nil
}

func (m *Model) apply(action game.Action) {
	result, err := m.game.Apply(action)
	if err != nil {
		m.message = err.Error()
		return
	}
	m.snapshot = m.game.Snapshot()
	m.dirty = true
	m.hint = nil
	m.message = fmt.Sprintf("Applied %s; board is %s.", result.Action, result.Status)
}

func (m Model) View() string {
	if m.width > 0 && (m.width < minWidth || m.height < minHeight) {
		return fmt.Sprintf("Sudoku TUI\n\nTerminal too small: %dx%d. Resize to at least %dx%d.\n", m.width, m.height, minWidth, minHeight)
	}
	return render(m)
}

package tui

import (
	"fmt"
	"strings"
)

const (
	bold  = "\x1b[1m"
	dim   = "\x1b[2m"
	reset = "\x1b[0m"
)

func render(m Model) string {
	var out strings.Builder
	mode := "VALUE"
	if m.mode == noteMode {
		mode = "NOTE"
	}
	fmt.Fprintf(&out, "%sSudoku%s  mode:%s  status:%s  focus:r%dc%d", bold, reset, mode, m.snapshot.Status, m.row+1, m.column+1)
	if m.dirty {
		out.WriteString("  unsaved:*")
	}
	out.WriteString("\n")
	out.WriteString(dim + "given <5>  player 5  invalid !5!  focus {...}  peer (...)" + reset + "\n")
	out.WriteString(renderBoard(m))
	out.WriteString("\n")
	if m.message != "" {
		out.WriteString(m.message + "\n")
	}
	switch m.modal {
	case quitModal:
		out.WriteString("Unsaved changes. Quit without saving? [y/N]\n")
	case resetModal:
		out.WriteString("Reset the puzzle and clear history? [y/N]\n")
	case saveModal:
		fmt.Fprintf(&out, "Save session path: %s_  (Enter saves, Esc cancels)\n", m.savePath)
	default:
		out.WriteString("arrows/hjkl move  1-9 set  0 clear  n notes  u/r undo/redo\n")
		out.WriteString("? preview hint  Enter apply hint  c check  S save  R reset  q quit\n")
	}
	return out.String()
}

func renderBoard(m Model) string {
	var out strings.Builder
	for row := 0; row < 9; row++ {
		if row > 0 {
			if row%3 == 0 {
				out.WriteString(strings.Repeat("━", 53))
			} else {
				out.WriteString(strings.Repeat("─", 53))
			}
			out.WriteByte('\n')
		}
		for noteRow := 0; noteRow < 3; noteRow++ {
			for column := 0; column < 9; column++ {
				if column > 0 {
					if column%3 == 0 {
						out.WriteString("┃")
					} else {
						out.WriteString("│")
					}
				}
				content := cellContent(m, row, column, noteRow)
				left, right := ' ', ' '
				if row == m.row && column == m.column {
					left, right = '{', '}'
				} else if row == m.row || column == m.column || (row/3 == m.row/3 && column/3 == m.column/3) {
					left, right = '(', ')'
				}
				fmt.Fprintf(&out, "%c%s%c", left, content, right)
			}
			out.WriteByte('\n')
		}
	}
	return strings.TrimSuffix(out.String(), "\n")
}

func cellContent(m Model, row, column, noteRow int) string {
	value := m.snapshot.Values[row][column]
	if value != 0 {
		if noteRow != 1 {
			return "   "
		}
		if m.snapshot.Invalid[row][column] {
			return fmt.Sprintf("!%d!", value)
		}
		if m.snapshot.Givens[row][column] != 0 {
			return fmt.Sprintf("<%d>", value)
		}
		return fmt.Sprintf(" %d ", value)
	}
	notes := m.snapshot.Notes[row][column]
	if notes.IsEmpty() {
		if noteRow == 1 {
			return " . "
		}
		return "   "
	}
	var content [3]byte
	for i := range content {
		content[i] = ' '
	}
	for offset := 0; offset < 3; offset++ {
		candidate := noteRow*3 + offset + 1
		if notes.Has(candidate) {
			content[offset] = byte('0' + candidate)
		}
	}
	return string(content[:])
}

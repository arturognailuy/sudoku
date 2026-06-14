package solver

import (
	"fmt"

	"github.com/gnailuy/sudoku/core"
)

// Move represents a solver's recommended action on the board, including
// which technique found it and a human-readable explanation.
type Move struct {
	Cell      core.Cell // The cell to fill (position + value).
	Technique string    // Technique identifier, e.g. "backtracker", "naked-single".
	Reason    string    // Human-readable explanation for hints.
}

// String returns a display-friendly representation of the move.
func (m Move) String() string {
	return fmt.Sprintf("%s at %s → %d (%s)",
		m.Technique, m.Cell.Position.ToString(), m.Cell.Value, m.Reason)
}

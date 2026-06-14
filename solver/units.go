package solver

import (
	"fmt"

	"github.com/gnailuy/sudoku/core"
)

// unit represents a Sudoku unit (row, column, or box) with its positions and name.
type unit struct {
	positions []core.Position
	name      string
}

// allUnits returns all 27 Sudoku units: 9 rows, 9 columns, 9 boxes.
func allUnits() []unit {
	units := make([]unit, 0, 27)

	// Rows.
	for row := 0; row < 9; row++ {
		positions := make([]core.Position, 9)
		for col := 0; col < 9; col++ {
			positions[col] = core.NewPosition(row, col)
		}
		units = append(units, unit{positions: positions, name: fmt.Sprintf("row %d", row+1)})
	}

	// Columns.
	for col := 0; col < 9; col++ {
		positions := make([]core.Position, 9)
		for row := 0; row < 9; row++ {
			positions[row] = core.NewPosition(row, col)
		}
		units = append(units, unit{positions: positions, name: fmt.Sprintf("column %d", col+1)})
	}

	// Boxes.
	for boxRow := 0; boxRow < 3; boxRow++ {
		for boxCol := 0; boxCol < 3; boxCol++ {
			positions := make([]core.Position, 0, 9)
			startRow, startCol := boxRow*3, boxCol*3
			for r := startRow; r < startRow+3; r++ {
				for c := startCol; c < startCol+3; c++ {
					positions = append(positions, core.NewPosition(r, c))
				}
			}
			units = append(units, unit{positions: positions, name: fmt.Sprintf("box %d", boxRow*3+boxCol+1)})
		}
	}

	return units
}

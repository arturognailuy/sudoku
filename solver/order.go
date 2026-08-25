package solver

import (
	"sort"

	"github.com/gnailuy/sudoku/core"
)

func sortedPositions[T any](values map[core.Position]T) []core.Position {
	positions := make([]core.Position, 0, len(values))
	for position := range values {
		positions = append(positions, position)
	}
	sort.Slice(positions, func(i, j int) bool {
		if positions[i].Row != positions[j].Row {
			return positions[i].Row < positions[j].Row
		}
		return positions[i].Column < positions[j].Column
	})
	return positions
}

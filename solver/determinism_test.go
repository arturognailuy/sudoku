package solver

import (
	"reflect"
	"testing"

	"github.com/gnailuy/sudoku/core"
)

func TestClassifyPuzzleIsDeterministicAcrossColoringComponents(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString("517...9..........39...2..6..247........89..1.1......3.2.5....4....47.6....6.53..8")
	want := ClassifyPuzzle(NewStore(), board)

	for repeat := 0; repeat < 25; repeat++ {
		got := ClassifyPuzzle(NewStore(), board)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("classification repeat %d differed\nwant: %+v\ngot:  %+v", repeat+1, want, got)
		}
	}
}

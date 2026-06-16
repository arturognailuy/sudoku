package solver_test

import (
	"testing"

	"github.com/gnailuy/sudoku/solver"
)

func TestUniqueRectangleSolver_KeyAndMetadata(t *testing.T) {
	s := solver.NewUniqueRectangleSolver()
	if s.GetKey() != "unique-rectangle" {
		t.Errorf("expected key %q, got %q", "unique-rectangle", s.GetKey())
	}
	if s.GetDisplayName() != "Unique Rectangle Type 1" {
		t.Errorf("expected display name %q, got %q", "Unique Rectangle Type 1", s.GetDisplayName())
	}
	if s.GetWeight() != solver.WeightUniqueRectangle {
		t.Errorf("expected weight %d, got %d", solver.WeightUniqueRectangle, s.GetWeight())
	}
}

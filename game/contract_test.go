package game

import (
	"errors"
	"testing"

	"github.com/gnailuy/sudoku/core"
)

func TestSnapshotIsDetached(t *testing.T) {
	game := newTestGame()
	position := core.NewPosition(0, 2)

	snapshot := game.Snapshot()
	snapshot.Givens[0][0] = 0
	snapshot.Values[0][2] = 9
	snapshot.Invalid[0][2] = true

	fresh := game.Snapshot()
	if fresh.Givens[0][0] != 5 {
		t.Fatal("mutating snapshot givens changed the game")
	}
	if fresh.Values[0][2] != 0 || fresh.Invalid[0][2] {
		t.Fatal("mutating snapshot state changed the game")
	}

	playBoard := game.PlayBoard()
	_ = playBoard.Set(position, 9)
	if game.Get(position) != 0 {
		t.Fatal("mutating PlayBoard copy changed the game")
	}
}

func TestApplyValueActionsAndHistory(t *testing.T) {
	game := newTestGame()
	position := core.NewPosition(0, 2)

	result, err := game.Apply(SetValue{Position: position, Value: 4})
	if err != nil {
		t.Fatalf("Apply(SetValue) returned error: %v", err)
	}
	assertSingleChange(t, result, position, 0, 4)
	if result.Action != ActionSetValue || !result.CanUndo || result.CanRedo {
		t.Fatalf("unexpected set result: %+v", result)
	}

	result, err = game.Apply(Undo{})
	if err != nil {
		t.Fatalf("Apply(Undo) returned error: %v", err)
	}
	assertSingleChange(t, result, position, 4, 0)
	if !result.CanRedo {
		t.Fatal("undo result should allow redo")
	}

	result, err = game.Apply(Redo{})
	if err != nil {
		t.Fatalf("Apply(Redo) returned error: %v", err)
	}
	assertSingleChange(t, result, position, 0, 4)

	result, err = game.Apply(ClearValue{Position: position})
	if err != nil {
		t.Fatalf("Apply(ClearValue) returned error: %v", err)
	}
	assertSingleChange(t, result, position, 4, 0)
}

func TestApplyReturnsTypedErrorsWithoutMutation(t *testing.T) {
	game := newTestGame()
	before := game.Snapshot()

	tests := []struct {
		name   string
		action Action
		code   ErrorCode
	}{
		{name: "invalid value", action: SetValue{Position: core.NewPosition(0, 2), Value: 10}, code: ErrorInvalidCell},
		{name: "given", action: SetValue{Position: core.NewPosition(0, 0), Value: 1}, code: ErrorImmutableCell},
		{name: "empty undo", action: Undo{}, code: ErrorNoUndo},
		{name: "empty redo", action: Redo{}, code: ErrorNoRedo},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := game.Apply(test.action)
			if !errors.Is(err, &EngineError{Code: test.code}) {
				t.Fatalf("expected error code %q, got %v", test.code, err)
			}
			if after := game.Snapshot(); after != before {
				t.Fatalf("failed action mutated game\nbefore: %+v\nafter: %+v", before, after)
			}
		})
	}
}

func TestApplyHintAndReset(t *testing.T) {
	game := newTestGame()

	result, err := game.Apply(ApplyHint{})
	if err != nil {
		t.Fatalf("Apply(ApplyHint) returned error: %v", err)
	}
	if result.Action != ActionApplyHint || len(result.Changes) != 1 || !result.CanUndo {
		t.Fatalf("unexpected hint result: %+v", result)
	}

	result, err = game.Apply(Reset{})
	if err != nil {
		t.Fatalf("Apply(Reset) returned error: %v", err)
	}
	if result.Action != ActionReset || result.CanUndo || result.CanRedo {
		t.Fatalf("unexpected reset result: %+v", result)
	}
	if game.Snapshot().Status != StatusInProgress {
		t.Fatal("reset game should be in progress")
	}
}

func TestApplyNilAction(t *testing.T) {
	game := newTestGame()
	before := game.Snapshot()
	_, err := game.Apply(nil)
	if !errors.Is(err, &EngineError{Code: ErrorInvalidAction}) {
		t.Fatalf("expected invalid action error, got %v", err)
	}
	if game.Snapshot() != before {
		t.Fatal("nil action mutated game")
	}
}

func assertSingleChange(t *testing.T, result Result, position core.Position, before, after int) {
	t.Helper()
	if len(result.Changes) != 1 {
		t.Fatalf("expected one cell change, got %+v", result.Changes)
	}
	change := result.Changes[0]
	if change.Position != position || change.Before != before || change.After != after {
		t.Fatalf("unexpected cell change: %+v", change)
	}
}

func TestNoteActionsAndUnifiedHistory(t *testing.T) {
	game := newTestGame()
	position := core.NewPosition(0, 2)

	result, err := game.Apply(ToggleNote{Position: position, Value: 4})
	if err != nil {
		t.Fatalf("Apply(ToggleNote) returned error: %v", err)
	}
	if result.Action != ActionToggleNote || !result.Changes[0].NotesAfter.Has(4) {
		t.Fatalf("unexpected toggle result: %+v", result)
	}
	if !game.Snapshot().Notes[0][2].Has(4) {
		t.Fatal("note was not stored in snapshot")
	}

	if _, err = game.Apply(SetValue{Position: position, Value: 4}); err != nil {
		t.Fatalf("Apply(SetValue) returned error: %v", err)
	}
	if !game.Snapshot().Notes[0][2].IsEmpty() {
		t.Fatal("setting a value should clear notes in that cell")
	}

	if _, err = game.Apply(Undo{}); err != nil {
		t.Fatalf("undo value returned error: %v", err)
	}
	snapshot := game.Snapshot()
	if snapshot.Values[0][2] != 0 || !snapshot.Notes[0][2].Has(4) {
		t.Fatalf("undo did not restore value and notes atomically: %+v", snapshot)
	}

	if _, err = game.Apply(Undo{}); err != nil {
		t.Fatalf("undo note returned error: %v", err)
	}
	if !game.Snapshot().Notes[0][2].IsEmpty() {
		t.Fatal("second undo should remove the note")
	}

	if _, err = game.Apply(Redo{}); err != nil {
		t.Fatalf("redo note returned error: %v", err)
	}
	if !game.Snapshot().Notes[0][2].Has(4) {
		t.Fatal("redo should restore the note")
	}
}

func TestSetValueCleansPeerNotesAtomically(t *testing.T) {
	game := newTestGame()
	target := core.NewPosition(0, 2)
	peers := []core.Position{
		core.NewPosition(0, 3), // row
		core.NewPosition(1, 2), // column and box
	}
	nonPeer := core.NewPosition(4, 1)

	for _, position := range append(peers, target, nonPeer) {
		if _, err := game.Apply(ToggleNote{Position: position, Value: 4}); err != nil {
			t.Fatalf("add note at %v: %v", position, err)
		}
	}

	if _, err := game.Apply(SetValue{Position: target, Value: 4}); err != nil {
		t.Fatalf("set value: %v", err)
	}
	snapshot := game.Snapshot()
	if !snapshot.Notes[target.Row][target.Column].IsEmpty() {
		t.Fatal("target notes were not cleared")
	}
	for _, position := range peers {
		if snapshot.Notes[position.Row][position.Column].Has(4) {
			t.Fatalf("peer note was not removed at %v", position)
		}
	}
	if !snapshot.Notes[nonPeer.Row][nonPeer.Column].Has(4) {
		t.Fatal("non-peer note should be preserved")
	}

	if _, err := game.Apply(Undo{}); err != nil {
		t.Fatalf("undo set value: %v", err)
	}
	snapshot = game.Snapshot()
	for _, position := range append(peers, target, nonPeer) {
		if !snapshot.Notes[position.Row][position.Column].Has(4) {
			t.Fatalf("undo did not restore note at %v", position)
		}
	}
}

func TestClearNotesAndRedoTruncation(t *testing.T) {
	game := newTestGame()
	position := core.NewPosition(0, 2)
	if _, err := game.Apply(ToggleNote{Position: position, Value: 3}); err != nil {
		t.Fatal(err)
	}
	if _, err := game.Apply(ToggleNote{Position: position, Value: 4}); err != nil {
		t.Fatal(err)
	}
	if _, err := game.Apply(ClearNotes{Position: position}); err != nil {
		t.Fatal(err)
	}
	if !game.Snapshot().Notes[0][2].IsEmpty() {
		t.Fatal("clear notes did not clear all notes")
	}
	if _, err := game.Apply(Undo{}); err != nil {
		t.Fatal(err)
	}
	if _, err := game.Apply(ToggleNote{Position: position, Value: 5}); err != nil {
		t.Fatal(err)
	}
	if _, err := game.Apply(Redo{}); !errors.Is(err, &EngineError{Code: ErrorNoRedo}) {
		t.Fatalf("new note action should truncate redo history, got %v", err)
	}
}

func TestNoteErrorsDoNotMutateGame(t *testing.T) {
	game := newTestGame()
	filled := core.NewPosition(0, 2)
	if _, err := game.Apply(SetValue{Position: filled, Value: 4}); err != nil {
		t.Fatal(err)
	}
	before := game.Snapshot()

	tests := []struct {
		action Action
		code   ErrorCode
	}{
		{ToggleNote{Position: core.NewPosition(0, 0), Value: 4}, ErrorImmutableCell},
		{ToggleNote{Position: filled, Value: 4}, ErrorNoteNotAllowed},
		{ToggleNote{Position: core.NewPosition(0, 3), Value: 0}, ErrorInvalidCell},
		{ClearNotes{Position: core.Position{Row: 9, Column: 0}}, ErrorInvalidCell},
	}
	for _, test := range tests {
		if _, err := game.Apply(test.action); !errors.Is(err, &EngineError{Code: test.code}) {
			t.Fatalf("expected %s, got %v", test.code, err)
		}
		if after := game.Snapshot(); after != before {
			t.Fatalf("failed note action mutated game\nbefore: %+v\nafter: %+v", before, after)
		}
	}
}

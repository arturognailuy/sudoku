package game

import (
	"fmt"

	"github.com/gnailuy/sudoku/core"
)

// Status describes the current state of a game session.
type Status string

const (
	StatusInProgress Status = "in-progress"
	StatusInvalid    Status = "invalid"
	StatusSolved     Status = "solved"
)

// ErrorCode identifies an expected engine failure without requiring callers
// to parse user-facing text.
type ErrorCode string

const (
	ErrorInvalidAction ErrorCode = "invalid-action"
	ErrorInvalidCell   ErrorCode = "invalid-cell"
	ErrorImmutableCell ErrorCode = "immutable-cell"
	ErrorNoUndo        ErrorCode = "no-undo"
	ErrorNoRedo        ErrorCode = "no-redo"
	ErrorNoHint        ErrorCode = "no-hint"
)

// EngineError is returned for invalid player actions. The game is unchanged
// when Apply returns an EngineError.
type EngineError struct {
	Code     ErrorCode
	Position *core.Position
	Detail   string
}

func (err *EngineError) Error() string {
	if err.Detail != "" {
		return err.Detail
	}
	return string(err.Code)
}

// Is allows errors.Is to match engine errors by code.
func (err *EngineError) Is(target error) bool {
	other, ok := target.(*EngineError)
	return ok && err.Code == other.Code
}

// ActionKind identifies a transition accepted by the engine.
type ActionKind string

const (
	ActionSetValue   ActionKind = "set-value"
	ActionClearValue ActionKind = "clear-value"
	ActionReset      ActionKind = "reset"
	ActionUndo       ActionKind = "undo"
	ActionRedo       ActionKind = "redo"
	ActionApplyHint  ActionKind = "apply-hint"
)

// Action is a typed player intent. Its unexported method keeps the set of
// accepted actions closed to this package.
type Action interface {
	actionKind() ActionKind
}

// SetValue sets an editable cell to a digit from 1 through 9.
type SetValue struct {
	Position core.Position
	Value    int
}

func (SetValue) actionKind() ActionKind { return ActionSetValue }

// ClearValue clears an editable cell.
type ClearValue struct{ Position core.Position }

func (ClearValue) actionKind() ActionKind { return ActionClearValue }

// Reset restores the puzzle and clears action history.
type Reset struct{}

func (Reset) actionKind() ActionKind { return ActionReset }

// Undo reverses the most recent recorded action.
type Undo struct{}

func (Undo) actionKind() ActionKind { return ActionUndo }

// Redo reapplies the next action in history.
type Redo struct{}

func (Redo) actionKind() ActionKind { return ActionRedo }

// ApplyHint applies the current hint as a recorded value action.
type ApplyHint struct{}

func (ApplyHint) actionKind() ActionKind { return ActionApplyHint }

// Snapshot is a detached read model. Its arrays may be modified by a caller
// without mutating the Game that produced it.
type Snapshot struct {
	Givens  [9][9]int
	Values  [9][9]int
	Invalid [9][9]bool
	Status  Status
	CanUndo bool
	CanRedo bool
}

// CellChange describes one visible cell changed by an accepted action.
type CellChange struct {
	Position      core.Position
	Before        int
	After         int
	InvalidBefore bool
	InvalidAfter  bool
}

// Result describes an accepted transition.
type Result struct {
	Action  ActionKind
	Changes []CellChange
	Status  Status
	CanUndo bool
	CanRedo bool
}

// Snapshot returns a detached representation suitable for rendering.
func (game *Game) Snapshot() Snapshot {
	var snapshot Snapshot
	for row := 0; row < 9; row++ {
		for column := 0; column < 9; column++ {
			position := core.NewPosition(row, column)
			snapshot.Givens[row][column] = game.problemBoard.Get(position)
			snapshot.Values[row][column] = game.Get(position)
			snapshot.Invalid[row][column] = game.invalidInput.Get(position) != 0
		}
	}
	snapshot.Status = game.status()
	snapshot.CanUndo = game.inputCursor >= 0
	snapshot.CanRedo = game.inputCursor < len(game.inputSequence)-1
	return snapshot
}

// ProblemBoard returns a detached copy of the immutable puzzle givens.
func (game *Game) ProblemBoard() core.Board { return game.problemBoard.Copy() }

// PlayBoard returns a detached copy of the solver-safe current board. Invalid
// player entries are represented separately in Snapshot.
func (game *Game) PlayBoard() core.Board { return game.playBoard.Copy() }

// Apply validates and applies one typed player action atomically.
func (game *Game) Apply(action Action) (Result, error) {
	before := game.Snapshot()
	if action == nil {
		return Result{}, &EngineError{Code: ErrorInvalidAction, Detail: "action is required"}
	}

	kind := action.actionKind()
	var err error
	switch typed := action.(type) {
	case SetValue:
		if !typed.Position.IsValid() || typed.Value < 1 || typed.Value > 9 {
			return Result{}, invalidCellError(typed.Position, typed.Value)
		}
		err = game.AddInputAndRecordHistory(core.Cell{Position: typed.Position, Value: typed.Value})
	case ClearValue:
		if !typed.Position.IsValid() {
			return Result{}, invalidCellError(typed.Position, 0)
		}
		err = game.AddInputAndRecordHistory(core.Cell{Position: typed.Position, Value: 0})
	case Reset:
		game.Reset()
	case Undo:
		err = game.Undo()
	case Redo:
		err = game.Redo()
	case ApplyHint:
		hint := game.Hint()
		if hint == nil {
			err = &EngineError{Code: ErrorNoHint, Detail: "no hint is available"}
		} else {
			err = game.AddInputAndRecordHistory(hint.Cell)
		}
	default:
		return Result{}, &EngineError{Code: ErrorInvalidAction, Detail: fmt.Sprintf("unsupported action %T", action)}
	}
	if err != nil {
		return Result{}, err
	}

	after := game.Snapshot()
	return resultFromSnapshots(kind, before, after), nil
}

func (game *Game) status() Status {
	if game.IsSolved() {
		return StatusSolved
	}
	if !game.IsValid() {
		return StatusInvalid
	}
	return StatusInProgress
}

func resultFromSnapshots(kind ActionKind, before, after Snapshot) Result {
	result := Result{
		Action:  kind,
		Status:  after.Status,
		CanUndo: after.CanUndo,
		CanRedo: after.CanRedo,
		Changes: make([]CellChange, 0),
	}
	for row := 0; row < 9; row++ {
		for column := 0; column < 9; column++ {
			if before.Values[row][column] == after.Values[row][column] &&
				before.Invalid[row][column] == after.Invalid[row][column] {
				continue
			}
			result.Changes = append(result.Changes, CellChange{
				Position:      core.NewPosition(row, column),
				Before:        before.Values[row][column],
				After:         after.Values[row][column],
				InvalidBefore: before.Invalid[row][column],
				InvalidAfter:  after.Invalid[row][column],
			})
		}
	}
	return result
}

func invalidCellError(position core.Position, value int) *EngineError {
	positionCopy := position
	return &EngineError{
		Code:     ErrorInvalidCell,
		Position: &positionCopy,
		Detail:   fmt.Sprintf("invalid cell at row %d, column %d with value %d", position.Row, position.Column, value),
	}
}

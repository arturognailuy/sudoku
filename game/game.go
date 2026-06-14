package game

import (
	"errors"
	"fmt"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
)

// MoveRecord stores a player move along with the previous cell value for undo support.
type MoveRecord struct {
	Input         core.Cell
	PreviousValue int
}

// Game holds the state for an interactive Sudoku session.
type Game struct {
	// Public fields.
	ProblemBoard core.Board // The problem board. Read-only.
	PlayBoard    core.Board // The board that the user can play with.

	// Private fields.
	invalidInput    core.Board              // Put the invalid input in another board to keep the play board solvable.
	inputSequence   []MoveRecord            // User input sequence.
	inputCursor     int                     // The cursor of the current user input.
	completeSolver  solver.CompleteSolver    // The complete solver for judging input and solving, must be reliable.
	strategySolvers []solver.StrategySolver  // An optional list of strategy solvers to give hints.
}

// NewGame creates a new game from a problem board and options.
func NewGame(problem core.Board, options Options) Game {
	if !problem.IsValid() {
		panic("Bug: Invalid problem board when creating a new Sudoku game")
	}

	return Game{
		ProblemBoard:    problem,
		PlayBoard:       problem.Copy(),
		invalidInput:    core.NewEmptyBoard(),
		inputSequence:   []MoveRecord{},
		inputCursor:     -1,
		completeSolver:  options.solverStore.GetDefaultSolver(),
		strategySolvers: options.GetStrategySolvers(),
	}
}

// Function to count the solutions of the current play board using the complete solver.
func (game *Game) countSolutions() int {
	return game.completeSolver.CountSolutions(&game.PlayBoard)
}

// Function to add a non-zero cell input.
func (game *Game) addNonZeroInput(input core.Cell) {
	if input.Value == 0 {
		panic("Bug: Cannot add a zero input with this function")
	}

	_ = game.PlayBoard.SetCell(input)       // cell validated by caller
	game.invalidInput.Unset(input.Position) // Reset the invalid input state when adding a new input.

	if game.countSolutions() <= 0 {
		// Store the invalid input in the invalidInput board and unset the cell in the play board.
		game.PlayBoard.Unset(input.Position)
		_ = game.invalidInput.SetCell(input) // cell validated by caller
	}
}

// Function to add a zero.
func (game *Game) addZeroInput(input core.Cell) {
	if input.Value != 0 {
		panic("Bug: Cannot add a non-zero input with this function")
	}

	game.PlayBoard.Unset(input.Position)
	game.invalidInput.Unset(input.Position) // Reset the invalid input state when adding a new input.

	// If the board has multiple solutions, we need to check if any previously invalid input is now valid.
	if !game.invalidInput.IsEmpty() && game.countSolutions() > 1 {
		for i := 0; i < 9; i++ {
			for j := 0; j < 9; j++ {
				value := game.invalidInput.Get(core.NewPosition(i, j))
				if value != 0 {
					// Try to add the previously invalid input to the play board.
					game.addNonZeroInput(core.NewCell(core.NewPosition(i, j), value))
				}
			}
		}
	}
}

// Function to get the cell value of the game boards.
func (game *Game) Get(position core.Position) int {
	if game.PlayBoard.Get(position) != 0 {
		return game.PlayBoard.Get(position)
	} else {
		return game.invalidInput.Get(position)
	}
}

// Function to add a cell input.
func (game *Game) AddInput(input core.Cell) (err error) {
	if !input.IsValid() {
		panic("Bug: Invalid input when adding input. Check user input before calling this function")
	}

	if game.ProblemBoard.Get(input.Position) != 0 {
		err = errors.New("cannot change the value of a problem cell")
		return
	}

	if input.Value == 0 {
		game.addZeroInput(input)
	} else {
		game.addNonZeroInput(input)
	}

	return
}

// Function to add a cell input and record the history.
func (game *Game) AddInputAndRecordHistory(input core.Cell) (err error) {
	previousValue := game.Get(input.Position)

	err = game.AddInput(input)
	if err != nil {
		return
	}

	// On new input, we remove all the input after the cursor.
	if len(game.inputSequence) > game.inputCursor+1 {
		game.inputSequence = game.inputSequence[:game.inputCursor+1]
	}

	// Then append the new input to the input sequence.
	game.inputSequence = append(game.inputSequence, MoveRecord{
		Input:         input,
		PreviousValue: previousValue,
	})
	game.inputCursor++

	return
}

// Function to undo the last cell input.
func (game *Game) Undo() (err error) {
	if game.inputCursor < 0 {
		err = errors.New("no input to undo")
		return
	}

	lastInput := game.inputSequence[game.inputCursor]
	game.inputCursor--

	_ = game.AddInput(core.Cell{
		Position: lastInput.Input.Position,
		Value:    lastInput.PreviousValue,
	})

	return
}

// Function to redo the last undone cell input.
func (game *Game) Redo() (err error) {
	if game.inputCursor >= len(game.inputSequence)-1 {
		err = errors.New("no input to redo")
		return
	}

	game.inputCursor++
	nextInput := game.inputSequence[game.inputCursor]

	_ = game.AddInput(nextInput.Input)

	return
}

// Function to repair the game to the last valid state.
func (game *Game) Repair() (undoSteps int) {
	for !game.IsValid() && game.inputCursor >= 0 {
		undoSteps++
		_ = game.Undo()
	}

	return undoSteps
}

// Function to reset the game to the initial state.
func (game *Game) Reset() {
	game.PlayBoard = game.ProblemBoard.Copy()
	game.invalidInput = core.NewEmptyBoard()
	game.inputSequence = []MoveRecord{}
	game.inputCursor = -1
}

// Function to solve the game.
func (game *Game) Solve() {
	game.completeSolver.Solve(&game.PlayBoard)
}

// Hint returns the next recommended move.
// It first checks for invalid inputs to clear, then tries strategy solvers,
// and falls back to the complete solver.
func (game *Game) Hint() *solver.Move {
	// If there is any invalid input, randomly remove one of them.
	if !game.invalidInput.IsEmpty() {
		positionPointer := game.invalidInput.GetRandomPositionWith(func(value int) bool {
			return value != 0
		})

		if positionPointer == nil {
			panic("Bug: Invalid input board is not empty but cannot find a valid position")
		}

		return &solver.Move{
			Cell: core.Cell{
				Position: *positionPointer,
				Value:    0,
			},
			Technique: "clear-invalid",
			Reason:    fmt.Sprintf("clear invalid input at %s", positionPointer.ToString()),
		}
	}

	// If any of the strategy solvers can find a move, use it.
	for _, s := range game.strategySolvers {
		move := s.Apply(&game.PlayBoard)
		if move != nil {
			return move
		}
	}

	// Otherwise, get a hint from the complete solver.
	return game.completeSolver.Hint(&game.PlayBoard)
}

// Function to check if the game is solved.
func (game *Game) IsSolved() bool {
	return game.PlayBoard.IsSolved()
}

// Function to check if the game is in a valid state.
func (game *Game) IsValid() bool {
	return game.invalidInput.IsEmpty()
}

// Function to print the Sudoku game to string.
func (game *Game) ToString() string {
	result := "Problem:\n"
	result += game.ProblemBoard.ToString()
	result += "\n"

	playBoardCopy := game.PlayBoard.Copy()
	playBoardCopy.Merge(game.invalidInput)

	status := "Valid"
	if game.IsSolved() {
		status = "Solved"
	} else if !game.IsValid() {
		status = "Invalid"
	}

	if playBoardCopy != game.ProblemBoard {
		result += "Current board (" + status + "):\n"
		result += playBoardCopy.ToString()
		result += "\n"
	}

	return result
}

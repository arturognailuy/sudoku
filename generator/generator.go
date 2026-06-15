package generator

import (
	"errors"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/solver"
	"github.com/gnailuy/sudoku/util"
)

// Function to generate a solved Sudoku board by solving an empty normalized board randomly.
func GenerateNormalizedSolvedBoard(options Options) core.Board {
	// The first row of a normalize empty board is always from 1 to 9.
	board := core.NewEmptyBoard()
	for col := 0; col < 9; col++ {
		if err := board.Set(core.NewPosition(0, col), col+1); err != nil {
			panic("Bug: " + err.Error())
		}
	}

	// To generate a solved board from an empty normalized board, we use the reliable default solver.
	defaultSolver := options.solverStore.GetDefaultSolver()
	defaultSolver.Solve(&board)

	return board
}

// Function to generate a Sudoku problem from a solved board.
func GenerateSudokuProblemFromSolvedBoard(board core.Board, options Options) core.Board {
	if !board.IsSolved() || !board.IsValid() {
		panic("Bug: The board is not solved or not valid to generate a problem")
	}

	// Initially, all cells are filled.
	nonEmptyPositions := make([]core.Position, 0)
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			nonEmptyPositions = append(nonEmptyPositions, core.NewPosition(row, col))
		}
	}

	// Remove numbers randomly from the solved board to create a problem.
	cluesNumberReached := false
	for i := 0; i < options.MaximumIterations; i++ {
		// Check if the number of clues reached the difficulty level.
		if options.Difficulty.IsWithinDifficultyLevel(board.GetFilledCellsCount()) {
			cluesNumberReached = true
		}

		if cluesNumberReached {
			// Stop if removing more numbers will exceed the difficulty level.
			if !options.Difficulty.IsWithinDifficultyLevel(board.GetFilledCellsCount() - 1) {
				break
			}

			// Use a simple geometric distribution to stop removing numbers with a probability of P.
			// The expected number of iterations after the difficulty level is reached will be 1/P.
			if util.RandomBool(0.125) {
				break
			}
		}

		// Stop removing numbers because it is impossible to have a unique solution with less than 17 filled cells.
		if options.MaximumSolutions == 1 && board.GetFilledCellsCount() <= 17 {
			break
		}

		// Test the non-empty positions in a random order and unset the first one that can be removed.
		util.ShuffleArray(nonEmptyPositions)

		removedPositionIndex := -1
		for j, position := range nonEmptyPositions {
			// Temporarily store the cell value.
			originalValue := board.Get(position)

			// Update the board.
			board.Unset(position)

			// Find out the maximum number of solutions using the default solver.
			numberOfSolutions := options.solverStore.GetDefaultSolver().CountSolutions(&board)

			// Check if the problem is solvable and has no more than maximum solutions.
			if numberOfSolutions > 0 && numberOfSolutions <= options.MaximumSolutions {
				canProgress := false

				allowedKeys := options.Difficulty.AllowedSolverKeys()
				if len(allowedKeys) > 0 {
					// If there are strategy solvers configured, we limit the problem to be solvable with the specified strategies.
					// Test the strategy solvers to ensure that at least one of them can make progress.
					for _, key := range allowedKeys {
						strategySolver := options.solverStore.GetStrategySolverByKey(key)
						if strategySolver == nil {
							panic("Bug: Invalid strategy solver key: " + key)
						}

						move := strategySolver.Apply(&board)
						if move != nil {
							canProgress = true
							break
						}
					}
				} else {
					// If there are no strategy solvers configured, we don't care about limiting the problem to specific strategies.
					// And the default solver can always make progress.
					canProgress = true
				}

				// Confirm the removal.
				if canProgress {
					removedPositionIndex = j
					break
				}
			}

			// If the problem is not solvable or has more than maximum solutions, revert the removal.
			_ = board.Set(position, originalValue) // value was just read from this position
		}

		// Remove the position from the non-empty positions list.
		if removedPositionIndex >= 0 {
			nonEmptyPositions = append(nonEmptyPositions[:removedPositionIndex], nonEmptyPositions[removedPositionIndex+1:]...)
		} else {
			// We did not find any position to remove in this iteration, so we stop the process.
			break
		}
	}

	return board
}

// Function to generate a Sudoku problem.
func GenerateSudokuProblem(options Options) core.Board {
	for {
		solvedBoard := GenerateNormalizedSolvedBoard(options)
		solvedBoard.Randomize()

		problem := GenerateSudokuProblemFromSolvedBoard(solvedBoard, options)

		if requiresThisTierSolver(problem, options) {
			return problem
		}
		// Lower-tier solvers alone can solve it — regenerate.
	}
}

// requiresThisTierSolver checks whether the puzzle requires at least one
// solver from this tier's SolverKeys. If there are no lower-tier solvers
// (lowest tier or unconstrained), any puzzle qualifies.
//
// The check works by attempting to solve the puzzle using only the
// lower-tier solvers (derived from tierRegistry/tierOrder). If those can fully
// solve the puzzle, it doesn't genuinely require this tier's techniques.
func requiresThisTierSolver(board core.Board, options Options) bool {
	lowerKeys := options.Difficulty.LowerTierSolverKeys()

	// If there are no lower-tier solvers, this is the lowest tier
	// (or unconstrained) — every puzzle qualifies.
	if len(lowerKeys) == 0 {
		return true
	}

	// Collect the lower-tier solvers.
	var lowerSolvers []solver.StrategySolver
	for _, key := range lowerKeys {
		s := options.solverStore.GetStrategySolverByKey(key)
		if s != nil {
			lowerSolvers = append(lowerSolvers, s)
		}
	}

	// Try to solve the puzzle using only lower-tier solvers.
	testBoard := board // copy (Board is a value type)
	for {
		var found bool
		for _, s := range lowerSolvers {
			move := s.Apply(&testBoard)
			if move != nil {
				if move.IsPlacement() {
					_ = testBoard.Set(move.Cell.Position, move.Cell.Value)
				}
				// Elimination-only moves are progress too.
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	// If lower-tier solvers alone can solve it, the puzzle doesn't
	// require any solver from this tier.
	return !testBoard.IsSolved()
}

// Function to generate a Sudoku problem from an input string.
func GenerateSudokuProblemFromString(input string) (boardPointer *core.Board, err error) {
	if !core.IsValidSudokuString(input) {
		return nil, errors.New("invalid Sudoku string: " + input)
	}

	board := core.NewEmptyBoard()
	board.FromString(input)

	if !board.IsValid() {
		return nil, errors.New("invalid Sudoku board: " + input)
	}

	boardPointer = &board

	return
}

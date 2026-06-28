package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnailuy/sudoku/cli"
	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/db"
	"github.com/gnailuy/sudoku/game"
	"github.com/gnailuy/sudoku/generator"
	"github.com/gnailuy/sudoku/solver"
)

// defaultDBPath returns the default puzzle database path.
// It uses $XDG_DATA_HOME/sudoku/puzzles.db if set,
// otherwise ~/.local/share/sudoku/puzzles.db.
func defaultDBPath() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "puzzles.db" // fallback to current directory
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "sudoku", "puzzles.db")
}

func main() {
	// Create and initialize the solver store.
	solverStore := solver.NewStore()

	// Parse the command line options.
	options := cli.NewCommandLineOptions()
	options.Parse()

	if *options.HelpRequested {
		cli.PrintHelp()
		os.Exit(0)
	}

	if *options.Input != "" {
		// Read the input as a Sudoku string
		problem, err := generator.GenerateSudokuProblemFromString(*options.Input)

		if err != nil {
			fmt.Fprintf(os.Stderr, "The input is not a valid Sudoku problem: %s\n", *options.Input)
			os.Exit(1)
		}

		solutionCount := solverStore.GetDefaultSolver().CountSolutions(problem)
		if solutionCount == 0 {
			fmt.Fprintf(os.Stderr, "The input is not a solvable Sudoku problem: %s\n", *options.Input)
			os.Exit(1)
		} else if solutionCount > 1 {
			fmt.Fprintf(os.Stderr, "The input has %d solutions: %s\n", solutionCount, *options.Input)
		}

		// Auto-store the input puzzle.
		autoStore(solverStore, *problem, "input")

		playCli(*problem, solverStore, solverStore.GetAllStrategySolverKeys())
	} else {
		// Generate a random problem using best-effort with DB fallback.
		difficulty := options.GetDifficultyOptions()
		levelName := options.Level.String()
		fmt.Printf("Generating a random %s Sudoku problem...\n", levelName)

		problem, keys := generateWithFallback(solverStore, difficulty, levelName)

		playCli(problem, solverStore, keys)
	}
}

// generateWithFallback tries to generate a puzzle at the target difficulty,
// falling back to the database if the generator can't match the target.
func generateWithFallback(solverStore solver.Store, difficulty generator.Difficulty, levelName string) (core.Board, []string) {
	// Best-effort generation.
	opts := generator.NewBestEffortOptions(solverStore, difficulty)
	result := generator.GenerateBestEffort(opts)

	// Auto-store the generated puzzle regardless.
	autoStore(solverStore, result.Puzzle, "generated")

	if result.Matched {
		// Generator matched the target difficulty.
		keys := difficulty.AllowedSolverKeys()
		if len(keys) == 0 {
			keys = solverStore.GetAllStrategySolverKeys()
		}
		return result.Puzzle, keys
	}

	// Generator didn't match — try DB fallback.
	dbPath := defaultDBPath()
	puzzleDB, err := db.Open(dbPath)
	if err == nil {
		defer puzzleDB.Close()

		if dbPuzzle, err := puzzleDB.GetRandom(levelName); err == nil && dbPuzzle != nil {
			// Found a puzzle in the database at the target difficulty.
			board := core.NewEmptyBoard()
			board.FromString(dbPuzzle.Puzzle)
			board.Randomize()
			keys := difficulty.AllowedSolverKeys()
			if len(keys) == 0 {
				keys = solverStore.GetAllStrategySolverKeys()
			}
			return board, keys
		}
	}

	// No DB match — use best-effort result with mismatch warning.
	actualLevel := result.Classification.Difficulty
	if actualLevel != levelName {
		fmt.Printf("Requested difficulty: %s. Generated puzzle difficulty: %s. Enjoy!\n",
			capitalize(levelName), capitalize(actualLevel))
	}

	keys := difficulty.AllowedSolverKeys()
	if len(keys) == 0 {
		keys = solverStore.GetAllStrategySolverKeys()
	}
	return result.Puzzle, keys
}

// autoStore stores a puzzle in the database in normalized form.
// Errors are silently ignored — this is best-effort background storage.
func autoStore(solverStore solver.Store, board core.Board, source string) {
	// Solve the board to get the full solution for normalization.
	solvedBoard := board.Copy()
	solverStore.GetDefaultSolver().Solve(&solvedBoard)
	if !solvedBoard.IsSolved() {
		return // unsolvable — don't store
	}

	// Normalize the solved board to get the normalization mapping.
	normalizedSolved := solvedBoard.Copy()
	normalizedSolved.Normalize()

	// Build the digit mapping from the solved board.
	var digitMap [10]int // digitMap[original] = normalized
	for col := 0; col < 9; col++ {
		original := solvedBoard.Get(core.NewPosition(0, col))
		normalized := normalizedSolved.Get(core.NewPosition(0, col))
		digitMap[original] = normalized
	}

	// Apply the same mapping to the puzzle board.
	normalizedPuzzle := core.NewEmptyBoard()
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			pos := core.NewPosition(row, col)
			val := board.Get(pos)
			if val != 0 {
				_ = normalizedPuzzle.Set(pos, digitMap[val])
			}
		}
	}

	puzzleStr := normalizedPuzzle.ToString()

	// Classify the puzzle.
	classification := solver.ClassifyPuzzle(solverStore, board)

	// Open DB and insert.
	dbPath := defaultDBPath()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return
	}
	puzzleDB, err := db.Open(dbPath)
	if err != nil {
		return
	}
	defer puzzleDB.Close()

	_, _ = puzzleDB.InsertPuzzle(db.Puzzle{
		Puzzle:       puzzleStr,
		Difficulty:   classification.Difficulty,
		Score:        classification.Score,
		MaxTechnique: classification.MaxTechnique,
		Source:       source,
	})
}

// Function to play a game in CLI.
func playCli(problem core.Board, solverStore solver.Store, strategySolverKeys []string) {
	opts := game.NewDefaultOptions(solverStore)
	opts.StrategySolverKeys = strategySolverKeys
	newGame := game.NewGame(problem, opts)
	ctrl := cli.NewController(&newGame)
	ctrl.Play()
}

// capitalize returns a string with the first letter uppercased.
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

package cmd

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
	"github.com/spf13/cobra"
)

func runPlay(cmd *cobra.Command) {
	input, _ := cmd.Flags().GetString("input")
	level, _ := cmd.Flags().GetString("level")

	if input != "" {
		problem, err := generator.GenerateSudokuProblemFromString(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "The input is not a valid Sudoku problem: %s\n", input)
			os.Exit(1)
		}

		solutionCount := solverStore.GetDefaultSolver().CountSolutions(problem)
		if solutionCount == 0 {
			fmt.Fprintf(os.Stderr, "The input is not a solvable Sudoku problem: %s\n", input)
			os.Exit(1)
		} else if solutionCount > 1 {
			fmt.Fprintf(os.Stderr, "The input has %d solutions: %s\n", solutionCount, input)
		}

		autoStore(solverStore, *problem, "input")
		playCli(*problem, solverStore, solverStore.GetAllStrategySolverKeys())
	} else {
		difficulty := parseDifficulty(level)
		levelName := level
		fmt.Printf("Generating a random %s Sudoku problem...\n", capitalize(levelName))

		problem, keys := generateWithFallback(solverStore, difficulty, levelName)
		playCli(problem, solverStore, keys)
	}
}

func parseDifficulty(level string) generator.Difficulty {
	switch level {
	case "easy":
		return generator.NewEasyDifficulty()
	case "medium":
		return generator.NewMediumDifficulty()
	case "hard":
		return generator.NewHardDifficulty()
	case "expert":
		return generator.NewExpertDifficulty()
	case "evil":
		return generator.NewEvilDifficulty()
	default:
		fmt.Fprintf(os.Stderr, "Invalid difficulty level: %s. Options: easy, medium, hard, expert, evil\n", level)
		os.Exit(1)
		return generator.Difficulty{} // unreachable
	}
}

func generateWithFallback(solverStore solver.Store, difficulty generator.Difficulty, levelName string) (core.Board, []string) {
	opts := generator.NewBestEffortOptions(solverStore, difficulty)
	result := generator.GenerateBestEffort(opts)

	autoStore(solverStore, result.Puzzle, "generated")

	if result.Matched {
		keys := difficulty.AllowedSolverKeys()
		if len(keys) == 0 {
			keys = solverStore.GetAllStrategySolverKeys()
		}
		return result.Puzzle, keys
	}

	dbPath := defaultDBPath()
	puzzleDB, err := db.Open(dbPath)
	if err == nil {
		defer puzzleDB.Close()

		if dbPuzzle, err := puzzleDB.GetRandom(levelName); err == nil && dbPuzzle != nil {
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

func autoStore(solverStore solver.Store, board core.Board, source string) {
	solvedBoard := board.Copy()
	solverStore.GetDefaultSolver().Solve(&solvedBoard)
	if !solvedBoard.IsSolved() {
		return
	}

	normalizedSolved := solvedBoard.Copy()
	normalizedSolved.Normalize()

	var digitMap [10]int
	for col := 0; col < 9; col++ {
		original := solvedBoard.Get(core.NewPosition(0, col))
		normalized := normalizedSolved.Get(core.NewPosition(0, col))
		digitMap[original] = normalized
	}

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
	classification := solver.ClassifyPuzzle(solverStore, board)

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

func playCli(problem core.Board, solverStore solver.Store, strategySolverKeys []string) {
	opts := game.NewDefaultOptions(solverStore)
	opts.StrategySolverKeys = strategySolverKeys
	newGame := game.NewGame(problem, opts)
	ctrl := cli.NewController(&newGame)
	ctrl.Play()
}

func defaultDBPath() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "puzzles.db"
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "sudoku", "puzzles.db")
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

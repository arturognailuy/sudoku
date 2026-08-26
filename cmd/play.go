package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gnailuy/sudoku/cli"
	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/db"
	"github.com/gnailuy/sudoku/generator"
	"github.com/gnailuy/sudoku/solver"
	"github.com/spf13/cobra"
)

func runPlay(cmd *cobra.Command) error {
	input, _ := cmd.Flags().GetString("input")
	level, _ := cmd.Flags().GetString("level")
	resume, _ := cmd.Flags().GetString("resume")
	newGame, _, err := createSession(sessionRequest{input: input, level: level, resume: resume}, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	ctrl := cli.NewController(&newGame)
	ctrl.Play()
	return nil
}

func parseDifficulty(level string) generator.Difficulty {
	difficulty, err := difficultyForLevel(level)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return difficulty
}

func generateWithFallbackTo(output io.Writer, solverStore solver.Store, difficulty generator.Difficulty, levelName string) (core.Board, []string, error) {
	opts := generator.NewBestEffortOptions(solverStore, difficulty)
	result := generator.GenerateBestEffort(opts)

	if result.RoundsUsed > 0 {
		autoStore(solverStore, result.Puzzle, "generated")
	}

	if result.Matched {
		keys := difficulty.AllowedSolverKeys()
		if len(keys) == 0 {
			keys = solverStore.GetAllStrategySolverKeys()
		}
		return result.Puzzle, keys, nil
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
			return board, keys, nil
		}
	}

	if result.RoundsUsed == 0 {
		return core.Board{}, nil, fmt.Errorf(
			"generation timed out before a puzzle completed and no %s puzzle is available in the database",
			levelName,
		)
	}

	actualLevel := result.Classification.Difficulty
	if actualLevel != levelName {
		fmt.Fprintf(output, "Requested difficulty: %s. Generated puzzle difficulty: %s. Enjoy!\n",
			capitalize(levelName), capitalize(actualLevel))
	}

	keys := difficulty.AllowedSolverKeys()
	if len(keys) == 0 {
		keys = solverStore.GetAllStrategySolverKeys()
	}
	return result.Puzzle, keys, nil
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

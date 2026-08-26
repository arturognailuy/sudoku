package cmd

import (
	"fmt"
	"io"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/game"
	"github.com/gnailuy/sudoku/generator"
	"github.com/gnailuy/sudoku/sessionfile"
)

type sessionRequest struct {
	input  string
	level  string
	resume string
}

func createSession(request sessionRequest, output, errorOutput io.Writer) (game.Game, string, error) {
	options := game.NewDefaultOptions(solverStore)
	options.StrategySolverKeys = solverStore.GetAllStrategySolverKeys()
	if request.resume != "" {
		data, err := sessionfile.Read(request.resume)
		if err != nil {
			return game.Game{}, "", fmt.Errorf("unable to read saved session: %w", err)
		}
		restored, err := game.Restore(data, options)
		if err != nil {
			return game.Game{}, "", fmt.Errorf("unable to resume saved session: %w", err)
		}
		return restored, request.resume, nil
	}

	var problem core.Board
	keys := solverStore.GetAllStrategySolverKeys()
	if request.input != "" {
		parsed, err := generator.GenerateSudokuProblemFromString(request.input)
		if err != nil {
			return game.Game{}, "", fmt.Errorf("the input is not a valid Sudoku problem: %s", request.input)
		}
		count := solverStore.GetDefaultSolver().CountSolutions(parsed)
		if count == 0 {
			return game.Game{}, "", fmt.Errorf("the input is not a solvable Sudoku problem: %s", request.input)
		}
		if count > 1 {
			fmt.Fprintf(errorOutput, "The input has %d solutions: %s\n", count, request.input)
		}
		problem = *parsed
		autoStore(solverStore, problem, "input")
	} else {
		difficulty, err := difficultyForLevel(request.level)
		if err != nil {
			return game.Game{}, "", err
		}
		fmt.Fprintf(output, "Generating a random %s Sudoku problem...\n", capitalize(request.level))
		problem, keys, err = generateWithFallbackTo(output, solverStore, difficulty, request.level)
		if err != nil {
			return game.Game{}, "", err
		}
	}
	options.StrategySolverKeys = keys
	return game.NewGame(problem, options), "", nil
}

func difficultyForLevel(level string) (generator.Difficulty, error) {
	switch level {
	case "easy":
		return generator.NewEasyDifficulty(), nil
	case "medium":
		return generator.NewMediumDifficulty(), nil
	case "hard":
		return generator.NewHardDifficulty(), nil
	case "expert":
		return generator.NewExpertDifficulty(), nil
	case "evil":
		return generator.NewEvilDifficulty(), nil
	default:
		return generator.Difficulty{}, fmt.Errorf("invalid difficulty level: %s. Options: easy, medium, hard, expert, evil", level)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/gnailuy/sudoku/cli"
	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/game"
	"github.com/gnailuy/sudoku/generator"
	"github.com/gnailuy/sudoku/solver"
)

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

		playCli(*problem, solverStore, solverStore.GetAllStrategySolverKeys())
	} else {
		// Generate a random problem.
		fmt.Printf("Generating a random %s Sudoku problem...\n", options.Level.String())
		difficulty := options.GetDifficultyOptions()
		problem := generator.GenerateSudokuProblem(generator.NewProblemOptions(solverStore, difficulty))

		// Use the difficulty's strategy solver keys for hints, falling back
		// to all registered solvers when the difficulty has no keys set
		// (e.g., Extreme/Evil levels).
		keys := difficulty.StrategySolverKeys
		if len(keys) == 0 {
			keys = solverStore.GetAllStrategySolverKeys()
		}

		playCli(problem, solverStore, keys)
	}
}

// Function to play a game in CLI.
func playCli(problem core.Board, solverStore solver.Store, strategySolverKeys []string) {
	opts := game.NewDefaultOptions(solverStore)
	opts.StrategySolverKeys = strategySolverKeys
	newGame := game.NewGame(problem, opts)
	ctrl := cli.NewController(&newGame)
	ctrl.Play()
}

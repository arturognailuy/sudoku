// Package cmd implements the CLI commands using cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/gnailuy/sudoku/solver"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sudoku",
	Short: "A CLI Sudoku game with strategy-based difficulty",
	Long: `Sudoku is a CLI puzzle game featuring 23 strategy solvers across five
difficulty tiers (Easy through Evil). Play interactively, generate batches
of puzzles, or import puzzle collections.`,
	Run: func(cmd *cobra.Command, args []string) {
		runPlay(cmd)
	},
}

// solverStore is shared across all subcommands.
var solverStore solver.Store

func init() {
	solverStore = solver.NewStore()

	// Play mode flags (on root command).
	rootCmd.Flags().StringP("input", "i", "", "Specify a Sudoku problem string to play")
	rootCmd.Flags().StringP("level", "l", "hard", "Difficulty level: easy, medium, hard, expert, evil")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

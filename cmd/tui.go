package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	front "github.com/gnailuy/sudoku/tui"
	"github.com/spf13/cobra"
)

func newTUICommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "tui",
		Short: "Play in the full-screen Bubble Tea interface",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			input, _ := command.Flags().GetString("input")
			level, _ := command.Flags().GetString("level")
			resume, _ := command.Flags().GetString("resume")
			current, resumePath, err := createSession(sessionRequest{input: input, level: level, resume: resume}, command.OutOrStdout(), command.ErrOrStderr())
			if err != nil {
				return err
			}
			program := tea.NewProgram(front.NewModel(current, resumePath), tea.WithAltScreen(), tea.WithInput(os.Stdin), tea.WithOutput(os.Stdout))
			_, err = program.Run()
			return err
		},
	}
	command.Flags().StringP("input", "i", "", "Specify a Sudoku problem string to play")
	command.Flags().StringP("level", "l", "hard", "Difficulty level: easy, medium, hard, expert, evil")
	command.Flags().String("resume", "", "Resume a saved game session")
	command.MarkFlagsMutuallyExclusive("resume", "input")
	command.MarkFlagsMutuallyExclusive("resume", "level")
	return command
}

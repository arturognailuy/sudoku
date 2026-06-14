package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/game"
)

// Controller owns all terminal I/O for the Sudoku game.
// It holds a Game and translates user commands into Game API calls.
type Controller struct {
	game         *game.Game
	closeChannel CloseChannel
}

// NewController creates a CLI controller for the given game.
func NewController(g *game.Game) *Controller {
	return &Controller{
		game:         g,
		closeChannel: NewCloseChannel(),
	}
}

// printError prints an error message with a prefix [ERROR].
func printError(message ...any) {
	fmt.Fprintln(os.Stderr, "[ERROR]", message)
}

// printColumnNumbers prints the column number header/footer.
func printColumnNumbers() {
	fmt.Print("    ")
	for i := 0; i < 9; i++ {
		if i%3 == 0 && i != 0 {
			fmt.Print("  ")
		}
		fmt.Printf(" %d", i+1)
	}
	fmt.Println()
}

// PrintBoard renders the 9×9 Sudoku grid with row/column numbers.
func (ctrl *Controller) PrintBoard() {
	fmt.Println()
	printColumnNumbers()

	for i := 0; i < 9; i++ {
		if i%3 == 0 {
			fmt.Println("    -------+-------+-------")
		}

		fmt.Printf(" %d ", i+1)
		for j := 0; j < 9; j++ {
			position := core.NewPosition(i, j)
			value := ctrl.game.Get(position)

			if j%3 == 0 {
				fmt.Print("| ")
			}
			if value == 0 {
				fmt.Print(". ")
			} else {
				fmt.Printf("%d ", value)
			}
		}
		fmt.Println("|", i+1)
	}
	fmt.Println("    -------+-------+-------")

	printColumnNumbers()
	fmt.Println()
}

// PrintHelp displays the available commands.
func (ctrl *Controller) PrintHelp() {
	fmt.Println("Supported commands:")
	fmt.Println("  - help, h                       : Print this help message.")
	fmt.Println("  - add, a <row> <column> <value> : Add the value to the cell at (row, column).")
	fmt.Println("  - clear, d <row> <column>       : Clear the value in a cell at (row, column).")
	fmt.Println("  - check, c                      : Check if the current board is correct.")
	fmt.Println("  - undo, u                       : Undo last move.")
	fmt.Println("  - redo, r                       : Redo last undo.")
	fmt.Println("  - repair, f                     : Undo all invalid inputs.")
	fmt.Println("  - hint, i                       : Apply a hint for the next move.")
	fmt.Println("  - solve, s                      : Solve the problem for me.")
	fmt.Println("  - reset, e                      : Reset the game and start over.")
	fmt.Println("  - quit, q                       : Quit the game.")
}

// setValue applies a cell value through the game API.
func (ctrl *Controller) setValue(rowInput, columnInput, valueInput int) (success bool, err error) {
	positionPointer, err := core.NewPositionFromInput(rowInput, columnInput)
	if err != nil {
		return false, fmt.Errorf("error in the input position: %w", err)
	}

	cellPointer, err := core.NewCellFromInput(*positionPointer, valueInput)
	if err != nil {
		return false, fmt.Errorf("error in the input value: %w", err)
	}

	if ctrl.game.Get(*positionPointer) == valueInput {
		return false, nil
	}

	err = ctrl.game.AddInputAndRecordHistory(*cellPointer)
	success = err == nil
	return
}

// runAddCommand handles the add command.
func (ctrl *Controller) runAddCommand(commandArguments string) (added bool, err error) {
	var row, column, value int
	_, err = fmt.Sscanf(commandArguments, "%1d%1d%1d", &row, &column, &value)
	if err != nil {
		return false, err
	}
	added, err = ctrl.setValue(row, column, value)
	return
}

// runClearCommand handles the clear command.
func (ctrl *Controller) runClearCommand(commandArguments string) (cleared bool, err error) {
	var row, column int
	_, err = fmt.Sscanf(commandArguments, "%1d%1d", &row, &column)
	if err != nil {
		return false, err
	}
	cleared, err = ctrl.setValue(row, column, 0)
	return
}

// runCommandWithArguments handles commands that take arguments (add/clear).
func (ctrl *Controller) runCommandWithArguments(commandFields []string) (success bool, err error) {
	if len(commandFields) != 2 {
		return false, errors.New("no argument specified for the command")
	}

	switch commandFields[0] {
	case "add", "a":
		return ctrl.runAddCommand(commandFields[1])
	case "clear", "d":
		return ctrl.runClearCommand(commandFields[1])
	default:
		return false, fmt.Errorf("unsupported command: %s", commandFields[0])
	}
}

// RunCommand parses and dispatches a single command.
func (ctrl *Controller) RunCommand(command string) bool {
	commandFields := strings.SplitN(command, " ", 2)

	if len(commandFields) == 0 || len(commandFields[0]) == 0 {
		return false
	}

	switch commandFields[0] {
	case "help", "h":
		ctrl.PrintHelp()
		return false
	case "add", "a", "clear", "d":
		success, err := ctrl.runCommandWithArguments(commandFields)
		if err != nil {
			printError("Failed to run the", commandFields[0], "command:", err)
		}
		return success
	case "check", "c":
		if ctrl.game.IsValid() {
			fmt.Println("The current board is correct.")
		} else {
			fmt.Println("You have entered incorrect values(s).")
		}
	case "undo", "u":
		err := ctrl.game.Undo()
		return err == nil
	case "redo", "r":
		err := ctrl.game.Redo()
		return err == nil
	case "repair", "f":
		return ctrl.game.Repair() > 0
	case "hint", "i":
		hint := ctrl.game.Hint()
		if hint != nil {
			added, err := ctrl.setValue(hint.Cell.Position.Row+1, hint.Cell.Position.Column+1, hint.Cell.Value)
			if err != nil {
				printError("Failed to apply hint:", err)
			}
			if added {
				fmt.Printf("Hint: %s\n", hint.Reason)
			}
			return added
		}
		return false
	case "solve", "s":
		ctrl.game.Solve()
		return true
	case "reset", "e":
		ctrl.game.Reset()
		return true
	case "quit", "q":
		ctrl.closeChannel.Close()
	default:
		// Shorthand: bare digits are treated as an add command.
		added, err := ctrl.runAddCommand(command)
		if err != nil {
			printError("Failed to run the command:", err)
		}
		return added
	}

	return false
}

// askUserInput prints the board, prompts, and reads one line of input.
func (ctrl *Controller) askUserInput(scanner *bufio.Scanner, inputChannel chan string) {
	if ctrl.closeChannel.IsClosed() {
		return
	}

	ctrl.PrintBoard()

	fmt.Println("Enter a command (Enter 'help' or 'h for help):")
	fmt.Print("> ")

	if scanner.Scan() {
		inputChannel <- strings.TrimSpace(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		printError("Failed to read the input command:", err)
	}
}

// Play runs the main interactive game loop.
func (ctrl *Controller) Play() {
	inputChannel := make(chan string)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		go ctrl.askUserInput(scanner, inputChannel)

		select {
		case command := <-inputChannel:
			ctrl.RunCommand(command)
		case <-ctrl.closeChannel:
			fmt.Println("\nExiting the game.")
			fmt.Println(ctrl.game.ToString())
			os.Exit(0)
		}

		if ctrl.game.IsSolved() {
			ctrl.PrintBoard()
			break
		}
	}

	fmt.Println("Congratulations! You have solved the problem.")
	fmt.Println(ctrl.game.ToString())
}

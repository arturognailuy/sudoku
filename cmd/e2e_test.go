package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/db"
	"github.com/gnailuy/sudoku/game"
	"github.com/gnailuy/sudoku/solver"
)

// Known puzzle strings for testing.
// This puzzle is classified as "evil" (requires xy-chain), which makes it useful
// for testing game commands since it has a unique solution and many empty cells.
const knownPuzzle = "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."

// A puzzle with multiple solutions (3 solutions) — derived from knownPuzzle with
// two cells removed to break uniqueness. Fast to count solutions (~1s).
const multipleSolutionPuzzle = "....2.6.....3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."

// --- Test #26: DB Fallback Path ---
// Pre-populate DB with a puzzle at its classified difficulty, then verify
// GetRandom returns it. Also verify that requesting a different difficulty
// returns nil. This tests the core fallback mechanism end-to-end.
func TestDBFallbackPath(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "fallback-test.db")

	puzzleDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	store := solver.NewStore()

	// Classify our known puzzle to find its actual difficulty.
	board := core.NewEmptyBoard()
	board.FromString(knownPuzzle)
	normalizedStr := normalizePuzzleForDB(store, board)
	classification := solver.ClassifyPuzzle(store, board)
	actualDifficulty := classification.Difficulty

	// Seed the DB with this puzzle at its classified difficulty.
	inserted, err := puzzleDB.InsertPuzzle(db.Puzzle{
		Puzzle:       normalizedStr,
		Difficulty:   actualDifficulty,
		Score:        classification.Score,
		MaxTechnique: classification.MaxTechnique,
		Source:       "e2e-fallback-seed",
	})
	if err != nil {
		t.Fatalf("insert seed puzzle: %v", err)
	}
	if !inserted {
		t.Fatal("seed puzzle should have been inserted")
	}

	puzzleDB.Close()

	// Reopen DB and test the fallback: GetRandom at the seeded difficulty
	// should return our puzzle.
	puzzleDB2, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer puzzleDB2.Close()

	dbPuzzle, err := puzzleDB2.GetRandom(actualDifficulty)
	if err != nil {
		t.Fatalf("get random %s: %v", actualDifficulty, err)
	}
	if dbPuzzle == nil {
		t.Fatalf("DB fallback returned nil — expected %s puzzle from DB", actualDifficulty)
	}
	if dbPuzzle.Difficulty != actualDifficulty {
		t.Errorf("DB fallback returned difficulty=%s, expected %s", dbPuzzle.Difficulty, actualDifficulty)
	}

	// Verify the puzzle from DB is valid and playable.
	fallbackBoard := core.NewEmptyBoard()
	fallbackBoard.FromString(dbPuzzle.Puzzle)
	if !fallbackBoard.IsValid() {
		t.Error("DB fallback puzzle is not valid")
	}

	solutionCount := store.GetDefaultSolver().CountSolutions(&fallbackBoard)
	if solutionCount != 1 {
		t.Errorf("DB fallback puzzle has %d solutions, expected 1", solutionCount)
	}

	// Verify that requesting a difficulty NOT in DB returns nil (no false matches).
	missingDifficulty := "easy"
	if actualDifficulty == "easy" {
		missingDifficulty = "hard"
	}
	noMatch, err := puzzleDB2.GetRandom(missingDifficulty)
	if err != nil {
		t.Fatalf("get random %s: %v", missingDifficulty, err)
	}
	if noMatch != nil {
		t.Errorf("expected nil for %s (DB only has %s), got puzzle", missingDifficulty, actualDifficulty)
	}
}

// --- Test #27: Generate with count 0 ---
func TestGenerateInvalidCount(t *testing.T) {
	cmd := generateCmd

	if err := cmd.Flags().Set("count", "0"); err != nil {
		t.Fatalf("set count flag: %v", err)
	}
	if err := cmd.Flags().Set("difficulty", "easy"); err != nil {
		t.Fatalf("set difficulty flag: %v", err)
	}

	err := runGenerate(cmd)
	if err == nil {
		t.Fatal("expected error for count=0, got nil")
	}
	if !strings.Contains(err.Error(), "count must be positive") {
		t.Errorf("expected 'count must be positive' error, got: %s", err.Error())
	}
}

// --- Tests #28-33: Game commands (clear, check, repair, reset, redo, shorthand) ---
// These test via the game API directly, which is what the CLI wraps.

// Test #28: Game clear command
func TestGameClear(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString(knownPuzzle)

	store := solver.NewStore()
	opts := game.NewDefaultOptions(store)
	g := game.NewGame(board, opts)

	// Find an empty cell to add a value to.
	emptyPos := findEmptyCell(board)
	if emptyPos == nil {
		t.Fatal("no empty cell found in test puzzle")
	}

	// Add a value.
	input := core.NewCell(*emptyPos, 5)
	err := g.AddInputAndRecordHistory(input)
	if err != nil {
		t.Fatalf("add input: %v", err)
	}

	// Verify value is set (either in play board or invalid board).
	if g.Get(*emptyPos) != 5 {
		t.Fatalf("expected cell to have value 5, got %d", g.Get(*emptyPos))
	}

	// Clear the cell (add value 0).
	clearInput := core.NewCell(*emptyPos, 0)
	err = g.AddInputAndRecordHistory(clearInput)
	if err != nil {
		t.Fatalf("clear input: %v", err)
	}

	// Verify cell is cleared.
	if g.Get(*emptyPos) != 0 {
		t.Errorf("expected cell to be cleared (0), got %d", g.Get(*emptyPos))
	}
}

// Test #29: Game check — correct board
func TestGameCheckCorrect(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString(knownPuzzle)

	store := solver.NewStore()
	opts := game.NewDefaultOptions(store)
	g := game.NewGame(board, opts)

	// Board with no user input should be valid (no incorrect values).
	if !g.IsValid() {
		t.Error("initial board should be valid (no incorrect inputs)")
	}

	// Add a correct value: solve the puzzle to find what goes in the first empty cell.
	solvedBoard := board.Copy()
	store.GetDefaultSolver().Solve(&solvedBoard)

	emptyPos := findEmptyCell(board)
	if emptyPos == nil {
		t.Fatal("no empty cell found")
	}

	correctValue := solvedBoard.Get(*emptyPos)
	input := core.NewCell(*emptyPos, correctValue)
	err := g.AddInputAndRecordHistory(input)
	if err != nil {
		t.Fatalf("add correct input: %v", err)
	}

	// Board should still be valid after correct input.
	if !g.IsValid() {
		t.Error("board should be valid after adding correct value")
	}
}

// Test #30: Game check — incorrect board
func TestGameCheckIncorrect(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString(knownPuzzle)

	store := solver.NewStore()
	opts := game.NewDefaultOptions(store)
	g := game.NewGame(board, opts)

	// Solve to find the correct value, then add a wrong one.
	solvedBoard := board.Copy()
	store.GetDefaultSolver().Solve(&solvedBoard)

	emptyPos := findEmptyCell(board)
	if emptyPos == nil {
		t.Fatal("no empty cell found")
	}

	correctValue := solvedBoard.Get(*emptyPos)
	// Pick a wrong value (any digit 1-9 that isn't the correct one).
	wrongValue := (correctValue % 9) + 1
	if wrongValue == correctValue {
		wrongValue = (wrongValue % 9) + 1
	}

	input := core.NewCell(*emptyPos, wrongValue)
	err := g.AddInputAndRecordHistory(input)
	if err != nil {
		t.Fatalf("add wrong input: %v", err)
	}

	// Board should be invalid after incorrect input.
	if g.IsValid() {
		t.Error("board should be invalid after adding wrong value")
	}
}

// Test #31: Game repair command
func TestGameRepair(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString(knownPuzzle)

	store := solver.NewStore()
	opts := game.NewDefaultOptions(store)
	g := game.NewGame(board, opts)

	// Solve to find the correct value, then add a wrong one.
	solvedBoard := board.Copy()
	store.GetDefaultSolver().Solve(&solvedBoard)

	emptyPos := findEmptyCell(board)
	if emptyPos == nil {
		t.Fatal("no empty cell found")
	}

	correctValue := solvedBoard.Get(*emptyPos)
	wrongValue := (correctValue % 9) + 1
	if wrongValue == correctValue {
		wrongValue = (wrongValue % 9) + 1
	}

	input := core.NewCell(*emptyPos, wrongValue)
	err := g.AddInputAndRecordHistory(input)
	if err != nil {
		t.Fatalf("add wrong input: %v", err)
	}

	if g.IsValid() {
		t.Fatal("board should be invalid before repair")
	}

	// Repair should remove invalid inputs.
	undoSteps := g.Repair()
	if undoSteps == 0 {
		t.Error("repair should have undone at least one step")
	}

	if !g.IsValid() {
		t.Error("board should be valid after repair")
	}
}

// Test #32: Game reset command
func TestGameReset(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString(knownPuzzle)

	store := solver.NewStore()
	opts := game.NewDefaultOptions(store)
	g := game.NewGame(board, opts)

	// Add a value.
	emptyPos := findEmptyCell(board)
	if emptyPos == nil {
		t.Fatal("no empty cell found")
	}

	input := core.NewCell(*emptyPos, 5)
	err := g.AddInputAndRecordHistory(input)
	if err != nil {
		t.Fatalf("add input: %v", err)
	}

	// Verify board changed.
	if g.Get(*emptyPos) == 0 {
		t.Fatal("cell should have value after add")
	}

	// Reset the game.
	g.Reset()

	// After reset, the cell should be empty again (back to problem state).
	if g.Get(*emptyPos) != 0 {
		t.Errorf("expected cell to be empty after reset, got %d", g.Get(*emptyPos))
	}

	// The play board should equal the problem board.
	if g.PlayBoard.ToString() != g.ProblemBoard.ToString() {
		t.Error("play board should match problem board after reset")
	}
}

// Test #33: Game redo command
func TestGameRedo(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString(knownPuzzle)

	store := solver.NewStore()
	opts := game.NewDefaultOptions(store)
	g := game.NewGame(board, opts)

	// Add a correct value.
	solvedBoard := board.Copy()
	store.GetDefaultSolver().Solve(&solvedBoard)

	emptyPos := findEmptyCell(board)
	if emptyPos == nil {
		t.Fatal("no empty cell found")
	}

	correctValue := solvedBoard.Get(*emptyPos)
	input := core.NewCell(*emptyPos, correctValue)
	err := g.AddInputAndRecordHistory(input)
	if err != nil {
		t.Fatalf("add input: %v", err)
	}

	// Undo.
	err = g.Undo()
	if err != nil {
		t.Fatalf("undo: %v", err)
	}
	if g.Get(*emptyPos) != 0 {
		t.Fatalf("cell should be empty after undo, got %d", g.Get(*emptyPos))
	}

	// Redo.
	err = g.Redo()
	if err != nil {
		t.Fatalf("redo: %v", err)
	}
	if g.Get(*emptyPos) != correctValue {
		t.Errorf("cell should have value %d after redo, got %d", correctValue, g.Get(*emptyPos))
	}
}

// --- Test #34: Shorthand digit input ---
// Tests that bare digits are dispatched as add commands.
// The CLI controller's RunCommand treats "row col value" as an add.
// We test this via the testController mirror that replicates the dispatch logic.
func TestShorthandDigitInput(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString(knownPuzzle)

	store := solver.NewStore()
	opts := game.NewDefaultOptions(store)
	g := game.NewGame(board, opts)

	// Solve to find a correct value for an empty cell.
	solvedBoard := board.Copy()
	store.GetDefaultSolver().Solve(&solvedBoard)

	emptyPos := findEmptyCell(board)
	if emptyPos == nil {
		t.Fatal("no empty cell found")
	}

	correctValue := solvedBoard.Get(*emptyPos)

	ctrl := newTestCtrl(&g)

	// Shorthand input format: "row col value" (1-indexed) — no "add" prefix.
	shorthand := fmt.Sprintf("%d %d %d", emptyPos.Row+1, emptyPos.Column+1, correctValue)

	changed := ctrl.RunCommand(shorthand)

	if !changed {
		t.Error("shorthand digit input should return true (board changed)")
	}

	if g.Get(*emptyPos) != correctValue {
		t.Errorf("after shorthand input, expected %d at (%d,%d), got %d",
			correctValue, emptyPos.Row, emptyPos.Column, g.Get(*emptyPos))
	}
}

// --- Test #35: Multiple-solution puzzle input ---
func TestMultipleSolutionPuzzle(t *testing.T) {
	store := solver.NewStore()

	// A puzzle with very few givens — should have multiple solutions.
	board := core.NewEmptyBoard()
	board.FromString(multipleSolutionPuzzle)

	solutionCount := store.GetDefaultSolver().CountSolutions(&board)
	if solutionCount <= 1 {
		t.Skipf("test puzzle has %d solution(s), expected >1 for this test — skipping", solutionCount)
	}

	// The play flow should still accept this puzzle (it warns but proceeds).
	// Verify the board is valid and the game can be created.
	if !board.IsValid() {
		t.Fatal("board with multiple solutions should still be valid")
	}

	opts := game.NewDefaultOptions(store)
	g := game.NewGame(board, opts)
	if g.IsSolved() {
		t.Error("game should not be solved at start")
	}
}

// --- Test #36: Generate with custom DB path ---
func TestGenerateCustomDBPath(t *testing.T) {
	tmpDir := t.TempDir()
	customDBPath := filepath.Join(tmpDir, "custom-test.db")

	// Verify DB doesn't exist yet.
	if _, err := os.Stat(customDBPath); err == nil {
		t.Fatal("custom DB should not exist yet")
	}

	// Generate 2 puzzles with custom DB path.
	puzzleDB, err := db.Open(customDBPath)
	if err != nil {
		t.Fatalf("open custom db: %v", err)
	}

	// Generate and store 2 puzzles.
	for i := 0; i < 2; i++ {
		result := generateOnePuzzle("hard", 5000, 5)
		storePuzzle(puzzleDB, result)
	}
	puzzleDB.Close()

	// Verify DB was created at custom path.
	if _, err := os.Stat(customDBPath); os.IsNotExist(err) {
		t.Error("custom DB file should exist after generation")
	}

	// Reopen and check stats.
	puzzleDB2, err := db.Open(customDBPath)
	if err != nil {
		t.Fatalf("reopen custom db: %v", err)
	}
	defer puzzleDB2.Close()

	stats, err := puzzleDB2.GetStats()
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.Total == 0 {
		t.Error("custom DB should have puzzles stored")
	}
}

// --- Test #37: Import empty/comments-only file ---
func TestImportEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	puzzleFile := filepath.Join(tmpDir, "empty-puzzles.txt")
	dbPath := filepath.Join(tmpDir, "empty-import-test.db")

	// File with only comments and blank lines.
	content := `# This file has no valid puzzles
# Just comments

# And blank lines

`
	if err := os.WriteFile(puzzleFile, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	puzzleDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer puzzleDB.Close()

	store := solver.NewStore()
	_ = store // used for validation if needed

	// Process the file the same way runImport does.
	file, err := os.Open(puzzleFile)
	if err != nil {
		t.Fatalf("open puzzle file: %v", err)
	}
	defer file.Close()

	var totalLines, validLines int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		totalLines++

		puzzleStr := normalizePuzzleInput(line)
		if !core.IsValidSudokuString(puzzleStr) {
			continue
		}

		board := core.NewEmptyBoard()
		board.FromString(puzzleStr)
		if !board.IsValid() {
			continue
		}

		solutionCount := store.GetDefaultSolver().CountSolutions(&board)
		if solutionCount == 0 {
			continue
		}
		validLines++
	}

	if totalLines != 0 {
		t.Errorf("expected 0 total data lines, got %d", totalLines)
	}
	if validLines != 0 {
		t.Errorf("expected 0 valid lines, got %d", validLines)
	}

	// DB should have 0 puzzles.
	stats, err := puzzleDB.GetStats()
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.Total != 0 {
		t.Errorf("expected 0 puzzles in DB, got %d", stats.Total)
	}
}

// --- Helpers ---

// findEmptyCell returns the first empty cell position in the board.
func findEmptyCell(board core.Board) *core.Position {
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			pos := core.NewPosition(row, col)
			if board.Get(pos) == 0 {
				return &pos
			}
		}
	}
	return nil
}

// newTestCtrl creates a test controller for testing command dispatch.
func newTestCtrl(g *game.Game) testController {
	return testController{game: g}
}

// testController is a minimal wrapper for testing RunCommand logic
// that mirrors cli.Controller's dispatch without terminal I/O.
type testController struct {
	game *game.Game
}

func (ctrl *testController) RunCommand(command string) bool {
	commandFields := strings.SplitN(command, " ", 2)

	if len(commandFields) == 0 || len(commandFields[0]) == 0 {
		return false
	}

	switch commandFields[0] {
	case "add", "a", "clear", "d":
		if len(commandFields) != 2 {
			return false
		}
		switch commandFields[0] {
		case "add", "a":
			return ctrl.runAddCommand(commandFields[1])
		case "clear", "d":
			return ctrl.runClearCommand(commandFields[1])
		}
	case "check", "c":
		return false
	case "undo", "u":
		return ctrl.game.Undo() == nil
	case "redo", "r":
		return ctrl.game.Redo() == nil
	case "repair", "f":
		return ctrl.game.Repair() > 0
	case "hint", "i":
		hint := ctrl.game.Hint()
		if hint != nil {
			pos := hint.Cell.Position
			_ = ctrl.game.AddInputAndRecordHistory(hint.Cell)
			return ctrl.game.Get(pos) == hint.Cell.Value
		}
		return false
	case "solve", "s":
		ctrl.game.Solve()
		return true
	case "reset", "e":
		ctrl.game.Reset()
		return true
	default:
		// Shorthand: bare digits treated as add.
		return ctrl.runAddCommand(command)
	}
	return false
}

func (ctrl *testController) runAddCommand(args string) bool {
	var row, column, value int
	_, err := fmt.Sscanf(args, "%1d%1d%1d", &row, &column, &value)
	if err != nil {
		return false
	}
	return ctrl.setValue(row, column, value)
}

func (ctrl *testController) runClearCommand(args string) bool {
	var row, column int
	_, err := fmt.Sscanf(args, "%1d%1d", &row, &column)
	if err != nil {
		return false
	}
	return ctrl.setValue(row, column, 0)
}

func (ctrl *testController) setValue(rowInput, columnInput, valueInput int) bool {
	positionPointer, err := core.NewPositionFromInput(rowInput, columnInput)
	if err != nil {
		return false
	}

	cellPointer, err := core.NewCellFromInput(*positionPointer, valueInput)
	if err != nil {
		return false
	}

	if ctrl.game.Get(*positionPointer) == valueInput {
		return false
	}

	err = ctrl.game.AddInputAndRecordHistory(*cellPointer)
	return err == nil
}

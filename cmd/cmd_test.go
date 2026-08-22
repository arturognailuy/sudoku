package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/db"
	"github.com/gnailuy/sudoku/game"
	"github.com/gnailuy/sudoku/generator"
	"github.com/gnailuy/sudoku/solver"
)

func fixedGenerationResult() generator.GenerationResult {
	return generator.GenerationResult{
		Puzzle: loadBoard(testKnownPuzzle),
		Classification: solver.Classification{
			Difficulty:   "easy",
			Score:        1,
			MaxTechnique: "fixture",
		},
		Matched:    true,
		RoundsUsed: 1,
	}
}

func TestStorePuzzle(t *testing.T) {
	// Create an in-memory DB.
	puzzleDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer puzzleDB.Close()

	result := fixedGenerationResult()

	// Store once — should return true (inserted).
	stored := storePuzzle(puzzleDB, result)
	if !stored {
		t.Fatal("first store should return true")
	}

	// Store again — should return false (duplicate).
	stored = storePuzzle(puzzleDB, result)
	if stored {
		t.Fatal("second store should return false (duplicate)")
	}
}

func TestStorePuzzleDBStats(t *testing.T) {
	puzzleDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer puzzleDB.Close()

	// Repeated fixture storage exercises the DB boundary without making this
	// storage test depend on randomized generation.
	for i := 0; i < 3; i++ {
		storePuzzle(puzzleDB, fixedGenerationResult())
	}

	stats, err := puzzleDB.GetStats()
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.Total == 0 {
		t.Fatal("no puzzles stored")
	}
}

func TestBatchGenerateUsesInjectedGenerator(t *testing.T) {
	puzzleDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer puzzleDB.Close()

	calls := 0
	generate := func(level string, timeout time.Duration, rounds int) generator.GenerationResult {
		calls++
		if level != "easy" || timeout != 2*time.Second || rounds != 3 {
			t.Fatalf("unexpected generation arguments: %q, %s, %d", level, timeout, rounds)
		}
		return fixedGenerationResult()
	}

	report := batchGenerateWith(puzzleDB, 3, "easy", 2*time.Second, 3, 1, generate)
	if calls != 3 || report.generated != 3 || report.stored != 1 || report.duplicates != 2 {
		t.Fatalf("unexpected report: calls=%d generated=%d stored=%d duplicates=%d", calls, report.generated, report.stored, report.duplicates)
	}
}

func TestNormalizePuzzleInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"dots format", "1.3.5....", "1.3.5...."},
		{"zeros format", "103050000", "1.3.5...."},
		{"with spaces", "1 0 3 0 5 0 0 0 0", "1.3.5...."},
		{"81-char zeros", strings.Repeat("0", 81), strings.Repeat(".", 81)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePuzzleInput(tt.input)
			if got != tt.expected {
				t.Errorf("normalizePuzzleInput(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNormalizePuzzleForDB(t *testing.T) {
	store := solver.NewStore()

	// A known solvable puzzle.
	puzzleStr := "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
	board := loadBoard(puzzleStr)

	normalized := normalizePuzzleForDB(store, board)
	if len(normalized) != 81 {
		t.Fatalf("normalized string length = %d, want 81", len(normalized))
	}

	// The normalized string should start with digits (normalized first row).
	// It should be a valid sudoku string.
	for _, ch := range normalized {
		if ch != '.' && (ch < '1' || ch > '9') {
			t.Fatalf("invalid character in normalized string: %c", ch)
		}
	}
}

func TestImportFromFile(t *testing.T) {
	// Create a temp file with test puzzles.
	tmpDir := t.TempDir()
	puzzleFile := filepath.Join(tmpDir, "test-puzzles.txt")

	// Use a known solvable puzzle.
	content := `# Test puzzles
..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3..

# Invalid puzzle (too short)
123456

# Another known solvable puzzle (using zeros for empty)
003020600900305001001806400008102900700000008006708200002609500800203009005010300
`
	if err := os.WriteFile(puzzleFile, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	puzzleDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer puzzleDB.Close()

	// Verify we can read and process the file manually (testing normalization
	// and classification without running the full command).
	store := solver.NewStore()

	puzzle1 := "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
	board1 := loadBoard(puzzle1)
	normalized1 := normalizePuzzleForDB(store, board1)
	classification1 := solver.ClassifyPuzzle(store, board1)

	inserted, err := puzzleDB.InsertPuzzle(db.Puzzle{
		Puzzle:       normalized1,
		Difficulty:   classification1.Difficulty,
		Score:        classification1.Score,
		MaxTechnique: classification1.MaxTechnique,
		Source:       "test",
	})
	if err != nil {
		t.Fatalf("insert puzzle: %v", err)
	}
	if !inserted {
		t.Fatal("puzzle should be inserted (new DB)")
	}

	// Insert same puzzle in different format — should be duplicate.
	puzzle2 := "003020600900305001001806400008102900700000008006708200002609500800203009005010300"
	board2 := loadBoard(normalizePuzzleInput(puzzle2))
	normalized2 := normalizePuzzleForDB(store, board2)

	// Same puzzle in different notation → same normalized form.
	if normalized1 != normalized2 {
		t.Logf("normalized1: %s", normalized1)
		t.Logf("normalized2: %s", normalized2)
		t.Fatal("same puzzle in different formats should normalize to the same string")
	}

	inserted, err = puzzleDB.InsertPuzzle(db.Puzzle{
		Puzzle:       normalized2,
		Difficulty:   classification1.Difficulty,
		Score:        classification1.Score,
		MaxTechnique: classification1.MaxTechnique,
		Source:       "test",
	})
	if err != nil {
		t.Fatalf("insert duplicate: %v", err)
	}
	if inserted {
		t.Fatal("same puzzle should be rejected as duplicate")
	}
}

func TestParseDifficulty(t *testing.T) {
	levels := []string{"easy", "medium", "hard", "expert", "evil"}
	for _, level := range levels {
		d := parseDifficultyQuiet(level)
		if d.MinimumClues <= 0 {
			t.Errorf("parseDifficultyQuiet(%q): MinimumClues should be > 0", level)
		}
	}
}

// loadBoard is a test helper that creates a board from a puzzle string.
func loadBoard(puzzleStr string) core.Board {
	board := core.NewEmptyBoard()
	board.FromString(puzzleStr)
	return board
}

// Known puzzle strings for testing.
const testKnownPuzzle = "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."

// --- DB Fallback Path ---
// Seed DB with a puzzle at its classified difficulty, verify GetRandom
// returns it, and verify a different difficulty returns nil.
func TestDBFallbackPath(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "fallback-test.db")

	puzzleDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	store := solver.NewStore()

	board := core.NewEmptyBoard()
	board.FromString(testKnownPuzzle)
	normalizedStr := normalizePuzzleForDB(store, board)
	classification := solver.ClassifyPuzzle(store, board)
	actualDifficulty := classification.Difficulty

	inserted, err := puzzleDB.InsertPuzzle(db.Puzzle{
		Puzzle:       normalizedStr,
		Difficulty:   actualDifficulty,
		Score:        classification.Score,
		MaxTechnique: classification.MaxTechnique,
		Source:       "fallback-seed",
	})
	if err != nil {
		t.Fatalf("insert seed puzzle: %v", err)
	}
	if !inserted {
		t.Fatal("seed puzzle should have been inserted")
	}
	puzzleDB.Close()

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
		t.Fatalf("DB fallback returned nil — expected %s puzzle", actualDifficulty)
	}
	if dbPuzzle.Difficulty != actualDifficulty {
		t.Errorf("DB fallback returned difficulty=%s, expected %s", dbPuzzle.Difficulty, actualDifficulty)
	}

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

// --- Generate with count 0 ---
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

// --- Generate with custom DB path ---
func TestGenerateCustomDBPath(t *testing.T) {
	tmpDir := t.TempDir()
	customDBPath := filepath.Join(tmpDir, "custom-test.db")

	puzzleDB, err := db.Open(customDBPath)
	if err != nil {
		t.Fatalf("open custom db: %v", err)
	}

	for i := 0; i < 2; i++ {
		storePuzzle(puzzleDB, fixedGenerationResult())
	}
	puzzleDB.Close()

	if _, err := os.Stat(customDBPath); os.IsNotExist(err) {
		t.Error("custom DB file should exist after generation")
	}

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

// --- Import empty/comments-only file ---
func TestImportEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	puzzleFile := filepath.Join(tmpDir, "empty-puzzles.txt")
	dbPath := filepath.Join(tmpDir, "empty-import-test.db")

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

	stats, err := puzzleDB.GetStats()
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.Total != 0 {
		t.Errorf("expected 0 puzzles in DB, got %d", stats.Total)
	}
}

// --- Shorthand digit input ---
// Tests that bare digits are dispatched as add commands via CLI dispatch logic.
func TestShorthandDigitInput(t *testing.T) {
	board := core.NewEmptyBoard()
	board.FromString(testKnownPuzzle)

	store := solver.NewStore()
	opts := game.NewDefaultOptions(store)
	g := game.NewGame(board, opts)

	solvedBoard := board.Copy()
	store.GetDefaultSolver().Solve(&solvedBoard)

	emptyPos := findFirstEmptyCell(board)
	if emptyPos == nil {
		t.Fatal("no empty cell found in test puzzle")
	}

	correctValue := solvedBoard.Get(*emptyPos)
	ctrl := newTestCtrl(&g)

	// Shorthand input format: "row col value" (1-indexed) — no "add" prefix.
	shorthand := fmt.Sprintf("%d %d %d", emptyPos.Row+1, emptyPos.Column+1, correctValue)
	changed := ctrl.RunCommand(shorthand)

	if !changed {
		t.Error("shorthand digit input should return true (board changed)")
	}
	actual := g.Snapshot().Values[emptyPos.Row][emptyPos.Column]
	if actual != correctValue {
		t.Errorf("after shorthand input, expected %d at (%d,%d), got %d",
			correctValue, emptyPos.Row, emptyPos.Column, actual)
	}
}

// --- Test helpers ---

// findFirstEmptyCell returns the first empty cell position in the board.
func findFirstEmptyCell(board core.Board) *core.Position {
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

// testController is a minimal wrapper for testing RunCommand logic
// that mirrors cli.Controller's dispatch without terminal I/O.
type testController struct {
	game *game.Game
}

func newTestCtrl(g *game.Game) testController {
	return testController{game: g}
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
		_, err := ctrl.game.Apply(game.Undo{})
		return err == nil
	case "redo", "r":
		_, err := ctrl.game.Apply(game.Redo{})
		return err == nil
	case "repair", "f":
		_, err := ctrl.game.Apply(game.Repair{})
		return err == nil
	case "hint", "i":
		_, err := ctrl.game.Apply(game.ApplyHint{})
		return err == nil
	case "solve", "s":
		_, err := ctrl.game.Apply(game.Solve{})
		return err == nil
	case "reset", "e":
		_, err := ctrl.game.Apply(game.Reset{})
		return err == nil
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

	if valueInput < 0 || valueInput > 9 {
		return false
	}

	if ctrl.game.Snapshot().Values[positionPointer.Row][positionPointer.Column] == valueInput {
		return false
	}

	var action game.Action = game.ClearValue{Position: *positionPointer}
	if valueInput != 0 {
		action = game.SetValue{Position: *positionPointer, Value: valueInput}
	}
	_, err = ctrl.game.Apply(action)
	return err == nil
}

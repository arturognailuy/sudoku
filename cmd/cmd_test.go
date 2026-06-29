package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnailuy/sudoku/core"
	"github.com/gnailuy/sudoku/db"
	"github.com/gnailuy/sudoku/solver"
)

func TestGenerateOnePuzzle(t *testing.T) {
	result := generateOnePuzzle("hard", 5000, 5)
	if result.Puzzle.GetFilledCellsCount() == 0 {
		t.Fatal("generated puzzle has no filled cells")
	}
	if result.Classification.Difficulty == "" {
		t.Fatal("classification has no difficulty")
	}
	if result.Classification.Score <= 0 {
		t.Fatal("classification has no score")
	}
}

func TestStorePuzzle(t *testing.T) {
	// Create an in-memory DB.
	puzzleDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer puzzleDB.Close()

	result := generateOnePuzzle("hard", 5000, 5)

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

	// Generate and store 3 puzzles.
	for i := 0; i < 3; i++ {
		result := generateOnePuzzle("evil", 5000, 5)
		storePuzzle(puzzleDB, result)
	}

	stats, err := puzzleDB.GetStats()
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.Total == 0 {
		t.Fatal("no puzzles stored")
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

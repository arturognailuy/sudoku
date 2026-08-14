---
domain: Designs
status: Active
entry_points:
  - main.go
  - cmd/root.go
dependencies:
  - .aidoc/INDEX.md
  - .aidoc/designs/game-engine.md
---

# E2E Test Scenarios

End-to-end test scenarios that treat the Sudoku CLI as a black box. Each scenario describes what a regular user would do, the expected behavior, and how to verify it.

These scenarios are designed to be run manually or via shell scripts against a built binary. They are **not** Go test files — E2E testing operates outside the codebase, treating it as a black box.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/INDEX.md` | Discovery index and project reading chains |
| `.aidoc/designs/game-engine.md` | Engine actions and state that frontends expose to users |
| `AGENT.md` | Required verification discipline for feature and bug-fix PRs |

## Prerequisites

```bash
go build -o sudoku .
export SUDOKU_E2E_DIR=$(mktemp -d)
export XDG_DATA_HOME=$SUDOKU_E2E_DIR/data
export SUDOKU_DB=$SUDOKU_E2E_DIR/test-puzzles.db
```

Root play commands auto-store puzzles under `XDG_DATA_HOME`; generate and import commands use `--db $SUDOKU_DB`. These settings isolate every database touched by the scenarios.

---

## 1. Interactive Play

### 1.1 Input a Known Puzzle
```bash
echo "quit" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Board is displayed. Game starts. Exits on "quit".

### 1.2 Input a Puzzle Using Dots Notation
```bash
echo "quit" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Dots (`.`) are treated as empty cells. Board displays correctly.

### 1.3 Input a Puzzle Using Zeros Notation
```bash
echo "quit" | ./sudoku --input "003020600900305001001806400008102900700000008006708200002609500800203009005010300"
```
**Expected:** Zeros are converted to empty cells. Same puzzle as 1.2.

### 1.4 Generate a Puzzle by Difficulty Level
```bash
echo "quit" | ./sudoku --level easy
```
**Expected:** A puzzle is generated (may show mismatch warning if best-effort misses). Game starts.

### 1.5 Invalid Input (Too Short)
```bash
./sudoku --input "123"
```
**Expected:** Error message about invalid puzzle string. Non-zero exit or error output.

### 1.6 Invalid Level Flag
```bash
./sudoku --level banana
```
**Expected:** Usage error shown with valid difficulty levels listed.

### 1.7 Help Flag
```bash
./sudoku --help
```
**Expected:** Usage printed with subcommands (`generate`, `import`) and flags listed.

### 1.8 Backward Compatibility: `--input` Flag
```bash
echo "quit" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Works the same as before the cobra migration.

### 1.9 Backward Compatibility: `--level` Flag
```bash
echo "quit" | ./sudoku --level easy
```
**Expected:** Works the same as before the cobra migration.

---

## 2. Game Commands

The game-command scenarios verify the stable engine boundary through the real terminal frontend: `cli.Controller` renders detached snapshots and submits typed actions while preserving the established command output and behavior.

For these scenarios, start a game with a known puzzle:
```bash
./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```

### 2.1 Add a Value (`add` / `a` / bare digits)
**Input:** `add 1 1 4` or `a 1 1 4` or `1 1 4`
**Expected:** Cell (1,1) is set to 4. Board updates.

### 2.2 Clear a Value (`clear` / `d`)
**Input:** After adding a value at (1,1): `clear 1 1` or `d 1 1`
**Expected:** Cell (1,1) is reset to empty.

### 2.3 Undo (`undo` / `u`)
**Input:** After adding a value: `undo` or `u`
**Expected:** Last move is reversed. Cell returns to previous state.

### 2.4 Redo (`redo` / `r`)
**Input:** After undo: `redo` or `r`
**Expected:** Undone move is re-applied, including whether the value is valid or invalid.

### 2.5 Check Board (`check` / `c`)
**Input:** `check` or `c`
**Expected (correct board):** "The current board is correct."
**Expected (incorrect board):** "You have entered incorrect value(s)."

### 2.6 Repair Board (`repair` / `f`)
**Input:** After entering wrong values: `repair` or `f`
**Expected:** All incorrect inputs are removed. Board returns to a valid state.

### 2.7 Reset Board (`reset` / `e`)
**Input:** After making several moves: `reset` or `e`
**Expected:** Board returns to the original problem state. All user inputs cleared.

### 2.8 Hint (`hint` / `i`)
**Input:** `hint` or `i`
**Expected:** A correct value is filled into a cell. The technique and reason are displayed.

### 2.9 Solve (`solve` / `s`)
**Input:** `solve` or `s`
**Expected:** The complete solution is displayed. "Congratulations" message shown.

### 2.10 Quit (`quit` / `q`)
**Input:** `quit` or `q`
**Expected:** Game exits cleanly.

### 2.11 Shorthand Digit Input
**Input:** `1 2 3` (no `add` prefix)
**Expected:** Treated as `add 1 2 3`. Cell (1,2) is set to 3.

### 2.12 Divergent Input Truncates Redo History
**Input:** Add a value, undo it, add a different value, then run `redo`.
**Expected:** The different value remains in the cell and the abandoned value does not reappear because the new action replaced the old redo branch.

---

## 3. Puzzle Generation CLI

### 3.1 Generate Help
```bash
./sudoku generate --help
```
**Expected:** Shows all flags: `-n`, `-d`, `-w`, `-t`, `--rounds`, `--db`.

### 3.2 Generate Easy Puzzles
```bash
./sudoku generate -n 5 -d easy --db $SUDOKU_DB
```
**Expected:** Generates 5 puzzles. Reports: generated count, stored count, duplicates, by-level breakdown. Some may be classified at a different difficulty than requested (mismatch is expected for easy/medium targets).

### 3.3 Generate Hard Puzzles
```bash
./sudoku generate -n 3 -d hard --db $SUDOKU_DB
```
**Expected:** Generates 3 puzzles. Each is classified and stored. Duration reported.

### 3.4 Generate with Parallel Workers
```bash
./sudoku generate -n 4 -d evil -w 4 --db $SUDOKU_DB
```
**Expected:** Uses 4 parallel goroutines. Completes generation. All puzzles stored.

### 3.5 Generate with Invalid Difficulty
```bash
./sudoku generate -d invalid
```
**Expected:** Error: "Invalid difficulty level". Exit code 1.

### 3.6 Generate with Count 0
```bash
./sudoku generate -n 0 -d easy
```
**Expected:** Error: "count must be positive". Exit code 1.

### 3.7 Generate with Custom DB Path
```bash
./sudoku generate -n 2 -d hard --db /tmp/custom-test.db
```
**Expected:** DB file created at `/tmp/custom-test.db`. Puzzles stored there.

### 3.8 Generate Dedup
```bash
./sudoku generate -n 10 -d evil --db $SUDOKU_DB
./sudoku generate -n 10 -d evil --db $SUDOKU_DB
```
**Expected:** Second run may report duplicates if the same normalized puzzles are generated.

---

## 4. Import CLI

### 4.1 Import Help
```bash
./sudoku import --help
```
**Expected:** Shows flags: `-f`, `--source`, `--db`.

### 4.2 Import Puzzles from File
Create `test-puzzles.txt`:
```
# Test puzzles
..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3..
003020600900305001001806400008102900700000008006708200002609500800203009005010300
```
```bash
./sudoku import -f test-puzzles.txt --source "test" --db $SUDOKU_DB
```
**Expected:** Puzzles are classified, normalized, and stored. The two lines represent the same puzzle (different notation), so one is stored and one is a duplicate.

### 4.3 Import with Invalid Lines
Create `mixed-puzzles.txt`:
```
..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3..
123456
abc
```
**Expected:** Valid puzzle stored. Invalid lines (too short, wrong characters) are skipped with stderr messages. Valid lines are processed.

### 4.4 Import with Source Label
```bash
./sudoku import -f puzzles.txt --source "top1465" --db $SUDOKU_DB
```
**Expected:** Custom source label ("top1465") is stored with each puzzle.

### 4.5 Import Missing File
```bash
./sudoku import -f nonexistent.txt
```
**Expected:** Error: "no such file". Exit code 1.

### 4.6 Import Empty / Comments-Only File
Create `empty.txt`:
```
# Only comments
# No puzzles here

```
```bash
./sudoku import -f empty.txt --db $SUDOKU_DB
```
**Expected:** Report shows 0 total lines, 0 stored. No errors.

### 4.7 Import Dedup
```bash
./sudoku import -f test-puzzles.txt --db $SUDOKU_DB
./sudoku import -f test-puzzles.txt --db $SUDOKU_DB
```
**Expected:** Second import reports all as duplicates.

---

## 5. Database and Fallback

### 5.1 Auto-Store on Play
```bash
echo "quit" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Puzzle is automatically stored in the DB at `~/.local/share/sudoku/puzzles.db`.

### 5.2 DB Fallback Path
1. Pre-populate the DB with easy puzzles: `./sudoku generate -n 20 -d easy --db $SUDOKU_DB`
2. Request an easy puzzle: `echo "quit" | ./sudoku --level easy --db $SUDOKU_DB`

**Expected:** If best-effort generation misses the target, the system falls back to the DB. If a match is found in the DB, no mismatch warning is shown. If the DB is also empty for that difficulty, the mismatch warning fires.

### 5.3 Mismatch Warning
```bash
echo "quit" | ./sudoku --level easy
```
**Expected (empty DB):** Best-effort likely misses easy target. Warning shown: "Requested difficulty: Easy. Generated puzzle difficulty: Medium/Hard. Enjoy!"

### 5.4 Multiple-Solution Puzzle Input
```bash
echo "quit" | ./sudoku --input "....7....6..195....98....6.8...6...34..8.3..17...2...6.6....28....419..5....8..79"
```
**Expected:** Warning about multiple solutions printed. Game still starts (plays with the first solution found).

---

## 6. Cobra Subcommand Structure

### 6.1 Root Command Shows Subcommands
```bash
./sudoku
```
**Expected:** Shows available commands: `generate`, `import`, and usage for the default play mode.

---

## 7. Future Scenarios (Not Yet Implemented)

These scenarios should be added as the project evolves:

- **Large import progress indicator:** Import 150+ puzzles → progress indicator fires every 100 puzzles.
- **Minimum-clues guard:** Import a puzzle with fewer than 17 clues → rejected or warned (prevents solver hang on near-empty boards).
- **Played tracking:** Mark puzzles as played → DB query skips played puzzles.
- **Concurrent DB access:** Multiple generate workers writing to the same DB → no corruption (WAL mode).
- **Manual notes:** After a frontend exposes note actions, toggle and clear notes through that frontend and verify the rendered candidates.
- **Automatic peer-note cleanup:** Add the same note to row, column, box, and non-peer cells; set the value and verify only target and peer notes are removed.
- **Unified note history:** Toggle or clear notes, set a value that removes peer notes, then undo and redo; verify values, invalid markers, and notes restore atomically.
- **Session save and restore:** After a frontend exposes persistence transport, save a session containing values, invalid entries, notes, and an undone action; restart, restore it, and verify the rendered state plus both undo and redo behavior.
- **Corrupt session rejection:** Attempt to load malformed or unsupported session data through the frontend and verify that it reports the failure without replacing the active session.

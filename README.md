# Sudoku

A CLI Sudoku game in Go. Generate puzzles at various difficulty levels, solve them, and play interactively with undo/redo and hints.

Features:
- 23 strategy solvers across 5 difficulty tiers (Easy → Evil)
- Strategy-based difficulty classification with HoDoKu-weighted scoring
- Best-effort puzzle generator with time/round limits
- SQLite puzzle database with automatic storage, dedup, and fallback lookup
- Batch generation CLI for offline puzzle creation
- Import CLI for loading puzzles from files
- Interactive play with manual notes, undo/redo, and technique-aware hints
- Explicit atomic session saves with validated resume
- Opt-in full-screen Bubble Tea interface with keyboard navigation

## Build

```bash
go build
```

## Play

```bash
# Random puzzle (default: hard)
./sudoku

# Choose difficulty
./sudoku -l medium
./sudoku -l hard
./sudoku -l expert
./sudoku -l evil

# Custom board (use . for empty cells)
./sudoku -i .56.4.7...1.5....6.......19...9.....3.58..2...4...6...1.....93....4....22.3.1....

# Resume a session saved from the interactive prompt
./sudoku --resume game.json
```

## Full-Screen TUI

The existing line-oriented game remains the default. Launch the optional terminal UI with the same puzzle sources:

```bash
./sudoku tui --level medium
./sudoku tui --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
./sudoku tui --resume game.json
```

Use arrows or `h`/`j`/`k`/`l` to move, `1`–`9` to enter a value, `n` to toggle note mode, `u`/`r` for history, `?` then Enter to preview/apply a hint, `S` to save, and `q` to quit. Unsaved sessions require confirmation.

## Batch Generate

Generate puzzles and store them in the database:

```bash
# Generate 100 puzzles targeting hard difficulty
./sudoku generate -n 100 -d hard

# Generate 500 evil puzzles with 4 parallel workers
./sudoku generate -n 500 -d evil -w 4

# Custom timeout and rounds per puzzle
./sudoku generate -n 50 -d expert -t 60s --rounds 20
```

## Import Puzzles

Import puzzles from a text file (one per line, 81 chars):

```bash
# Import from file (supports . or 0 for empty cells)
./sudoku import -f puzzles.txt

# Custom source label
./sudoku import -f top1465.txt --source "top1465"
```

## In-Game Commands

During play, enter moves as `row col value` (for example, `1 2 5`). Commands accept the displayed long form or alias:

- `add`, `a <row> <column> <value>` — set a value
- `clear`, `d <row> <column>` — clear a value
- `note`, `n <row> <column> <value>` — toggle a manual candidate note
- `notes-clear`, `x <row> <column>` — clear a cell's notes
- `save <path>` — atomically save values, notes, and undo/redo history
- `undo`, `u` / `redo`, `r` — move through value and note history
- `hint`, `i` — apply a technique-aware hint
- `check`, `c` / `repair`, `f` — inspect or remove invalid entries
- `reset`, `e` / `solve`, `s` — restart or solve the puzzle
- `help`, `h` / `quit`, `q` — show command help or exit without saving

Manual notes use a fixed 3×3 candidate layout. Saving is explicit: use `save` before `quit`, then pass the same file to `--resume` later.

## Development

```bash
go test ./...    # Run tests
go vet ./...     # Static analysis
```

For AI agents: start with [`AGENT.md`](AGENT.md) → [`.aidoc/INDEX.md`](.aidoc/INDEX.md).

## License

MIT

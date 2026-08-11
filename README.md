# Sudoku

A CLI Sudoku game in Go. Generate puzzles at various difficulty levels, solve them, and play interactively with undo/redo and hints.

Features:
- 23 strategy solvers across 5 difficulty tiers (Easy → Evil)
- Strategy-based difficulty classification with HoDoKu-weighted scoring
- Best-effort puzzle generator with time/round limits
- SQLite puzzle database with automatic storage, dedup, and fallback lookup
- Batch generation CLI for offline puzzle creation
- Import CLI for loading puzzles from files
- Interactive play with undo/redo and technique-aware hints

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
```

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

During play, enter moves as `row col value` (e.g., `1 2 5`). Additional commands:

- `u` — undo last move
- `r` — redo
- `h` — get a hint
- `q` — quit

## Development

```bash
go test ./...    # Run tests
go vet ./...     # Static analysis
```

For AI agents: start with [`AGENT.md`](AGENT.md) → [`.aidoc/INDEX.md`](.aidoc/INDEX.md).

## License

MIT

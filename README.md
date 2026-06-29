# Sudoku

A CLI Sudoku game in Go. Generate puzzles at various difficulty levels, solve them, and play interactively with undo/redo and hints.

Features:
- 23 strategy solvers across 5 difficulty tiers (Easy → Evil)
- Strategy-based difficulty classification with HoDoKu-weighted scoring
- Best-effort puzzle generator with time/round limits
- SQLite puzzle database with automatic storage, dedup, and fallback lookup
- Interactive play with undo/redo and technique-aware hints

## Build

```bash
go build
```

## Play

```bash
# Random puzzle (default: easy)
./sudoku

# Choose difficulty
./sudoku -l medium
./sudoku -l hard
./sudoku -l expert
./sudoku -l evil

# Custom board (use . for empty cells)
./sudoku -i .56.4.7...1.5....6.......19...9.....3.58..2...4...6...1.....93....4....22.3.1....
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

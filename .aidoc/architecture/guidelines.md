# Architecture Guidelines

## Package Structure

Each package has a single responsibility:

| Package | Responsibility |
|---------|---------------|
| `core` | Board representation, cell, position, validation, normalization, string serialization |
| `solver` | Solver interface, solver registry, solver implementations |
| `generator` | Puzzle generation — filled board creation, cell removal with uniqueness checks |
| `game` | Game state management, undo/redo/hints, interactive CLI play loop |
| `cli` | Command-line flag parsing |
| `util` | Generic helpers (shuffle, array generation) |

## Go Conventions

- **Formatting:** `gofmt` / `goimports`. No exceptions.
- **Naming:** Follow Go conventions — exported names are `PascalCase`, unexported are `camelCase`.
- **Error handling:** Return errors explicitly. No panics in normal code paths (panics are reserved for bug detection in invariants).
- **Dependencies:** Minimize external dependencies. Current deps: `spf13/pflag`, `spf13/cobra`, `thediveo/enumflag/v2`.

## Solver Interface Contract

All solvers implement `ISudokuSolver` (defined in `solver/sudoku_solver.go`):

```go
type ISudokuSolver interface {
    GetKey() string
    GetDisplayName() string
    GetDescription() string
    IsReliable() bool
    Solve(board *core.SudokuBoard) bool
    Hint(board *core.SudokuBoard) *core.Cell
    CountSolutions(board *core.SudokuBoard) int
}
```

### Key Properties

- **`GetKey()`** — unique string identifier, used as the key in `SudokuSolverStore`.
- **`IsReliable()`** — a reliable solver can solve any valid Sudoku. The default backtracking solver is reliable. Strategy solvers (naked singles, hidden singles, etc.) are **not** reliable — they can only solve puzzles within their technique scope.
- **`Solve()`** — attempts to solve the board in-place. Returns `false` if the solver cannot fully solve it.
- **`Hint()`** — returns the next cell the solver can determine, without modifying the board. Returns `nil` if stuck.
- **`CountSolutions()`** — unreliable solvers should return `0` (inherited from `BaseSolver`).

### Adding a New Solver

1. Create `solver/<technique>_solver.go` embedding `BaseSolver`.
2. Set `Reliable: false` for strategy solvers.
3. Implement `Solve()`, `Hint()`. Inherit `CountSolutions()` from `BaseSolver`.
4. Register in `NewSudokuSolverStore()`.
5. Add the solver key to the appropriate `StrategySolverKeys` in difficulty definitions.
6. Write tests in `solver/<technique>_solver_test.go`.

## Layer Separation

```
CLI (flags) → main.go → Generator (create puzzle) → Game (play loop)
                              ↓
                         Solver Store → ISudokuSolver implementations
                              ↓
                         Core (board, cell, position, validation)
```

- **Core** has no dependencies on other packages.
- **Solver** depends only on `core` and `util`.
- **Generator** depends on `core`, `solver`, and `util`.
- **Game** depends on `core` and `solver`.
- **CLI** depends on `generator` (for difficulty enum).

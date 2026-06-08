---
domain: Architecture
status: Active
entry_points:
  - solver/sudoku_solver.go
dependencies:
  - .aidoc/INDEX.md
  - .aidoc/designs/difficulty-model.md
---

# Architecture Guidelines

The Sudoku project follows a layered architecture where each package owns a single concern.
This doc captures the design constraints, layer boundaries, and solver contract that code alone doesn't express.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/INDEX.md` | Discovery index with reading chains |
| `.aidoc/designs/difficulty-model.md` | Difficulty model design and target state |
| `AGENT.md` | Operational rules for AI agents working on this repo |

## Why This Structure

The package layout ensures strategy solvers can be added independently without touching the generator, game, or CLI.
Each solver is self-contained: implement the interface, register in the store, and add the key to difficulty definitions.
The generator already has plumbing for strategy-based validation — it checks `StrategySolverKeys` during cell removal — so new solvers plug in without generator changes.

## Layer Boundaries

```
CLI (flags) → main.go → Generator (create puzzle) → Game (play loop)
                              ↓
                         Solver Store → ISudokuSolver implementations
                              ↓
                         Core (board, cell, position, validation)
```

**Dependency rules:**
- `core` has zero dependencies on other packages (leaf layer).
- `solver` depends only on `core` and `util`.
- `generator` depends on `core`, `solver`, and `util`.
- `game` depends on `core` and `solver`. Does **not** depend on `cli` or `generator`.
- `cli` depends on `generator` (for difficulty enum). Only `main.go` imports `cli`.
- `util` has no internal dependencies (pure helpers).

Violations of these boundaries indicate a design problem.

## Solver Interface Contract

Defined in `solver/sudoku_solver.go`. All solvers implement `ISudokuSolver` with these semantics:

- **`GetKey()`** — unique string identifier used as the map key in `SudokuSolverStore`.
- **`IsReliable()`** — a reliable solver can solve any valid board (the backtracking solver). Strategy solvers are **not** reliable — they only handle puzzles within their technique scope.
- **`Solve(board)`** — attempts to solve in-place. Returns `false` if the solver cannot fully solve.
- **`Hint(board)`** — returns the next determinable cell without modifying the board. Returns `nil` if stuck.
- **`CountSolutions(board)`** — only meaningful for reliable solvers. Unreliable solvers inherit `BaseSolver.CountSolutions` which returns `0`.

`BaseSolver` (`solver/sudoku_solver.go`) provides default implementations for metadata methods and `CountSolutions`. Strategy solvers embed `BaseSolver` with `Reliable: false`.

## Adding a New Solver

1. Create `solver/<technique>_solver.go` embedding `BaseSolver`.
2. Implement `Solve()` and `Hint()`. Inherit `CountSolutions()` from `BaseSolver`.
3. Register in `solver.NewSudokuSolverStore()`.
4. Add the solver key to the appropriate difficulty level's `StrategySolverKeys` in `generator/sudoku_generator_difficulty.go`.
5. Write tests in `solver/<technique>_solver_test.go`.

## Design Constraints

- **Interface naming:** `ISudokuSolver` uses the `I` prefix (non-standard Go). Planned for rename in the core refactoring phase — see `.aidoc/designs/roadmap.md` Phase 2.
- **Error vs panic:** Methods called with invalid state from within the system `panic` (bug detection). Methods processing user input return errors. This split is intentional.
- **Geometric distribution stop:** The generator uses `util.RandomBool(0.125)` to probabilistically stop cell removal after reaching the target clue range. This produces natural variation within a difficulty band.

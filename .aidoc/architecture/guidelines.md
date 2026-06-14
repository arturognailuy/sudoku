---
domain: Architecture
status: Active
entry_points:
  - solver/solver.go
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
The generator already has plumbing for strategy-based validation — it calls `StrategySolver.Apply()` during cell removal — so new solvers plug in without generator changes.

## Layer Boundaries

```
CLI (flags) → main.go → Generator (create puzzle) → Game (play loop)
                              ↓
                         Solver Store → Solver implementations
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

Defined in `solver/solver.go`. Three interfaces form the solver hierarchy:

### `Solver` (base metadata)

All solvers implement `Solver`:
- **`GetKey()`** — unique string identifier used as the map key in `Store`.
- **`GetDisplayName()`** — human-readable solver name.
- **`GetDescription()`** — description of the solver's approach.

`Base` (`solver/solver.go`) provides default implementations for these metadata methods.

### `StrategySolver` (technique-based)

Extends `Solver`. Used by solvers that apply a single technique (e.g., naked singles, hidden singles):
- **`Apply(board)`** — finds the next move using this technique. Returns `*Move` with the cell to fill, the technique name, and a human-readable explanation. Returns `nil` if the technique cannot make progress.

Strategy solvers are **not reliable** — they only handle puzzles within their technique scope.

### `CompleteSolver` (full solve)

Extends `Solver`. Used by solvers that can fully solve any valid board (e.g., backtracking):
- **`Solve(board)`** — attempts to solve in-place. Returns `false` if the solver cannot fully solve.
- **`Hint(board)`** — returns the next determinable move as `*Move` without modifying the board. Returns `nil` if stuck.
- **`CountSolutions(board)`** — returns the number of solutions for the board.

### `Move` struct

`Move` (`solver/move.go`) is the return type for `StrategySolver.Apply()` and `CompleteSolver.Hint()`:
- **`Cell`** — the cell to fill (position + value).
- **`Technique`** — technique identifier (e.g., `"naked-single"`, `"backtracker"`).
- **`Reason`** — human-readable explanation for display in hints.

### `Store`

`Store` (`solver/store.go`) holds both `CompleteSolver` and `StrategySolver` implementations:
- **`GetDefaultSolver()`** — returns the default `CompleteSolver` (backtracker).
- **`GetStrategySolverByKey(key)`** — returns a `StrategySolver` by key, or `nil`.
- **`RegisterStrategy(s)`** — registers a `StrategySolver`.

## Adding a New Strategy Solver

1. Create `solver/<technique>_solver.go` embedding `Base`.
2. Implement `StrategySolver`: write `Apply()` that returns `*Move` with the technique name and reason.
3. Register in `solver.NewStore()` using `store.RegisterStrategy(s)`.
4. Add the solver key to the appropriate difficulty level's `StrategySolverKeys` in `generator/difficulty.go`.
5. Write tests in `solver/<technique>_solver_test.go`.

## Design Constraints

- **Interface naming:** Types follow Go conventions. Examples: `Solver` (base interface), `StrategySolver`, `CompleteSolver`, `Base`, `Backtracker`, `Store`, `Move`, `Board`, `Game`, `Difficulty`, `Options`, `MoveRecord`, `CandidateSet`.
- **Candidate computation:** `Board.Candidates(pos)` computes valid candidates on the fly by scanning row, column, and box peers. The `CandidateSet` bitfield type provides compact representation (`uint16`, bits 1–9) for the result. Board itself stores only the grid — no cached candidate state to maintain. Strategy solvers call `board.Candidates(pos)` when they need candidates.
- **Error vs panic:** Methods called with invalid state from within the system `panic` (bug detection). Methods processing user input return errors. This split is intentional.
- **Geometric distribution stop:** The generator uses `util.RandomBool(0.125)` to probabilistically stop cell removal after reaching the target clue range. This produces natural variation within a difficulty band.

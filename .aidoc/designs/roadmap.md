---
domain: Designs
status: Active
entry_points: []
dependencies:
  - .aidoc/architecture/guidelines.md
  - .aidoc/designs/difficulty-model.md
---

# Roadmap

Future development phases for the Sudoku project, from core refactoring through UI readiness.
Refactoring comes first — clean up while the codebase is small, then build new solvers on solid foundations.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/architecture/guidelines.md` | Current layer boundaries and solver contract |
| `.aidoc/designs/difficulty-model.md` | Difficulty model design (strategy-based, with puzzle database) |
| `.aidoc/INDEX.md` | Discovery index |

## Phase 2: Core Refactoring

Refactor the codebase while it's small, before adding new solvers.
New strategy solvers should be built on clean, modern infrastructure.

- **Core data structures:** Revisit `Board`, `Cell`, `Position`, and related types
  for clarity, performance, and extensibility. Add `CandidateSet` bitfield for
  constraint propagation (Option B approved).
- **Solver architecture:** Split the solver interface into `Solver` + `StrategySolver` +
  `CompleteSolver` with a `Move` struct for richer hints (Option B approved).
- **Game design:** Refactor the game loop and state management for cleaner separation
  of concerns.

## Phase 3: Strategy Solvers

Implement solving techniques incrementally, one per PR.
Each solver implements the solver interface (renamed in Phase 2), gets tests,
and registers in the solver store.

Priority order:
1. Naked Singles
2. Hidden Singles
3. Naked/Hidden Pairs and Triples
4. Pointing Pairs / Box-Line Reduction
5. X-Wing (stretch)

See `.aidoc/designs/difficulty-model.md` for the strategy tier definitions
and `.aidoc/architecture/guidelines.md` for the step-by-step solver addition guide.

## Phase 4: Generator Integration and Puzzle Database

Wire strategy solvers into puzzle generation and build a puzzle database:

1. Generate puzzles offline using the existing generator.
2. Classify each puzzle by technique tier (highest strategy required to solve) and clue count.
3. Store puzzles with difficulty metadata in a database.
4. Serve puzzles by database lookup — filter by requested difficulty level.

Clue-count ranges are a secondary constraint: a puzzle must fall within the expected
clue band *and* require techniques at the target tier.

## Phase 5: UI-Ready Core Engine

Refactor the core game into a reusable engine that can serve as the backend for
multiple UI implementations — similar to how GNU Go provides a core engine
used by various graphical frontends.

Goals:
- **Clean API boundary:** The engine exposes game state, moves, undo/redo, hints,
  and validation through a well-defined API. No terminal I/O assumptions in the core.
- **Note-taking support:** Players can annotate cells with candidate values
  (pencil marks). The engine tracks notes as part of the game state, including
  undo/redo for note operations.
- **CLI as one frontend:** The existing CLI becomes one consumer of the engine API,
  not the only interface.
- **UI independence:** The engine makes no assumptions about rendering, input method,
  or platform. A web UI, TUI, or mobile app should all be viable frontends.

This phase depends on the core refactoring (Phase 2) to establish clean interfaces first.

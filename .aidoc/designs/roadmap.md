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

## Phase 3: Strategy Solvers

Implement solving techniques incrementally, organized by difficulty tier.
Each solver implements `StrategySolver` (with `Apply()` returning `*Move`),
gets tests, and registers in the solver store.

Implementation by difficulty tier:

**Easy tier (PR #6):** ✅
- Naked Singles — cell has exactly one candidate left.
- Hidden Singles — candidate appears in only one cell within a row, column, or box.
- Both registered in store, wired into Easy difficulty (`SolverKeys`).

**Intermediate tier (PR #7):** In review
- Naked Pairs/Triples — two/three cells in a unit share the same candidates exclusively; eliminating those candidates from other cells in the unit reveals singles.
- Pointing Pairs / Box-Line Reduction — candidate confined to single row/column within a box (or vice versa); elimination reveals singles.
- Both registered in store, wired into Medium difficulty (`SolverKeys`).
- Single-field design: `SolverKeys` holds the solvers introduced at this tier; `LowerTierSolverKeys` holds cumulative lower-tier solvers; `AllowedSolverKeys()` returns the full allowed set.

**Advanced tier (PR #8, stretch):**
- X-Wing — candidate in exactly two cells in two rows sharing the same columns.
- Wire into Hard difficulty.

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
- **CLI as one frontend:** The existing CLI is already separated into `cli/controller.go`,
  consuming the Game API. Additional UIs follow the same pattern.
- **UI independence:** The engine makes no assumptions about rendering, input method,
  or platform. A web UI, TUI, or mobile app should all be viable frontends.

The core Game struct is already pure state (no I/O) after Phase 2. This phase
extends it with note-taking and formalizes the API as a stable engine boundary.

---
domain: Designs
status: Active
entry_points:
  - generator/sudoku_generator_difficulty.go
dependencies:
  - .aidoc/architecture/guidelines.md
---

# Difficulty Model

Difficulty is the project's core design challenge: the current model uses clue count (a poor proxy),
while the target model uses solving techniques (meaningful difficulty). This doc captures the gap and the migration path.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/architecture/guidelines.md` | Solver interface contract and layer boundaries |
| `.aidoc/INDEX.md` | Discovery index |

## Current Model (Clue-Count)

Difficulty is defined solely by pre-filled cell count in `generator/sudoku_generator_difficulty.go`:

| Level | Clues (min–max) |
|-------|-----------------|
| Easy | 45–59 |
| Medium | 32–44 |
| Hard | 25–31 |
| Extreme | 20–24 |
| Evil | 17–19 |

### Why This Is Insufficient

A 25-clue puzzle might be trivially solvable with naked singles, or might require advanced techniques — clue count doesn't distinguish. All `StrategySolverKeys` are currently empty (`[]string{}`), so the generator never validates technique requirements. The geometric distribution stop means difficulty is random within each band.

Users cannot request puzzles that require specific solving techniques, making "Hard" meaningless beyond "fewer clues."

## Target Model (Strategy-Based)

Difficulty should be defined by the **hardest technique required to solve** the puzzle.

### Strategy Tiers

| Tier | Techniques |
|------|-----------|
| Basic | Naked singles, hidden singles |
| Intermediate | Naked/hidden pairs/triples, pointing pairs, box/line reduction |
| Advanced | X-wing, swordfish, XY-wing, simple coloring |
| Expert | Jellyfish, finned X-wing, ALS, forcing chains |

### Difficulty Mapping

| Level | Required Tier | Meaning |
|-------|---------------|---------|
| Easy | Basic only | Solvable with naked/hidden singles alone |
| Medium | Up to Intermediate | Requires at least one intermediate technique |
| Hard | Up to Advanced | Requires at least one advanced technique |
| Expert/Evil | Expert or guessing | Requires expert techniques or trial-and-error |

### Generation Flow

1. Generate a puzzle using the existing generator.
2. Attempt to solve with strategy solvers in tier order (basic → intermediate → advanced → expert).
3. Record the highest tier required.
4. If the required tier doesn't match the requested difficulty, reject and regenerate.

### Architecture Support Already In Place

The plumbing exists in the generator (`generator/sudoku_generator.go`):
- `SudokuDifficulty.StrategySolverKeys` lists solver keys to check during cell removal.
- The generator calls `solver.Hint()` on each listed strategy solver before confirming a removal.
- `SudokuSolverStore` maps solver keys to implementations.

What's missing: actual strategy solver implementations to register.

## Open Questions

<!-- TODO: (arturo) Determine if clue-count ranges should be preserved as a secondary constraint alongside technique requirements, or replaced entirely. -->
<!-- TODO: (arturo) Decide on rejection/regeneration limits — how many retries before falling back to a less constrained difficulty. -->

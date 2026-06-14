---
domain: Designs
status: Active
entry_points:
  - generator/difficulty.go
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

Difficulty is defined solely by pre-filled cell count in `generator/difficulty.go`:

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

### Clue Count as Secondary Constraint

Clue-count ranges are preserved as a secondary constraint alongside technique requirements.
A puzzle that requires a Jellyfish but has 50 clues isn't fun — the strategy difficulty
and the clue count must both fall within reasonable bounds for a satisfying experience.
The existing clue-count ranges define the acceptable band; technique requirements define
the minimum solving complexity.

### Generation and Storage Flow

Rather than reject-and-regenerate (which is expensive and may loop indefinitely for rare
technique requirements), puzzles are generated offline and stored:

1. Generate a puzzle using the existing generator.
2. Attempt to solve with strategy solvers in tier order (basic → intermediate → advanced → expert).
3. Record the highest tier required and the clue count.
4. Save the puzzle with its difficulty metadata to a database.
5. To serve a puzzle of a given difficulty, look up the database for a match.

This decouples generation (slow, offline, batch) from serving (fast, database lookup),
and avoids the need for retry limits or fallback logic.

### Architecture Support Already In Place

The plumbing exists in the generator (`generator/generator.go`):
- `Difficulty.StrategySolverKeys` lists solver keys to check during cell removal.
- The generator calls `solver.Apply()` on each listed strategy solver before confirming a removal.
- `Store` maps solver keys to implementations.

**Easy difficulty is now wired:** `StrategySolverKeys: ["naked-single", "hidden-single"]`.
The generator produces Easy puzzles that are solvable using only naked and hidden singles.
Medium and above still use empty keys (no technique constraint).


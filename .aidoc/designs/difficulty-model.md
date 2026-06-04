# Difficulty Model

## Current Model (Clue-Count)

Difficulty is defined solely by how many cells are pre-filled:

| Level | Clues (min–max) |
|-------|-----------------|
| Easy | 45–59 |
| Medium | 32–44 |
| Hard | 25–31 |
| Extreme | 20–24 |
| Evil | 17–19 |

### Limitations

- A puzzle with 25 clues might be trivially solvable with naked singles, or might require advanced techniques — clue count doesn't distinguish.
- All `StrategySolverKeys` are empty — no technique requirements enforced.
- The generator uses a geometric distribution to randomly stop removing cells within a range, so difficulty is random within a band.
- Users cannot request puzzles that require specific solving techniques.

## Target Model (Strategy-Based)

Difficulty should be defined by the **hardest technique required to solve** the puzzle.

### Strategy Tiers

| Tier | Techniques |
|------|-----------|
| Basic | Naked singles, hidden singles |
| Intermediate | Naked pairs/triples, hidden pairs/triples, pointing pairs, box/line reduction |
| Advanced | X-wing, swordfish, XY-wing, simple coloring |
| Expert | Jellyfish, finned X-wing, ALS, forcing chains |

### Difficulty Mapping

| Level | Required Tier | Meaning |
|-------|---------------|---------|
| Easy | Basic only | Solvable with naked/hidden singles alone |
| Medium | Up to Intermediate | Requires at least one intermediate technique |
| Hard | Up to Advanced | Requires at least one advanced technique |
| Expert/Evil | Expert or guessing | Requires expert techniques or trial-and-error |

### How It Works

1. Generate a puzzle (existing generator).
2. Attempt to solve with strategy solvers in tier order (basic → intermediate → advanced → expert).
3. Record the highest tier required to solve.
4. If the required tier doesn't match the requested difficulty, reject and regenerate.

### Architecture Support

The plumbing already exists:
- `SudokuDifficulty.StrategySolverKeys` — list of solver keys the generator checks during cell removal.
- `SudokuSolverStore` — registry to look up solvers by key.
- The generator already calls `solver.Hint()` on listed strategy solvers to verify solvability during cell removal.

What's missing: actual strategy solver implementations to register.

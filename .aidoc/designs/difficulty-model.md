---
domain: Designs
status: Active
entry_points:
  - generator/difficulty.go
dependencies:
  - .aidoc/architecture/guidelines.md
---

# Difficulty Model

Difficulty combines clue count with solving-technique requirements.
This doc captures the current model and the migration path toward fully strategy-based difficulty.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/architecture/guidelines.md` | Solver interface contract and layer boundaries |
| `.aidoc/INDEX.md` | Discovery index |

## Current Model (Clue-Count + Strategy Tiers)

Difficulty combines clue count with technique requirements in `generator/difficulty.go`:

| Level | Clues (min–max) | Strategy Tier | Solver Keys |
|-------|-----------------|---------------|-------------|
| Easy | 45–59 | Basic | naked-single, hidden-single |
| Medium | 32–44 | Intermediate | naked-pair, naked-triple, pointing-pair, hidden-pair |
| Hard | 25–31 | Advanced | x-wing, xy-wing, hidden-triple |
| Expert | 22–24 | Expert | swordfish, naked-quad, simple-coloring, hidden-quad |
| Evil | 17–22 | Evil | jellyfish |

Each level's allowed solvers = its own SolverKeys + all solvers from lower tiers.
During generation, the generator verifies that lower-tier solvers alone cannot solve
the puzzle — ensuring it genuinely requires at least one technique from this tier.

### Solver Inventory (14 solvers)

Solvers are split into per-size variants for accurate difficulty tiering. Shared
algorithms use a factory/parameterized pattern — e.g., `FishSolver` (X-Wing/Swordfish/
Jellyfish) and `NakedSubsetSolver` / `HiddenSubsetSolver` (pair/triple/quad).

| Solver Key | Display Name | Weight | Tier | Algorithm |
|------------|-------------|--------|------|-----------|
| naked-single | Naked Single | 4 | Easy | Direct |
| hidden-single | Hidden Single | 14 | Easy | Direct |
| naked-pair | Naked Pair | 60 | Medium | NakedSubsetSolver(size=2) |
| naked-triple | Naked Triple | 80 | Medium | NakedSubsetSolver(size=3) |
| pointing-pair | Pointing Pair | 50 | Medium | Direct |
| hidden-pair | Hidden Pair | 70 | Medium | HiddenSubsetSolver(size=2) |
| x-wing | X-Wing | 140 | Hard | FishSolver(size=2) |
| xy-wing | XY-Wing | 160 | Hard | Direct |
| hidden-triple | Hidden Triple | 100 | Hard | HiddenSubsetSolver(size=3) |
| swordfish | Swordfish | 150 | Expert | FishSolver(size=3) |
| naked-quad | Naked Quad | 120 | Expert | NakedSubsetSolver(size=4) |
| simple-coloring | Simple Coloring | 150 | Expert | Direct |
| hidden-quad | Hidden Quad | 150 | Expert | HiddenSubsetSolver(size=4) |
| jellyfish | Jellyfish | 300 | Evil | FishSolver(size=4) |

### Tier Rationale

Tiers are based on SudokuWiki's human-difficulty ordering (frequency × difficulty):

- **Easy:** Trivial techniques — scan for cells/units with one candidate.
- **Medium:** Basic pattern recognition — pairs, triples, pointing pairs. Hidden pairs
  are easier than X-Wing for humans.
- **Hard:** Requires systematic row/column scanning (X-Wing, XY-Wing) or identifying
  three hidden digits in three cells (Hidden Triple).
- **Expert:** Very hard to spot manually — 3-row/col fish patterns (Swordfish), four-cell
  subsets (Naked/Hidden Quad), graph coloring (Simple Coloring).
- **Evil:** Near-impossible to spot manually — 4-row/col fish patterns (Jellyfish). Future
  additions: BUG+1, Unique Rectangles.

## Difficulty Mapping

| Level | Required Tier | Meaning |
|-------|---------------|--------|
| Easy | Basic only | Solvable with naked/hidden singles alone |
| Medium | Up to Intermediate | Requires at least one naked-pair, naked-triple, pointing-pair, or hidden-pair |
| Hard | Up to Advanced | Requires at least one X-Wing, XY-Wing, or hidden-triple step |
| Expert | Up to Expert | Requires at least one swordfish, naked-quad, simple-coloring, or hidden-quad step |
| Evil | Up to Evil | Requires at least one jellyfish step |

### Clue Count as Secondary Constraint

Clue-count ranges are preserved as a secondary constraint alongside technique requirements.
The existing clue-count ranges define the acceptable band; technique requirements define
the minimum solving complexity.

### Architecture Support

The plumbing in `generator/difficulty.go`:
- `tierRegistry` (map) + `tierOrder` (slice) define the tier hierarchy — single source of truth.
- `Difficulty.SolverKeys` lists solver keys introduced at this tier.
- `Difficulty.AllowedSolverKeys()` returns the full allowed set (lower tiers + this tier).
- `Difficulty.LowerTierSolverKeys()` returns cumulative keys from all tiers below.
- During cell removal, the generator calls `solver.Apply()` on each allowed solver.
- After generation, `requiresThisTierSolver()` verifies lower-tier solvers alone can't solve.
- `Store` maps solver keys to implementations.

**Easy:** `SolverKeys: ["naked-single", "hidden-single"]`.
`LowerTierSolverKeys()` returns nil (lowest tier).

**Medium:** `SolverKeys: ["naked-pair", "naked-triple", "pointing-pair", "hidden-pair"]`.
`LowerTierSolverKeys()` returns Easy keys.

**Hard:** `SolverKeys: ["x-wing", "xy-wing", "hidden-triple"]`.
`LowerTierSolverKeys()` returns Easy + Medium keys.

**Expert:** `SolverKeys: ["swordfish", "naked-quad", "simple-coloring", "hidden-quad"]`.
`LowerTierSolverKeys()` returns Easy + Medium + Hard keys.

**Evil:** `SolverKeys: ["jellyfish"]`.
`LowerTierSolverKeys()` returns Easy + Medium + Hard + Expert keys.

## Scoring System

Each solver carries a `Weight` field representing its difficulty cost per application,
based on HoDoKu's established weights. A puzzle's total difficulty score is the sum
of all technique weights used during solving:

```
score = Σ(weight[technique] × times_used)
```

The `ScorePuzzle(store, moves)` function in `solver/scoring.go` computes the score
from a list of moves. Moves from unknown techniques (e.g., backtracker) contribute zero.

### Configuration

All tunable parameters — solver weights and clue-count ranges — are centralized in
`solver/config.go`. This is the single file to update when tuning parameters or
calibrating the difficulty system.

### Future: Score-Based Difficulty Ranges

`MinScore` and `MaxScore` fields on `Difficulty` will be added in Phase 4,
when the puzzle database provides enough data to calibrate score ranges.
The scoring infrastructure is ready — each new solver just sets its weight
in the constructor.

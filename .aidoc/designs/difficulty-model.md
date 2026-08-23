---
domain: Designs
status: Active
entry_points:
  - generator/difficulty.go
dependencies:
  - .aidoc/architecture/guidelines.md
  - .aidoc/designs/difficulty-calibration.md
---

# Difficulty Model

Difficulty combines clue count with solving-technique requirements and weighted solving moves. This document defines the current invariants and the evidence required before calibration changes those parameters.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/architecture/guidelines.md` | Solver interface contract and layer boundaries |
| `.aidoc/designs/difficulty-calibration.md` | Measurement contract and evidence required for policy changes |
| `.aidoc/INDEX.md` | Discovery index |

## Current Model (Clue-Count + Strategy Tiers)

Difficulty combines clue count with technique requirements in `generator/difficulty.go`:

| Level | Clues (min–max) | Strategy Tier | Solver Keys |
|-------|-----------------|---------------|-------------|
| Easy | 45–59 | Basic | naked-single, hidden-single |
| Medium | 32–44 | Intermediate | naked-pair, naked-triple, pointing-pair, hidden-pair |
| Hard | 25–31 | Advanced | x-wing, xy-wing, hidden-triple, w-wing |
| Expert | 22–24 | Expert | swordfish, naked-quad, simple-coloring, hidden-quad, xyz-wing |
| Evil | 17–21 | Evil | jellyfish, bug-plus-one, unique-rectangle, unique-rectangle-2, unique-rectangle-3, unique-rectangle-4, x-cycles, xy-chain |

Each level's allowed solvers = its own SolverKeys + all solvers from lower tiers.
During generation, the generator verifies that lower-tier solvers alone cannot solve
the puzzle — ensuring it genuinely requires at least one technique from this tier.

### Solver Inventory (23 solvers)

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
| w-wing | W-Wing | 150 | Hard | Direct |
| swordfish | Swordfish | 150 | Expert | FishSolver(size=3) |
| naked-quad | Naked Quad | 120 | Expert | NakedSubsetSolver(size=4) |
| simple-coloring | Simple Coloring | 150 | Expert | Direct |
| hidden-quad | Hidden Quad | 150 | Expert | HiddenSubsetSolver(size=4) |
| xyz-wing | XYZ-Wing | 180 | Expert | Direct |
| jellyfish | Jellyfish | 300 | Evil | FishSolver(size=4) |
| bug-plus-one | BUG+1 | 250 | Evil | Direct |
| unique-rectangle | Unique Rectangle Type 1 | 200 | Evil | Direct |
| unique-rectangle-2 | Unique Rectangle Type 2 | 220 | Evil | Direct |
| unique-rectangle-3 | Unique Rectangle Type 3 | 240 | Evil | Direct |
| unique-rectangle-4 | Unique Rectangle Type 4 | 250 | Evil | Direct |
| x-cycles | X-Cycles | 280 | Evil | Direct (DFS chain search) |
| xy-chain | XY-Chain | 300 | Evil | Direct (DFS chain search) |

### Tier Rationale

Tiers are based on SudokuWiki's human-difficulty ordering (frequency × difficulty):

- **Easy:** Trivial techniques — scan for cells/units with one candidate.
- **Medium:** Basic pattern recognition — pairs, triples, pointing pairs. Hidden pairs
  are easier than X-Wing for humans.
- **Hard:** Requires systematic row/column scanning (X-Wing, XY-Wing) or identifying
  three hidden digits in three cells (Hidden Triple). W-Wing uses bi-value cells
  connected by a strong link.
- **Expert:** Very hard to spot manually — 3-row/col fish patterns (Swordfish), four-cell
  subsets (Naked/Hidden Quad), graph coloring (Simple Coloring). XYZ-Wing extends
  XY-Wing with a three-candidate pivot.
- **Evil:** Near-impossible to spot manually — 4-row/col fish patterns (Jellyfish),
  bivalue universal grave detection (BUG+1), deadly-pattern elimination
  (Unique Rectangle Types 1–4), single-digit alternating inference chains
  (X-Cycles), and multi-cell bi-value chains (XY-Chain).

## Difficulty Mapping

| Level | Required Tier | Meaning |
|-------|---------------|--------|
| Easy | Basic only | Solvable with naked/hidden singles alone |
| Medium | Up to Intermediate | Requires at least one naked-pair, naked-triple, pointing-pair, or hidden-pair |
| Hard | Up to Advanced | Requires at least one X-Wing, XY-Wing, Hidden Triple, or W-Wing step |
| Expert | Up to Expert | Requires at least one Swordfish, Naked Quad, Simple Coloring, Hidden Quad, or XYZ-Wing step |
| Evil | Up to Evil | Requires at least one Jellyfish, BUG+1, Unique Rectangle, X-Cycles, or XY-Chain step |

### Clue Count as Secondary Constraint

Clue-count ranges are preserved as a secondary constraint alongside technique requirements.
The existing clue-count ranges define the acceptable band; technique requirements define
the minimum solving complexity.

### Architecture Support

The canonical hierarchy lives behind `solver.StrategyTierNames`, `solver.StrategySolverKeysForTier`, and `solver.StrategyTierForTechnique`. `generator/difficulty.go` uses a detached package-local view of that hierarchy so generation and classification cannot drift.

The generation plumbing:
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

**Hard:** `SolverKeys: ["x-wing", "xy-wing", "hidden-triple", "w-wing"]`.
`LowerTierSolverKeys()` returns Easy + Medium keys.

**Expert:** `SolverKeys: ["swordfish", "naked-quad", "simple-coloring", "hidden-quad", "xyz-wing"]`.
`LowerTierSolverKeys()` returns Easy + Medium + Hard keys.

**Evil:** `SolverKeys: ["jellyfish", "bug-plus-one", "unique-rectangle", "unique-rectangle-2", "unique-rectangle-3", "unique-rectangle-4", "x-cycles", "xy-chain"]`.
`LowerTierSolverKeys()` returns Easy + Medium + Hard + Expert keys.

## Classification Outcome

`solver.Classification.Outcome` reports either `solved` or `strategy-unsolved`. Difficulty remains the highest explicit technique tier reached, but a stalled trace with no applicable technique has no tier and is never promoted to Evil or backtracking. Generator target matching accepts only completed strategy solves.

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

## Calibration Boundary

Score ranges, solver weights, clue bands, generation budgets, and success-rate policy must not change from intuition alone. `.aidoc/designs/difficulty-calibration.md` defines the corpus, deterministic classification semantics, measurements, report artifacts, and decision gates required before policy changes.

The current scoring infrastructure records calibration inputs, while `solver/config.go` remains the single configuration boundary. Any proposal to add score bands or alter classification requires baseline distributions, pathological-input analysis, and before-and-after validation.

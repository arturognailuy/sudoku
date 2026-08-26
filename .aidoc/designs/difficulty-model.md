---
domain: Designs
status: Active
entry_points:
  - generator/difficulty.go
dependencies:
  - .aidoc/architecture/guidelines.md
  - .aidoc/designs/difficulty-calibration.md
---

# Strategy Rating Model

Easy through Evil are strategy grades: each label names the highest technique tier required by the canonical deterministic solver. The labels do not predict a person's experience; weighted score orders puzzles within a grade, while clue count remains generation guidance.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/architecture/guidelines.md` | Solver interface contract and layer boundaries |
| `.aidoc/designs/difficulty-calibration.md` | Measurement contract and evidence required for policy changes |
| `.aidoc/INDEX.md` | Discovery index |

## Why Strategy Grades Exist

Human difficulty depends on experience, recognition speed, interface, and play conditions that the repository does not observe. A canonical strategy trace is objective, reproducible, explainable, and under the product's control, so it is the authoritative rating contract.

The familiar Easy through Evil names remain the public tier names, but their meaning is strictly solver-relative. Human ratings or telemetry may support a separate player-difficulty model in the future; they are neither a prerequisite for strategy grades nor grounds for silently changing them.

## Current Strategy Hierarchy

Generation uses clue guidance and technique requirements in `generator/difficulty.go`:

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

The hierarchy groups the canonical solver's techniques into five stable capability tiers. Published Sudoku technique references informed the original grouping, but the product contract does not claim that the resulting order predicts human effort.

Each grade means that the canonical deterministic solver completed the puzzle and required at least one technique from that grade while using no technique above it. Changing a technique's tier changes product semantics and therefore requires a separately reviewed policy proposal with reproducible before-and-after evidence.

## Strategy Grade Mapping

| Level | Required Tier | Meaning |
|-------|---------------|--------|
| Easy | Basic only | Solvable with naked/hidden singles alone |
| Medium | Up to Intermediate | Requires at least one naked-pair, naked-triple, pointing-pair, or hidden-pair |
| Hard | Up to Advanced | Requires at least one X-Wing, XY-Wing, Hidden Triple, or W-Wing step |
| Expert | Up to Expert | Requires at least one Swordfish, Naked Quad, Simple Coloring, Hidden Quad, or XYZ-Wing step |
| Evil | Up to Evil | Requires at least one Jellyfish, BUG+1, Unique Rectangle, X-Cycles, or XY-Chain step |

### Clue Count as Generation Guidance

Clue-count ranges guide generation and preserve the current search space, but clue count does not assign or override a strategy grade. A puzzle's grade comes only from its completed canonical strategy trace.

### Architecture Support

The canonical hierarchy lives behind `solver.StrategyTierNames`, `solver.StrategySolverKeysForTier`, and `solver.StrategyTierForTechnique`. `generator/difficulty.go` uses a detached package-local view so generation and classification cannot drift. Existing Go and wire fields retain the name `Difficulty` for compatibility; their values carry strategy grades.

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

## Within-Grade Scoring

Each solver carries a `Weight` field representing its relative cost per application,
based on HoDoKu's established weights. A puzzle's total score is the sum of all
technique weights used during solving:

```
score = Σ(weight[technique] × times_used)
```

The `ScorePuzzle(store, moves)` function in `solver/scoring.go` computes the score from a list of moves. Moves from unknown techniques (e.g., backtracker) contribute zero.

Score ranks completed puzzles within the same strategy grade. Score must not demote or promote a puzzle across grades, because numeric weights are subordinate to the explicit tier hierarchy. Cross-grade score overlap is expected and is not evidence that the grade contract failed.

### Configuration

All tunable parameters — solver weights and clue-count ranges — are centralized in `solver/config.go` as one auditable strategy-rating configuration boundary.

## Calibration Boundary

Solver weights, technique tiers, clue guidance, generation budgets, and success-rate policy must not change from intuition alone. `.aidoc/designs/difficulty-calibration.md` defines the corpus, deterministic classification semantics, measurements, report artifacts, and decision gates required before policy changes.

The current scoring infrastructure records calibration inputs, while `solver/config.go` remains the single configuration boundary. Any proposal to add score bands or alter classification requires baseline distributions, pathological-input analysis, and before-and-after validation.

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
| Medium | 32–44 | Intermediate | naked-subset, pointing-pair |
| Hard | 25–31 | Advanced | x-wing |
| Expert | 22–24 | Expert | swordfish, hidden-subset |
| Evil | 17–22 | Evil | xy-wing, simple-coloring |

Each level's allowed solvers = its own SolverKeys + all solvers from lower tiers.
During generation, the generator verifies that lower-tier solvers alone cannot solve
the puzzle — ensuring it genuinely requires at least one technique from this tier.

## Target Model (Strategy-Based)

Difficulty should be defined by the **hardest technique required to solve** the puzzle.

### Strategy Tiers

| Tier | Techniques | CLI Level |
|------|-----------|----------|
| Basic | Naked singles, hidden singles | Easy |
| Intermediate | Naked pairs/triples, pointing pairs / box-line reduction | Medium |
| Advanced | X-Wing | Hard |
| Expert | Swordfish, hidden pairs/triples | Expert |
| Evil | XY-Wing, simple coloring | Evil |

### Difficulty Mapping

| Level | Required Tier | Meaning |
|-------|---------------|--------|
| Easy | Basic only | Solvable with naked/hidden singles alone |
| Medium | Up to Intermediate | Requires at least one intermediate technique |
| Hard | Up to Advanced | Requires at least one X-Wing step |
| Expert | Up to Expert | Requires at least one swordfish or hidden-subset step |
| Evil | Up to Evil | Requires at least one XY-Wing or simple coloring step |

### Clue Count as Secondary Constraint

Clue-count ranges are preserved as a secondary constraint alongside technique requirements.
A puzzle that requires a Swordfish but has 50 clues isn't fun — the strategy difficulty
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
- `Difficulty.SolverKeys` lists solver keys introduced at this tier (the only field).
- `tierRegistry` (package-level map keyed by difficulty level name, e.g. `"easy"`, `"medium"`) + `tierOrder` (ordered slice of tier names) define the tier hierarchy. Lower-tier solver keys are derived from these — there is no separate field. This is the single source of truth for tier ordering.
- `Difficulty.AllowedSolverKeys()` returns the full allowed set (lower tiers + this tier), computed from `tierRegistry`/`tierOrder`.
- `Difficulty.LowerTierSolverKeys()` returns the cumulative keys from all tiers below, derived from `tierRegistry`/`tierOrder`.
- During cell removal, the generator calls `solver.Apply()` on each allowed solver before confirming a removal.
- After generation, `requiresThisTierSolver()` checks that lower-tier solvers alone can't solve the puzzle.
- `Store` maps solver keys to implementations.

**Easy difficulty:** `SolverKeys: ["naked-single", "hidden-single"]`.
The generator produces Easy puzzles solvable using only naked and hidden singles.
As the lowest tier in `tierOrder`, `LowerTierSolverKeys()` returns nil.

**Medium difficulty:** `SolverKeys: ["naked-subset", "pointing-pair"]`.
`LowerTierSolverKeys()` returns `["naked-single", "hidden-single"]` (derived from Easy tier in registry/tierOrder).
Allowed set = all four solvers. The generator produces Medium puzzles that genuinely
require at least one intermediate technique — basic techniques alone cannot solve them.

**Hard difficulty:** `SolverKeys: ["x-wing"]`.
`LowerTierSolverKeys()` returns `["naked-single", "hidden-single", "naked-subset", "pointing-pair"]`
(derived from Easy + Medium tiers in registry/tierOrder).
Allowed set = all five solvers. The generator produces Hard puzzles that genuinely
require at least one X-Wing step — basic and intermediate techniques alone cannot solve them.

**Expert difficulty:** `SolverKeys: ["swordfish", "hidden-subset"]`.
`LowerTierSolverKeys()` returns `["naked-single", "hidden-single", "naked-subset", "pointing-pair", "x-wing"]`
(derived from Easy + Medium + Hard tiers in registry/tierOrder).
Allowed set = all seven solvers. The generator produces Expert puzzles that genuinely
require at least one expert technique — lower-tier techniques alone cannot solve them.

**Evil difficulty:** `SolverKeys: ["xy-wing", "simple-coloring"]`.
`LowerTierSolverKeys()` returns `["naked-single", "hidden-single", "naked-subset", "pointing-pair", "x-wing", "swordfish", "hidden-subset"]`
(derived from Easy + Medium + Hard + Expert tiers in registry/tierOrder).
Allowed set = all nine solvers. The generator produces Evil puzzles that genuinely
require at least one evil technique — lower-tier techniques alone cannot solve them.

All five difficulty levels are now fully strategy-based with technique requirements.

## Scoring System

Each solver carries a `Weight` field representing its difficulty cost per application,
based on HoDoKu's established weights. A puzzle's total difficulty score is the sum
of all technique weights used during solving:

```
score = Σ(weight[technique] × times_used)
```

Scoring is purely additive infrastructure — it does not change any existing behavior.
The `ScorePuzzle(store, moves)` function in `solver/scoring.go` computes the score
from a list of moves. Moves from unknown techniques (e.g., backtracker) contribute zero.

### Current Solver Weights

| Solver | Key | Weight | Rationale |
|--------|-----|--------|----------|
| Naked Single | naked-single | 4 | Trivial — scan cells |
| Hidden Single | hidden-single | 14 | Easy — scan units |
| Pointing Pairs / Box-Line | pointing-pair | 50 | Easy — box/line intersection |
| Naked Pairs/Triples | naked-subset | 70 | Moderate — combined pair+triple solver |
| Hidden Pairs/Triples/Quads | hidden-subset | 100 | Hard — combined subset solver |
| X-Wing | x-wing | 140 | Hard — row/column pattern scanning |
| Swordfish | swordfish | 150 | Very Hard — 3×3 fish pattern |
| Simple Coloring | simple-coloring | 150 | Hard — graph 2-coloring |
| XY-Wing | xy-wing | 160 | Hard — pivot + two pincers |
| Backtracker | default | 0 | Not scored — fallback solver |

Weights for combined solvers (naked-subset, hidden-subset) use representative midpoints.
When these are split into per-size solvers (Phase 3.5), each variant will get its own
weight matching its specific human difficulty.

### Future: Score-Based Difficulty Ranges

`MinScore` and `MaxScore` fields on `Difficulty` will be added in Phase 4,
when the puzzle database provides enough data to calibrate score ranges.
The scoring infrastructure is ready — each new solver just sets its weight
in the constructor.

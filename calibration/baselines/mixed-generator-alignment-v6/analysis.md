# Generator Alignment and Strategy-Coverage Analysis

This evidence-only analysis interprets `report.json` and `observations.jsonl` for the 101-record `mixed-generator-alignment-v6` corpus. It changes no strategy tiers, weights, clue guidance, generation budgets, fallback behavior, or acceptance policy.

## Selection and Reproducibility

The v6 manifest preserves all 76 v5 records and appends five sequential `GenerateBestEffort` calls per requested tier. Calls used `MaxRounds=4`, `MaxDurationMs=2000`, and default options; no result was rejected or replaced after observation. All 101 records reproduced exactly under double classification.

The generated stratum now contains ten calls per requested tier. Random seeds remain unavailable at the public generator boundary, so the immutable puzzles, provenance, call order, elapsed time, rounds, and classifications are the reproducible evidence.

## Target Alignment

| Requested grade | Solved target hits | Other completed grades | Strategy-unsolved |
|---|---:|---|---:|
| Easy | 10/10 | none | 0 |
| Medium | 0/10 | Easy 10 | 0 |
| Hard | 1/10 | Easy 5, Medium 3 | 1 Evil-trace stall |
| Expert | 0/10 | Easy 3, Medium 2, Hard 1, Evil 3 | 1 Evil-trace stall |
| Evil | 4/10 | Easy 2, Hard 1, Expert 1, Medium 1 | 1 Evil-trace stall |

Easy is aligned in this sample. Medium consistently collapses to Easy. Hard and Expert are broadly dispersed rather than merely one adjacent grade away. Evil reaches its requested completed grade in four calls, but one additional call reaches an Evil technique before stalling and therefore remains `strategy-unsolved`.

Clue guidance does not explain the misses as a grade boundary. Medium misses span 32–44 clues, Hard misses span 25–31 clues, Expert misses span 24–26 clues, and Evil misses span 24–26 clues. The canonical grade remains the highest required technique in a completed trace.

## Trace and Budget Signals

Maximum-technique signatures reinforce the alignment result. All ten Medium calls finish with `naked-single`; the request does not yield a Medium technique. Hard and Expert calls range from single techniques through Evil techniques, while Evil calls span Easy through Evil techniques.

| Requested grade | Maximum-technique signature across ten calls |
|---|---|
| Easy | `naked-single` 10 |
| Medium | `naked-single` 10 |
| Hard | `naked-single` 4; `hidden-single`, `naked-pair`, `naked-triple`, `pointing-pair`, `x-wing`, and `x-cycles` 1 each |
| Expert | `naked-single` 3; `unique-rectangle` 2; `naked-pair`, `pointing-pair`, `unique-rectangle-2`, `xy-wing`, and `x-cycles` 1 each |
| Evil | `x-cycles` 2; `hidden-single`, `naked-single`, `naked-pair`, `simple-coloring`, `unique-rectangle`, `unique-rectangle-4`, `x-wing`, and `xy-chain` 1 each |

The nominal two-second duration is checked between full generation rounds rather than during a round. Calls above 2,000 ms occurred in 0/10 Easy, 1/10 Medium, 1/10 Hard, 10/10 Expert, and 8/10 Evil samples. Observed maxima were 7 ms, 2,034 ms, 2,816 ms, 15,988 ms, and 20,289 ms respectively. Any budget proposal must therefore separate a soft between-round budget from a hard wall-clock deadline.

## Strategy-Coverage Signals

The full corpus contains 17 strategy-unsolved records: 10/31 external, 3/50 generated, 1/15 imported, and 3/5 pathological. The three new generated stalls are:

| Record | Requested | Moves before stall | Highest technique reached |
|---|---|---:|---|
| `generated-hard-7` | Hard | 21 | `x-cycles` (Evil) |
| `generated-expert-6` | Expert | 19 | `unique-rectangle-2` (Evil) |
| `generated-evil-8` | Evil | 22 | `x-cycles` (Evil) |

External stalls remain concentrated in `top95` and `hardest`: four and six records respectively. Their traces range from one move to 25 moves and reach Easy, Expert, or Evil techniques before no registered strategy can progress. The imported stall reaches `unique-rectangle-2` after 11 moves. Empty and three-given pathological boards make no strategy move, while the dedicated stall fixture stops after 48 naked singles.

The source pattern distinguishes two concerns. Generator alignment is poor even when classification completes, especially for Medium through Expert. Strategy coverage is independently incomplete because diverse external/imported puzzles and three newly generated high-tier puzzles stall after canonical techniques have made progress. A stalled trace identifies an inventory boundary but does not prove which missing technique should be added.

## Decision Boundary

The evidence supports a separately reviewed generator-policy proposal, not an immediate policy change. That proposal should choose whether requested grades are strict targets or best-effort preferences, define soft versus hard time semantics, and compare candidate behavior on the preserved exploratory and held-out splits.

Strategy inventory changes should remain separate from generator-budget changes. Representative stalls need technique-level diagnosis and regression fixtures before any solver is added or moved between tiers. Database behavior remains after calibration policy, as required by the roadmap.

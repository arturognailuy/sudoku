---
domain: Designs
status: Active
entry_points:
  - calibration/runner.go
  - cmd/calibrate.go
  - solver/classify.go
  - generator/generator.go
dependencies:
  - .aidoc/designs/difficulty-model.md
  - .aidoc/designs/roadmap.md
---

# Strategy Rating Calibration

Calibration tests whether the canonical strategy grades are reproducible, internally coherent, and useful for generation. Human ratings are optional evidence for a separate player-difficulty model, not a prerequisite or validation gate for the strategy rating contract.

## Related Docs

| Document | Relationship |
|----------|--------------|
| `.aidoc/designs/difficulty-model.md` | Strategy grades, weights, clue guidance, and configuration boundary |
| `.aidoc/designs/roadmap.md` | Sequencing after baseline stabilization |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box behavior catalog for later policy changes |

## Why Calibration Exists

The current model combines technique tiers, accumulated HoDoKu-derived weights, and clue guidance. Calibration must distinguish canonical grade behavior from accidental effects of clue count, solver ordering, repeated moves, or generator search budgets.

Calibration reports support policy decisions; calibration tooling does not silently make them. Technique-tier changes, weight changes, clue guidance, generation budgets, and fallback semantics require a separate proposal with before-and-after evidence.

The strategy rating contract is intentionally narrower than human difficulty. Completion time, mistakes, hints, subjective ratings, and player experience may later inform a separately named empirical layer, but absence of that data does not block or weaken a canonical strategy grade.

## Measurement Preconditions

`solver.ClassifyPuzzle` applies strategies in the canonical registration order defined by the shared tier hierarchy. `solver.StrategyTierForTechnique` selects the highest required tier independently of numeric weights, so overlapping weight ranges cannot demote an Expert technique below a Hard technique.

Classification preserves total weighted score and the ordered move trace as independent measurements. `solver.Classification.Outcome` distinguishes a completed strategy solve from `strategy-unsolved`; a stalled trace is not promoted to Evil or represented as backtracking. A solved input with no moves belongs to the lowest tier, while a stalled input with no applicable technique has no assigned tier.

Identical input, store configuration, and code version must produce identical outcome, score, maximum technique, and trace. The baseline runner must persist these fields and verify repeated exact equality before treating observations as authoritative.

## Corpus Contract

The corpus combines traceable external puzzles, puzzles generated for every requested tier, provenance-permitted database/import samples, and deliberately pathological inputs. External records may preserve published labels for comparison, but independently rated human data is optional. Pathological groups include minimal-clue boards, boundary clue counts, unusually long traces, and uniquely solvable puzzles that the strategy inventory cannot finish.

Each record carries a normalized 81-cell puzzle, a content hash, source category, source identifier or citation, original rating when available, license or redistribution constraint, and collection method. Normalization and hash-based deduplication occur before sampling. Reports aggregate restricted sources without republishing their puzzle strings.

Generated samples record the requested tier, generator configuration, attempt or round number, elapsed duration, resulting clue count, and classification outcome. Generator-only evidence cannot validate the assumptions that created the generator, so conclusions must be shown separately for generated, external, imported, and pathological groups.

Corpus versions are immutable manifests. Version 2 requires every member to declare its normalized content hash, source category (`external`, `generated`, `imported`, or `pathological`), source identifier, license, redistribution constraint, collection method, and split (`exploratory` or `held-out`). Original ratings preserve both their source system and source label. Generated records additionally require requested difficulty, a generator-configuration description, attempt number, elapsed time, clue count, and classification outcome.

`sudoku calibrate prepare --input <candidate.json> --output <manifest.json>` is the ingestion boundary. It removes whitespace, converts zero notation to canonical dots, computes SHA-256 over the normalized 81 cells, rejects duplicate hashes and incomplete provenance, and refuses to overwrite an existing manifest. Candidate order is preserved so sampling decisions remain reviewable. Exploratory data informs candidate policy; a held-out validation subset measures the final candidate once and prevents thresholds from being tuned to every observed puzzle.

## Metrics and Methods

The baseline report answers distinct questions with distinct measurements:

| Question | Evidence |
|----------|----------|
| Reproducibility | Repeated classifications with exact equality of outcome, score, maximum tier, move count, and trace digest |
| Grade integrity | Technique usage and maximum required tier by assigned grade and source group |
| Within-grade ordering | Score and move-count distributions within each assigned grade; cross-grade overlap is descriptive, not a boundary |
| Optional external comparison | Source-specific agreement matrices for published labels; disagreements do not invalidate strategy grades |
| Generation quality | Target-hit rate by round and elapsed budget, mismatch matrix, p50/p95 latency, rounds, clue count, and failure rate |
| Strategy coverage | Strategy-unsolved rate by source group with representative trace-stall cases |

Confidence intervals accompany rates and percentiles where sample size permits. Sample counts and missing-data rules appear beside every aggregate. External rating systems are analyzed separately and never normalized into the canonical grade without an explicit new product decision, because equally named tiers need not mean the same thing.

## Reproducibility and Report Artifacts

Every run records the corpus manifest hash, repository commit, Go version, operating system and architecture, solver configuration digest, generator options, random seed where the public boundary supports one, start time, and command invocation. No user telemetry or network service is required to reproduce local analysis.

The local `sudoku calibrate` boundary accepts an immutable, ordered version 2 JSON manifest and an output directory. Direct measurement accepts only canonical dot notation and verifies every declared content hash, preventing the preparation and measurement boundaries from disagreeing. `calibration.Run` binds checkpoints and observations to the exact manifest SHA-256, rejects changed manifests and non-prefix logs, classifies every puzzle twice for exact reproducibility, appends one JSON Lines observation only after agreement, and resumes from the durable log when a checkpoint is missing or lagging.

The runner emits `observations.jsonl`, `checkpoint.json`, `report.json`, and `report.md`. Raw observations are append-only run artifacts; checkpoints and deterministic reports are derived state that can be rebuilt without changing observations. Reports stratify outcomes and Wilson intervals by source and split, summarize nearest-rank score/move/clue distributions by assigned tier, show neighboring score-range overlap, preserve each external rating system as a separate agreement matrix, and report generated target mismatches, hit rates, latency, and rounds.

## Current Evidence

`calibration/testdata/mixed-pilot-v2.json` preserves the immutable 30-record pilot across external, generated, imported, and pathological sources. `calibration/testdata/mixed-external-expansion-v3.json` adds source-order records from the same pinned MIT-licensed `norvig/pytudes` groups without outcome-based selection, increasing the external stratum from 9 to 31 puzzles. `calibration/testdata/mixed-generated-expansion-v4.json` then appends three sequential generation calls per requested tier without outcome-based rejection or replacement, increasing the generated stratum from 10 to 25 puzzles. `calibration/testdata/mixed-imported-expansion-v5.json` appends all nine remaining hash-unique named integration fixtures in source order, increasing the imported stratum from 6 to 15 puzzles. Each manifest has a matching directory under `calibration/baselines/` with append-only observations, deterministic reports, and exact run metadata.

The external expansion confirms that score is not a cross-grade boundary: all four neighboring assigned grades have overlapping observed score ranges, and 10 of 31 external puzzles are strategy-unsolved. The `easy50`, `top95`, and `hardest` groups remain separate because their labels do not define a shared rating scale.

The generated expansion provides five observations per requested tier. Easy reached its target in 5/5 calls, Evil in 2/5, and Medium, Hard, and Expert in 0/5; the latter tiers often classified below their request, while Expert calls also exceeded the nominal duration because the budget is checked only between full generation rounds. These measurements expose generator behavior but remain too small to justify changing budgets or fallback semantics.

The imported expansion classifies 76 total records reproducibly. Fourteen of 15 imported fixtures solve, one is strategy-unsolved, and two fixtures documented as Expert classify as Hard under the canonical full solver order. This shows that constrained integration-fixture labels are not automatically classifier labels.

The published evidence remains a measurement baseline, not permission to tune policy. Pathological data remain small, the external corpus comes from one source family, and generated target-hit intervals remain wide. The next evidence priority is generator alignment and strategy-coverage analysis under the approved strategy contract; human-rating data is not required.

## Rating Contract and Remaining Decision Gates

The approved rating contract fixes these invariants:

1. Easy through Evil mean the highest strategy tier required by the canonical deterministic solver, not predicted player experience.
2. Score orders completed puzzles within a grade and never overrides the explicit grade hierarchy.
3. Clue ranges guide generation but do not assign a grade.
4. `strategy-unsolved` remains a separate outcome and is never relabeled Evil.
5. Human observations, if collected later, belong to a separate empirical player-difficulty layer.

Future calibration proposals still require review of technique-tier or weight changes, round and time budgets, storage and fallback treatment of strategy-unsolved puzzles, and target-hit, reproducibility, latency, and coverage acceptance thresholds. Every proposal must state rejected alternatives and compare current and candidate behavior on exploratory and held-out data.

## Delivery Sequence

1. Preserve deterministic classification semantics with regression tests.
2. Maintain immutable corpus manifests, append-only observations, and deterministic reports without tuning product policy.
3. Expand only pilot strata whose target-hit, coverage, or rare-failure estimates remain unstable.
4. Propose any policy change with before-and-after comparisons under the fixed strategy contract.
5. Apply approved policy with unit coverage and applicable built-binary E2E scenarios.

Primary code boundaries are `calibration.Run`, `cmd.newCalibrateCommand`, `solver.ClassifyPuzzle`, `solver.StrategyTierForTechnique`, `solver.ScorePuzzle`, `generator.Difficulty`, `generator.GenerateBestEffort`, and `solver/config.go`.

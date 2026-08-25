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

# Difficulty Calibration

Difficulty calibration turns solver traces and generation attempts into reproducible evidence for product policy. The first calibrated contract is solver-relative strategy difficulty; external human ratings validate that approximation but do not redefine it without evidence.

## Related Docs

| Document | Relationship |
|----------|--------------|
| `.aidoc/designs/difficulty-model.md` | Current tiers, weights, clue constraints, and configuration boundary |
| `.aidoc/designs/roadmap.md` | Sequencing after baseline stabilization |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box behavior catalog for later policy changes |

## Why Calibration Exists

The current model combines technique tiers, accumulated HoDoKu-derived weights, and clue bands, but those inputs have not been measured together on a controlled corpus. Calibration must distinguish actual tier separation from accidental effects of clue count, solver ordering, repeated moves, or generator search budgets.

Calibration reports support policy decisions; calibration tooling does not silently make them. Weight changes, score bands, clue bands, generation budgets, and fallback semantics require a separate proposal with before-and-after evidence.

## Measurement Preconditions

`solver.ClassifyPuzzle` applies strategies in the canonical registration order defined by the shared tier hierarchy. `solver.StrategyTierForTechnique` selects the highest required tier independently of numeric weights, so overlapping weight ranges cannot demote an Expert technique below a Hard technique.

Classification preserves total weighted score and the ordered move trace as independent measurements. `solver.Classification.Outcome` distinguishes a completed strategy solve from `strategy-unsolved`; a stalled trace is not promoted to Evil or represented as backtracking. A solved input with no moves belongs to the lowest tier, while a stalled input with no applicable technique has no assigned tier.

Identical input, store configuration, and code version must produce identical outcome, score, maximum technique, and trace. The baseline runner must persist these fields and verify repeated exact equality before treating observations as authoritative.

## Corpus Contract

The corpus combines independently rated puzzles, puzzles generated for every requested tier, provenance-permitted database/import samples, and deliberately pathological inputs. Pathological groups include minimal-clue boards, boundary clue counts, unusually long traces, and uniquely solvable puzzles that the strategy inventory cannot finish.

Each record carries a normalized 81-cell puzzle, a content hash, source category, source identifier or citation, original rating when available, license or redistribution constraint, and collection method. Normalization and hash-based deduplication occur before sampling. Reports aggregate restricted sources without republishing their puzzle strings.

Generated samples record the requested tier, generator configuration, attempt or round number, elapsed duration, resulting clue count, and classification outcome. Generator-only evidence cannot validate the assumptions that created the generator, so conclusions must be shown separately for generated, external, imported, and pathological groups.

Corpus versions are immutable manifests. Version 2 requires every member to declare its normalized content hash, source category (`external`, `generated`, `imported`, or `pathological`), source identifier, license, redistribution constraint, collection method, and split (`exploratory` or `held-out`). Original ratings preserve both their source system and source label. Generated records additionally require requested difficulty, a generator-configuration description, attempt number, elapsed time, clue count, and classification outcome.

`sudoku calibrate prepare --input <candidate.json> --output <manifest.json>` is the ingestion boundary. It removes whitespace, converts zero notation to canonical dots, computes SHA-256 over the normalized 81 cells, rejects duplicate hashes and incomplete provenance, and refuses to overwrite an existing manifest. Candidate order is preserved so sampling decisions remain reviewable. Exploratory data informs candidate policy; a held-out validation subset measures the final candidate once and prevents thresholds from being tuned to every observed puzzle.

## Metrics and Methods

The baseline report answers distinct questions with distinct measurements:

| Question | Evidence |
|----------|----------|
| Reproducibility | Repeated classifications with exact equality of outcome, score, maximum tier, move count, and trace digest |
| Tier meaning | Technique usage, move-count, score, and clue distributions by assigned tier and source group |
| Tier separation | Neighboring-tier overlap, monotonic trends, and representative inversions rather than averages alone |
| External validity | Agreement matrix plus ordinal association against independently rated puzzles; disagreements remain inspectable cases |
| Generation quality | Target-hit rate by round and elapsed budget, mismatch matrix, p50/p95 latency, rounds, clue count, and failure rate |
| Strategy coverage | Strategy-unsolved rate by source group with representative trace-stall cases |

Confidence intervals accompany rates and percentiles where sample size permits. Sample counts and missing-data rules appear beside every aggregate. External rating systems are analyzed separately before any justified normalization because equally named tiers need not mean the same thing.

## Reproducibility and Report Artifacts

Every run records the corpus manifest hash, repository commit, Go version, operating system and architecture, solver configuration digest, generator options, random seed where the public boundary supports one, start time, and command invocation. No user telemetry or network service is required to reproduce local analysis.

The local `sudoku calibrate` boundary accepts an immutable, ordered version 2 JSON manifest and an output directory. Direct measurement accepts only canonical dot notation and verifies every declared content hash, preventing the preparation and measurement boundaries from disagreeing. `calibration.Run` binds checkpoints and observations to the exact manifest SHA-256, rejects changed manifests and non-prefix logs, classifies every puzzle twice for exact reproducibility, appends one JSON Lines observation only after agreement, and resumes from the durable log when a checkpoint is missing or lagging.

The runner emits `observations.jsonl`, `checkpoint.json`, `report.json`, and `report.md`. Raw observations are append-only run artifacts; checkpoints and deterministic reports are derived state that can be rebuilt without changing observations. Reports stratify outcomes and Wilson intervals by source and split, summarize nearest-rank score/move/clue distributions by assigned tier, show neighboring score-range overlap, preserve each external rating system as a separate agreement matrix, and report generated target mismatches, hit rates, latency, and rounds.

## Current Evidence

`calibration/testdata/mixed-pilot-v2.json` preserves the immutable 30-record pilot across external, generated, imported, and pathological sources. `calibration/testdata/mixed-external-expansion-v3.json` adds source-order records from the same pinned MIT-licensed `norvig/pytudes` groups without outcome-based selection, increasing the external stratum from 9 to 31 puzzles. Each manifest has a matching directory under `calibration/baselines/` with append-only observations, deterministic reports, and exact run metadata.

The external expansion makes the pilot's apparent score separation less credible: all four neighboring assigned tiers now have overlapping observed score ranges, and 10 of 31 external puzzles are strategy-unsolved. The `easy50`, `top95`, and `hardest` groups remain separate because their labels do not define a shared human-difficulty scale.

The published evidence remains a measurement baseline, not calibration policy. Generated records still number only two per requested tier, imported and pathological strata remain small, and the external corpus comes from one source family. The evidence supports targeted generated/imported expansion and an independent human-rating source before any proposal changes weights, thresholds, clue bands, budgets, or fallback behavior.

## Decision Gates

The baseline report must let a reviewer decide:

1. whether labels promise solver-relative difficulty only or also claim an externally validated human approximation;
2. whether score orders puzzles only within a tier or contributes to cross-tier boundaries;
3. whether clue bands remain constraints, become generation guidance, or change from measured evidence;
4. whether round and time budgets differ by requested tier;
5. how strategy-unsolved puzzles appear in classification, storage, and fallback behavior;
6. which stability, separation, hit-rate, and latency thresholds define an acceptable calibrated model.

No threshold is fixed in this design because the baseline exists to supply the evidence. A calibration proposal must state rejected alternatives and compare current and candidate behavior on both exploratory and held-out data.

## Delivery Sequence

1. Preserve deterministic classification semantics with regression tests.
2. Maintain immutable corpus manifests, append-only observations, and deterministic reports without tuning product policy.
3. Expand only pilot strata whose intervals, overlap, or rare-failure estimates remain unstable.
4. Review policy choices and propose calibration with before-and-after comparisons.
5. Apply approved policy with unit coverage and applicable built-binary E2E scenarios.

Primary code boundaries are `calibration.Run`, `cmd.newCalibrateCommand`, `solver.ClassifyPuzzle`, `solver.StrategyTierForTechnique`, `solver.ScorePuzzle`, `generator.Difficulty`, `generator.GenerateBestEffort`, and `solver/config.go`.

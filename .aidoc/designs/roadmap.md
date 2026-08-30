---
domain: Designs
status: Active
entry_points:
  - .github/workflows/ci.yml
  - scripts/e2e_api.py
  - scripts/e2e_cli.py
  - scripts/e2e_tui.py
  - scripts/coverage_report.py
  - solver/config.go
dependencies:
  - .aidoc/designs/e2e-test-scenarios.md
  - .aidoc/designs/difficulty-model.md
  - .aidoc/designs/difficulty-calibration.md
  - .aidoc/designs/future-directions.md
  - .aidoc/designs/database-puzzle-selection.md
  - .aidoc/designs/database-play-statistics.md
  - .aidoc/designs/database-concurrency.md
---

# Stabilization Roadmap

Sudoku's feature baseline, calibration, and stabilization gates are complete: the Go repository owns the engine, CLI, TUI, persistence, recovery, client-neutral HTTP API, bounded generation semantics, and independent CI lanes. Played-state selection, acquisition/completion statistics, and explicit history reset are implemented database capabilities.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/e2e-test-scenarios.md` | Canonical black-box behavior catalog and current automation pointers |
| `.aidoc/designs/difficulty-model.md` | Strategy-grade invariants and the calibration boundary |
| `.aidoc/designs/difficulty-calibration.md` | Strategy measurement methodology, report artifacts, and review decisions |
| `.aidoc/designs/future-directions.md` | Deliberately non-priority product and production directions |
| `.aidoc/designs/web-api.md` | Current HTTP contract, security boundary, and deployment defaults |
| `.aidoc/designs/database-puzzle-selection.md` | Current played-state selection and migration boundary |
| `.aidoc/designs/database-play-statistics.md` | Current completion counters, statistics, and explicit history reset |
| `.aidoc/designs/database-concurrency.md` | Next bounded database increment: connection policy and deterministic mixed-workload stress |

## Why Stabilization Remains the Gate

The current product spans terminal interaction, durable local state, generation, SQLite, and a network API. Independent CI and black-box lanes keep changes to any one boundary from weakening the established baseline.

The stabilization policy prioritizes repeatable evidence over speculative behavior. Reliable CI and deterministic black-box tests make calibration results meaningful and reduce the risk of changing database selection or import policy.

## Completed Baseline

All previously planned phases 1–10 are complete. The repository provides the following maintained product boundaries rather than open roadmap items:

- reusable game actions, detached snapshots, notes, unified history, and versioned sessions;
- line-oriented CLI and opt-in accessible TUI with automatic candidates;
- private autosave and crash recovery;
- bounded puzzle generation, import, SQLite storage, and fallback;
- contract-first OpenAPI 3.1.1 HTTP API with recovery, revisions, authentication, and CORS controls.

A TypeScript client and browser UI are explicitly outside this repository. A separate project will own their design, implementation, release cycle, and tests.

## Current Evidence and Remaining Work

### 1. Maintain CI and Black-Box E2E

Pull-request CI separates unit tests, race detection, vet, lint, API contract checks, API E2E, and TUI E2E into independent jobs. Independent jobs keep a slow boundary from hiding fast failures and allow every gate to report its own timeout and diagnostics.

Storage and command-wiring tests use fixed classified puzzles through the `cmd.batchGenerateWith` generation seam. Real randomized generation remains covered in `generator`, while `cmd` tests prove reporting and SQLite composition without waiting for a target difficulty. Solver fallback fixtures use `solver.Backtracker.SolveDeterministic`; randomized `solver.Backtracker.Solve` remains available for diverse full-board generation without making race-test duration depend on a lucky search path. These boundaries keep `go test -race -count=1 ./...` viable as a mandatory gate without weakening generation or fallback coverage.

The API, TUI, and line-CLI harnesses build and execute the real binary with isolated temporary state. The line-CLI lane covers parsing, gameplay/history, durable sessions, import normalization and deduplication, bounded generation, and SQLite-visible composition. The public `--from-db` boundary deterministically covers exact-grade acquisition, migration, and balanced reuse; generated-fallback accounting uses focused package coverage.

### 2. Maintain Boundary Unit and Integration Coverage

- `webapi` tests cover malformed input, lifecycle and persistence failures, exact authentication and CORS boundaries, revision conflicts, concurrent sessions, and process-lock exclusion.
- `cmd` tests cover deterministic generation/storage composition, session restoration and source rejection, API startup-policy validation, and process-lock ownership. CLI dispatch and Cobra workflows remain in built-binary E2E instead of being reimplemented in test-only controllers.
- The unit job records a cross-package Go coverage profile and publishes a package summary through `scripts/coverage_report.py`. Coverage is review evidence rather than a pass/fail threshold.
- Review prioritizes meaningful branches by risk: `webapi`, `cmd`, `recovery`, and `sessionfile` failure/lifecycle paths; `game` state invariants; and generator/solver correctness. Low-risk wrappers and generated boundary code do not justify artificial tests solely to raise a percentage.
- Require every fixed defect to gain a regression test at the narrowest layer that proves the behavior.

### 3. Preserve the Completed Calibration Contract

Calibration runs from the stable CI baseline with deterministic classifier semantics and versioned mixed corpora. Easy through Evil are canonical strategy grades rather than predictions of player experience; score orders puzzles within a grade, clue count guides generation, and strategy-unsolved remains separate. `.aidoc/designs/difficulty-calibration.md` owns the current evidence, corpus contract, reproducibility metadata, measurements, and remaining decision gates.

Calibration output remains local and telemetry-free. The 101-record corpus separates target-alignment failures from strategy-inventory stalls. Batch generation remains best-effort and stores each completed puzzle under its actual grade; per-puzzle wall-clock budgets are hard deadlines. Interactive play first uses an exact requested-grade result or database puzzle, then explicitly reports any actual-grade fallback. Technique-inventory changes remain separate; human data may support a later empirical player-difficulty layer but is not a prerequisite for strategy calibration.

### 4. Maintain Played-State Selection

Calibration and stabilization now provide the stable strategy-grade boundary required for database selection. `.aidoc/designs/database-puzzle-selection.md` defines the current exact-grade acquisition, never-played-first selection, balanced reuse, in-place additive migration, and deterministic public `--from-db` boundary.

The black-box scenarios in `.aidoc/designs/e2e-database-scenarios.md` and focused transaction/migration regression tests protect this boundary. `.aidoc/designs/database-play-statistics.md` keeps acquisition metrics separate from completion count/latest-completion fields, exposes both through `sudoku db stats`, and resets only explicitly selected history. The database does not infer abandonment or elapsed duration.

The next bounded increment is `.aidoc/designs/database-concurrency.md`: make SQLite connection policy apply predictably across pooled connections and prove deterministic mixed access from goroutines, independent handles, and built-binary processes. It preserves the schema and commands. Minimum-clue policy and measured large-import behavior remain separate later decisions.

## Maintained Stabilization Gates

- Pull-request CI runs race detection, API E2E, TUI PTY E2E, and automated line-CLI/command E2E.
- Boundary failures and lifecycles in `webapi` and `cmd` have focused regression coverage without duplicating black-box command tests.
- CI publishes reproducible package coverage for risk-based review without a global threshold.
- All black-box lanes run against built artifacts with isolated state and deterministic fixtures.
- Calibration evidence and policy proposals begin from a green, stable baseline.

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
---

# Stabilization Roadmap

Sudoku's feature baseline is complete: the Go repository owns the engine, CLI, TUI, persistence, recovery, and client-neutral HTTP API. The next work cycle stabilizes that baseline through stronger CI, unit coverage, and automated black-box testing before difficulty calibration or database behavior changes.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/e2e-test-scenarios.md` | Canonical black-box behavior catalog and current automation pointers |
| `.aidoc/designs/difficulty-model.md` | Difficulty invariants and the calibration boundary |
| `.aidoc/designs/difficulty-calibration.md` | Measurement methodology, report artifacts, and review decisions |
| `.aidoc/designs/future-directions.md` | Deliberately non-priority product and production directions |
| `.aidoc/designs/web-api.md` | Current HTTP contract, security boundary, and deployment defaults |

## Why Stabilization Comes Next

The current product spans terminal interaction, durable local state, generation, SQLite, and a network API. Additional features would increase the number of failure modes before the existing boundaries have consistent regression protection.

The stabilization cycle prioritizes repeatable evidence over new behavior. Reliable CI and black-box tests make later calibration results meaningful and reduce the risk of changing database selection or import policy.

## Completed Baseline

All previously planned phases 1–10 are complete. The repository provides the following maintained product boundaries rather than open roadmap items:

- reusable game actions, detached snapshots, notes, unified history, and versioned sessions;
- line-oriented CLI and opt-in accessible TUI with automatic candidates;
- private autosave and crash recovery;
- bounded puzzle generation, import, SQLite storage, and fallback;
- contract-first OpenAPI 3.1.1 HTTP API with recovery, revisions, authentication, and CORS controls.

A TypeScript client and browser UI are explicitly outside this repository. A separate project will own their design, implementation, release cycle, and tests.

## Next Work Cycle

### 1. Strengthen CI and Black-Box E2E

Pull-request CI separates unit tests, race detection, vet, lint, API contract checks, API E2E, and TUI E2E into independent jobs. Independent jobs keep a slow boundary from hiding fast failures and allow every gate to report its own timeout and diagnostics.

Storage and command-wiring tests use fixed classified puzzles through the `cmd.batchGenerateWith` generation seam. Real randomized generation remains covered in `generator`, while `cmd` tests prove reporting and SQLite composition without waiting for a target difficulty. Solver fallback fixtures use `solver.Backtracker.SolveDeterministic`; randomized `solver.Backtracker.Solve` remains available for diverse full-board generation without making race-test duration depend on a lucky search path. These boundaries keep `go test -race -count=1 ./...` viable as a mandatory gate without weakening generation or fallback coverage.

The API, TUI, and line-CLI harnesses build and execute the real binary with isolated temporary state. The line-CLI lane covers parsing, gameplay/history, durable sessions, import normalization and deduplication, bounded generation, and SQLite-visible composition. Probabilistic database fallback remains excluded until it has a deterministic public behavior or a package-level seam.

### 2. Raise Boundary Unit and Integration Coverage

- `webapi` tests cover malformed input, lifecycle and persistence failures, exact authentication and CORS boundaries, revision conflicts, concurrent sessions, and process-lock exclusion.
- `cmd` tests cover deterministic generation/storage composition, session restoration and source rejection, API startup-policy validation, and process-lock ownership. CLI dispatch and Cobra workflows remain in built-binary E2E instead of being reimplemented in test-only controllers.
- The unit job records a cross-package Go coverage profile and publishes a package summary through `scripts/coverage_report.py`. Coverage is review evidence rather than a pass/fail threshold.
- Review prioritizes meaningful branches by risk: `webapi`, `cmd`, `recovery`, and `sessionfile` failure/lifecycle paths; `game` state invariants; and generator/solver correctness. Low-risk wrappers and generated boundary code do not justify artificial tests solely to raise a percentage.
- Require every fixed defect to gain a regression test at the narrowest layer that proves the behavior.

### 3. Calibrate Difficulty with Data

Calibration starts from the stable CI baseline with deterministic classifier semantics, then measures a versioned mixed corpus before changing `solver/config.go`. `.aidoc/designs/difficulty-calibration.md` defines the corpus groups, reproducibility metadata, score and tier distributions, external-rating validation, generation hit-rate and latency measurements, pathological inputs, and strategy-unsolved reporting.

Calibration output remains local and telemetry-free. The baseline report informs separately reviewed decisions about label meaning, score use, clue bands, tier-specific budgets, unsolved states, and acceptance thresholds before any product-policy change.

### 4. Complete Deferred Database Behaviors

Database behavior changes follow calibration so selection policy is based on measured difficulty rather than unstable assumptions. Candidate work includes played-state filtering, large-import progress verification, minimum-clue handling, and concurrent SQLite stress coverage.

Each database behavior requires explicit semantics, migration impact analysis, black-box scenarios, and regression tests before implementation.

## Exit Criteria for Stabilization

- Pull-request CI runs race detection, API E2E, TUI PTY E2E, and automated line-CLI/command E2E.
- Boundary failures and lifecycles in `webapi` and `cmd` have focused regression coverage without duplicating black-box command tests.
- CI publishes reproducible package coverage for risk-based review without a global threshold.
- All black-box lanes run against built artifacts with isolated state and deterministic fixtures.
- The difficulty calibration plan can begin from a green, stable baseline.

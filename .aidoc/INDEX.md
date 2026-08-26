---
domain: Conventions
status: Active
entry_points: []
dependencies: []
---

# .aidoc/INDEX.md — Discovery Index

This index provides reading chains for common starting points and a complete document map.

## Reading Chains

### Understanding the Architecture
1. `AGENT.md` — global rules, repo layout
2. `.aidoc/architecture/guidelines.md` — design constraints, layer boundaries, solver interface contract
3. `core/candidates.go` — `CandidateSet` bitfield type
4. `core/board.go` — `Board` struct with compute-on-fly `Candidates()` method
5. `solver/solver.go` — `Solver`, `StrategySolver`, `CompleteSolver` interfaces and `Base`
6. `solver/move.go` — `Move` struct (cell + technique + reason)
7. `solver/store.go` — solver registry with typed access
8. `game/game.go` — private session state and compatibility adapters
9. `game/contract.go` — typed actions, detached snapshots, results, and engine errors
10. `game/serialization.go` — versioned complete-session persistence and atomic restoration
11. `cli/controller.go` — line-oriented CLI controller (terminal I/O, commands, display)
12. `sessionfile/session_file.go` — presentation-neutral bounded and atomic session transport
13. `tui/model.go` — Bubble Tea event model and action translation
14. `tui/render.go` — deterministic full-screen renderer
15. `recovery/recovery.go` — private XDG recovery records, validation, retention, and atomic transport

### Understanding Puzzle Generation
1. `.aidoc/designs/difficulty-model.md` — current model, tier invariants, and configuration boundary
2. `.aidoc/designs/difficulty-calibration.md` — measurement contract, evidence, reports, and decision gates
3. `calibration/runner.go` — immutable manifests, reproducibility checks, observations, checkpoints, and reports
4. `calibration/baselines/mixed-imported-expansion-v5/report.md` — current imported-stratum expansion results and limitations
5. `calibration/baselines/mixed-generated-expansion-v4/report.md` — preserved generated-stratum expansion results
6. `calibration/baselines/mixed-external-expansion-v3/report.md` — preserved expanded external baseline
7. `calibration/baselines/mixed-pilot-v2/report.md` — preserved initial pilot baseline
8. `cmd/calibrate.go` — resumable local measurement command
9. `generator/difficulty.go` — difficulty levels and `StrategySolverKeys`
10. `generator/generator.go` — board generation, cell removal, best-effort generation with limits
11. `generator/options.go` — `Options` and `BestEffortOptions` (time/round limits)
12. `solver/classify.go` — puzzle classification (difficulty tier, score, max technique)
13. `db/db.go` — SQLite database open/close/migrate
14. `db/puzzle.go` — puzzle CRUD, random query by difficulty, dedup
15. `cmd/play.go` — fallback flow (generator → DB lookup → graceful degradation) and auto-store
16. `cmd/generate.go` — batch generation CLI (parallel workers, progress, report)
17. `cmd/import.go` — import CLI (file parsing, normalization, dedup, report)

### Understanding the Roadmap
1. `.aidoc/designs/roadmap.md` — stabilization priorities, sequencing, and exit criteria
2. `.aidoc/designs/e2e-test-scenarios.md` — compatibility and black-box acceptance scenarios
3. `.aidoc/designs/difficulty-model.md` — calibration boundary and current difficulty invariants
4. `.aidoc/designs/difficulty-calibration.md` — measurement methodology, report contract, and product decisions
5. `.aidoc/designs/future-directions.md` — deliberately non-priority product and production directions
6. `.aidoc/designs/web-api.md` — client-neutral HTTP resources, revisions, recovery, client access, and security boundary
7. `api/openapi.yaml` — canonical OpenAPI 3.1.1 wire contract, schemas, errors, and examples
8. `.aidoc/designs/game-engine.md` — stable engine API, notes, history, and serialization design
9. `.aidoc/designs/background-autosave.md` — recovery lifecycle, privacy, storage, retention, and conflict policy
10. `.aidoc/designs/tui-frontend.md` — current full-screen interaction and rendering semantics
11. `.aidoc/architecture/guidelines.md` — current architecture and solver contract

### Adding a New Strategy Solver
1. `.aidoc/architecture/guidelines.md` — constraints, interface contract, step-by-step
2. `solver/solver.go` — implement `StrategySolver`
3. `solver/move.go` — return `*Move` from `Apply()`
4. `solver/store.go` — register with `RegisterStrategy()`
5. Write tests in `solver/<name>_solver_test.go`
6. Update `generator/difficulty.go` to reference the new solver key

## Document Map

| Path | Purpose |
|------|---------|
| `AGENT.md` | AI operator entry point — rules and repo layout |
| `.aidoc/INDEX.md` | This file — discovery index and reading chains |
| `.aidoc/architecture/guidelines.md` | Design constraints, layer boundaries, solver contract |
| `.aidoc/designs/difficulty-model.md` | Difficulty invariants, scoring model, and calibration boundary |
| `.aidoc/designs/difficulty-calibration.md` | Calibration methodology, corpus contract, evidence, reports, and decision gates |
| `.aidoc/designs/roadmap.md` | Stabilization priorities, sequencing, and exit criteria |
| `.aidoc/designs/future-directions.md` | Non-priority product and production directions with decision gates |
| `.aidoc/designs/web-api.md` | Contract-first OpenAPI workflow, resources, revisions, recovery, client access, and network security boundary |
| `.aidoc/designs/background-autosave.md` | Background autosave lifecycle, privacy, retention, and conflict design |
| `.aidoc/designs/automatic-candidates.md` | Automatic-candidate engine contract, TUI interaction, and rendering constraints |
| `.aidoc/designs/tui-frontend.md` | TUI interaction model, persistence policy, rendering, and dependency boundaries |
| `.aidoc/designs/cli-sessions.md` | CLI manual notes, rendering, save, and resume design |
| `.aidoc/designs/game-engine.md` | Engine API, notes, unified history, snapshots, and serialization design |
| `.aidoc/designs/e2e-test-scenarios.md` | E2E test scenarios — black-box user scenarios for manual/script testing |
| `api/openapi.yaml` | Canonical OpenAPI 3.1.1 HTTP wire contract |
| `README.md` | Human-facing project summary |
| `cmd/root.go` | Cobra root command and shared state |
| `cmd/play.go` | Interactive play mode, fallback flow, auto-store |
| `cmd/session.go` | Shared CLI/TUI session startup and restore validation |
| `cmd/tui.go` | Opt-in full-screen TUI command and terminal lifecycle |
| `cmd/api.go` | API flags, dependency wiring, and signal-aware server shutdown |
| `webapi/server.go` | HTTP security boundary, session registry, recovery, and engine translation |
| `webapi/generated.go` | Generated OpenAPI models and strict server interface |
| `tui/model.go` | TUI event model, focus, modes, confirmations, and persistence |
| `tui/render.go` | Deterministic color-independent board rendering |
| `recovery/recovery.go` | Private XDG recovery records, discovery, validation, retention, and deletion |
| `sessionfile/session_file.go` | Bounded reads and atomic mode-0600 session writes |
| `calibration/runner.go` | Immutable corpus manifests, append-only observations, resumable checkpoints, and derived reports |
| `calibration/testdata/mixed-pilot-v2.json` | Immutable traceable mixed-corpus pilot manifest |
| `calibration/baselines/mixed-pilot-v2/report.md` | Preserved initial pilot baseline and statistical limitations |
| `calibration/testdata/mixed-external-expansion-v3.json` | Immutable source-order external-stratum expansion manifest |
| `calibration/baselines/mixed-external-expansion-v3/report.md` | Preserved external expansion evidence and limitations |
| `calibration/testdata/mixed-generated-expansion-v4.json` | Immutable sequential generated-stratum expansion manifest |
| `calibration/baselines/mixed-generated-expansion-v4/report.md` | Preserved generated expansion evidence and limitations |
| `calibration/testdata/mixed-imported-expansion-v5.json` | Immutable source-order imported-stratum expansion manifest |
| `calibration/baselines/mixed-imported-expansion-v5/report.md` | Current imported expansion evidence and remaining limitations |
| `cmd/calibrate.go` | Local difficulty measurement CLI boundary |
| `cmd/generate.go` | Batch generation CLI (parallel workers, progress, report) |
| `cmd/import.go` | Import CLI (file parsing, normalization, dedup, report) |
| `scripts/e2e_cli.py` | Built-binary line CLI, session, calibration, import, generation, and SQLite E2E harness |
| `scripts/e2e_api.py` | Built-binary HTTP lifecycle E2E harness |
| `scripts/e2e_tui.py` | Built-binary PTY TUI and recovery E2E harness |
| `scripts/coverage_report.py` | Package-level Go coverage summary for risk-based CI review |
| `db/db.go` | SQLite puzzle database — open, close, schema migration |
| `db/puzzle.go` | Puzzle CRUD, random query by difficulty, statistics |
| `game/contract.go` | Stable engine actions, snapshots, results, and typed errors |
| `game/serialization.go` | Versioned JSON session serialization, validation, and restoration |
| `solver/classify.go` | Puzzle classification — difficulty tier, score, max technique |

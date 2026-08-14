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
11. `cli/controller.go` — CLI controller (terminal I/O, commands, display)

### Understanding Puzzle Generation
1. `.aidoc/designs/difficulty-model.md` — current model (clue-count), limitations, target model
2. `generator/difficulty.go` — difficulty levels and `StrategySolverKeys`
3. `generator/generator.go` — board generation, cell removal, best-effort generation with limits
4. `generator/options.go` — `Options` and `BestEffortOptions` (time/round limits)
5. `solver/classify.go` — puzzle classification (difficulty tier, score, max technique)
6. `db/db.go` — SQLite database open/close/migrate
7. `db/puzzle.go` — puzzle CRUD, random query by difficulty, dedup
8. `cmd/play.go` — fallback flow (generator → DB lookup → graceful degradation) and auto-store
9. `cmd/generate.go` — batch generation CLI (parallel workers, progress, report)
10. `cmd/import.go` — import CLI (file parsing, normalization, dedup, report)

### Understanding the Roadmap
1. `.aidoc/designs/roadmap.md` — Phase 7 scope, delivery order, and exit criteria
2. `.aidoc/designs/tui-frontend.md` — full-screen interaction, persistence, rendering, and dependency design
3. `.aidoc/designs/game-engine.md` — stable engine API, notes, history, and serialization design
4. `.aidoc/designs/cli-sessions.md` — existing manual-note and persistence frontend design
5. `.aidoc/architecture/guidelines.md` — current architecture and solver contract
6. `.aidoc/designs/e2e-test-scenarios.md` — CLI compatibility and future TUI scenarios

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
| `.aidoc/designs/difficulty-model.md` | Difficulty model: current state, limitations, and target design |
| `.aidoc/designs/roadmap.md` | Phase 7 scope, delivery plan, and exit criteria |
| `.aidoc/designs/tui-frontend.md` | TUI interaction model, persistence policy, rendering, and dependency boundaries |
| `.aidoc/designs/cli-sessions.md` | CLI manual notes, rendering, save, and resume design |
| `.aidoc/designs/game-engine.md` | Engine API, notes, unified history, snapshots, and serialization design |
| `.aidoc/designs/e2e-test-scenarios.md` | E2E test scenarios — black-box user scenarios for manual/script testing |
| `README.md` | Human-facing project summary |
| `cmd/root.go` | Cobra root command and shared state |
| `cmd/play.go` | Interactive play mode, fallback flow, auto-store |
| `cmd/generate.go` | Batch generation CLI (parallel workers, progress, report) |
| `cmd/import.go` | Import CLI (file parsing, normalization, dedup, report) |
| `db/db.go` | SQLite puzzle database — open, close, schema migration |
| `db/puzzle.go` | Puzzle CRUD, random query by difficulty, statistics |
| `game/contract.go` | Stable engine actions, snapshots, results, and typed errors |
| `game/serialization.go` | Versioned JSON session serialization, validation, and restoration |
| `solver/classify.go` | Puzzle classification — difficulty tier, score, max technique |

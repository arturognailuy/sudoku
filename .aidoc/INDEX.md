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
3. `solver/sudoku_solver.go` — `ISudokuSolver` interface and `BaseSolver`
4. `solver/sudoku_solver_store.go` — solver registry

### Understanding Puzzle Generation
1. `.aidoc/designs/difficulty-model.md` — current model (clue-count), limitations, target model
2. `generator/sudoku_generator_difficulty.go` — difficulty levels and `StrategySolverKeys`
3. `generator/sudoku_generator.go` — board generation and cell removal logic

### Understanding the Roadmap
1. `.aidoc/designs/roadmap.md` — future phases: strategy solvers, puzzle database, core refactoring, UI-ready engine
2. `.aidoc/designs/difficulty-model.md` — difficulty model and generation/storage flow
3. `.aidoc/architecture/guidelines.md` — current architecture and solver contract

### Adding a New Strategy Solver
1. `.aidoc/architecture/guidelines.md` — constraints, interface contract, step-by-step
2. `solver/sudoku_solver.go` — implement `ISudokuSolver`
3. `solver/sudoku_solver_store.go` — register in `NewSudokuSolverStore()`
4. Write tests in `solver/<name>_solver_test.go`
5. Update `generator/sudoku_generator_difficulty.go` to reference the new solver key

## Document Map

| Path | Purpose |
|------|---------|
| `AGENT.md` | AI operator entry point — rules and repo layout |
| `.aidoc/INDEX.md` | This file — discovery index and reading chains |
| `.aidoc/architecture/guidelines.md` | Design constraints, layer boundaries, solver contract |
| `.aidoc/designs/difficulty-model.md` | Difficulty model: current state, limitations, and target design |
| `.aidoc/designs/roadmap.md` | Future phases: strategy solvers, puzzle database, core refactoring, UI-ready engine |
| `README.md` | Human-facing project summary |

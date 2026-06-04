# .aidoc/INDEX.md — Discovery Index

## Reading Chains

### Understanding the Architecture
1. `AGENT.md` — global rules, repo layout, key interfaces
2. `.aidoc/architecture/guidelines.md` — Go conventions, layer separation, solver interface contract
3. `solver/sudoku_solver.go` — `ISudokuSolver` interface and `BaseSolver`
4. `solver/sudoku_solver_store.go` — solver registry
5. `solver/sudoku_solver_default.go` — backtracking solver (reference implementation)

### Understanding Puzzle Generation
1. `generator/sudoku_generator_difficulty.go` — difficulty levels and `StrategySolverKeys`
2. `generator/sudoku_generator_options.go` — generation options
3. `generator/sudoku_generator.go` — board generation and cell removal

### Understanding the Difficulty Model
1. `.aidoc/designs/difficulty-model.md` — current model (clue-count), limitations, target model (strategy-based)

### Adding a New Strategy Solver
1. `.aidoc/architecture/guidelines.md` — conventions and interface contract
2. `solver/sudoku_solver.go` — implement `ISudokuSolver`
3. `solver/sudoku_solver_store.go` — register in `NewSudokuSolverStore()`
4. Write tests in `solver/<name>_test.go`
5. Update `generator/sudoku_generator_difficulty.go` to reference the new solver key

## Document Map

| Path | Purpose |
|------|---------|
| `AGENT.md` | AI operator entry point |
| `.aidoc/INDEX.md` | This file — discovery index |
| `.aidoc/architecture/guidelines.md` | Go conventions, layer separation, solver contract |
| `.aidoc/designs/difficulty-model.md` | Difficulty model: current state and target design |
| `README.md` | Human-facing project summary |

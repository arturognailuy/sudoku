# AGENT.md — AI Operator Instructions

You are an AI agent working on **Sudoku**, a CLI Sudoku game written in Go.

## Quick Start

```bash
go build && go test ./...
```

## Entry Points

- **This file** — global rules, branching, CI requirements.
- **`.aidoc/INDEX.md`** — discovery index. Start here for architecture, designs, and reading chains.
- **`README.md`** — human-facing project summary.

## Repository Layout

```
.
├── main.go              # Entry point — CLI parsing → game loop
├── cli/                 # Command-line flag parsing (difficulty enum, help)
├── core/                # Board, cell, position, validator, normalizer, string serialization
├── solver/              # ISudokuSolver interface, solver store, default backtracking solver
├── generator/           # Puzzle generation — solved board + cell removal with uniqueness checks
├── game/                # Game state, undo/redo/hints, interactive CLI play loop, signal handling
├── util/                # Random shuffle, array helpers
├── .aidoc/              # AI-native documentation
│   ├── INDEX.md
│   ├── architecture/
│   └── designs/
└── .github/workflows/   # CI (go test, go vet, golangci-lint)
```

## Rules

### Branching

- Work on feature branches (`feature/<short>` or `fix/<short>`), never directly on `main`.
- PRs target `main`. Squash-merge only.

### Code Style

- Follow existing Go conventions in the codebase.
- Use `gofmt` / `goimports` formatting.
- No new dependencies without justification.
- Keep packages focused: one responsibility per package.

### Testing

- Every new solver must have tests.
- Run `go test ./...` before committing.
- Run `go vet ./...` to catch issues.
- CI must pass before merge.

### Commit Messages

- Use conventional style: `feat:`, `fix:`, `docs:`, `test:`, `chore:`.
- Keep subject line under 72 characters.

### Documentation

- Keep `.aidoc/` docs in sync with code changes in the same PR.
- Follow DocGuidelines: docs capture the *why* and *constraints*, not the *how* that code already expresses.
- `README.md` is for humans; `.aidoc/` is for AI agents.

## Domain Context

This is a Sudoku puzzle game. The current difficulty model is clue-count-based (how many cells are pre-filled). The goal is to evolve toward strategy-based difficulty, where difficulty is determined by the hardest solving technique required. See `.aidoc/designs/difficulty-model.md` for the design.

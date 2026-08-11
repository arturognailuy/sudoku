---
domain: Designs
status: Active
entry_points:
  - game/game.go
dependencies:
  - .aidoc/designs/ui-ready-engine.md
  - .aidoc/architecture/guidelines.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Roadmap

The next milestone is Phase 5: formalize the existing pure game logic as a stable, serializable engine with note-taking support. The CLI remains the first frontend and proves the boundary before any graphical UI begins.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/ui-ready-engine.md` | Canonical Phase 5 behavior and constraints |
| `.aidoc/architecture/guidelines.md` | Current package and dependency boundaries |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box compatibility contract for the CLI |

## Phase 5 Outcome

Phase 5 provides one engine contract for game state, typed actions, validation, hints, undo/redo, notes, snapshots, and versioned serialization. Frontends consume detached snapshots and submit actions without accessing mutable boards or reproducing Sudoku rules.

Phase 5 does not add a TUI, web UI, or mobile app. A new frontend should begin only after the CLI is fully migrated and the engine API has survived implementation and black-box verification.

## Delivery Plan

| PR | Scope | Acceptance boundary |
|----|-------|---------------------|
| 1 | Engine contract and encapsulation | Introduce typed actions, detached snapshots, structured results/errors, and compatibility adapters; remove public mutable-state dependencies from new API usage. |
| 2 | Notes and unified history | Add manual notes, automatic peer cleanup, and atomic undo/redo across value and note changes. |
| 3 | Versioned serialization | Save and restore complete sessions, including notes and undo/redo history; reject corrupt or unsupported state atomically. |
| 4 | CLI migration and verification | Move `cli.Controller` to actions/snapshots, retire obsolete adapters, and run the documented black-box E2E scenarios. |

Implementation PRs are sequential because each one stabilizes the contract consumed by the next. Every PR includes package tests and keeps existing CLI behavior green.

## Exit Criteria

- `game.Game` exposes no mutable internal board or history state to frontends.
- Value changes, notes, hints, reset, undo, and redo use one action model.
- Snapshots contain everything a frontend needs to render a session.
- Serialization round trips preserve visible state and undo/redo behavior.
- The CLI depends only on the stable engine contract for game operations.
- Existing black-box CLI scenarios, package tests, lint, vet, and CI pass.

## Deferred Work

Graphical frontends, network protocols, cloud sync, multi-user play, localization, and automatic note population are outside Phase 5. Those features may consume the engine later without changing its core state-transition semantics.

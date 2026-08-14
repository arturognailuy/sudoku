---
domain: Designs
status: Active
entry_points:
  - game/game.go
dependencies:
  - .aidoc/designs/game-engine.md
  - .aidoc/architecture/guidelines.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Roadmap

Phase 5 formalizes the pure game logic as a stable, serializable engine with note-taking support. The migrated CLI is the first frontend and proves the action/snapshot boundary before any graphical UI begins.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/game-engine.md` | Canonical Phase 5 behavior and constraints |
| `.aidoc/architecture/guidelines.md` | Current package and dependency boundaries |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box compatibility contract for the CLI |

## Phase 5 Outcome

Phase 5 provides one engine contract for game state, typed actions, validation, hints, undo/redo, notes, snapshots, and versioned serialization. Frontends consume detached snapshots and submit actions without accessing mutable boards or reproducing Sudoku rules.

Phase 5 does not add a TUI, web UI, or mobile app. A new frontend should begin only after the CLI is fully migrated and the engine API has survived implementation and black-box verification.

## Implementation Boundaries

- `game.Game.Apply` accepts typed value, note, hint, history, reset, repair, and solve actions; command-shaped mutation methods remain private.
- `game.Game.Snapshot` returns all renderable state as detached values.
- Unified history restores values, invalid markers, and notes atomically.
- Versioned serialization preserves complete sessions and validates restored data before replacing state.
- `cli.Controller` renders snapshots and submits actions without reproducing engine rules.

## Exit Criteria

- `game.Game` exposes no mutable internal board or history state to frontends.
- Value changes, notes, hints, reset, undo, and redo use one action model.
- Snapshots contain everything a frontend needs to render a session.
- Serialization round trips preserve visible state and undo/redo behavior.
- The CLI depends only on the stable engine contract for game operations.
- Existing black-box CLI scenarios, package tests, lint, vet, and CI pass.

## Deferred Work

Graphical frontends, network protocols, cloud sync, multi-user play, localization, and automatic note population are outside Phase 5. Those features may consume the engine later without changing its core state-transition semantics.

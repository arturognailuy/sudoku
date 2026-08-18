---
domain: Designs
status: Active
entry_points:
  - cmd/tui.go
  - tui/model.go
dependencies:
  - .aidoc/designs/automatic-candidates.md
  - .aidoc/designs/tui-frontend.md
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Roadmap

Phase 8 adds opt-in automatic candidate assistance as the next gameplay improvement after the full-screen TUI. The feature strengthens the shared engine read boundary without introducing persistence changes, hosting, accounts, or network infrastructure.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/automatic-candidates.md` | Canonical Phase 8 engine, interaction, rendering, and compatibility decisions |
| `.aidoc/designs/game-engine.md` | Stable state and snapshot boundary consumed by every frontend |
| `.aidoc/designs/tui-frontend.md` | Current full-screen interaction, rendering, and accessibility contract |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box compatibility contract and Phase 8 acceptance scenarios |

## Why Phase 8 Is Next

The current TUI makes keyboard play pleasant and proves that two frontends can share actions, snapshots, hints, and serialization. Automatic candidates now provide more player value than another frontend while exercising a small missing part of the read contract: presentation-neutral, derived legal-candidate data.

A web frontend would first require API or WebAssembly, hosting, authentication, and deployment decisions. Background autosave would require path, retention, recovery, and privacy policy. Candidate assistance is lower risk because `core.Board.Candidates` already defines the rule and the feature does not create mutable session state.

## Phase 8 Outcome

Players can toggle legal candidates in the full-screen TUI without changing the default board, manual notes, undo history, dirty state, or saved sessions. Candidate sets update from accepted values after every engine transition and remain visually distinct from manual notes in dark, light, and no-color modes.

The line-oriented CLI remains unchanged. The engine exposes detached candidate data so later frontends can choose their own presentation without recomputing Sudoku rules.

## Delivery Order

1. Add detached legal candidates to `game.Snapshot` using `core.Board.Candidates`, with no cache or second algorithm.
2. Add the opt-in TUI toggle and combined automatic-candidate/manual-note rendering.
3. Update help, package tests, the pseudo-terminal harness, black-box scenarios, and player documentation.

Each layer preserves serialization version 1 and all current CLI/TUI behavior. The implementation adds no dependency and no candidate-specific action or history record.

## Exit Criteria

- Candidate sets are correct for editable empty cells and refresh after value entry, clear, undo/redo, hints, repair, solve, reset, and restore.
- Invalid visible entries do not constrain peers and suppress candidates in their own occupied cells.
- Automatic candidates start off, toggle without dirtying the session, and are never persisted.
- Manual notes remain player-owned, including stale notes, and remain distinguishable from automatic candidates without color alone.
- Dark, light, and `NO_COLOR` rendering preserve the current board geometry and accessibility semantics.
- Root CLI output and commands, version 1 save bytes, restore behavior, hints, and solver behavior remain compatible.
- Package tests, pseudo-terminal scenarios, applicable root CLI E2E scenarios, build, vet, lint, diff checks, and CI pass.

## Later Work

Background autosave and crash recovery follow Phase 8 only after explicit path, privacy, retention, and conflict policies are designed. A web or API boundary follows local-session reliability work. Mouse support, localization, cloud sync, mobile clients, and multi-user features remain deferred until a concrete product need justifies their architectural cost.

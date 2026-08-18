---
domain: Designs
status: Active
entry_points:
  - cmd/tui.go
  - tui/model.go
dependencies:
  - .aidoc/designs/tui-frontend.md
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Roadmap

Phase 7 delivers an optional full-screen terminal interface while preserving the line-oriented CLI. The TUI is the smallest second frontend for validating the reusable game engine before the project commits to a web, mobile, or network boundary.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/tui-frontend.md` | Canonical Phase 7 interaction, persistence, rendering, and dependency decisions |
| `.aidoc/designs/game-engine.md` | Stable state and serialization boundary consumed by every frontend |
| `.aidoc/designs/cli-sessions.md` | Existing CLI behavior that Phase 7 must preserve |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box compatibility contract and future TUI scenarios |

## Why Phase 7 Exists

Phase 6 proved that a frontend can render manual notes, transport complete sessions, and preserve mixed value/note history through the engine contract. A second frontend now needs to prove that the contract is presentation-independent rather than merely a cleaner CLI implementation.

A TUI is the next frontend because it consumes the Go engine in-process and delivers a materially better keyboard experience without first deciding on HTTP APIs, WebAssembly, hosting, authentication, or mobile distribution. The root CLI remains the compatibility interface for scripts and redirected input; the TUI is opt-in through `sudoku tui`.

## Phase 7 Outcome

Players can launch a full-screen board, navigate with the keyboard, enter values and manual notes, inspect and apply hints, use unified undo/redo, save explicitly, resume safely, and quit without accidental loss. The interface handles terminal resize and limited color while preserving all engine and session invariants.

Phase 7 does not add automatic candidate population, background autosave, mouse support, a web service, a browser frontend, or a mobile app. Those features remain separate product decisions after the multi-frontend boundary is proven.

## Delivered Layers

1. Extracted presentation-neutral session-file transport and shared game startup while preserving every root CLI behavior and black-box scenario.
2. Added the TUI model, deterministic renderer, focus movement, value/note input, status messages, and engine action translation.
3. Added hint preview/apply, destructive-action confirmations, explicit save, resume-path handling, dirty-state protection, and small-terminal fallback.
4. Added pseudo-terminal E2E coverage, help and README guidance, then reassess whether Phase 8 should target automatic candidates, autosave, or a web frontend.

Each layer leaves the built program usable. The event loop serializes all `game.Game` mutations; persistence and generation may run outside the loop only when they return results as messages.

## Exit Criteria

- `sudoku tui` starts from explicit input, generated, and restored puzzles without changing root CLI defaults or output.
- Keyboard navigation, value mode, note mode, peer-note cleanup, hint preview/apply, undo, and redo expose engine behavior without duplicating Sudoku rules.
- Given, focused, peer, invalid, and editable cells remain distinguishable without color alone.
- Save remains explicit and atomic; failed save or restore preserves existing data and displays an actionable error.
- Unsaved changes require confirmation before quit, while a clean session exits directly.
- Resize events produce a usable board or a clear minimum-size fallback without corrupting state.
- Package tests, pseudo-terminal scenarios, existing CLI E2E scenarios, `go test`, `go vet`, lint, diff checks, and CI pass.

## Deferred Work

Automatic candidate population, background autosave and crash recovery, mouse support, themes, localization, web and mobile frontends, network protocols, cloud sync, and multi-user play remain deferred. Later frontends should continue consuming the engine action, snapshot, hint, and serialization contracts rather than moving their rules into transport or presentation code.

---
domain: Designs
status: Active
entry_points:
  - cli/controller.go
  - cmd/play.go
dependencies:
  - .aidoc/designs/cli-sessions.md
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Roadmap

Phase 6 turns the engine-backed command line into a complete session frontend. Players can use manual notes, leave and resume a game without losing history, and recover safely from invalid session files before work begins on a graphical frontend.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/cli-sessions.md` | Canonical Phase 6 command, rendering, and persistence behavior |
| `.aidoc/designs/game-engine.md` | Stable state and serialization boundary consumed by the CLI |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box compatibility contract for session workflows |

## Why Phase 6 Exists

The game engine already owns notes and complete serialized sessions, but the CLI cannot expose either capability. Adding a different frontend now would leave the first frontend unable to exercise important parts of the engine contract and would postpone validation of persistence transport and note rendering.

Phase 6 keeps the project CLI-first while making interactive play practical across multiple terminal sessions. The CLI remains a thin adapter: note rules stay in `game.Game.Apply`, session validation stays in `game.Restore`, and files remain frontend-owned transport.

## Phase 6 Outcome

The interactive CLI renders manual notes, accepts note actions, saves complete sessions, and resumes saved sessions. Restore failures preserve existing data and produce actionable terminal errors rather than silently starting a different puzzle.

Phase 6 does not add automatic candidate population, background autosave, TUI dependencies, network storage, or a graphical frontend. Those features require separate product decisions after the explicit CLI workflow is proven.

## Delivery Order

1. Add note commands and a deterministic note-aware board rendering without changing value-entry shorthand.
2. Add explicit save and resume transport around `game.Game.Serialize` and `game.Restore`, including atomic file replacement.
3. Complete black-box session scenarios, help text, and player documentation, then reassess whether the next frontend should be a TUI, web UI, or mobile app.

Each step should remain independently reviewable and keep the built CLI usable. Package tests verify adapters and failure paths; black-box scenarios verify the compiled program and filesystem behavior.

## Exit Criteria

- Players can toggle and clear notes on editable empty cells and see the resulting notes.
- Value entry removes target and peer notes visibly; undo and redo restore values and notes together.
- Players can save to an explicit path and resume the same visible state and undo/redo cursor.
- Saving replaces the destination atomically and does not leave a partial session file.
- Malformed, unsupported, or semantically invalid sessions fail without overwriting source data.
- Help output and the black-box scenario catalog describe every new command and flag.
- Tests, vet, lint, diff checks, applicable built-binary E2E scenarios, and CI pass.

## Deferred Work

Automatic candidate population, autosave policy, graphical frontends, network protocols, cloud sync, multi-user play, and localization remain deferred. Later frontends should continue consuming the same engine action, snapshot, and serialization contracts rather than moving their rules into transport or presentation code.

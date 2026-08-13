---
domain: Designs
status: Draft
entry_points:
  - game/game.go
dependencies:
  - .aidoc/architecture/guidelines.md
  - .aidoc/designs/roadmap.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# UI-Ready Game Engine

Phase 5 turns `game.Game` into a stable, reusable engine for the CLI and future TUI, web, or mobile frontends. The engine owns game rules and state transitions; frontends own input, rendering, navigation, and persistence transport.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/architecture/guidelines.md` | Current package boundaries that the engine must preserve |
| `.aidoc/designs/roadmap.md` | Phase 5 delivery order |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box CLI behavior that must remain compatible |

## Why the Engine Boundary Exists

`game.Game` already avoids terminal I/O, but public mutable boards, command-shaped methods, and unexported history state prevent a frontend from treating it as a durable engine contract. A stable boundary lets every frontend observe the same state, submit the same actions, and receive the same validation results without duplicating Sudoku rules.

The Phase 5 work extends the existing `game` package rather than introducing a second game model. The CLI remains the first consumer until the contract is proven; no graphical frontend is part of this phase.

## Engine Responsibilities

The engine owns:

- immutable puzzle givens and the current player state;
- accepted values, invalid player entries, and player notes;
- action validation and deterministic state transitions;
- unified undo/redo history for value and note actions;
- solved, valid, and conflict status;
- hint calculation through the configured solver store;
- versioned state serialization and restoration.

Frontends own presentation, command parsing, keyboard or pointer gestures, accessibility, localization, file/network transport, and autosave timing. The engine returns structured state and errors; it does not print, read input, exit a process, or choose UI wording.

## Public Contract

The stable contract is organized around four concepts:

- `Game` is the authoritative mutable session owned by one caller at a time.
- `Snapshot` is a detached read model containing givens, visible values, invalid markers, notes, status, and undo/redo availability. Mutating a snapshot cannot mutate the game.
- `Action` is a typed player intent: set or clear a value, toggle or clear notes, reset, or apply a hint. Frontends submit actions instead of reproducing rules.
- `Result` describes the accepted transition, changed cells, current status, and whether undo or redo is available. Invalid actions return typed errors and leave state unchanged.

`Hint` remains a query: it returns a structured recommendation with position, value, technique, and explanation. Applying the recommendation is a separate action so hints participate in history exactly like player moves.

Existing convenience methods may remain temporarily as adapters, but `cli.Controller` must migrate to the action/snapshot contract before those adapters are removed. Public callers must not receive mutable references to the engine's internal boards or history.

## State and Validation Invariants

- Puzzle givens never change after game creation or restoration.
- Every accepted action is atomic: all value, note, status, and history changes succeed together or not at all.
- New actions after an undo discard the redo tail.
- Undo and redo restore the complete prior state, including invalid markers and all note changes caused by an action.
- User mistakes remain observable for rendering but never contaminate the solver's valid board.
- Public input errors return typed errors; panics remain reserved for violated internal invariants.
- Snapshots are self-consistent and safe for asynchronous rendering after the engine advances.
- A `Game` is not concurrently mutable. Frontends serialize actions and may pass snapshots between goroutines.

## Note-Taking Semantics

Notes are player annotations, distinct from `core.Board.Candidates`, which computes legal solver candidates. Notes use `core.CandidateSet` as a value representation but are stored by the game engine.

Notes are allowed only on editable empty cells and contain digits 1–9. Setting a value clears notes in that cell and removes that value from notes in peer cells. Clearing a value does not recreate notes. Every automatic note cleanup is part of the same action delta, so undo restores the exact previous notes.

The engine does not automatically fill all legal candidates or continuously synchronize notes with `core.Board.Candidates`. A frontend may request an explicit future helper for that behavior without changing the meaning of manual notes.

## Serialization Contract

Serialized state is a versioned data contract, not formatted CLI output. The representation includes the original puzzle, player values, invalid entries, notes, action history, and history cursor so a restored game preserves undo/redo behavior.

Restoration validates the schema version, puzzle, positions, values, notes, and history before constructing a game. Solver configuration is supplied by the host during restoration rather than embedded as executable configuration. Unknown future schema versions fail explicitly; migrations can be added per version.

`Game.ToString` remains diagnostic text and is not a persistence format.

## Compatibility and Failure Boundaries

Phase 5 preserves existing CLI commands and black-box behavior unless a separately reviewed UX change says otherwise. The CLI renders snapshots and translates commands into actions; it must not inspect engine internals.

Serialization failures, invalid actions, unavailable undo/redo, and attempts to edit givens are expected errors. Solver inability to produce a hint is a valid no-hint result, not an engine failure. Corrupt restored state must never produce a partially initialized game.

## Verification

Engine tests cover every action, typed error, atomic rollback, note cleanup, mixed value/note undo-redo sequences, redo truncation, immutable snapshots, and serialization round trips. Restoration tests include malformed and unsupported versions.

CLI migration is verified against `.aidoc/designs/e2e-test-scenarios.md` by building the binary and exercising it as a black box. Package tests, `go vet`, `golangci-lint`, and CI must pass for each implementation PR.

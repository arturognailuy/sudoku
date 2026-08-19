---
domain: Designs
status: Active
entry_points:
  - cmd/tui.go
  - tui/model.go
  - sessionfile/session_file.go
dependencies:
  - .aidoc/designs/background-autosave.md
  - .aidoc/designs/tui-frontend.md
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Roadmap

Phase 9 adds private background autosave and crash recovery for the full-screen TUI. The phase improves local-session reliability while preserving explicit saves, line-oriented CLI behavior, the stable game-engine contract, and version 1 serialized gameplay state.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/background-autosave.md` | Canonical Phase 9 lifecycle, privacy, storage, conflict, and retention decisions |
| `.aidoc/designs/tui-frontend.md` | Current TUI event loop, dirty state, persistence, and modal behavior |
| `.aidoc/designs/game-engine.md` | Stable serialization and restoration boundary |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box compatibility contract and Phase 9 acceptance scenarios |

## Why Phase 9 Is Next

The TUI now supports long-running play with notes, history, hints, and automatic candidates, but process or host interruption can still lose unsaved progress. Recovery provides immediate player value and completes the local session lifecycle before an API, browser frontend, cloud sync, or account model introduces remote state and broader conflict policy.

The engine already serializes complete sessions, and `sessionfile` already provides bounded reads and atomic private writes. Phase 9 can therefore remain a frontend lifecycle feature instead of changing Sudoku rules, history, solver behavior, or the serialized gameplay schema.

## Phase 9 Outcome

The TUI writes debounced recovery records after successful gameplay mutations and offers valid recent records when a plain TUI session starts. Recovery files use a private XDG state location, support concurrent local games, survive abnormal termination, and disappear after successful explicit save or confirmed discard.

Autosave is enabled by default with an explicit opt-out. Intentional startup using `--input`, `--level`, or `--resume` remains deterministic and bypasses recovery discovery. The line-oriented CLI never creates recovery files.

## Delivery Order

1. Add a presentation-neutral recovery package for wrapper validation, secure XDG paths, discovery, retention, atomic writes, and deletion.
2. Add TUI mutation generations, one-second debounce, single-flight writes, warning and retry behavior, and cleanup policy.
3. Add startup recovery selection, explicit discard, `--no-autosave`, and deterministic bypass for explicit puzzle sources.
4. Update player help, package tests, pseudo-terminal coverage, black-box scenarios, and current-state documentation.

Implementation keeps recovery metadata outside `game`, uses `game.Serialize` and `game.Restore` as opaque boundaries, and adds no external dependency.

## Exit Criteria

- Only successful gameplay mutations create or refresh recovery state; presentation-only actions do not.
- Recovery uses mode-`0700` directories, mode-`0600` regular files, bounded atomic replacement, random opaque names, and no symlink traversal.
- One-second debounce and single-flight writes preserve the newest generation without concurrent-session overwrites.
- Plain TUI startup can resume or discard recent valid records, while explicit startup sources and opt-out remain deterministic.
- Successful explicit save, clean unmodified exit, and confirmed discard remove the current record; abnormal termination leaves the latest successful record.
- Invalid records are ignored, records older than 30 days are pruned, and write failures remain visible without terminating play.
- Root CLI behavior, explicit save bytes, version 1 restore semantics, engine actions, TUI gameplay, and automatic candidates remain compatible.
- Package tests, pseudo-terminal scenarios, applicable root CLI E2E scenarios, build, vet, lint, diff checks, documentation audit, and CI pass.

## Later Work

A web or API boundary follows local-session reliability work only when hosting and deployment requirements are concrete. Mouse support, localization, cloud sync, encryption, mobile clients, account identity, cross-device merge, and multi-user features remain deferred until their product need justifies new architectural and privacy costs.

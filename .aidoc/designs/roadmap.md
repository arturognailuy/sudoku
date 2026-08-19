---
domain: Designs
status: Active
entry_points:
  - cmd/root.go
  - cmd/session.go
  - game/contract.go
  - recovery/recovery.go
dependencies:
  - .aidoc/designs/web-api.md
  - .aidoc/designs/browser-frontend.md
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Roadmap

Phase 10 adds a local-first HTTP API and embedded browser frontend after the reusable engine, complete sessions, TUI, candidate assistance, and crash recovery have stabilized. The phase proves a non-terminal boundary without prematurely committing the project to hosting, accounts, or remote data storage.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/web-api.md` | Canonical API resources, revisions, recovery, and local security boundary |
| `.aidoc/designs/browser-frontend.md` | Browser interaction, accessibility, assets, and client-state decisions |
| `.aidoc/designs/game-engine.md` | Stable actions, snapshots, hints, and serialization reused by the API |
| `.aidoc/designs/e2e-test-scenarios.md` | Existing compatibility contract and Phase 10 acceptance scenarios |

## Why Phase 10 Exists

The game engine now supports multiple terminal presentations without leaking UI concerns, and private recovery protects long-running local sessions. A browser frontend is the next useful stress test: HTTP introduces serialization, concurrency, lifecycle, and security boundaries while the browser introduces responsive pointer and accessibility requirements.

A hosted service would add unrelated decisions about identity, abuse, deployment, TLS, persistence, and privacy. Phase 10 keeps the server on loopback and ships one local executable so those product choices do not distort the first API contract.

## Phase 10 Outcome

`sudoku web` starts a loopback HTTP server and serves an embedded browser application. Players can create a puzzle by difficulty or puzzle string, enter values and notes, use undo/redo and hints, toggle automatic candidates, recover interrupted sessions, and discard or export a game.

The Go process remains authoritative. Browser state contains presentation preferences only; all Sudoku state transitions pass through `game.Game.Apply`, and all rendering starts from detached snapshots. The line CLI, TUI, serialized gameplay format, and recovery security guarantees remain compatible.

## Delivery Structure

1. **Application and API boundary**
   - Add a transport-independent session registry around `game.Game` with opaque IDs and monotonic revisions.
   - Add strict `/api/v1` JSON models for session creation, snapshots, hint preview, actions, conflicts, and discard.
   - Keep generation, fallback, and solver policy injected from command wiring rather than duplicated in handlers.

2. **Durability and local security**
   - Persist accepted web mutations through the validated private recovery store.
   - Restrict Phase 10 to loopback, same-origin browser requests, bounded JSON, safe HTTP timeouts, private logging, and clean shutdown.
   - Preserve recoverable sessions across browser refresh and server restart without accepting arbitrary host paths.

3. **Embedded browser frontend**
   - Add framework-free TypeScript, semantic HTML, and CSS compiled to embedded assets.
   - Support responsive keyboard, pointer, and touch play with visible focus, notes, candidates, hints, history, confirmations, and accessible status.
   - Keep UI preferences separate from authoritative session state.

4. **Black-box verification and documentation**
   - Exercise the built server through HTTP and a real browser with isolated XDG roots.
   - Cover stale revisions, restart recovery, invalid input, request limits, responsive layouts, accessibility, and browser console health.
   - Capture desktop and mobile screenshots and keep all existing CLI/TUI scenarios green.

## Exit Criteria

- `sudoku web` serves one self-contained application from a loopback-only address with no runtime CDN or cloud dependency.
- API version 1 has strict bounded requests, stable errors, opaque session IDs, per-session serialization, and revision conflicts that prevent silent stale writes.
- Every accepted action uses the game-engine contract and returns an authoritative detached snapshot; browser code contains no duplicate Sudoku rules.
- Accepted web mutations are recoverable after process restart, and discard removes only the selected record.
- The browser supports values, notes, candidates, hints, history, reset, new/discard, and session export across keyboard, pointer, touch, desktop, and narrow layouts.
- Semantic roles, labels, focus behavior, non-color distinctions, reduced motion, and status announcements pass accessibility checks.
- API handler tests, browser tests, built-binary HTTP/browser E2E, build, vet, lint, diff checks, documentation audit, existing CLI/TUI E2E, and CI pass.

## Deferred Work

Remote binding, hosted deployment, TLS, accounts, authentication, analytics, cloud synchronization, shared games, cross-device merge, mobile-native clients, localization, and multi-user collaboration remain outside Phase 10. Each deferred capability changes the threat model or product contract and requires a separately reviewed design.

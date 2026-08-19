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
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Roadmap

Phase 10 adds a client-neutral local HTTP API after the reusable engine, complete sessions, TUI, candidate assistance, and crash recovery have stabilized. Browser UX belongs in a separate TypeScript project; this repository focuses only on the Go backend contract, durability, and security boundary.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/web-api.md` | Canonical API resources, revisions, recovery, client access, and local security boundary |
| `.aidoc/designs/game-engine.md` | Stable actions, snapshots, hints, and serialization reused by the API |
| `.aidoc/designs/e2e-test-scenarios.md` | Existing compatibility contract and Phase 10 backend acceptance scenarios |

## Why Phase 10 Exists

The game engine already supports multiple terminal presentations without leaking UI concerns, and private recovery protects long-running local sessions. An HTTP boundary is the next useful stress test because it introduces serialization, concurrency, lifecycle, and security concerns while proving that a separately maintained client can consume the engine safely.

A bundled browser application would mix backend and UX release cycles, dependencies, tests, and product decisions. Phase 10 keeps the Go repository client-neutral so a separate TypeScript project can evolve its interaction model without embedding frontend build tooling or assets into the Sudoku binary.

## Phase 10 Outcome

`sudoku api` starts a loopback-only HTTP server. Clients can create a puzzle by difficulty or puzzle string, inspect authoritative snapshots, enter values and notes, use undo/redo and hints, recover interrupted sessions, and discard or export a game.

The Go process remains authoritative. Every Sudoku state transition passes through `game.Game.Apply`, and every response derives from detached snapshots. The line CLI, TUI, serialized gameplay format, and recovery security guarantees remain compatible.

## Delivery Structure

1. **Application and API boundary**
   - Add a transport-independent session registry around `game.Game` with opaque IDs and monotonic revisions.
   - Add strict `/api/v1` JSON models for session creation, snapshots, hint preview, actions, conflicts, export, and discard.
   - Keep generation, fallback, and solver policy injected from command wiring rather than duplicated in handlers.

2. **Durability and local security**
   - Persist accepted API mutations through the validated private recovery store.
   - Restrict Phase 10 to loopback, bounded JSON, safe HTTP timeouts, private logging, and clean shutdown.
   - Preserve recoverable sessions across process restarts without accepting arbitrary host paths.

3. **External-client boundary**
   - Keep API models independent from Go engine structs and frontend-specific view models.
   - Disable browser cross-origin access by default; allow only explicitly configured exact loopback origins.
   - Publish stable machine-readable errors and a versioned schema that a separate TypeScript client can consume.

4. **Black-box verification and documentation**
   - Exercise the built server through HTTP with isolated XDG roots.
   - Cover stale revisions, restart recovery, invalid input, request limits, origin policy, and concurrent sessions.
   - Keep all existing CLI/TUI scenarios green; browser rendering and accessibility tests belong to the frontend project.

## Exit Criteria

- `sudoku api` starts a backend-only loopback server and does not embed, build, open, or serve frontend assets.
- API version 1 has strict bounded requests, stable errors, opaque session IDs, per-session serialization, and revision conflicts that prevent silent stale writes.
- Every accepted action uses the game-engine contract and returns an authoritative detached snapshot; API code contains no duplicate Sudoku rules.
- Accepted API mutations are recoverable after process restart, and discard removes only the selected record.
- Cross-origin browser access is disabled by default and limited to explicit exact loopback origins when enabled.
- Handler tests, built-binary HTTP E2E, build, vet, lint, diff checks, documentation audit, existing CLI/TUI E2E, and CI pass.

## Deferred Work

The browser frontend, frontend hosting, remote binding, hosted backend deployment, TLS, accounts, authentication, analytics, cloud synchronization, shared games, cross-device merge, mobile-native clients, localization, and multi-user collaboration remain outside this repository's Phase 10 scope. Each capability changes the threat model or product contract and requires a separately reviewed design in its owning project.

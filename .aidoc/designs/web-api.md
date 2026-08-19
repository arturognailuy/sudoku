---
domain: Designs
status: Active
entry_points:
  - cmd/root.go
  - cmd/session.go
  - game/contract.go
  - recovery/recovery.go
dependencies:
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/background-autosave.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Local HTTP API

Phase 10 adds a versioned, client-neutral HTTP boundary around the existing game engine without adding frontend code or turning Sudoku into a hosted service. A local `sudoku api` process remains authoritative for gameplay, recovery, validation, and concurrency while separately released clients own presentation.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/game-engine.md` | Canonical state, action, hint, and serialization contract |
| `.aidoc/designs/background-autosave.md` | Private recovery storage available to API sessions |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box HTTP acceptance coverage |
| `.aidoc/architecture/guidelines.md` | Package dependency boundaries |

## Why the API Is Client-Neutral

The first HTTP boundary proves that `game.Game` supports non-terminal clients before hosting introduces authentication, accounts, deployment, abuse controls, and remote data policy. Keeping frontend source and build tooling in a separate TypeScript project lets backend and UX contracts evolve independently.

The local boundary still behaves like a real API: transport models are independent from Go engine structs, errors and revisions are machine-readable, and handlers are tested through HTTP. Later hosting requires separate security and product review rather than exposing these loopback endpoints unchanged.

## Process and Package Boundaries

`cmd/api.go` owns command flags, dependency construction, startup output, signal shutdown, and listener lifecycle. `webapi` owns HTTP routing, request limits, JSON translation, session lookup, revision checks, origin policy, and error envelopes. The Sudoku repository contains no frontend asset package, browser launcher, TypeScript toolchain, or SPA routing.

Each active API session owns one `game.Game`, one opaque random session ID, one monotonically increasing revision, and one recovery record. A registry serializes access per session; different sessions may proceed independently. `game.Game` remains non-concurrent and unaware of HTTP.

Session creation reuses existing puzzle input, difficulty generation, solver configuration, and database fallback policies through dependencies wired by `cmd`. The API must not import `cmd` or duplicate generation rules.

## Version 1 Resource Model

All JSON endpoints live below `/api/v1`; `/healthz` is outside the API namespace. Version 1 exposes:

- `GET /api/v1/sessions` lists active and recovered sessions using bounded non-payload summaries;
- `POST /api/v1/sessions` creates a game from a difficulty or puzzle string;
- `POST /api/v1/sessions/import` restores bounded versioned session bytes supplied by a client;
- `GET /api/v1/sessions/{id}` returns the current detached snapshot and revision;
- `GET /api/v1/sessions/{id}/hint` previews a structured hint without mutation;
- `GET /api/v1/sessions/{id}/export` returns versioned session bytes;
- `POST /api/v1/sessions/{id}/actions` submits one typed engine action;
- `DELETE /api/v1/sessions/{id}` explicitly discards the session and recovery record.

Create requests use a tagged source object so difficulty and puzzle input cannot conflict. Import and export transfer bounded validated bytes; no endpoint accepts or returns an arbitrary host filesystem path.

API snapshots use explicit JSON fields for givens, visible values, invalid markers, manual notes, legal candidates, status, and undo/redo availability. Rows and columns are numbered 1 through 9 at the transport boundary. API models do not expose Go type names, internal history records, frontend view models, or the version 1 persistence document.

## Actions, Revisions, and Errors

Action requests contain `expected_revision`, an action kind, and only the fields required by that action. The API translates accepted action kinds to `game.Action`; unknown kinds or extra fields fail before the engine is called. Presentation preferences such as theme, focus, or candidate visibility remain client-owned and have no API action.

Every accepted mutation increments the session revision exactly once and returns the resulting snapshot, action result, and new revision. Rejected actions leave both game state and revision unchanged. A stale `expected_revision` returns HTTP `409 Conflict` with the current revision and snapshot so delayed clients cannot silently overwrite newer state.

Expected failures use a stable envelope with a code, human-readable message, and optional field or cell context. Invalid JSON and transport validation use `400`; absent sessions use `404`; stale revisions use `409`; oversized bodies use `413`; engine rejections use `422`; unexpected failures use `500` without leaking paths or internal details.

## Recovery and Lifecycle

Accepted mutations persist the newest serialized session through the private recovery store before the success response is committed. A persistence failure returns an error and must not claim durable success; the in-memory engine is reconstructed from the last durable bytes when necessary to preserve the contract.

Server startup discovers valid records from a dedicated API recovery namespace and exposes bounded summaries through the session collection. The TUI recovery namespace remains separate, so simultaneous frontends cannot adopt the same record; explicit import/export is the transfer boundary. Explicit discard deletes one API record, while graceful server shutdown retains active records for restart.

Only one `sudoku api` process may own the API recovery namespace at a time. Startup acquires an exclusive process-lifetime lock and fails clearly rather than allowing two registries to write the same records.

## Local Security and Client Access

`sudoku api` listens on `127.0.0.1` by default and Phase 10 rejects non-loopback bind addresses. The server validates `Host`, requires JSON content types for mutations, limits request/header/body sizes, applies read/write/idle timeouts, logs metadata without puzzle or session payloads, and shuts down cleanly.

Browser cross-origin access is disabled by default. A repeatable `--allow-origin` option may enable exact `http://127.0.0.1:<port>`, `http://localhost:<port>`, or `[::1]` development origins; wildcards, `null`, non-loopback hosts, and credentialed CORS are rejected. Preflight responses advertise only required methods and headers, and mutating requests from browsers must match the configured origin exactly.

Opaque session IDs prevent accidental collisions but are not authentication. Remote binding, TLS termination, authentication, rate limiting, and deployment configuration are later work requiring a separate threat model.

## Verification

Handler tests exercise real HTTP requests, strict decoding, body limits, action translation, typed errors, revision conflicts, recovery failure, concurrent access, default CORS denial, and exact-origin allowlisting. Black-box tests start the built binary with isolated XDG roots, call the versioned API, restart the process, and confirm recovery without importing Go packages or relying on a frontend.

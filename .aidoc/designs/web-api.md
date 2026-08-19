---
domain: Designs
status: Active
entry_points:
  - cmd/root.go
  - cmd/session.go
  - game/contract.go
  - recovery/recovery.go
dependencies:
  - .aidoc/designs/browser-frontend.md
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/background-autosave.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Local Web API

Phase 10 adds a versioned HTTP boundary around the existing game engine without turning Sudoku into a hosted service. A local `sudoku web` process remains authoritative for gameplay, recovery, and validation while an embedded same-origin browser client handles presentation.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/browser-frontend.md` | Browser interaction, accessibility, and asset decisions |
| `.aidoc/designs/game-engine.md` | Canonical state, action, hint, and serialization contract |
| `.aidoc/designs/background-autosave.md` | Private recovery storage available to local web sessions |
| `.aidoc/designs/e2e-test-scenarios.md` | Black-box HTTP and browser acceptance coverage |
| `.aidoc/architecture/guidelines.md` | Package dependency boundaries |

## Why the API Is Local First

The first web boundary proves that `game.Game` supports a non-terminal frontend before hosting introduces authentication, accounts, deployment, abuse controls, and remote data policy. `sudoku web` therefore binds to loopback only, serves one self-contained application, and does not promise a remotely stable service.

The local-first boundary still behaves like a real API: transport models are independent from Go engine structs, errors and revisions are machine-readable, and handlers are tested through HTTP. Later hosting can reuse the application boundary after separate security and product review rather than exposing local endpoints unchanged.

## Process and Package Boundaries

`cmd/web.go` owns command flags, dependency construction, startup output, signal shutdown, and browser opening. `webapi` owns HTTP routing, request limits, JSON translation, session lookup, revision checks, and same-origin enforcement. `webui` owns embedded browser assets. Neither `webapi` nor `webui` contains Sudoku rules.

Each active API session owns one `game.Game`, one opaque random session ID, one monotonically increasing revision, and one recovery record. A registry serializes access per session; different sessions may proceed independently. `game.Game` remains non-concurrent and unaware of HTTP.

Session creation reuses the existing puzzle input, difficulty generation, solver configuration, and database fallback policies through dependencies wired by `cmd`. The API must not import `cmd` or duplicate generation rules.

## Version 1 Resource Model

All JSON endpoints live below `/api/v1`; static assets and `/healthz` are outside the API namespace. Version 1 exposes:

- `GET /api/v1/sessions` lists active and recovered sessions using bounded non-payload summaries;
- `POST /api/v1/sessions` creates a game from a difficulty or puzzle string;
- `POST /api/v1/sessions/import` restores bounded versioned session bytes uploaded by the client;
- `GET /api/v1/sessions/{id}` returns the current detached snapshot and revision;
- `GET /api/v1/sessions/{id}/hint` previews a structured hint without mutation;
- `GET /api/v1/sessions/{id}/export` downloads versioned session bytes;
- `POST /api/v1/sessions/{id}/actions` submits one typed engine action;
- `DELETE /api/v1/sessions/{id}` explicitly discards the session and recovery record.

Create requests use a tagged source object so difficulty and puzzle input cannot conflict. Session restore uses recovery discovery on server startup. Import and export transfer bounded validated bytes through the browser; no endpoint accepts or returns an arbitrary host filesystem path.

API snapshots use explicit JSON fields for givens, visible values, invalid markers, manual notes, legal candidates, status, and undo/redo availability. Rows and columns are numbered 1 through 9 at the transport boundary. API models do not expose Go type names, internal history records, or the version 1 persistence document.

## Actions, Revisions, and Errors

Action requests contain `expected_revision`, an action kind, and only the fields required by that action. The API translates accepted action kinds to `game.Action`; unknown kinds or extra fields fail before the engine is called.

Every accepted mutation increments the session revision exactly once and returns the resulting snapshot, action result, and new revision. Rejected actions leave both game state and revision unchanged. A stale `expected_revision` returns HTTP `409 Conflict` with the current revision so duplicate tabs or delayed requests cannot silently overwrite newer state.

Expected failures use a stable envelope with a code, human-readable message, and optional field/cell context. Invalid JSON and transport validation use `400`; absent sessions use `404`; stale revisions use `409`; oversized bodies use `413`; engine rejections use `422`; unexpected failures use `500` without leaking paths or internal details.

## Recovery and Lifecycle

Accepted mutations persist the newest serialized session through the private recovery store before the success response is committed. A persistence failure returns an error and must not claim durable success; the in-memory engine may be reconstructed from the last durable bytes to preserve the API contract.

Server startup discovers valid records from a dedicated web recovery namespace and exposes bounded summaries through the session collection; list responses never include serialized games or filesystem paths. The TUI recovery namespace remains separate, so simultaneous frontends cannot adopt and overwrite the same record; explicit import/export is the transfer boundary. Explicit discard deletes the web record, while graceful server shutdown retains active records for restart. Existing private-path, bounded-write, validation, and 30-day retention guarantees still apply.

Only one `sudoku web` process may own the web recovery namespace at a time. Startup acquires an exclusive process-lifetime lock and fails clearly rather than allowing two registries to write the same records.

The browser URL may identify the current random session, but possession of an ID is not treated as remote authentication. Phase 10 has no accounts, sharing, cross-device sync, or multi-user editing.

## Local Security Boundary

`sudoku web` listens on `127.0.0.1` by default and Phase 10 rejects non-loopback bind addresses. The server validates `Host` and mutating-request `Origin`, serves no permissive CORS headers, requires JSON content types, limits request/header/body sizes, applies read/write/idle timeouts, and shuts down cleanly.

Static assets use a restrictive Content Security Policy and no third-party origins. The server logs request metadata without puzzle/session payloads. Remote binding, TLS termination, authentication, rate limiting, and deployment configuration are later work requiring a separate threat model.

## Verification

Handler tests exercise real HTTP requests, strict decoding, body limits, action translation, typed errors, revision conflicts, recovery failure, and concurrent access. Black-box tests start the built binary with isolated XDG roots, call the versioned API, restart the process, and confirm recovery without importing Go packages.

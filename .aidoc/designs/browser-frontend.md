---
domain: Designs
status: Active
entry_points:
  - tui/model.go
  - tui/render.go
  - game/contract.go
dependencies:
  - .aidoc/designs/web-api.md
  - .aidoc/designs/tui-frontend.md
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/e2e-test-scenarios.md
---

# Browser Frontend

Phase 10 adds a responsive, keyboard- and pointer-accessible browser client for the local web API. The browser renders detached server snapshots and submits typed intents; it never reimplements Sudoku validation, candidate calculation, history, or hints.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/web-api.md` | Authoritative session, action, revision, recovery, and security boundary |
| `.aidoc/designs/tui-frontend.md` | Existing interaction semantics and visual-state distinctions |
| `.aidoc/designs/game-engine.md` | Canonical values, notes, candidates, status, and hint concepts |
| `.aidoc/designs/e2e-test-scenarios.md` | Browser black-box acceptance coverage |

## Why a Separate Browser Design Exists

The TUI proves the engine can support a rich frontend, but terminal rendering constraints should not become browser architecture. The browser preserves gameplay semantics while adopting semantic HTML, responsive layout, pointer input, focus management, and browser accessibility conventions.

The first browser client remains deliberately local and single-player. Phase 10 validates frontend reuse and API quality; it does not add accounts, advertising, analytics, cloud persistence, social features, or remote deployment.

## Asset and Dependency Strategy

Browser source uses TypeScript, semantic HTML, and CSS with no UI framework. A small build step produces hashed static assets that `webui/assets.go` embeds into the Go binary, preserving a single executable at runtime. Dependencies are development-only, pinned by a lockfile, and justified by type checking, bundling, and browser automation.

The repository keeps browser source as the canonical asset. CI builds assets before the Go embed/build gate and fails if generated output is stale; runtime never fetches a CDN script, font, icon, or stylesheet. The server provides an SPA fallback only for browser routes and never for `/api/` paths.

## Interaction Model

The board is a semantic 9×9 grid with one focusable cell at a time. Arrow keys move focus without wrapping; digits enter values or notes according to the visible mode; Delete or Backspace clears; standard buttons expose note mode, automatic candidates, undo, redo, hint, reset, save/export, and new/discard actions.

Pointer and touch select a cell and then use an on-screen digit pad. Keyboard-only and pointer flows submit the same API actions. The browser disables only controls that the current snapshot proves unavailable and never predicts whether an engine action will succeed.

Hint preview remains non-mutating and shows technique plus explanation before an explicit apply action. Destructive reset, discard, and navigation away from unsaved work require clear confirmation. Automatic candidates remain a local display preference; manual notes remain authoritative engine state.

## Rendering and Accessibility

The layout scales from a compact phone viewport to desktop without horizontal scrolling. The board keeps strong 3×3 boundaries and distinguishes givens, player values, invalid entries, focus, peers, manual notes, and automatic candidates without relying on color alone.

Every control has an accessible name and visible focus indicator. Status and errors use a polite live region; completion uses a deliberate announcement without trapping focus. The board exposes row, column, value, given/editable state, invalid state, and notes to assistive technology without causing 81 simultaneous tab stops.

Dark, light, high-contrast, reduced-motion, and narrow-viewport behavior follow browser preferences. Animations are optional decoration and disabled by `prefers-reduced-motion`. User-facing text stays centralized so later localization does not require extracting strings from engine or transport code.

## Client State and Concurrency

The client stores only presentation state locally: focused cell, note mode, candidate visibility, open dialog, theme preference, and the latest server revision. Authoritative values, invalid markers, manual notes, status, and history always come from API snapshots.

The browser sends one mutation at a time. A `409 Conflict` replaces stale client data with the returned current revision/snapshot and explains that another tab changed the game; the client never retries a mutation automatically. Network or server errors keep the visible last confirmed snapshot and provide an explicit retry path.

A current random session ID may be retained in the same-origin URL or storage so refresh can reconnect. The startup screen uses the server's bounded session summaries to offer recovered games and explicit discard without loading every serialized game. Stored presentation data contains no serialized game or analytics payload; recovery after server restart is driven by validated server records rather than browser storage.

## Verification

Unit tests cover transport-model decoding and presentation reducers without reproducing engine rules. Browser E2E tests run against the built `sudoku web` binary and verify creation, values, notes, candidates, hints, history, conflict recovery, restart recovery, responsive layout, keyboard/pointer parity, focus order, accessible names, and no unexpected console errors.

Browser verification captures desktop and mobile screenshots for dark, light, and high-contrast states. Existing line CLI and TUI black-box scenarios remain green so the web frontend cannot redefine shared engine behavior.

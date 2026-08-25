---
domain: Designs
status: Active
entry_points:
  - main.go
  - cmd/root.go
dependencies:
  - .aidoc/INDEX.md
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/web-api.md
---

# E2E Test Scenarios

End-to-end test scenarios that treat the Sudoku CLI as a black box. Each scenario describes what a regular user would do, the expected behavior, and how to verify it.

These scenarios are designed to be run manually or via shell scripts against a built binary. They are **not** Go test files — E2E testing operates outside the codebase, treating it as a black box.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/INDEX.md` | Discovery index and project reading chains |
| `.aidoc/designs/game-engine.md` | Engine actions and state that frontends expose to users |
| `.aidoc/designs/cli-sessions.md` | Manual-note rendering and persistence transport contract |
| `AGENT.md` | Required verification discipline for feature and bug-fix PRs |
| `.aidoc/designs/roadmap.md` | Stabilization sequencing and deterministic-test constraints |

## Automation Map

| Boundary | Automated entry point | CI status |
|----------|-----------------------|-----------|
| HTTP API lifecycle | `scripts/e2e_api.py` | Independent mandatory `api-e2e` job |
| TUI, sessions, and recovery | `scripts/e2e_tui.py` | Independent mandatory `tui-e2e` job |
| Root line CLI, sessions, calibration, generate, import, and SQLite composition | `scripts/e2e_cli.py` | Independent mandatory `cli-e2e` job |
| OpenAPI contract and compatibility | `scripts/check-api-contract.sh` | Independent mandatory `api-contract` job |
| Generator behavior | `generator/*_test.go` | Unit and race jobs |
| Deterministic difficulty classification | `solver/classify_test.go` | Unit and race jobs; composed through generate/import black-box flows |
| Command storage and reporting | `cmd.batchGenerateWith` tests with fixed puzzles | Unit and race jobs |
| Package coverage evidence | `scripts/coverage_report.py` | Unit job summary; no pass/fail threshold |
| Probabilistic database fallback | Section 6.2 | Manual; deterministic package boundaries cover DB lookup |
| Deferred database behavior | Section 12 | Deferred until calibration |

Every automated black-box entry point builds or receives the repository binary and owns isolated temporary XDG roots. Package tests do not count as E2E coverage; they keep deterministic seams narrow while the scripts verify complete user-facing composition.

Run the line-oriented boundary locally with:

```bash
go build -o sudoku .
python3 scripts/e2e_cli.py ./sudoku
```

The harness uses fixed puzzle fixtures for gameplay and import assertions. Its generation smoke test bounds real generation to one round and one millisecond; it verifies command/worker/database composition without asserting a random difficulty result. Section 6.2 remains manual because the public binary has no deterministic way to force the fallback branch.

## Prerequisites

```bash
go build -o sudoku .
export SUDOKU_E2E_DIR=$(mktemp -d)
export XDG_DATA_HOME=$SUDOKU_E2E_DIR/data
export SUDOKU_DB=$SUDOKU_E2E_DIR/test-puzzles.db
```

Root play commands auto-store puzzles under `XDG_DATA_HOME`; generate and import commands use `--db $SUDOKU_DB`. These settings isolate every database touched by the scenarios.

---

## 1. Interactive Play

### 1.1 Input a Known Puzzle
```bash
echo "quit" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Board is displayed. Game starts. Exits on "quit".

### 1.2 Input a Puzzle Using Dots Notation
```bash
echo "quit" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Dots (`.`) are treated as empty cells. Board displays correctly.

### 1.3 Input a Puzzle Using Zeros Notation
```bash
echo "quit" | ./sudoku --input "003020600900305001001806400008102900700000008006708200002609500800203009005010300"
```
**Expected:** Zeros are converted to empty cells. Same puzzle as 1.2.

### 1.4 Generate a Puzzle by Difficulty Level
```bash
echo "quit" | ./sudoku --level easy
```
**Expected:** A puzzle is generated (may show mismatch warning if best-effort misses). Game starts.

### 1.5 Invalid Input (Too Short)
```bash
./sudoku --input "123"
```
**Expected:** Error message about invalid puzzle string. Non-zero exit or error output.

### 1.6 Invalid Level Flag
```bash
./sudoku --level banana
```
**Expected:** Usage error shown with valid difficulty levels listed.

### 1.7 Help Flag
```bash
./sudoku --help
```
**Expected:** Usage printed with subcommands (`calibrate`, `generate`, `import`, `tui`) and flags listed.

### 1.8 Backward Compatibility: `--input` Flag
```bash
echo "quit" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Works the same as before the cobra migration.

### 1.9 Backward Compatibility: `--level` Flag
```bash
echo "quit" | ./sudoku --level easy
```
**Expected:** Works the same as before the cobra migration.

---

## 2. Game Commands

The game-command scenarios verify the stable engine boundary through the real terminal frontend: `cli.Controller` renders detached snapshots and submits typed actions while preserving the established command output and behavior.

For these scenarios, start a game with a known puzzle:
```bash
./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```

### 2.1 Add a Value (`add` / `a` / bare digits)
**Input:** `add 1 1 4` or `a 1 1 4` or `1 1 4`
**Expected:** Cell (1,1) is set to 4. Board updates.

### 2.2 Clear a Value (`clear` / `d`)
**Input:** After adding a value at (1,1): `clear 1 1` or `d 1 1`
**Expected:** Cell (1,1) is reset to empty.

### 2.3 Undo (`undo` / `u`)
**Input:** After adding a value: `undo` or `u`
**Expected:** Last move is reversed. Cell returns to previous state.

### 2.4 Redo (`redo` / `r`)
**Input:** After undo: `redo` or `r`
**Expected:** Undone move is re-applied, including whether the value is valid or invalid.

### 2.5 Check Board (`check` / `c`)
**Input:** `check` or `c`
**Expected (correct board):** "The current board is correct."
**Expected (incorrect board):** "You have entered incorrect value(s)."

### 2.6 Repair Board (`repair` / `f`)
**Input:** After entering wrong values: `repair` or `f`
**Expected:** All incorrect inputs are removed. Board returns to a valid state.

### 2.7 Reset Board (`reset` / `e`)
**Input:** After making several moves: `reset` or `e`
**Expected:** Board returns to the original problem state. All user inputs cleared.

### 2.8 Hint (`hint` / `i`)
**Input:** `hint` or `i`
**Expected:** A correct value is filled into a cell. The technique and reason are displayed.

### 2.9 Solve (`solve` / `s`)
**Input:** `solve` or `s`
**Expected:** The complete solution is displayed. "Congratulations" message shown.

### 2.10 Quit (`quit` / `q`)
**Input:** `quit` or `q`
**Expected:** Game exits cleanly.

### 2.11 Shorthand Digit Input
**Input:** `1 2 3` (no `add` prefix)
**Expected:** Treated as `add 1 2 3`. Cell (1,2) is set to 3.

### 2.12 Divergent Input Truncates Redo History
**Input:** Add a value, undo it, add a different value, then run `redo`.
**Expected:** The different value remains in the cell and the abandoned value does not reappear because the new action replaced the old redo branch.

---

## 3. Difficulty Measurement CLI

### 3.0 Prepare a Canonical Corpus
Create a candidate version 2 manifest with a name and ordered `puzzles` array. Include unique IDs, puzzle strings, source category and identifier, license and redistribution status, collection method, and exploratory/held-out split; generated records also carry generator metadata. Run `sudoku calibrate prepare --input <candidate> --output <manifest>`.

**Expected:** Zero notation and whitespace are normalized to 81-cell dot notation, each record receives a matching SHA-256 content hash, order is preserved, and an existing output is never overwritten. Duplicate normalized puzzle content or incomplete provenance exits non-zero.

### 3.1 Start and Resume an Immutable Corpus
Use the prepared version 2 manifest. Run `sudoku calibrate --manifest <path> --output <directory>` twice with the same paths.

**Expected:** The first run appends one observation per puzzle and writes a manifest-bound checkpoint plus JSON and Markdown reports. The reports include reproducibility count, source/split outcome groups, tier distributions, neighboring score overlap where both tiers exist, external agreement, and generated-target measurements. The second run reports zero new observations, leaves `observations.jsonl` unchanged, and keeps `report.json` deterministic. The built-binary harness also verifies the completed checkpoint index and representative stratified fields.

### 3.2 Reject a Changed Manifest
Complete a measurement run, then change the manifest name, puzzle order, IDs, or puzzle text and reuse the existing output directory.

**Expected:** The command exits non-zero because existing observations are bound to the original exact manifest SHA-256. Existing observations remain unchanged.

## 4. Puzzle Generation CLI

### 4.1 Generate Help
```bash
./sudoku generate --help
```
**Expected:** Shows all flags: `-n`, `-d`, `-w`, `-t`, `--rounds`, `--db`.

### 4.2 Generate Easy Puzzles
```bash
./sudoku generate -n 5 -d easy --db $SUDOKU_DB
```
**Expected:** Generates 5 puzzles. Reports: generated count, stored count, duplicates, by-level breakdown. Some may be classified at a different difficulty than requested (mismatch is expected for easy/medium targets).

### 4.3 Generate Hard Puzzles
```bash
./sudoku generate -n 3 -d hard --db $SUDOKU_DB
```
**Expected:** Generates 3 puzzles. Each is classified and stored. Duration reported.

### 4.4 Generate with Parallel Workers
```bash
./sudoku generate -n 4 -d evil -w 4 --db $SUDOKU_DB
```
**Expected:** Uses 4 parallel goroutines. Completes generation. All puzzles stored.

### 4.5 Generate with Invalid Difficulty
```bash
./sudoku generate -d invalid
```
**Expected:** Error: "Invalid difficulty level". Exit code 1.

### 4.6 Generate with Count 0
```bash
./sudoku generate -n 0 -d easy
```
**Expected:** Error: "count must be positive". Exit code 1.

### 4.7 Generate with Custom DB Path
```bash
./sudoku generate -n 2 -d hard --db /tmp/custom-test.db
```
**Expected:** DB file created at `/tmp/custom-test.db`. Puzzles stored there.

### 4.8 Generate Dedup
```bash
./sudoku generate -n 10 -d evil --db $SUDOKU_DB
./sudoku generate -n 10 -d evil --db $SUDOKU_DB
```
**Expected:** Second run may report duplicates if the same normalized puzzles are generated.

---

## 5. Import CLI

### 5.1 Import Help
```bash
./sudoku import --help
```
**Expected:** Shows flags: `-f`, `--source`, `--db`.

### 5.2 Import Puzzles from File
Create `test-puzzles.txt`:
```
# Test puzzles
..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3..
003020600900305001001806400008102900700000008006708200002609500800203009005010300
```
```bash
./sudoku import -f test-puzzles.txt --source "test" --db $SUDOKU_DB
```
**Expected:** Puzzles are classified, normalized, and stored. The two lines represent the same puzzle (different notation), so one is stored and one is a duplicate.

### 5.3 Import with Invalid Lines
Create `mixed-puzzles.txt`:
```
..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3..
123456
abc
```
**Expected:** Valid puzzle stored. Invalid lines (too short, wrong characters) are skipped with stderr messages. Valid lines are processed.

### 5.4 Import with Source Label
```bash
./sudoku import -f puzzles.txt --source "top1465" --db $SUDOKU_DB
```
**Expected:** Custom source label ("top1465") is stored with each puzzle.

### 5.5 Import Missing File
```bash
./sudoku import -f nonexistent.txt
```
**Expected:** Error: "no such file". Exit code 1.

### 5.6 Import Empty / Comments-Only File
Create `empty.txt`:
```
# Only comments
# No puzzles here

```
```bash
./sudoku import -f empty.txt --db $SUDOKU_DB
```
**Expected:** Report shows 0 total lines, 0 stored. No errors.

### 5.7 Import Dedup
```bash
./sudoku import -f test-puzzles.txt --db $SUDOKU_DB
./sudoku import -f test-puzzles.txt --db $SUDOKU_DB
```
**Expected:** Second import reports all as duplicates.

---

## 6. Database and Fallback

### 6.1 Auto-Store on Play
```bash
echo "quit" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Puzzle is automatically stored in the DB at `~/.local/share/sudoku/puzzles.db`.

### 6.2 DB Fallback Path
1. Pre-populate the DB with easy puzzles: `./sudoku generate -n 20 -d easy --db $SUDOKU_DB`
2. Request an easy puzzle: `echo "quit" | ./sudoku --level easy --db $SUDOKU_DB`

**Expected:** If best-effort generation misses the target, the system falls back to the DB. If a match is found in the DB, no mismatch warning is shown. If the DB is also empty for that difficulty, the mismatch warning fires.

### 6.3 Mismatch Warning
```bash
echo "quit" | ./sudoku --level easy
```
**Expected (empty DB):** Best-effort likely misses easy target. Warning shown: "Requested difficulty: Easy. Generated puzzle difficulty: Medium/Hard. Enjoy!"

### 6.4 Multiple-Solution Puzzle Input
```bash
echo "quit" | ./sudoku --input "....7....6..195....98....6.8...6...34..8.3..17...2...6.6....28....419..5....8..79"
```
**Expected:** Warning about multiple solutions printed. Game still starts (plays with the first solution found).

---

## 7. Cobra Subcommand Structure

### 7.1 Root Command Shows Subcommands
```bash
./sudoku
```
**Expected:** Shows available commands: `generate`, `import`, `tui`, and usage for the default play mode.

---

## 8. Manual Notes and Durable Sessions

Use the known puzzle from section 2 and paths under `$SUDOKU_E2E_DIR`.

### 8.1 Toggle and Clear Manual Notes
```bash
printf 'n 1 1 1\nn 1 1 9\nx 1 1\nq\n' | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Notes 1 and 9 appear in fixed candidate positions, then `x` returns the board to compact rendering.

### 8.2 Peer Cleanup and Unified History
```bash
printf 'n 1 1 4\nn 1 2 4\nn 3 1 4\nn 2 2 4\nn 4 5 4\n1 1 4\nu\nr\nq\n' | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Setting (1,1) removes notes from the target and its row, column, and box peers but preserves the note at non-peer (4,5). Undo restores the value and all removed notes atomically; redo reapplies cleanup.

### 8.3 Save, Resume, and Preserve Redo
```bash
SESSION=$SUDOKU_E2E_DIR/session.json
printf 'n 1 1 5\n1 1 4\nu\nsave %s\nq\n' "$SESSION" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
printf 'r\nq\n' | ./sudoku --resume "$SESSION"

INVALID_SESSION=$SUDOKU_E2E_DIR/invalid-session.json
printf '1 1 5\nn 1 2 4\nsave %s\nq\n' "$INVALID_SESSION" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
printf 'c\nq\n' | ./sudoku --resume "$INVALID_SESSION"
```
**Expected:** The first restored board includes note 5 at (1,1), `redo` restores value 4, and the session file has mode `0600`. The second restore retains the invalid entry and note, and `check` reports the invalid board.

### 8.4 Reject Corrupt, Unsupported, and Oversized Sessions
```bash
printf '{bad json' > "$SUDOKU_E2E_DIR/corrupt.json"
printf '{"version":999}' > "$SUDOKU_E2E_DIR/unsupported.json"
head -c 1048577 /dev/zero > "$SUDOKU_E2E_DIR/oversized.json"
./sudoku --resume "$SUDOKU_E2E_DIR/corrupt.json"
./sudoku --resume "$SUDOKU_E2E_DIR/unsupported.json"
./sudoku --resume "$SUDOKU_E2E_DIR/oversized.json"
```
**Expected:** Every command exits non-zero before interactive play with a concise restore error. Source files remain unchanged.

### 8.5 Resume Flag Conflicts
```bash
./sudoku --resume "$SUDOKU_E2E_DIR/session.json" --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
./sudoku --resume "$SUDOKU_E2E_DIR/session.json" --level easy
```
**Expected:** Cobra rejects both combinations before reading or generating a puzzle.

### 8.6 Failed Save Preserves the Destination
```bash
mkdir "$SUDOKU_E2E_DIR/existing-destination"
printf 'save %s\nq\n' "$SUDOKU_E2E_DIR/existing-destination" | ./sudoku --input "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
```
**Expected:** Save reports an error, the existing destination remains a directory, and no `.sudoku-session-*` temporary file remains.

---

## 9. Full-Screen TUI

The TUI scenarios require a pseudo-terminal. The standard-library harness exercises the built binary rather than package internals:

```bash
python3 scripts/e2e_tui.py ./sudoku
```

### 9.1 Startup and Resize
**Action:** Start `sudoku tui --input <known puzzle>` in a terminal of at least 72×40, then resize below the minimum and back.
**Expected:** The centered board uses clean digits, fixed-position notes, strong 3×3 boundaries, a strong focus background, and subtle peer backgrounds. A small terminal displays only the minimum-size instruction; resizing restores the unchanged game.

### 9.2 Values, Notes, and History
**Action:** Move with arrows or `h`/`j`/`k`/`l`, enter a value, toggle note mode with `n`, enter a note, then use `u` and `r`.
**Expected:** Focus stops at board edges. Keys submit engine actions, note cleanup follows peer rules, and undo/redo restore whole transitions.

### 9.3 Help, Hint Preview, and Apply
**Action:** Press `?`, inspect and close the keyboard-help overlay, press `i`, inspect the technique/reason, then press Enter.
**Expected:** Help does not mutate the board. Hint preview does not mutate the board; Enter applies that hint through `game.ApplyHint` and marks the session dirty.

### 9.4 Explicit Save and Safe Quit
**Action:** Change a cell, press `q` and decline, press `S`, enter a path, then quit. Resume with `sudoku tui --resume <path>`.
**Expected:** Dirty quit asks for confirmation. Save atomically creates a mode-0600 session, clears the dirty marker, and resumed values, notes, and history match.

### 9.5 Restore Rejection and CLI Compatibility
**Action:** Run TUI restore against corrupt, unsupported, and oversized files; then run scenarios 1.1, 7.3, and 7.4 against root play.
**Expected:** Invalid restores fail before entering full-screen mode and preserve source files. The line-oriented CLI output and persistence behavior remain compatible.

### 9.6 Themes and No-Color Accessibility
**Action:** Start the known puzzle with the default environment, `SUDOKU_THEME=light`, and `NO_COLOR=1`.
**Expected:** Dark and light runs use their deterministic palettes. The no-color run emits no color codes while bold givens, underlined invalid values, reverse-video focus, faint peers/automatic candidates, bold manual notes, clean digits, and heavy box boundaries preserve semantic distinctions.

### 9.7 Automatic-Candidate Toggle
**Action:** Start TUI, press `a`, and press `a` again.
**Expected:** Legal candidates appear only while enabled, board geometry remains stable, the status changes between `AUTO ON` and `AUTO OFF`, and no dirty marker is created.

### 9.8 Candidate Transition Refresh
**Action:** With candidates enabled, set and clear a value, undo and redo, apply a hint, repair an invalid entry, solve, and reset.
**Expected:** Every displayed candidate set follows the accepted solver-safe board without an independent candidate action or history entry.

### 9.9 Automatic and Manual Coexistence
**Action:** Add legal and stale manual notes with candidates enabled under dark, light, and `NO_COLOR` themes.
**Expected:** Manual styling wins at overlapping positions, stale manual notes remain visible, and automatic candidates remain distinguishable without color alone.

### 9.10 Candidate Save and Resume
**Action:** Save with candidates enabled, then resume the session.
**Expected:** Session bytes contain no candidate preference or derived data, and the resumed TUI starts with `AUTO OFF`.

### 9.11 Invalid-Entry Handling
**Action:** Enter an invalid value with candidates enabled, then clear or repair it.
**Expected:** The invalid cell shows its visible value, its peers ignore that value during candidate calculation, and candidates reappear in the cell after clear or repair.

---

## 10. TUI Autosave and Recovery

Use an isolated state root with `export XDG_STATE_HOME=$SUDOKU_E2E_DIR/state`.

### 10.1 Crash Recovery and Durable Record Discovery
**Action:** Start `sudoku tui --input <known puzzle>`, make a gameplay change, wait past the one-second debounce, and terminate the process abnormally. Then run plain `sudoku tui`.
**Expected:** One mode-`0600` record exists under a mode-`0700` recovery directory. Plain startup offers the changed game; selecting it restores the value without relying on the old process ID.

### 10.2 Concurrent Recovery Selection
**Action:** Interrupt two changed TUI instances, then run plain `sudoku tui` and move through the recovery list.
**Expected:** Each instance has a different random record. The list is newest-first, either game can be selected, and `d` discards only the selected record.

### 10.3 Explicit Sources and Opt-Out
**Action:** Leave a valid recovery record, then start with each of `--input`, explicit `--level`, `--resume`, and `--no-autosave`.
**Expected:** Every intentional source bypasses recovery selection. `--no-autosave` also creates no recovery record after mutations.

### 10.4 Recovery Cleanup
**Action:** Recover a game and explicitly save it; separately confirm dirty quit/discard and cleanly quit an unmodified recovered game.
**Expected:** Each successful cleanup path removes the active recovery record. `Ctrl-C` and abnormal termination retain the latest successful record.

### 10.5 Recovery Failure Is Non-Fatal
**Action:** Use an inaccessible or unsafe state path and mutate the game.
**Expected:** Gameplay continues with a persistent concise autosave warning. A later mutation retries recovery after the path is fixed.

## 11. HTTP API Backend

Run the automated black-box lifecycle against the built backend with isolated state roots:

```bash
python3 scripts/e2e_api.py ./sudoku
```

The harness calls the running `sudoku api` process rather than importing Go handlers. It covers startup safety, process-lock exclusion, health, strict input, exact-origin CORS configuration, exact bearer authentication, lifecycle operations, revision conflict, export/import, restart recovery, and discard. Handler and command tests complement the black-box lane with malformed-input envelopes, preflight method/header rejection, concurrent session isolation, allowed-origin validation, and lock ownership. The scenarios below define the broader acceptance contract; frontend behavior remains in the separate client project.

### 11.1 Startup, Binding, and Health
**Action:** Start `sudoku api` with isolated XDG roots using the default listener and an explicit network listener, call `/healthz`, and request an unknown `/api/` path.
**Expected:** The default listener is loopback, the explicit listener requires authentication configuration, startup prints the listening address, health reports healthy, the unknown route returns a stable JSON `404`, and no frontend assets or SPA fallback are served.

### 11.2 Session Creation and Strict Input
**Action:** Create sessions by difficulty and puzzle string, then send conflicting sources, unknown fields, malformed JSON, wrong content types, and oversized bodies.
**Expected:** Valid requests return opaque IDs, revision zero, and authoritative snapshots. Invalid requests return bounded stable errors without creating sessions or leaking host details.

### 11.3 OpenAPI Contract and Runtime Conformance
**Action:** Validate and lint `api/openapi.yaml` with the pinned Redocly CLI, regenerate the strict Go boundary and confirm a clean diff, compare the contract with the target branch using `oasdiff`, then execute every declared operation and representative examples against the built server.
**Expected:** The OpenAPI 3.1.1 contract is valid and lint-clean, generation is reproducible, no unapproved breaking change is reported, documented examples match runtime responses, and no implemented route is absent from the contract.

### 11.4 Actions, Hints, and Revision Conflicts
**Action:** Enter values and notes, preview/apply a hint, undo/redo, and submit two actions with the same expected revision.
**Expected:** Accepted mutations increment revisions once and match engine semantics. Hint preview is read-only. The delayed mutation returns `409` with current state and never overwrites the accepted action.

### 11.5 Restart Recovery and Discard
**Action:** Mutate two API sessions, stop and restart the server, reconnect to both, then discard one.
**Expected:** Both sessions restore from separate private records with complete values, notes, and history. Discard removes only the selected record, and another restart retains the other session.

### 11.6 Concurrent Sessions and Process Lock
**Action:** Mutate separate sessions concurrently, submit concurrent actions to one session, and start a second `sudoku api` process against the same state root.
**Expected:** Different sessions proceed independently, one session remains revision-ordered, and the second process fails clearly without modifying recovery records.

### 11.7 Origin Policy
**Action:** Send browser-style preflight and mutation requests with no configured origin, exact allowed local and remote HTTP/HTTPS origins, a different port, `null`, a wildcard, and a path-bearing origin.
**Expected:** Cross-origin browser access is denied by default. Only exact configured origins succeed; responses never enable wildcard CORS, and authenticated preflight permits only the required authorization header.

### 11.8 Authentication and Remote Access
**Action:** Bind to a non-loopback address with no token, then with a configured token; call API resources with a missing, incorrect, and correct bearer credential.
**Expected:** Unsafe startup is rejected. Missing and incorrect credentials receive bounded unauthorized responses, the correct credential succeeds, and logs never contain the token.

### 11.9 Existing Frontend Compatibility
**Action:** Run all applicable root CLI, TUI, serialization, candidate, and recovery scenarios after API tests.
**Expected:** Existing output, actions, session bytes, recovery behavior, and terminal rendering remain compatible.

## 12. Deferred Database Scenarios

These scenarios begin only after stabilization and difficulty calibration, as sequenced in `.aidoc/designs/roadmap.md`:

- **Large import progress indicator:** Import 150+ puzzles → progress indicator fires every 100 puzzles.
- **Minimum-clues guard:** Import a puzzle with fewer than 17 clues → rejected or warned (prevents solver hang on near-empty boards).
- **Played tracking:** Mark puzzles as played → DB query skips played puzzles.
- **Concurrent DB access:** Multiple generate workers writing to the same DB → no corruption (WAL mode).

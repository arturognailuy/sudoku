---
domain: Designs
status: Active
entry_points:
  - cmd/root.go
  - db/db.go
dependencies:
  - .aidoc/designs/e2e-test-scenarios.md
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/database-puzzle-selection.md
  - .aidoc/designs/database-play-statistics.md
  - .aidoc/designs/database-concurrency.md
---

# E2E Database Scenarios

The database scenario catalog protects root-command database composition, played-state acquisition behavior, acquisition/completion statistics, Cobra discovery, and explicitly deferred database work.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/e2e-test-scenarios.md` | E2E discovery map, isolation rules, and automation entry points |
| `.aidoc/designs/database-puzzle-selection.md` | Current exact-grade acquisition and recycling contract |
| `.aidoc/designs/database-play-statistics.md` | Current completion, statistics, and history-reset contract |
| `.aidoc/designs/database-concurrency.md` | Proposed mixed-workload, lock-bound, and multi-process reliability contract |
| `AGENT.md` | Required black-box verification discipline |

## Why This Boundary

Database behavior crosses generation, classification, persistence, and startup. Deterministic cases belong in automation through the public `--from-db` boundary; generation fallback accounting remains covered at the narrowest deterministic package seam.

## 6. Database and Fallback

### 6.1 Auto-Store on Play
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** Puzzle is automatically stored in the DB at `~/.local/share/sudoku/puzzles.db`.

### 6.2 DB Fallback Path
1. Pre-populate the DB with easy puzzles: `./sudoku generate -n 20 -d easy --db $SUDOKU_DB`
2. Request an easy puzzle: `echo "quit" | ./sudoku --level easy --db $SUDOKU_DB`

**Expected:** If best-effort generation misses the target, the system falls back to the DB. If a match is found in the DB, no mismatch warning is shown. If the DB is also empty for that difficulty, the mismatch warning fires.

### 6.3 Mismatch Warning
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected (empty DB):** Best-effort likely misses easy target. Warning shown: "Requested difficulty: Easy. Generated puzzle difficulty: Medium/Hard. Enjoy!"

### 6.4 Multiple-Solution Puzzle Input
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** Warning about multiple solutions printed. Game still starts (plays with the first solution found).

---

## 7. Cobra Subcommand Structure

### 7.1 Root Command Shows Subcommands
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** Shows available commands: `generate`, `import`, `tui`, and usage for the default play mode.

---

## 12. Played-State Acquisition

These built-binary cases run in `scripts/e2e_cli.py` with an isolated database:

### 12.1 Never-Played Puzzles First
**Setup:** Import two distinct puzzles with the same exact strategy grade.
**Action:** Run `sudoku --from-db --level <grade> --db <path>` twice and quit each game.
**Expected:** Each stored puzzle is selected once before either repeats; both rows record one acquisition.

### 12.2 Balanced Reuse After Exhaustion
**Action:** Acquire a third puzzle from the two-puzzle fixture.
**Expected:** One least-played puzzle is returned and only its acquisition count increments. Repeated acquisitions keep counts within one of each other.

### 12.3 In-Place Migration
**Setup:** Create a pre-change database containing exact-grade puzzle rows, then open it with the new binary.
**Expected:** Migration preserves every puzzle and classification, initializes each row as unplayed, and the first acquisition succeeds.

### 12.4 Source and Failure Boundaries
**Action:** Exercise `--from-db` with an empty requested grade, a custom database path, and conflicting `--input`/`--resume` flags.
**Expected:** The command reports stable errors, does not generate a substitute, and does not mutate another database. Explicit input and resumed sessions leave acquisition history unchanged.

### 12.5 Generated Fallback Accounting
**Action:** Use the narrowest deterministic package seam to cover matched generation, generated mismatch with exact-grade DB fallback, and mismatch without a DB fallback.
**Expected:** Only the puzzle ultimately selected for play is marked played; a stored but unused generated mismatch remains unplayed.

## 13. Acquisition And Completion Statistics

These scenarios are executable through the package suites and built-binary harness described below:

### 13.1 Separate History Dimensions
**Setup:** Use a fixed normalized puzzle fixture. Acquire it twice, quit one run unfinished, and complete the other with player actions.
**Action:** Run `sudoku db stats --db <path>`.
**Expected:** The row reports two acquisitions and one completion. Acquisition is never labeled as completion or abandonment, and the overall row agrees with the per-grade snapshot.

### 13.2 Completion Boundaries
**Action:** Exercise quit, save/recovery, invalid moves, `solve`, a final player value, a hint-assisted final value, and undo/re-solve in isolated runs through the applicable built frontend.
**Expected:** Quit, persistence, invalid actions, and `solve` do not increment completion. A player or hint-assisted solve increments once per run; undo/re-solve does not increment twice. Loading an already solved session does not count.

### 13.3 Normalized Identity And Migration
**Setup:** Open a pre-completion-schema database, then submit digit-relabelled forms that normalize to the same existing puzzle.
**Expected:** Migration preserves the row and acquisition history, initializes completion history to zero, and every equivalent form contributes to the same normalized statistics row without creating a duplicate.

### 13.4 Statistics Filtering And Snapshot
**Action:** Request all-grade and single-grade statistics while a focused package test exercises a concurrent counter update.
**Expected:** Unknown grades fail before database work. Each successful command reports stored, selected, acquisition, completed-puzzle, completion, and latest-time fields from one read snapshot; empty timestamps render as `-`.

### 13.5 Explicit Reset Scope
**Action:** Preview `reset-history` for `acquisition`, `completion`, and `all`; cancel once; then confirm with `--yes`, both with and without `--level`.
**Expected:** The preview names the database, scope, filter, rows, and counters. Cancellation changes nothing. Confirmation resets exactly the requested counter/timestamp pairs atomically while preserving puzzle rows, classification, source, saved sessions, recovery records, and the non-selected history dimension.

### 13.6 Frontend And Failure Consistency
**Action:** Complete a puzzle through the line CLI, TUI, and HTTP API where each boundary applies; separately force completion persistence and reset failures.
**Expected:** Every frontend applies the same completion rule. A completion-write failure leaves the game solved and surfaces a concise warning; a reset failure exits non-zero with no partial reset.

## 14. Concurrent SQLite Reliability

These scenarios become mandatory with the implementation of `.aidoc/designs/database-concurrency.md`:

### 14.1 Multi-Process Import And Read
**Action:** Run several built `sudoku import` processes with overlapping fixed fixtures against one temporary database while bounded `sudoku db stats` readers execute.
**Expected:** Every process exits successfully, imports produce exactly the unique normalized rows, snapshots are internally consistent, and no process hangs. Deliberate lock exhaustion is tested separately at the package boundary.

### 14.2 Post-Contention Acquisition And Integrity
**Action:** After the writers close, acquire fixed-grade rows through `--from-db`, inspect exact counters, close all clients, and run SQLite `PRAGMA quick_check`.
**Expected:** Acquisition totals equal successful selections, balanced reuse still holds, committed counters survive reopen, and the integrity result is `ok`.

### 14.3 Deterministic Lock Bound
**Action:** In a focused package test, hold a write transaction and attempt a write from an independent handle before and after releasing the lock.
**Expected:** The blocked operation returns within the configured five-second bound without partial mutation; a later operation succeeds.

## 15. Other Deferred Database Scenarios

Keep these independently reviewed after acquisition/completion statistics:

- **Large import progress indicator:** Import 150+ puzzles → progress indicator fires every 100 puzzles.
- **Minimum-clues guard:** Import a puzzle with fewer than 17 clues → rejected or warned (prevents solver hang on near-empty boards).

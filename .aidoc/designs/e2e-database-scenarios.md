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
---

# E2E Database Scenarios

The database scenario catalog protects root-command database composition, probabilistic fallback behavior, Cobra discovery, and explicitly deferred database work.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/e2e-test-scenarios.md` | E2E discovery map, isolation rules, and automation entry points |
| `AGENT.md` | Required black-box verification discipline |

## Why This Boundary

Database behavior crosses generation, classification, persistence, and startup. Deterministic cases belong in automation; the public fallback branch remains manual until a stable forcing seam exists.

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

## 12. Planned Played-State Acquisition

These cases define built-binary acceptance for the next database implementation. Deterministic public cases belong in `scripts/e2e_cli.py` with an isolated database:

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

## 13. Other Deferred Database Scenarios

Keep these independently reviewed after played-state selection:

- **Large import progress indicator:** Import 150+ puzzles → progress indicator fires every 100 puzzles.
- **Minimum-clues guard:** Import a puzzle with fewer than 17 clues → rejected or warned (prevents solver hang on near-empty boards).
- **Concurrent DB access:** Multiple generate workers writing to the same DB → no corruption (WAL mode).

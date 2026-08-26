---
domain: Designs
status: Active
entry_points:
  - cmd/root.go
  - db/db.go
dependencies:
  - .aidoc/designs/e2e-test-scenarios.md
  - .aidoc/designs/game-engine.md
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

## 12. Deferred Database Scenarios

The deferred database scenarios begin only after stabilization and difficulty calibration, as sequenced in `.aidoc/designs/roadmap.md`:

- **Large import progress indicator:** Import 150+ puzzles → progress indicator fires every 100 puzzles.
- **Minimum-clues guard:** Import a puzzle with fewer than 17 clues → rejected or warned (prevents solver hang on near-empty boards).
- **Played tracking:** Mark puzzles as played → DB query skips played puzzles.
- **Concurrent DB access:** Multiple generate workers writing to the same DB → no corruption (WAL mode).

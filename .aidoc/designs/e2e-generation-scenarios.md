---
domain: Designs
status: Active
entry_points:
  - cmd/generate.go
  - generator/generator.go
dependencies:
  - .aidoc/designs/e2e-test-scenarios.md
  - .aidoc/designs/game-engine.md
---

# E2E Generation Scenarios

The generation scenario catalog verifies generation flags, worker composition, target reporting, database storage, validation, and deduplication through the built command.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/e2e-test-scenarios.md` | E2E discovery map, isolation rules, and automation entry points |
| `AGENT.md` | Required black-box verification discipline |

## Why This Boundary

Generation combines probabilistic puzzle construction with deterministic command and storage contracts. Assertions therefore protect composition and bounded outcomes rather than requiring a random run to hit an exact grade.

## 4. Puzzle Generation CLI

### 4.1 Generate Help
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** Shows all flags: `-n`, `-d`, `-w`, `-t`, `--rounds`, `--db`.

### 4.2 Generate Easy Puzzles
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** Generates 5 puzzles. Reports: generated count, stored count, duplicates, by-level breakdown. Some may be classified at a different difficulty than requested (mismatch is expected for easy/medium targets).

### 4.3 Generate Hard Puzzles
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** Generates 3 puzzles. Each is classified and stored. Duration reported.

### 4.4 Generate with Parallel Workers
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** Uses 4 parallel goroutines. Completes generation. All puzzles stored.

### 4.5 Generate with Invalid Difficulty
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** Error: "Invalid difficulty level". Exit code 1.

### 4.6 Generate with Count 0
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** Error: "count must be positive". Exit code 1.

### 4.7 Generate with Custom DB Path
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** DB file created at `/tmp/custom-test.db`. Puzzles stored there.

### 4.8 Generate Dedup
**Action:** Execute the matching case in `scripts/e2e_cli.py`, which owns the canonical command sequence and fixture.
**Expected:** Second run may report duplicates if the same normalized puzzles are generated.

---

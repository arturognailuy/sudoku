---
domain: Designs
status: Active
entry_points:
  - cmd/root.go
  - cmd/session.go
  - cli/controller.go
  - tui/model.go
  - webapi/server.go
  - db/db.go
  - db/puzzle.go
dependencies:
  - .aidoc/designs/database-puzzle-selection.md
  - .aidoc/designs/game-engine.md
  - .aidoc/designs/e2e-database-scenarios.md
  - .aidoc/designs/database-concurrency.md
---

# Database Play Statistics and History Reset

The database keeps completion counters beside acquisition counters, exposes both as separate concepts through `sudoku db stats`, and provides an explicitly confirmed `sudoku db reset-history` command. A completion is counted once per play run when a player action first solves the puzzle; the automatic `solve` action does not count. This behavior does not infer abandonment or elapsed play duration, and it preserves the normalized puzzle key and `INSERT OR IGNORE` deduplication contract.

## Related Docs

| Document | Relationship |
|----------|-------------|
| `.aidoc/designs/database-puzzle-selection.md` | Defines acquisition counters and the selection policy that consumes them |
| `.aidoc/designs/game-engine.md` | Defines solved status and typed actions used to detect completion |
| `.aidoc/designs/e2e-database-scenarios.md` | Owns black-box acceptance scenarios for statistics and reset behavior |
| `.aidoc/designs/database-concurrency.md` | Defines mixed-workload snapshot/reset contention and lock bounds |
| `.aidoc/designs/roadmap.md` | Sequences this increment before broader database reliability work |

## Why Completion Is Separate From Acquisition

`play_count` means that a stored puzzle was selected and exposed to a player. It increments before play begins so the selector can prefer unseen puzzles and balance reuse even if the process later quits, crashes, or saves for another day. It cannot answer whether a puzzle was solved.

Completion is the smallest reliable gameplay outcome. The engine reports exactly when a valid action reaches `game.StatusSolved`, while abandonment cannot be inferred from process exit: an exit may be a deliberate stop, a crash, or a session that will be resumed. The database therefore keeps acquisition and completion as separate dimensions and does not derive one from the other.

This increment interprets “completion times” as the number of successful completions plus the latest completion timestamp. It deliberately does not measure elapsed solving duration. Duration would require reviewed start, pause, idle, resume, and clock-change semantics and is not needed to count completions accurately.

## What Counts As A Completion

A **play run** begins when a frontend creates or restores a playable `game.Game` and ends when that frontend exits or discards it.

- Count at most one completion in each play run.
- Count the first successful action whose result changes the run from not solved to `game.StatusSolved`.
- Do not count `game.ActionSolve`; asking the complete solver to finish is not a player completion.
- A completion reached with hints still counts. Hints assist play but remain explicit player actions and do not automatically solve the whole puzzle.
- Undoing and re-solving within the same run does not add another completion.
- Restoring a saved or recovered session creates a new run. If the restored game is unfinished and the player completes it, count one completion for that run.
- Starting from an already solved serialized session does not count merely because it was loaded.
- Quitting, process termination, saving, recovery creation, and invalid moves never create completion or abandonment records.

Counting is frontend-neutral. The line CLI, TUI, and HTTP API all observe accepted `game.Result` values through one shared play-run tracker rather than reimplementing completion rules independently. `game.Game` remains storage-neutral: it reports actions and solved status but does not open SQLite or suppress a successful game action when statistics persistence fails.

## What The Database Stores

Apply an additive migration in `db.DB.migrate` that adds `completion_count` with a non-null zero default and nullable `last_completed_at` to `puzzles`. Existing rows migrate to zero completions and no completion timestamp. `db.DB.RecordCompletion` atomically increments the counter and assigns SQLite's current timestamp for the existing normalized puzzle key.

The puzzle key is the existing normalized 81-character puzzle string. `cmd/play.go` already remaps digits from the solved first row and stores that normalized string as the primary key. Imports, generated puzzles, and direct input continue to use `INSERT OR IGNORE`; duplicate input never creates another puzzle row or clears either history.

The run tracker retains that normalized key and active database path from `cmd.createSession`. On restore, it derives the key from the immutable givens and uses the current `--db` path (or the XDG default). The root and TUI commands must both accept the same `--db` behavior. If an old session’s normalized puzzle is absent from the selected database, completion persistence reports a warning and leaves gameplay/session persistence intact rather than silently inserting or mutating another database.

The migration does not add a starts table, attempt/event log, abandonment field, elapsed-duration field, schema-version table, or broader Sudoku-symmetry normalization. It preserves the existing digit-relabeling normalization exactly; rotations, reflections, transposition, and row/column symmetry remain outside the deduplication contract.

## What Users Can Inspect

Add a `db` Cobra command group with:

```text
sudoku db stats [--db <path>] [--level <easy|medium|hard|expert|evil>]
```

The command prints one row per included strategy grade and one overall row. Each row reports:

- stored puzzles;
- puzzles never selected and selected at least once;
- total selections;
- puzzles completed at least once;
- total completions;
- latest selection time;
- latest completion time.

Labels use **selected/acquisitions** for `play_count` and **completed/completions** for `completion_count`; neither is presented as abandonment, elapsed duration, or unique players. Empty timestamps render as `-`. `--level` rejects unknown grades before opening or mutating the database. `--db` follows the existing explicit-path/XDG-default behavior.

This command is a read-only snapshot. It may observe a state immediately before or after a concurrent acquisition/completion, but each returned aggregate comes from one SQLite read transaction so rows and the overall total describe the same snapshot.

## How History Reset Works

Add:

```text
sudoku db reset-history --history <acquisition|completion|all>
                        [--level <easy|medium|hard|expert|evil>]
                        [--db <path>] [--yes]
```

`--history` is required so the destructive scope is explicit:

- `acquisition` sets `play_count = 0` and `last_played_at = NULL`;
- `completion` sets `completion_count = 0` and `last_completed_at = NULL`;
- `all` resets both dimensions in one transaction.

Before mutation, print the selected database, optional grade filter, affected puzzle count, and counters that will be cleared. In an interactive terminal, require an exact affirmative confirmation; any other response cancels without mutation. In non-interactive use, require `--yes` after printing the preview. An empty selection succeeds without writing.

Reset never deletes puzzle rows, changes difficulty/classification/source, alters normalized keys, or touches explicit save files and recovery records. A concurrent acquisition or completion either happens before or after the reset transaction; no partial counter/timestamp pair is visible. After an acquisition reset, selection again prefers every zero-count puzzle. After a completion reset, a still-running play may record a later completion normally.

## How Completion Persistence Integrates

Introduce a small frontend-neutral tracker around `game.Game.Apply`:

1. `cmd.createSession` resolves the database path and normalized puzzle key and constructs one tracker per play run.
2. The tracker delegates the typed action to `game.Game.Apply`.
3. After a successful non-`ActionSolve` result first reaches `StatusSolved`, it atomically records completion and marks that run as recorded.
4. CLI, TUI, and API action paths call the tracker; rendering and protocol responses continue to consume the same `game.Result`.
5. A database failure does not roll back the already accepted game action. The frontend surfaces one concise warning and the tracker may retry only when another accepted action again reaches solved without having recorded the run.

This keeps SQLite concerns out of `game`, centralizes once-per-run behavior, and gives package tests a deterministic recorder seam. `cli/controller.go`, `tui/model.go`, and `webapi/server.go` remain presentation/protocol adapters rather than independent statistics implementations.

## Failure, Compatibility, And Privacy Boundaries

- Existing databases and session files remain readable; new completion columns default safely.
- Existing `play_count`/`last_played_at` selection semantics do not change.
- Statistics are local to the selected SQLite file. No telemetry, account identifier, or network reporting is added.
- Completion updates use the existing bounded SQLite busy timeout and return promptly under lock contention.
- A statistics write failure warns but never changes whether the puzzle is solved.
- A reset failure rolls back the complete requested scope and exits non-zero.
- Exact once-per-run is an in-memory guarantee. Replaying the same saved unfinished session in a later process is another run and may produce another completion; no durable attempt identity is introduced in this increment.

## Verification Plan

Package tests cover additive migration, atomic completion increments, filtered aggregate snapshots, reset scopes, rollback/error behavior, missing normalized rows, `ActionSolve` exclusion, hints, undo/re-solve suppression, and concurrent increments.

Built-binary scenarios in `.aidoc/designs/e2e-database-scenarios.md` cover:

- separate acquisition and completion output;
- unfinished and solver-completed runs not incrementing completion;
- one player completion incrementing once;
- normalized duplicates sharing one statistics row;
- grade-filtered preview/reset and explicit confirmation;
- preservation of puzzle rows, classifications, save files, and the non-reset history dimension;
- line CLI, TUI, and API completion consistency where each public boundary applies.

The implementation PR must update the executable scenario harnesses and run every affected unit, race, vet, lint, API-contract, and built-binary E2E lane before review.

## Deferred Decisions

- Abandonment and give-up definitions.
- Elapsed solving duration, pause/idle policy, and timing across restore.
- Durable play-attempt identities or append-only event history.
- Player/account attribution or telemetry.
- Large-import changes, minimum-clue policy, and full Sudoku-symmetry canonicalization. Mixed-workload SQLite stress is designed separately in `.aidoc/designs/database-concurrency.md`.

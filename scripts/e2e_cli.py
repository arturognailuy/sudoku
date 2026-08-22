#!/usr/bin/env python3
"""Deterministic black-box E2E checks for the built line CLI and commands."""

import argparse
import json
import os
from pathlib import Path
import sqlite3
import stat
import subprocess
import tempfile

PUZZLE_DOTS = "..3.2.6..9..3.5..1..18.64....81.29..7.......8..67.82....26.95..8..2.3..9..5.1.3.."
PUZZLE_ZEROS = PUZZLE_DOTS.replace(".", "0")
MULTIPLE_SOLUTIONS = "....7....6..195....98....6.8...6...34..8.3..17...2...6.6....28....419..5....8..79"


def run(binary, args, root, input_text=None, expected=0, timeout=20):
    env = os.environ.copy()
    env["XDG_DATA_HOME"] = str(root / "data")
    env["XDG_STATE_HOME"] = str(root / "state")
    result = subprocess.run(
        [binary, *args],
        input=input_text,
        capture_output=True,
        text=True,
        env=env,
        timeout=timeout,
    )
    output = result.stdout + result.stderr
    if result.returncode != expected:
        raise AssertionError(
            f"{' '.join(args)} exited {result.returncode}, want {expected}\n{output[-4000:]}"
        )
    return output


def contains(output, *needles):
    missing = [needle for needle in needles if needle not in output]
    if missing:
        raise AssertionError(f"output missing {missing}\n{output[-4000:]}")


def excludes(output, *needles):
    present = [needle for needle in needles if needle in output]
    if present:
        raise AssertionError(f"output unexpectedly contains {present}\n{output[-4000:]}")


def puzzle_rows(database):
    with sqlite3.connect(database) as connection:
        return connection.execute(
            "SELECT puzzle, difficulty, source FROM puzzles ORDER BY puzzle"
        ).fetchall()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("binary", nargs="?", default="./sudoku")
    args = parser.parse_args()
    binary = os.path.abspath(args.binary)

    with tempfile.TemporaryDirectory(prefix="sudoku-cli-e2e-") as directory:
        root = Path(directory)

        # Startup, parsing, help, and backward-compatible root flags.
        output = run(binary, ["--help"], root)
        contains(output, "generate", "import", "tui", "--input", "--level")
        output = run(binary, ["generate", "--help"], root)
        contains(output, "--count", "--difficulty", "--workers", "--timeout", "--rounds", "--db")
        output = run(binary, ["import", "--help"], root)
        contains(output, "--file", "--source", "--db")
        contains(run(binary, ["--input", PUZZLE_DOTS], root, "q\n"), "Exiting the game.", PUZZLE_DOTS)
        contains(run(binary, ["--input", PUZZLE_ZEROS], root, "q\n"), "Exiting the game.", PUZZLE_DOTS)
        contains(run(binary, ["--input", "123"], root, expected=1), "not a valid Sudoku problem")
        contains(run(binary, ["--level", "banana"], root, expected=1), "invalid difficulty level")
        contains(
            run(binary, ["--level", "easy"], root, "q\n", timeout=60),
            "Requested difficulty: Easy.",
            "Exiting the game.",
        )
        contains(
            run(binary, ["--input", MULTIPLE_SOLUTIONS], root, "q\n"),
            "has 2 solutions",
            "Exiting the game.",
        )

        # Real line-controller lifecycle: invalid state, repair, values, clear,
        # history, hint metadata, reset, notes, and clean quit.
        commands = "\n".join(
            [
                "1 1 5",
                "check",
                "repair",
                "add 1 1 4",
                "clear 1 1",
                "undo",
                "redo",
                "note 1 1 1",
                "notes-clear 1 1",
                "hint",
                "reset",
                "q",
                "",
            ]
        )
        output = run(binary, ["--input", PUZZLE_DOTS], root, commands)
        contains(
            output,
            "You have entered incorrect value(s).",
            "Hint:",
            "Exiting the game.",
        )

        # A new action after undo must truncate the abandoned redo branch.
        output = run(
            binary,
            ["--input", PUZZLE_DOTS],
            root,
            "1 1 4\nu\n1 1 5\nr\nc\nq\n",
        )
        contains(output, "You have entered incorrect value(s).", "Exiting the game.")
        excludes(output, "The current board is correct.")
        contains(
            run(binary, ["--input", PUZZLE_DOTS], root, "solve\n"),
            "Congratulations! You have solved the problem.",
        )

        # Durable sessions preserve notes, invalid values, and redo history.
        session = root / "session.json"
        output = run(
            binary,
            ["--input", PUZZLE_DOTS],
            root,
            f"note 1 2 1\n1 1 5\n1 1 4\nu\nsave {session}\nq\n",
        )
        contains(output, f"Session saved to {session}.")
        if stat.S_IMODE(session.stat().st_mode) != 0o600:
            raise AssertionError("session file mode is not 0600")
        saved = json.loads(session.read_text(encoding="utf-8"))
        if saved.get("version") != 1:
            raise AssertionError("saved session does not use version 1")
        output = run(binary, ["--resume", str(session)], root, "check\nr\ncheck\nq\n")
        contains(
            output,
            "You have entered incorrect value(s).",
            "The current board is correct.",
        )

        corrupt = root / "corrupt.json"
        corrupt.write_text("{bad json", encoding="utf-8")
        unsupported = root / "unsupported.json"
        unsupported.write_text('{"version":999}', encoding="utf-8")
        oversized = root / "oversized.json"
        oversized.write_bytes(b"x" * (1024 * 1024 + 1))
        for source, message in (
            (corrupt, "resume saved session"),
            (unsupported, "resume saved session"),
            (oversized, "session file is too large"),
        ):
            contains(run(binary, ["--resume", str(source)], root, expected=1), message)
        contains(
            run(binary, ["--resume", str(session), "--input", PUZZLE_DOTS], root, expected=1),
            "none of the others can be",
        )
        contains(
            run(binary, ["--resume", str(session), "--level", "easy"], root, expected=1),
            "none of the others can be",
        )
        destination = root / "existing-destination"
        destination.mkdir()
        output = run(binary, ["--input", PUZZLE_DOTS], root, f"save {destination}\nq\n")
        contains(output, "Failed to run the save command")
        if not destination.is_dir() or list(root.glob(".sudoku-session-*")):
            raise AssertionError("failed save changed the destination or leaked a temporary file")

        # Import normalization, invalid lines, source labels, empty input, and dedup.
        database = root / "commands.db"
        puzzles = root / "puzzles.txt"
        puzzles.write_text(
            "# fixture\n" + PUZZLE_DOTS + "\n" + PUZZLE_ZEROS + "\n123456\nabc\n",
            encoding="utf-8",
        )
        output = run(
            binary,
            ["import", "--file", str(puzzles), "--source", "e2e-fixture", "--db", str(database)],
            root,
        )
        contains(output, "Total lines: 4", "Valid: 2", "Invalid (skipped): 2", "Stored (new): 1", "Duplicates: 1")
        rows = puzzle_rows(database)
        if len(rows) != 1 or rows[0][2] != "e2e-fixture" or len(rows[0][0]) != 81 or "0" in rows[0][0]:
            raise AssertionError(f"unexpected imported database rows: {rows}")
        output = run(
            binary,
            ["import", "--file", str(puzzles), "--source", "second", "--db", str(database)],
            root,
        )
        contains(output, "Stored (new): 0", "Duplicates: 2")
        empty = root / "empty.txt"
        empty.write_text("# comments only\n\n", encoding="utf-8")
        contains(
            run(binary, ["import", "--file", str(empty), "--db", str(database)], root),
            "Total lines: 0",
            "Stored (new): 0",
        )
        contains(
            run(binary, ["import", "--file", str(root / "missing.txt")], root, expected=1),
            "open file",
        )

        # Generation validation and a tightly bounded real-worker smoke run.
        contains(run(binary, ["generate", "--count", "0"], root, expected=1), "count must be positive")
        contains(run(binary, ["generate", "--difficulty", "invalid"], root, expected=1), "invalid difficulty level")
        generated = root / "generated.db"
        output = run(
            binary,
            [
                "generate",
                "--count",
                "1",
                "--difficulty",
                "hard",
                "--workers",
                "2",
                "--timeout",
                "1ms",
                "--rounds",
                "1",
                "--db",
                str(generated),
            ],
            root,
            timeout=60,
        )
        contains(output, "Generated: 1", "=== Generation Report ===")
        if not generated.is_file() or len(puzzle_rows(generated)) != 1:
            raise AssertionError("bounded generation did not create one SQLite puzzle")

        # Root play auto-stores through the default XDG data path.
        auto_database = root / "data" / "sudoku" / "puzzles.db"
        contains(run(binary, ["--input", PUZZLE_DOTS], root, "q\n"), "Exiting the game.")
        if not auto_database.is_file() or not puzzle_rows(auto_database):
            raise AssertionError("root play did not auto-store the puzzle")

    print("PASS: line CLI gameplay, sessions, import, generation, and SQLite composition")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Enforce a ratcheted readability budget for maintained Go sources."""

from __future__ import annotations

import argparse
import json
from dataclasses import asdict, dataclass, field
from pathlib import Path

MAX_FILE_LINES = 1_000
MAX_LINE_LENGTH = 140
BASELINE_VERSION = 1
DEFAULT_BASELINE = Path("scripts/review_size_baseline.json")
EXEMPT_PATHS = {
    Path("webapi/generated.go"),
    Path("message/english.go"),
}


@dataclass(frozen=True)
class Finding:
    path: str
    file_lines: int | None = None
    long_line_count: int = 0
    widest_line: int = 0
    widest_line_number: int = field(default=1, compare=False)


def findings(root: Path) -> list[Finding]:
    result: list[Finding] = []
    for path in sorted(root.rglob("*.go")):
        relative = path.relative_to(root)
        if relative in EXEMPT_PATHS or "testdata" in relative.parts:
            continue
        lines = path.read_text(encoding="utf-8").splitlines()
        long_lines = [
            (number, len(line))
            for number, line in enumerate(lines, 1)
            if len(line) > MAX_LINE_LENGTH
        ]
        if len(lines) <= MAX_FILE_LINES and not long_lines:
            continue
        widest_line_number, widest_line = max(long_lines, key=lambda item: item[1], default=(1, 0))
        result.append(
            Finding(
                path=relative.as_posix(),
                file_lines=len(lines) if len(lines) > MAX_FILE_LINES else None,
                long_line_count=len(long_lines),
                widest_line=widest_line,
                widest_line_number=widest_line_number,
            )
        )
    return result


def baseline_payload(current: list[Finding]) -> dict[str, object]:
    return {
        "version": BASELINE_VERSION,
        "limits": {"file_lines": MAX_FILE_LINES, "line_length": MAX_LINE_LENGTH},
        "findings": [asdict(item) for item in current],
    }


def write_baseline(path: Path, current: list[Finding]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(baseline_payload(current), indent=2) + "\n", encoding="utf-8")


def read_baseline(path: Path) -> list[Finding]:
    payload = json.loads(path.read_text(encoding="utf-8"))
    expected_limits = {"file_lines": MAX_FILE_LINES, "line_length": MAX_LINE_LENGTH}
    if payload.get("version") != BASELINE_VERSION or payload.get("limits") != expected_limits:
        raise ValueError("baseline version or limits do not match the checker")
    return [Finding(**item) for item in payload.get("findings", [])]


def differences(expected: list[Finding], current: list[Finding]) -> list[tuple[str, Finding | None, Finding | None]]:
    expected_by_path = {item.path: item for item in expected}
    current_by_path = {item.path: item for item in current}
    return [
        (path, expected_by_path.get(path), current_by_path.get(path))
        for path in sorted(expected_by_path.keys() | current_by_path.keys())
        if expected_by_path.get(path) != current_by_path.get(path)
    ]


def describe(item: Finding | None) -> str:
    if item is None:
        return "no signal"
    parts: list[str] = []
    if item.file_lines is not None:
        parts.append(f"{item.file_lines} file lines")
    if item.long_line_count:
        parts.append(f"{item.long_line_count} long lines, widest {item.widest_line}")
    return "; ".join(parts) or "no signal"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("root", nargs="?", type=Path, default=Path("."))
    parser.add_argument("--baseline", type=Path, default=DEFAULT_BASELINE)
    parser.add_argument("--write-baseline", action="store_true")
    args = parser.parse_args()

    current = findings(args.root)
    if args.write_baseline:
        write_baseline(args.baseline, current)
        print(f"Wrote {len(current)} readability baseline signal(s) to {args.baseline}.")
        return 0

    try:
        expected = read_baseline(args.baseline)
    except (OSError, ValueError, json.JSONDecodeError, TypeError) as error:
        print(f"::error file={args.baseline}::Cannot load readability baseline: {error}")
        return 1

    drift = differences(expected, current)
    for path, old, new in drift:
        line = new.widest_line_number if new is not None else 1
        print(
            f"::error file={path},line={line}::Readability budget changed: "
            f"baseline [{describe(old)}], current [{describe(new)}]. "
            "Fix regressions; after an improvement, run "
            f"python3 scripts/review_size.py --write-baseline to ratchet the budget."
        )
    if drift:
        print(f"Readability review: {len(drift)} baseline difference(s).")
        return 1

    print(f"Readability review: {len(current)} accepted baseline signal(s); no regressions.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

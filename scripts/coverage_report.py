#!/usr/bin/env python3
"""Render a Go cover profile as a package-level Markdown report."""

from __future__ import annotations

import argparse
from collections import defaultdict
from pathlib import Path


def read_profile(path: Path) -> dict[str, tuple[int, int]]:
    blocks: dict[tuple[str, str], tuple[int, int]] = {}
    with path.open(encoding="utf-8") as profile:
        header = profile.readline().strip()
        if not header.startswith("mode: "):
            raise ValueError("missing Go coverage mode header")
        for line_number, line in enumerate(profile, 2):
            fields = line.split()
            if len(fields) != 3:
                raise ValueError(f"invalid coverage record on line {line_number}")
            location, statements_text, count_text = fields
            source, block = location.rsplit(":", 1)
            statements = int(statements_text)
            count = int(count_text)
            key = (source, block)
            previous = blocks.get(key)
            if previous is None or count > previous[1]:
                blocks[key] = (statements, count)

    packages: dict[str, list[int]] = defaultdict(lambda: [0, 0])
    for (source, _), (statements, count) in blocks.items():
        package = source.rsplit("/", 1)[0]
        packages[package][0] += statements
        if count > 0:
            packages[package][1] += statements
    return {package: (values[0], values[1]) for package, values in packages.items()}


def render(packages: dict[str, tuple[int, int]]) -> str:
    lines = [
        "## Go package coverage",
        "",
        "Coverage is review evidence, not a pass/fail threshold. Prioritize meaningful boundary and failure-path tests over percentage chasing.",
        "",
        "| Package | Statements | Covered | Coverage |",
        "|---|---:|---:|---:|",
    ]
    total_statements = 0
    total_covered = 0
    for package, (statements, covered) in sorted(packages.items()):
        percent = 100 * covered / statements if statements else 0
        lines.append(f"| `{package}` | {statements} | {covered} | {percent:.1f}% |")
        total_statements += statements
        total_covered += covered
    total_percent = 100 * total_covered / total_statements if total_statements else 0
    lines.append(f"| **Total** | **{total_statements}** | **{total_covered}** | **{total_percent:.1f}%** |")
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("profile", type=Path)
    args = parser.parse_args()
    try:
        print(render(read_profile(args.profile)))
    except (OSError, ValueError) as error:
        parser.error(str(error))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

#!/usr/bin/env python3

import tempfile
import unittest
from pathlib import Path

from scripts import coverage_report


class CoverageReportTest(unittest.TestCase):
    def test_combines_repeated_blocks_and_packages(self) -> None:
        content = """mode: set
example.com/project/a/a.go:1.1,2.1 2 0
example.com/project/a/a.go:1.1,2.1 2 1
example.com/project/a/a.go:3.1,4.1 1 0
example.com/project/b/b.go:1.1,2.1 3 1
"""
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory, "coverage.out")
            path.write_text(content, encoding="utf-8")
            report = coverage_report.render(coverage_report.read_profile(path))

        self.assertIn("| `example.com/project/a` | 3 | 2 | 66.7% |", report)
        self.assertIn("| `example.com/project/b` | 3 | 3 | 100.0% |", report)
        self.assertIn("| **Total** | **6** | **5** | **83.3%** |", report)

    def test_rejects_missing_header(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory, "coverage.out")
            path.write_text("bad\n", encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "missing Go coverage mode header"):
                coverage_report.read_profile(path)


if __name__ == "__main__":
    unittest.main()

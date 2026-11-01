from pathlib import Path
from tempfile import TemporaryDirectory
import unittest

from scripts import review_size


class ReviewSizeTest(unittest.TestCase):
    def test_baseline_matches_exact_findings(self) -> None:
        with TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "debt.go").write_text("package debt\n" + ("x" * 141) + "\n", encoding="utf-8")
            current = review_size.findings(root)
            baseline = root / "baseline.json"
            review_size.write_baseline(baseline, current)

            self.assertEqual(review_size.read_baseline(baseline), current)
            self.assertEqual(review_size.differences(current, current), [])

    def test_regression_and_improvement_both_require_review(self) -> None:
        original = [
            review_size.Finding(
                path="debt.go",
                long_line_count=1,
                widest_line=141,
                widest_line_number=2,
            )
        ]
        regression = [
            review_size.Finding(
                path="debt.go",
                long_line_count=2,
                widest_line=150,
                widest_line_number=3,
            )
        ]

        moved_only = [
            review_size.Finding(
                path="debt.go",
                long_line_count=1,
                widest_line=141,
                widest_line_number=20,
            )
        ]

        self.assertEqual(len(review_size.differences(original, regression)), 1)
        self.assertEqual(len(review_size.differences(original, [])), 1)
        self.assertEqual(review_size.differences(original, moved_only), [])

    def test_exempt_generated_sources_and_testdata(self) -> None:
        with TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "webapi").mkdir()
            (root / "webapi" / "generated.go").write_text("x" * 200, encoding="utf-8")
            (root / "pkg" / "testdata").mkdir(parents=True)
            (root / "pkg" / "testdata" / "fixture.go").write_text("x" * 200, encoding="utf-8")

            self.assertEqual(review_size.findings(root), [])


if __name__ == "__main__":
    unittest.main()

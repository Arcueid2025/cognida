import importlib.util
from pathlib import Path
from types import SimpleNamespace
import unittest


MODULE_PATH = Path(__file__).with_name("run_retrieval_baseline.py")
SPEC = importlib.util.spec_from_file_location("payment_baseline", MODULE_PATH)
baseline = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(baseline)


class RetrievalBaselineTests(unittest.TestCase):
    def test_evidence_ids_are_deduplicated_and_sorted(self):
        items = [
            {"content": "【PI-KB-002-4】x 【PI-KB-001-3】y"},
            {"content": "【PI-KB-002-4】duplicate"},
        ]
        self.assertEqual(baseline.evidence_ids(items), ["PI-KB-001-3", "PI-KB-002-4"])

    def test_percentile95_uses_nearest_rank(self):
        self.assertEqual(baseline.percentile95([1, 2, 3, 4, 5]), 5)
        self.assertEqual(baseline.percentile95([]), None)

    def test_mode_metrics_exclude_evidence_free_cases(self):
        original = baseline.call_search
        baseline.call_search = lambda args, prompt, mode: (
            [{"content": "【PI-KB-001-1】", "chunk_id": "chunk-1"}],
            12.5,
            None,
        )
        try:
            result = baseline.evaluate_mode(
                SimpleNamespace(),
                [
                    {"id": "with-evidence", "prompt": "p", "required_evidence": ["PI-KB-001-1"]},
                    {"id": "evidence-free", "prompt": "p", "required_evidence": []},
                ],
                "hybrid",
            )
        finally:
            baseline.call_search = original

        summary = result["summary"]
        self.assertEqual(summary["mean_evidence_recall_at_k"], 1.0)
        self.assertEqual(summary["evidence_free_cases"], 1)
        self.assertEqual(summary["failure_case_ids"], [])


if __name__ == "__main__":
    unittest.main()

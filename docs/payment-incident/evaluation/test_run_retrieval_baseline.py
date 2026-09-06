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
        baseline.call_search = lambda args, prompt, mode, enable_rerank=False: (
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

    def test_mode_metrics_passes_rerank_flag_to_search(self):
        received = []

        def fake_search(args, prompt, mode, enable_rerank=False):
            received.append(enable_rerank)
            return [{"content": "【PI-KB-001-1】"}], 1.0, None

        original = baseline.call_search
        baseline.call_search = fake_search
        try:
            baseline.evaluate_mode(
                SimpleNamespace(),
                [{"id": "rerank", "prompt": "p", "required_evidence": ["PI-KB-001-1"]}],
                "hybrid",
                enable_rerank=True,
            )
        finally:
            baseline.call_search = original

        self.assertEqual(received, [True])

    def test_rerank_comparison_keeps_rerank_off_without_recall_gain(self):
        comparisons = baseline.compare_rerank_variants(
            {
                "hybrid": {"summary": {"mean_evidence_recall_at_k": 0.6, "p95_latency_ms": 10.0, "api_errors": 0}},
                "hybrid+rerank": {"summary": {"mean_evidence_recall_at_k": 0.6, "p95_latency_ms": 20.0, "api_errors": 1}},
            }
        )

        self.assertEqual(comparisons, [{
            "mode": "hybrid",
            "baseline": "hybrid",
            "candidate": "hybrid+rerank",
            "recall_delta": 0.0,
            "p95_latency_delta_ms": 10.0,
            "api_error_delta": 1,
            "recommendation": "keep_rerank_off",
        }])


if __name__ == "__main__":
    unittest.main()

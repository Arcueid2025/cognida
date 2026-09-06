#!/usr/bin/env python3
"""Run a reproducible, read-only retrieval baseline for IncidentPilot.

The evaluator deliberately calls Cognida's public retrieval API rather than
Milvus directly. That keeps the measurement aligned with the retrieval path
used by both the UI and the Agent's rag_query tool.
"""

from __future__ import annotations

import argparse
import json
import math
import os
import re
import sys
import time
import urllib.error
import urllib.request
from datetime import datetime, timezone
from pathlib import Path
from statistics import mean
from typing import Any


EVIDENCE_ID = re.compile(r"【(PI-KB-\d+-\d+)】")
DEFAULT_MODES = ("vector", "bm25", "hybrid")
RERANK_VARIANTS = {"off": False, "on": True}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run IncidentPilot retrieval baseline")
    parser.add_argument("--kb-id", required=True, help="支付系统故障知识库 ID")
    parser.add_argument(
        "--cases",
        type=Path,
        default=Path(__file__).with_name("payment_incident_dev_cases.json"),
        help="开发评测集路径",
    )
    parser.add_argument("--base-url", default="http://localhost:8080", help="Cognida Go 服务地址")
    parser.add_argument("--token", default=os.getenv("COGNIDA_TOKEN", ""), help="Bearer token；也可设 COGNIDA_TOKEN")
    parser.add_argument("--top-k", type=int, default=5, help="每题返回片段数，默认 5")
    parser.add_argument("--min-score", type=float, default=0.0, help="透传至检索 API 的最低分")
    parser.add_argument("--modes", default=",".join(DEFAULT_MODES), help="逗号分隔：vector,bm25,hybrid")
    parser.add_argument(
        "--rerank-variants",
        default="off",
        help="逗号分隔：off,on；使用 off,on 对同一检索配置做重排 A/B 对比",
    )
    parser.add_argument("--timeout", type=float, default=60, help="单次 API 调用超时（秒）")
    parser.add_argument("--output", type=Path, required=True, help="结果 JSON 输出路径（建议 evaluation/results/）")
    return parser.parse_args()


def percentile95(values: list[float]) -> float | None:
    if not values:
        return None
    return sorted(values)[max(0, math.ceil(len(values) * 0.95) - 1)]


def load_cases(path: Path) -> tuple[str, list[dict[str, Any]]]:
    data = json.loads(path.read_text(encoding="utf-8"))
    cases = data.get("cases")
    if not isinstance(cases, list) or not cases:
        raise ValueError(f"评测集缺少非空 cases: {path}")
    return str(data.get("dataset_id", path.stem)), cases


def call_search(
    args: argparse.Namespace, prompt: str, mode: str, enable_rerank: bool = False
) -> tuple[list[dict[str, Any]], float, str | None]:
    body = json.dumps(
        {
            "kb_ids": [args.kb_id],
            "query": prompt,
            "top_k": args.top_k,
            "min_score": args.min_score,
            "retrieval_mode": mode,
            "enable_rerank": enable_rerank,
        }
    ).encode("utf-8")
    request = urllib.request.Request(
        args.base_url.rstrip("/") + "/api/v1/knowledge/search",
        data=body,
        headers={"Content-Type": "application/json", **({"Authorization": f"Bearer {args.token}"} if args.token else {})},
        method="POST",
    )
    started = time.perf_counter()
    try:
        with urllib.request.urlopen(request, timeout=args.timeout) as response:
            payload = json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as error:
        detail = error.read().decode("utf-8", errors="replace")
        return [], (time.perf_counter() - started) * 1000, f"HTTP {error.code}: {detail}"
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as error:
        return [], (time.perf_counter() - started) * 1000, str(error)

    elapsed_ms = (time.perf_counter() - started) * 1000
    if payload.get("code") != 0:
        return [], elapsed_ms, f"API code {payload.get('code')}: {payload.get('message', '')}"
    items = payload.get("data", {}).get("items", [])
    if not isinstance(items, list):
        return [], elapsed_ms, "API response data.items is not a list"
    return items, elapsed_ms, None


def evidence_ids(items: list[dict[str, Any]]) -> list[str]:
    found: set[str] = set()
    for item in items:
        found.update(EVIDENCE_ID.findall(str(item.get("content", ""))))
    return sorted(found)


def evaluate_mode(
    args: argparse.Namespace, cases: list[dict[str, Any]], mode: str, enable_rerank: bool = False
) -> dict[str, Any]:
    results: list[dict[str, Any]] = []
    latencies: list[float] = []
    recalls: list[float] = []
    errors = 0
    evidence_free = 0

    for case in cases:
        required = sorted(set(case.get("required_evidence", [])))
        items, latency_ms, error = call_search(args, str(case["prompt"]), mode, enable_rerank)
        latencies.append(round(latency_ms, 3))
        retrieved = evidence_ids(items)
        matched = sorted(set(required) & set(retrieved))
        recall = len(matched) / len(required) if required else None
        if recall is None:
            evidence_free += 1
        else:
            recalls.append(recall)
        if error:
            errors += 1

        results.append(
            {
                "case_id": case["id"],
                "category": case.get("category"),
                "latency_ms": round(latency_ms, 3),
                "error": error,
                "required_evidence": required,
                "retrieved_evidence": retrieved,
                "matched_evidence": matched,
                "evidence_recall_at_k": recall,
                "results": [
                    {
                        "rank": rank,
                        "chunk_id": item.get("chunk_id"),
                        "knowledge_id": item.get("knowledge_id"),
                        "knowledge_title": item.get("knowledge_title"),
                        "score": item.get("score"),
                    }
                    for rank, item in enumerate(items, start=1)
                ],
            }
        )

    failures = [r["case_id"] for r in results if r["error"] or (r["evidence_recall_at_k"] is not None and r["evidence_recall_at_k"] < 1)]
    return {
        "summary": {
            "cases": len(cases),
            "cases_with_required_evidence": len(recalls),
            "evidence_free_cases": evidence_free,
            "api_errors": errors,
            "mean_evidence_recall_at_k": round(mean(recalls), 6) if recalls else None,
            "mean_latency_ms": round(mean(latencies), 3) if latencies else None,
            "p95_latency_ms": round(percentile95(latencies), 3) if percentile95(latencies) is not None else None,
            "failure_case_ids": failures,
        },
        "cases": results,
    }


def compare_rerank_variants(mode_results: dict[str, dict[str, Any]]) -> list[dict[str, Any]]:
    """Return paired deltas without inventing a latency acceptance threshold."""
    comparisons: list[dict[str, Any]] = []
    for mode in DEFAULT_MODES:
        baseline = mode_results.get(mode, {}).get("summary")
        reranked = mode_results.get(f"{mode}+rerank", {}).get("summary")
        if baseline is None or reranked is None:
            continue

        baseline_recall = baseline["mean_evidence_recall_at_k"]
        reranked_recall = reranked["mean_evidence_recall_at_k"]
        recall_delta = (
            round(reranked_recall - baseline_recall, 6)
            if baseline_recall is not None and reranked_recall is not None
            else None
        )
        baseline_p95 = baseline["p95_latency_ms"]
        reranked_p95 = reranked["p95_latency_ms"]
        p95_delta = round(reranked_p95 - baseline_p95, 3) if baseline_p95 is not None and reranked_p95 is not None else None
        error_delta = reranked["api_errors"] - baseline["api_errors"]

        if recall_delta is None:
            recommendation = "manual_review_required"
        elif recall_delta <= 0:
            recommendation = "keep_rerank_off"
        else:
            recommendation = "review_latency_and_errors_before_enabling"

        comparisons.append(
            {
                "mode": mode,
                "baseline": mode,
                "candidate": f"{mode}+rerank",
                "recall_delta": recall_delta,
                "p95_latency_delta_ms": p95_delta,
                "api_error_delta": error_delta,
                "recommendation": recommendation,
            }
        )
    return comparisons


def main() -> int:
    args = parse_args()
    if args.top_k <= 0:
        raise SystemExit("--top-k 必须为正整数")
    modes = tuple(mode.strip() for mode in args.modes.split(",") if mode.strip())
    invalid = set(modes) - set(DEFAULT_MODES)
    if not modes or invalid:
        raise SystemExit(f"--modes 仅支持 {', '.join(DEFAULT_MODES)}，当前无效值: {', '.join(sorted(invalid))}")
    rerank_variant_names = tuple(variant.strip() for variant in args.rerank_variants.split(",") if variant.strip())
    invalid_variants = set(rerank_variant_names) - set(RERANK_VARIANTS)
    if not rerank_variant_names or invalid_variants:
        raise SystemExit(
            f"--rerank-variants 仅支持 {', '.join(sorted(RERANK_VARIANTS))}，当前无效值: {', '.join(sorted(invalid_variants))}"
        )

    dataset_id, cases = load_cases(args.cases)
    mode_results = {
        f"{mode}{'+rerank' if enable_rerank else ''}": evaluate_mode(args, cases, mode, enable_rerank)
        for mode in modes
        for enable_rerank in (RERANK_VARIANTS[variant] for variant in rerank_variant_names)
    }
    report = {
        "report_type": "incidentpilot_retrieval_baseline",
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "dataset_id": dataset_id,
        "configuration": {
            "base_url": args.base_url,
            "kb_id": args.kb_id,
            "top_k": args.top_k,
            "min_score": args.min_score,
            "rerank_variants": list(rerank_variant_names),
            "modes": list(modes),
            "timeout_seconds": args.timeout,
            "token_source": "COGNIDA_TOKEN/--token" if args.token else "none (requires DEV_MODE bypass)",
        },
        "notes": [
            "证据命中通过检索片段正文中的【PI-KB-xxx-x】编号判定。",
            "无必要证据案例不计入 Recall@K，需另行评估拒答行为。",
            "本脚本只调用 POST /api/v1/knowledge/search，不写入知识库、Milvus 或评测集。",
            "启用 on 只用于离线对比；除非 Recall@K 改善且延迟、错误率均可接受，不应修改线上核验台的默认检索设置。",
        ],
        "modes": mode_results,
        "rerank_comparisons": compare_rerank_variants(mode_results),
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    for mode, detail in report["modes"].items():
        summary = detail["summary"]
        print(f"{mode}: Recall@{args.top_k}={summary['mean_evidence_recall_at_k']} P95={summary['p95_latency_ms']}ms failures={len(summary['failure_case_ids'])}")
    print(f"已写入: {args.output}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

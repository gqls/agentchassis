#!/usr/bin/env python3
"""
level1.py — structural analysis of iter_0 vs Claude eval results.

No API. Runs locally on the eval JSONL.

For each row computes:
  - JSON validity (iter_0 and Claude separately)
  - Schema field comparison (when both valid)
  - Length ratio (iter_0 / Claude in characters)
  - Forbidden-phrase hits (extracted from prompt's "Avoid phrases:" list)
  - Fabrication markers (regex heuristics — flags, not verdicts)

Produces level1_metrics.json with per-row data + aggregates.

Usage:
    python level1.py --results iter0_eval_results_v1.jsonl \
                     --output  level1_metrics.json
"""
from __future__ import annotations

import argparse
import json
import re
from pathlib import Path
from typing import Any


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description=__doc__.split("\n", 1)[0])
    p.add_argument("--results", required=True, help="eval results JSONL")
    p.add_argument("--output",  required=True, help="output metrics JSON")
    return p.parse_args()


def safe_json(s: str) -> Any:
    try:
        return json.loads(s.strip())
    except Exception:
        return None


def extract_avoid_phrases(prompt: str) -> list[str]:
    """
    Extract 'Avoid phrases:' bullet list from the brief.
    Returns lowercased phrases, or [] if no parseable list found.
    Inline bracketed lists are not reliably tokenisable so are skipped.
    """
    m = re.search(
        r'Avoid (?:these )?phrases?:?\s*\n((?:\s*[-•]\s*[^\n]+\n?){2,})',
        prompt, re.IGNORECASE,
    )
    if not m:
        return []
    phrases: list[str] = []
    for line in m.group(1).split('\n'):
        line = re.sub(r'^\s*[-•]\s*', '', line).strip()
        if line and len(line) >= 4:
            phrases.append(line.lower())
    return phrases


def find_phrase_hits(text: str, phrases: list[str]) -> list[str]:
    text_low = text.lower()
    return [p for p in phrases if p in text_low]


def fabrication_markers(text: str) -> dict:
    """Regex heuristics. Flags content that *might* be fabricated for L3 review."""
    return {
        "percentages": re.findall(r'\b\d+(?:\.\d+)?\s?%', text),
        "specific_durations": re.findall(
            r'\b\d+[\s-]?(?:minute|hour|day|week|month|year|second)s?\b',
            text, re.IGNORECASE,
        ),
        "specific_counts": re.findall(
            r'\b\d+[\s-]?(?:client|customer|user|company|business|page|deployment)s?\b',
            text, re.IGNORECASE,
        ),
        "named_titles": re.findall(
            r'\b([A-Z][a-z]{2,}\s+[A-Z][a-z]{2,}),?\s+(?:CEO|CTO|Founder|Director|Manager|President|VP|Head\s+of)\b',
            text,
        ),
    }


def analyze_row(row: dict) -> dict:
    iter0 = row.get("iter0_response", "")
    claude = row.get("claude_response", "")
    prompt = row.get("prompt", "")

    iter0_obj = safe_json(iter0)
    claude_obj = safe_json(claude)

    iter0_keys = sorted(iter0_obj.keys()) if isinstance(iter0_obj, dict) else None
    claude_keys = sorted(claude_obj.keys()) if isinstance(claude_obj, dict) else None

    avoid = extract_avoid_phrases(prompt)
    iter0_fab = fabrication_markers(iter0)
    claude_fab = fabrication_markers(claude)

    return {
        "source_log_id":            row.get("source_log_id"),
        "iter0_valid_json":         iter0_obj is not None,
        "claude_valid_json":        claude_obj is not None,
        "iter0_keys":               iter0_keys,
        "claude_keys":              claude_keys,
        "schema_match":             (iter0_keys == claude_keys) if (iter0_keys and claude_keys) else None,
        "iter0_extra_fields":       sorted(set(iter0_keys or []) - set(claude_keys or [])),
        "iter0_missing_fields":     sorted(set(claude_keys or []) - set(iter0_keys or [])),
        "iter0_chars":              len(iter0),
        "claude_chars":             len(claude),
        "length_ratio":             round(len(iter0) / len(claude), 2) if claude else None,
        "iter0_gen_tokens":         row.get("generation_tokens"),
        "iter0_gen_time_s":         row.get("generation_time_s"),
        "avoid_phrases_extracted":  len(avoid),
        "iter0_forbidden_hits":     find_phrase_hits(iter0, avoid),
        "claude_forbidden_hits":    find_phrase_hits(claude, avoid),
        "iter0_fabrication":        {k: v for k, v in iter0_fab.items() if v},
        "claude_fabrication":       {k: v for k, v in claude_fab.items() if v},
        "iter0_fab_count":          sum(len(v) for v in iter0_fab.values()),
        "claude_fab_count":         sum(len(v) for v in claude_fab.values()),
    }


def aggregate(rows: list[dict]) -> dict:
    n = len(rows)
    valid_pairs = [r for r in rows if r["iter0_valid_json"] and r["claude_valid_json"]]
    rows_with_avoid = [r for r in rows if r["avoid_phrases_extracted"] > 0]

    length_ratios = sorted(r["length_ratio"] for r in valid_pairs if r["length_ratio"])

    return {
        "n_total":                              n,
        "n_iter0_valid_json":                   sum(r["iter0_valid_json"] for r in rows),
        "n_claude_valid_json":                  sum(r["claude_valid_json"] for r in rows),
        "n_both_valid":                         len(valid_pairs),
        "n_schema_match":                       sum(1 for r in valid_pairs if r["schema_match"]),
        "n_rows_with_extracted_avoid_list":     len(rows_with_avoid),
        "iter0_forbidden_phrase_hits_total":    sum(len(r["iter0_forbidden_hits"]) for r in rows),
        "claude_forbidden_phrase_hits_total":   sum(len(r["claude_forbidden_hits"]) for r in rows),
        "iter0_rows_with_forbidden":            sum(1 for r in rows if r["iter0_forbidden_hits"]),
        "claude_rows_with_forbidden":           sum(1 for r in rows if r["claude_forbidden_hits"]),
        "iter0_fabrication_total":              sum(r["iter0_fab_count"] for r in rows),
        "claude_fabrication_total":             sum(r["claude_fab_count"] for r in rows),
        "iter0_rows_with_fabrication":          sum(1 for r in rows if r["iter0_fab_count"] > 0),
        "claude_rows_with_fabrication":         sum(1 for r in rows if r["claude_fab_count"] > 0),
        "length_ratio_mean":                    round(sum(length_ratios) / len(length_ratios), 2) if length_ratios else None,
        "length_ratio_median":                  length_ratios[len(length_ratios)//2] if length_ratios else None,
        "length_ratio_min":                     length_ratios[0] if length_ratios else None,
        "length_ratio_max":                     length_ratios[-1] if length_ratios else None,
        "iter0_mean_gen_time_s":                round(sum(r["iter0_gen_time_s"] or 0 for r in rows) / n, 1) if n else None,
        "iter0_mean_gen_tokens":                round(sum(r["iter0_gen_tokens"] or 0 for r in rows) / n, 1) if n else None,
    }


def main() -> None:
    args = parse_args()

    rows = []
    with open(args.results) as f:
        for line in f:
            line = line.strip()
            if line:
                rows.append(json.loads(line))

    print(f"Analyzing {len(rows)} rows...")
    per_row = [analyze_row(r) for r in rows]
    aggregates = aggregate(per_row)

    out = {"per_row": per_row, "aggregates": aggregates}
    Path(args.output).write_text(json.dumps(out, indent=2, ensure_ascii=False))

    a = aggregates
    print(f"\nLevel 1 — structural ({a['n_total']} rows)")
    print(f"  JSON valid:   iter_0 {a['n_iter0_valid_json']}/{a['n_total']}   "
          f"Claude {a['n_claude_valid_json']}/{a['n_total']}")
    print(f"  Both valid:   {a['n_both_valid']}/{a['n_total']}")
    print(f"  Schema match: {a['n_schema_match']}/{a['n_both_valid']}")
    if a['length_ratio_mean']:
        print(f"  Length (iter_0/Claude):  mean {a['length_ratio_mean']}  median {a['length_ratio_median']}  "
              f"range {a['length_ratio_min']}-{a['length_ratio_max']}")
    print(f"  Forbidden hits:   iter_0 {a['iter0_forbidden_phrase_hits_total']} ({a['iter0_rows_with_forbidden']} rows)   "
          f"Claude {a['claude_forbidden_phrase_hits_total']} ({a['claude_rows_with_forbidden']} rows)   "
          f"[from {a['n_rows_with_extracted_avoid_list']} parseable avoid-lists]")
    print(f"  Fabrication markers:  iter_0 {a['iter0_fabrication_total']} ({a['iter0_rows_with_fabrication']} rows)   "
          f"Claude {a['claude_fabrication_total']} ({a['claude_rows_with_fabrication']} rows)")
    print(f"\nWritten: {args.output}")


if __name__ == "__main__":
    main()


    
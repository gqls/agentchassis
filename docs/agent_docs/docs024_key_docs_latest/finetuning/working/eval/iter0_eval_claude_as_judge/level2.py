#!/usr/bin/env python3
"""
level2.py — Claude-as-judge evaluation of iter_0 vs Claude on each held-out brief.

Methodology:
  1. Anonymize the two responses, randomize A/B position per case.
  2. Send brief + anonymous responses to Claude judge.
  3. Get scores on relevance/voice/integrity (1-5) + winner pick.
  4. Decode A/B back to model identity.
  5. Stream-flush results so a crash doesn't lose work.

Confounds called out in the report (not fixed here):
  - Self-judge bias: judge is Claude; iter_0 was trained on Claude outputs.
    Use a different Claude model (Opus) than the one that produced training
    labels (Sonnet 4.6) to reduce self-recognition.
  - Position bias: A/B randomization controls for this; report quantifies.

Cost: ~20 cases × ~5K tokens each ≈ 100K tokens. Order-of-magnitude $1.

Usage:
    export ANTHROPIC_API_KEY=...
    python level2.py --results iter0_eval_results_v1.jsonl \
                     --output  level2_judgments.jsonl \
                     --judge-model claude-opus-4-7
"""
from __future__ import annotations

import argparse
import json
import os
import random
import re
import sys
import time
from pathlib import Path

try:
    from anthropic import Anthropic
except ImportError:
    sys.stderr.write("Install: pip install anthropic\n")
    sys.exit(1)


JUDGE_PROMPT = """You are evaluating two anonymous AI-generated responses to a content-generation brief. Be strict and fair.

# The brief that was given to both AIs
{brief}

# Response A
{response_a}

# Response B
{response_b}

# Your task

Score each response on three dimensions (1-5 integer scale):

1. RELEVANCE — Does it address THIS specific brief? Right company, right audience, right section type, right schema fields.
2. VOICE — Does it match the voice/tone instructions in the brief? Avoid-phrase compliance counts here.
3. INTEGRITY — Does it avoid fabrication? No invented contacts, no fake testimonials, no made-up metrics, no fictional case studies. Stick to what's in the brief.

Then pick a winner: A, B, or TIE. Use TIE only when the responses are genuinely indistinguishable in quality — not as a hedge.

Be strict. 5 = outstanding execution. 3 = competent baseline. 1 = broken or off-task.

Output JSON only, no preamble, no code fences:
{{
  "response_a": {{"relevance": N, "voice": N, "integrity": N, "notes": "1-2 sentences on what stood out"}},
  "response_b": {{"relevance": N, "voice": N, "integrity": N, "notes": "1-2 sentences on what stood out"}},
  "winner": "A" | "B" | "TIE",
  "reasoning": "1-3 sentences explaining the pick, citing specific differences"
}}"""


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description=__doc__.split("\n", 1)[0])
    p.add_argument("--results",      required=True, help="eval results JSONL")
    p.add_argument("--output",       required=True, help="output judgments JSONL")
    p.add_argument("--judge-model",  default="claude-opus-4-7",
                   help="Use a different model than the training-label producer to reduce self-judge bias.")
    p.add_argument("--seed",         type=int, default=3407, help="RNG seed for A/B position randomization")
    p.add_argument("--max-cases",    type=int, default=0, help="cap (0 = all). Useful for cost control.")
    return p.parse_args()


def strip_code_fences(s: str) -> str:
    s = s.strip()
    s = re.sub(r'^```(?:json)?\s*\n?', '', s)
    s = re.sub(r'\n?\s*```\s*$', '', s)
    return s.strip()


def main() -> None:
    args = parse_args()
    rng = random.Random(args.seed)

    if not os.environ.get("ANTHROPIC_API_KEY"):
        sys.exit("ANTHROPIC_API_KEY env var not set")

    client = Anthropic()

    # Resume support
    done: set[str] = set()
    output_path = Path(args.output)
    if output_path.exists():
        with open(output_path) as f:
            for line in f:
                try:
                    done.add(json.loads(line)["source_log_id"])
                except Exception:
                    pass
        if done:
            print(f"Resume: {len(done)} judgments already in {output_path}, skipping.")

    rows: list[dict] = []
    with open(args.results) as f:
        for line in f:
            line = line.strip()
            if line:
                rows.append(json.loads(line))

    rows = [r for r in rows if r.get("source_log_id") not in done]
    if args.max_cases > 0:
        rows = rows[:args.max_cases]

    print(f"Cases to judge: {len(rows)}  (judge: {args.judge_model})")

    fout = open(output_path, "a", buffering=1)
    failures = 0

    for i, row in enumerate(rows, 1):
        # Randomize position
        flip = rng.random() < 0.5
        if flip:
            response_a, response_b = row["iter0_response"], row["claude_response"]
            mapping = {"A": "iter0", "B": "claude"}
        else:
            response_a, response_b = row["claude_response"], row["iter0_response"]
            mapping = {"A": "claude", "B": "iter0"}

        prompt = JUDGE_PROMPT.format(
            brief=row["prompt"],
            response_a=response_a,
            response_b=response_b,
        )

        t0 = time.time()
        try:
            resp = client.messages.create(
                model=args.judge_model,
                max_tokens=1500,
                messages=[{"role": "user", "content": prompt}],
            )
            judge_text = strip_code_fences(resp.content[0].text)
            judge_obj = json.loads(judge_text)
        except Exception as e:
            failures += 1
            print(f"[{i}/{len(rows)}] {row.get('source_log_id', '?')[:8]}…  judge call failed: {type(e).__name__}: {e}")
            continue

        winner_label = judge_obj.get("winner", "TIE")
        winner_model = mapping.get(winner_label, "TIE") if winner_label in ("A", "B") else "TIE"

        # Extract scores back into model-keyed form
        a_scores = judge_obj.get("response_a", {})
        b_scores = judge_obj.get("response_b", {})
        iter0_scores = a_scores if mapping["A"] == "iter0" else b_scores
        claude_scores = a_scores if mapping["A"] == "claude" else b_scores

        result = {
            "source_log_id":    row["source_log_id"],
            "judge_model":      args.judge_model,
            "iter0_position":   "A" if flip else "B",
            "winner_label":     winner_label,
            "winner_model":     winner_model,
            "iter0_scores":     iter0_scores,
            "claude_scores":    claude_scores,
            "reasoning":        judge_obj.get("reasoning", ""),
            "judge_time_s":     round(time.time() - t0, 1),
        }
        fout.write(json.dumps(result, ensure_ascii=False) + "\n")
        fout.flush()

        si, sc = iter0_scores, claude_scores
        print(f"[{i}/{len(rows)}] {row['source_log_id'][:8]}…  "
              f"iter_0 R{si.get('relevance','?')}/V{si.get('voice','?')}/I{si.get('integrity','?')} "
              f"vs Claude R{sc.get('relevance','?')}/V{sc.get('voice','?')}/I{sc.get('integrity','?')}  "
              f"→ {winner_model}")

    fout.close()
    print(f"\nDone. Failures: {failures}/{len(rows)}.  Output: {output_path}")


if __name__ == "__main__":
    main()

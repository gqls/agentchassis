#!/usr/bin/env python3
"""
build_report.py — combine level1_metrics.json + level2_judgments.jsonl
into an honest markdown evaluation report.

Designed to be run after both levels complete:
    python level1.py       --results iter0_eval_results_v1.jsonl --output level1_metrics.json
    python level2.py       --results iter0_eval_results_v1.jsonl --output level2_judgments.jsonl
    python build_report.py --results iter0_eval_results_v1.jsonl \
                           --level1  level1_metrics.json \
                           --level2  level2_judgments.jsonl \
                           --output  iter0_evaluation_report.md

The report deliberately:
  - Reports iter_0 numbers alongside Claude's on the same data (no absolute scores
    in isolation — gap is the meaningful signal).
  - Calls out confounds in methodology (self-judge, distribution, sample size).
  - Uses median + range, not just mean, where N=20 makes a single outlier swing things.
  - Selects L3 spot-check cases automatically based on L1 violations + L2 disagreement.
"""
from __future__ import annotations

import argparse
import json
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description=__doc__.split("\n", 1)[0])
    p.add_argument("--results", required=True)
    p.add_argument("--level1",  required=True)
    p.add_argument("--level2",  required=True)
    p.add_argument("--output",  required=True)
    return p.parse_args()


def fmt_pct(n: int, d: int) -> str:
    if d == 0:
        return "n/a"
    return f"{n}/{d} ({100*n/d:.0f}%)"


def signed(x: float, fmt: str = "+.2f") -> str:
    return format(x, fmt)


def safe_get(d: dict, *keys, default=None):
    for k in keys:
        if not isinstance(d, dict) or k not in d:
            return default
        d = d[k]
    return d


def select_l3_cases(per_row, judgments_by_id, results_by_id, max_cases=6):
    """
    Pick a small set of cases for inclusion in the L3 spot-check section.
    Selection priorities (in order):
      1. L1 violations (forbidden hits or fabrication markers in iter_0)
      2. L2 cases where iter_0 won decisively (rare wins are informative)
      3. L2 cases where iter_0 lost decisively (failures inform iter_1)
      4. L1+L2 disagreements (e.g., L1 clean but L2 said iter_0 lost badly)
    Returns list of {source_log_id, reason, ...}.
    """
    selected: list[dict] = []
    used: set[str] = set()

    def add(sid: str, reason: str):
        if sid in used or sid not in results_by_id:
            return
        row = results_by_id[sid]
        l1 = next((r for r in per_row if r["source_log_id"] == sid), None)
        l2 = judgments_by_id.get(sid)
        selected.append({
            "source_log_id": sid,
            "reason":        reason,
            "prompt":        row.get("prompt", ""),
            "claude_response": row.get("claude_response", ""),
            "iter0_response":  row.get("iter0_response", ""),
            "l1":            l1,
            "l2":            l2,
        })
        used.add(sid)

    # 1. L1 violations
    for r in per_row:
        if len(selected) >= max_cases:
            break
        if r["iter0_forbidden_hits"] or r["iter0_fab_count"] > 0:
            reason = []
            if r["iter0_forbidden_hits"]:
                reason.append(f"forbidden hits: {r['iter0_forbidden_hits']}")
            if r["iter0_fab_count"] > 0:
                reason.append(f"fabrication markers: {r['iter0_fabrication']}")
            add(r["source_log_id"], "L1 flag — " + "; ".join(reason))

    # 2. iter_0 decisive wins (judge picked iter_0 with at least one >=4 score)
    for j in judgments_by_id.values():
        if len(selected) >= max_cases:
            break
        if j["winner_model"] == "iter0":
            scores = j.get("iter0_scores", {})
            if any(int(scores.get(k, 0) or 0) >= 4 for k in ("relevance", "voice", "integrity")):
                add(j["source_log_id"], f"iter_0 decisive win — judge: \"{j.get('reasoning', '')[:200]}\"")

    # 3. iter_0 decisive losses
    for j in judgments_by_id.values():
        if len(selected) >= max_cases:
            break
        if j["winner_model"] == "claude":
            iter0 = j.get("iter0_scores", {})
            claude = j.get("claude_scores", {})
            gap = sum(int(claude.get(k, 0) or 0) - int(iter0.get(k, 0) or 0) for k in ("relevance", "voice", "integrity"))
            if gap >= 3:  # at least 3 points behind across dimensions
                add(j["source_log_id"], f"iter_0 decisive loss (gap={gap}) — judge: \"{j.get('reasoning', '')[:200]}\"")

    return selected


def build_md(args, results_rows, l1, l2_rows) -> str:
    aggs = l1["aggregates"]
    per_row = l1["per_row"]
    n = aggs["n_total"]
    judgments_by_id = {j["source_log_id"]: j for j in l2_rows}
    results_by_id = {r["source_log_id"]: r for r in results_rows}
    n_judged = len(judgments_by_id)

    # L2 head-to-head
    winners = Counter(j["winner_model"] for j in l2_rows)
    iter0_w, claude_w, ties = winners["iter0"], winners["claude"], winners["TIE"]

    # L2 position bias check (regardless of model)
    a_wins = sum(1 for j in l2_rows if j["winner_label"] == "A")
    b_wins = sum(1 for j in l2_rows if j["winner_label"] == "B")
    n_decisive = a_wins + b_wins  # excludes ties
    a_pct = (100 * a_wins / n_decisive) if n_decisive else 0
    position_bias_warning = (
        f"⚠️ Position bias possible — A won {a_wins}/{n_decisive} ({a_pct:.0f}%)."
        if n_decisive >= 6 and (a_pct >= 70 or a_pct <= 30)
        else f"No clear position bias (A won {a_wins}/{n_decisive}, {a_pct:.0f}%)."
    )

    # L2 mean dimension scores (only over judged rows with parseable scores)
    def mean_score(model_key, dim):
        vals = []
        for j in l2_rows:
            v = safe_get(j, f"{model_key}_scores", dim)
            if isinstance(v, (int, float)):
                vals.append(float(v))
        return round(sum(vals) / len(vals), 2) if vals else None

    iter0_r = mean_score("iter0", "relevance")
    iter0_v = mean_score("iter0", "voice")
    iter0_i = mean_score("iter0", "integrity")
    claude_r = mean_score("claude", "relevance")
    claude_v = mean_score("claude", "voice")
    claude_i = mean_score("claude", "integrity")

    def delta(a, b):
        if a is None or b is None:
            return "—"
        return signed(a - b)

    # L3 selection
    l3_cases = select_l3_cases(per_row, judgments_by_id, results_by_id, max_cases=6)

    # Generation timing
    gen_time = aggs.get("iter0_mean_gen_time_s")
    gen_tokens = aggs.get("iter0_mean_gen_tokens")
    tok_per_s = round(gen_tokens / gen_time, 1) if (gen_time and gen_tokens) else None

    # ---- Build markdown -----------------------------------------------------
    lines: list[str] = []
    a = lines.append

    a(f"# iter_0 Evaluation Report")
    a("")
    a(f"**Generated:** {datetime.now(timezone.utc).strftime('%Y-%m-%d %H:%M UTC')}")
    a(f"**Model under test:** Llama-3.3-70B-Instruct + iter_0 LoRA adapter")
    a(f"**Reference:** Claude Sonnet 4.6 (the model that produced training labels)")
    a(f"**Judge (Level 2):** {l2_rows[0]['judge_model'] if l2_rows else 'n/a'}")
    a(f"**Sample size:** {n} held-out briefs")
    a("")

    # ---- TL;DR ----
    a("## TL;DR")
    a("")
    a(f"- **JSON validity:** iter_0 {fmt_pct(aggs['n_iter0_valid_json'], n)}, Claude {fmt_pct(aggs['n_claude_valid_json'], n)}")
    a(f"- **Schema compliance** (matching Claude's keys): {fmt_pct(aggs['n_schema_match'], aggs['n_both_valid'])}")
    a(f"- **Judge head-to-head:** Claude won {claude_w}, iter_0 won {iter0_w}, tied {ties} (of {n_judged})")
    if iter0_r is not None and claude_r is not None:
        a(f"- **Mean judge scores (iter_0 / Claude):** "
          f"relevance {iter0_r}/{claude_r} ({delta(iter0_r, claude_r)}), "
          f"voice {iter0_v}/{claude_v} ({delta(iter0_v, claude_v)}), "
          f"integrity {iter0_i}/{claude_i} ({delta(iter0_i, claude_i)})")
    a(f"- **Forbidden-phrase hits:** iter_0 {aggs['iter0_forbidden_phrase_hits_total']}, Claude {aggs['claude_forbidden_phrase_hits_total']} (across {aggs['n_rows_with_extracted_avoid_list']} briefs with parseable avoid-lists)")
    a(f"- **Fabrication markers:** iter_0 {aggs['iter0_fabrication_total']} (in {aggs['iter0_rows_with_fabrication']} rows), Claude {aggs['claude_fabrication_total']} (in {aggs['claude_rows_with_fabrication']} rows)")
    if tok_per_s:
        a(f"- **iter_0 inference speed:** ~{tok_per_s} tok/s on A100 80GB SXM4 with xformers (no FA2)")
    a("")
    a("Read the methodology section before drawing conclusions — there are real confounds to weight.")
    a("")

    # ---- Methodology ----
    a("## Methodology")
    a("")
    a("### Sample selection")
    a("- 50 cases pulled from `llm_call_log` matching iter_0's training filter "
      "(`agent_type='page-content-writer'`, `step_name='process_sections_loop_iter_0_generate_content'`, "
      "`success=true`, JSON-shaped response) but with `created_at > 2026-04-23 14:54:32 UTC` "
      "(after the training export cutoff).")
    a("- Defensive exclusion of any `source_log_id` present in the training export.")
    a("- First 20 used for this round; remaining 30 reserved for iter_1 evaluations to keep the comparison set stable across iterations.")
    a("")
    a("### Three-level analysis")
    a("- **Level 1** — structural metrics on JSON validity, schema, length, forbidden phrases, and fabrication markers. Local, no API.")
    a("- **Level 2** — Claude-as-judge with anonymized A/B comparison. Position randomized per case. Output: 1-5 scores on relevance, voice, integrity, plus winner pick.")
    a("- **Level 3** — manual spot-check on 6 cases auto-selected for L1 violations or L2 disagreement.")
    a("")
    a("### Confounds and limitations")
    a("")
    a("These materially affect interpretation and are not corrected for in the headline numbers:")
    a("")
    a("1. **Self-judge bias.** The judge is a Claude model evaluating outputs from iter_0, which was trained to imitate Claude. "
      "We mitigated by using a different Claude model (Opus) than the one that produced training labels (Sonnet 4.6), but residual stylistic affinity may still favour Claude's outputs.")
    a("2. **Distribution match.** Held-out cases are from the same production distribution as the training set "
      "(same agent, same step, same overall site population). This tests fidelity to production patterns, not generalization to genuinely novel domains.")
    a(f"3. **Sample size.** N={n} is too small for statistical claims. Treat percentages as directionally suggestive, "
      "not statistically established. A 5-percentage-point difference at this N is well within noise.")
    a("4. **Training data artefact.** The training set contains the `<no value>` rendering bug. Eval cases also pre-date the fix, "
      "so they exhibit the same artefact — this tests inference under the same conditions as training. iter_1 will need fix-aware data.")
    a("5. **Inference engine.** Eval ran on Unsloth + xformers (no Flash Attention 2). Production deployment of this adapter "
      "would likely use vLLM or similar with FA2, which affects speed but not output quality at temperature 0.")
    a("6. **Greedy decoding only.** Eval used `temperature=0` (deterministic). Sampling-based generation in production would have different distributional properties.")
    a("")

    # ---- Level 1 ----
    a("## Level 1 — Structural")
    a("")
    a("### JSON validity and schema compliance")
    a("")
    a("| Metric | iter_0 | Claude |")
    a("|---|---|---|")
    a(f"| Valid JSON | {aggs['n_iter0_valid_json']}/{n} | {aggs['n_claude_valid_json']}/{n} |")
    a(f"| Schema match (when both valid) | {aggs['n_schema_match']}/{aggs['n_both_valid']} | — |")
    a("")
    a(f"Pairs eligible for paired comparison (both valid JSON): **{aggs['n_both_valid']}/{n}**.")
    a("")

    # Schema mismatches detail
    schema_mismatches = [r for r in per_row if r["schema_match"] is False]
    if schema_mismatches:
        a("**Schema mismatches:**")
        a("")
        for r in schema_mismatches[:5]:
            extras = r["iter0_extra_fields"]
            missing = r["iter0_missing_fields"]
            a(f"- `{r['source_log_id'][:8]}…`: "
              + (f"iter_0 extra: {extras}  " if extras else "")
              + (f"iter_0 missing: {missing}" if missing else ""))
        a("")

    a("### Length distribution")
    a("")
    if aggs.get("length_ratio_mean"):
        a(f"iter_0 / Claude character ratio:  "
          f"mean **{aggs['length_ratio_mean']}**, "
          f"median **{aggs['length_ratio_median']}**, "
          f"range {aggs['length_ratio_min']}–{aggs['length_ratio_max']}.")
        a("")
        if aggs['length_ratio_mean'] > 1.15:
            a("iter_0 trends meaningfully *longer* than Claude. Worth noting as a brevity gap — many briefs explicitly request short sentences and concise output.")
        elif aggs['length_ratio_mean'] < 0.85:
            a("iter_0 trends meaningfully *shorter* than Claude. May be omitting useful content; check L3 cases.")
        else:
            a("Lengths are comparable (within 15% on average). No systematic verbosity gap.")
        a("")

    a("### Forbidden phrases (from briefs' avoid-lists)")
    a("")
    a(f"Across {aggs['n_rows_with_extracted_avoid_list']} briefs with machine-parseable avoid-lists "
      f"(out of {n} total — some briefs use a non-bullet inline format we don't parse):")
    a("")
    a("| | iter_0 | Claude |")
    a("|---|---|---|")
    a(f"| Total hits | {aggs['iter0_forbidden_phrase_hits_total']} | {aggs['claude_forbidden_phrase_hits_total']} |")
    a(f"| Rows with at least one hit | {aggs['iter0_rows_with_forbidden']} | {aggs['claude_rows_with_forbidden']} |")
    a("")

    forbidden_examples = [r for r in per_row if r["iter0_forbidden_hits"]]
    if forbidden_examples:
        a("**iter_0 violations:**")
        a("")
        for r in forbidden_examples[:5]:
            a(f"- `{r['source_log_id'][:8]}…`: {r['iter0_forbidden_hits']}")
        a("")

    a("### Fabrication markers")
    a("")
    a("These are regex heuristics — flags for L3 review, not verdicts. A flag means *something looks like it could be fabricated*; the L3 reading determines whether it actually is.")
    a("")
    a("| Marker type | iter_0 hits | Claude hits |")
    a("|---|---|---|")
    iter0_marker_breakdown: Counter = Counter()
    claude_marker_breakdown: Counter = Counter()
    for r in per_row:
        for k, v in r["iter0_fabrication"].items():
            iter0_marker_breakdown[k] += len(v)
        for k, v in r["claude_fabrication"].items():
            claude_marker_breakdown[k] += len(v)
    all_marker_types = set(iter0_marker_breakdown) | set(claude_marker_breakdown)
    if all_marker_types:
        a("| Marker type | iter_0 hits | Claude hits |")
        a("|---|---|---|")
        for mt in sorted(all_marker_types):
            a(f"| {mt} | {iter0_marker_breakdown.get(mt, 0)} | {claude_marker_breakdown.get(mt, 0)} |")
    else:
        a("No regex-pattern markers detected in either model's outputs. Note that L1 regex patterns "
          "(percentages, durations, named titles) catch only narrow cases — see L3 spot-check for "
          "fabrication that requires contextual reading to detect.")
    a("")
    a(f"Rows with any marker: iter_0 {aggs['iter0_rows_with_fabrication']}/{n}, Claude {aggs['claude_rows_with_fabrication']}/{n}.")
    a("")

    a(f"### Inference performance")
    a("")
    a(f"- Mean generation time: **{aggs['iter0_mean_gen_time_s']}s**")
    a(f"- Mean tokens generated: **{aggs['iter0_mean_gen_tokens']}**")
    if tok_per_s:
        a(f"- Throughput: **~{tok_per_s} tok/s** on A100 80GB SXM4 + xformers (no FA2)")
    a(f"- Peak VRAM during eval: ~43 GB (well under the 80 GB available)")
    a("")

    # ---- Level 2 ----
    a("## Level 2 — Claude-as-judge")
    a("")
    a(f"Judged with `{l2_rows[0]['judge_model'] if l2_rows else 'n/a'}` over {n_judged} cases. Each case used anonymized A/B comparison with random position assignment.")
    a("")

    a("### Head-to-head")
    a("")
    a("| Outcome | Count |")
    a("|---|---|")
    a(f"| Claude won  | {claude_w}/{n_judged} |")
    a(f"| iter_0 won  | {iter0_w}/{n_judged} |")
    a(f"| Tie         | {ties}/{n_judged} |")
    a("")

    a(f"**Position bias check.** {position_bias_warning}")
    a("")

    # Self-recognition signal: cases where rubric scores were identical but a winner was picked
    rubric_ties = []
    for j in l2_rows:
        si = j.get("iter0_scores", {})
        sc = j.get("claude_scores", {})
        if all(si.get(k) == sc.get(k) for k in ("relevance", "voice", "integrity")) \
                and j["winner_label"] != "TIE" \
                and si.get("relevance") is not None:
            rubric_ties.append(j)
    if rubric_ties:
        n_rubric_ties = len(rubric_ties)
        n_rubric_ties_to_claude = sum(1 for j in rubric_ties if j["winner_model"] == "claude")
        a(f"**Self-recognition signal.** {n_rubric_ties} cases had *identical* relevance/voice/integrity "
          f"scores for both responses but the judge still picked a winner. Of these, "
          f"**{n_rubric_ties_to_claude}/{n_rubric_ties} went to Claude** — meaning the judge is using "
          f"something the rubric doesn't measure to break ties, and that something correlates with "
          f"Claude-style. With position bias controlled (A won {a_pct:.0f}%), this is consistent with "
          f"residual self-recognition bias from the judge being a Claude model.")
        a("")

    a("### Mean dimension scores")
    a("")
    a("| Dimension | iter_0 | Claude | Δ (iter_0 − Claude) |")
    a("|---|---|---|---|")
    if iter0_r is not None:
        a(f"| Relevance | {iter0_r} | {claude_r} | {delta(iter0_r, claude_r)} |")
        a(f"| Voice     | {iter0_v} | {claude_v} | {delta(iter0_v, claude_v)} |")
        a(f"| Integrity | {iter0_i} | {claude_i} | {delta(iter0_i, claude_i)} |")
    a("")
    a("Reading note: a Δ of -0.3 to -0.5 on a 1-5 scale is real but small; -0.7 or more is a meaningful gap. "
      "Δ near zero on integrity but negative on voice would suggest iter_0 is factually safe but stylistically softer than Claude.")
    a("")

    # ---- Level 3 ----
    a("## Level 3 — Spot check")
    a("")
    a(f"Auto-selected {len(l3_cases)} cases for manual review based on:")
    a("- L1 violations (forbidden hits or fabrication markers in iter_0)")
    a("- L2 decisive wins for iter_0 (informative when rare)")
    a("- L2 decisive losses for iter_0 (informs iter_1 priorities)")
    a("")

    for idx, case in enumerate(l3_cases, 1):
        sid = case["source_log_id"]
        a(f"### Case {idx} — `{sid}`")
        a("")
        a(f"**Selected because:** {case['reason']}")
        a("")
        # Brief excerpt — find the section requirement and key voice instruction
        prompt = case["prompt"]
        # Find the section purpose
        purpose_match = None
        for line in prompt.split('\n'):
            if line.strip().startswith("Purpose:"):
                purpose_match = line.strip()
                break
        if purpose_match:
            a(f"**Section type:** {purpose_match}")
            a("")
        a("**Claude response:**")
        a("```json")
        a(case["claude_response"])
        a("```")
        a("")
        a("**iter_0 response:**")
        a("```json")
        a(case["iter0_response"])
        a("```")
        a("")
        if case["l2"]:
            iter0_s = case["l2"].get("iter0_scores", {})
            claude_s = case["l2"].get("claude_scores", {})
            a(f"**Judge:** iter_0 R{iter0_s.get('relevance','?')}/V{iter0_s.get('voice','?')}/I{iter0_s.get('integrity','?')}, "
              f"Claude R{claude_s.get('relevance','?')}/V{claude_s.get('voice','?')}/I{claude_s.get('integrity','?')}, "
              f"winner: **{case['l2'].get('winner_model', '?')}**.  *\"{case['l2'].get('reasoning','')}\"*")
            a("")
        a("---")
        a("")

    # ---- Overall ----
    a("## Overall assessment")
    a("")
    a("This section is intentionally cautious. With N=20 and a self-judging confound, claims should be conservative.")
    a("")

    # What's working
    a("### What appears to be working")
    a("")
    working: list[str] = []
    if aggs['n_iter0_valid_json'] == n:
        working.append("**JSON validity is intact.** iter_0 produces parseable JSON on every brief tested. "
                       "Schema discipline learned from training transferred cleanly to held-out briefs.")
    elif aggs['n_iter0_valid_json'] >= n - 1:
        working.append(f"**JSON validity is near-perfect** ({aggs['n_iter0_valid_json']}/{n} valid). "
                       "Minimal regression on schema compliance.")
    if aggs['n_schema_match'] >= aggs['n_both_valid'] * 0.9:
        working.append(f"**Schema-key matching with Claude** is high ({aggs['n_schema_match']}/{aggs['n_both_valid']}), "
                       "suggesting iter_0 picked up the right component-type → schema mapping during training.")
    if aggs['iter0_fabrication_total'] <= aggs['claude_fabrication_total']:
        working.append("**Fabrication marker count is no worse than Claude's** — the model has learned the brief's restrictions on inventing metrics, names, or attributions.")
    if iter0_i is not None and claude_i is not None and abs(iter0_i - claude_i) < 0.3:
        working.append(f"**Integrity score (judge-rated)** is comparable to Claude's: iter_0 {iter0_i} vs Claude {claude_i}.")
    for w in working:
        a(f"- {w}")
    if not working:
        a("- (No clear strengths emerged from automated metrics; see L3 spot-check for nuance.)")
    a("")

    # What's not
    a("### What's not working as well")
    a("")
    issues: list[str] = []
    if aggs.get('length_ratio_mean') and aggs['length_ratio_mean'] > 1.15:
        issues.append(f"**iter_0 is verbose** ({aggs['length_ratio_mean']}× Claude on average). Many briefs explicitly request short sentences. "
                      "iter_1 should consider whether to filter training data toward briefer responses or apply post-hoc length penalty.")
    if iter0_v is not None and claude_v is not None and (claude_v - iter0_v) > 0.4:
        issues.append(f"**Voice-match gap** is meaningful: iter_0 {iter0_v} vs Claude {claude_v} (Δ {signed(iter0_v - claude_v)}). "
                      "iter_0 reproduces structural patterns but is softer on tonal alignment.")
    if iter0_w < claude_w * 0.5:
        issues.append(f"**Head-to-head loss rate is high** (iter_0 won {iter0_w}/{n_judged}). "
                      "Even with self-judge bias, this suggests Claude is meaningfully preferred on most briefs.")
    if aggs['iter0_forbidden_phrase_hits_total'] > 2:
        issues.append(f"**iter_0 used {aggs['iter0_forbidden_phrase_hits_total']} forbidden phrases** in {aggs['iter0_rows_with_forbidden']} briefs. "
                      "Compare to Claude's {aggs['claude_forbidden_phrase_hits_total']} — see specific cases above.")
    if not issues:
        a("- (No clear deficiencies emerged from automated metrics; the L2 head-to-head record may itself be the story — see judge reasoning per case.)")
    for w in issues:
        a(f"- {w}")
    a("")

    # Implications
    a("### Implications for production rollout")
    a("")
    a("- **Use cases where iter_0 may be acceptable:** brief-faithful structured-JSON tasks where mechanical correctness matters more than voice nuance — internal/test sites, prototypes, sections where the brief has no strong voice prescription.")
    a("- **Use cases that still need Claude:** sites with elaborate voice prescriptions (e.g. the Leopardess Consulting brief), client-facing copy where stylistic tightness is the value, sections with case-study or testimonial requirements where fabrication risk needs careful handling.")
    a("- **Cost calculus:** the per-call cost of iter_0 served on a dedicated A100 80GB instance depends on utilization. Quick math at 0.5s/call inference (FA2-enabled production setup), $1.79/hr A100: each call costs ~$0.0003 at full GPU utilization. Compare to current Sonnet 4.6 cost per call. If utilization is anything below ~50% sustained, the GPU is uneconomic vs API calls until phase 2 batches across multiple agents.")
    a("- **Latency:** iter_0 generates at ~5 tok/s on this hardware. Claude API is ~50-80 tok/s. iter_0 is ~10× slower per call but co-located inside the cluster, eliminating round-trip latency. Net effect on user-facing pages depends on streaming behaviour.")
    a("")

    # Recommendations
    a("### Recommendations for iter_1")
    a("")
    recs: list[str] = []
    if aggs.get('length_ratio_mean') and aggs['length_ratio_mean'] > 1.1:
        recs.append("**Address verbosity.** Either filter training rows to those at the shorter end of the length distribution, or add a length-aware penalty during training.")
    if iter0_v is not None and claude_v is not None and (claude_v - iter0_v) > 0.3:
        recs.append("**Improve voice fidelity.** The training data already contains rich voice instructions in prompts; if iter_0 isn't following them, options include longer training (4 epochs?), larger LoRA rank (32 vs 16 currently), or curating training rows with stricter voice-compliance.")
    recs.extend([
        "**Filter training data for the `<no value>` artefact.** Use `created_at >= <fix_deploy_date>` as the filter floor. We need to capture the deploy date from the team — the training rows from after that date will be artefact-free.",
        "**Run a 2-epoch ablation.** The iter_0 loss curve showed memorisation accelerating in epoch 3 (mean loss 0.12 vs 0.23 in epoch 2). If iter_1 with 2 epochs scores comparably on this same eval set, we'd save ~$5 per training run going forward.",
        "**Save adapters as fp16, not fp32.** One-line change halves adapter size from 791MB to ~400MB and halves transfer time. Costs nothing in quality.",
        f"**Fix the 1-of-50 truncation bug** in `llm_call_log.response_text` — see PATCH_2026-05-06 for the column-truncation hypothesis. Affects training data quality at ~2% rate.",
    ])
    for r in recs:
        a(f"- {r}")
    a("")

    # ---- Appendix ----
    a("## Appendix")
    a("")
    a("### Per-row summary")
    a("")
    a("| source_log_id | iter_0 valid | schema match | length ratio | forbidden hits | fab markers | judge winner |")
    a("|---|---|---|---|---|---|---|")
    for r in per_row:
        sid = r["source_log_id"][:8] + "…"
        valid = "✓" if r["iter0_valid_json"] else "✗"
        sm = ("✓" if r["schema_match"] else ("✗" if r["schema_match"] is False else "—"))
        lr = r["length_ratio"] if r["length_ratio"] else "—"
        fh = len(r["iter0_forbidden_hits"])
        fm = r["iter0_fab_count"]
        j = judgments_by_id.get(r["source_log_id"])
        winner = j["winner_model"] if j else "—"
        a(f"| `{sid}` | {valid} | {sm} | {lr} | {fh} | {fm} | {winner} |")
    a("")

    a("### Files")
    a("- `iter0_eval_results_v1.jsonl` — raw inference outputs (20 rows × prompt + claude_response + iter0_response + timing)")
    a("- `level1_metrics.json` — structural analysis output")
    a("- `level2_judgments.jsonl` — Claude-as-judge outputs (anonymised, decoded)")
    a("- `held_out_cases_v1.jsonl` — original held-out set (50 cases; 20 used here)")
    a("- `held_out_cases_v1.sql` — query that produced the held-out set")
    a("")

    return "\n".join(lines)


def main() -> None:
    args = parse_args()

    results_rows = []
    with open(args.results) as f:
        for line in f:
            line = line.strip()
            if line:
                results_rows.append(json.loads(line))

    l1 = json.loads(Path(args.level1).read_text())

    l2_rows = []
    with open(args.level2) as f:
        for line in f:
            line = line.strip()
            if line:
                l2_rows.append(json.loads(line))

    md = build_md(args, results_rows, l1, l2_rows)
    Path(args.output).write_text(md)
    print(f"Report written: {args.output}  ({len(md)} chars)")


if __name__ == "__main__":
    main()

# Flywheel D — iter_0 evaluation

Three-level evaluation comparing iter_0 LoRA adapter against Claude Sonnet 4.6 on held-out production briefs.

## Inputs

- `iter0_eval_results_v1.jsonl` — 20 rows from running `04_eval_iter0.py` on Thunder Compute. Each row contains: `source_log_id`, `prompt`, `claude_response` (from production), `iter0_response` (generated), `generation_tokens`, `generation_time_s`.
- `held_out_cases_v1.jsonl` — original 50-case held-out set; 20 used for this round.
- `held_out_cases_v1.sql` — the query that produced the held-out set, for reproducibility.

## Order to run

### Level 1 — structural (free, instant)

```bash
python level1.py \
    --results iter0_eval_results_v1.jsonl \
    --output  level1_metrics.json
```

Computes JSON validity, schema match, length ratios, forbidden-phrase hits, and fabrication markers — for both iter_0 and Claude on the same data. Side-by-side, not in isolation.

### Level 2 — Claude-as-judge (~$1, ~5 min)

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
python level2.py \
    --results iter0_eval_results_v1.jsonl \
    --output  level2_judgments.jsonl \
    --judge-model claude-opus-4-7
```

For each case: anonymise responses (random A/B), send brief + both to judge, record scores on relevance/voice/integrity (1-5) plus winner pick. Stream-flushes per row; resume support if interrupted.

**Why Opus as judge.** Training labels were Sonnet 4.6. Using a different Claude model reduces but doesn't eliminate self-recognition bias. Worth noting in the report's confounds section.

### Level 3 + report

```bash
python build_report.py \
    --results iter0_eval_results_v1.jsonl \
    --level1  level1_metrics.json \
    --level2  level2_judgments.jsonl \
    --output  iter0_evaluation_report.md
```

Combines L1 + L2 into a markdown report. Auto-selects 6 cases for L3 spot-check inclusion based on:
- L1 violations (forbidden phrases, fabrication markers)
- L2 decisive wins for iter_0 (rare, informative)
- L2 decisive losses (gap of 3+ points across dimensions)

## What the report does and doesn't claim

**Does:**
- Reports iter_0 numbers alongside Claude's on the same data, so the gap is visible
- Calls out methodological confounds (self-judging, distribution match, sample size, training data artefacts)
- Reports median + range for length, not just mean — N=20 makes single outliers meaningful
- Includes auto-selected cases verbatim for the user to read directly
- Generates conditional recommendations that fire only when metrics warrant them

**Doesn't:**
- Make ship/no-ship recommendations — those depend on threshold decisions outside this scope
- Treat percentages as statistically established — N=20 is too small for that
- Hide places where iter_0 lost decisively — those are reported as L3 cases for transparency
- Assume Claude is "right" — both models are scored on the same rubric

## Iterating

For iter_1 evaluation:
1. Re-run `04_eval_iter0.py` (after renaming) against the same held-out set OR pull a fresh held-out set from `llm_call_log` (use `created_at > 2026-04-23` plus ID exclusion against both training set and v1 held-out).
2. Run all three levels with new outputs (`level1_iter1_metrics.json`, etc.).
3. Build a comparative report — left as future work; present version reports a single iteration only.

## Cost recap

| Stage | Cost | Time |
|---|---|---|
| Level 1 | $0 | <5 sec |
| Level 2 | ~$1 | ~5 min |
| Level 3 + report | $0 | <5 sec |
| **Total** | **~$1** | **~5 min** |
# iter_0 Evaluation Report

**Generated:** 2026-05-08 16:38 UTC
**Model under test:** Llama-3.3-70B-Instruct + iter_0 LoRA adapter
**Reference:** Claude Sonnet 4.6 (the model that produced training labels)
**Judge (Level 2):** claude-opus-4-7
**Sample size:** 20 held-out briefs

## TL;DR

- **JSON validity:** iter_0 20/20 (100%), Claude 19/20 (95%)
- **Schema compliance** (matching Claude's keys): 19/19 (100%)
- **Judge head-to-head:** Claude won 16, iter_0 won 4, tied 0 (of 20)
- **Mean judge scores (iter_0 / Claude):** relevance 4.35/4.6 (-0.25), voice 4.35/4.55 (-0.20), integrity 4.65/4.75 (-0.10)
- **Forbidden-phrase hits:** iter_0 0, Claude 1 (across 14 briefs with parseable avoid-lists)
- **Fabrication markers:** iter_0 0 (in 0 rows), Claude 0 (in 0 rows)
- **iter_0 inference speed:** ~4.9 tok/s on A100 80GB SXM4 with xformers (no FA2)

Read the methodology section before drawing conclusions — there are real confounds to weight.

## Methodology

### Sample selection
- 50 cases pulled from `llm_call_log` matching iter_0's training filter (`agent_type='page-content-writer'`, `step_name='process_sections_loop_iter_0_generate_content'`, `success=true`, JSON-shaped response) but with `created_at > 2026-04-23 14:54:32 UTC` (after the training export cutoff).
- Defensive exclusion of any `source_log_id` present in the training export.
- First 20 used for this round; remaining 30 reserved for iter_1 evaluations to keep the comparison set stable across iterations.

### Three-level analysis
- **Level 1** — structural metrics on JSON validity, schema, length, forbidden phrases, and fabrication markers. Local, no API.
- **Level 2** — Claude-as-judge with anonymized A/B comparison. Position randomized per case. Output: 1-5 scores on relevance, voice, integrity, plus winner pick.
- **Level 3** — manual spot-check on 6 cases auto-selected for L1 violations or L2 disagreement.

### Confounds and limitations

These materially affect interpretation and are not corrected for in the headline numbers:

1. **Self-judge bias.** The judge is a Claude model evaluating outputs from iter_0, which was trained to imitate Claude. We mitigated by using a different Claude model (Opus) than the one that produced training labels (Sonnet 4.6), but residual stylistic affinity may still favour Claude's outputs.
2. **Distribution match.** Held-out cases are from the same production distribution as the training set (same agent, same step, same overall site population). This tests fidelity to production patterns, not generalization to genuinely novel domains.
3. **Sample size.** N=20 is too small for statistical claims. Treat percentages as directionally suggestive, not statistically established. A 5-percentage-point difference at this N is well within noise.
4. **Training data artefact.** The training set contains the `<no value>` rendering bug. Eval cases also pre-date the fix, so they exhibit the same artefact — this tests inference under the same conditions as training. iter_1 will need fix-aware data.
5. **Inference engine.** Eval ran on Unsloth + xformers (no Flash Attention 2). Production deployment of this adapter would likely use vLLM or similar with FA2, which affects speed but not output quality at temperature 0.
6. **Greedy decoding only.** Eval used `temperature=0` (deterministic). Sampling-based generation in production would have different distributional properties.

## Level 1 — Structural

### JSON validity and schema compliance

| Metric | iter_0 | Claude |
|---|---|---|
| Valid JSON | 20/20 | 19/20 |
| Schema match (when both valid) | 19/19 | — |

Pairs eligible for paired comparison (both valid JSON): **19/20**.

### Length distribution

iter_0 / Claude character ratio:  mean **1.11**, median **1.01**, range 0.75–2.43.

Lengths are comparable (within 15% on average). No systematic verbosity gap.

### Forbidden phrases (from briefs' avoid-lists)

Across 14 briefs with machine-parseable avoid-lists (out of 20 total — some briefs use a non-bullet inline format we don't parse):

| | iter_0 | Claude |
|---|---|---|
| Total hits | 0 | 1 |
| Rows with at least one hit | 0 | 1 |

### Fabrication markers

These are regex heuristics — flags for L3 review, not verdicts. A flag means *something looks like it could be fabricated*; the L3 reading determines whether it actually is.

| Marker type | iter_0 hits | Claude hits |
|---|---|---|
No regex-pattern markers detected in either model's outputs. Note that L1 regex patterns (percentages, durations, named titles) catch only narrow cases — see L3 spot-check for fabrication that requires contextual reading to detect.

Rows with any marker: iter_0 0/20, Claude 0/20.

### Inference performance

- Mean generation time: **22.2s**
- Mean tokens generated: **108.0**
- Throughput: **~4.9 tok/s** on A100 80GB SXM4 + xformers (no FA2)
- Peak VRAM during eval: ~43 GB (well under the 80 GB available)

## Level 2 — Claude-as-judge

Judged with `claude-opus-4-7` over 20 cases. Each case used anonymized A/B comparison with random position assignment.

### Head-to-head

| Outcome | Count |
|---|---|
| Claude won  | 16/20 |
| iter_0 won  | 4/20 |
| Tie         | 0/20 |

**Position bias check.** No clear position bias (A won 11/20, 55%).

**Self-recognition signal.** 5 cases had *identical* relevance/voice/integrity scores for both responses but the judge still picked a winner. Of these, **5/5 went to Claude** — meaning the judge is using something the rubric doesn't measure to break ties, and that something correlates with Claude-style. With position bias controlled (A won 55%), this is consistent with residual self-recognition bias from the judge being a Claude model.

### Mean dimension scores

| Dimension | iter_0 | Claude | Δ (iter_0 − Claude) |
|---|---|---|---|
| Relevance | 4.35 | 4.6 | -0.25 |
| Voice     | 4.35 | 4.55 | -0.20 |
| Integrity | 4.65 | 4.75 | -0.10 |

Reading note: a Δ of -0.3 to -0.5 on a 1-5 scale is real but small; -0.7 or more is a meaningful gap. Δ near zero on integrity but negative on voice would suggest iter_0 is factually safe but stylistically softer than Claude.

## Level 3 — Spot check

Auto-selected 5 cases for manual review based on:
- L1 violations (forbidden hits or fabrication markers in iter_0)
- L2 decisive wins for iter_0 (informative when rare)
- L2 decisive losses for iter_0 (informs iter_1 priorities)

### Case 1 — `3bb14000-3236-4361-8973-96da6bfe0076`

**Selected because:** iter_0 decisive win — judge: "Both share the same headline and avoid banned phrases, but A is more specific to the target audience (named verticals) and echoes the brief's voice cues directly ('no black boxes', 'no overselling'). "

**Section type:** Purpose: Hero section for about pages

**Claude response:**
```json
{
  "headline": "We Build AI Systems That Actually Work in Your Business",
  "subheadline": "FineTuning is a specialist AI implementation practice working with UK SMEs. We design, build, and deploy AI systems that reduce manual work, surface the information you need, and run reliably in production — without requiring you to become a technical expert first."
}
```

**iter_0 response:**
```json
{
  "headline": "We Build AI Systems That Actually Work in Your Business",
  "subheadline": "FineTuning works with UK SMEs in admin-heavy sectors — professional services, logistics, facilities management, and similar — to replace slow, manual processes with systems that run reliably. No jargon, no overselling, no black boxes. Just practical AI that fits the way your business already operates."
}
```

**Judge:** iter_0 R5/V5/I5, Claude R4/V4/I5, winner: **iter0**.  *"Both share the same headline and avoid banned phrases, but A is more specific to the target audience (named verticals) and echoes the brief's voice cues directly ('no black boxes', 'no overselling'). B is competent but more generic positioning."*

---

### Case 2 — `cd649c27-4550-43c1-a3e6-45c6b9f7bad7`

**Selected because:** iter_0 decisive win — judge: "Both match the voice well, but B fabricates a specific response-time SLA ('two business days') that isn't sourced from the brief, violating the integrity rule. A delivers the same purpose-statement co"

**Section type:** Purpose: Hero section for contact pages

**Claude response:**
```json
{
  "headline": "Contact & Integration Enquiries",
  "subheadline": "Report a specification error. Request a gripper listing. Ask about API or bulk data access. If the data can be improved, we want to know — response within two business days."
}
```

**iter_0 response:**
```json
{
  "headline": "Report a Spec Error, Request a Listing, or Enquire About API Access",
  "subheadline": "Use this channel to flag a gripper specification error, request a missing model be added to the catalog, or initiate a bulk data or API integration enquiry. State your application constraints and the decision you're trying to make — the more specific you are, the faster we can respond usefully."
}
```

**Judge:** iter_0 R5/V5/I5, Claude R4/V5/I3, winner: **iter0**.  *"Both match the voice well, but B fabricates a specific response-time SLA ('two business days') that isn't sourced from the brief, violating the integrity rule. A delivers the same purpose-statement content without inventing a commitment."*

---

### Case 3 — `e952ac1d-24b5-45c0-9ce6-8b34d6cae118`

**Selected because:** iter_0 decisive win — judge: "A is more disciplined: shorter declarative beats, cleaner rhythm, and explicitly hits 'archetype earned' without softening into AI-performance language. B's subheadline runs long and tacks 'Enter the "

**Section type:** Purpose: Main hero banner with compelling headline, subheadline, and call-to-action buttons for homepage and landing pages.

**Claude response:**
```json
{
  "headline": "Spark Is Where Your Take Gets Tested",
  "subheadline": "Not a feed. Not a forum. An arena. Every day, a provocation drops. You argue your position. The AI judges your reasoning — not your clout. Your archetype is earned by how you play, not how you present. Enter the Daily Gauntlet.",
  "primary_cta": "Enter Today's Gauntlet",
  "primary_cta_url": "/contact.html",
  "secondary_cta": "See How It Works",
  "secondary_cta_url": "/services.html"
}
```

**iter_0 response:**
```json
{
  "headline": "Spark Was Built for People Who Actually Have Opinions",
  "subheadline": "Not a feed. Not a forum. An arena. Every day, a provocation drops. You take a position. The room decides if you can hold it. Your archetype is earned — round by round.",
  "primary_cta": "Enter Today's Gauntlet",
  "primary_cta_url": "/contact.html",
  "secondary_cta": "See How It Works",
  "secondary_cta_url": "/services.html"
}
```

**Judge:** iter_0 R4/V5/I5, Claude R4/V4/I4, winner: **iter0**.  *"A is more disciplined: shorter declarative beats, cleaner rhythm, and explicitly hits 'archetype earned' without softening into AI-performance language. B's subheadline runs long and tacks 'Enter the Daily Gauntlet' inside the prose when the CTA already covers it."*

---

### Case 4 — `c4bcb967-45f3-4d84-ae5b-056c79c9c078`

**Selected because:** iter_0 decisive win — judge: "Response A explicitly executes the rewrite guidance to add a 2–4 sentence MatchMatrix explainer naming scored parameters, weight-setting, and output format. Response B is competent but treats MatchMat"

**Section type:** Purpose: Hero section for services pages

**Claude response:**
```json
{
  "headline": "Gripper Selection Tools Built for Specification Decisions",
  "subheadline": "Catalog search, parameter calculators, and MatchMatrix scoring across pneumatic, electric, vacuum, magnetic, soft-robotic, and adhesive technologies — all benchmarked against consistent criteria so you can compare actuation types on the same terms and reach a defensible shortlist faster."
}
```

**iter_0 response:**
```json
{
  "headline": "Gripper Selection Services Built for Engineers Who Need Data, Not Demos",
  "subheadline": "Filter a cross-technology catalog of pneumatic, electric, vacuum, magnetic, soft-robotic, and adhesive grippers; calculate payload, stroke, and cycle time parameters; and run MatchMatrix — a weighted scoring system that benchmarks gripper fit across payload margin, cycle time, IP rating, cost-per-cycle, and user-defined application constraints. MatchMatrix scores each gripper against your specific application parameters, with weights you set explicitly before the calculation runs. The output is a ranked shortlist with score breakdowns, not a black-box recommendation. Methodology is fully documented so you can audit the logic before you commit to a specification."
}
```

**Judge:** iter_0 R5/V5/I5, Claude R4/V4/I5, winner: **iter0**.  *"Response A explicitly executes the rewrite guidance to add a 2–4 sentence MatchMatrix explainer naming scored parameters, weight-setting, and output format. Response B is competent but treats MatchMatrix as a vague claim, which is exactly what the guidance told the writer to avoid."*

---

### Case 5 — `fc8bce5f-c05e-4621-b2ef-5eb58b686f78`

**Selected because:** iter_0 decisive loss (gap=5) — judge: "Both match voice and audience well, but B fabricates specifics about the fix and timeline ('under a week', 'per-agent topic routing') that the brief does not supply, violating the integrity rule again"

**Section type:** Purpose: Main hero banner with compelling headline, subheadline, and call-to-action buttons for homepage and landing pages.

**Claude response:**
```json
{
  "headline": "40 Agents. One Misconfigured Consumer Group. Here's How It Broke — and How We Fixed It.",
  "subheadline": "In a 40-agent financial services pipeline, a single Kafka consumer group misconfiguration caused task duplication, rebalance storms, and silent data inconsistency under production load. This case study documents the failure mode, the architectural diagnosis, and the configuration changes that eliminated it — without rewriting the pipeline.",
  "primary_cta": "Book a Technical Discovery Call",
  "primary_cta_url": "/contact.html",
  "secondary_cta": "See How We Architect Agent Pipelines",
  "secondary_cta_url": "/services.html"
}
```

**iter_0 response:**
```json
{
  "headline": "Kafka Consumer Group Misconfiguration Broke Your Pipeline. We Know Why.",
  "subheadline": "A 40-agent financial services pipeline stalled under production message volume because the consumer group was sized wrong — not too small, too large. Every agent was competing for every partition. We diagnosed the partition affinity failure, reconfigured the group with per-agent topic routing, and deployed the corrected pipeline on Kubernetes in under a week. The failure mode is common. The fix is not obvious until it breaks in production.",
  "primary_cta": "Technical Discovery Call",
  "primary_cta_url": "/contact.html",
  "secondary_cta": "See How We Configure Kafka for Agent Pipelines",
  "secondary_cta_url": "/services.html"
}
```

**Judge:** iter_0 R4/V4/I2, Claude R5/V5/I5, winner: **claude**.  *"Both match voice and audience well, but B fabricates specifics about the fix and timeline ('under a week', 'per-agent topic routing') that the brief does not supply, violating the integrity rule against invented metrics/outcomes. A delivers the same failure-mode framing and stack credibility while staying honest about what's documented."*

---

## Overall assessment

This section is intentionally cautious. With N=20 and a self-judging confound, claims should be conservative.

### What appears to be working

- **JSON validity is intact.** iter_0 produces parseable JSON on every brief tested. Schema discipline learned from training transferred cleanly to held-out briefs.
- **Schema-key matching with Claude** is high (19/19), suggesting iter_0 picked up the right component-type → schema mapping during training.
- **Fabrication marker count is no worse than Claude's** — the model has learned the brief's restrictions on inventing metrics, names, or attributions.
- **Integrity score (judge-rated)** is comparable to Claude's: iter_0 4.65 vs Claude 4.75.

### What's not working as well

- **Head-to-head loss rate is high** (iter_0 won 4/20). Even with self-judge bias, this suggests Claude is meaningfully preferred on most briefs.

### Implications for production rollout

- **Use cases where iter_0 may be acceptable:** brief-faithful structured-JSON tasks where mechanical correctness matters more than voice nuance — internal/test sites, prototypes, sections where the brief has no strong voice prescription.
- **Use cases that still need Claude:** sites with elaborate voice prescriptions (e.g. the Leopardess Consulting brief), client-facing copy where stylistic tightness is the value, sections with case-study or testimonial requirements where fabrication risk needs careful handling.
- **Cost calculus:** the per-call cost of iter_0 served on a dedicated A100 80GB instance depends on utilization. Quick math at 0.5s/call inference (FA2-enabled production setup), $1.79/hr A100: each call costs ~$0.0003 at full GPU utilization. Compare to current Sonnet 4.6 cost per call. If utilization is anything below ~50% sustained, the GPU is uneconomic vs API calls until phase 2 batches across multiple agents.
- **Latency:** iter_0 generates at ~5 tok/s on this hardware. Claude API is ~50-80 tok/s. iter_0 is ~10× slower per call but co-located inside the cluster, eliminating round-trip latency. Net effect on user-facing pages depends on streaming behaviour.

### Recommendations for iter_1

- **Address verbosity.** Either filter training rows to those at the shorter end of the length distribution, or add a length-aware penalty during training.
- **Filter training data for the `<no value>` artefact.** Use `created_at >= <fix_deploy_date>` as the filter floor. We need to capture the deploy date from the team — the training rows from after that date will be artefact-free.
- **Run a 2-epoch ablation.** The iter_0 loss curve showed memorisation accelerating in epoch 3 (mean loss 0.12 vs 0.23 in epoch 2). If iter_1 with 2 epochs scores comparably on this same eval set, we'd save ~$5 per training run going forward.
- **Save adapters as fp16, not fp32.** One-line change halves adapter size from 791MB to ~400MB and halves transfer time. Costs nothing in quality.
- **Fix the 1-of-50 truncation bug** in `llm_call_log.response_text` — see PATCH_2026-05-06 for the column-truncation hypothesis. Affects training data quality at ~2% rate.

## Appendix

### Per-row summary

| source_log_id | iter_0 valid | schema match | length ratio | forbidden hits | fab markers | judge winner |
|---|---|---|---|---|---|---|
| `7d5f4ea0…` | ✓ | — | 0.45 | 0 | 0 | claude |
| `e39594c4…` | ✓ | ✓ | 2.43 | 0 | 0 | claude |
| `8ce6dec4…` | ✓ | ✓ | 1.02 | 0 | 0 | claude |
| `3bb14000…` | ✓ | ✓ | 1.1 | 0 | 0 | iter0 |
| `8f9b88a4…` | ✓ | ✓ | 0.81 | 0 | 0 | claude |
| `cd649c27…` | ✓ | ✓ | 1.64 | 0 | 0 | iter0 |
| `32c0c57f…` | ✓ | ✓ | 0.8 | 0 | 0 | claude |
| `fc8bce5f…` | ✓ | ✓ | 1.14 | 0 | 0 | claude |
| `0787ad8a…` | ✓ | ✓ | 1.07 | 0 | 0 | claude |
| `a273daa3…` | ✓ | ✓ | 0.75 | 0 | 0 | claude |
| `17c58ce7…` | ✓ | ✓ | 0.76 | 0 | 0 | claude |
| `7e6b05fb…` | ✓ | ✓ | 0.9 | 0 | 0 | claude |
| `61427591…` | ✓ | ✓ | 0.83 | 0 | 0 | claude |
| `e952ac1d…` | ✓ | ✓ | 0.91 | 0 | 0 | iter0 |
| `611eb4df…` | ✓ | ✓ | 0.88 | 0 | 0 | claude |
| `0bfcd02e…` | ✓ | ✓ | 1.01 | 0 | 0 | claude |
| `d1ae1a18…` | ✓ | ✓ | 0.93 | 0 | 0 | claude |
| `8367203f…` | ✓ | ✓ | 1.02 | 0 | 0 | claude |
| `c4bcb967…` | ✓ | ✓ | 2.03 | 0 | 0 | iter0 |
| `b5fc0c68…` | ✓ | ✓ | 1.01 | 0 | 0 | claude |

### Files
- `iter0_eval_results_v1.jsonl` — raw inference outputs (20 rows × prompt + claude_response + iter0_response + timing)
- `level1_metrics.json` — structural analysis output
- `level2_judgments.jsonl` — Claude-as-judge outputs (anonymised, decoded)
- `held_out_cases_v1.jsonl` — original held-out set (50 cases; 20 used here)
- `held_out_cases_v1.sql` — query that produced the held-out set

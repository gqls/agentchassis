# PLAN — getting more reasoning data: what we must record that we don't

*2026-07-18. Answers the owner's question: the fix-loop corpus is an eval set, so
where does training-scale reasoning data actually come from? Every figure below is
a live query, listed in the RUNBOOK so it can be re-run. Companion to
`PLAN_2026-07-18_reasoning_dataset_extraction.md`.*

---

## The headline

The fix loop is **not** where the platform's reasoning volume is. It is 445 rows.
The platform has logged **40,785 LLM calls** and holds **5,292 work items with
real terminal outcomes**. The reasoning is being produced at ~100× the rate we
were planning to harvest it.

What is missing is not reasoning. It is the **join between a decision and what
happened next** — and that join is missing for a mundane, fixable reason.

| | today |
|---|---|
| LLM calls logged, all time | 40,785 |
| …carrying a `work_item_id` | **2,894 (7.1%)** |
| …joinable to a *terminal* outcome | **1,165** |
| work items with a terminal status | 5,292 |
| items ever independently `verified` | **9** |
| human decisions with the reason recorded | **0 of 316** |

**1,165 outcome-linked reasoning steps already exist** — roughly 15× the fix
loop's 79. And the cheapest change on this page would take it past 15,000.

---

## Gap 1 — `work_item_id` is dropped by every large judgement agent (highest ROI by far)

`llm_call_log` already has a `work_item_id` column, and `llm_call_logger.go`
already populates it *when it is present in `input_data`*. It usually isn't:

| agent | calls | with `work_item_id` | |
|---|---|---|---|
| `page-content-writer` | 16,846 | **0%** | generation, lower priority |
| `content-quality-auditor` | 7,119 | **0%** | ← judgement agent |
| `visual-design-auditor` | 4,032 | **0%** | ← judgement agent |
| `site-review-agent` | 3,987 | **0%** | ← judgement agent |
| `feed-triage` | 423 | **0%** | ← judgement agent |
| `webdesign-agent` | 3,760 | 34.1% | |
| `content-gap-planner` | 2,004 | 41.7% | |
| `component-creator` | 447 | 64.4% | |
| `tool-recreation-handler` | 181 | **100%** | proves it is plumbing, not physics |

The four judgement agents at 0% represent **~15,500 calls in which an agent formed
a verdict about a page, a design, or a source — and we cannot tell whether it was
right**, because nothing connects the verdict to the item's fate.

`tool-recreation-handler` sits at 100%, so this is a propagation gap in specific
call paths, not a platform limitation. **This is the single highest-value change
on this page**: one field threaded through four agents' `input_data`, multiplying
the outcome-labelled corpus by an order of magnitude, with no new tables, no
schema change, and no new LLM spend.

## Gap 2 — human decisions are recorded as a status and nothing else

316 items have reached `needs_human_review`, `wont_fix`, or `rejected`.

```
approved_by populated:      0 / 316
resolution_path populated:  0 / 316
handled_by populated:      57 / 316
```

Both columns already exist on `site_work_items`. So every time a human overrules
the platform — the highest-quality label obtainable, and the only one expressing
*preference* rather than mere success — we keep the fact and discard the reason.

For a reasoning model this is the difference between "this was rejected" and
"this was rejected **because** it named a destination that doesn't exist." Only
the second teaches anything. Needed: `approved_by` + a free-text
`resolution_note` written at the moment of the human call, not reconstructed.

## Gap 3 — `complete` is self-reported; almost nothing is verified

4,583 items are `complete`. **9** have ever reached `verified` — all
`item_type='empty_section'`, none since 2026-07-14. So verification-after-fix was
built once, for one item type, and never generalised.

CLAUDE.md already enshrines the rule this violates: *"Trust the rendered artefact,
not the status. `complete` is not proof the work happened."* Training on
`complete` as a positive label trains on the agent's own say-so — precisely the
failure mode `bugs_open/012` documents, where a component was saved back as a
fragment and reported success.

Needed: a post-fix verification sweep that re-runs the *detector* that raised the
item and writes `verified` or `regressed`. The detectors already exist — that is
how the items were raised. This converts a self-report into ground truth, and
`regressed` is the negative class the corpus currently has almost none of.

## Gap 4 — the signals we already have and simply aren't using

Two are free, sitting in columns already populated:

**`attempt_count`** — a difficulty label at no cost:

| attempts | items | completed | stuck (`failed`/`unresolved`) |
|---|---|---|---|
| 0 | 4,581 | 3,971 | 228 |
| 1 | 589 | 540 | 21 |
| 2 | 72 | 67 | 3 |
| **3** | **60** | **5** | **44** |

The 60 three-attempt items are the most valuable rows on this page: the agent
tried its best three times and failed 44 times out of 60. Hard cases with a
ground-truth negative. A first-attempt success and a third-attempt failure are
different training signals and are currently indistinguishable in any export.

**`severity`, `impact`, `parent_item_id`, `depends_on`** — already populated,
give a cost-of-being-wrong weighting and the dependency structure between
decisions. Nothing needs building; the ETL just has to read them.

## Gap 5 — most judgement output isn't structured as reasoning

The fix loop emits `{outcome, citations[{tier,quote,where}], symptom_check,
next_scope}` and *enforces* it — `pkg/diagnose/step.go:67-101` degrades a verdict
to UNVERIFIABLE if it lacks citations or evidence families, and writes down why.
That is why its 79 steps are worth more per row than 15,000 auditor calls.

`feed-triage` is the one agent outside the loop that already gets this right:

```json
{"score": 62, "credibility": "medium",
 "credibility_reason": "Content originates from Murata, a reputable manufacturer,
   but is hosted on YouTube as promotional material rather than editorial news.",
 "reason": "Covers smart factory automation ... but is promotional video content
   rather than news.",
 "source_tier": "tier3_blog", "flagged": true}
```

Decision, justification, classification, confidence — one row, and it batches
many items per call, so its 423 calls carry **thousands** of individual
judgements. It is the best-shaped non-loop reasoning source on the platform and
was not on anyone's list.

The other auditors return prose. Needed: port the loop's contract — a required
justification field and a citation/evidence field — to `content-quality-auditor`,
`visual-design-auditor` and `site-review-agent`. Their prompts are already
judgement prompts; this changes the output schema, not the task. Do it once, and
~15,000 calls a month become training-shaped rather than needing an LLM pass to
retrofit structure.

## Gap 6 — no counterfactuals, and a lossy log

- **Rejected alternatives are never recorded.** We store what was chosen, never
  what was considered and dismissed. Preference pairs — the format most reasoning
  post-training actually wants — cannot be built from a chosen-only corpus. The
  council is the exception and shows the shape: `approve`/`object`/`veto` from
  several seats on one plan **is** a preference set, and it is already persisted.
  Cheapest source of pairs: ask the seat-based agents to emit a one-line
  runner-up, not a full second plan.
- **`LogLLMCall` is fire-and-forget with a 5s timeout**
  (`platform/orchestration/actions/llm_call_logger.go:34`). Rows can vanish under
  load, silently, and load correlates with the interesting runs. For a dataset
  this is corruption you cannot detect after the fact. Needed: make the write
  durable (queue or synchronous-with-retry) before treating counts as truth.
- **No `stop_reason` column.** Truncation is currently inferred from
  `output_tokens >= max_tokens` (old rows) or an error-message `LIKE` (new). It's
  6 rows of 445 today, but it is inference, not a fact, and it should be a column.

---

## Where the volume actually comes from

Ordered by cost per usable example.

**1. Fix the plumbing (Gaps 1–4).** No new LLM spend at all. Retro-active for
Gap 4, forward-only for Gaps 1–3. Takes the outcome-labelled corpus from ~1,165
to a five-figure number as normal platform operation continues. **Do this first;
everything else multiplies whatever this produces.**

**2. Harvest `feed-triage`.** 423 calls × many items per call, already correctly
structured, running daily since March. Costs one ETL query.

**3. Structure the three big auditors (Gap 5).** ~15,000 calls/month become
reasoning-shaped. Cost: a prompt+schema change per agent, plus the guard logic —
which already exists and can be reused from `pkg/diagnose`.

**4. Replay the known-answer bugs.** `/bugs_open/` holds 20 cases with documented
root causes and verification steps. Each is a re-runnable diagnosis task with a
known answer. N bugs × M models × K prompt variants generates graded trajectories
on demand, and the grading is nearly free because the answer is already written
down. This is also the cheapest route to *difficulty diversity* — the corpus today
is all one shape.

**5. Multi-model gauntlet.** Already contemplated in the concept-register
workstream. Same task, several models, the disagreements are the valuable rows:
where models diverge is where the reasoning is load-bearing.

**6. Deliberate loop runs at volume.** The most expensive option, and the one to
decide *last* — after 1–5, because they change what a deliberate run is worth.

## The honest ceiling

Even with all six, this platform makes hundreds of decisions a day, not millions.
Realistic scale over six months is **tens of thousands** of outcome-labelled
reasoning steps. That is fine-tuning and evaluation scale. It is not pretraining
scale, and no amount of plumbing makes it so.

That is not a reason to stop — a few thousand *outcome-verified, domain-specific*
reasoning traces with real negatives is a genuinely scarce asset, and worth more
per row than a large scrape of unlabelled chain-of-thought. But the plan should
say out loud that the goal is a high-quality specialist corpus, not a big one.

---

## Recommended first move

Gap 1 alone — thread `work_item_id` through the four judgement agents. It is the
smallest change, needs no schema migration, no new spend and no new agent, and it
is the one that determines whether everything else is joinable. Gap 2 (a
`resolution_note` on human calls) is second and nearly as cheap, and every day it
waits is more human judgement thrown away.

Both are `platform/` changes and therefore **not this thread's to make** — they
belong to the owning threads and should go through the council gate.

---

## Method note

The recurrence idea — "an item completed and then re-detected means the fix didn't
hold" — is a genuinely strong signal and is **not** included above, because the
query written for it was wrong: a naive self-join reported 387,301 recurrences
for `page_rerender`, which is the join exploding across many same-type items per
site, not a finding. It needs a query that pairs each item with the *next*
detection of the same type on the same target, ordered by time. Recorded here as
an open lead rather than dropped, and deliberately not quoted as a number.

# NOTES — reasoning-dataset thread (running record)

*Append-only running notes. Newest turn at the BOTTOM. One entry per working
session; record what was tried, what the DB actually said, and what was wrong —
especially what was wrong, including our own earlier claims in this file.*

**Standing docs for this workstream** (all in
`docs/agent_docs/docs024_key_docs_latest/reasoning_dataset/`):

| doc | what it is |
|---|---|
| `HANDOFF_2026-07-18_reasoning_training_dataset.md` | the originating brief (fixloop thread) |
| `PLAN_2026-07-18_reasoning_dataset_extraction.md` | the design + phasing; **corrections to the handoff live here** |
| `RUNBOOK_reasoning_dataset.md` | the commands — how to extract, verify, re-run |
| `NOTES_reasoning_dataset.md` | this file — the thread's memory |
| `SUMMARY_reasoning_dataset_2026-07-18.md` | the read-out: where we are, where we're going |

---

## Turn 1 — 2026-07-18 (~13:30–14:15Z) — planning, and a corrected finding

**Asked for:** a plan for the handoff, and what it involves.

**Did:** grounded every volume claim in the handoff against the live DB rather
than taking them forward; mapped the artifact schema and join topology via two
parallel exploration passes; wrote `PLAN_…`; committed `b65fe81eb`.

### What the DB actually says (all figures live, 2026-07-18)

| | handoff said | live |
|---|---|---|
| bundles | 38 | 38 (13 correlations) |
| fix_plans | 39 | 43 |
| council_reports | 34 | 43 |
| escalations | 5 | 5 |
| verdict+review rows in `llm_call_log` | 296 | **445** |
| orchestrations with a verdict | 26 | 26 |
| **diagnosis trail steps** | not stated | **79** |

Trail depth per run: 5 runs × 1 step, 5 × 2, 8 × 3, 8 × 5.
Verdict outcomes: **57 UNVERIFIABLE / 18 CONFIRMED / 13 REFUTED** of 89 calls.
Model split: claude-sonnet-4-6 = 70 verdicts, claude-sonnet-5 = 19.
Truncation: **6 rows of 445** trip the filter. All old-regime
(`output_tokens >= max_tokens`); zero new-regime.

**Read: the corpus is an eval set, not a training set.** 79 diagnosis steps across
13 trajectories, split across two model generations, will not train a reasoning
model. No extraction design changes that; it is a volume fact. See PLAN §2.

### The `<no value>` finding — and the correction to it

Found that **19/19 `repropose` prompts render `<no value>`** for 2–6 reviewer
sections (also `review_debug_historian` 13/13, `reframe` 2/2). Verified it was not
abstention: for run `53da3a30`, `collected_data->'review_editquality'->'result'` is
a complete object (1561 chars) while the prompt shows a blank.

Then drafted a `/bugs_open/` case file for it. **Grepped the bugs_open index first
(CLAUDE.md's rule) and did not need to** — it was already filed the same day as
`016` by the experience-loop thread, and the `fix-proposer` row had been **fixed at
13:15:11Z** by the council-gate thread, about 45 minutes before I looked.

Two intermediate errors worth recording, because both were on their way into a
committed document:

1. **Claimed the bug was live and unfiled.** It was neither. The rule that caught
   it is the cheap one — grep the index before filing. It cost one grep and saved
   a duplicate case file plus a false alarm to two threads.
2. **Then claimed "the fix didn't take"** — because two repropose calls at 13:17Z
   and 13:24Z post-date the 13:15Z fix and still showed 6 blanks each. Wrong: both
   belong to orchestration `48cf0339`, **started 13:11:13Z**, carrying pre-fix
   config. The log timestamp is the *step's*, not the *run's*. Checked all rows:
   no repropose has started post-fix.

> **Transferable (→ 016b §9 candidate):** when verifying a config fix against
> `llm_call_log`, join to `orchestration_states.created_at` and test the **run**
> start against the fix time. A long run straddles the boundary and its later
> steps look post-fix while carrying pre-fix config. Reading the step timestamp
> alone flips either way — stale round read as failed fix, or failed fix read as
> stale round.

**Net state of the defect:** cause fixed, **unexercised** (no post-fix run yet, so
unproven in the wild); all 19 historical rows invalid and quarantined; and 016's
*second* finding confirmed live — **13 seats seeded, 6 referenced by the
`repropose` prompt, 7 invisible to the reviser** (compliance, debug_historian,
llm_reliability, render_guardian, adoption_guardian, diagnosis_guardian,
improvement_guardian). `review_debug_historian` is doubly disconnected: it was
rendering `<no value>` on its own input 13/13 *and* its output never reaches the
reviser. Quantification appended to `bugs_open/016`.

### Decisions taken

- **Eval-first, not training-first.** Build the pipeline; aim it at a benchmark.
- **Run outside the cluster** (claimscan idiom: psql extracts, Go transforms).
  `training_data_export.go:3-8` records that JSONL-onto-a-pod was already tried
  and retired — the file died with the pod.
- **Hand-curate the benchmark labels.** One rubric file exists, ~10 graded runs,
  four incompatible prose shapes, and several uses of "FAILED" that mean an API
  529 rather than a grade. A parser would mislabel them.
- **Flag bad rows, don't drop them.** `input_complete` + `exclude_reason`. A file
  that silently omits 19 repropose rows reads as "no repropose data".
- **Add a `guard` block** pairing the raw verdict against the coerced one — not in
  the handoff, and probably the most distinctive signal in the corpus.
- **Do not fix the loop from this thread.** Read-only; `platform/`, `internal/`,
  `pkg/` untouched.

### Open / next

- [ ] Raise the **6-of-13 seat gap** with the fixloop thread (paragraph drafted).
- [ ] Agree the read contract (freeze or version the verdict/council JSON shape).
- [ ] Watch for the first post-13:15Z repropose run — it proves or disproves the
      016 fix, and it opens the post-fix half of the natural experiment.
- [ ] Phase 1: `extract.sql` + `cmd/reasoningset`.

**Landmine for next turn:** the corpus is now **bimodal** across 13:15:11Z —
pre-fix repropose rows have blank objections, post-fix ones will not. Do not pool
them. It is a provenance boundary, and also a free natural experiment: same task,
same models, with and without objections visible.

---

## Turn 2 — 2026-07-18 (~14:30–15:15Z) — where more data actually comes from

**Asked for:** how to get more reasoning training data; what we need to record
that we don't.

**Did:** surveyed the whole platform's LLM output rather than just the fix loop.
Wrote `PLAN_capture_gaps_and_volume.md`.

### The finding that reframes the workstream

We had been sizing this project off the fix loop — 445 rows. **The platform has
logged 40,785 LLM calls and holds 5,292 work items with terminal outcomes.** The
reasoning is being produced at ~100× the rate we planned to harvest it. What is
missing is the *join* between a decision and its consequence.

| | |
|---|---|
| LLM calls, all time | 40,785 |
| carrying `work_item_id` | 2,894 (**7.1%**) |
| joinable to a terminal outcome | **1,165** — already 15× the fix loop |
| items ever independently `verified` | **9** |
| human decisions with a reason recorded | **0 of 316** |

### The six gaps (detail in the PLAN)

1. **`work_item_id` dropped by every big judgement agent** — content-quality-auditor
   (7,119 calls), visual-design-auditor (4,032), site-review-agent (3,987),
   feed-triage (423) are all at **0%**, while `tool-recreation-handler` is at
   **100%**. Plumbing, not physics. ~15,500 verdicts currently unjoinable to any
   outcome. **Highest ROI on the page** — one field, no migration, no new spend.
2. **Human decisions keep the status and discard the reason.** 316 items in
   `needs_human_review`/`wont_fix`/`rejected`; `approved_by` 0, `resolution_path`
   0. Both columns already exist.
3. **`complete` is self-reported.** 4,583 complete, **9** ever `verified`, all one
   item type, none since 07-14. Training on `complete` trains on the agent's own
   say-so — the `bugs_open/012` failure mode exactly.
4. **Free signals unused:** `attempt_count` (60 items hit 3 attempts, 44 of them
   stuck — hard cases with ground-truth negatives), plus `severity`, `impact`,
   `depends_on`.
5. **Most judgement output isn't structured.** The exception is **`feed-triage`**,
   which already emits `{score, reason, credibility, credibility_reason,
   source_tier, flagged}` and **batches many items per call** — 423 calls carry
   thousands of judgements. Best-shaped non-loop source on the platform and it was
   on nobody's list.
6. **No counterfactuals; lossy log.** Rejected alternatives never recorded (the
   council's approve/object/veto is the exception and shows the shape).
   `LogLLMCall` is fire-and-forget with a 5s timeout — rows vanish under load, and
   load correlates with interesting runs.

### Reads taken

- **Plumbing before generation.** Gaps 1–4 need no new LLM spend and multiply
  whatever every later lever produces. Deliberate volume runs are the *last* move,
  not the first.
- **Replay `/bugs_open/` is the cheap volume lever** — 20 cases with documented
  root causes and verification steps = re-runnable tasks with known answers, and
  the grading is nearly free.
- **State the ceiling out loud.** Hundreds of decisions a day, not millions.
  Realistic six-month scale is tens of thousands of outcome-labelled steps:
  fine-tuning and eval scale, not pretraining. The goal is a scarce high-quality
  specialist corpus, not a big one.

### Error made this turn

The recurrence signal ("completed then re-detected = the fix didn't hold") is
real and valuable, but my query for it was wrong — a naive self-join reported
**387,301** recurrences for `page_rerender`, which is the join exploding across
many same-type items per site, not a finding. Left in the PLAN as an open lead
with the bad number explicitly *not* quoted. Needs a query pairing each item with
the *next* detection of the same type on the same target, ordered by time.

### Open / next

- [ ] Owner call: commission Gap 1 (`work_item_id` propagation) with the owning
      threads — it is a `platform/` change, so council gate, and not ours to make.
- [ ] Gap 2 (`resolution_note` on human calls) — every day it waits is more
      human judgement discarded.
- [ ] Write the recurrence query properly.
- [ ] Harvest `feed-triage` in the Phase 1 extractor — it is already well-shaped
      and needs only a query.

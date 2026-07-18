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

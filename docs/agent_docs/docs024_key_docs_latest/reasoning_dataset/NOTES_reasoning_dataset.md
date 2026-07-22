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

---

## Turn 3 — 2026-07-18 (~15:15–16:00Z) — drafting the submissions killed two of my own gaps

**Asked for:** draft council-gate submissions for Gaps 1 and 2.

**Did:** read the actual code paths before writing the plans. Neither gap
survived. Drafted two submissions for what is really there instead.

### Gap 1 was wrong — a statistic read as a mechanism

The 0% `work_item_id` figures are real; the conclusion was not. Those agents are
item **producers**, not handlers — they *raise* work items and are never
dispatched to work on one, so no `work_item_id` exists at spawn time and there is
nothing to propagate. `tool-recreation-handler` hits 100% purely because
`build-dispatch-loop` injects `"work_item_id": "pending.first_item.id"`
(`051_build_dispatch_loop.sql:78`) — the handler path, which none of the four is
on. And **`feed-triage` never touches `site_work_items` at all**
(`feed_triage_actions.go:241` updates `content_feed_items`), so it would sit at
0% for ever. Including it was a measurement artefact.

> **Transferable:** a 0% column is evidence of *absence*, not of *dropping*.
> I inferred a plumbing failure from a statistic without checking whether the
> value was ever in scope. The join we want runs the **other way** — from the
> created item back to the run that raised it.

### Gap 2's premise collapsed — and the real finding is worse

Drafted "write `approved_by`/`resolution_path` in the admin handlers", then
checked whether the reason was being captured elsewhere. The handlers do write
`result = jsonb_build_object('resolution', $2, 'resolved_by', 'admin')`. The
JSONB is empty too:

```
complete items: 4,599   with result->'resolution': 0   with result->'approved_by': 0
```

**Zero of 4,599.** The human-resolution path has never been called. Those routes
are live and uninvoked; `HandleConfirmWorkItem` is fully implemented and **never
registered in `server.go`** — unreachable. You cannot improve the capture of a
path nobody takes. 275 items sit in `needs_human_review` as a dead-letter queue;
whether that queue should be worked is a product question for the owner.

Consolation: `error` **is** populated where it matters (77/96 `failed`, 42/50
`wont_fix`). The failure reason is already recorded and just not exported — ETL,
not a platform change.

### What was submitted instead

- **A — origin provenance** (`submission_A_work_item_origin_provenance.json`):
  `origin_correlation_id` on `site_work_items`, populated at the two INSERT
  paths. The corrected Gap 1, in the direction that exists.
- **B — register more verifiers** (`submission_B_register_more_item_verifiers.json`):
  promoted from Gap 3. The completion-gate framework, the policy and a reference
  implementation all exist (`verifiers.go`, `complete_work_item_verification.go`,
  `VerifyEmptySectionResolved`) — but `RegisterVerifier` has been called
  **once**, for `empty_section`, which is exactly why 4,594 of 4,599 `complete`
  items were never checked and only 9 items ever reached `verified`.

Both validated against every 097 client-side check (scope, ≤8 edits, non-empty
`grounded_in`, size). **Neither submitted** — that spends credits and is the
owner's call.

### Open / next

- [ ] Owner: submit A and/or B via `097_TRIGGER_council_review_v1.sh`.
- [ ] Owner/product: is the `needs_human_review` queue (275 items) meant to be
      worked? If yes it is a UI/process gap, not a data one.
- [ ] Correct the `feed-triage` line in Turn 2 above — it remains a good
      *reasoning-shape* source (structured, batched judgements) but is NOT a
      work-item outcome source and never will be.

---

## Turn 4 — 2026-07-19 (~13:10Z) — both submissions fired

**Asked for:** update the docs with where we are including the missteps, then
fire both submissions.

### Pre-flight re-verification (do this every time; it caught drift)

A day had passed, so every `grounded_in` claim was re-checked against live before
spending credits. All three load-bearing claims held:

- `RegisterVerifier(` still appears **once** (`check_empty_sections.go:38`)
- `origin_correlation_id` still does not exist anywhere in `platform/ internal/ pkg/`
- nothing had landed on either target file (last relevant commit `3b52da8ec`,
  a `generic_theme` fallback fix, untouching)

Two things *had* drifted and were corrected in the JSONs before firing:

- work-item counts: `complete` 4,599 → **4,570**, `failed` 96 → **152**,
  `unresolved` 200 → **236**. No claim changed meaning, but an unexplained
  mismatch between a submission and the live DB reads as carelessness to a
  reviewer who checks.
- **the gate roster grew 9 → 13 seats overnight.** Exactly the churn CLAUDE.md
  warns about. Re-read it rather than assuming which seats fire.

### Fired

| | submission | SUBMISSION_CORR | RUN_ORCH_ID |
|---|---|---|---|
| **A** | work-item origin provenance | `61105914-fe50-4e23-b36f-70654ed25727` | `e19504c7-b812-4072-ba41-9033c5d878c0` |
| **B** | register more item verifiers | `66dbd0dd-de5f-4f50-acd3-f5f3d817dbd9` | `0bef2b48-56f7-4aae-8b42-05c35d58cdf7` |

Two runs, not one: they are separate coherent tasks, which is the credit policy's
unit. Plan sizes 7,498 and 8,997 bytes, both well inside the 64KB cap.

### The missteps this workstream has made, collected in one place

Four now, and the pattern in them is worth more than any single correction:

1. **Claimed a bug was live and unfiled** (Turn 1). It was filed the same day
   *and already fixed*. Caught by grepping `/bugs_open/` before filing — the
   cheapest possible check, and the only reason a duplicate didn't go out.
2. **Claimed the fix hadn't taken** (Turn 1). Two calls post-dated the fix and
   still showed the defect — but belonged to a run that *started* before it. The
   log timestamp is the step's, not the run's.
3. **Read a 0% column as a dropped field** (Turn 3). It was an absent one: those
   agents are producers, never dispatched with a work item. A statistic became a
   wrong recommendation because I didn't check whether the value was ever in
   scope.
4. **Proposed writes to a code path nobody calls** (Turn 3). 0 of 4,570 complete
   items carry a resolution in *either* the column or the JSONB, because the
   human-resolve handlers have never been invoked and `HandleConfirmWorkItem`
   isn't even registered.

> **The common thread:** every one came from reasoning about the system from its
> *data* without reading its *code*, or from reading a doc without checking the
> live row. Both are fast and both feel like evidence. The corrective is
> mechanical, not clever — before asserting a mechanism, read the code that
> implements it; before asserting a state, query the thing itself. All four were
> caught before anything shipped, three of them by checks this repo's own
> CLAUDE.md already mandates.

### Open / next

- [ ] Read both verdicts; APPROVED → hand the submission to the owning thread
      with the trailer id. REVISE → the objections come back with the reviewers'
      own read-only checks already answered.
- [ ] Neither change is ours to implement — `platform/` belongs to the owning
      threads. This thread's role ends at an approved plan.
- [ ] Phase 1 of the extractor is still unstarted and is not blocked by either.

---

## Turn 5 — 2026-07-19 (~13:10–15:00Z) — six council runs, and a self-inflicted one

### Run ledger

| submission | round | RUN_ORCH_ID | verdict | objectors |
|---|---|---|---|---|
| A | 1 | `e19504c7` | revise | editquality ×3, tooling_provenance ×2 |
| A | 2 | `8c085080` | revise | debug_historian ×2 (both procedural) |
| A | 3 | `b52d9694` | revise | **reuse_agent ×2** (new seat, first fire) |
| B | 1 | `0bef2b48` | revise | editquality ×2, bug_historian ×3, guardian ×3 |
| B | 2 | `1d534983` | revise | editquality ×2, bug_historian ×2, guardian ×3 |
| B | 2-dup | `4e7a1d1e` | revise | **wasted run — see misstep below** |

SUBMISSION_CORRs: A `61105914-fe50-4e23-b36f-70654ed25727`,
B `66dbd0dd-de5f-4f50-acd3-f5f3d817dbd9`.

### MISSTEP — "the spawn was dropped". It never was, three times over.

After firing, `orchestration_states` showed no row for the run. I polled for ~7
minutes, concluded the spawn had been silently dropped, cited CLAUDE.md's ~300s
post-restart rule, checked the pod (15h old, so the rule didn't apply), declared
it lost anyway — and **re-fired B, spending a council run on a duplicate**.

Every one of those runs landed. `1d534983` completed at 13:38, well after I had
written it off. The orchestration row is created when the coordinator picks the
message up, which under normal platform load lags the Kafka publish by **several
minutes** — and the platform was busy (constant orchestrations at 13:52).

> **Transferable — the polling window is not the timeout.** Absence of an
> `orchestration_states` row is not evidence of a dropped spawn until you have
> waited longer than the queue depth, which is not a number you can see. Poll for
> at least 10–15 minutes before concluding loss, and check whether *other*
> orchestrations are being created in the meantime — a busy platform is the
> explanation, not a counter-argument.

This is the same failure as the earlier three: concluding a **mechanism** from an
**absence**, confidently, without the check that would have settled it. Cost this
time was real money, not just a wrong sentence.

> **Checked whether `bugs_open/029` excuses this. It does not — and the near-miss
> is instructive.** 029 (filed the same day, relojistas thread) documents twelve
> orchestrations hanging in `AWAITING_RESPONSES` and saturating the `dispatch`
> concurrency group **between 12:52 and 13:28** — almost exactly this session's
> submission window. It would have been very easy, and wrong, to file my latency
> under it. The bug's own observation section rules it out: *"Meanwhile
> `council-gate`, `endpoint-health-checker` and other groups continued normally"*
> — council-gate is a different concurrency group and was unaffected throughout.
> My runs were ordinary queueing, and my diagnosis was ordinary impatience.
> A plausible, same-day, same-window bug report is exactly the kind of evidence
> that makes a wrong conclusion feel confirmed; the thing that settled it was
> reading the bug rather than pattern-matching its title.

### Where each submission actually stands

**A — converging.** 5 objections from 2 reviewers → 2 from 1 (both procedural:
pod-verification step, migration verify/rollback files — both added) → round 3
drew a *different* seat, `reuse_agent`, firing for the first time because the
plan now touches a migration. Its two points:
1. `batch_id` already exists on this INSERT path — show it isn't already the
   per-run key before adding a parallel column. **Checked: it is not.**
   `batch_id := uuid.New()` (`write_audit_findings_action.go:600`) is a fresh
   random uuid per invocation, mapped to no correlation or orchestration id
   anywhere. It groups an audit run's items but cannot identify *which* run. The
   objection is refuted — but the reviewer was right that the plan asserted this
   by omission rather than showing it.
2. Two insert paths writing the same table, one now provenance-aware and one not,
   is a fork with no unification plan — architecture-level, and belongs in the
   plan body, not a doc_notes footnote. Fair; unaddressed.

**B — not converging, and the reason is legitimate.** Objection count held at 8
across both rounds. The hits that matter:
- **`VerifyPhantomInternalLinkResolved` is "a stub dressed as an edit"** —
  comments only, no query, no return, while edits 1 and 2 carry real compilable
  logic. Completely fair. I wrote prose where the plan needed code.
- **guardian caught an internal contradiction in my own plan**: it cites 9 items
  at `status='verified'` while asserting "no Go code sets that today". Both are
  true (the 9 are historical/hand-set; the grep genuinely finds no Go writer) but
  the plan presents them side by side without reconciling, which reads as one of
  them being wrong.
- **`VerifierCoverage()` only sees item_types with a registered discovery check** —
  any item_type created by another path is invisible to the coverage guard, so
  the guard under-reports the very gap it exists to expose.
- The ~47-entry allowlist is sketched as `"... the remaining current gaps"`, so
  the test cannot compile as written.

### Decision: stop firing

Six runs is enough. A is two small edits from plausible approval. B's remaining
objections require *writing the actual predicate extraction and enumerating 47
item types* — that is implementation work in `platform/`, which this thread is
explicitly not allowed to do (see PLAN §6). Iterating a plan toward the level of
detail the council wants would mean doing the owning thread's job in a JSON file.

> **What six runs taught about the gate that is worth keeping:** it is most
> valuable on the *first* round, where it caught two real design defects (the
> silent-null UUID parse; the absent-row blind spot in live code). By round three
> it is mostly enforcing plan-completeness, and a plan detailed enough to satisfy
> it is nearly the change itself. The gate reviews plans; past a point the honest
> move is to hand over the plan and let the implementing thread take the next
> verdict on real code.

### Open / next

- [ ] **Owner call:** hand A and B to the owning threads as-is (recommended), or
      spend a 7th run on A after adding the two `reuse_agent` answers.
- [ ] If A goes another round: add the `batch_id` refutation above verbatim, and
      promote the two-insert-path fork from doc_notes into the plan body.
- [ ] Phase 1 extractor — still unstarted, still unblocked by any of this.

---

## Turn 6 — 2026-07-20 — bug 033 filed, A assigned, Phase 1 BUILT

### 033 — the human-review queue, and a correction to our own claim

Filed `bugs_open/033`. 292 items in `needs_human_review`, oldest 2026-03-15,
arriving faster than ever (216 in July, 47 of them `cta_names_unknown_destination`)
and never drained. The three admin routes that could action one have **never
run**; a fourth (`HandleConfirmWorkItem`) is fully implemented and **never
registered** in `server.go`. `approved_by` and `resolution_path` are dead columns.

**MISSTEP corrected while filing.** We had written "0 of 4,570 have a
resolution". That query was scoped to `status='complete'`. Unscoped it returns
**8** — good prose, written by working threads via direct SQL (all with empty
`resolved_by`, which the API would have stamped `'admin'`), seven `cancelled`
and one `section_source_drift`. So reasons ARE captured: rarely, ad hoc, without
identity, never yet for a `needs_human_review` item. Corrected in
`PLAN_capture_gaps_and_volume.md` too.

033 deliberately proposes **no code**. It asks the prior question — queue or bin?
Wiring the surface is wasted if it is a bin; re-tuning the producers is wrong if
it is a queue. It also records that "record who decided" is blocked on auth:
those handlers have no user context at all.

### A assigned

To `work_item_completion_integrity` — it adds a provenance column to
`site_work_items` and their remit is whether such a row can be trusted to mean
what it says; they already hold `017`, `032` and the verifier-coverage handoff.
Stated weakness (this is the *creation* end, their name says *completion*) with
an explicit written route to decline.

### Phase 1 EXTRACTOR — built, run, verified

`cmd/reasoningset` + `extract.sql`. **820 records / 112 trajectories.**
689 `input_complete`; 131 flagged (69 `no_value_injection`, 43 `blinded_docs`,
16 `truncated`, 3 `call_failed`); 7 guard trips; 14 guard-unaligned; 10 graded.

**Verification found two real defects in my own first cut.** Both fixed; both
would have shipped silently.

1. **Guard misalignment.** `llm_call_log` has no iteration column, so I paired
   verdict calls to trail entries **by index**. Unsound — runs exist with 5
   verdict calls against a trail of length 1 (retries and failures leave no
   trail entry). It emitted a raw `UNVERIFIABLE` against a coerced `CONFIRMED`,
   which the coercion logic **cannot produce** (it only ever degrades). That
   impossibility is what exposed it. Now the guard asserts only when the counts
   agree, and records `alignment` otherwise.

   > **Transferable:** when pairing two sequences with no shared key, assert the
   > invariant the pairing must satisfy and let it fail loudly. Here the
   > invariant was free — "coercion never upgrades" — and it caught an error that
   > eyeballing the output would not have.

2. **Blinding leak.** 47 rows carried fixloop doc text. Not from the diagnosis
   corpus — from **council-gate submissions legitimately proposing changes to
   those scripts**. Flagged `blinded_docs` rather than dropped: an eval consumer
   must drop them, a training-only consumer may keep them, and that is the
   consumer's call to make, not the extractor's.

**Passed:** `input_state` byte-matches the DB bundle (30,269 bytes both sides);
zero impossible inversions; `tripped` set only when aligned; every excluded row
carries a reason; zero unflagged blinded rows.

### Two traps, both now in the RUNBOOK

- **`psql -At -f -` under `kubectl exec -i` silently truncates after the first
  statement.** First run: 781 `step` rows, zero `trail`/`bundle`, exit 0, and
  only "Waiting for server to close stdin failed" on stderr. psql reads stdin by
  default — drop the `-f -`. Silent partial success, the family this whole
  workstream keeps meeting.
- **Reconcile against a snapshot, not a live count.** JSONL held 676 review rows
  against a live 694. The 18 "missing" were council runs — *including my own* —
  that landed mid-extract. Bounding the count by the extract's newest row gives
  676 exactly.

  > **MISSTEP (#7).** I first blamed SQL `LIKE 'review_%'` treating `_` as a
  > wildcard. Plausible, tidy, and false: `WHERE step_name LIKE 'review_%' AND
  > step_name !~ '^review_'` returns zero rows. I checked before acting, which is
  > the only reason it cost nothing. The pattern is now familiar enough to name:
  > **a plausible mechanism is not a diagnosis.**

### Open / next

- [ ] Complete `LABELS_benchmark.json` — 3 of ~10 entries curated. Needs a
      careful read of `NOTES_running_fixloop(10).md`; deliberately not parsed.
- [ ] Phase 3 quality report — surviving steps per model slice is the go/no-go.
      Early read: 231 sonnet-4-6 / 589 sonnet-5, so the slices are usable for
      eval and still far short of training scale.
- [ ] Harvest `feed-triage` (423 calls × many judgements each, already
      structured) — needs a second extract query, no new machinery.
- [ ] Owner call on 033 (queue or bin) and on whether to generate volume.

---

## 2026-07-22 — new chassis v1.0.1149 adds a 4th provenance boundary (council lane)

Owner flagged a fresh production image. Checked whether it touches this thread's
**read contract** (the verdict / council_report / score_relevance JSON shapes),
since the extractor runs outside the cluster and needs nothing rebuilt of its own.

**Finding: the council DECISION RULE changed, live.** Commits `9e91999a4`
(severity adjudicator) + `872c830a8` (`severityGates`) ship in v1.0.1149. Before
them, ANY seat objection at any severity gated a round to `revise`; now only a
**high**-severity objection gates (low/medium are advisory). Pod-verified against
the running binary (not the tag): `severityGates` = 2 hits, `objectionGates` = 2,
`council_decide` positive control = 16. Pod `agent-chassis-7d4ff8b54-cm786`
started **2026-07-22T13:56:14Z**.

**Why it matters here:** this overturns the exact fact my dissent metric was
*corrected against* — the addendum says "the council returns `revise` if any seat
objects". That is now false for post-boundary rows, so `round_decision` carries
two different meanings across the instant and must not be pooled.

**Blast radius, measured not assumed:** the council is running now — 1
council_report row after 13:56:14Z, 5 after the 11:02Z commit, newest
14:07:50Z. The **current** corpus (last extracted 07-20) has zero post-boundary
rows, so nothing already emitted is wrong; the *next* council-lane extraction
will cross it.

**What I did NOT over-claim:** `dissent` and `contested` are computed by the
extractor from raw per-seat votes, independent of how the round decision was
reached — so they are unaffected and I said so rather than flagging the whole
lane. [VERIFIED by reading buildCouncil: dissent = seat-vs-same-round-majority;
contested = approve>0 && other>0; neither reads RoundDecision.]

**Recorded structurally, not as a scrollback note:** RUNBOOK §5b boundary table
extended (now four). Extractor now emits `labels.round_decision_rule` on every
council row (`any_objection_gates` | `high_severity_gates`, keyed on the row's own
created_at vs `councilSeverityGate = 2026-07-22T13:56:14Z`) so a consumer never
redoes the arithmetic. Compiles, vet clean, pattern-check silent.

**Boundary uncertainty, stated honestly:** the two commits landed 09:06Z/11:02Z;
an earlier roll that day may already have carried them. The pod start is the
LATEST instant it could have gone live, so the boundary is `≤13:56:14Z` — rows
before it are conservatively old-rule, rows on/after confirmed new-rule. [The
prior pod's start time is unrecoverable, so I cannot tighten this further without
it — marked `≤` rather than guessing.]

Two other 1141→1149 changes noted, NOT boundaries for my read contract: feed
render window 72h→720h (`d3c2f95db`) shifts the `expired` distribution downstream
but not the score_relevance shape — reinforces "don't train `expired` as a
negative"; and `7be33718f` fixed a NEW `.result` blank-render recurrence in the
CLASSIFIER's mission seat — a different subsystem from the fix-loop council I
mine, so [ASSUMED out of lane, not verified against a council_report sample].

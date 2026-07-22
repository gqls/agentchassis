# HANDOFF / RESUME — reasoning dataset, current state 2026-07-21

*Cold-start entry point for this workstream. The founding brief
(`HANDOFF_2026-07-18_reasoning_training_dataset.md`) is kept as the record of the
original intent and the reframe; read it for the **why**. This file is the **where
we are** — read it first, then that one for background. Self-sufficient.*

---

## One-paragraph status

The read-only extractor is **built, proven, and committed** (`cmd/reasoningset/`).
It has mined all three reasoning lanes the plan identified. The corpus stands at
**7,712 records across 448 trajectories**, up from a planned handful. The blocking
question the plan set in advance — *is this worth a training run* — has a
**measured** answer: **no to training, yes (modestly) to evaluation, and the real
volume is in the two lanes that need no new spend.** Nothing here changes the
platform; it reads the live DB and emits JSONL. This thread stays **read-only for
`platform/`, `internal/`, `pkg/`** — the two platform fixes it found were handed
to the threads that own that code.

## What exists and is committed

| piece | what it is | commit |
|---|---|---|
| `cmd/reasoningset/extract.sql` | 5 tagged queries: `step`, `trail`, `bundle` (diagnosis), `council` (per report×seat), `feedjudge`/`feeditem` (news-feed triage) | — |
| `cmd/reasoningset/main.go` | Go transformer: `build` / `buildCouncil` / `buildFeed`, shared `commonExclusions()` gate, provenance boundaries, derived labels | `c08b5ad7a` (latest) |
| standing docs | this dir — see the map below | — |

Run it: **`RUNBOOK_reasoning_dataset.md`** has every command with its gotcha. The
one that bites: `kubectl exec` stdout **silently truncates** large streams at
~47 MB (706 of 5,662 rows, exit 0, no error) — pipe through `gzip` *inside* the
pod. `psql -f -` under `exec` also truncates after the first statement — drop
`-f -`, psql reads stdin by default.

## The three lanes, as mined

**Diagnosis lane — 820 records / 112 trajectories.** 689 usable, 131 flagged with
the reason (blinded docs, truncation, premise-shift). This is the fix loop's own
verdict trail: cited theory → evidence → CONFIRMED/REFUTED/UNVERIFIABLE. **102
usable diagnosis steps**, spread across two model generations — too few and too
mixed to train. **24 gold-graded steps over 6 trajectories** hand-curated in
`LABELS_benchmark.json` (twelve runs read by hand — "FAILED" means *overloaded
API* in one note and *never started* in another, so no scraper). Two of the
twelve correctly yield zero records because those runs never ran.

**Council lane — 836 seat reviews.** Labels are **derived, not curated** — that is
what makes it usable at scale. Several seats judging **one plan** with each seat's
agreement/dissent from its **same-round peers** attached = a preference set, the
format reasoning post-training wants. Live findings, all corrected in place after
first getting them wrong:
- Dissent is **201 (24%)** measured against same-round peers, **not** 62% against
  the round decision (a `revise` fires if *any* seat objects, so that made every
  approver a "dissenter").
- **`contested` is near-useless** — 822 of 836 votes sit in a split round.
- **Six seats have never objected and are not rubber stamps** — they write scope
  determinations ("outside my footprint"), a legitimate approve. Their `approve`
  ≠ `editquality`'s `approve`; 155 records flagged `scope_approve_risk` so a
  consumer can separate the two.
- **Trailer terminal-outcome idea is dead:** only **4 `Council-Reviewed:` trailers
  exist** against 44 submissions. An anecdote, not a label source.

**Feed-triage lane — 5,671 judgements, 5,658 with a terminal outcome** (rejected
3,472 · expired 1,385 · relevant 782 · review 19). The **only** lane where
judgement and outcome sit on the same row. Checked the circularity worry rather
than assuming: **492 judgements scored ≥60 and were rejected anyway**, and
`duplicate_of`/expiry explain none — flagged `score_outcome_divergence`, the most
interesting subset, and an **open question** (something downstream decides, not
yet identified).

## The go/no-go verdict (the gate the plan set)

| question | answer |
|---|---|
| Train a reasoning model on this? | **No** — 102 usable diagnosis steps, two model generations. |
| Use it as an evaluation set? | **Yes, modestly** — 24 gold-graded steps / 6 trajectories. Two of three passes are *abstentions graded correct*; the one failure is a *confident* wrong answer. A benchmark where declining to answer is the gold behaviour is scarce. |
| Where is the unexploited volume? | **Council + feed lanes** — derived/co-located labels, no new spend, no platform change. |

The honest ceiling: this platform makes **hundreds** of decisions a day, not
millions. Six months of all lanes is *tens of thousands* of outcome-labelled
steps — fine-tuning/eval scale, not pretraining. The goal is a scarce, high-quality
specialist corpus, not a big one.

## Standing-five docs map

- `PLAN_2026-07-18_reasoning_dataset_extraction.md` — extraction design + corrections
- `PLAN_capture_gaps_and_volume.md` — how to get MORE data (the volume question)
- `RUNBOOK_reasoning_dataset.md` — every command, each gotcha attached
- `NOTES_reasoning_dataset.md` — technical log, append-only
- `NOTES_corpus_quality.md` + `_council_addendum.md` (ADDENDUM 2 = feed lane) — the report and its threats to validity
- `README_where_we_are.md` — owner's plain-prose log (append-only, never rewrite)
- `SUMMARY_*_reasoning_dataset.md` — snapshot series; latest is `2026-07-20b` (predates council+feed being mined — a new SUMMARY is due once the self-containment join lands or the volume decision is taken)
- `LABELS_benchmark.json` — the twelve curated gold runs, each quoting its source

## Next actions

1. **Make council records self-contained** (proposed, not started). Two joins are
   available but not materialised: (a) the **plan under review** — `input_state`
   is empty on council rows; the `fix_plan` artifact lives on the same
   `trajectory_id`; (b) **per-seat model provenance** — joinable to `llm_call_log`
   on `(correlation, review_<seat>)`, not joined today, so council rows can't be
   sliced by model generation the way verdicts can. Both are in the addendum's
   "Honest limits".
2. **Owner decision — generate data at volume?** Cheapest form: replay the ~30
   bugs already solved, where the correct diagnosis is written down and grading is
   near-free — this *also* refills the dead revise lane (all 23 revise steps
   reasoned against blank objections; they predate the `bugs_open/016` render fix).
3. **Do not extract again for its own sake.** The pipeline is proven; more passes
   over the same 13 diagnosis runs add nothing.

## Landmines carried forward

- **Read the RUN start, never the step timestamp**, when grading a config change —
  step timestamps post-date a fix whose run began before it (WRONG_CALLS #2).
- **Provenance boundaries are dated cutoffs baked into the extractor** (RUNBOOK §5b
  is the authoritative table): the `bugs_open/016` render fix (2026-07-18
  13:15:11Z); chassis v1.0.1140's `bugs_open/032` fix and the `VerifyTarget`
  widening (2026-07-20 17:58:20Z); and — added 2026-07-22 — chassis **v1.0.1149**'s
  **council decision-rule change** (≤2026-07-22 13:56:14Z, pod-verified): only a
  HIGH-severity objection now gates a round to `revise`, so `round_decision` means
  two things across the instant. The extractor stamps `labels.round_decision_rule`
  (`any_objection_gates` | `high_severity_gates`) per council row; `dissent`/
  `contested` are raw-vote measures and are UNAFFECTED. Rows before a boundary
  carry different meaning — the extractor flags, doesn't drop.
- **The coercion guard only ever DEGRADES a verdict, never upgrades** — an extract
  showing raw-UNVERIFIABLE → coerced-CONFIRMED is a bug in the extractor, not the
  data (it was, once; fixed by asserting only when verdict/trail counts agree).
- **Silent truncation the enshrined rule misses**: 50 of 439 feed `score_relevance`
  responses are unparseable JSON, cut mid-array, `success=true`, and **zero** trip
  `output_tokens >= max_tokens`. That is the `bugs_open/008` stop_reason family in
  a lane nobody checks — surfaced to feed-triage's owner, not ours to fix.
- **Blinding leak**: fixloop docs are excluded from the loop's own input to keep
  benchmarks honest; keep them out of any eval set or you leak the answers.
  `commonExclusions()` gates all three lanes on this (it caught 3 council rows that
  would otherwise have reached the eval set).

## Platform defects found (handed off — NOT this thread's to fix)

- `bugs_open/032` — verifier reads a **deleted** target as a successful fix. Now
  CLOSED by the empty-sections thread; moved to `bugs_closed/`.
- `bugs_open/021` INSTANCE 2 — completion verifier registered for **1 of ~50**
  item types. Owned by the work-item-completion-integrity thread (submission B
  handed to them; they corrected the `ItemVerifier` signature that made a
  site-aggregate verifier unwritable — now `VerifyTarget`).
- `bugs_open/033` — human-review queue nobody could action. **Corrected**: a review
  surface *does* exist in `frontends/admin-dashboard` but shows only the newest 50
  (0 of the build-pipeline items). My "no surface" claim was a one-Go-layer search
  miss — the recurring WRONG_CALLS shape on this workstream.

## Two questions only the owner can answer

1. Is the human-review queue meant to be **worked or emptied**? (It is known now to
   be a near-blind UI.)
2. Is generating training data at volume worth the spend?

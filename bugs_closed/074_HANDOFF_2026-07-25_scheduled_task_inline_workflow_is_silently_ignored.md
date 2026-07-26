# 074 — a scheduled task's inline workflow is silently ignored; the task reports success and does nothing

> **CLOSED 2026-07-26 — FIXED AND LIVE.** The shape can no longer be authored (a CHECK
> constraint, migration `217`, live on apply), no row carries it, and the one remaining casualty
> now runs for real. Full resolution at the bottom of this file; working docs in
> `docs/agent_docs/docs024_key_docs_latest/bugfix_074_inline_workflow/`.
>
> **CORRECTION to the account below, and it is the load-bearing one.** This file says "the chassis
> runs the target agent's own workflow, not the one in `input_data`" — true, but it stops one step
> short, and the missing step decides the fix. **The chassis DOES honour an inline workflow — at
> `body.config.workflow`** (`platform/messaging/processor.go:897-903`, `selectWorkflow` Priority 1,
> taken ahead of group discovery), and that is a live production path, not a testing hook: 58
> `orchestration_states` rows carry the inline workflow `DispatchFeedSourcesAction` dispatches that
> way (`dispatch_feed_sources_action.go:224`). What is broken is that **the scheduler cannot
> express it** — `fireTrigger` builds `config` from the row's *columns*, so a workflow authored in
> `input_data` lands at `body.input_data.config.workflow`, one level below anything that reads it.
> Caught by reading `selectWorkflow` before writing the fix, which is the only reason the fix is a
> refusal rather than a lift.

**Filed:** 2026-07-25, by the model_directory_pipeline session, after inducing a fault in a
sweep it had already described as "live and proven".
**Severity:** High — this is a **silent no-op with a green status**. Both timestamps
advance, no error is logged, and the only way to notice is to check whether the work
actually happened. At least three tasks are affected, one of them another workstream's.
**Class:** structural (a config shape that looks supported, is accepted without complaint,
and is never read).
**Status:** ~~OPEN~~ **CLOSED 2026-07-26 — fixed and live** (see RESOLUTION at the foot of this
file). As filed: the model_directory half was fixed and applied; the `evidence-freshness` half
belonged to the claims_verification workstream and was untouched.

---

## The shape that doesn't work

A `scheduled_tasks` row carrying its workflow inline:

```json
target_agent_type: "generic",
input_data: { "config": { "agent_type": "generic",
                          "workflow": { "start_step": "refresh_claims", "steps": { ... } } } }
```

The chassis runs the **target agent's own** workflow, not the one in `input_data`.
`generic`'s own workflow is a single no-op:

```json
{"start_step":"complete","processing_mode":"task","timeout_seconds":10,
 "steps":{"complete":{"action":"complete_workflow",
          "description":"No-op — scheduled task pre_query already did the work"}}}
```

So the task fires, an orchestration is created, it goes straight to `complete`, and
`last_triggered_at` **and** `last_completed_at` are both stamped. Everything a
health check would look at says the sweep ran.

## Measured, not inferred

```sql
-- orchestrations that have EVER carried the action, since the pipeline was built:
SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_directory_claims%';
--  0

-- claims ever re-verified after registration (the sweep's only fingerprint):
SELECT count(*) FILTER (WHERE verified_at > created_at + interval '1 minute'), count(*) FROM directory_claims;
--  0 | 108

-- the workflow one of those "successful" runs actually executed:
SELECT jsonb_pretty(workflow_plan) FROM orchestration_states
WHERE correlation_id='a9225ec2-06fa-4da0-b40f-0800e7c8379f';
--  {"start_step":"complete","steps":{"complete":{"action":"complete_workflow",
--    "description":"No-op — scheduled task pre_query already did the work"}}}
```

## Affected tasks

```sql
SELECT name, target_agent_type, jsonb_typeof(input_data->'config'->'workflow'), last_triggered_at
FROM scheduled_tasks WHERE input_data->'config' ? 'workflow';
```

| task | owner | status |
|---|---|---|
| `model-directory-freshness` | model_directory_pipeline | **FIXED 2026-07-25** — repointed at a new `directory-freshness` agent that carries the workflow; now sweeps all kinds daily |
| `adoption-tracker-freshness` | model_directory_pipeline | **disabled 2026-07-25** — redundant once one task covers every kind; row kept as evidence of the shape |
| `evidence-freshness` | claims_verification | **UNTOUCHED — their call.** Same shape, same `generic` target. Whether their V5 evidence re-verification has ever run is a question for them, and the two queries above answer it in seconds |

## How it was found, which is the transferable part

Not by inspection — by **inducing a fault**. The register had just taken 17 company claims
with 17 verifications and zero rejections, which my own seed file said to treat as
suspicious rather than good. So one stored quote was corrupted to a sentence that appears
on no page and its `verified_at` backdated past its staleness, and the sweep forced. The
expected flip to `citation_lost` never came — and chasing *why* is what exposed that the
sweep had never run at all, for any claim, ever.

The happy path and the broken path are **indistinguishable** here: task fires → completes
→ timestamps advance. Every check short of "did the work actually happen" passes. This is
the `verify the failing branch` rule earning its place: a green run proves deployment, not
correctness, and for anything whose job is to DETECT a fault, only an induced fault proves
it works.

## Fix candidates

1. **What the model_directory half did (applied):** move the workflow to where the chassis
   reads it — an `agent_definitions` row — and point the task's `target_agent_type` at it.
   Uses the mechanism already proven in the same pipeline (`directory-researcher` runs this
   way and works). Config-only, no image roll. See
   `docs/.../model_directory_pipeline/SEED_directory_freshness_agent.sql`.
2. **Make the silent case loud (the real fix).** Either honour `input_data.config.workflow`
   in the scheduler/chassis, or **refuse the task** when it carries a workflow that will be
   ignored. A third option — log a WARN naming the task — is the cheapest and would have
   turned three months of nothing into one grep. Whichever: the current behaviour of
   accepting the field and discarding it is the defect, not the field's absence.
3. **A guard**: assert at seed time (or in a periodic check) that no enabled
   `scheduled_tasks` row has `input_data->'config' ? 'workflow'` while targeting an agent
   whose own workflow is the `generic` no-op. One query; catches the whole class.

## How to verify any fix

Not the timestamps. Run the two count queries above and expect non-zero, **then induce a
fault** and watch for the `citation_lost` transition. Restore from the backup table after.

---

# RESOLUTION — 2026-07-26

**Owner call taken with both options on the table: reject the shape, do not teach the scheduler to
read it.** `bugs_closed/054` settled that `scheduled_tasks.input_data` is the payload only and that
`fireTrigger` must not reach into it (*"a legitimate payload field named `input_data`/`action`/
`config` would be silently eaten"*). Lifting `input_data.config.workflow` would reopen exactly
that, and would leave two competing homes for a workflow. A workflow's home is an
`agent_definitions` row.

## Three legs — prevent, repair, detect

**A. Prevent — live on apply, no image roll.** Migration
`217_scheduled_tasks_reject_inline_workflow.sql`:

```sql
ALTER TABLE scheduled_tasks ADD CONSTRAINT scheduled_tasks_no_inline_workflow
  CHECK (NOT (input_data -> 'config' ? 'workflow'));
```

Verified on the **failing** branch — a probe row carrying the shape is refused with SQLSTATE 23514
naming the constraint — and with a **positive control**, an ordinary payload that still inserts, so
the refusal is not passing for an unrelated reason. Deliberately narrow: `diagnose-pipeline-trigger`
and `ai-endpoint-health-check` carry 054-family envelope keys but no workflow, and stay legal.
Recorded in `schema_migrations` by `--record-only` after a hand apply (never `--apply`: 13 files
were pending, most of them other threads').

**B. Repair — the casualty now runs.** `evidence-freshness` got its own `agent_definitions` row
carrying the workflow, mirroring `SEED_directory_freshness_agent.sql`; the task points at it and
its `input_data` is `{}`. Staged through a **`dry_run: true` pass first**, because the sweep
regenerates `writer_block` on two sites other threads are actively working.

The discriminating check — this count could not have been non-zero before, for any reason:

```sql
SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_evidence_base%';
--  before: 0     after: 3
```

The first real pass (correlation `4c18e688-…`) wrote **3 new `site_specs` revisions** as
`created_by='evidence-refresher'`, each noted `V4 freshness pass: N live-verifiable fact(s)
checked…`, `pinned` preserved exactly (leopardess `t`→`t`, checked against a backup taken first),
and raised **3 `stale_evidence` items** for human review. 8 sites swept, 24 sql-sourced facts, 0
errors.

**Then an induced fault**, because a green pass proves deployment, not detection: one leopardess
fact corrupted to `9370` against a live `937`, `verified_at` backdated. The next pass returned it
`drifted` — *"live value 937 is outside tolerance \"exact\" of the published 9370 (moved down)"* —
and re-synced the number itself. **That induced fault also found a new defect, filed as
`bugs_open/091`**: while an earlier `stale_evidence` item is open, a second, *different* drift is
dropped by the work-item dedup and the run still reports `work_item_created: true`.

`adoption-tracker-freshness` (disabled, superseded) had its never-read workflow stripped so the
constraint could go on VALID. Preserved verbatim, since the row was being kept as evidence:

```json
{"config": {"agent_type": "generic", "workflow": {"start_step": "refresh_claims",
 "processing_mode": "orchestrator", "timeout_seconds": 600, "steps": {
   "refresh_claims":    {"action": "refresh_directory_claims", "config": {"kind": "company"},
                         "next_step": "refresh_protocols", "output_field": "refresh_result"},
   "refresh_protocols": {"action": "refresh_directory_claims", "config": {"kind": "protocol"},
                         "next_step": "complete", "output_field": "refresh_protocol_result"},
   "complete":          {"action": "complete_workflow",
                         "config": {"output_fields": ["refresh_result", "refresh_protocol_result"]}}}}}}
```

**C. Detect — committed, INERT until the next `kafka-scheduler` build** (`16315d5ab`). The
scheduler now refuses to fire a task whose workflow it cannot deliver, logging the task, the
ignored `start_step` and the remedy. It refuses rather than fires because the manufactured green
orchestration is the actual harm; it still stamps both timestamps so the row rotates instead of
re-winning its group's only slot every tick (`bugs_open/048`), and claims no concurrency slot,
because nothing is running. With leg A in place this branch is unreachable in a healthy database —
it stays because a constraint can be dropped and an old row restored, and because a discard should
be visible where it happens. The test was falsified before being trusted (blinding the detector
fails three subtests; they pass on restore).

**Why this closes on legs A and B alone:** the defect is "the shape is accepted and discarded in
silence". It can no longer be authored, and no row carries it. Leg C is belt-and-braces and is
named as inert rather than implied to have shipped. After the next build, pod-grep a string the
change *created*: `strings /app/kafka-scheduler | grep -c "cannot deliver"`, with
`"Pre-query found no rows"` as the positive control.

## Not council-reviewed, and why

`097_TRIGGER` refuses submissions touching none of `platform/`, `internal/`, `pkg/`. This work is
`cmd/` + SQL + docs, so the gate declines it client-side. Stated here rather than left as a silent
gap in the `098` coverage report.

## Follow-ons left for their owners

- **`bugs_open/091`** — the dedup/reporting defect the induced fault exposed. `insertWorkItem` is
  shared machinery (work_item_completion_integrity's remit); V4 is claims_verification's. Not
  fixed here on purpose.
- **`evidence-freshness` is now live and daily**, and the sweep **supersedes** the evidence_base
  spec (`is_current=false` + INSERT). Any thread holding a `site_specs.id` for an evidence base
  must re-SELECT the current row before writing. Flagged to claims_verification and to the
  brochure/043 lanes in their own notes.
- **`016b` §9 — done.** It already carried an entry for this case (written by the filing session)
  whose recommended fix pattern is the one taken; it is now amended with the two things it did not
  have — "the receiving end DID support the field, the sending end could not express it", and
  "refuse at authorship, not at use" — and its pointer corrected to `bugs_closed/074`. The
  amendment waited about an hour: another session had 98 uncommitted lines in that file, and a
  same-file passenger cannot be excluded by a pathspec, so it was parked in the workstream dir
  until they committed. Nothing is owed.

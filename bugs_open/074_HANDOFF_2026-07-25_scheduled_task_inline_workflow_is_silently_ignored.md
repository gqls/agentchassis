# 074 — a scheduled task's inline workflow is silently ignored; the task reports success and does nothing

**Filed:** 2026-07-25, by the model_directory_pipeline session, after inducing a fault in a
sweep it had already described as "live and proven".
**Severity:** High — this is a **silent no-op with a green status**. Both timestamps
advance, no error is logged, and the only way to notice is to check whether the work
actually happened. At least three tasks are affected, one of them another workstream's.
**Class:** structural (a config shape that looks supported, is accepted without complaint,
and is never read).
**Status:** OPEN. **The model_directory half is FIXED and applied** (see below); the
`evidence-freshness` half belongs to the claims_verification workstream and is untouched.

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

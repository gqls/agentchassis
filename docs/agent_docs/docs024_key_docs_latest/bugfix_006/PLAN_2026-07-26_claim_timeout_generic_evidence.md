# PLAN 2026-07-26 — bugs_open/006 §C: generic completion evidence for the claim-timeout sweep

Sibling of `PLAN_2026-07-24_contact_form_hardening.md` (which covers §B). This one covers
**§C only**, the last unanswered part of `006`, and the decisions behind the shape it took.

## The problem, restated from measurement rather than from the file

`claimed-item-timeout` (a `scheduled_tasks` row, `interval_seconds=120`) is the fleet's only
self-heal for a claim whose handler finished but whose `mark_complete` write never ran. It has
two branches: **auto-complete on evidence at 15 minutes**, and **reset to `triaged` at 40
minutes**.

The evidence branch is a hand-written artifact test **per `item_type`**, and there are three:
`needs_content_page`, `page_rerender`, `needs_design`. Everything else falls through to reset —
including work that demonstrably succeeded.

Re-measured 2026-07-26 (the bug file's figures were from 07-20 and are stale):

| item_type | timed out | auto-completed |
|---|---|---|
| `page_rerender` | 28 | 9 |
| **`needs_page`** | **27** | **0** |
| `content_rewrite` | 15 | 0 |
| `needs_imagery` | 6 | 0 |
| `needs_content_page` | 0 | 5 |
| *(5 further types)* | 8 | 0 |

84 timeouts against 14 auto-completions over 14 days. Outcome split: most items reached
`complete` on a retry — i.e. **the work was done twice** — and 11 ended `failed` with 1
`unresolved`, attempts exhausted on work that had already succeeded.

## The decision: one generic branch, not fifteen more per-type ones

The bug file already framed the choice: *"The structural fix this section asks for (complete
atomically, or retry the completion write idempotently) is still the right one, and would make
the per-item-type evidence branches unnecessary rather than needing fifteen more of them."*

**What made it tractable was noticing the evidence is already recorded, once, for every item
type.** Both dispatch loops pass the item id into the handler via
`call_handler.config.input_mapping`, so the handler's own orchestration row carries it:

```
initial_request_data->'input_data'->>'work_item_id'
```

That is not item-type-specific, and it is not something we had to add. So the branch is:

```sql
wi.item_type NOT IN (<verifier-backed types>)
AND EXISTS (SELECT 1 FROM orchestration_states o
            WHERE o.initial_request_data->'input_data'->>'work_item_id' = wi.id::text
              AND o.status = 'COMPLETED'
              AND o.updated_at > wi.claimed_at)
```

### Decision 1 — the standard is PARITY with the lost write, not correctness

The tempting instinct is to make the sweep *smarter* than the `complete_work_item` call it is
standing in for. **That is wrong.** The dispatch loop completes an item whose handler saga
reached its own `complete_workflow`; the sweep is recovery for that write being lost. A recovery
stricter than the write it replaces leaves finished work re-running — which is the defect being
fixed. Quality judgement belongs at the write (where the verifier registry already sits), not in
the recovery.

### Decision 2 — the one place parity is unsafe, excluded exactly

`complete_work_item` is **not** unconditional: `verifyBeforeComplete` consults a per-`item_type`
verifier that can **block** completion (`bugs_open/017`, `/021`). SQL cannot call a Go verifier,
so the three item types that have one are excluded and keep falling through to reset. Strictly
safer than parity, and free today — none of the three appears anywhere in the 14-day timeout
data.

### Decision 3 — the exclusion list is pinned to the registry by a test that reads the migration

Two hand-maintained lists that must stay identical is the drift class this repo keeps paying
for, and here the drift is **silent and one-sided**: register a fourth verifier, leave the SQL
alone, and the sweep auto-completes an item the verifier would have blocked, with no caller to
notice because a sweep has none.

`TestRegisteredVerifiersMatchClaimTimeoutExclusion` **reads the migration file** and extracts its
list rather than holding a third literal copy in the test — *a guard whose own copy can drift is
not a guard*. It asserts both directions and `t.Fatal`s rather than passing quietly if the
migration is missing or the SQL was reshaped past its regex.

### Decision 4 — the three artifact branches stay

`orchestration_states` is purged at roughly 2 days. They are the fallback for a claim whose
orchestration row is gone or was never persisted. Removing them would trade one gap for another.

## The alternative that was NOT taken, and why

**"Mark the item complete atomically with the last deploy step"** — the bug file's other
suggestion. Rejected:

- it means adding a `complete_work_item` step to ~18 handler agent definitions, duplicating the
  dispatch loop's own step in each — more surface, and eighteen places to drift;
- it does not actually close the hole. A pod that dies between writing the artifact and stamping
  the item loses the write wherever the stamp lives. "Atomic" here would need the stamp inside
  the same transaction as the artifact write, which handlers do not have (the artifact is written
  by a different action, often a different agent, sometimes a git commit);
- the generic evidence branch removes the *need* for it: the orchestration reaching `COMPLETED`
  is the durable record that a stamp would have been.

## Scope boundary, stated rather than implied

This fixes the **recovery** layer. It does **not** fix why the completion write is lost — that is
`bugs_open/003`'s pod-death / lost-child-response question (F2+F3 live `v1.0.1159`,
owner-ratified). §C asked for evidence that works for every item type; that is what is delivered.

## Why this could close §C today

`scheduled_tasks.pre_query` is **DB config — live immediately, no image roll.** The Go half is a
test, which changes no runtime behaviour. So the `bugs_closed` bar ("fixed AND live") is
reachable in one session, which is why §C closed while §B's tail is still healing.

## Verification design (all executed — see NOTES)

1. Shadow the new branch as a `SELECT` over the live `claimed` set **before** applying.
2. Three transaction-aborting guards in the migration, the third executing the **real stored
   `pre_query`** against a positive and a negative probe — not a restatement of the predicate,
   because a check sharing the fix's own expression cannot falsify it.
3. Fault-inject every assertion and watch it fail with the right diagnostic.
4. Then prove it **through the running scheduler**, not through psql, with the negative control
   present.

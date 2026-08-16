# 287 — build-dispatch-loop marks work items `complete` with the SPAWN RECORD as their result, not the handler's reply — appears at the 2026-08-15 10:14Z roll, ~75% of completions since

**Filed 2026-08-16** by the mortgagecalculator adoption lane, found while verifying that
`bugs_closed/274`'s fix (delivery of a child's reply to its parent) is in the live chassis. **It is,
and it works** — zero cannot-deliver rows against 859 child completions since the v1.0.1303 roll.
This is the *next* defect on the same seam, and it is a **regression carried by the same roll**.

## 1. The shape

A work item routed by `build-dispatch-loop` is marked `complete` and its `result` column holds:

```json
{"role":"handler","topics":{"requests":"job.<corr>-<x>-<handler>-process_item_iter_N_spawn_handler.requests",…},
 "agent_id":"…","agent_type":"<handler>"}
```

That is the **spawn record** — `spawn_agent`'s output (`extractSpawnData`, stored at
`process_item_iter_N_spawn_handler` and at output_field `handler_spawned`). It is NOT the
handler's reply, which `mark_complete` is configured to record:
`complete_work_item` with `"result": "handler_result"` (live `build-dispatch-loop` config,
`process_item.sub_workflow.steps.mark_complete`; `call_handler` is `call_agent` with
`output_field: handler_result`).

The child does its work and COMPLETES; the reply DOES arrive; the parent's own persisted
`collected_data.handler_result` DOES end up holding the correct `{"response":{…}}`. Only the
item's stored `result` is wrong. So: **the work is done, the record of it is someone else's.**

## 2. Scale [MEASURED 2026-08-16 ~10:10Z]

```sql
SELECT date_trunc('hour', updated_at),
       count(*) FILTER (WHERE result ? 'response')                        AS own_envelope,
       count(*) FILTER (WHERE result ? 'topics' AND result ? 'agent_id')  AS spawn_record,
       count(*) AS total
FROM site_work_items WHERE status='complete' AND updated_at > '2026-08-15 04:00' AND handler_agent IS NOT NULL
GROUP BY 1 ORDER BY 1;
```

| window | own envelope | spawn record |
|---|---|---|
| 08-15 08:00–10:13 (pre-roll) | 136 | **0** |
| 08-15 10:14–11:00 | 8 | 5 |
| 08-15 12:00 | 5 | 26 |
| 08-15 14:00–16:00 | 55 | 168 |
| 08-15 19:00 (post-1303 roll) | 22 | 120 |
| 08-16 08:00 | 3 | 3 (+27 "other") |

**Zero instances in any hour before the 10:14Z roll; the shape appears in the roll's first
hour and dominates thereafter** (~270 vs 70 since 18:46Z). By handler since 18:46Z:
page-rerender 221, page-build-handler 19, image-build-handler 8, component-template-fixer 7,
tool-improver 4, color-variable-fixer 4, content-gap-planner 4, section-editor 4.
The pre-roll "other" bucket (42 at 08:00) is a different, older shape — not this.

## 3. A worked instance, timeline intact (`c0aee25f…`, page-rerender, gaswholesalers)

| t (UTC) | event |
|---|---|
| 08:39:56 | child `page-rerender` created (parent `build-dispatch-loop`) |
| 08:40:05 | child COMPLETED |
| **08:40:06** | item `c773ee3d` marked `complete` — result = spawn record |
| 08:40:46 | parent COMPLETED; its `collected_data.handler_result` = `{"response":{"deploy_result":{"success":true,…}}}` and `process_item_iter_1_call_handler` = same |

So `mark_complete` ran ONE second after the child completed and read `handler_result` as the
spawn record; the parent's final state shows the reply merged at `handler_result` — **after**, or
onto a copy `mark_complete` did not read. `call_agent`'s own pre-park return (`buildCallResult`,
`call_agent.go:647`) has keys `agent_called/agent_type/request_id/child_orchestration/…` — NOT
`role/topics/agent_id` — so what the item holds is not call_agent's placeholder either; it is
the **previous step's** output, which is what `resolveFieldValue`'s fallback search finds when
the asked-for key is absent (`[INFERRED — the resolution at mark_complete time is the thing
the 090 run is asked to read]`).

## 4. Why "the same roll" — what changed on this seam that day

Two commits rode the 10:14Z roll on exactly this path:
- `919cc6976` (274 fix) — reply envelope headers, so the child's reply now VALIDATES and is
  delivered. Before it, the reply never arrived, the parent was told the child FAILED, and
  the item was completed with whatever `handler_result` resolved to (that produced the
  08-11/08-14 §D shape: a *content-planner* payload). **So the pre-roll `own_envelope`
  rows were NOT reply-driven** — with the reply never validating, they must have come
  from another path `[UNMEASURED which]`.
- `3ba384c63` (WFA-014, RFC_012 (a)) — `persistAwaitingStateWithRetry` now CARRIES the
  dispatching step's in-memory `collected_data` onto the fresh DB load at park
  (`carryCollectedDataOntoFreshState`, coordinator.go:2169). Before it, the park
  discarded them.

The register entry for WFA-014 already names the residual it does not cover: *"this fixes
the PARK only — `applyResponseToState` still REPLACES wholesale on its `output_mapping` and
default branches, so a carried key can still die at reply time."* The `call_agent` branch is
the additive `isAgentResponse` one (coordinator.go ~2820-2855), so wholesale-replace is not
this — but the ORDER of (reply merged onto which state) vs (next step `mark_complete` reads
which state) is the open question. `[HYPOTHESIS, filed to the loop, not asserted]`: the
resume-from-park path loads a fresh state, merges the reply, and continues — while
`mark_complete`'s read of `handler_result` happens against a state where the key holds the
carried pre-reply value or is absent, so the recursive fallback lands on `handler_spawned`.

## 5. Consequence

- Every completion-time reader of `site_work_items.result` for these ~270 items sees a spawn
  record: verifiers (`bugs_closed/213`'s gate reads the handler's self-report — 4-of-14
  became ~0-of-N), retraction logic, operators, and the two-strike rule (a `complete` is a
  strike whatever it holds).
- The WORK is fine — children complete, artefacts land (verified: mortgagecalculator's logo
  redeploy `undeployed_asset:e766370e` at 19:12Z is one of these rows; `logo.png` was and is
  200). So this is a records defect, not a delivery defect — the mirror image of 274.

## 6. Diagnosis status

`090` filed 2026-08-16 ~10:15Z: intake `de3b436a-d8da-4c4c-a8fe-8b3abcb0d461`, run
`fb7ae3bc-e9bf-4a96-b540-d593b91bc79c`. **⚠ The trigger warned that local HEAD is 853 commits
ahead of origin/087, and the loop reads ORIGIN — so it may be reading a coordinator WITHOUT
WFA-014.** Read its verdict with that in mind; if it cites pre-3ba384c63 line numbers, its
mechanism read is of the wrong binary. Grep-before-file: `bugs_open/236` recorded the park
discard (fixed by WFA-014); `bugs_closed/274` the delivery half; `bugs_closed/213` §D the
foreign-payload symptom. None names this post-roll shape.

## 7. Fix candidates (ordered by what closes the door — NOT yet reviewed, do not act on §7 alone)

1. **Make `mark_complete` read the reply, not a path that can resolve elsewhere**: have
   `complete_work_item` refuse to record a `result` that lacks the `response` envelope when
   the preceding step was a `call_agent` (or read `process_item_iter_N_call_handler.response`
   directly, which the parent row shows IS correct at the end). Closes the door for this
   loop; does not fix the ordering.
2. **Fix the ordering at the seam**: ensure the reply merge lands on the same state the next
   step reads before that step executes — RFC_012 territory (every await-using pipeline in
   the blast radius); route to that decision with §3's timeline attached.
3. **Kill the fallback**: `resolveFieldValue`'s recursive search returning a sibling step's
   data for a missing key is the same landmine `bugs_open/248`'s R1 objection named
   (`findFieldRecursive`); a missing `handler_result` should be an error at
   `complete_work_item`, not a silent substitution.

## 8. How to verify a fix

Post-roll, `SELECT count(*) FILTER (WHERE result ? 'topics' AND result ? 'agent_id')` over
`status='complete' AND updated_at > <roll>` must be 0 **while** `count(*) FILTER (WHERE result
? 'response')` is non-zero in the same window (demand). And one item followed end to end:
child COMPLETED → item `complete` with `result.response` = the child's payload.

# 343 — a parent freezes silently after ABANDONING an await, instead of failing

**Filed 2026-08-20.** Split out of `bugs_closed/029_HANDOFF_2026-07-19_hung_spawns_saturate_dispatch_group_and_halt_builds_fleetwide.md`
(owner ruling, 2026-08-20) — that file's retry-window defect is fixed and live; **this is the half
that is neither**, carried out under its own number so a fixed Part A can never read as a fixed hang.

**Status: OPEN. Not reproducing. Rare, bursty, and fully characterised except for its mechanism.**

> **Read this first, because the number it replaces was misleading in both directions.** 029's title
> said "fleet-wide outage class"; that framing belonged to 2026-07-19 and to its *consequence* half.
> This bug is **not** a fleet-wide outage. It is a narrow, silent, recoverable-but-slow freeze that
> needs an unusual trigger. It is also **ours**, which is why it stays open.

## What it is, in one paragraph

A `build-dispatch-loop` orchestration sends a request to a child, the child never answers, and the
parent gives up after three retries. **That give-up is where the defect is: instead of failing the
orchestration, the parent stops writing state altogether.** It sits in `EXECUTING_STEP` holding no
waiting awaited request, so nothing routine notices it — only the 4-hour stale reaper. The build it
was running simply never finishes.

## What is measured, and how confident each part is

All figures `[MEASURED 2026-08-19/20]` from `awaited_requests`, reproduced independently by two
sessions. Evidence is **preserved**, so none of this expires:
`docs/agent_docs/docs024_key_docs_latest/bugfix_029_retry_kills_live_child/EVIDENCE_2026-08-15_to_17_awaited_requests.tsv`
— 6,484 rows, round-trip proven to rebuild 31/20/11/0 from the file alone.

| fact | evidence |
|---|---|
| **The entry condition is 100% terminal** | **31** `call_handler`s reached `retry_version=3, status='error'` in the retained window; **0** ever registered another `call_handler`. Positive control that could have come out otherwise, and did: **1,387** healthy `call_handler`s, **1,054 (76%)** continued |
| **It splits into two modes** | **20 WEDGED** — iteration N+1's `spawn_handler` IS registered, its `call_handler` never is (this bug). **11 STOPPED** — no N+1 `spawn_handler` either |
| **The requests were HUNG, not slow** | the rv0 window is **1200 s**, uniform over 1,386 rows. The slowest healthy child ever observed answered in **971.3 s** (n=3,150); **zero** exceed 1200 s. So a request reaching rv1 missed a window **229 s longer than any response on record** |
| **A retry cannot help** | `handleRecoverableError` (`coordinator.go:3108`) replays the **original request** to `awaited.RequestsTopic` — the topic stored on the row — and `UpdateAwaitedRequestRetry` writes only `retry_version` and `timeout_at`. The topic is **per-child-instance** (30/30 distinct; control: 1,387 healthy → 1,387 distinct). **Same child, no fresh clock** |
| **The process did NOT die** | `build-dispatch-loop` runs an ephemeral pod per orchestration (`spawn_actions.go:2388`, `BackoffLimit: 3`). The 17 duplicate spawn pairs are **17/17 the same `processing_pod`**, and all 98 rows across those orchestrations resolve to **exactly 17 pods** — one each, never replaced. `ActiveDeadlineSeconds` is 86400, so a job deadline is out too |
| **Nothing in `awaited_requests` separates the 20 from the 11** | iteration number, pod, error duration, timeout budget, rows-per-orchestration and `target_agent_type` all fail. **This route is closed** — do not re-run it hopefully |
| **Not reproducing** | wedged instances: 08-17 **20**, then **0** on 08-18 (1,594 spawns), 08-19 (736), 08-20 (429). Entry condition **0** since 08-18 |

## Why the one burst happened — CONTEXT, not cause

The 08-17 burst was triggered by an **external GitHub outage**: **954** GitHub-503 errors that day
against **1–3** on every other retained day, hours 13:00–18:00 with the wedge window inside it.
Controlled two independent ways — by correlation, **30/30 (100%)** of abandoned calls vs **71/337
(21.1%)** healthy; by page identity, **17/17 (100%)** vs **86/413 (20.8%)**.

**Do not mistake this for the cause of the bug.** It explains why 30 children stopped answering on
one day. It does not explain what the parent then did, and **~1 in 5 healthy pages hit a 503 and
completed anyway**, so a 503 is not sufficient. **Any** future cause of a child not answering within
1200 s — a slow dependency, a crashed child, a network partition — re-triggers this. The trigger was
external; the freeze is ours.

## Where to look

`PLAN_2026-08-19_wedge_fix_park_advance_divergence.md` in the lane directory is **still valid and was
never wrong** — but it only ever spoke to *this* half, so read it as this bug's plan, not 029's. Its
three verified source sites:

1. `coordinator.go:2113` `persistAwaitingStateWithRetry` — the beat-the-park check is keyed on
   **StepName**, not the request id; on a hit it returns `nil` **without** persisting, and the caller
   reads `nil` as success. Row in the table, nothing in the map.
2. `coordinator.go:2671` `handleCompleteResponse` — `allDone := len(freshState.AwaitedRequests)==0`,
   from the **map alone**. The table is never consulted.
3. `coordinator.go:848` `continueExecution` — silent early `return nil` when the loaded status is
   `AWAITING_RESPONSES`.

Plus **P2** (plan §2), which the 0-of-31 result **promoted**: `skipToNextLoopIterationForAsync`
(`loop_error_handler.go:243-260`) marks the row terminal and deletes the map entry in memory, then
persists with a single **non-retrying** `UpdateState`. An optimistic-lock failure there loses the
delete-plus-advance. It sits on exactly the path that now carries **31 terminal outcomes and no
survivors**. Filed there as `[VERIFIED as a path; UNVERIFIED as ever having fired]` — that is still
true, but it is the best-motivated candidate.

The diagnosis loop's own `NextScope`, reached independently: `executeStep`, `processAwaitResponse`,
`createContinuationContext`, `ProcessResponse`, `handleRecoverableError`.

## ⚠ Two traps that will cost you a day each

**1. `agent_error_log` has NO key that spans parent and child.** They are logged under *different*
`orchestration_id`s; there is no `correlation_id` column (it lives in `context` jsonb). `[MEASURED]`
**0 of 367** correlation ids appear as an `orchestration_id`, and neither 8-hex prefix in the child
topic matches a child row — the second is the **parent's** id (367/367). So a parent-keyed join sees
**only parent rows**, a correlation-keyed join sees **only child rows**, and nothing returns both.
The obvious key — the parent's `orchestration_id` — is populated, non-null and **structurally
blind**: it returns **27% vs 20%**, which *reads as a refutation of the outage finding*. **The only
sound linkage is the payload's `page_name`**
(`request_payload → message.body.input_data.spec.page_name` = the child's `context->>'page_name'`).
Four joins in this family came out blind across two sessions. See `LANDMINES.md`.

**2. Put the must-be-non-zero control IN THE SAME QUERY as the claim.** That is what caught the
27%-vs-20% one: a control column showed 30/30 abandoned parents *do* have error rows, so the join
demonstrably worked while the GitHub count stayed low — the signature of *looking in the wrong place*
rather than of *nothing being there*. An after-the-fact control gets skipped the moment the number
looks plausible.

## How to verify a fix

The entry condition has been 0 for three days, so **you cannot wait for it**. Either:

- **Induce it.** Make a child fail to answer within its 1200 s window and assert the parent **fails
  the orchestration** rather than going quiet — i.e. that it does not end in `EXECUTING_STEP` holding
  no waiting awaited request. Mutation-prove it: revert the fix and the test must fail.
- **Or capture the next natural one.** RSH-011 `wedge-evidence-capture` is live and hourly, and its
  capture path is induction-proven; it records the full `awaited_requests` set for a freeze while it
  is still happening.

**Do not treat quiet as evidence of a fix.** Six of the eight days around the 08-17 burst were also
zero, *before* anything was changed.

## Cold start

`docs/agent_docs/docs024_key_docs_latest/bugfix_029_retry_kills_live_child/HANDOFF_2026-08-19b_continue_here.md`
— the 2026-08-20 blocks. Detail in `NOTES_retry_kills_live_child.md` §§19–25 (§19 the control, §20
the refuted pod hypothesis, §21+§23 what Part A does and does not fix, §§24–25 the outage, confirmed
twice). The lane directory keeps its `bugfix_029_...` name; that is history, not a claim about which
bug it serves.

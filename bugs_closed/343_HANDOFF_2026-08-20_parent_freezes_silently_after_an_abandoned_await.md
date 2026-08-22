# 343 — a parent freezes silently after ABANDONING an await, instead of failing

> # ⛔ CLOSED 2026-08-22 — BY OWNER RULING, AND **NOT** BECAUSE THE BUG WAS SOLVED
>
> **This is a deliberate override of the fixed-AND-live bar, not a case that met it.** The owner was
> given the choice — hold the lane armed indefinitely, or close on the half that is fixed — and ruled
> **close it**. Read the next two paragraphs before citing this file for anything.
>
> **WHAT IS FIXED AND LIVE: the second wedge only.** `persistAwaitingStateWithRetry`'s arrival check
> is now keyed on request identity, `parkOutcome` makes "success without persisting" unrepresentable,
> and the advance decision cross-checks the table. Shipped in `ca5e41122` / `7f3875d3c` / `bf1fbc5b7`,
> live on `agent-chassis` **v1.0.1322** since 2026-08-21 16:54Z and proven at the binary with a
> control that came back absent. That work is sound and is not in question.
>
> **WHAT IS STILL UNEXPLAINED: the first death.** Why a child stopped answering, and why the parent
> never registered iteration N+1's `call_handler` **at all**, are exactly as unknown today as on
> 2026-08-20. Nothing shipped speaks to them. **So if you arrive here with a silent parent freeze,
> this file does not tell you it is fixed — it tells you a related mechanism was fixed and this one
> was closed anyway.** Re-open under a new number rather than assuming regression.
>
> **Why the owner could reasonably close it.** The 08-17 burst's entry condition was an external
> **GitHub 503 outage** (~300× base rate, that day only), not a coordinator phenomenon. The natural
> trigger may not return for months, and the route that needed no burst — the 20-wedged vs
> 10-stopped within-day control — was **run and came back negative** (`NOTES_retry_kills_live_child.md`
> §21: nothing in the retained columns separates the two modes, and `orchestration_states` is purged).
> So the lane had no live path left that did not begin with "wait for an unrelated outage".
>
> **`[MEASURED 2026-08-22]` at the close, with a demand control** — abandoned rv3 `call_handler`
> awaits are **30 on 2026-08-17 and zero on every other day in the table**, against a live volume of
> 592 `call_handler` awaits on 08-21 and 53 by mid-morning on 08-22. The zero is measured against
> real traffic rather than read off a quiet fleet. ⚠ **It is still not evidence of a fix** — six of
> the eight days around the burst were also zero *before anything changed*. It is why closing costs
> little, not why closing is correct.
>
> **RSH-011 `wedge-evidence-capture` stays ARMED and its trigger stays BROAD.** Closing the file does
> not retire the instrument: if the freeze recurs, the evidence that was missing in August gets
> captured. Its labels have been retargeted at this closed path (see the dead-label note below) —
> a pointer into `bugs_open/` would land nowhere, which is the exact defect fixed on 08-21.

**Filed 2026-08-20.** Split out of `bugs_closed/029_HANDOFF_2026-07-19_hung_spawns_saturate_dispatch_group_and_halt_builds_fleetwide.md`
(owner ruling, 2026-08-20) — that file's retry-window defect is fixed and live; **this is the half
that is neither**, carried out under its own number so a fixed Part A can never read as a fixed hang.

**Status: ~~OPEN~~ CLOSED 2026-08-22 by owner ruling (see banner). Not reproducing. Rare, bursty, and
fully characterised except for its mechanism — which remains unexplained at close.**

> ## 2026-08-21 — THE WEDGE MECHANISM IS FIXED AT SOURCE AND LIVE. The bug stays OPEN, and the reason is precise.
>
> Two commits, both **live on `agent-chassis` v1.0.1322 since 16:54Z** and proven at the binary with
> controls (own literals PRESENT; a pre-existing positive control PRESENT; a nonsense negative control
> ABSENT; and a literal from a commit made *after* the build ABSENT — so the probe demonstrably
> discriminates, rather than returning PRESENT for everything).
>
> **What was actually wrong, and it is smaller and sharper than this file's title suggests.**
> `persistAwaitingStateWithRetry`'s arrival check asked *"is there a `response` key under this step
> name?"* — a question that is **true of every iteration after the first** on a loop step. So each
> re-registration of an already-answered step name read as "the reply beat the park", the park returned
> a bare `nil` **without persisting**, and the caller read `nil` as success and went on to insert a
> table row and arm a timeout for a request that exists in no map. Row in the table, nothing in the map,
> nothing routine to notice. That is the **second wedge** — the one that keeps an iteration N+1
> `spawn_handler` registered while its `call_handler` never is, which is 20 of the 31 measured cases.
>
> **Shipped (`ca5e41122`, `7f3875d3c` — register RSH-012, WFA-021):**
> 1. `applyResponseToState` now records **which** request a reply answered, in **all four** branches.
>    Two of them wrote no arrival marker at all, and one of those — `output_mapping` — is **live on real
>    `call_agent` await paths** (`sql_for_agents/107_image_build_handler.sql:589`, `:1119`). So the check
>    was blind on a live path as well as wrong on the others.
> 2. The check is **keyed on that id**: same id → genuine arrival, skip; **different id → stale residue,
>    park normally** (this single discrimination is the fix); no id → dated legacy branch for the
>    mixed-fleet window, which has its own LANDMINES entry because it looks exactly like the logic the
>    rest of the diff condemns.
> 3. `parkOutcome` makes *"returned success without persisting"* **unrepresentable at the type level**;
>    on a skip the caller now inserts **no** row and arms **no** timeout, killing the orphan-row producer.
> 4. The advance decision now **cross-checks the table** it had never consulted
>    (`reconcileAllDoneAgainstTable` → `OutstandingAwaitedRequests`). Detection is **unconditional**
>    (`agent_error_log`, `error_code='await_divergence'`); enforcement is opt-in behind
>    `WorkflowPlan.AwaitReconcileEnforce`, default false, **no live workflow sets it**.
>
> **Also shipped (`bf1fbc5b7` — RSH-013), and explicitly NOT claimed as the mechanism:** P2's
> `skipToNextLoopIteration` wrote its advance with one **unretried** `UpdateState`; an optimistic-lock
> failure lost the advance *and* the awaited-map delete while the row was already terminal, so the
> redelivery was eaten by the `processed_at` duplicate guard. It now retries by reloading and
> re-applying, and marks the row terminal only once the advance is durable. **Verified as a path,
> UNVERIFIED as ever having fired.**
>
> **WHY THIS DOES NOT CLOSE 343.** The **first death** is untouched and unexplained: why a child stopped
> answering, and why the parent never registered iteration N+1's `call_handler` at all. Nothing above
> speaks to it. The bar is fixed AND live AND understood; this delivers the first two for the *second*
> wedge only.
>
> **Do not read a quiet fleet as confirmation.** The entry condition has been ~0 since 08-18 and six of
> the eight days around the 08-17 burst were also zero *before anything changed*. The informative
> reads are: `SELECT count(*) FROM agent_error_log WHERE error_code='await_divergence' AND occurred_at
> > '2026-08-21 16:54Z'` — **with** the demand control `SELECT count(*) FROM awaited_requests WHERE
> sent_at > '<same>'` non-zero, because a zero on a quiet fleet is a dead path, not a fix. And an
> RSH-011 capture showing **no marker** under the step name at re-park time would mean this mechanism
> is not the one that fired.
>
> Council: round 1 REVISE — the `editquality` seat read the submission's sketch as dropping the
> `exists`/`hasResponse` guards, which would make a *fresh* park skip. **The guards are in the shipped
> code; the sketch omitted them, and the sketch is what reviewers judge.** It nonetheless exposed a real
> TEST GAP — the two guards cover each other *in series*, so removing either alone was unobservable and
> nothing discriminated the branch. Closed by
> `TestParkPersistsWhenTheStepHasDataButNoResponseMarker`. Round 2 **APPROVED** (corr
> `a9b9d4b8-c30f-432a-b35d-5e5864733fc5`).

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

### The observed wedge is ENTIRELY inside the outage window — state this limit, do not overstate it

`[MEASURED 2026-08-20]` **All 20 wedge instances are on 08-17, inside the outage.** There is **no
observed wedge outside an outage window at all**, and that is a real bound on what we know: every
inference in this file about the wedge comes from one day on which an external dependency was
failing at ~300× its base rate.

**The 31st instance is NOT a counter-example, and it is worth being precise about because it looks
like one.** One entry-condition occurrence sits outside the outage — `51e9a384`, 2026-08-15 08:16Z,
on a day with **zero** GitHub errors (control: 08-17 returns 954, so the query sees what it should).
But it is a **STOPPED** case, not a wedge: no iteration-1 `spawn_handler`, and `max_iter_seen = 0`,
meaning **iteration 0 was the only iteration that orchestration ever had.** That is the most benign
shape available — the error fell on the loop's first and only item, so no next iteration was ever
owed. It therefore **strengthens** the benign reading of the 11 stopped cases rather than providing
non-outage evidence for the wedge.

> **So: `n=0` outside the outage for the wedge, not `n=1`.** If you want to know whether this freeze
> can happen without a 503 storm, the honest answer today is **that is unobserved**, and the way to
> learn it is the next natural occurrence (RSH-011 is armed), not this row.

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
  is still happening. **But read the next section before you interpret anything it banks.**

### ⚠ "The capture has banked N rows" is NOT evidence this bug recurred — and that is the FIRST check anyone will run

`[MEASURED 2026-08-21]` The capture has banked **4** orchestrations, **all after the burst**, all
titled `WEDGE EVIDENCE CAPTURE (live) — bugs_open/029`. **None of them is this bug**, and the check
that settles it is one line rather than four inspections:

> **This bug's entry condition is an abandoned rv3 await.** There have been **0** of those since
> 2026-08-18 — positive control, which had to be non-zero and is: **30** in the 08-17 burst window.
> **So nothing captured since 08-18 can be an instance of 343, whatever the note is titled.**

Independently: **0 of the 4** are even on a `build-dispatch-loop`. What they are —
`page-content-writer` at `process_sections_loop_iter_0_generate_content`, `endpoint-health-checker`
and `availability-discovery-agent` at `complete`, `generic` at `spawn_verifier` — with freeze times
of **0.039 s to 1m36s**, against this bug's shape of a freeze *after* a 300 s exhaustion.

**Why it will mislead:** the capture's trigger is *"an `EXECUTING_STEP` row older than the
threshold"*, which is far broader than this signature. Four rows in a table named `wedge-evidence`
reads as four wedges. It is not.

> **Label, fixed at source 2026-08-21 — so the four historical rows and any new one differ.** The
> script said `bugs_open/029`, a number that no longer exists. It now says
> `bugs_open/343 (was 029, split 2026-08-20)` in all three places
> (`deployments/kustomize/services/wedge-evidence-capture/base/check.py` — docstring, per-capture
> note title, run-summary title). **Verified at the artefact, not at the apply:** the CronJob mounts
> `wedge-evidence-capture-script-6bd8f58h7g`, whose data carries the new string, while the previous
> configmap still carries `bugs_open/029` — the control that shows the grep discriminates.
> **The four existing rows keep the old label and were deliberately not rewritten**, so searching for
> either string alone finds only part of the history. The script also now carries the
> entry-condition test below, because whoever reads a capture is reading a note body, not this file.

**RULING on the trigger, since it lives in this bug now: LEAVE IT BROAD.** Do not narrow it to the
343 signature. Three reasons, and the third is the one that decides it:
1. A narrow filter **would have missed all four**, and **two of the four rows have since been deleted
   by the cleanup** — i.e. the capture pre-empted exactly the loss it exists for, on rows a signature
   filter would have skipped.
2. The signature was derived from **one outage day** (`n=0` outside it, see above). Filtering on it
   would bind the instrument to the only sample we have, and a variant would be invisible.
3. **The two failure modes are not symmetric.** Breadth costs a misreading, which a written caveat
   fixes for free — this section. Narrowness costs the one capture that mattered, and that is
   **unrecoverable**. Prefer the recoverable failure.

**The two `complete`-step captures: CHASED AND CLOSED, 2026-08-21 — they belong to `bugs_open/040`.**
Two of the four were parked at `current_step = 'complete'` after **0.09 s** and **1.1 s**. Resolved:
both are Kafka write failures at the terminal step
(`complete_workflow: failed to send response: ... Kafka write errors (1/1)`), which is a class
`bugs_open/040` (kafka dial timeouts, OPEN) **already documents** — its 2026-08-15 section describes
the same `complete_workflow` failure. **Not a new bug, not filed as one; contributed into 040
instead.** One of the pair sat 4 h before the reaper failed it, and that is the reaper working as
designed, not this bug.

> **What I got wrong on the way, because it is the same trap this file warns about twice.** I read
> the pair as "one failed loudly, one failed **silently**" — the second being 343-shaped. It is not:
> `[MEASURED]` the two recorded the **same** failure in **different tables**, and `agent_error_log`
> and `orchestration_states.error` turn out to be **completely disjoint** for Kafka errors
> (**125** / **1** / **0** in both — not a retention artefact, identical over the common window). I
> had checked `orchestration_states.error` for both and **inferred silence from a single surface.**
>
> **The general form is stronger and is what to remember** (measured by the peer session, verified
> here): the sinks are **not** disjoint in general — **23,230** orchestrations have `agent_error_log`
> rows, **22** have `orchestration_states.error` set, **9** of those 22 in both. The Kafka zero is a
> property of that class. **What generalises is the denominator: `orchestration_states.error` is
> populated for ~0.1% of what the other sink covers, so it is not a rate instrument for anything.**
> Contributed to `bugs_open/040`, whose own "How to verify" points readers at that near-empty column.
> A per-instance "this one was silent" claim is unsafe unless both surfaces are read.

**One retention correction that CONFIRMS this file rather than undermining it.** `min(created_at)` on
`orchestration_states` reads **2026-07-19**, which looks like month-long retention and would reopen
every `[UNVERIFIED]` here that rests on "the rows are purged". It does not: **CANCELLED** rows (24)
appear never to be pruned, while `COMPLETED` and `FAILED` start ~26 h back. Checked in the direction
that mattered — **0 of 21** of the 08-17 wedged orchestrations survive there. **So the purge claim
holds, and the closed route above stays closed.**

**Do not treat quiet as evidence of a fix.** Six of the eight days around the 08-17 burst were also
zero, *before* anything was changed.

## Cold start

`docs/agent_docs/docs024_key_docs_latest/bugfix_029_retry_kills_live_child/HANDOFF_2026-08-19b_continue_here.md`
— the 2026-08-20 blocks. Detail in `NOTES_retry_kills_live_child.md` §§19–25 (§19 the control, §20
the refuted pod hypothesis, §21+§23 what Part A does and does not fix, §§24–25 the outage, confirmed
twice). The lane directory keeps its `bugfix_029_...` name; that is history, not a claim about which
bug it serves.

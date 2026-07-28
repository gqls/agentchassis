# 129 — the spawned child ADOPTS the parent's orchestration row and silently declines the work

> ## UPDATE 2026-07-28 ~22:15 — THE FIX IS **LIVE** on v1.0.1194, carried out by another session's roll. Still OPEN: the REPLAY half is not yet witnessed.
>
> I deliberately did not deploy (council veto, below). **It shipped anyway** — the
> fleet rolled to **v1.0.1194 at 20:48:11Z**, and `make build-<service>` builds from
> committed `HEAD`, so my 19:57Z and 20:16Z commits went out with someone else's
> build. Pod-grepped on **both** pods rather than inferred: `is_retry` **0** (the
> string this fix deletes), `RETRY_PAYLOAD_UNAVAILABLE` / `RETRY_SELF_ADDRESSED` /
> `MISROUTED_REQUEST` / `RETRY_PAYLOAD_BACKFILL_MISSED` all **1**.
>
> **CAPTURE HALF — PROVEN on live traffic.** Both wired actions recorded a faithful
> payload; the child's own id is stored, not the awaiting orchestration's:
>
> | step | awaiting orch | target orch | action | body | payload size |
> |---|---|---|---|---|---|
> | `call_scraper` (call_agent) | `b89b6e5e` | **`ef7e2ddb`** | `process` | 210 B | 1186 B |
> | `spawn_scraper` (spawn_agent) | `b89b6e5e` | **`8b5dc669`** | `initialize` | 447 B | 1124 B |
>
> The invariant query returns **0** self-addressed payloads, as it must. Payload
> size is ~1.1 KB, which retires the council's size risk with a number.
>
> **REPLAY HALF — NOT YET WITNESSED, and I am not calling it proven.**
> `Replaying original request` appears **0** times in 90 minutes of live logs — but
> so do `Retrying request` and all four new markers, i.e. **no timeout has occurred
> yet**, not "replays are failing". Induced run `c54b3fdf-b556-45f1-8e59-8237bec64d2a`
> (`TRIGGER_code_indexer_v2.sh`, the lane this file records as failing 2 of 3 on
> v1.0.1184): `spawn_indexer` **processed at retry_version 0** → `call_indexer`, and
> the child reached `EXECUTING_STEP / index_symbols`, i.e. step 2 of its 3-step
> workflow, so `request_analysis` ran. **That is the lane working — but it succeeded
> on the FIRST attempt, so it exercised the capture path and not the replay path.**
> The defect was bursty (2 of 3), so one green run is weak evidence on its own.
>
> **WHAT IS STILL OWED TO CLOSE THIS** — one witnessed replay:
> ```sql
> SELECT step_name, retry_version, status FROM awaited_requests
>  WHERE retry_version > 0 AND sent_at > '2026-07-28 20:48:11' ORDER BY sent_at DESC;
> ```
> ```bash
> kubectl logs -n ai-persona-system -l app=agent-chassis --since=6h \
>   | grep -E 'Replaying original request|RETRY_PAYLOAD_UNAVAILABLE|RETRY_SELF_ADDRESSED'
> ```
> A `Replaying original request` line whose `child_orchestration_id` differs from the
> awaiting orchestration, followed by that request reaching `processed`, closes it.
>
> ---
>
> ## ROOT CAUSE FOUND AND FIXED IN CODE — 2026-07-28 evening (bugsearch 4 thread).
>
> **The child was never the defect.** It was handed the parent's identity, by the
> coordinator's own retry path. `handleRecoverableError`
> (`platform/orchestration/coordinator.go`) did not resend the request that timed
> out — it **synthesised a new one out of the AWAITING orchestration's own state**:
>
> ```go
> OrchestrationID: state.OrchestrationID,   // the PARENT's id  → this file's swallow
> Action:          "execute",               // never the original action
> Body:            {"is_retry": true, …},   // the original payload, gone
> ```
>
> So all three of this file's fix candidates aim at the wrong end. Candidate 2 is
> still worth having and is included, but as defence in depth — **it does not fix
> the bug on its own** (the parent still times out; only the silence goes away).
>
> **Measured before writing anything** (live `clients_db`, 14 days): **430 of 430**
> retried `awaited_requests` took that path and **294 (68%) exhausted the budget**.
> All-history the distribution is 93 at `retry_version` 1, 45 at 2 and **294 at 3** —
> a retry that recovers decays; this one accumulates at the cap. The adapter
> re-execute branch eight lines above, whose comment already says *"adapters need the
> full payload"*, has been taken **zero** times.
> `[UNMEASURED]` how many of the 108 rows that ended `processed` at
> `retry_version` ≥ 1 were rescued by the retry rather than by a late original
> response — the two are indistinguishable here, so the durable claim is "68%
> exhausted", not "retries never work".
>
> **The fix — one invariant: a retry is a REPLAY of the original request**, differing
> in `retry_version`, `message_id` and `timestamp` and nothing else. Commit
> `eb70c3dd3` + follow-up; migration `263` (`awaited_requests.request_payload`,
> nullable) **applied and recorded**; chassis **v1.0.1193 built and pushed**.
> Concept register: **RSH-003** (`resilience-self-heal.md`).
>
> **WHY THIS STAYS OPEN — the bar is fixed AND live, and it is not live.**
> The council gate **REJECTED** it on **SCOPE** (`75cb2fdc-e74c-4d3d-99b7-9264548e65d6`,
> `decided_by: hard veto from guardian`, `unreadable: 0` — a judgement, not the
> harness). Six of ten seats approved and **no seat disputed the diagnosis**; the
> veto is about venue — a shared mechanism plus a schema column arriving inside a
> bug patch, the same finding as `bugs_closed/124`. Per the owner ruling of
> 2026-07-28 that is **not** answered by resubmitting, and the seats **contradict
> each other on the remedy** (the guardian's contained alternative is the child-side
> patch that `constitution` and `editquality` approved the plan for *not* treating as
> primary), so it needs a human. **The deploy is deliberately not done.**
>
> **COLD-START:** `docs/agent_docs/docs024_key_docs_latest/bugfix_129_retry_replay/`
> — `REVIEW_2026-07-28_council_scope_veto.md` first (the three options, costed),
> then `RUNBOOK_retry_replay.md` for the verification commands.
>
> **LANDMINE 1:** the live retry path is `coordinator.go handleRecoverableError`,
> **not** `helpers.go retryTimedOutRequest`. The latter has the identical defect and
> is fixed too, but it is dormant — grepping for the *mechanism* lands you there and
> you will fix nothing. Grep the **error string from the evidence**
> (`timed out after 3 retries`), which lands in `coordinator.go`.
> **LANDMINE 2:** the discriminating pod-grep is a string this fix **DELETED** —
> `is_retry`. v1.0.1192 has it and none of the new markers; v1.0.1193 is the
> reverse. The obvious positive markers alone are vacuous.
>
> **Residual, NOT fixed here and deliberately not bundled:** 6 of the 428 retried
> requests come from `scrape_web`/`web_search`, which put the **caller's own**
> orchestration id on the *original* outbound message
> (`web_search_action.go:139`), so there is no child identity to replay. They now
> fail fast rather than retry — 4 of the 6 exhausted anyway, so the delta is ≤2
> requests a fortnight. Own mechanism, own diagnosis.

**Filed 2026-07-28 (bugs thread 2).** Status: OPEN, unowned — but read the family notes
below before routing: the *measurement* discipline and prior refuted theory live in the
handshake record, and `platform/orchestration/coordinator.go` is being actively worked by
the work-item-parallelisation workstream (CS-1 landed there hours before this filing).

This is the **child-side mechanism** behind (at least part of) the known spawn→call
handshake failure — the family measured in `agent_error_log` as
`"Request <id> timed out after 3 retries"` (dominant sources `generic/call_dispatch`,
`build-pipeline-trigger/spawn_dispatch`; the 07-27 ordering-race theory for it was
loop-REFUTED, corr `eb8df254`, leaving "genuine non-response" as the cause). This file
shows what "genuine non-response" actually is on the child, with logs: **the request
arrives, the child processes it "successfully", and chooses to do nothing.**

## Symptom

`index-orchestrator` (spawn_indexer → call_indexer → complete) FAILED at `spawn_indexer`
twice consecutively, ~11:01 and ~11:10 on 2026-07-28, chassis `v1.0.1184`:

```
orchestration_states: FAILED / spawn_indexer
error: Request 91f55361-3ef3-457d-a1bd-5c583d9130c0 timed out after 3 retries
```

Both spawned pods came up healthy (1/1 Running, image v1.0.1184) and stayed idle.

## Evidence — the child's own log (captured before pod reaping; orchestrations 11a6e647, 58a53a6a)

The child received **all three retries** (`retry_version` 1, 2, 3 all logged), and for
each one walked the same path (pod `agent-code-indexer-25dc09ae-sjvbh`, 11:16:34, the
same lines for retries 1 and 2):

```
processor.go:285   "Workflow validated successfully, executing workflow"
processor.go:1748  "Executing workflow ... start_step: request_analysis, total_steps: 3"
coordinator.go:615 "Found existing state by orchestration_id"  orchestration_id=58a53a6a   <-- THE PARENT'S ROW
coordinator.go:153 "Orchestration state retrieved" is_new=false status=AWAITING_RESPONSES
coordinator.go:729 "Handling orchestration status" state.CurrentStep=spawn_indexer
coordinator.go:780 "Orchestration is awaiting responses in handleOrchestrationStatus" awaited_count=1
processor.go:293   "Workflow started and is now waiting for a response"                   <-- IT IS NOT
processor.go:1612  "ProcessMessage completed successfully"
```

`grep -c "fetched repo source"` on both pods' full logs: **0**. The child never ran
`request_analysis`, never fetched, never replied. The parent times out after 3 retries.

**The mechanism in one sentence:** the spawn request carries the PARENT's
`orchestration_id` (58a53a6a); the child's `SagaCoordinator.ExecuteWorkflow` loads state
by that id, finds the parent's row at `AWAITING_RESPONSES / spawn_indexer`,
and `handleOrchestrationStatus` concludes "already awaiting — nothing to execute" —
about a workflow (`request_analysis → index_symbols → complete`) that has never started.
The child's "success" log line and the parent's timeout describe the same event.

## What this is NOT (checked, same window)

- **Not a fleet-wide spawn break on v1.0.1184:** a spawned council (orchestration
  `32a1e2bc`, plan contains `spawn_agent`) completed with approval 11:12→11:17, between
  the two failures.
- **Not proven to be CS-1** (`afbd005f9`, staleness guard in the claim-recovery reset,
  same file, landed in the same roll). Attractive theory — pre-CS-1 CLAIM_RECOVERY
  stealing the parent's "live" claim could have been the accident that let children run —
  **but the hourly `agent_error_log` counts refute the clean version of it**: the 10:00
  hour's 15 `spawn_dispatch` timeouts ran on `v1.0.1182`, built BEFORE CS-1 was
  committed (09:52:37 UTC), and the family long predates CS-1 (07-27: 25/12, 23/13).
  `[HYPOTHESIS, UNTESTED]`: CS-1 may still have *changed the odds* on this branch;
  nothing here measures that.
- **Not the persist-ordering race** — that theory was already refuted by the loop on
  07-27 (`persistAwaitingStateWithRetry` re-loads and returns early when a response
  already arrived; verified citations in the handshake record).

## Why it matters

- The 24h `code-index-refresh` cadence goes through this exact lane — when this fires,
  the index refresh **fails silently from the cadence's point of view** (the scheduled
  task's `last_completed_at` still updates on trigger, not on outcome).
- The diagnosis loop and feature-builder spawn through the same coordinator path — the
  two prior `needs_diagnosis` filings for this family (SpawnAgentAction 07-27,
  ProcessResponse 07-20) are both `status='failed'`: **the defect eats the diagnosis
  runs sent to diagnose it.**
- It blocked the live verification of `bugs_open/108`'s fix for ~an hour (three manual
  dispatches to get one reindex through).

## Fix candidates — ordered by what makes the bad state unrepresentable

1. **A spawned child executing its own workflow must never satisfy itself with the
   PARENT's state row.** Key the child's state by its own orchestration identity (or
   namespace the row by executing agent), so "found existing state" can only ever find
   the child's own progress. Makes the swallow structurally impossible.
2. **`handleOrchestrationStatus` must not treat `AWAITING_RESPONSES` as "nothing to do"
   when the awaited request is THE ONE IT IS CURRENTLY PROCESSING.** The child is
   holding request `052f24bb` — the very request the parent's row is awaiting; declining
   it because the row says "awaiting" is self-referential. Narrower than 1.
3. **Reply something.** Even if the child declines, an explicit decline-response would
   convert a 6-minute-3-retry timeout into an immediate, diagnosable failure. Does not
   fix the defect; ends the silence.

## Burstiness confirmed on the same image, same hour

The THIRD consecutive dispatch (corr `7e89536a`, ~11:20) went straight through:
`request_analysis` fetched and analysed, `index_symbols` ran. Same image
(`v1.0.1184`), same lane, same input — so the swallow is **bursty, not
deterministic**, consistent with the family's history (25/12 → 24/1 → 23/13 windows).
2-of-3 failure rate for this lane today. Whatever gates the
`Found existing state → decline` branch varies per run; that variance is a diagnosis
lead in itself (timing of the parent's state write vs the child's first consume?
which pod consumed the earlier retries?).

## How to verify a fix

- Induced: dispatch `index-orchestrator` (TRIGGER_code_indexer_v2.sh) — currently fails
  ~consistently on v1.0.1184 within ~8 min; after a fix, `spawn_indexer → call_indexer`
  must advance and the child's log must show `fetched repo source`.
- The free reproducer for the wider family remains `build-pipeline-trigger` (every 30s,
  bursty): measure in `agent_error_log`, never `orchestration_states` (which showed
  166 COMPLETED / 0 FAILED while `agent_error_log` held 79 timeouts).

## Evidence preservation

Full pod logs from both failures are ephemeral (job cleanup). Salient excerpts above are
the durable copy; the failing orchestration rows `11a6e647-37c1-4d22-b081-e790c22abbb3`
and `58a53a6a-18f0-4aa0-878b-4ecc727c3f94` were **deliberately not cancelled** so a
diagnosis run can read them (the 07-27 lesson: cancel destroyed the evidence).

---

## 2026-07-28 12:4x — diagnosis `dcde1ed9` returned CONFIRMED, with an honest corroboration boundary

The loop CONFIRMED the timeout mechanism from the state tier: both orchestrations'
own `spawn_indexer` `awaited_requests` rows show `status=error` with a ~2-minute
sent→timeout window ("consistent with the request itself timing out in the
spawn/await-response completion path"). Its `next_scope` matches this file's open
questions (child execution/response for requests `91f55361`/`052f24bb`; whether those
rows ever went through a CLAIM_RECOVERY claim/reset cycle; spawn_agent timeout/retry
handling).

**What it could NOT corroborate, and why that is not a refutation:** the child-side
observations ("loads the parent's row", "declines while logging success") are marked
`explained: false` because *"no code-indexer / child pod log evidence is present in
this bundle"* — the pods were reaped, and this file is markdown, which the code tier
cannot read (the exact 108-residual). The log excerpts in this file are the only
surviving copy of that evidence; they were captured directly from both pods before
reaping. One genuine correction from the verdict: rows read AFTER the timeout show
`status=FAILED` (the child's later deliveries would hit the `StatusFailed` "Workflow
previously failed" branch); the `StatusAwaitingResponses` decline captured in the logs
happened DURING the 2-minute window. Both branches decline; neither replies.

**Adjacent and possibly curative:** the CS-3a work (`f4d24252f`, `353e98781` — chassis
response-consumer seed-to-latest, "response replay closed", 0/5 → 4/5 in 47s) shipped
the same day and targets response deafness after restarts. If the child DOES reply in
some fraction of cases and the reply goes unconsumed, that is the response-side half
of this family. Env-gated, default unchanged — check whether it is enabled before
measuring this bug again.

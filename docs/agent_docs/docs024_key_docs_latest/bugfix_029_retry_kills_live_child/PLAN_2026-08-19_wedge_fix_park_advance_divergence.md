# PLAN — 2026-08-19 — `bugs_open/029` wedge: the park/advance divergence

**Provenance.** Drafted by a Fable subagent on this session's brief, then **graded against the live
system by this session before being recorded.** Per MEMORY `a-subagent-report-is-another-doc`, a
subagent's report has no seam showing where measuring stopped — so the grading below is not
ceremony. **Two of its claims did not survive; both are corrected in place and neither changes the
plan's shape.**

**Status: PLAN ONLY.** No code written, nothing submitted to the council, nothing committed but this
document. Prepared alongside 090 run `d52c3407-14e7-4b9e-be46-c8ee741b2532`, whose verdict is now in
(NOTES §11: `UNVERIFIABLE`, a real abstention). §7 states what is contingent on a confirmation and
what is not.

---

## §0a — GRADING of the draft (this session, 2026-08-19 ~16:4xZ)

| draft claim | grade | what the check said |
|---|---|---|
| the three source sites (`coordinator.go:2113` / `:2671` / `:848`) | **HOLDS** | independently re-read; the draft additionally found `state.Status` is set to `StatusAwaitingResponses` in memory at `:2017` *before* the park, so the skip path leaves both `executeStep` and `continueExecution` reading "parked" from memory — a sharper statement of why the freeze is immediate and silent than this lane had |
| the retry path cannot mint a second rv0 row | **HOLDS, wrong citation** | it cited `state.go:1962` `UpdateAwaitedRequestForRetry`, which has **zero callers**. The live writer is `UpdateAwaitedRequestRetry`, `coordinator.go:3337`. Conclusion right, evidence dead — and the correct proof is stronger (only two `INSERT INTO awaited_requests` sites exist, neither on a retry path). NOTES §11a |
| **"live divergence census, both directions: 0 and 0 `[MEASURED]`" offered as a noise baseline for detection** | **❌ VACUOUS — do not carry it into a submission** | re-run here: **there are ZERO orchestrations in `AWAITING_RESPONSES` and ZERO `awaited_requests` rows in any non-terminal status fleet-wide.** `orchestration_states` holds 2919 COMPLETED / 47 FAILED / 24 CANCELLED and nothing else; `awaited_requests` holds 30986 processed / 399 expired / 128 error. **The denominator is empty, so 0/0 was arithmetically forced and could not have come out otherwise.** That is the `WRONG_CALLS` 2026-08-03 shape exactly. A real baseline has to be taken over a window with live parked orchestrations in it |
| `await_reconcile` consumer count = 0 | **true but tautological** | it is a name the draft invented; a new field necessarily has zero consumers. The RFC_022 conditions that do work are the other two (opt-in; unsafe side default OFF). State it that way in the submission rather than dressing a tautology as a measurement |
| duplicate-pair timing: same `processing_pod` both rows, gap 489–563 s | **[UNVERIFIED by this session]** — plausible and matches the lane's 414–577 s, but I did not re-run it. Marked so a reader does not inherit it as checked |

---

## §1 — The narrow correctness fix: key the arrival check on the request id

**The id is not recoverable today.** `applyResponseToState` (`coordinator.go:2784`) records `response`,
`response_received_at`, `response_status`, and `initialized` for spawns. **Nothing records which
request was answered.** `response_received_at` is a time, not an identity.

**Minimum addition:** `applyResponseToState` already receives `awaitedReq *AwaitedRequest`, which
carries `RequestID` (`state.go:67`). Write one sibling key alongside the reply in **all three**
branches (nil-guard `awaitedReq` — it can be nil on the dynamic-step path, `:2828`):

```go
existingData[awaitedResponseIDMarker] = awaitedReq.RequestID   // new const beside awaitedResponseMarker (:2179)
```

Then `:2114` becomes: marker present **and** recorded id == the reqID being registered → a reply to
THIS request beat the park → skip (today's safe back-off, consumer owns the continuation). Marker
present with a **different** id → stale residue of an earlier answered request on the same step name
→ **park normally.** That single discrimination closes the re-registration mechanism.

**Legacy rows:** a marker written pre-roll carries no id. Explicit branch — id key absent → treat as
arrived (today's exact behaviour), dated comment, dead once no pre-roll orchestration is live.

**Two findings from the draft worth keeping:**
- The genuine beat-the-park race is **spawn-shaped**: `preRegisterAwaitedRequest`
  (`spawn_actions.go:115,160`) inserts the row *before* the send, so a fast ack can be handled while
  the park is still persisting. On paths with no pre-registration the consumer cannot claim, so any
  marker hit there is stale **by construction**.
- The arrival check is also **blind where it should see**: `applyResponseToState`'s `output_mapping`
  branch (`:2795-2821`) replaces `CollectedData[stepName]` wholesale and **never writes the marker**.
  `[UNVERIFIED]` whether any live spawn/call step uses `output_mapping` — grep before implementing.
- `withoutResponseMarker` (`:2240+`) strips the marker on the carry path; it must strip the id too.

## §2 — The structural fix, ranked by what makes the bad state unrepresentable

The claim stands: two representations of "what is outstanding" — the `awaited_requests` table (drives
the response consumer) and the `AwaitedRequests` JSONB map (drives the advance decision) — **with no
reconciliation anywhere.**

| rank | option | why it ranks here | mutation proof |
|---|---|---|---|
| **1** | **D — typed park outcome.** `persistAwaitingStateWithRetry` returns `(parkOutcome, error)`, ∈ {`parkPersisted`, `parkSkippedReplyArrived`}; caller must switch. On skip: no `InsertAwaitedRequest` (`:2047`), no timeout goroutine | Makes "returned success without persisting" **unrepresentable at the type level** — the compiler forces every caller to say what it does. Kills the orphan-row producer outright. No new authority, no default-path behaviour change | same-id marker → assert **no** table row and no timeout armed; stale marker → assert entry and status **are** persisted. Reverting the signature fails compilation |
| **2** | **B — reconcile at the decision point**, split. *Detection* (unconditional): cross-check `allDone` against a new `CountOutstandingAwaitedRequests`; any disagreement → CRITICAL + metric + breadcrumb. *Enforcement* (opt-in, default OFF): map says done, table disagrees → adopt and re-park | Detection changes no decision; it makes divergence visible **at the moment it becomes a wrong decision**, which is what `:2671` hides today. **Two real hazards, both must be stated in the submission:** (a) a naive table-only `allDone` has a mutual-back-off race — two replies on two pods each see the other's row `'processing'` and neither advances; the map's optimistic-lock CAS is what serialises this today, so the table is a cross-check, **never a naive replacement**; (b) adopting a bogus `'waiting'` row means waiting out its timeout and the retry driver then **replays the request — real side effects re-run** | mutate detector and enforcer **separately** (MEMORY: a mutation that passes may have hit a guard in series); a third test asserts field OFF advances exactly as today |
| **3** | **C — periodic reconciler**, extending RSH-011 `wedge-evidence-capture` | Catches **every** producer including ones not yet found — the only option robust to a new mechanism. Lagging, detection-only. Not `platform/`, so a separate small task outside the council round | must write ONE row per run **including clean runs** (the RFC_022 cron precedent: a missing row must not read as "nothing is wrong"); induce, then delete, per RSH-011 |
| **4** | **A — table as single source of truth** | The only option that makes the dual representation itself unrepresentable, and the right eventual destination — but it re-engineers the advance decision's concurrency control while the divergence topology is unconfirmed. **Defer, and route on its own merits, not inside this bug** (the 124 lesson: a seam rework inside a bug patch draws a scope veto) | not designable until a burst is captured |

**Recommendation: §1 + D + B's detection half in one council round; B's enforcement opt-in in the
same diff, default OFF; C separately; A deferred.**

### Two further divergence producers the draft found — record, do not fix in this round

- **P2** `[VERIFIED as a path; UNVERIFIED as ever having fired]`: `skipToNextLoopIterationForAsync`
  (`loop_error_handler.go:243-260`) marks the table row terminal and deletes the map entry **in
  memory**, then `skipToNextLoopIteration` persists with a single **non-retrying** `UpdateState`
  (`:184-187`). An optimistic-lock failure there loses the delete-plus-advance while the row is
  already terminal; the reply's redelivery is then eaten by `DUPLICATE_SKIPPED` (`coordinator.go:378`,
  keyed on `processed_at`). **A lost continuation on exactly the error path the wedge enters
  through.** Own 090 symptom, not a passenger edit.
- **P3**: two "complete" writers with different guards on one table — `MarkAwaitedRequestComplete`
  (`coordinator.go:4399`, unconditional, can flip `'error'`→`'processed'`) vs `CompleteAwaitedRequest`
  (`state.go:1929`, guarded `status='processing'`). Register-entry material.

## §3 — Why the step executed twice at rv0: candidates, not a cause

`StuckOrchestrationTimeout = 5 * time.Minute` (`coordinator.go:38`).

| candidate | prediction | testable from `awaited_requests` alone? |
|---|---|---|
| **C1 stuck-orchestration takeover** (`handleOrchestrationStatus`, `StatusExecutingStep` arm, `:761-775`): incoming message + `LastActivity` > 5 min → `ClearExecutingStep` → `continueExecution` re-runs `CurrentStep` — fresh execution, hence rv0 and a new request id | gap = 300 s + trigger cadence | **Yes** — 17/17 gaps must exceed 300 s; any pair under it refutes |
| **C2 `StatusRunning` arm** (`:782-801`) | identical table signature | **No** — needs logs (pods two days gone) or an RSH-011 capture |
| **C3 retry driver** | bumps in place, never a second row | **REFUTED at source**, NOTES §11a |
| **C4 Kafka redelivery** | duplicate guard skips on `processed_at` without re-running the body | **Disfavoured** — row 1 is `processed` in all pairs |
| **C5 broken park interlock double-drive** | seconds-scale gap, same goroutine | **Disfavoured for these 17** by the 8–9 min gaps, but a real latent consequence of §1's defect |

**What no table query can answer, and what this plan does not claim:** the **first** death — reply 1
was handled and the parent never registered `iter_{N+1}_call_handler`. NOTES §6 is right that the
arrival-check defect does **not** fit as the first failure (spawn #1's park has no prior marker to
trip on). This plan closes the second-wedge mechanism and the recovery path. It does not explain the
first death, and no edit here should be described as if it did.

## §4 — Owner-ruling compliance

- **What carries new authority:** §1, D, and B's detection half carry **none** — no field, no changed
  decision on the default path. Per 2026-07-29 §1, restoring "a park persists the awaited entry" is
  **repairing** the stated guarantee, not altering it. **B's enforcement half is new authority on the
  shared park/response seam**, so per 2026-08-02 §2 it ships as an opt-in workflow-plan key
  (`await_reconcile_enforce`, bool) with **default false = today's behaviour; the unsafe side OFF.**
- **RFC_022 as narrowed:** opt-in ✓, unsafe default OFF ✓, zero live consumers ✓ — **but say plainly
  in the submission that the third is tautological for a newly-named field** (§0a). → **not
  architecture-scope; council gate, no RFC.** As a workflow-plan key it does not enter the WFA-013
  optional-key budget; **if implementation moves it onto an action's config it does**, and the cron
  parity test (`cmd/config-key-audit/optional_budget_cron_parity_test.go`) applies.
- **2026-07-29 §3 — tell the consumers, don't just measure them.** Consumers are every workflow with
  `await_response` steps. The message that matters: *a park that could previously fail silently to
  persist now always persists; a re-registration of an already-answered step name now genuinely parks
  instead of no-opping.* Name `build-dispatch-loop` explicitly.
- **Ordering-exemption condition (2):** register entry **in the same commit** (arrival-check contract,
  the marker/id pair, the opt-in field, P3's drift), plus a LANDMINES entry for the legacy-marker
  compat branch — a session seeing "no id key reads as arrived" must not "fix" it to stale.

## §5 — Council submission sketch (NOT submitted)

**rationale:** *`bugs_open/029`'s wedge: the outstanding-awaited set has two representations that
nothing reconciles — the `awaited_requests` table (drives the response consumer) and the
`AwaitedRequests` JSONB (drives the advance decision). A defect verified at source
(`coordinator.go:2114`: the beat-the-park check is keyed on step name, not request id, and returns
success without persisting) manufactures divergence on every re-registration of an answered step
name — the shape of 17 of 20 wedged instances from 2026-08-17, measured in `awaited_requests` and
reproduced independently, and read back by diagnosis run `d52c3407` which cited the rows. This round:
record the request id where the reply is recorded; key the check on it; make a non-persisting park
outcome unrepresentable via a typed result; make divergence visible at the decision it corrupts, with
enforcement behind an opt-in key, default OFF. The wedge's FIRST-death mechanism is not explained by
this change and nothing here asserts it.*

| # | file | operation | sketch |
|---|---|---|---|
| 1 | `coordinator.go` | edit `applyResponseToState`, all three branches | `existingData[awaitedResponseIDMarker] = awaitedReq.RequestID`, nil-guarded; new const beside `:2182`; also write it in `output_mapping` + default branches, which write no marker today |
| 2 | `coordinator.go` | edit arrival check `:2114-2122` | hit requires marker AND id == reqID; different id → park; no id → dated legacy branch preserving today |
| 3 | `coordinator.go` | `persistAwaitingStateWithRetry` → `(parkOutcome, error)`; update `processAwaitResponse` `:2031-2047` | on skip: no insert, no timeout goroutine, report handled |
| 4 | `coordinator.go` | `withoutResponseMarker` / `carryCollectedDataOntoFreshState` `:2189-2260` | strip both keys; update doc comment |
| 5 | `state.go` | add `CountOutstandingAwaitedRequests(ctx, orchID, excludeReqID)` | `status IN ('waiting','processing','retrying')` |
| 6 | `coordinator.go` | edit `handleCompleteResponse` after `:2675` | detection unconditional; enforcement only under `await_reconcile_enforce`; default path byte-identical |
| 7 | `park_arrival_check_test.go` (new) | tests, mutation-provable | stale → parks; same-id → skip, no orphan row; legacy → compat; detector and enforcer mutated separately |
| 8 | register + `LANDMINES.md` | same commit | seam, marker/id pair, opt-in field, P3 drift; landmine for the legacy branch |

## §6 — What could make this plan wrong

| claim | disconfirming observation |
|---|---|
| the arrival-check defect explains the second wedge | an RSH-011-captured wedge showing **no marker** under the spawn step name at re-park time. **Retrospective check impossible** (`orchestration_states` purged) — prospective only, hence `[INFERRED]` for the 08-17 cohort |
| the beat-the-park race is spawn-only | any call-path action found pre-registering (grep `preRegisterAwaitedRequest` callers at implementation time) |
| legacy no-id markers are safe to treat as arrived | a post-roll wedge whose park skipped on a legacy marker |
| C1 (5-min takeover) fits the doubles | any of the 17 gaps under 300 s — **checkable now**, and not yet done |
| detection starts quiet | unknown: **the 0/0 census was vacuous** (§0a). The park writes JSONB *before* the insert (`:2031`→`:2047`), so a benign window exists in one direction; if noise appears the query needs a grace age, not deletion |

## §7 — Sequencing

**Worth doing regardless of any verdict** — these are defects verified at source on their own terms:
edits 1–5, 7, 8, plus detection (6's unconditional half). If a future capture confirms a different
first-death mechanism, none of this becomes *wrong*; it becomes *insufficient*, which it already
admits to being.

**Contingent on a confirmed topology:** turning `await_reconcile_enforce` on for any workflow; Option
A; and any sentence claiming the wedge is *explained*.

**Not in this round:** P2 (own 090 symptom — one coherent bug per run); Option C (CronJob, separate);
the `### awaited_requests` bundle-rows gap (NOTES §10 — **different subsystem, its own council
round**, and arguably higher value than this plan because it unblocks every future orchestration-hang
diagnosis).

**029 stays OPEN throughout.** The bar is fixed AND live; the first death is neither.

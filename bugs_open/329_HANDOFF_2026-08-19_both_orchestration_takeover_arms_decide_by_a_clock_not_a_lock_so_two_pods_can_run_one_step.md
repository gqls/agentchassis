# 329 — both orchestration takeover arms decide by a CLOCK, not a lock, so two pods can run the same step

**Filed 2026-08-19** by the `orchestration_status_lifecycle` lane (`bugs_closed/294` → `310` → the
class fix). Status: **OPEN, UNOWNED.**

**Raised by the council, not by a symptom.** The `guardian` seat objected at *medium* on migration
`465`'s round (`1c212b15`, APPROVED) and was right. Filing it because an objection that lives only
inside a council verdict is effectively invisible — nobody greps `diagnosis_artifacts`.

**No known incident.** This heuristic has run for months and predates the change that surfaced it.
It is a correctness gap, not a fire.

---

## The one-paragraph version

When a pod sees an orchestration that looks stuck, it takes it over and resumes it. It decides
"stuck" by a **clock** — `time.Since(state.LastActivity) > StuckOrchestrationTimeout` (5 minutes) —
and then simply proceeds. Nothing claims the row. If the original pod is in fact still working and
merely has not refreshed `last_activity`, **two pods run the same step**, with whatever external
side effects that step has, twice. The estate already has the right primitive —
`TakeOverOrchestration` is a guarded compare-and-swap built for exactly this shape — and neither arm
uses it.

## The two arms

| arm | location | gate |
|---|---|---|
| `EXECUTING_STEP` takeover (pre-existing) | `coordinator.go` `case StatusExecutingStep` | `state.CurrentlyExecuting != nil && time.Since(state.LastActivity) > StuckOrchestrationTimeout` |
| `RUNNING` takeover (added by `465`) | `coordinator.go` `case StatusRunning` | `time.Since(state.LastActivity) > StuckOrchestrationTimeout` |

**The second deliberately mirrors the first.** `465` reused the estate's existing takeover guard
rather than inventing a weaker one, so this bug is **not a regression** — it is the pre-existing
shape, now present twice and therefore worth naming.

## Why there is no lock today `[MEASURED 2026-08-19]`

- `SetExecutingStep` (`state.go:1346`) ends in `r.UpdateState(ctx, state)` — **not**
  `UpdateStateWithVersion`. So the write that marks a step executing does **no** version check.
- `ClearExecutingStep` (`state.go:~1428`), the pre-existing takeover's own write, likewise ends in
  `r.UpdateState(ctx, state)`.
- `UpdateStateWithVersion` **exists** (`state.go:961`) and is used elsewhere, so optimistic locking
  is available and simply not on this path.
- `TakeOverOrchestration` (`state.go:1366+`) is a **guarded CAS**: it re-stamps `processing_node`
  from a named previous holder and returns whether the handover was won. Its own doc comment
  describes it as the fix for `bugs_open/075`.

⚠ **Do not build an ownership gate on `processing_node` naively.** `SetExecutingStep` assigns it but
`UpdateStateWithVersion`'s UPDATE does not list the column, so the assignment **never reaches the
database** (comment at `state.go:1354-1360`, `bugs_open/075`). The column therefore records the pod
that *created* the row, not the pod driving it. `TakeOverOrchestration` is the only writer that
moves it after creation — which is precisely why it, and not a hand-rolled check, is the primitive
to use.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Route both arms through `TakeOverOrchestration`.** A takeover becomes *claimed* rather than
   *assumed*: win the CAS, then resume; lose it, and return without acting. Uses the primitive that
   already exists and closes the door for both arms symmetrically.
2. **Version the write instead** — have `SetExecutingStep`/`ClearExecutingStep` use
   `UpdateStateWithVersion`, so a second pod's takeover fails the optimistic-lock check. Smaller
   diff, but it converts a race into a retry rather than into a refusal, and the existing
   `ExecuteWithOptimisticLocking` retry loop would then re-drive it.
3. **Leave the heuristic and widen the threshold.** Rejected as a fix — it changes the odds, not the
   representability, and the estate's own rule is to rank by what closes the door.

## How to verify a fix

- **Induce the race:** two concurrent callers against one orchestration whose `last_activity` is
  older than `StuckOrchestrationTimeout`. Exactly one must proceed; the other must return without
  executing the step. **The disconfirming result is that both proceed** — which is what today's code
  does, so run it against the unfixed path first and watch it fail.
- **Negative control in the same run:** a fresh row (`last_activity = NOW()`) must be taken over by
  *neither*, or a fix that refuses everything passes the test above.
- **Do not verify by reading the function** — this lane got the same class of thing wrong three
  times in two days by reading a body instead of its callers. Walk the caller graph, and confirm at
  the artefact.

## Notes for whoever takes it

- **Core dispatch plumbing.** The council's own history shows `coordinator.go` deflected upward six
  times in seven days; expect the architecture seat to look hard at anything beyond a point fix.
- **Needs a roll**, unlike this lane's other work — it is Go, inert until the next build.
- Related: `bugs_open/075` (why `processing_node` is inert), `bugs_closed/294`, migration `465`,
  council `1c212b15` for the objection verbatim, and
  `docs024_key_docs_latest/orchestration_status_lifecycle/` for the lane's runbook and notes.

---

# UPDATE 2026-09-03 — taken up by the `bugfix_329_takeover_claim` lane; still valid, and the reachability story is materially different from the one above

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_329_takeover_claim/`. Taken as an **unowned**
bug on the evidence this file's filing lane left: `orchestration_status_lifecycle/HANDOFF_2026-08-19_continue_here.md:10`
disclaims it in writing, no live session names it, and that lane's directory has been cold since
2026-08-24.

## 1. The mechanism above is unchanged — re-read at the code, not grepped

`coordinator.go:758` (`StatusExecutingStep`) and `:780` (`StatusRunning`) both still gate on
`time.Since(state.LastActivity) > StuckOrchestrationTimeout` (`:38`, 300 s) and then proceed with
nothing claiming the row. `TakeOverOrchestration` still has exactly **one** caller — `:290`, the
response-routing path, which deliberately proceeds win or lose. Neither arm uses it.

## 2. Three facts this file predates, all measured at the live system

**(a) The race is physically possible now.** `agent-chassis` is **2/2 replicas** on `v1.0.1356`
`[MEASURED 2026-09-03 ~11:0xZ]`. This file states no replica count, so a reader could reasonably have
assumed one pod and closed it.

**(b) There is a guard IN SERIES that this file never mentions — and it would have passed a naive
test of the fix.** `agent-chassis` runs `CHASSIS_INTAKE_MODE=worker_pool_all`. Under that mode
(`platform/agentbase/intake.go`, `intake_workers.go`) messages are persisted and executed by a
claim-worker pool that first acquires the message's **serialisation key** — `intakeSerialisationKey`
(intake.go:215) derives it as the **orchestration_id** for requests and as the **parent**
orchestration for responses — via `ClaimSerialisationKey` (intake_repo.go:136), an
`INSERT … ON CONFLICT DO UPDATE … WHERE lease_expires_at <= NOW()` CAS on a table both pods share.
That is one holder per orchestration **fleet-wide**. ⚠ **A "two callers, exactly one proceeds" test
run on the chassis can therefore pass with the fix reverted.**

**(c) It is a THIRD guard, not a second, and the outermost two do not cover the estate.** With
thanks to the `dispatch_throughput` lane, who supplied the one I had missed. Layers, outermost in:

| # | guard | where it applies |
|---|---|---|
| i | intake serialisation claim | `agent-chassis` **only** |
| ii | **the coordinator arms — this defect** | every agent binary |
| iii | work-item claim CAS (`claim_work_item_action.go`, conditional UPDATE valid only while `triaged`/`approved`) | the dispatch path |

## 3. Why the fix belongs in the coordinator: the guard covers a MINORITY of the drivers

`SagaCoordinator` is constructed in `platform/agentbase/agent.go`, so both arms run in **every** agent
binary. `CHASSIS_INTAKE_MODE` is set on `agent-chassis` **only** — and `intake.go` disables the mode
for **spawned** pods structurally (the `a.spawned` guard plus the `system.agent.` topic-prefix check).

Who actually creates orchestrations, `orchestration_states.processing_node` over 14 d
`[MEASURED 2026-09-03 ~11:3xZ]`, pod-family suffix stripped: `agent-chassis` **3,332** ·
`agent-page-rerender` **2,215** · `agent-build-dispatch-loop` **660** · `agent-page-build-handler`
**412** · `agent-internal-link-resolver` **392** · `agent-page-content-writer` **392** ·
`agent-asset-deployer` **236** · 9 more families 82–142 each. **Every family below the first is a
spawned pod, not a Deployment** (checked against `kubectl get deploy`; 8 `agent-page-rerender` pods
were alive at the time of the reading). So the intake claim covers **under 40%** of the processes that
drive orchestrations, and the remainder run the same two clock-only arms with no serialisation claim
of any kind.

⚠ This measurement could have come out `agent-chassis: 100%`, which would have pushed the fix toward
the intake layer instead. It did not.

## 4. Blast radius on the busiest path: **clean**, and that narrows the case honestly

The `dispatch_throughput` lane's DOUBLE-HANDLE CENSUS (their `RUNBOOK_dispatch_throughput.md`,
§"Concurrency meters that actually measure concurrency"), 24 h `[MEASURED 2026-09-03 ~11:4xZ]`:
**3,044 handler orchestrations · 2,911 distinct work items · 71 items with ≥2 handlers · 0
overlapping pairs on one item.** The 71 are sequential retries. Also **0** rows with a repeated step
name inside one `execution_path` over 7 d `[MEASURED 2026-09-03 ~11:0xZ]`.

⚠ Their caveat, which cuts against their own lane: that census bounds this bug **only where a CAS
exists**; on a path with external side effects and no claim it says nothing. ⚠ And on a re-run, a
stale-reaped handler's `updated_at` is the REAP stamp, not end-of-life, so a legitimate successor
re-claim inside the reap window reads as an overlapping pair (discriminator: status FAILED +
`error LIKE 'Orchestration stale%'` on the first-started member, second started minutes not seconds
later).

**So the justification is NOT "stop an active fire", and should not be argued as one.** It is: where a
path sits behind a CAS the defect is absorbed; where it does not, nothing stands between a
five-minute clock and a double execution with external side effects. The fix's value is that it stops
depending on a backstop that is not present on every path, that nobody chose as this defect's
mitigation, and that no one maintains as such — guard (iii) exists to make work-item claiming
exclusive, and its protection here is a **by-product** a future refactor could remove without knowing
it was load-bearing.

## 5. A finding this lane owns, because no one else does

`intakeLeaseDefault` is **180 s** (`intake_workers.go:43`) against a **300 s**
`StuckOrchestrationTimeout`. **180 < 300**, so a serialisation key can change hands *before* the row
is old enough to look stuck — the handover window and the takeover window are adjacent, not exclusive.
And `drainKey` only tests `claimLost` **between events**, while `processMessage` takes no context
(the file header says so), so after a handover the old holder **finishes the event it is already
inside** while the new holder starts the next one on the same orchestration — a window bounded by the
heartbeat period (`lease/3` = 60 s).

The `dispatch_throughput` lane was offered this and declined it on the correct ground: their
concurrency ground is the **work-item claim** seam, they have never measured `intakeLeaseDefault`, and
they have **no evidence the ordering is deliberate**. Recorded here so a later reader does not
attribute it to them. **[UNVERIFIED]** whether the ordering was ever chosen rather than defaulted.

## 6. Notes for whoever reviews the fix

- **Do not verify on the dispatch path.** Guard (iii) absorbs a double-takeover there even with (i)
  and the fix both removed, so a pass proves nothing. Pick a locus with no CAS behind it and say why
  it has none.
- **No `090` run, and here is why rather than merely that** (owner ruling 2026-07-31): `LANDMINES.md`
  records that a run on a symbol in a file over ~60 KB returns bundles and no verdict, looking exactly
  like a run still in progress. `coordinator.go` is **199,136 bytes**, `state.go` **77,392**
  `[MEASURED 2026-09-03]`. Substituted three artefact-level reads — both arms in full, the live
  deployment env and replica counts, and the intake claim path end to end.
- **`processing_node` intra-pod caveat for candidate (1).** The CAS is `WHERE processing_node = $3`.
  Where the observed value already equals the acting pod's own name, two actors in that pod would both
  match and both report `rowsAffected = 1`. Whether that is reachable depends on the claim layers
  above; it must be stated either way, not assumed away.
- `claim_work_item_action.go` is **another lane's seam** (it now carries an opt-in
  `honour_spend_governor` pre-claim read, default OFF, live on `build-dispatch-loop` only). Not to be
  edited from here; they have asked to be told if the claim primitive the arms use changes.

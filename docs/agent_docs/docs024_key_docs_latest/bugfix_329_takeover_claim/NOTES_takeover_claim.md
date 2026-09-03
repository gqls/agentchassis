# NOTES — bugfix 329, the takeover arms decide by a clock, not a lock

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## (a) 2026-09-03 ~10:5xZ — picking the bug up, and why this one

Asked to pick up an **unowned** bug. Method, so the next thread can repeat it rather than
re-derive it:

- `scripts/who-owns.py <number>` is **useless as a batch filter here** and I nearly drew a false
  conclusion from it. Ran it over all 166 `bugs_open/` files looking for its "no activity … and no
  owning workstream" verdict: **zero hits.** That does not mean every bug is owned — it means a
  bare 3-digit number matches prose in dozens of unrelated lane docs, so the "mentions" heuristic
  finds an owner for everything. `[MEASURED 2026-09-03]` The tool is built for the single-bug
  question it documents ("before routing work AT an existing bug"), and I misused it as a census.
  What actually worked: last-commit age on the bug FILE + the existence and age of a
  `bugfix_<n>_*` lane dir + the live session list from `ListAgents`.
- **`ListAgents` is the check `who-owns.py` cannot do**, and CLAUDE.md says so ("it reads COMMITS,
  so a session mid-fix is invisible"). 160 peer sessions; 14 name a bug number. That is how I
  confirmed `bugs_open/450` was live (session `[6a8285]`, busy, commits 9 minutes old) and left it.

329 was chosen because the lane that FILED it disclaims it in writing —
`orchestration_status_lifecycle/HANDOFF_2026-08-19_continue_here.md:10`: *"not this lane's work:
… `bugs_open/329` is filed with no [owner]"* — no live session names it, its lane dir has not been
touched since 2026-08-24, and it is core plumbing shared by every agent binary, which is the
"framework as a whole" shape the owner asks for.

## (b) 2026-09-03 ~11:0xZ — the bug is still valid, verified at the code

`platform/orchestration/coordinator.go`, the switch in `handleOrchestrationStatus`:

- `case StatusExecutingStep` (**line 758**): `if state.CurrentlyExecuting != nil &&
  time.Since(state.LastActivity) > StuckOrchestrationTimeout` → `ClearExecutingStep` → reload →
  `continueExecution`. Nothing claims the row.
- `case StatusRunning` (**line 780**): `if time.Since(state.LastActivity) >
  StuckOrchestrationTimeout` → `continueExecution`. Likewise.
- `StuckOrchestrationTimeout = 5 * time.Minute` (**line 38**).

`TakeOverOrchestration` has exactly **one** caller (**coordinator.go:290**), and it is the
response-routing path, which deliberately proceeds whether or not it wins the CAS. Neither
takeover arm uses it. So the bug file's mechanism reads true today.

Tree state checked before starting: `platform/orchestration/coordinator.go` and `state.go` are
both **clean** — no other session is mid-edit in them (17 files under
`platform/orchestration/actions/` are dirty, none of them mine or on this path).

## (c) 2026-09-03 ~11:1xZ — two live facts the bug file (2026-08-19) predates

Both measured at the cluster, both change the reachability story:

1. **The race is physically possible today.** `agent-chassis` is **2/2 replicas**, image
   `v1.0.1356`, pods ~148m old `[MEASURED 2026-09-03 ~11:0xZ]`. The bug file says "no known
   incident" and never states a replica count — so a reader could reasonably have assumed one pod.
2. **There is a guard IN SERIES on the chassis, and the bug file never mentions it.**
   `agent-chassis` runs `CHASSIS_INTAKE_MODE=worker_pool_all` (live deployment env, alongside
   `CHASSIS_DB_MAX_OPEN_CONNS=12`). Under that mode (`platform/agentbase/intake.go`,
   `intake_workers.go`) messages are persisted to `chassis_intake_events` and executed by a
   claim-worker pool that first acquires the message's **serialisation key** —
   `intakeSerialisationKey` (intake.go:215) derives it as **the orchestration_id** for requests,
   and for responses as the request id resolved through `awaited_requests` to the **parent**
   orchestration. `ClaimSerialisationKey` (intake_repo.go:136) is an
   `INSERT … ON CONFLICT DO UPDATE … WHERE lease_expires_at <= NOW()` CAS, and the claims table is
   shared by both pods, so it is **one holder per orchestration fleet-wide**.

   ⚠ **This is the "a mutation that PASSES may have hit a guard in SERIES" trap in its natural
   habitat.** A test of my fix could pass on the chassis for this reason and not because of the
   fix. Any verification has to isolate the arm under test.

## (d) 2026-09-03 ~11:2xZ — where the masking STOPS, which is where the bug still lives

Read, not inferred:

- **Every other agent deployment has no intake claim at all.** `SagaCoordinator` is constructed in
  `platform/agentbase/agent.go`, which every agent binary uses, so both takeover arms run in all of
  them. Deployment census `[MEASURED 2026-09-03 ~11:1xZ]`: `CHASSIS_INTAKE_MODE` is set on
  **`agent-chassis` only**. Running ≥2 replicas without it: `auth-service` 3, `reasoning-agent` 3,
  `web-scrape-adapter` 3, `core-manager` 2, `git-adapter` 2, `github-actions-runner` 2,
  `image-generator-adapter` 2, `admin-dashboard` 2. **[UNVERIFIED]** whether each of those actually
  reaches `handleOrchestrationStatus` in traffic, or only structurally — that is the open question
  in the PLAN, with the query attached.
- **Even on the chassis the claim is not total.** The lease is `intakeLeaseDefault = 180s`
  (intake_workers.go:43); `StuckOrchestrationTimeout` is **300s**. `180 < 300`, so a key can change
  hands *before* the row is old enough to look stuck — the handover and the takeover window are not
  mutually exclusive, they are adjacent.
- **And the old holder does not stop immediately.** `drainKey` (intake_workers.go:~190) heartbeats
  every `lease/3` = 60s and sets `claimLost` when the heartbeat reports the claim gone — but the
  drain loop only tests `claimLost` **between events**, and `processMessage` takes no context, so an
  in-flight event **runs to completion regardless**. The file's own header says as much ("this pool
  cannot impose a hard deadline on an event"). So after a lease handover there is a window, bounded
  by the heartbeat period, in which the old holder is still inside its event while the new holder
  starts the next one **on the same orchestration** — which is exactly 329's precondition.
- Intra-pod is genuinely excluded at the intake layer: `workerID` is
  `fmt.Sprintf("%s/worker-%d", a.PodName, i)` (intake_workers.go:99), unique per worker, and the
  claim CAS is keyed on it. So the chassis's four workers per pod cannot collide *through that
  door*. They are not excluded on any door that bypasses intake.

**Net:** the bug is valid and the fix belongs in the coordinator — the guard that masks it is a
chassis-layer guard, and the defect is in the shared coordinator every agent runs. That is the
framework-level-over-individual-case argument, and it got sharper, not weaker, from measuring.

## (e) 2026-09-03 ~11:2xZ — why no `090` run, stated plainly

Owner ruling 2026-07-31: a `bugs_open/` file asserting a structural cause is not filed until it has
been through the diagnosis loop, **or** the session states why it substituted first-hand
verification. I am not re-filing 329 — it was filed with a council objection behind it — but I am
about to assert a good deal about its reachability, so the same bar applies to me.

**Substituted, and here is why rather than merely that:** `LANDMINES.md` records that a `090` run on
a symbol in a file over ~60KB returns bundles and **no verdict**, and that this looks exactly like a
run still in progress. `platform/orchestration/coordinator.go` is **199,136 bytes** and `state.go`
is **77,392** `[MEASURED 2026-09-03]` — both over the threshold, and the symbols at issue
(`handleOrchestrationStatus`, `TakeOverOrchestration`) live in them. A run would burn ~30 minutes of
queue and return nothing gradable.

Substitute, three independent artefact-level reads: (i) both arms read in full at the lines above,
not grepped; (ii) the live deployment env and replica counts read from the cluster; (iii) the intake
claim path read end to end — key derivation, CAS SQL, lease constant, heartbeat, and the
between-events `claimLost` test.

## (f) 2026-09-03 ~11:3xZ — the measurement that settles where the fix belongs

I had been treating "does the intake claim cover the estate?" as an open question to be answered by
reading env vars on Deployments. That framing was wrong, and the row census corrected it.

**Who actually creates orchestrations** `[MEASURED 2026-09-03 ~11:3xZ]`, from
`orchestration_states.processing_node` over 14 days (the column records the pod that CREATED the
row — `bugs_open/075`; that is exactly the fact I want here), pod-family suffix stripped:

| pod family | orchestrations |
|---|---|
| `agent-chassis` | 3,332 |
| `agent-page-rerender` | 2,215 |
| `agent-build-dispatch-loop` | 660 |
| `agent-page-build-handler` | 412 |
| `agent-internal-link-resolver` | 392 |
| `agent-page-content-writer` | 392 |
| `agent-asset-deployer` | 236 |
| …9 more families | 82–142 each |

**Every family below the first is a SPAWNED pod, not a Deployment** — confirmed against
`kubectl get deploy` (none of them appears in it) and there are **8 `agent-page-rerender` pods
alive right now**. And `platform/agentbase/intake.go` disables intake mode for spawned pods
**structurally**, by the `a.spawned` guard and the `system.agent.` topic-prefix check, with the
header explaining why (a spawned Job pod inherits `personae-prod-config` wholesale and would
execute chassis orchestrations under the wrong identity).

**So the intake claim covers a MINORITY of the processes that drive orchestrations** — `agent-chassis`
is under 40% of the rows and the remainder run the same `SagaCoordinator`, with the same two
clock-only takeover arms, and no serialisation claim of any kind.

This is the whole argument for fixing it in the coordinator rather than leaning on the intake pool,
and I did not have it an hour ago — I had a deployment-env census, which was the question I could
most easily ask rather than the one that mattered.

⚠ **Note what makes this a measurement and not a dressed-up assumption:** it could have come out
`agent-chassis: 100%`, which would have said the guard covers everything that matters and pushed the
fix toward the intake layer. It did not.

**Still `[UNVERIFIED]`, and it is the honest remaining gap:** that two actors have ever *in fact*
been inside `handleOrchestrationStatus` for one orchestration simultaneously. I looked for the
footprint — repeated step names within one row's `execution_path`, 7 days — and found **0**
`[MEASURED 2026-09-03 ~11:0xZ]`, consistent with the bug file's "no known incident". A negative
there is weak evidence either way (a takeover need not leave a duplicate entry), so I am claiming a
**reachable correctness gap**, not an observed incident, and the fix candidates are ranked by
representability rather than by incident count.

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

## (g) 2026-09-03 ~11:4xZ — the throughput lane answers, and there are THREE guards, not two

Messaged the `dispatch_throughput` lane (live session `throughput`) because the 180s/300s ordering
looked like it might be theirs. Their reply corrected me twice and cost me nothing:

- **The ordering is NOT theirs and stays in 329.** Their concurrency ground is the **work-item claim**
  seam (`site_work_items`, selector/loader/claim), not the intake serialisation lease. They have never
  measured `intakeLeaseDefault` and have no reason to think `180 < 300` is deliberate. They asked me
  explicitly not to let it land in their lane by default. **So it is a 329 finding, cited to nobody.**
- **There is a THIRD guard in series and I had listed two.** `claim_work_item_action.go` claims via an
  atomic conditional UPDATE on `site_work_items` that only succeeds while the row is still
  `triaged`/`approved`. Layers, outermost in: (i) the intake serialisation claim (chassis only);
  (ii) **the coordinator arms — the defect**; (iii) the work-item claim CAS on the dispatch path.
  **A coordinator double-takeover cannot double-execute a work item on that path even with (i) and my
  fix both removed.** My "guard in series" instinct was right and my count was wrong, which is the
  more useful correction — I had stopped counting at the layer I happened to be reading.

### The census that bounds the blast radius, and it came out clean

Ran their DOUBLE-HANDLE CENSUS (their RUNBOOK §"Concurrency meters that actually measure
concurrency") over 24 h `[MEASURED 2026-09-03 ~11:4xZ]`:

| handlers | distinct items | items with ≥2 handlers | **overlapping pairs on one item** |
|---|---|---|---|
| 3,044 | 2,911 | 71 | **0** |

The 71 are sequential retries (handlers = attempt_count, not overlapped). **The outcome 329 predicts
is not occurring on the busiest path** — because guard (iii) absorbs it.

⚠ Their caveat, which cuts against their own lane and is the reason it is worth quoting: this census
bounds 329 **only where a CAS exists**. On a path with external side effects and no claim it says
nothing. ⚠ And a shape not to misread on a re-run: a stale-reaped handler's `updated_at` is the REAP
stamp, not end-of-life, so a legitimate successor re-claim inside the reap window reads as an
overlapping pair (discriminator: status FAILED + `error LIKE 'Orchestration stale%'` on the
first-started member, second started minutes not seconds later).

### What this does to the case for fixing it, stated straight

**The justification is no longer "stop an active fire", and I am not going to write it as one.** It is:
where a path happens to sit behind a CAS the defect is absorbed; where it does not, nothing stands
between a five-minute clock and a double execution with external side effects. The fix's value is that
it stops depending on a backstop that is **not present on every path**, that **nobody chose as this
defect's mitigation**, and that **no one maintains as such** — guard (iii) exists to make work-item
claiming exclusive, and its protection of 329 is a by-product that any future refactor may remove
without knowing it was load-bearing.

That is a real argument, and it is a *smaller* one than the bug file's framing. Recorded here before
the design lands so the plan is judged against it rather than against the version of the bug I
believed at 11:00.

### One more thing they gave me, filed so it does not get lost

`claim_work_item_action.go` now carries a pre-claim read for a spend governor (opt-in
`honour_spend_governor`, default OFF, live on `build-dispatch-loop` only) returning
`claimed:false, reason:"spend_governor_shed"` — a distinct non-claim outcome from
`ai_endpoint_unavailable`. **Not mine to touch**, and they have asked to be told if the claim
primitive the arms use changes. Also: **`398` is one of the ambiguous numbers** — two unrelated files,
and most recent commits saying "398" mean the CTA-gradient one. The single-flight bug is
`bugs_open/398_HANDOFF_2026-08-25_scheduled_tasks_row_is_not_single_flight.md`. Resolve by slug.

## (h) 2026-09-03 ~12:1xZ — TWO CORRECTIONS, one of them mine and already propagated

The design agent came back and refuted a premise in the bug file **and** a premise of mine. I
verified both myself before writing either down — an agent's report is another doc, not a
measurement.

### CORRECTION 1 — the bug file's stated mechanism is wrong, and fix candidate (2) is a no-op

> **CORRECTED 2026-09-03:** `bugs_open/329` says `[MEASURED 2026-08-19]` that `SetExecutingStep` and
> `ClearExecutingStep` "end in `r.UpdateState(ctx, state)` — **not** `UpdateStateWithVersion`. So the
> write that marks a step executing does **no** version check." **That is false.** `state.go:883-885`:
>
> ```go
> // UpdateState updates an existing orchestration state with optimistic locking
> func (r *StateRepository) UpdateState(ctx context.Context, state *OrchestrationState) error {
> 	return r.UpdateStateWithVersion(ctx, state, state.Version)
> }
> ```
>
> `UpdateState` **is** the version CAS. So the bug file's fix candidate (2) — "have
> `SetExecutingStep`/`ClearExecutingStep` use `UpdateStateWithVersion`" — is **already the case** and
> would change nothing.

**What the defect actually is,** which is a better bug than the one filed: a **check-then-act across
two reads**. The arm judges "stale" on the *caller's snapshot* (`coordinator.go:761`, `:796`), and
the write that follows does a *fresh* `GetState` → mutate → CAS that **never re-evaluates the
predicate**. Two takers arriving seconds apart therefore both win, because each CASes against the
version it just read. ⚠ And the corollary inverts the intuitive test: **exactly-simultaneous takers
do NOT double-execute today** — the loser's CAS fails and the arm returns the error. So a
`sync.WaitGroup` start-line test would show the unfixed code behaving correctly. The disconfirming
case is the **sequential interleaving**, not the simultaneous one.

**How the filing lane got it wrong, and it is a cheap check I would also have skipped:** they read
the last line of the two helpers, saw the name `UpdateState`, and did not open it. One-line
delegating wrappers are exactly where a name stops describing behaviour.

### CORRECTION 2 — MINE. "Every agent binary runs the arms" is FALSE, and I put it in four places

> **CORRECTED 2026-09-03 ~12:1xZ:** in NOTES (d), in the `bugs_open/329` update §3, in the commit
> message `108791548`, and in a message to the `throughput` lane, I wrote that because
> `SagaCoordinator` is constructed in `platform/agentbase/agent.go`, **"both arms run in every agent
> binary"**, and I listed eight Deployments (auth-service 3 replicas, reasoning-agent 3,
> web-scrape-adapter 3, core-manager 2, git-adapter 2, github-actions-runner 2,
> image-generator-adapter 2, admin-dashboard 2) as unguarded exposure. **They are not exposure. They
> do not run the arms at all.**
>
> Verified by me, not taken from the agent: **only `cmd/agent-chassis` imports
> `platform/agentbase`.** `cmd/reasoning-agent/main.go:13` imports `internal/agents/reasoning`, which
> contains no `SagaCoordinator` and no `ExecuteWorkflow`. Only `cmd/test-spawning` (a dev tool) and
> `internal/core-manager` import `platform/orchestration` directly, and core-manager's use only scans
> state and publishes to `system.commands.workflow.resume`.

**What I did wrong**, precisely: I read "constructed in `platform/agentbase/agent.go`", observed that
agentbase is the shared chassis library, and inferred "therefore every agent binary". I then went and
measured something adjacent and easy — `CHASSIS_INTAKE_MODE` across Deployments — and the fact that
*that* census came back rich made the whole paragraph feel measured. **An env census over
Deployments cannot tell you which binary contains a symbol**, and I never ran the one-line check that
could: `for d in cmd/*/; do grep -rq platform/agentbase $d && echo $d; done`.

### ⚠ The conclusion SURVIVES, and I am not going to quietly enjoy that

The claim "the intake serialisation claim covers a minority of the processes that drive
orchestrations, so the fix belongs in the coordinator" is **still true**, on a **different and
narrower basis** than I gave:

- Not "eight other Deployments run it unguarded" — they do not run it at all.
- But **spawned Job pods run the chassis image inline**, because `intake.go` refuses the mode when
  `a.spawned`. And my own `processing_node` census already showed those are the **majority** of
  orchestration creators: `agent-chassis` 3,332 against 4,900+ from `agent-page-rerender` (2,215),
  `agent-build-dispatch-loop` (660), `agent-page-build-handler` (412) and the rest — every one a
  spawned pod, confirmed against `kubectl get deploy`.

So the evidence I actually gathered supported the conclusion; **the sentence I wrote to justify it
did not.** That is the more embarrassing shape, because a right answer with a wrong reason survives
review and propagates, and the next reader inherits the reason, not the answer.

## (i) 2026-09-03 ~12:5xZ — council round 1: REVISE, and it earned its keep

**Verdict: REVISE**, gated by `prior_art_librarian`. **8 of 11 seats approved — including
`architecture`**, which settles the scope question I had argued (guarantee NARROWED, not widened,
2026-07-29 §1) rather than leaving it asserted. Objectors: `prior_art_librarian` (high + medium +
low), `guardian` (medium + 2 low), `bug_historian` (medium).

### ⚠ MISSTEP FIRST: I read another lane's verdict as mine

I ran `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1`
— the query CLAUDE.md prints for reading a verdict — and got a REVISE about **finetuning.uk's
playground page**, a tools-api route change with CORS and Ollama objections. Nothing in it is mine.
`LIMIT 1` returns the most recent council note **on the fleet**, and on this estate another lane's
round lands between yours and your read as a matter of course.

**The tell I nearly missed:** the objections were coherent, specific, and about a submission — they
just were not about *my* submission. Nothing errored. **The check:** CLAUDE.md already says to use
`SUBMISSION_CORR`, because *"the correlation is the key the artifacts are written under"* — so read
`diagnosis_artifacts WHERE correlation_id LIKE '<your corr>%' AND kind='council_report'`, never the
LIMIT 1. Logged in `WRONG_CALLS.md`.

### The gating objection was RIGHT to be raised and WRONG in its worry — the answer is a citation

`prior_art_librarian [high]`: the design rests on `ExecuteWithOptimisticLocking` reloading **and
re-invoking the mutator closure** each attempt; its sibling `UpdateStateWithRetry` is landmined for
the OPPOSITE (`*state = *reloaded`, discarding local mutations); if it shares that defect, *"the plan
would ship the bug it claims to close."*

It does not, and I had asserted "safe machinery to reuse" **without citing a line**, which is the
objection. `state.go:1417-1437`: fresh `GetState` at `:1421`, `currentVersion` at `:1426`,
**`fn(state)` re-invoked at `:1429`**, CAS against that attempt's version at `:1434`. The sibling
(`:1079-1100`) calls `UpdateState` on the **caller's** pointer and overwrites it from `reloaded` on
conflict — which is exactly why it is landmined and exactly why this design uses the other one.

### The objection I thought was procedural found a real defect

`guardian [low]` + `prior_art_librarian [medium]` both asked what a caller of the deleted
`ClearExecutingStep` would look like **if the compiler could not see it**. I had `go build ./...`
clean, no reflection on `StateRepository`, and only comment references in `.go`. So I grepped
**outside** `.go` — and found a **live database row**: `orchestration_status_vocabulary.RUNNING.written_by`
= `'StateRepository.ClearExecutingStep (state.go:1428)'`, seeded by migration 466.

**Not a caller, invisible to the compiler, and now false.** Nothing would ever have flagged it.
Migration `736` fixes it (documentation column — `is_terminal`/`is_pausable`, which the reaper and
cleanup actually read, are untouched).

⚠ **The lesson is not "grep harder", it is that "no caller" and "no reference" are different
questions,** and a deletion needs the second. A live row seeded by a migration is a reference that
outlives every compile-time check.

### And answering the gating objection turned up a false claim in my OWN comment

Checking whether the claim really refreshes `last_activity`, I found my comment said
`UpdateStateWithVersion` *"bumped Version and LastActivity on this pointer"*. It bumps `Version` in
memory (`:1074`) and stamps `last_activity` in the **database** from a local `now` (`:1051`) — it
does **not** write `LastActivity` back onto the struct. Harmless to the design (the next taker reads
the database, which is where the refresh lives) but **false where a reader would rely on it**. Fixed.

### `bug_historian [medium]` — accepted in full, and acted on

The residual (a live driver judged stuck) is an incomplete fix of a class this platform has
documented repeatedly. **Filed as `bugs_open/461`** rather than re-disclosed in a risk block — with
the 24× gap between `defaultLocalActionTimeout` (7200 s) and `StuckOrchestrationTimeout` (300 s),
why widening the threshold cannot work, and the heartbeat flagged as architecture-scope.

⚠ One thing that changes the seat's calculus and is in 461: **the instrument to size it did not
exist until this fix.** `processing_history` action `stale_takeover_claimed` is the first durable
record a takeover has ever left. 461 says plainly not to size the bug until that query has data, and
that today it necessarily returns 0 because the fix is not rolled.

**Round 2 resubmitted** under the same trail (`RESUBMIT_CORR=3beb3f54…`, run orch `43181b1a`).

## (j) 2026-09-03 ~13:3xZ — the fix MISSED the build that rolled, proven at the artefact with a control

A fresh chassis rolled while I was working: `agent-chassis` is now **`v1.0.1358`** (was `v1.0.1356`),
pods ~6 minutes old. The tempting inference — *"a build went out after I committed, so my fix is
live"* — is exactly the one `MEMORY.md` warns about, so I asked the binary instead of the clock.

```
build provenance  git_commit d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85
git merge-base --is-ancestor b55f837ef d0252fd4…   ->  NO
git merge-base --is-ancestor 367f5a7fd d0252fd4…   ->  YES   (control)
```

**`v1.0.1358` does NOT contain `b55f837ef`.** The control matters and is why this is a measurement
rather than an assertion: an older commit *is* an ancestor of the stamp, so the test discriminates —
a `NO` on both would have meant my command was broken, not that the fix was absent.

**So the fix is committed and still INERT, and rides the build after this one.** Nothing about 329's
behaviour has changed in production. ⚠ Do not read `processing_history @> '[{"action":
"stale_takeover_claimed"}]'` returning 0 as evidence of anything until a build containing
`b55f837ef` is running — today it returns 0 because the code is not there.

⚠ Note also what the provenance line does NOT tell you: it is a **startup** line, so on a busy
service it scrolls out of reach within hours. It was still in range here only because the pods were
minutes old. Later, the binary probe with a present-and-absent control pair is the check with no
shelf life.

## (k) 2026-09-03 ~14:0xZ — council round 2: APPROVED, and the two advisories were both worth acting on

**APPROVED** — *"approved with 2 advisory objection(s) — none high-severity"*, corr `3beb3f54`
(read from `diagnosis_artifacts` by MY correlation, not the `doc_notes` LIMIT 1 that burned me in
(i)). 9 of 11 approve; `prior_art_librarian` moved from its **high** objection to **approve**, so the
citation answered it. `architecture` approved in both rounds.

### `reuse_agent [medium]` — two "takeover" methods with overlapping names, ACTED ON

> *"`TakeOverOrchestration` … a doc comment describing exactly the concept this plan is naming
> `ClaimStaleOrchestration` … never explains … why a second, differently-named, differently-behaving
> 'takeover' method is being added instead of extending it. Two takeover mechanisms with overlapping
> names and different version semantics … is precisely the pattern this platform has already been
> burned by … This needs to be resolved or explicitly distinguished before the edit ships, not left
> implicit in a citation."*

**Right, and the last clause is the bit I had got wrong** — the distinction existed only inside a
submission nobody will read again. Fixed **in the code, at BOTH sites**, so whichever one a reader
meets first tells them about the other: `TakeOverOrchestration` = ownership **bookkeeping**, records
WHO is driving, leaves version/last_activity alone, callers proceed win or lose, excludes nothing;
`ClaimStaleOrchestration` = **mutual exclusion**, decides WHETHER YOU MAY drive, and losing means do
not act. Plus why they must not be merged (the two guarded mechanisms must never govern one column).

Not renamed: `TakeOverOrchestration` is pre-existing, has a live caller and its own council history
(`bugs_open/075`), so a rename is a bigger blast radius than the objection warrants.

### `guardian [low]` — "run the symbol census against the code index" — **THE INSTRUMENT CANNOT DO IT**

The guardian wanted the `ClearExecutingStep` removal checked against the code index rather than
re-asserted from grep, noting that my `.go`-only grep *"already missed one non-Go reference"* (true —
the vocabulary row). I went to run it, and checked the instrument first.

`SELECT count(*) FROM code_symbols WHERE content ILIKE '%ClearExecutingStep%'` → **1**. Looks like a
clean census. **Controls, run in the same breath** `[MEASURED 2026-09-03]`:

| symbol | index content hits | reality |
|---|---|---|
| `ClearExecutingStep` | 1 | declaration only |
| `NewStateRepository` | **1** | **20 files** in the tree contain it |
| `UpdateStateWithVersion` | **1** | many callers |
| `handleOrchestrationStatus` | **1** | called at coordinator.go:165 |

`content` holds each symbol's **own body**, so the only row matching a name is the declaration.
**The query returns 1 for everything and cannot come out otherwise** — it is not evidence, and a
"clean" answer from it would have been the same answer as for a symbol with twenty callers. Filed as
a LANDMINE (footprint `code_symbols`, `code_symbols.content`, `index_code_symbols`).

**So the honest answer to the guardian is that the census it asked for cannot be run there**, and the
evidence stands as: `go build ./...` clean with the function deleted (complete for compiled callers
**by construction** — no index can match that), plus greps for what the compiler cannot see, which is
exactly how the vocabulary row was found in the first place.

### `guardian [low]` on migration 736's prose-matching verify — accepted, not fixed

The verify RAISEs on `written_by LIKE '%ClearExecutingStep%'`. The seat is right that a prose pointer
rots; it is also right that this is one-shot and documentation-only. Left as written and **recorded
as a residual**: `orchestration_status_vocabulary.written_by` is a prose column with no guard, and
nothing will notice the next time a rename makes it false. That is a real (small) class and it is
named here rather than silently accepted.

### Applied

Migration `736` applied 2026-09-03 after approval — `UPDATE 1`, verify block printed
`NOTICE: 736 OK`, and the live row now reads
`StateRepository.ClaimStaleOrchestration (state.go, the EXECUTING_STEP arm of the claim)`.
Recorded in the ledger with `--record-only` (hand-applied, so the runner would otherwise not know).
⚠ I did **not** use `run-migrations.sh --apply`: it takes EVERY pending file, and other lanes have
migrations in that directory.

## (l) 2026-09-03 ~16:0xZ — THE FIX IS LIVE, proven at both binaries with a three-way control

`agent-chassis` is now **`v1.0.1359`**, pods ~155 min old. ⚠ The `build provenance` startup line had
already **scrolled out of reach on both pods** (`--tail=200000` → not found), exactly as the landmine
says it does on a busy service. **An empty result there means "not in range", not "unstamped"** — so
I used the binary probe, which has no shelf life.

**Probe the CAPABILITY, not the commit** — and never a bare positive:

| probe | pod `…nrqf7` | pod `…phgh2` | reads as |
|---|---|---|---|
| `Orchestration is actively executing` (unchanged by me) | PRESENT | PRESENT | **control** — the probe mechanism works |
| `STALE_TAKEOVER_CLAIMED` (created by the fix) | **PRESENT** | **PRESENT** | the fix **shipped** |
| `Found stuck orchestration, taking over` (deleted by the fix) | ABSENT | ABSENT | the old code is **gone** |

Three-way, on **both** pods (one release tag can ship several revisions — `bugs_open/249`). The
control is what makes it a measurement: a broken `grep -aq` would have returned ABSENT for all three
and read as "the fix is not there".

**So `bugs_open/329`'s fix is LIVE as of 2026-09-03 ~13:28Z.**

### ⚠ BUT THE OBVIOUS "IS IT WORKING?" METER IS BROKEN HERE, and my first reading of it was worthless

I grepped the pod logs for `STALE_TAKEOVER_CLAIMED` and `STALE_TAKEOVER_LOST`: **0 and 0**. That
looks like an answer. It is not — **the demand control was also 0**
(`Orchestration is actively executing`, the arm's own non-stale branch, which must fire constantly).

Checking the instrument instead of the result: `kubectl logs -l app=agent-chassis --since=3h
--tail=200000` returns **68 LINES TOTAL** `[MEASURED 2026-09-03 ~16:0xZ]`, on a service that created
**1,588 orchestration rows in the same window**. The log is not a small sample of the traffic; it is
essentially nothing. **Any zero read from it is uninformative**, and would have been just as zero
before the fix shipped.

**Use the durable needle instead — which is precisely why the fix writes one:**

```sql
SELECT count(*) FROM orchestration_states
WHERE processing_history @> '[{"action":"stale_takeover_claimed"}]';   -- 0 as of 2026-09-03 16:0xZ
```

That zero **is** trustworthy: `processing_history` is persisted, not windowed. It means **no takeover
has been claimed since the roll**, which is consistent with everything measured — this was always a
correctness gap rather than a fire.

### Current population, for whoever picks this up

`[MEASURED 2026-09-03 ~16:0xZ]` — the arms' raw material, and it churns fast:

- Traffic control: **1,588** orchestration rows created in the 150 min since the roll. The chassis is busy.
- `EXECUTING_STEP` rows: **18**, of which **8** idle > 5 min at first read — and **1** two minutes
  later (`med-price-collector`, idle 14 min). So stale rows appear and resolve continuously.
- ⚠ **A stale row is not a takeover.** The arms fire only when a MESSAGE ARRIVES for an orchestration
  whose row is already stale. That conjunction is rare, which is why the needle is still 0 and why
  nobody should read 0 as "the fix does not work" — only as "it has not been exercised yet".

## (m) 2026-09-04 ~08:1xZ — re-verified on `v1.0.1360`; still live, still not exercised

A second build rolled overnight. Re-probed rather than assumed the fix survived it — a new image is
built from whatever HEAD was, and "it was in the last one" is not evidence about this one.

`agent-chassis` on **`v1.0.1360`**; fleet **36** images, **min tag = max tag = 1360**, 0 pre-fix.
Binary probe: control `Orchestration is actively executing` **PRESENT** · `STALE_TAKEOVER_CLAIMED`
**PRESENT** · `Found stuck orchestration, taking over` **ABSENT**. **The fix is still live.**

**Needle: still 0** — `processing_history @> '[{"action":"stale_takeover_claimed"}]'` — against
**6,890** orchestrations created in the last 24 h `[MEASURED 2026-09-04]`.

⚠ **That traffic control is now much stronger than yesterday's**, and it is what makes the zero
readable: 6,890 rows means the chassis is thoroughly exercised, so a zero is "the arms have not been
reached with a stale row", not "nothing is running". **Still not a fault, and still not evidence the
fix works** — the arms fire only when a message arrives for an orchestration whose row is *already*
stale, which is rare by design. Do not read the zero either way; read it as "not yet exercised", and
keep watching per the handoff.

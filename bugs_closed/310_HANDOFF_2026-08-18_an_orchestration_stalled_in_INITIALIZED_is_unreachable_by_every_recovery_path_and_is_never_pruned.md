# 310 — an orchestration stalled in `INITIALIZED` is unreachable by EVERY recovery path, and is never pruned

> ## ✅ CLOSED 2026-08-18 — FIXED AND LIVE, and **fixed by another lane, not by this file's migration**
>
> **Migration `464_reaper_initialized_arm.sql`, applied by hand 18:43:03Z by a concurrent lane and
> owner-approved (`42375f96a`, `Council-Submitted: e973d2aa`).** The reaper fired on its own
> scheduled tick at **18:45:53Z** and failed both rows with
> `reaper: stale INITIALIZED for >4h; step=check_health`. `INITIALIZED` is now **0 rows** fleet-wide.
> That is the real-tick verification this file said was still owed — delivered, just not by me.
>
> **This file's own migration `470` is a DUPLICATE and has been retired.** Two sessions worked the
> same gap in parallel and produced **byte-identical** SQL (both payloads md5 `421dfc4f…`, 3,068 B);
> the live text is that value. See "Two lanes, one gap" below.
>
> **⚠ TWO CLAIMS IN THIS FILE ARE CORRECTED BELOW — read them before reusing its reasoning:**
> (a) my "the window is milliseconds" was **derived, and the measurement disagrees**; (b) my
> "463's structural argument transfers" is **wrong**, and the other lane's reason why is the more
> useful finding in this whole episode.

**Filed 2026-08-18** by the `bugfix_281_tool_audit_ported` lane, found while verifying the close of
`bugs_closed/294`. Status: **CLOSED 2026-08-18 — fixed and live.**

**This is `294` one status over.** 294 closed the `RUNNING` gap with migration `463`
(a `failed_running` reaper arm). `INITIALIZED` has the identical shape — a non-terminal status that
no automated path reaches — and was named in 294's own post-close observation as the standing
argument for its fix candidate 2. It is not the same bug and it is not covered by 463.

**Do not read this as an outage.** Health-checking, the subsystem both stranded rows come from,
is working: see the demand control in Evidence. The defect is a permanent slow leak and a
structural hole, not a stopped pipeline.

---

## The one-paragraph version

A row is created in `INITIALIZED` and is meant to leave that status milliseconds later, inside the
same message handler that created it. If the process dies in that window, the row stays there for
ever. The stale-orchestration reaper has four arms — `AWAITING_RESPONSES` (30 min / 90 min),
`EXECUTING_STEP` (4 h) and, since `463`, `RUNNING` (4 h) — and **none for `INITIALIZED`**. The
`database-cleanup` task deletes `COMPLETED`/`FAILED` after 24 h and `EXECUTING_STEP`/
`AWAITING_RESPONSES` after 4 h, and **never mentions `INITIALIZED`** either. The only in-process
consumer of the status is the `case StatusInitialized` arm of `handleOrchestrationStatus`, which
runs only when an inbound Kafka message arrives for that orchestration — and for a stranded row,
none ever will. So nothing will look at it again, and nothing will delete it.

Unlike `RUNNING`, it pins no Kafka topics, so this is **not** a contributor to `bugs_open/240`.
The harm is a permanent row (~4 KB), one silently-lost unit of work per occurrence, and a status
that is invisible to every operator surface.

## Evidence `[MEASURED 2026-08-18, DB clock 18:12Z]`

**The live reaper's arms** — read from the `pre_query` the task is actually wired with, not from a
migration file's intent (`SELECT pre_query FROM scheduled_tasks WHERE name='stale-orchestration-reaper'`):

| arm | predicate |
|---|---|
| `failed_dispatch` | `status='AWAITING_RESPONSES'` AND `owner_agent_type='build-dispatch-loop'` AND idle > 30 min |
| `failed_orchs` | `status='AWAITING_RESPONSES'` AND idle > 90 min |
| `failed_wedged` | `status='EXECUTING_STEP'` AND idle > 4 h |
| `failed_running` | `status='RUNNING'` AND idle > 4 h  *(added by `463`, verified live)* |
| — | **no `INITIALIZED` arm exists** |

**The live pruner** (`scheduled_tasks.name='database-cleanup'`, hourly), clause 3 and clause 4:
deletes `status IN ('COMPLETED','FAILED') AND updated_at < NOW() - INTERVAL '24 hours'`, and
`status IN ('EXECUTING_STEP','AWAITING_RESPONSES') AND updated_at < NOW() - INTERVAL '4 hours'`.
**`INITIALIZED` appears in neither.**

**No live scheduled task mentions the status at all:**
`SELECT name FROM scheduled_tasks WHERE pre_query ILIKE '%INITIALIZED%'` → **0 rows**. Only three
tasks reference `orchestration_states`: the two above plus `claimed-item-timeout`, which is
read-only with respect to it (0 `UPDATE`/`DELETE` against the table in its `pre_query`).

**The census, and why it is a LIFETIME figure rather than a window.** `orchestration_states` *is*
pruned — `COMPLETED`/`FAILED` reach back only to 08-17 — so a count over this table is normally a
~24 h window and must be treated as one. The predicates above are the licence for treating this
particular status differently: **no automated path deletes an `INITIALIZED` row**, which is why one
survives from 2026-07-13. Scope the claim honestly: this is lifetime *with respect to every
automated path*, and a session could have swept rows by hand (24 `CANCELLED` rows dated
2026-07-19→24, written by nothing in this repository, are what a manual sweep looks like here).

| orchestration | owner_agent_type | current_step | idle |
|---|---|---|---|
| `0dcdd076-6515-4c60-b015-a73af84e7c3b` | `generic` | `check_health` | **870.78 h (36.3 d)** |
| `8a3adf9b-e7a0-4fe9-90b1-3fc7237fa125` | `endpoint-health-checker` | `check_health` | **10.24 h** |

Both have `awaited_requests = {}` and `last_activity = created_at` — they never advanced past
creation, which is the signature of the window below rather than of a later stall.

**DEMAND CONTROL — the subsystem works; do not read two rows as a broken pipeline.**
`endpoint-health-checker` logged **964 `COMPLETED`** and 4 `FAILED` runs inside the 24 h retention
window, most recent 16:37Z. So the 07:58Z stall is roughly **1 in 969** for that agent over that
window, `n=1`, and the rate is not otherwise estimable from this table because the successes are
pruned. A zero here would have been meaningless; 964 is what makes the one row informative.

**A natural experiment, and it did NOT reproduce.** A chassis roll is the exact event that should
strand these rows. Pods rolled at **18:00:06Z / 18:00:29Z**; re-censused at 18:12Z, `INITIALIZED`
was **still 2 rows** and `RUNNING` **0**. Read as weak corroboration that the window is genuinely
tiny — not as evidence of a fix, and not as a licence to close this.

## Root cause, in the code

**Exactly one live writer, exactly one reader.** `grep -rn "StatusInitialized" --include="*.go" .`
returns four hits and no more:

1. `platform/orchestration/state.go:26` — the constant.
2. `platform/orchestration/state.go:734` — `StateRepository.CreateInitialState`, the `$11` status
   parameter of the INSERT. **The only live writer.**
3. `platform/orchestration/coordinator.go:741` — `case StatusInitialized` in
   `handleOrchestrationStatus`. **The only reader.** It calls `SetExecutingStep` and then
   `continueExecution`, so it is what normally moves a row on.
4. `platform/orchestration/helpers.go:615` — a second writer that **never fires**: it is inside
   `OrchestratorHelper`, whose constructor `NewOrchestratorHelper` (`helpers.go:581`) has **zero
   callers** repo-wide. Re-verified here by grep rather than cited from `294`, which is what that
   file's Correction 3 asks the next reader to do. The same greps show `NewTimeoutMonitor`'s only
   caller is `NewOrchestratorHelper` (`helpers.go:587`), so the whole `TimeoutMonitor` cluster is
   dormant and is not a recovery path for this or any status. Scope: "no caller in this
   repository", not "uncallable" — both constructors are exported.

**The window is milliseconds, in one process.** `coordinator.go:149` calls `getOrCreateState`,
which runs `CreateInitialState` (row now `INITIALIZED`) and returns; `coordinator.go:165` then
calls `handleOrchestrationStatus` on that same state, which takes the `StatusInitialized` arm and
calls `SetExecutingStep`. Between those two commits there is a `GetState`, three log lines and a
map assignment. **`INITIALIZED` is an inter-step transition, never a durable healthy state** — the
same structural argument `463` used to license its threshold on the *code* rather than on a census,
and it does not expire the way a census does.

A row is therefore stranded iff the process dies inside that window (eviction, OOM, roll), or
`handleOrchestrationStatus` returns before `SetExecutingStep` and the message is not redelivered.
As with `RUNNING`: **no in-process recovery can exist, because the process that would do it is
gone.** An external sweeper is the correct primary fix, not merely the cheapest.

**It is also invisible to every operator surface**, which is why it went 36 days unnoticed:

- `internal/core-manager/admin/dashboard_handlers.go:353` counts "active workflows" as
  `status IN ('RUNNING','AWAITING_RESPONSES','PAUSED_FOR_HUMAN_INPUT')` — `INITIALIZED` is absent,
  so a corpse is not reported as active *or* as anything else.
- `platform/orchestration/actions/cleanup_stale_topics.go:209,213` protects
  `('RUNNING','AWAITING_RESPONSES','PAUSED_FOR_HUMAN_INPUT','EXECUTING_STEP')` — `INITIALIZED`
  absent. **This is the one place the omission helps**: it is why these rows pin no Kafka topics
  and why this is not a `240` contributor.

## ADJACENT FINDING — a status string that no code path can produce `[MEASURED 2026-08-18]`

Found while enumerating the status vocabulary for fix candidate 2. **Filed here rather than
separately because it is latent; promote it the moment anyone wires pause-for-human.**

The Go constant is `StatusPausedForHuman OrchestrationStatus = "PAUSED_FOR_HUMAN"`
(`state.go:28`). **Five production SQL sites guard the string `'PAUSED_FOR_HUMAN_INPUT'`**, which
that constant cannot produce:

| site | what it would get wrong |
|---|---|
| `actions/cleanup_stale_topics.go:209` and `:213` | a paused run's Kafka topics unprotected → deleted mid-pause |
| `internal/core-manager/admin/topic_cleanup_handler.go:85` | same, on the admin path |
| `internal/core-manager/admin/dashboard_handlers.go:353` | a paused run not counted as active |
| `platform/orchestration/monitoring.go:168` | the `paused` metric can only ever report 0 |

`test/tools/db-inspector/check_state.go:127` uses `'PAUSED_FOR_HUMAN'` — **the test tool has the
correct string and production has the wrong one, in five places.**

**Why it is not biting today, stated so nobody "fixes" a live incident that isn't one:**
nothing writes the constant. `grep -rn "StatusPausedForHuman"` returns its definition and one e2e
assertion (`test/e2e/scenarios/human_in_loop_test.go:107`); `pause_for_human_input` appears only in
a fuel-cost table and a doc comment, with no registered action. No `PAUSED_*` row has ever existed
live (`SELECT DISTINCT status FROM orchestration_states` → 6 values, none paused). So the feature
is unwired and the mismatch is **latent**. It becomes live and silent the day someone wires it.

## Fix candidates, ordered by what makes the bad state unrepresentable

**1 — an `INITIALIZED` arm on the reaper, mirroring `463`.** Live config, effective on save, no
build. Closes the immortality; does not stop the producer, which is correct — the producer is a
process dying mid-window and cannot be stopped from inside that process. Lowest risk: the idiom,
the guards, the rollback shape and the induced test are all proven by `463`, applied one day
earlier against the same row. **Recommended, and the one I would take.** Threshold 4 h, matching `failed_wedged`
and `failed_running`, licensed by the code argument above rather than by a census.

**2 — reap on the INVARIANT instead of enumerating statuses** (this is `294`'s residual 1, and the
only candidate that closes the *class*). Any non-terminal row, idle beyond a threshold, with no
`awaited_requests`, is dead regardless of which status it stopped in — so no future status can be
forgotten the way `RUNNING` was and `INITIALIZED` is. **It needs care that candidate 1 does not:**
it must exclude genuinely-durable non-terminal states, and the vocabulary is not what a reader
would assume — `PAUSED_FOR_HUMAN` is the real durable one, it is spelled two different ways across
the estate (see the adjacent finding), and an invariant arm that inherits the wrong spelling would
reap paused runs and look correct while doing it. Do not attempt this without settling the string.

**3 — never persist the status: create the row directly in `EXECUTING_STEP`.** Makes the state
unrepresentable at source, which is the strongest form. But it is a Go change on the hottest path
in the coordinator, it needs a roll, and `INITIALIZED` is load-bearing for the duplicate-key →
`GetState` idempotency branch at `coordinator.go:620` and `:673`. Higher risk than the leak.
**Not recommended now**; the right shape if the class is ever revisited wholesale.

**4 — a topic-pin age guard** — not applicable here, unlike in 294: these rows pin nothing.

## How to verify a fix

Induce it in **both** directions, with the negative control in the same tick — the shape `463`
used, because an arm only ever observed *not* firing is indistinguishable from one that cannot:

```sql
-- (a) induce: a row that MUST be reaped
UPDATE orchestration_states SET status='INITIALIZED', last_activity = now() - interval '5 hours'
 WHERE orchestration_id = '<a disposable test row>';
-- (b) negative control, same tick: a row that must NOT be reaped
UPDATE orchestration_states SET status='INITIALIZED', last_activity = now() - interval '10 minutes'
 WHERE orchestration_id = '<a second disposable row>';
-- then wait one reaper tick (interval_seconds=180) and read BOTH back.
-- pass = (a) FAILED, (b) still INITIALIZED. Both halves are required.
```

Then confirm the pre-existing arms survived the rewrite — a clumsy full-text `pre_query` update
that dropped `failed_wedged` would otherwise pass every assertion about the new arm.

## Notes for whoever takes it

- **Re-read the live `pre_query` before writing a migration against it.** `scheduled_tasks` is
  edited by more than one lane and the update is a full-text rewrite, so it clobbers silently.
  `463`'s rollback gates on `md5(pre_query)` with a three-way branch for exactly this reason —
  and **not** on `updated_at`, which moves for scheduler stamping with the text unchanged.
- **A substring assertion proves nothing about whether the SQL parses.** A stored `pre_query`
  parses only when the task next ticks, so a typo commits happily and then takes out *all* the
  reaper's arms minutes later with no earlier signal. Use the `EXECUTE`-with-sentinel guard from
  `463_..._ROLLBACK.sql` (house idiom: `210_report_pipeline_scheduled_tasks.sql`). The council
  gated `463` at HIGH on precisely this and it is the first thing a reviewer will look for.
- **`463` itself still has no parse check** — the council accepted that as an advisory because it
  was already applied and ledger-recorded. A new migration has no such excuse: put the guard in
  from the start.
- Statuses are disjoint, so a new CTE needs no `NOT IN` exclusion against its siblings.
- **The wider observation, which is the RFC signal and NOT this fix:** the status vocabulary is
  enumerated by hand in at least six places — the reaper, the pruner, two topic-protection sets,
  the dashboard count and the monitoring metric — and **no two of those lists agree**. That is the
  same defect class as the `PAUSED_FOR_HUMAN` mismatch and as `294`, and it is why forgetting a
  status is the default outcome rather than an accident. Candidate 2 is the local answer; a single
  owned definition of "non-terminal" that all six read is the real one.

---

## FIX BUILT 2026-08-18 — candidate 1, written and fully proven, **NOT YET APPLIED**

`docs/agent_docs/sql_for_agents/470_reaper_initialized_arm.sql` (+ `470_..._ROLLBACK.sql`).
Adds a `failed_initialized` CTE to the reaper: `status='INITIALIZED' AND last_activity < NOW() -
INTERVAL '4 hours'`, wired into the projection and the `HAVING` clause like its four siblings.

**It is deliberately unapplied.** This is live config on a fleet-wide reaper — it takes effect the
instant it is saved, with no build step to catch a mistake — so the apply is the owner's call, not
a session's. Everything that can be proven without applying has been.

**Numbering note:** built as `469`, renumbered to `470` mid-session because another lane created
`469_render_audit_rotation_three_day_window.sql` two minutes earlier. Both embedded `pre_query`
texts were re-verified byte-exact by md5 *after* the rename, because a global `sed` over a file
whose payload is SQL is exactly how a payload silently acquires an edit.

**Construction, and why it was not retyped.** The new `pre_query` was built by anchored insertion
into the **live** text captured byte-exactly from the row (md5 `91ba9704…`, 2,582 B → `421dfc4f…`,
3,068 B). `diff` of before/after shows **only** the three intended additions and nothing else. A
reaper rewrite is full-text: retyping it is how a sibling arm quietly disappears.

**Both guards are in the migration itself from the start**, which `463` did not have — its council
round gated at HIGH on the substring-assertion gap and its `editquality` seat logged the advisory
that `463` still lacked the parse check. Guard 1 is the three-way `md5(pre_query)` concurrency
branch; Guard 2 is the functional `EXECUTE`-with-sentinel parse check.

**Proven, every branch, all inside rolled-back transactions:**

| what | result |
|---|---|
| Guard 2 on the new text | `PARSE OK` — parses and executes |
| Guard 2 on a corrupted copy (`failed_initialized AS ((((`) | **CAUGHT** — `syntax error at or near "UPDATE"` |
| Branch A — live is `463`'s text | `UPDATE 1`, arm wired, prior arms intact, parse-checked |
| Branch B — migration run twice | `UPDATE 0`, *"already present — this run is a no-op"*, verify still passes |
| Branch C — a third lane's edit | **`REFUSED`**, naming the remedy — no clobber |
| ROLLBACK sidecar | restores byte-exactly to `91ba9704…` and parse-checks the restored text |

**The induced test, both directions in the SAME tick** — the check this file's own "How to verify"
section prescribes, run against the two real rows with the new `pre_query` executed directly:

| row | idle | outcome |
|---|---|---|
| `0dcdd076…` `generic`/`check_health` | 870.98 h | **`FAILED`** — `reaper: stale INITIALIZED for >4h; step=check_health` |
| `8a3adf9b…` (temporarily set fresh, the **negative control**) | 0.17 h | **still `INITIALIZED`**, untouched |

`initialized_failed = 1` and **every sibling arm reported 0**, so nothing was reaped collaterally.
An arm only ever observed *not* firing is indistinguishable from one that cannot fire, which is why
both halves are recorded and why the control shares the tick.

**Control after every test above:** the live row is **unchanged** — `md5 = 91ba9704…`, and
`INITIALIZED` is still 2 rows. Nothing in this section touched production.

**Still owed before this can close:**
1. The owner's decision to apply, then the apply itself, then re-running the induced test against
   the *scheduled* tick (these tests bypass the scheduler by executing the `pre_query` directly).
2. Council verdict — submitted, correlation recorded in the commit trailer.
3. **Re-read the live `pre_query` immediately before applying.** Guard 1 will refuse rather than
   clobber if another lane has edited it since, but knowing that in advance beats an aborted apply.

## Council — APPROVED at round 1, corr `473553c0-009d-4279-b4db-70b23f32fe16`

Nine seats recorded, **all approve**, decided 2026-08-18 18:32Z. **Read the signal honestly: the
report's own counter says `abstained: 8`**, i.e. most seats declared the change outside their
footprint (`render_guardian`: "Out of scope for this seat"). The substantive reviews are
`editquality`, `reuse_agent`, `guardian` and `debug_historian`. A path-scoped roster largely
abstaining on a `docs/`-path SQL file is a weaker endorsement than `463`'s 12-reviewer round, and
it should be weighed as such rather than quoted as unanimity.

`debug_historian` — the seat that gated `463` at HIGH — approved, citing the byte-exact md5 capture
"not retyped", the content-hash concurrency guard "correctly avoiding the `updated_at` is not a
content-change signal trap", the `EXECUTE`+sentinel parse check, the sibling-survival negative
control, and that the induced test ran inside a rolled-back transaction, "correctly avoiding the
'running a `pre_query` by hand advances the rotation' trap".

**Guardian's objection 1 (low) — ANSWERED, and it was worth checking rather than arguing.** It
asked whether adding a fifth projection column (`initialized_failed`) changes the result shape in a
way that breaks a consumer expecting a fixed column list, and correctly said it could not verify
that from SQL. **It does not.** `cmd/scheduler/main.go:444` `runPreQuery` is fully generic: it calls
`rows.Columns()` at runtime, sizes both slices from `len(cols)`, and builds a
`map[string]interface{}` keyed by column **name** before marshalling to JSON. There is no fixed
column list anywhere in the path. Corroborating: `463` already added a fourth column
(`running_failed`) and the reaper has ticked normally since.

**Guardian's objection 2 (low) — accepted, and this is where it is recorded.** It asked that the
`FORCE=1` scope override be explicit and reviewable rather than merely asserted. It was stated in
the submission's rationale, and it is stated here: the client-side gate refuses submissions whose
edits touch none of `platform/`, `internal/`, `pkg/`, and it keys on **path**. This change lives
under `docs/agent_docs/sql_for_agents/` but is not documentation — it rewrites the live `pre_query`
of a fleet-wide scheduled task, effective on save, with no build step in between. `463` is the
precedent for this class going through the gate. **If that reasoning is wrong, the override is the
thing to challenge, not the migration.**

**Guardian's `missing` note — folded into the apply checklist, not waved away.** It observed that
none of the guard, concurrency or parse-check claims are verifiable from its side, because neither
`scheduled_tasks` nor `orchestration_states` is in the schema it can reach, so they "rest entirely
on the author's own induced tests, which a human should re-run independently before apply". That is
the correct posture and it converges with the `090` limitation recorded below: **three separate
automated reviewers today could not reach live `scheduled_tasks` content.** Re-run the tests at
apply time; do not take them from this file.

## The `090` diagnosis loop: UNVERIFIABLE, not confirmed — and it found a real gap in my evidence

Run per CLAUDE.md's 2026-07-31 ruling (a `bugs_open/` file asserting a cross-cutting or structural
root cause is not "filed" until it has been through the loop, or the session says plainly why it
substituted equivalent first-hand verification). Intake `2747e87b`, run `22bd0dd6`, item
`4504d16f`. **Verdict: `UNVERIFIABLE`, `stopped_by: scope-not-narrowing`. It did NOT refute
anything.** Stating that plainly because an UNVERIFIABLE is not a confirmation and must not be
cited as one — **the claims in this file rest on the first-hand verification above, not on this run.**

**What it did corroborate**, with a Tier-0 citation: the `case StatusInitialized` arm exists in
`handleOrchestrationStatus`; and, from live state, that two rows have sat in `INITIALIZED` with
`updated_at == created_at == last_activity` for over a month — which it correctly characterised as
"state evidence of the SYMPTOM (staleness) but not yet of the CAUSE".

**Why it could not finish, and it is a limitation rather than a failure:** the decisive evidence is
the verbatim `pre_query` text of two `scheduled_tasks` rows, and **the loop could not fetch it** —
its own trail records that it queried columns belonging to `agent_definitions`
(`type`, `display_name`, `task_workflow`) and got back an unrelated `thunder-reaper` row. So it
never saw the reaper or the cleanup task at all. **This is the second independent instance today of
the same limitation on this same area:** the landmine-verifier returned `NEEDS_HUMAN_REVIEW` on a
neighbouring entry with the explicit reason that the central claim "lives in database content and
migration files outside the .go-only index scope and cannot be verified or contradicted".
**Generalise it: a claim whose evidence is live `scheduled_tasks` config is outside what these
verifiers can reach, so their silence on it carries no information either way.**

**The gap it found in my evidence, which was fair and is now closed.** It objected that my premise
"the only consumer is message-driven" rested on function signatures rather than on a cited call
path — `ExecuteWorkflow`/`ProcessResponse` were outside its bundle. Correct objection. Verified
directly and now cited rather than inferred:

- `handleOrchestrationStatus` has **exactly one caller** — `coordinator.go:165`, inside
  `ExecuteWorkflow`. (`grep -rn "handleOrchestrationStatus(" --include="*.go" .` returns the
  definition at `:700` and that one call. `ProcessResponse` does **not** call it, contrary to the
  loop's guess.)
- `ExecuteWorkflow` has exactly **two** callers: `platform/messaging/processor.go:1864`, inside
  `MessageProcessor.executeWorkflow(ctx, msgCtx *MessageContext, …)` — the inbound-message path —
  and `cmd/test-spawning/main.go:141`, a test binary that is not in the production path.

So the chain is now established end to end: **the only production route into the `INITIALIZED`
arm is an inbound message, and a stranded row is one for which no further message will ever
arrive.** That is the premise the whole filing turns on, and it is now a citation rather than an
inference.

---

# CLOSED 2026-08-18 — and the useful part is that another lane got the reasoning right where I did not

## What actually shipped

`464_reaper_initialized_arm.sql`, applied by hand at **18:43:03Z** by a concurrent lane, owner-
approved, recorded in `schema_migrations` (`applied_by='record-only'`, note: *"runner's `--apply`
would sweep ~17 other threads' pending files"*). Committed as `42375f96a`.

**Verified at the artefact, on a real scheduled tick — not by executing the `pre_query` by hand,
which is all my own tests had done:**

| row | idle | after the 18:45:53Z tick |
|---|---|---|
| `0dcdd076…` `generic`/`check_health` | 871 h | **`FAILED`** — `reaper: stale INITIALIZED for >4h; step=check_health` |
| `8a3adf9b…` `endpoint-health-checker`/`check_health` | 10.7 h | **`FAILED`** — same error |

`INITIALIZED` is now **0 rows** fleet-wide, as are non-terminal rows older than 4 h.

## Two lanes, one gap — and they produced byte-identical SQL

I filed this and built `470`; another lane was working the same gap from `294`'s residual and built
`464`. Neither knew about the other. **Both payloads are byte-identical: md5 `421dfc4f…`, 3,068
bytes.** That convergence is reassuring about the change and damning about the coordination — I
checked `/bugs_open/`, `/bugs_closed/` and the work-item queue before filing, and none of them could
show me a session that had written no file yet. `scripts/who-owns.py` reads commits, so a lane
mid-fix is invisible to it; that limitation is already recorded in MEMORY as
`who-owns-is-blind-to-uncommitted-sessions`, and this is a clean instance of the cost.

**`470` is retired as a duplicate, not deleted** — its header now says so. If it is ever run, Guard 1
takes the "already present — this run is a no-op" branch, which is the branch I proved. It is not in
`schema_migrations`.

## CORRECTION 1 — "the window is milliseconds" was DERIVED, and the measurement disagrees

I wrote, in this file and in `470`'s header, that the `INITIALIZED` window is "one `GetState`, three
log lines and a map assignment… MILLISECONDS", and licensed the 4 h threshold as "~7 orders of
magnitude of headroom".

**The other lane measured it instead of deriving it**, over every retained row that has *left*
`INITIALIZED`, using `(processing_history->0->>'timestamp') - created_at`:

> 5,736 rows, all agent types · avg **0.22 s** · p99 **2.01 s** · **max 6.31 s** · over 5 min: **0**

So the real occupancy is up to **6.31 seconds**, not milliseconds — my figure was wrong by roughly
three orders of magnitude. The threshold is still amply safe (4 h is ~2,280× the observed max), but
**it is licensed by that measurement, not by my derivation.** This is the `[INFERRED]` vs
`[MEASURED]` rule biting exactly where it is supposed to: I read a code path, inferred a duration
from it, and stated the inference in the same voice as a finding. The disconfirming result existed
and was one query away.

## CORRECTION 2 — "463's structural argument transfers" is WRONG, and this is the finding worth keeping

I reused `463`'s licence wholesale: *`RUNNING` is an inter-step transition, never a durable healthy
state, therefore so is `INITIALIZED`.* The other lane's commit refuses that reuse, and is right:

> **463'S LICENCE DOES NOT TRANSFER, and not reusing it is half the point.** `RUNNING` is transient
> BY CONSTRUCTION (one writer, one caller, next write flips it back). `INITIALIZED` is a genuine
> WAITING state: created at `state.go:734`, left only when the first message is handled
> (`coordinator.go:741`). Between those it waits on Kafka, so its duration is a **QUEUE property —
> measurable, not derivable.**

That distinction is the real content. `RUNNING`'s bound comes from the code shape and cannot be
otherwise; `INITIALIZED`'s bound is a fact about queue latency that could change tomorrow if Kafka
backs up, and therefore has to be re-measured rather than re-argued. **My one-writer/one-reader
trace was accurate and was answering the wrong question** — it establishes what *writes* the status,
not how long a row *sits* in it. Anyone revisiting this threshold should re-run the p99 query, not
re-read `coordinator.go`.

**Their stuck-signature note, which my filing had as an observation and they turned into a
discriminator:** both live examples had `last_activity == created_at` exactly *and*
`processing_history = []` — "a spawn that never arrived, not a slow one; a slow start would carry
history".

## What remains open, and where it lives

- **The CLASS is still not closed.** This shut one more instance. `294`'s residual 1 — reap on the
  invariant (any non-terminal row, stale, with no `awaited_requests`) — remains the only fix that
  stops a *future* status being forgotten, and the owner has taken it as separate work. The blocker
  named in this file stands: settle the `PAUSED_FOR_HUMAN` / `PAUSED_FOR_HUMAN_INPUT` vocabulary
  first, or an invariant arm will inherit the wrong spelling and reap paused runs while looking
  correct.
- **The adjacent `PAUSED_FOR_HUMAN` mismatch is untouched and still latent.** Recorded above and in
  `LANDMINES.md`.
- **A build trap the other lane paid for and wrote down**, worth repeating because it defeated an
  md5 check: they stripped a trailing newline belonging to the value (`head -c -1` is right for a
  psql dump, wrong for exact python-written text), fusing `> 0` onto the terminator and truncating
  the file to 166 of 215 lines — **and the md5 assertion PASSED on the truncated file**, because the
  extractor's end-pattern never matched and ran to EOF. Assert **structure** as well as content:
  `BEGIN`/`COMMIT`/terminator counts exactly 1, fused lines 0.

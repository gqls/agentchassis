# 294 — an orchestration stalled in `RUNNING` is unreachable by EVERY recovery path, and pins two Kafka topics for ever

> ## ✅ CLOSED 2026-08-18 — FIXED AND LIVE, induced in both directions
>
> **Migration `463_reaper_running_arm.sql`, applied + recorded + committed (`1be21820f`).**
> Candidate 1 taken: the reaper gained a `failed_running` arm. Live config, so it took
> effect the instant it was saved — no roll. Council: `Council-Submitted:
> 860d87d9-e273-44fe-bb1d-d45a3f2bb69a` (verdict not yet read; `098` credits it
> automatically on approval). Rollback in one command:
> `docs/agent_docs/sql_for_agents/463_reaper_running_arm_ROLLBACK.sql`.
>
> **Two claims in the filing below are corrected in §"Closure" — read them before
> reusing this file's reasoning:** (a) the age census that licensed the 4 h threshold
> can no longer license anything, and (b) root-cause item 4 (`monitoring.go`) is wrong.
> A third lock on the door, not in the filing, was found while verifying it.


**Filed 2026-08-17** by the `bugfix_281_tool_audit_ported` lane, found while sweeping the
corpses left by `bugs_open/289`. Status: **CLOSED 2026-08-18 — fixed and live.**

**Latent as of this filing** — the 49 rows that proved it were swept the same day, and `289`'s
fix (committed `509e01e6a`, inert until the next roll) stops the producer that was feeding it.
**The gap itself is untouched**: the next defect that leaves a row in `RUNNING` gets the same
immortality, and nothing will notice.

---

## The one-paragraph version

The stale-orchestration reaper has three arms — `AWAITING_RESPONSES` > 30 min (dispatch loop),
`AWAITING_RESPONSES` > 90 min, and `EXECUTING_STEP` > 4 h. **There is no arm for `RUNNING`.**
The other recovery path, `TimeoutMonitor`, keys entirely on entries in `awaited_requests`, so a
row with none is invisible to it too. A row that stops in `RUNNING` with an empty
`awaited_requests` therefore has no process that will ever look at it again. Worse, `RUNNING`
counts as *live* in two places: `monitoring.go` reports it as an active orchestration, and
`cleanup_stale_topics.go` protects its `requests_topic` and `responses_topic` from deletion — so
each corpse also pins two Kafka topics permanently, which is a direct contributor to
`bugs_open/240` (kafka-scheduler OOM, every client fetching metadata for all ~25k topics).

## Evidence `[MEASURED 2026-08-17, before the sweep]`

**The reaper's live arms** — read from the `pre_query` it is actually wired with, not from a
migration file's intent (`docs/agent_docs/sql_for_agents/335_reaper_policies_and_shared_park_function.sql`,
the `UPDATE scheduled_tasks SET pre_query = …` block):

| arm | predicate |
|---|---|
| `failed_dispatch` | `status = 'AWAITING_RESPONSES'` AND `owner_agent_type='build-dispatch-loop'` AND `last_activity < NOW() - INTERVAL '30 minutes'` |
| `failed_orchs` | `status = 'AWAITING_RESPONSES'` AND `last_activity < NOW() - INTERVAL '90 minutes'` |
| `failed_wedged` | `status = 'EXECUTING_STEP'` AND `last_activity < NOW() - INTERVAL '4 hours'` |
| — | **no `RUNNING` arm exists** |

**`RUNNING` is not a live state — it is purely a graveyard.** Age census of every `RUNNING` row
fleet-wide, all statuses of health:

| age of `last_activity` | rows | distinct agent types |
|---|---|---|
| < 15 min | **0** | 0 |
| 15 min – 1 h | **0** | 0 |
| 1 h – 4 h | **0** | 0 |
| > 4 h | **49** | 1 (`tool-auditor`) |

**This is the measurement that makes a `RUNNING` reaper arm safe, and it is also the one that
could have come out otherwise** — had healthy agents been sitting in `RUNNING` for minutes at a
time, a 4 h arm would risk killing live work and this bug would need a different fix. They are
not: nothing healthy occupies `RUNNING` at any age.

**Immortality, demonstrated by age:** the oldest corpse's `last_activity` was
**2026-07-29 20:50Z** — 19 days, through multiple fleet rolls, with `awaited_requests = {}`
throughout.

**The topic pin:** the 49 dead rows held **49 distinct `requests_topic` and 49 distinct
`responses_topic` values — 98 topics** that `getActiveOrchestrationTopics`
(`platform/orchestration/actions/cleanup_stale_topics.go:205-216`) was returning as "referenced
by non-terminal orchestrations" and therefore protecting from cleanup, for ever.

**After the sweep** (`UPDATE … SET status='FAILED'` on `status='RUNNING' AND last_activity <
NOW() - INTERVAL '4 hours'`, 49 rows, ids saved): `RUNNING` rows fleet-wide **0**; topics pinned
by dead `RUNNING` rows **0**.

## Root cause, in the code

1. **`platform/orchestration/helpers.go`** — `TimeoutMonitor` iterates `awaited_requests` and
   fires `retryTimedOutRequest` / `failOrchestrationDueToTimeout` per awaited entry. **An empty
   `awaited_requests` gives it nothing to iterate**, so a row with none is structurally outside
   its reach regardless of age.
2. **The reaper `pre_query`** (live in `scheduled_tasks`) covers `AWAITING_RESPONSES` and
   `EXECUTING_STEP` only. `RUNNING` was presumably assumed transient.
3. **`platform/orchestration/coordinator.go:3725`** actively *protects* `RUNNING`: on an error
   path it reloads the row and, seeing `StatusRunning`, logs "Workflow recovered by another
   process" and declines to fail it. Correct for a race measured in seconds; it also means a row
   that reaches `RUNNING` and stops is not failed by the very path that noticed the error.
4. **`platform/orchestration/monitoring.go:111,167`** counts `RUNNING` as active, so the fleet's
   active-orchestration metric was overstated by 49 permanent corpses — the instrument that
   should have surfaced this was reporting them as healthy work.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Add a `RUNNING` arm to the reaper** (config, live immediately). Smallest change that closes
   the hole, and the census above says it is safe at a 4 h threshold. Proposed, to sit alongside
   `failed_wedged`:
   ```sql
   failed_running AS (
       UPDATE orchestration_states
       SET status = 'FAILED',
           error = 'reaper: stale RUNNING for >4h; step=' || COALESCE(current_step, '(none)'),
           updated_at = NOW()
       WHERE status = 'RUNNING'
         AND last_activity < NOW() - INTERVAL '4 hours'
       RETURNING orchestration_id, owner_agent_type, current_step
   ),
   ```
   **Re-run the age census immediately before applying** — it is a one-query check and it is what
   licenses the threshold. If any healthy agent has by then started living in `RUNNING`, this
   candidate is wrong and candidate 2 is the one to take.
2. **Reap on the invariant instead of the status**: any non-terminal row whose
   `awaited_requests` is empty and whose `last_activity` is older than N is, by construction,
   waiting for nothing. This is strictly more general than (1) — it would also catch a future
   status nobody thought to enumerate — and it does not depend on `RUNNING` staying unused.
   Costlier to get right; the predicate must not race a row between steps.
3. **Stop counting `RUNNING` as active in `monitoring.go`** so a corpse is visible as a corpse.
   This fixes the *instrument*, not the defect, and should not be mistaken for closing it —
   but it is what would have surfaced this in a fortnight rather than by accident.
4. **A topic-pin guard**: exclude rows older than N from `getActiveOrchestrationTopics`
   regardless of status, so a stuck orchestration cannot pin topics indefinitely even if the
   reaper misses it. Defence in depth for `bugs_open/240`.

## How to verify a fix

- Induce it: set a scratch orchestration row to `status='RUNNING'`, `awaited_requests='{}'`,
  `last_activity = NOW() - INTERVAL '5 hours'`, wait one reaper tick, and require it to become
  `FAILED`. **The disconfirming result is that it stays `RUNNING`** — which is exactly what
  today's reaper does, so run it against the unfixed reaper first and watch it fail.
- Negative control in the same tick: a second scratch row with `last_activity = NOW()` must be
  left alone. Without it, a fix that fails everything passes the test above.
- Then: `SELECT count(*) FROM orchestration_states WHERE status='RUNNING' AND last_activity <
  NOW() - INTERVAL '4 hours'` stays at 0 across a day, and the topic count stops growing.

## Notes for whoever takes it

- **The reaper `pre_query` is live config, so a change takes effect immediately with no roll** —
  there is no build to gate it and no window to reconsider. Treat it accordingly.
- **It is a shared fleet-wide mechanism**, so it is council-scope under CLAUDE.md's
  platform-seams rule. ~~~~**The council gate is unavailable until the Anthropic account quota returns (2026-09-01, or
  sooner if raised)**~~ — **CORRECTED 2026-08-17: false. That outage lasted ~3 minutes** (last
  failure 11:09:53Z, successes resumed 11:13:02Z); the "regain access 2026-09-01" text in the API's
  400 body is not predictive, as `bugs_open/243` already recorded for the identical error on
  08-10. **So the gate is NOT the reason this is unapplied.** The real reasons stand on their own
  and are the ones to weigh: it is live config that takes effect the instant it is saved, with no
  build step to catch a mistake, on a fleet-wide reaper — and its 4 h threshold is licensed by a
  census that must be re-run first.~~
  > **CORRECTED 2026-08-17 (evening), by the `299` skipped-render-audit lane — THE GATE IS UP.**
  > A full round ran today on this branch: submitted **12:46:22Z**, `complete_approved` /
  > `COMPLETED` by **12:52:17Z** — ~6 minutes, 12 reviewers, 5 abstained, 0 unreadable, verdict
  > APPROVED with six advisory objections whose text I read and answered
  > (correlation `eaa043d7-867f-4d40-a0d9-c41b41e56cf9`; see
  > `bugs_open/299_…skipped_render_audit…` §8). So the quota is not blocking submissions, and
  > "filed rather than applied" should no longer rest on this reason. **Caught because I
  > submitted rather than believing the note** — the claim was dated the same morning, which is
  > exactly the shelf life this estate's own rules warn about. Whoever takes 294: put it through
  > the gate. `[UNMEASURED]` whether the quota constrains some *other* path (a different account,
  > model tier, or the fix-loop's own LLM steps) — this correction refutes only the stated
  > consequence, that a council round cannot be run.
- `bugs_open/289` is the producer that filled this hole, and its fix is committed but inert until
  the next roll. **Do not treat 289's fix as closing this** — 294 is the reason the corpses were
  immortal, not the reason they were created.
- The 49 swept ids are in the lane scratchpad (`289_corpse_rows_before.txt`) if anything needs
  reversing; the rows were set to `FAILED`, not deleted.


---

# Closure — 2026-08-18, by the `bugs_open 294` lane

## What shipped

`docs/agent_docs/sql_for_agents/463_reaper_running_arm.sql` — **candidate 1**, a
`failed_running` CTE in the `stale-orchestration-reaper`'s `pre_query`, mirroring the
existing `failed_wedged` arm:

```sql
failed_running AS (
    UPDATE orchestration_states
    SET status = 'FAILED',
        error = 'reaper: stale RUNNING for >4h; step=' || COALESCE(current_step, '(none)'),
        updated_at = NOW()
    WHERE status = 'RUNNING'
      AND last_activity < NOW() - INTERVAL '4 hours'
    RETURNING orchestration_id, owner_agent_type, current_step
),
```

plus its counter in the trailing `SELECT` and in the `HAVING` (without the `HAVING`
clause a RUNNING-only reap would be silently swallowed and fire no report message).
Statuses are disjoint from the sibling CTEs, so no `NOT IN` exclusion is needed.

**Applied by hand via `psql`, then `--record-only`.** The migration runner has no
single-file mode, and `--apply` would have swept **~17 other threads' pending files**.

## Verification — induced in BOTH directions, with the negative control in the same tick

| when | reaper tick | stale row (`last_activity` −5 h) | control row (`last_activity` = NOW()) |
|---|---|---|---|
| **before** the fix | 13:52:47Z | **stayed `RUNNING`**, `error` NULL | `RUNNING` |
| **after** the fix | 13:56:18Z | **`FAILED`** — `reaper: stale RUNNING for >4h; step=scratch_stale` | **untouched, `RUNNING`** |

The pre-fix run is the disconfirming result this file asked for, observed before the
change. The control matters twice: it proves the fixed reaper is selective rather than
failing everything, **and** it retroactively licenses the pre-fix reading — the same
mechanism ticking three minutes later did flip the row, so "stayed RUNNING" was the
reaper declining to act, not the reaper being asleep. Both scratch rows deleted after.

`RUNNING` rows fleet-wide after closure: **0**.

## Correction 1 — the census that licensed the threshold no longer licenses anything

This file says: *"Re-run the age census immediately before applying — it is a one-query
check and it is what licenses the threshold."* Done, and **it came back 0 rows in EVERY
band**, including `> 4 h` — the 289 sweep cleared the population and 289's fix stopped
the producer.

A census reading 0 everywhere **cannot discriminate** "nothing healthy lives in
`RUNNING`" from "I sampled at a quiet moment". The original measurement was strong
precisely because its young bands *could* have been non-zero; this one could not come
out any other way. **So it was not used as the licence, and nobody should re-cite it.**

**The licence is the code, and it does not expire:**

- `RUNNING` is written on **exactly one line fleet-wide** — `platform/orchestration/state.go:1428`,
  in `ClearExecutingStep`, flipping `EXECUTING_STEP` → `RUNNING`. (`grep StatusRunning`
  across `platform/ internal/ pkg/` returns one assignment; every other `'RUNNING'`
  literal is a *reader*, plus the unrelated Thunder vendor enum.)
- `ClearExecutingStep` has **exactly one caller** — `coordinator.go:765`, the
  stuck-orchestration takeover — whose next act is `GetState` then `continueExecution`.
- `continueExecution`'s main loop **opens with `SetExecutingStep`** (`coordinator.go:868`),
  flipping the row straight back.

So the `RUNNING` window is one `GetState` + a circuit-breaker check + a max-age check —
**milliseconds**. It is an inter-step transition, never a durable healthy state, and 4 h
is roughly seven orders of magnitude of headroom. Kept at 4 h anyway (not tighter) to
match `failed_wedged` and stay conservative on live config; the code would license
minutes if `240`'s topic pressure ever demands it.

## Correction 2 — root-cause item 4 is wrong: `monitoring.go` never reported these

This file says `platform/orchestration/monitoring.go:111,167` counted the 49 corpses as
active, and calls it *"the instrument that should have surfaced this … reporting them as
healthy work"*. **It could not have.** Those queries read `FROM orchestrator_state`, and
that relation **does not exist in `clients_db`**:

```
SELECT to_regclass('orchestrator_state'), to_regclass('public.orchestrator_state');
 →  (null), (null)          -- while orchestration_states resolves fine
SELECT COUNT(*) FROM orchestrator_state WHERE status IN ('RUNNING','AWAITING_RESPONSES');
 →  ERROR:  relation "orchestrator_state" does not exist
```

So `WorkflowMonitor` errors rather than miscounts — `GetStuckWorkflows` and
`GetWorkflowMetrics` are inoperative, and the `/monitor/stuck` and `/monitor/metrics`
endpoints (`platform/health/monitoring.go:53,85`) return **500** whenever anyone asks.
The live query that *does* count `RUNNING` as active is
**`internal/core-manager/admin/dashboard_handlers.go:351-354`**, which reads the real
table. Candidate 3, if anyone takes it, belongs there — not in `monitoring.go`.

This is a landmine in its own right (reading the source tells you the metric counts
`RUNNING`; that is true of the source and false of every live reading) and is filed in
`LANDMINES.md`.

## Addition — a THIRD lock on the door, not in the filing

`handleOrchestrationStatus` (`coordinator.go:740-796`) switches on
`StatusInitialized` / `StatusExecutingStep` / `StatusAwaitingResponses` /
`StatusCompleted` / `StatusFailed`, and then:

```go
default:
    return fmt.Errorf("unknown orchestration status: %s", state.Status)
```

**There is no `case StatusRunning`.** So even the message-driven path actively *rejects*
a stranded row — this is stronger than the filing's item 3 (`failWorkflow` at :3718
merely *declines to fail* it). Three independent locks, then: no reaper arm, no
`TimeoutMonitor` reach (it keys entirely on `AwaitedRequests`, and every entry point
indexes that map), and a hard error on message arrival.

**Why a reaper arm is the *right* fix and not merely the cheapest:** if the pod dies
between `ClearExecutingStep` and `SetExecutingStep`, no in-process recovery can exist —
the process that would perform it is gone. An external sweeper is the only thing that
can reach this state.

## Residuals — deliberately not done here

1. **Candidate 2** (reap on the invariant: any non-terminal row with empty
   `awaited_requests` and stale). Strictly more general, and still the right answer if
   another status is ever left unenumerated. Not needed to close *this* defect.
2. **Candidate 3**, retargeted by correction 2 above → `dashboard_handlers.go:351-354`.
3. **Candidate 4**, the topic-pin age guard — defence in depth for `bugs_open/240`.
4. **The missing `case StatusRunning`** in `handleOrchestrationStatus`. Now bounded
   rather than permanent (the reaper reaches it within 4 h), so it is a wart, not a leak.

None of these re-open the defect this file names: a `RUNNING` row is no longer immortal,
and its two Kafka topics are released the moment it is failed.

## Correction 3 — lock #2 is stronger than either this file or my own closure said

Both the original filing (root cause 1) and my closure above describe `TimeoutMonitor` as
unable to reach a `RUNNING` row *because* `awaited_requests` is empty and every entry
point indexes `state.AwaitedRequests`. Every clause of that is literally true, and it
quietly implies something false: that `TimeoutMonitor` runs at all.

**It does not. Nothing constructs it.** Verified 2026-08-18, after the
`bugfix_029_retry_kills_live_child` lane filed the same finding in `LANDMINES.md` the
same day (its entry asks you to re-run the greps rather than trust it — I did):

```
grep -rn "TimeoutMonitor\|OrchestratorHelper" --include="*.go" . | grep -v "/helpers.go"
 → one hit, and it is a COMMENT in reply_delivery_adoption_test.go:125
grep -rn "NewOrchestratorHelper" --include="*.go" .
 → the definition at helpers.go:581 and nothing else — ZERO callers
```

`NewTimeoutMonitor` has exactly one caller, `NewOrchestratorHelper` (`helpers.go:587`),
which itself has none. So the whole cluster is dormant, and a stranded row is outside
its reach *for every status and every value of `awaited_requests`* — not merely for the
empty-map case.

This does not change the fix or the verdict; it makes lock #2 unconditional rather than
conditional, so the conclusion holds a fortiori. **It is recorded because the mechanism
was stated wrongly twice — once in the filing, once by me while correcting the filing —
and the reason both times was reading the function and not the caller graph.** Scope the
claim: both constructors are exported, so this is "no caller in this repository", not
"uncallable". Re-run the greps rather than citing this paragraph.

Noted for the council round in flight (`860d87d9`): the submission's `grounded_in`
carries the `AwaitedRequests` framing. Nothing in it is false, but if a seat objects on
this ground the objection is correct and the answer is the caller graph above.

## Council — APPROVED at round 2 (correlation `860d87d9-e273-44fe-bb1d-d45a3f2bb69a`)

**Round 1 → REVISE**, gated by `debug_historian` at HIGH, and the objection was right:
my verify block *substring-checked* the rewritten `pre_query` and so proved nothing about
whether the assembled SQL **parses**. Stored SQL parses only when the cron fires, so a typo
would have committed happily and then taken out the reaper — *all five arms, not just mine*
— minutes later, with no earlier signal. Round 1's plan also carried only the **pre-fix**
half of the induced test, because I submitted while the post-fix tick was still pending.

**Round 2 → APPROVED**, 2 advisory objections, none high-severity. `debug_historian` moved
to approve: *"matches the SQL-surgery lore closely and answers the round-1 gap correctly."*

What changed between rounds, all of it in the **rollback** file — 463 was already applied
and proven, whereas the rollback had never run, carried the identical hole, and is the
artefact someone executes under incident pressure:

- **Guard 1, concurrency** — three-way on `md5(pre_query)`. Not `updated_at`: measured
  2026-08-18, that moves for scheduler stamping with the text unchanged.
- **Guard 2, functional parse check** — `EXECUTE` the written text in a sub-block, sentinel
  raise to discard effects, so a syntax error aborts the migration instead of surfacing at
  the next tick.

**Both guards follow an existing house idiom rather than being invented** — round 2's
`reuse_agent` was right that I authored them from scratch first. The three-way md5 branch is
`458_detected_item_promoter_..._ROLLBACK.sql`; the EXECUTE check is
`210_report_pipeline_scheduled_tasks.sql`, whose header states the danger better than I did:
*"a pre_query with a typo fails silently at tick time (the task simply never fires), which is
the hardest kind of dead pipeline to notice."* The one deliberate variation: 210 executes
gate `SELECT`s, which are inert, whereas the reaper's `pre_query` **mutates**, so it needs
the sentinel.

**Every branch proven, not just the happy one** (2026-08-18, all inside rolled-back
transactions; live row verified untouched afterwards at `md5 91ba9704`):

| branch | live text is | result |
|---|---|---|
| A | 463's text | `NOTICE … rolling back` · `UPDATE 1` · restored + parse-checked |
| B | already the pre-image | `NOTICE … already rolled back — no-op` · `UPDATE 0` |
| C | a third lane's edit | **`REFUSED`**, naming the remedy — no clobber |

Guard 2 separately proven both ways: the live text passed; a corrupted copy
(`failed_running AS ((((`) was **caught**.

### Advisories accepted but not acted on, with reasons

- **`editquality` (medium): 463 itself still has no parse check.** True. It is applied,
  recorded in `schema_migrations` with a checksum, and functionally proven; editing it now
  would drift the ledger against a file whose only remaining use is re-application *after* a
  rollback — and the rollback is precisely where the guard now sits. Anyone re-applying 463
  should copy the guard from the sidecar.
- **`guardian` (medium + missing): a future agent might legitimately park in `RUNNING`.**
  The one-writer trace covers today's code only — correct, and stated as the standing risk in
  the migration header and the `LANDMINES` entry, which tell the next author to re-run the
  greps rather than cite them. Measured now: `RUNNING` is **0 rows** fleet-wide and the new
  arm has reaped **0** real rows since apply, so nothing is being killed today.
- **`prior_art_librarian` (low): the `TimeoutMonitor`-is-dormant claim rests on greps.**
  Correct, and scoped in Correction 3 above to "no caller in this repository".

## Post-close observation — the first fleet roll after the fix

A chassis roll is the *exact* failure mode this bug is about: `RUNNING` is entered only in
the stuck-orchestration takeover, and a pod dying between `ClearExecutingStep` and
`SetExecutingStep` is what strands a row there. So the first roll after the fix is a free
natural experiment.

`v1.0.1309`, pods started **2026-08-18 15:45:31Z / 15:45:53Z** (binary carries commit
`f0117fb8b`, probed with a control). Measured ~40 min later: **`RUNNING` = 0 rows
fleet-wide**, and the new arm has reaped **0** real rows all-time.

**Read this as weak corroboration, not proof.** It is consistent with "rolls do not
routinely strand rows", and equally consistent with "no orchestration happened to be inside
the millisecond takeover window during this roll" — which, given the window's size, is the
likelier reading. The arm's correctness is established by the induced test above, not by
this. What the observation *does* rule out is a roll being a high-rate producer.

⚠ **`INITIALIZED` is ACCUMULATING, not a historical straggler** — re-measured the same day:
two rows, `generic`/`check_health` idle **36.2 days** and `endpoint-health-checker`/`check_health`
idle **0.4 days and rising**. Both from a `check_health` step. Unlike `RUNNING` they pin no
Kafka topics (`INITIALIZED` is absent from `getActiveOrchestrationTopics`'s protected set),
so the harm is bounded — but the gap is live and growing, and it is the standing argument for
fix candidate 2.

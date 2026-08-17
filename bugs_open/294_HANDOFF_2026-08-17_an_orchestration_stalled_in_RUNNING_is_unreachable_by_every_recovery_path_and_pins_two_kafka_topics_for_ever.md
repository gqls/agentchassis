# 294 — an orchestration stalled in `RUNNING` is unreachable by EVERY recovery path, and pins two Kafka topics for ever

**Filed 2026-08-17** by the `bugfix_281_tool_audit_ported` lane, found while sweeping the
corpses left by `bugs_open/289`. Status: **OPEN, UNOWNED.**

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

# 060 — the platform keeps no durable record of which agents have ever run (`usage_count` is dead; `orchestration_states` is pruned at 24h)

**Filed:** 2026-07-22 by the dormant-agents thread, out of `bugs_open/044`.
**Severity:** latent, structural. Nothing errors. It is the reason `044`'s
detector can only measure *recent* activity, not lifetime "has this ever run".
**Status: CLOSED 2026-07-26 — fixed AND live on chassis v1.0.1167.** See
**§CLOSED** at the bottom for the fix, evidence, and the follow-up filed while
closing it.

## The finding

Building `044`'s dormant-agents detector, I needed a durable answer to *"has
this agent ever run?"* — the exact question that was answered wrong for
`section-editor` (declared nonexistent while it had **3 lifetime runs**). There
is **no durable signal for it anywhere**:

- **`agent_definitions.usage_count` is unmaintained** — it is `0` for **every**
  active agent, including demonstrably-live ones:
  ```
  SELECT type, usage_count FROM agent_definitions
  WHERE type IN ('fix-proposer','page-build-handler','section-editor','generic')
    AND is_active AND COALESCE(is_snapshot,false)=false;
  -- all 0
  SELECT bool_and(COALESCE(usage_count,0)=0) FROM agent_definitions
  WHERE is_active AND COALESCE(is_snapshot,false)=false
    AND jsonb_typeof(default_config#>'{workflow,steps}')='object';   -- t (162/162)
  ```
  The column exists (it is even SELECTed by `AgentDefinitionDiscovery.FindByType`,
  `platform/discovery/agent_discovery.go`) but nothing ever increments it.
- **`orchestration_states` is pruned hourly at 24h.** The `database-cleanup`
  scheduled task runs `DELETE FROM orchestration_states WHERE status IN
  ('COMPLETED','FAILED') AND updated_at < now() - INTERVAL '24 hours'`. Measured
  2026-07-22: 1,737 rows, oldest 9 days, **94% from the last ~36h**. (On
  2026-07-20 it briefly held 106k rows back to 05-28 — the cleanup was not
  pruning then; it is now, so ~24-48h is the going-forward window.)
- **`orchestration_state_audit`** retains ~2 days — same order, not durable.

So the only run-history substrate is pruned to ~1-2 days, and the one column that
was clearly *meant* to be the durable counter is dead.

## Why it matters

`044`'s detector uses the step-fingerprint over `orchestration_states`. With a
~24h window, "never observed" over-flags massively: any active agent that has not
run *today* (e.g. `fix-proposer` between fixes) reads as dormant. That is a
**recent-activity** signal, not the **lifetime** signal `044` actually wants. The
detector now guards against this (it refuses to EMIT while the observation window
is shorter than the age floor, and says so), so emission is effectively gated off
until this bug is fixed — the detector is report-only in practice. The lifetime
"has this ever run" question — the one that would have saved `section-editor` —
**cannot be answered at all** with today's schema.

## Fix candidates (design decision — owner)

1. **Revive `usage_count` + add `last_ran_at`, written in the completion path.**
   The obstacle is the SAME attribution problem that makes `owner_agent_type`
   useless: orchestrations record `owner_agent_type='generic'`, so the completion
   handler must resolve *which `agent_definitions.type` it ran as* before it can
   credit the right row. That resolution is the crux — get it wrong and the
   counter is as misleading as `owner_agent_type`. Forward-only (no backfill).
2. **A dedicated `agent_run_log`** (type, orchestration_id, ran_at) written once
   per run and NOT swept by `database-cleanup` (or swept on a long retention,
   e.g. 400 days). Cheapest reliable lifetime signal; bounded by one row/run.
3. **Exempt a compact projection from the cleanup** — e.g. keep
   `(agent_type, max(created_at))` in a summary table updated before the DELETE.

## How to verify a fix

> **CORRECTED 2026-07-26.** This bar as originally written cannot be met by the
> design that actually shipped (`PLAN_2026-07-24_durable_run_record.md`): the new
> `agent_run_stats` table is explicitly **forward-only, no backfill** — it cannot
> know pre-deploy history. `section-editor` has only 3 lifetime runs in 5 months,
> so gating verification on it reappearing is impractical; it will simply take
> however long it takes to run again after deploy, which is the correct forward-
> only behaviour, not a fix defect. The practical verify bar is a **frequently-
> running** agent (`fix-proposer` or `page-build-handler`) getting a row within
> minutes of the image roll — that proves the mechanism, and `section-editor`
> reappearing later is confirmatory, not required.

A frequently-running agent (`fix-proposer` / `page-build-handler`) must acquire an
`agent_run_stats` row within minutes of the new image going live, even though
`orchestration_states` would have pruned the same evidence within 24h. Then
`044`'s detector, pointed at the durable signal, stops flagging active agents and
its emission guard converges (un-gates) once the tracking window matures.

## References

- `bugs_open/044` — the dormant-agents detector this blocks; its window guard and
  report already surface this gap. Workstream:
  `docs024.../dormant_agents_inventory/`.
- `platform/orchestration/actions/diagnose_dormant_agents_action.go` — the
  detector; see its WINDOW GUARD comment.
- `database-cleanup` scheduled task — the 24h prune.

## §CLOSED 2026-07-26

**Fix shipped: candidate 1 + 2 combined.** A 2026-07-24 session on this same
workstream had already designed and written the fix; it sat uncommitted. Adopted
2026-07-26 after independently reading every line, compiling clean, and running
its tests: commit `baf887a8e`.

- **New table `agent_run_stats`** (migration
  `docs/agent_docs/sql_for_agents/203_agent_run_stats.sql`, applied by hand +
  `--record-only`): one row per RESOLVED real agent type, upserted at every
  orchestration start, never pruned. This is candidate 2 (a dedicated log-like
  table), not a revived `usage_count` — a new table stays isolated from the
  versioned/snapshotted `agent_definitions` row.
- **Resolution fix** (candidate 1's crux): `processor.go`'s `executeWorkflow`
  resolves the real dispatched type via `extractGroupInfo` (the same value
  `selectWorkflow` used to load the workflow), sets
  `ExecutionContext.RunAgentType`, and fires a detached-context
  `discovery.RecordAgentRun` write. `coordinator.determineOwnerAgentType` prefers
  `RunAgentType` over the dispatch-path `Sender.AgentType`/`generic` fallback
  (this half — the field + the preference — had already landed as a same-file
  passenger in an unrelated commit, `3af7b9d8d`, 2026-07-24).
- **Defensive fallback** in `ai_actions.go`'s `ExecuteLLMPromptAction`: now that
  `owner_agent_type` can carry a real resolved type instead of always `generic`,
  its one execution-driving reader falls back to loading `generic` on a failed
  lookup instead of hard-failing.
- **`bugs_open/044`'s detector rewired** to read `agent_run_stats` directly,
  replacing the step-fingerprint method entirely (see that bug's own closure).

**Live evidence (chassis v1.0.1167, 2026-07-26 ~15:00–15:10 UTC):**
```
-- fresh agent_run_stats rows, filled by real traffic within ~2 minutes of the roll:
agent_type              | run_count | first_ran_at
council-gate            | 1         | 2026-07-26 15:02:13
endpoint-health-checker | 2         | 2026-07-26 15:00:50

-- owner_agent_type on fresh generic-dispatched orchestrations now names the real agent:
orchestration_name                       | owner_agent_type
council-gate-orchestrate-0726-1502       | council-gate
endpoint-health-checker-orchestrate-...  | endpoint-health-checker
```
`council-gate` is the exact agent the original step-fingerprint method could
**never** measure (its 099 roster mirror copies `fix-proposer`'s steps
verbatim, so it had no unique step key — the mirrored-agent blind spot). It now
records under its own real type on the first live run. This is the strongest
available proof the resolution fix works, stronger than waiting for
`section-editor` (only 3 lifetime runs in 5 months) as the original verify bar
asked — see the correction above.

**Corrected verify bar met:** `endpoint-health-checker` (frequent) and
`council-gate` acquired `agent_run_stats` rows within ~2 minutes of the image
roll, and both correctly stopped appearing in the rewired `044` detector's
never-run list on the very next sweep (below).

**Two genuine message drops during verification, root-caused, not a defect in
this fix:** `postgres-clients-0` crash-looped several times in the ~15 minutes
after the roll (`bugs_open/082`, filed the same day by another workstream,
independently: a 1s exec liveness probe kills the pod under CPU contention from
a neighbour). Two dispatches (a council-gate submission, one dormant-agents
sweep trigger) landed exactly in a restart window; chassis logs show the
message WAS consumed but `FindByType` failed with `SQLSTATE 08P01` mid-lookup,
so the processor fell back to a synthetic dynamic workflow that is itself
invalid, and returned an error before any `orchestration_states` row was ever
created — hence "message consumed, no row, no error visible from outside."
Confirmed by reading the pod logs directly, not inferred. Retried after
`postgres-clients-0` held stable for 5 consecutive checks; both retries
succeeded end-to-end. **Not a residual of this fix** — filed against 082, not
here.

**Council review:** submitted (`SUBMISSION_CORR=2d2748e8-8a60-45a2-8cce-68148af9076e`)
and resubmitted twice; both of the first two runs stalled mid-review — the
same `bugs_open/082` DB instability, confirmed by an idle audit trail rather
than any error. Advisory only (CLAUDE.md norm) — does not block this closure.
If a verdict lands later, it will be reconciled against commit `baf887a8e` the
same way other threads in this repo have reconciled a late verdict against an
already-landed commit (no trailer was added at commit time; forward-only
precludes amending).

**Verify recipe for anyone re-checking:** `SELECT agent_type, run_count,
first_ran_at, last_ran_at FROM agent_run_stats ORDER BY last_ran_at DESC;` —
should be growing steadily as real traffic runs. `section-editor` will appear
here once it runs again post-2026-07-26 (forward-only, no backfill — expected
to take a while given its historical 3-runs-in-5-months rate; its absence is
not a regression).

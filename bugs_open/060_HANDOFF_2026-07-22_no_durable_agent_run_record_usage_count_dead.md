# 060 — the platform keeps no durable record of which agents have ever run (`usage_count` is dead; `orchestration_states` is pruned at 24h)

**Filed:** 2026-07-22 by the dormant-agents thread, out of `bugs_open/044`.
**Severity:** latent, structural. Nothing errors. It is the reason `044`'s
detector can only measure *recent* activity, not lifetime "has this ever run".

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

`section-editor` (3 lifetime runs, all pruned) must read as *has-run* from the
durable signal even though `orchestration_states` no longer contains it. Then
`044`'s detector, pointed at the durable signal, stops flagging active agents and
its emission guard can be lifted.

## References

- `bugs_open/044` — the dormant-agents detector this blocks; its window guard and
  report already surface this gap. Workstream:
  `docs024.../dormant_agents_inventory/`.
- `platform/orchestration/actions/diagnose_dormant_agents_action.go` — the
  detector; see its WINDOW GUARD comment.
- `database-cleanup` scheduled task — the 24h prune.

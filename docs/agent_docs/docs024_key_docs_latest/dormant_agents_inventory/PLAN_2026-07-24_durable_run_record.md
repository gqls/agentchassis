# PLAN — durable agent-run record + make owner_agent_type useful (bugs_open/060, 044)

**Started:** 2026-07-24. Owner approved building both halves of `bugs_open/060`:
a durable per-agent run record, AND making `owner_agent_type` reflect the real
agent instead of `'generic'`. This unblocks `044`'s dormant detector (gives it a
lifetime signal that survives the 24h `orchestration_states` prune).

## Root of the "generic" problem (diagnosed from code, 2026-07-24)

- `determineOwnerAgentType` (`coordinator.go:3106`) returns `execCtx.Sender.AgentType`
  → `AGENT_TYPE` env → `'generic'`. For an orchestration dispatched through the
  GENERIC path (scheduled tasks, manual kcat to `system.agent.generic.requests`),
  the sender is `user`/scheduler and the pod's `AGENT_TYPE` is `generic`, so
  `owner_agent_type='generic'` — even though the pod runs a SPECIFIC agent's
  workflow.
- The real target type IS known: `selectWorkflow`/`extractGroupInfo`
  (`processor.go`) reads `config.agent_type` from the message and loads THAT
  agent's workflow via `FindBestGroup`. So the authoritative resolved type is
  `resolvedType = extractGroupInfo(msg) || p.agentType`.
- Dedicated-pod agents (`feature-implementer`, `page-rerender`, …) ALREADY record
  their real type — the bug is specifically the generic-dispatch path.

## Safety (mapped exhaustively 2026-07-24, subagent + read)

**No routing/topic logic derives from `owner_agent_type`** — request topics come
from `state.RequestsTopic`/`MY_REQUESTS_TOPIC`, response topics from the awaited
record. Nothing branches on the value being exactly `'generic'`. The change is the
DOCUMENTED structural fix (a prior handoff prescribed "`selectWorkflow` should set
`owner_agent_type` to the resolved group/agent type"). Every observability/
provenance reader (`agent_error_log.agent_type`, immune-system triage, finetune
export, monitoring SQL that filters `owner_agent_type='<realtype>'`) gets MORE
accurate. Two places need care:

1. **`ai_actions.go:196`** — a FALLBACK that loads an agent definition by
   `params.AgentType` (= `OwnerAgentType`). Today it loads `generic`; after the
   change it loads the real target (more correct). RISK: if the value is ever not
   a loadable active `agent_definitions.type`, the action fails where it used to
   succeed. MITIGATION (both): (a) only write real agent types to
   `owner_agent_type`; (b) make that fallback DEFENSIVE — on load failure, fall
   back to the previous behaviour instead of hard-failing.
2. **`SEED_code_index_refresh_cadence.sql` self-exclusion** (`owner_agent_type NOT
   IN ('index-orchestrator','code-indexer')`) — currently inert (those run
   generic), will START working as designed. Verify the first reindex after ship
   still derives the expected ref. Benign (worst case the GATE skips the tick).

## Design

### A. Durable run record — new table `agent_run_stats` (migration 203)

Keyed on the RESOLVED agent type, NOT pruned. One upsert per orchestration start
(a start = work was ROUTED to the agent, which is exactly the dormancy signal).
```
agent_run_stats(agent_type PK, run_count, first_ran_at, last_ran_at,
                last_orchestration_id, updated_at)
```
Why a new table, not reviving `usage_count`: `usage_count` lives on the versioned/
snapshotted `agent_definitions` (its `UpdateUsageCount` writer exists but is never
wired). Writing per-run there churns `updated_at` on the config table and touches
snapshot rows. A dedicated table is isolated, carries timestamps, and is not
pruned.

**Why keyed on resolved type is strictly better than the fingerprint method:** it
records the ACTUAL dispatched type, so (a) no mirrored-agent blind spot
(council-gate is recorded as itself if it runs), (b) no council/subtree
false-negatives, (c) survives the 24h prune. The fingerprint method becomes
redundant.

### B. Writer — `RecordAgentRun` (fire-and-forget) in `executeWorkflow`

`processor.executeWorkflow` computes `resolvedType` once and:
- fires a detached-context upsert into `agent_run_stats` (best-effort — a lost
  write is re-set on the next run; never blocks the request);
- sets `execCtx.RunAgentType = resolvedType` so the owner_agent_type fix (C) picks
  it up.

### C. owner_agent_type — new `ExecutionContext.RunAgentType`, preferred at creation

- Add field `RunAgentType` to `ExecutionContext`; thread through `ToHeaders`/
  `FromHeaders` (header `run_agent_type`).
- `determineOwnerAgentType` prefers `execCtx.RunAgentType` (a real agent type)
  above Sender/env/generic.
- Make `ai_actions.go:196` defensive (fall back on load failure).

### D. Dormant detector — read `agent_run_stats` (lifetime signal)

- never-ran = active `agent_definitions.type` with NO `agent_run_stats` row.
- observation window = `now - min(first_ran_at)` — it GROWS (not pruned), so the
  window guard added 2026-07-22 now converges: after ~ageFloor days of tracking,
  emission un-gates automatically. No mirrored-agent blind spot to carry.
- Keep the age floor + window guard. Report states "tracked since <first_ran_at>".

## Cold-start (stated honestly)

`agent_run_stats` is forward-only — it cannot know pre-deploy history. Right after
ship it is nearly empty, so "not in agent_run_stats" is weak until the table has
accumulated ≥ the age floor. The window guard handles this exactly: it emits
nothing until `now - min(first_ran_at) ≥ age_floor`. So the detector stays
report-only for ~2 weeks post-deploy, then begins emitting a reliable signal. This
is the correct behaviour, not a limitation to work around.

## Sequencing

1. Migration 203 (create table) — safe to apply anytime (additive, empty).
2. Go: writer + owner_agent_type + detector rewire — inert until image roll.
3. Commit; council-review (platform change; CLAUDE.md norm 2026-07-24).
4. Apply migration; owner rolls image.
5. Verify: `agent_run_stats` fills; recent generic-dispatched orchestrations show
   real `owner_agent_type`; `ai_actions.go:196` fallback still safe; reindex ref
   unchanged.

## Risk register

- Hot-path write: fire-and-forget, detached ctx, short timeout, single-row upsert.
  Bounded contention per type.
- `owner_agent_type` semantic change: safe per the map; the one execution-driving
  reader (`ai_actions.go:196`) is made defensive.
- Migration number 203 could collide with a concurrent session — verify pending
  list before applying.

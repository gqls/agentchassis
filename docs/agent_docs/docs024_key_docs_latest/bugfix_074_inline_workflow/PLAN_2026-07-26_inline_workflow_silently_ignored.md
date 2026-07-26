# PLAN — bugs_open/074: a scheduled task's inline workflow is silently ignored

**Opened 2026-07-26.** Bug filed 2026-07-25 by the model_directory_pipeline session, which
repaired its own half and left the class — and one live casualty — open.

## The defect, restated after reading the code

A `scheduled_tasks` row that carries its workflow inline —
`input_data: {"config": {"agent_type": "generic", "workflow": {…}}}` — fires, creates an
orchestration, runs the **target agent's own** workflow, and stamps both timestamps. For
`target_agent_type='generic'` that workflow is a single `complete_workflow` no-op, so the run is
green and does nothing.

> **CORRECTION to the bug file's account (2026-07-26).** The bug file says "the chassis runs the
> target agent's own workflow, not the one in `input_data`". True, but it stops one step short of
> the mechanism, and the missing step is what decides the fix:
>
> **The chassis DOES honour an inline workflow — it just has to be at `body.config.workflow`.**
> `platform/messaging/processor.go:897-903` (`selectWorkflow`, Priority 1, ahead of group
> discovery). That path is live in production, not a testing curiosity:
> `DispatchFeedSourcesAction` (`dispatch_feed_sources_action.go:224`) dispatches exactly that
> shape, and **58 `orchestration_states` rows carry its inline workflow** (measured 2026-07-26:
> `SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%call_ingester%'`).
>
> What is actually broken is that **the scheduler cannot express it**. `fireTrigger`
> (`cmd/scheduler/main.go:427-433`) owns the envelope: it builds `config` itself as
> `{"agent_type": task.TargetAgentType}` and puts the whole `input_data` column underneath. A
> workflow authored inside `input_data` therefore lands at `body.input_data.config.workflow` —
> one level too deep, read by nothing.

Same nesting family as **`bugs_closed/054`** (a task's `input_data` authored as a whole message
envelope), whose ruling is the reason this fix takes the shape it does.

## Decision: reject the shape, do not teach the scheduler to read it

Owner call, 2026-07-26, taken with the alternative on the table.

- `bugs_closed/054` settled that `scheduled_tasks.input_data` is **payload only** and that
  `fireTrigger` must not reach into it: *"a legitimate payload field named
  `input_data`/`action`/`config` would be silently eaten."* Lifting `input_data.config.workflow`
  would re-open precisely that.
- A workflow's home is an `agent_definitions` row, where it is versioned, snapshotted and
  discoverable. Two homes is the drift class this repo keeps filing bugs about.
- The repair pattern is already proven in the same bug: model_directory's
  `SEED_directory_freshness_agent.sql`, verified live (see NOTES).

## The three legs

| leg | what | live when |
|---|---|---|
| **A — prevent** | `CHECK (NOT (input_data->'config' ? 'workflow'))` on `scheduled_tasks`, migration `217` | on apply — no image roll |
| **B — repair** | `evidence-freshness` → its own `agent_definitions` row, dry-run staged; `adoption-tracker-freshness` cleared; both seeds rewritten | on apply |
| **C — detect** | scheduler WARNs and skips rather than firing a task whose workflow it is about to discard | **inert** until the next `kafka-scheduler` build |

Leg C is belt-and-braces by construction: with A in place the state it detects cannot be
authored. It stays in because a constraint can be dropped or a row restored, and because the
discard happens in the scheduler — that is where it should be visible.

## Sequencing (B before A)

The constraint cannot go on VALID while two rows violate it, and `evidence-freshness` is another
workstream's live, enabled task that has never once run. So: repair first, with a **`dry_run: true`
pass** to see what the sweep would do to all 8 evidence bases before letting it write anything.

## Out of scope

- **Council gate**: `097_TRIGGER` refuses submissions touching none of `platform/`, `internal/`,
  `pkg/` (its line 78). This work is `cmd/` + SQL + docs — no council run, said here rather than
  left as a silent gap in the coverage report.
- Re-enabling or re-cadencing anything else in `scheduled_tasks`. The other two envelope-shaped
  rows (`diagnose-pipeline-trigger`, `ai-endpoint-health-check`) are the 054 family, carry no
  workflow, and are left alone.

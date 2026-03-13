# 014 — Scheduler & Scheduled Tasks Guide

How the kafka-scheduler works, how tasks are configured, and the rules for creating new ones.

---

## How the Scheduler Works

The scheduler is a single Go binary (`cmd/scheduler/main.go`) running as a 1-replica Kubernetes Deployment. It has no Kafka consumer — it only produces messages. Every 30 seconds (configurable via `TICK_INTERVAL_SECONDS`) it runs a "tick":

```
Every 30 seconds:
  1. Load due tasks     — SELECT tasks where interval has elapsed
  2. Count in-flight    — per concurrency group
  3. For each due task:
     a. Concurrency check  — skip if group is at max
     b. Run pre_query      — skip if no rows returned
     c. Merge pre_query results into input_data
     d. Publish trigger message to target_topic
     e. Update last_triggered_at
```

---

## When Does a Task Fire?

A task is "due" when:

```sql
enabled = true
AND (last_triggered_at IS NULL
     OR last_triggered_at + interval_seconds <= NOW())
```

That's it. The scheduler doesn't check `last_completed_at` when deciding whether to fire — only `last_triggered_at` and `interval_seconds`. After firing, `last_triggered_at` is set to NOW(), so the task won't fire again until the interval elapses.

`last_completed_at` is only used for concurrency group tracking (see below).

---

## Pre-Queries

A `pre_query` is SQL that runs before each trigger. It serves two purposes:

**Dynamic input** — column values from the first row are merged into `input_data`. For example, finding the next site that needs work and passing its `site_id`.

**Gating** — if the query returns no rows, the task is skipped for this tick.

### Pre-Query as the Actual Work (CTE-only tasks)

Some tasks don't need a Kafka message at all — the pre_query does the work directly via CTEs:

```sql
WITH reset AS (
    UPDATE site_work_items
    SET status = 'triaged', claimed_by = NULL, claimed_at = NULL
    WHERE status = 'claimed' AND claimed_at < NOW() - INTERVAL '10 minutes'
    RETURNING id
)
SELECT COALESCE(COUNT(*)::text, '0') as reset_count FROM reset
```

The UPDATE runs inside the CTE. The SELECT at the end returns a row so the scheduler knows it executed. A Kafka message is still sent (the scheduler doesn't distinguish), but the target agent treats it as a no-op orchestrate.

### The "Always Return a Row" Rule

**If the pre_query does work via CTEs, it must always return at least one row.** Otherwise:

1. CTE runs the UPDATE/DELETE (work is done)
2. SELECT returns no rows (e.g. because of a `HAVING` clause)
3. Scheduler sees "no rows" → skips the task → never updates `last_triggered_at`
4. Task fires again next tick → CTE runs again (idempotent, no harm)
5. But `last_triggered_at` never advances → task always fires every tick
6. And `last_completed_at` never gets set → concurrency group slot stays occupied

Use `COALESCE(COUNT(*)::text, '0')` or a plain `SELECT 'done'` to guarantee a row.

**Bad:**
```sql
SELECT COUNT(*) FROM deleted_items HAVING COUNT(*) > 0
-- Returns nothing when 0 items deleted → scheduler skips → timestamps stuck
```

**Good:**
```sql
SELECT COALESCE(COUNT(*)::text, '0') as deleted_count FROM deleted_items
-- Always returns a row, even when count is 0
```

---

## Concurrency Groups

Tasks in the same `concurrency_group` share a `max_concurrent` limit. This prevents resource contention — e.g. two dispatch loops running simultaneously for different sites.

### How In-Flight Is Counted

A task counts as "in-flight" when ALL of these are true:

```sql
enabled = true
AND concurrency_group IS NOT NULL
AND last_triggered_at IS NOT NULL
AND (last_completed_at IS NULL OR last_completed_at < last_triggered_at)
AND last_triggered_at + timeout_seconds > NOW()
```

In plain English: the task has been triggered, has NOT completed since that trigger, and has NOT timed out.

### The Three Ways a Task Stops Being In-Flight

1. **Completion** — `last_completed_at` is updated to a value >= `last_triggered_at`. The agent's workflow must do this explicitly (e.g. `notify_scheduler` step).

2. **Timeout** — `last_triggered_at + timeout_seconds` passes. The scheduler considers the task done (whether it actually finished or not). This is the safety valve for crashed agents.

3. **Manual reset** — You UPDATE `last_completed_at = NOW()` in the database.

### The Group Starvation Problem

If a task in a group is permanently in-flight (never completes, timeout keeps being reset by re-triggering), it blocks all other tasks in the same group.

This happens when:
- `interval_seconds` < `timeout_seconds` AND the task re-triggers before timing out
- The re-trigger resets `last_triggered_at`, which resets the timeout window
- The task never completes → the group slot is permanently occupied

**Prevention rules:**
1. Always ensure tasks update `last_completed_at` on ALL completion paths (success, idle, error)
2. Set `timeout_seconds` < `interval_seconds` so the timeout expires before the next trigger
3. Don't put unrelated tasks in the same group
4. Tasks whose work is done in pre_query CTEs should have a `notify_scheduler` equivalent — or just ensure the pre_query always returns a row

### Recommended Group Assignments

| Group | Purpose | max_concurrent | Tasks |
|-------|---------|---------------|-------|
| `dispatch` | Pipeline dispatch work | 2 | `build-pipeline-trigger`, `improvement-sweep` |
| `claim-management` | Stale claim cleanup | 1 | `claimed-item-timeout` |
| `maintenance` | Database/system maintenance | 1 | `database-cleanup` |
| `vet-data` | Veterinary data collection | 1 | `vet-batch-verify`, `vet-sweep-continue` |
| (none) | Independent tasks | — | `feasibility-recheck`, `stale-orchestration-reaper` |

**Rule: don't share groups between unrelated tasks.** The original `maintenance` group had both `claimed-item-timeout` and `database-cleanup`. When `database-cleanup` stalled, it blocked claim resets, which blocked the entire pipeline.

---

## last_completed_at — Who Updates It?

The scheduler sets `last_triggered_at` automatically. But `last_completed_at` must be set by the agent or the pre_query. The scheduler never sets it.

### For fire_message=true tasks (Kafka-triggered agents)

The agent's workflow must include a step that updates `last_completed_at`:

```json
"notify_scheduler": {
    "action": "query_database",
    "config": {
        "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = 'build-pipeline-trigger'",
        "output_format": "object"
    },
    "next_step": "complete",
    "description": "Tell scheduler this execution finished"
}
```

This must be on ALL completion paths — success, idle (nothing to do), and error. If any path skips it, the task stays in-flight until timeout.

### For CTE-only tasks (work done in pre_query)

The pre_query itself should update `last_completed_at` as part of its CTE chain:

```sql
WITH work AS (
    UPDATE site_work_items SET status = 'triaged' WHERE ...
    RETURNING id
),
mark_complete AS (
    UPDATE scheduled_tasks SET last_completed_at = NOW()
    WHERE name = 'claimed-item-timeout'
)
SELECT COALESCE(COUNT(*)::text, '0') FROM work
```

Or simpler: if `timeout_seconds` is set shorter than `interval_seconds`, the in-flight count clears itself via timeout. But explicit completion is more reliable.

---

## The fire_message Column

The `scheduled_tasks` table has a `fire_message` boolean column. **The scheduler Go code does not currently read this column.** It always sends a Kafka message after a successful pre_query.

For CTE-only tasks (`claimed-item-timeout`, `database-cleanup`), the Kafka message goes to `system.agent.generic.requests` where the chassis processes it as a no-op orchestrate. This is wasteful but harmless.

A future scheduler update could skip the Kafka publish when `fire_message = false`. Until then, the column serves as documentation of intent.

---

## Task Lifecycle Diagram

```
                          ┌──────────────────────┐
                          │  scheduled_tasks row  │
                          │  enabled = true       │
                          └──────────┬───────────┘
                                     │
                    ┌────────────────┴────────────────┐
                    │ Is interval elapsed?              │
                    │ last_triggered_at + interval <= NOW │
                    └────────┬──────────┬─────────────┘
                             │          │
                           Yes          No → wait for next tick
                             │
                    ┌────────┴────────────────────────┐
                    │ Concurrency check                 │
                    │ Is group at max_concurrent?        │
                    └────────┬──────────┬─────────────┘
                             │          │
                            No          Yes → skip this tick
                             │
                    ┌────────┴────────────────────────┐
                    │ Run pre_query (if configured)     │
                    │ Returns rows?                     │
                    └────────┬──────────┬─────────────┘
                             │          │
                           Yes          No → skip this tick
                             │
                    ┌────────┴────────────────────────┐
                    │ Merge pre_query results           │
                    │ Publish Kafka message             │
                    │ Update last_triggered_at = NOW    │
                    └────────┬────────────────────────┘
                             │
                    ┌────────┴────────────────────────┐
                    │ Agent processes the message       │
                    │ On completion:                    │
                    │   UPDATE last_completed_at = NOW  │
                    └──────────────────────────────────┘
```

---

## Current Tasks Reference

### build-pipeline-trigger

| Field | Value |
|-------|-------|
| Interval | 60s |
| Group | dispatch |
| max_concurrent | 2 |
| Timeout | 300s |
| fire_message | true |
| Agent | build-pipeline-trigger → spawns build-dispatch-loop |

**Pre-query:** Seeds build queue from `build_queue` table, finds a site with dispatchable items (no active claims).

**What it does:** Finds one site with triaged work items, spawns a dispatch loop to process them. The dispatch loop claims, spawns handlers, and marks items complete.

**Completion:** The `build-pipeline-trigger` workflow has `notify_scheduler` on both success and idle paths.

### improvement-sweep

| Field | Value |
|-------|-------|
| Interval | 600s |
| Group | dispatch |
| max_concurrent | 2 |
| Timeout | 300s |
| fire_message | true |
| Agent | improvement-loop |

**Pre-query:** Finds deployed sites not currently being worked on. Round-robins through sites by `last_built_at`.

**What it does:** Runs quality, design, and completeness discovery checks. Triages findings into work items. Dispatches fixes via build-dispatch-loop. Triggers rerender after fixes.

### claimed-item-timeout

| Field | Value |
|-------|-------|
| Interval | 120s |
| Group | claim-management |
| max_concurrent | 1 |
| Timeout | 60s |
| fire_message | false |
| Agent | (CTE-only, no agent) |

**Pre-query (CTE):** Resets work items that have been `status = 'claimed'` for more than 10 minutes. Items at max attempts go to `failed`, others go back to `triaged`.

**Why timeout < interval:** The CTE runs in milliseconds. Setting `timeout_seconds = 60` (less than `interval_seconds = 120`) ensures the in-flight count always clears before the next trigger, preventing group starvation.

### database-cleanup

| Field | Value |
|-------|-------|
| Interval | 21600s (6 hours) |
| Group | maintenance |
| max_concurrent | 1 |
| Timeout | 120s |
| fire_message | false |
| Agent | (CTE-only, no agent) |

**Pre-query (CTE):** Deletes old agent_error_log entries, audit trail rows, completed/failed orchestration_states, and stale stuck orchestrations.

**Pre_query must always return a row** — no `HAVING` clause. Otherwise `last_completed_at` is never set and the group slot stays occupied.

### feasibility-recheck

| Field | Value |
|-------|-------|
| Interval | 600s |
| Group | (none) |
| Timeout | 60s |
| fire_message | false |

**Pre-query (CTE):** Promotes `blocked` work items to `triaged` when their handler_agent now exists in `agent_definitions`. Checks if agents have been deployed since the item was blocked.

### vet-batch-verify / vet-sweep-continue

| Field | Value |
|-------|-------|
| Interval | 900s / 1800s |
| Group | vet-data |
| max_concurrent | 1 |
| Timeout | 900s |
| fire_message | true |

**What they do:** Drive the veterinary data collection pipeline. `vet-batch-verify` processes verification tasks. `vet-sweep-continue` resumes area sweeps.

### stale-orchestration-reaper

| Field | Value |
|-------|-------|
| Interval | 300s |
| Group | (none) |
| Timeout | 60s |
| fire_message | false |

**Pre-query (CTE):** Fails orchestrations stuck in `EXECUTING_STEP` or `WAITING_FOR_RESPONSE` for more than 24 hours. These are leftovers from pod restarts whose topics were cleaned up.

---

## Creating a New Task — Checklist

### 1. Decide: CTE-only or agent-triggered?

**CTE-only** (pre_query does all the work, no agent needed):
- Database cleanup, timeout resets, status promotions
- Set `fire_message = false`
- Ensure the SELECT always returns a row (`COALESCE(COUNT(*)::text, '0')`)
- Set `timeout_seconds < interval_seconds`

**Agent-triggered** (Kafka message triggers an agent workflow):
- Set `fire_message = true`
- The agent's workflow MUST update `last_completed_at` on ALL paths
- Set `timeout_seconds` >= expected agent execution time

### 2. Choose a concurrency group

- If the task is independent, leave `concurrency_group` NULL
- If it competes for resources with another task, share a group
- **Never group unrelated tasks** — a stalled cleanup should not block claim resets
- Set `max_concurrent` to the number of simultaneous instances allowed

### 3. Set timeout correctly

| Scenario | Rule |
|----------|------|
| CTE-only task | `timeout_seconds` < `interval_seconds` |
| Fast agent (< 2 min) | `timeout_seconds` = 120–300 |
| Slow agent (dispatch loops, LLM work) | `timeout_seconds` >= typical completion time |
| Never: | `timeout_seconds` = 0 (means immediately times out) |
| Never: | `timeout_seconds` >> `interval_seconds` without completion tracking |

### 4. Write the pre_query

```sql
-- Pattern: gate on a condition, return dynamic input
SELECT s.id::text as site_id, s.domain
FROM sites s
WHERE s.status = 'deployed'
  AND ... your condition ...
ORDER BY ... your priority ...
LIMIT 1

-- Pattern: CTE does work, always returns a row
WITH work_done AS (
    UPDATE ... WHERE ... RETURNING id
)
SELECT COALESCE(COUNT(*)::text, '0') as affected_count FROM work_done
```

### 5. Insert the task

```sql
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type,
    target_topic, enabled, fire_message, timeout_seconds,
    concurrency_group, max_concurrent, pre_query
) VALUES (
    'my-new-task',
    'What this task does and why',
    300,            -- every 5 minutes
    'my-agent',     -- target agent type
    'system.agent.generic.requests',  -- almost always this
    true,           -- enabled
    true,           -- true for agent-triggered, false for CTE-only
    120,            -- timeout in seconds
    'my-group',     -- or NULL for independent
    1,              -- max concurrent in group
    'SELECT ...'    -- or NULL if no gating needed
);
```

---

## Debugging

### Task isn't firing

```sql
-- 1. Is it enabled?
SELECT name, enabled FROM scheduled_tasks WHERE name = 'my-task';

-- 2. Is the interval elapsed?
SELECT name, last_triggered_at,
       last_triggered_at + (interval_seconds || ' seconds')::interval as next_due,
       NOW() > last_triggered_at + (interval_seconds || ' seconds')::interval as is_due
FROM scheduled_tasks WHERE name = 'my-task';

-- 3. Is a concurrency group blocking it?
SELECT concurrency_group, COUNT(*) as in_flight,
       array_agg(name) as tasks
FROM scheduled_tasks
WHERE enabled = true
  AND concurrency_group IS NOT NULL
  AND last_triggered_at IS NOT NULL
  AND (last_completed_at IS NULL OR last_completed_at < last_triggered_at)
  AND last_triggered_at + (timeout_seconds || ' seconds')::interval > NOW()
GROUP BY concurrency_group;

-- 4. Is the pre_query returning no rows?
-- Run the pre_query manually and check.
```

### Task fires but nothing happens

```sql
-- Check if the agent exists
SELECT type, status FROM agent_definitions
WHERE type = 'my-agent' AND deleted_at IS NULL;

-- Check agent_error_log for failures
SELECT * FROM agent_error_log
WHERE agent_type = 'my-agent'
ORDER BY occurred_at DESC LIMIT 5;
```

### Concurrency group stuck

```sql
-- Find which task is blocking the group
SELECT name, concurrency_group, last_triggered_at, last_completed_at,
       timeout_seconds,
       last_triggered_at + (timeout_seconds || ' seconds')::interval as times_out_at
FROM scheduled_tasks
WHERE concurrency_group = 'the-stuck-group'
  AND last_triggered_at IS NOT NULL
  AND (last_completed_at IS NULL OR last_completed_at < last_triggered_at)
ORDER BY last_triggered_at;

-- Force-unstick: mark as completed
UPDATE scheduled_tasks
SET last_completed_at = NOW()
WHERE name = 'the-blocking-task';
```

### Manual kick

```sql
-- Force a task to fire on the next tick
UPDATE scheduled_tasks
SET last_triggered_at = NOW() - (interval_seconds || ' seconds')::interval
WHERE name = 'my-task';

-- Or just reset the completion to unblock the group
UPDATE scheduled_tasks
SET last_completed_at = NOW()
WHERE name = 'my-task';
```

---

## Known Issues & Future Work

**`fire_message` column is not read by the scheduler.** The Go struct doesn't include it. Every task sends a Kafka message. CTE-only tasks waste a message to generic.requests where it runs a no-op orchestrate. Low priority fix — functionally harmless.

**No-refire guard confusion.** The term "no-refire guard" was used in earlier docs to describe concurrency group blocking. There is no explicit refire guard — just the interval-based timing. The apparent "guard" was a concurrency group slot being occupied by an uncompleted task.

**`last_completed_at` is fragile.** If an agent crashes before updating it, the group slot stays occupied until timeout. The `notify_scheduler` step is the only mechanism, and it's easy to miss on error paths. A more robust approach would be for the scheduler to track completion via the `orchestration_states` table status (COMPLETED/FAILED) rather than requiring agents to explicitly update a timestamp.

**Concurrency counting doesn't handle re-trigger correctly.** When a task re-triggers before its timeout expires, `last_triggered_at` resets, which resets the timeout window. If the task never completes and `interval < timeout`, the group slot is permanently occupied. Setting `timeout < interval` for CTE-only tasks prevents this, but it's a trap for agent-triggered tasks that can run longer than their interval.

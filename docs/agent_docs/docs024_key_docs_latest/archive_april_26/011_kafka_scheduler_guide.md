# 011 — Kafka Scheduler

Standalone service that reads the `scheduled_tasks` table and publishes trigger messages to Kafka on configured intervals. It replaces ad-hoc cron jobs and manual triggering with a database-driven, concurrency-aware scheduling system.

---

## Why It Exists

Several agents need periodic triggering but have no caller. `build-pipeline-trigger` finds sites with pending work items and kicks off the dispatch loop — but nothing invokes it. `improvement-loop` runs discovery checks and queues fixes — but only fires when manually triggered or as a post-build step. The scheduler closes this gap by acting as the heartbeat for the entire system.

Adding a new schedule is an INSERT into `scheduled_tasks`. No code changes, no redeployment.

---

## Architecture

The scheduler is a single Go binary (`cmd/scheduler/main.go`) that runs as a 1-replica Kubernetes Deployment. It has no Kafka consumer — it only produces messages. It connects to:

- **Postgres** (clients_db) — reads `scheduled_tasks`, runs pre-queries, updates timestamps
- **Kafka** — publishes trigger messages using the existing `platform/kafka` producer

It ticks every 30 seconds (configurable via `TICK_INTERVAL_SECONDS`). Each tick it queries for tasks whose `interval_seconds` has elapsed since `last_triggered_at`, checks concurrency constraints, and fires any that are due.

### Tick Loop

```
every 30s:
  1. Load due tasks from scheduled_tasks (interval elapsed, enabled)
  2. Count in-flight tasks per concurrency group
  3. For each due task:
     a. Check concurrency group — skip if at max_concurrent
     b. Run pre_query if configured — skip if no rows returned
     c. Merge pre_query results into input_data
     d. Publish trigger message to target_topic
     e. Update last_triggered_at
```

### Double-Fire Prevention

Postgres `last_triggered_at` is checked atomically in the query:

```sql
WHERE enabled = true
  AND (last_triggered_at IS NULL
       OR last_triggered_at + (interval_seconds || ' seconds')::interval <= NOW())
```

Even if two scheduler instances briefly overlap during a rolling update, the 30-second tick interval plus the `last_triggered_at` update means the worst case is a task firing twice in quick succession — which the agents themselves handle via idempotent work item deduplication.

---

## The `scheduled_tasks` Table

Created by migration `066_kafka_scheduler.sql`.

| Column | Type | Purpose |
|---|---|---|
| `id` | UUID | Primary key |
| `name` | TEXT UNIQUE | Human-readable identifier, used in orchestration names |
| `description` | TEXT | What this schedule does |
| `interval_seconds` | INT | How often to trigger (seconds) |
| `target_agent_type` | TEXT | Agent type to include in the trigger message body |
| `target_topic` | TEXT | Kafka topic to publish to (default: `system.agent.generic.requests`) |
| `input_data` | JSONB | Static input data merged into the message body |
| `pre_query` | TEXT | Optional SQL query run before each trigger (see below) |
| `concurrency_group` | TEXT | Tasks sharing a group respect `max_concurrent` across the group |
| `max_concurrent` | INT | Max in-flight tasks in this group (default: 1) |
| `enabled` | BOOLEAN | Toggle without deleting |
| `last_triggered_at` | TIMESTAMPTZ | Set by scheduler after each trigger |
| `last_completed_at` | TIMESTAMPTZ | Reserved for future use (agents can update this on completion) |
| `timeout_seconds` | INT | How long a task is considered "in-flight" for concurrency counting |

---

## Pre-Queries

A `pre_query` is a SQL SELECT that runs before each trigger. It serves two purposes:

1. **Dynamic input** — column values from the first row are merged into `input_data`. For example, finding the next site that needs checking and passing its `site_id` and `domain` to the agent.

2. **Gating** — if the query returns no rows, the task is skipped for this tick. This prevents triggering when there's nothing to do.

Example pre_query for `improvement-sweep`:

```sql
SELECT s.id::text as site_id, s.domain
FROM sites s
WHERE s.status IN ('active', 'deployed')
  AND NOT EXISTS (
    SELECT 1 FROM site_work_items wi
    WHERE wi.site_id = s.id AND wi.status = 'claimed'
  )
ORDER BY s.last_built_at ASC NULLS FIRST
LIMIT 1
```

This finds the least-recently-built site with no active claims. If all sites have claimed items (meaning they're already being worked on), the query returns nothing and the sweep is skipped.

---

## Concurrency Groups

Tasks in the same `concurrency_group` share a `max_concurrent` limit. A task is considered "in-flight" when `last_triggered_at` is set and either `last_completed_at` is null / older than `last_triggered_at`, and `timeout_seconds` hasn't elapsed.

The two seed schedules both belong to the `dispatch` group with `max_concurrent = 2`:

| Task | Interval | Group | Max Concurrent |
|---|---|---|---|
| `build-pipeline-trigger` | 120s | dispatch | 2 |
| `improvement-sweep` | 600s | dispatch | 2 |

This means at most 2 dispatch-related orchestrations run at once across both schedules. If `build-pipeline-trigger` fires and the previous run is still processing, the second slot is still available for `improvement-sweep` or a second pipeline trigger. If both slots are occupied, tasks wait until the next tick.

The `timeout_seconds` acts as a safety valve — if an agent hangs or crashes without completing, the in-flight count resets after the timeout and the scheduler retries.

---

## Message Format

The scheduler publishes messages in the same format as `CallAgentAction` and `DispatchAreaDiscoverersAction`. This means existing agents process them without any changes.

**Headers:**

```
correlation_id:     <new UUID>
request_id:         <new UUID>
message_id:         <new UUID>
orchestration_id:   <new UUID>
orchestration_name: sched-<task-name>-<timestamp>
step_name:          start
message_type:       request
action:             orchestrate
from_agent_type:    kafka-scheduler
from_agent_id:      kafka-scheduler-singleton
responses_topic:    system.scheduler.responses
```

**Body:**

```json
{
  "action": "orchestrate",
  "config": {
    "agent_type": "<target_agent_type>"
  },
  "input_data": { ... merged static + pre_query data ... }
}
```

The `responses_topic` points to `system.scheduler.responses`. Agent responses land there rather than polluting other agent topics. Nothing consumes this topic yet — it's there for future observability (tracking completion times, failure rates).

---

## Adding a New Schedule

Insert a row. No code changes needed.

**Simple periodic trigger** (e.g. data scraper every 30 minutes):

```sql
INSERT INTO scheduled_tasks (name, description, interval_seconds, target_agent_type, concurrency_group, max_concurrent)
VALUES (
    'competitor-scraper',
    'Scrapes competitor sites for pricing and content changes',
    1800,
    'competitor-scraper-agent',
    'scraping',
    3
);
```

**With dynamic input from pre_query** (e.g. rotating through sites):

```sql
INSERT INTO scheduled_tasks (name, description, interval_seconds, target_agent_type, concurrency_group, max_concurrent, pre_query)
VALUES (
    'content-freshness-check',
    'Checks one site at a time for stale content',
    900,
    'content-freshness-agent',
    'scraping',
    3,
    'SELECT s.id::text as site_id, s.domain
     FROM sites s
     WHERE s.status = ''deployed''
       AND (s.content_data->>''last_freshness_check'')::timestamptz < NOW() - interval ''7 days''
     ORDER BY (s.content_data->>''last_freshness_check'')::timestamptz ASC NULLS FIRST
     LIMIT 1'
);
```

**Disabling temporarily:**

```sql
UPDATE scheduled_tasks SET enabled = false WHERE name = 'competitor-scraper';
```

**Changing frequency:**

```sql
UPDATE scheduled_tasks SET interval_seconds = 3600 WHERE name = 'competitor-scraper';
```

---

## Deployment

| Component | Location |
|---|---|
| Go source | `cmd/scheduler/main.go` |
| Dockerfile | `build/docker/backend/kafka-scheduler.dockerfile` |
| Kustomize base | `deployments/kustomize/services/kafka-scheduler/base/` |
| Production overlay | `deployments/kustomize/services/kafka-scheduler/overlays/production/uk_001/` |
| Terraform | `deployments/terraform/environments/production/uk001/services/agents/2270-kafka-scheduler/` |
| SQL migration | `066_kafka_scheduler.sql` |

**Environment variables:**

| Variable | Required | Default | Purpose |
|---|---|---|---|
| `CLIENTS_DATABASE_URL` | Yes | — | Postgres connection string |
| `KAFKA_BROKERS` | Yes | — | Comma-separated broker list |
| `TICK_INTERVAL_SECONDS` | No | 30 | How often to check for due tasks |
| `LOG_LEVEL` | No | info | debug, info, warn, error |

**Resources:** 25m CPU / 64Mi memory requests. The scheduler does very little work — one DB query and zero-to-few Kafka publishes per tick.

**Health endpoints:** `/health` (liveness), `/ready` (checks DB connectivity).

---

## Observability

The scheduler logs at INFO level for each triggered task:

```json
{
  "level": "info",
  "msg": "Triggered task",
  "task": "build-pipeline-trigger",
  "agent_type": "build-pipeline-trigger",
  "topic": "system.agent.generic.requests"
}
```

At DEBUG level it logs skipped tasks (concurrency at max, pre_query returned no rows).

To see what's scheduled and when it last ran:

```sql
SELECT name, interval_seconds, enabled,
       last_triggered_at,
       last_triggered_at + (interval_seconds || ' seconds')::interval AS next_due,
       concurrency_group
FROM scheduled_tasks
ORDER BY name;
```

To check if concurrency is blocking triggers:

```sql
SELECT concurrency_group, COUNT(*) as in_flight
FROM scheduled_tasks
WHERE enabled = true
  AND concurrency_group IS NOT NULL
  AND last_triggered_at IS NOT NULL
  AND (last_completed_at IS NULL OR last_completed_at < last_triggered_at)
  AND last_triggered_at + (timeout_seconds || ' seconds')::interval > NOW()
GROUP BY concurrency_group;
```
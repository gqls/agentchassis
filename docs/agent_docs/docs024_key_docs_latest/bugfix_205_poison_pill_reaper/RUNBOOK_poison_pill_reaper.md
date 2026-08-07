# RUNBOOK — bug 205

DB access:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## Is the loop alive?

```sql
-- non-zero rows with ok=0 means still burning; NO rows means the sweep stopped
-- (different finding) — never read absence as success
SELECT date_trunc('hour',created_at) hr, count(*), count(*) FILTER (WHERE success) ok
  FROM llm_call_log WHERE step_name='extract_and_reconcile'
   AND created_at > now() - interval '8 hours' GROUP BY 1 ORDER BY 1 DESC;
```

## Who is looping (task identity, not counts)

```sql
SELECT os.collected_data->'input_data'->>'task_id' task_id, count(*) dispatches,
       count(*) FILTER (WHERE os.status='FAILED') failed
  FROM orchestration_states os
 WHERE os.owner_agent_type='vet-practice-verifier'
   AND os.created_at > now() - interval '24 hours'
 GROUP BY 1 HAVING count(*) > 1 ORDER BY 2 DESC;
```

## Read the reaper's live pre_query (no repo seed exists — the row IS the source)

```sql
SELECT pre_query FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
```

Gotcha: the reaper's work happens in the pre_query CTEs regardless of whether the
HAVING emits a row; `fire_message` only controls the follow-on message.

## Task states

```sql
SELECT status, count(*) FROM business_intel.collection_tasks
 WHERE task_type='initial_verification' GROUP BY status;
-- parked rows after the fix:
SELECT id, retry_count, left(error_message,80) FROM business_intel.collection_tasks
 WHERE status='failed';
```

Gotcha: `started_at` is `timestamp without time zone` in UTC (DB tz is UTC); local
wall-clock is BST — check `SELECT now()` before calling a row stale.

## Un-park a task deliberately (operator action, after fixing its cause)

```sql
UPDATE business_intel.collection_tasks
   SET status='pending', retry_count=0, error_message=NULL, scheduled_for=NULL
 WHERE id='<task-id>' AND status='failed';
```

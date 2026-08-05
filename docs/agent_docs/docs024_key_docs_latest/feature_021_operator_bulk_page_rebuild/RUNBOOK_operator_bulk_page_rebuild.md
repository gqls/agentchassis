# RUNBOOK — operator bulk page rebuild (`features_open/021`)

## The script

`scripts/rebuild_pages.sh <domain> <page1,page2,...> "<reason>"`

Defaults to `DRY_RUN=1` (safe — no DB write, no dispatch, local report only).
Set `DRY_RUN=0` to actually queue and dispatch. See the script's own header
comment for the full env var list (`PRIORITY`, `REQUESTED_BY`, `INTENT`) and
the three things it is NOT yet wired to do (intent enforcement, a true preview
of the operator's own page list, a dedicated Kafka topic).

## Pre-flight, every time (the script prints these itself, but know what they mean)

```sql
-- how many OTHER pages already sit needs_rebuild on this site — these ride
-- along on a REAL (DRY_RUN=0) dispatch, not just the pages you named
SELECT name, updated_at FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = '<domain>')
  AND build_status = 'needs_rebuild'
ORDER BY updated_at;

-- existing pending/claimed page_rebuild tasks for this site (yours will union
-- with these, not run separately)
SELECT id, reason, requested_by, payload, status, created_at
FROM maintenance_queue
WHERE site_id = (SELECT id FROM sites WHERE domain = '<domain>')
  AND task_type = 'page_rebuild' AND status IN ('pending','claimed');
```

## After a REAL dispatch (`DRY_RUN=0`)

```sql
-- orchestration state, by correlation (never by created_at)
SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - updated_at))::int AS since_s
FROM orchestration_states WHERE correlation_id = '<CORRELATION_ID>'::uuid ORDER BY created_at;

-- the queue row's own lifecycle: pending -> claimed -> complete
SELECT status, claimed_at, completed_at, result, error_message
FROM maintenance_queue WHERE id = '<TASK_ID>';

-- did the pages actually flag / rebuild
SELECT name, build_status, updated_at FROM pages
WHERE site_id = '<SITE_ID>' AND name = ANY(string_to_array('<pages_csv>', ','));

-- watch it run live (page-rebuild's own step timeout is 5400s)
kubectl -n ai-persona-system logs -f -l agent-type=page-rebuild --tail=200 | grep '<CORRELATION_ID>'
```

**Do not trust a clean-looking dispatch alone** — this repo's own landmine
(hit twice already in the `bugfix_154`/`bugs_open/178` lane, 2026-08-04):
`kcat -P` can exit 0 having silently dropped the message. Check for an
`orchestration_states` row within a few minutes, not just once.

## Mechanism grounding queries (re-run before trusting any of this — 11 days is
## a long time on this repo; these were last verified 2026-08-05)

```sql
-- the dormant road, still dormant
SELECT task_type, status, count(*), max(created_at) FROM maintenance_queue GROUP BY 1,2;
SELECT type, is_active FROM agent_definitions WHERE type='maintenance-triage' AND deleted_at IS NULL;
SELECT name FROM scheduled_tasks WHERE target_agent_type='maintenance-triage';  -- expect 0 rows

-- the reaper predicate bugs_closed/070 fixed — confirm it keys on updated_at, not created_at
SELECT pre_query FROM scheduled_tasks WHERE name='stale-work-item-reaper';

-- the claim cadence a resurrected site_work_items row would have raced (N/A to this path, kept for context)
SELECT interval_seconds FROM scheduled_tasks WHERE name='build-pipeline-trigger';

-- the full maintenance-triage workflow, to re-check nothing has changed
SELECT jsonb_pretty(default_config->'workflow'->'steps') FROM agent_definitions
WHERE type='maintenance-triage' AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;

-- the full page-rebuild workflow
SELECT jsonb_pretty(default_config->'workflow'->'steps') FROM agent_definitions
WHERE type='page-rebuild' AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
```

## DB access

`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

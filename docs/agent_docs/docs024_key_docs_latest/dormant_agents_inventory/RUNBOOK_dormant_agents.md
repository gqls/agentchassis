# RUNBOOK — dormant-agents inventory

DB shell:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## Reproduce the measurement (no code needed — this is what the action runs)

```sql
WITH observed_steps AS (
  SELECT DISTINCT jsonb_object_keys(workflow_plan->'steps') AS step
  FROM orchestration_states
  WHERE workflow_plan ? 'steps' AND jsonb_typeof(workflow_plan->'steps')='object'
),
agent_steps AS (
  SELECT a.type, jsonb_object_keys(a.default_config#>'{workflow,steps}') AS step
  FROM agent_definitions a
  WHERE a.is_active AND a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false
    AND jsonb_typeof(a.default_config#>'{workflow,steps}')='object'
),
fingerprints AS (            -- step keys unique to ONE agent
  SELECT step, min(type) AS type FROM agent_steps
  GROUP BY step HAVING count(DISTINCT type)=1
),
agent_fp AS (
  SELECT f.type, count(*) u,
         count(*) FILTER (WHERE f.step IN (SELECT step FROM observed_steps)) o,
         min(f.step) sample
  FROM fingerprints f GROUP BY f.type
)
SELECT f.type, u AS unique_steps, sample,
       round(extract(epoch FROM (now()-min(a.created_at)))/86400.0,1) AS age_days,
       count(a.*) AS active_rows
FROM agent_fp f
JOIN agent_definitions a ON a.type=f.type
  AND a.is_active AND a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false
WHERE f.o=0                  -- never observed
GROUP BY f.type, u, sample ORDER BY age_days DESC;
```

GOTCHA: `owner_agent_type` is NOT a valid signal (95k+ rows are `'generic'`).
Never count runs that way — it reports fix-proposer/council-gate as never-run.

GOTCHA: only TOP-LEVEL `workflow_plan->'steps'` keys count as "observed". An
agent that runs only via a council/subtree path whose steps never surface as
top-level keys (feature-designer) reads as never-observed. That is a triage
signal, not a verdict.

## Verify the action is in the running chassis (before applying the seed)

```
kubectl -n ai-persona-system get pods | grep agent-chassis
kubectl -n ai-persona-system exec <chassis-pod> -- sh -c 'grep -ac diagnose_dormant_agents /proc/1/exe'
# must be >= 1. Also confirm the image tag is the one you built:
kubectl -n ai-persona-system get pod <chassis-pod> -o jsonpath='{.spec.containers[0].image}'
```

## Apply the seed (AFTER the image is live)

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  < docs/agent_docs/docs024_key_docs_latest/dormant_agents_inventory/seed_diagnosis_dormant_agents.sql
```

## Read the latest report

```sql
SELECT body FROM doc_notes WHERE categories ? 'dormant-agents'
ORDER BY created_at DESC LIMIT 1;
```

## See what it emitted (after dry_run is flipped off)

```sql
SELECT item_key, status, summary
FROM site_work_items
WHERE created_by='diagnosis-dormant-agents' AND item_type='dormant_agent'
ORDER BY updated_at DESC;
```

## Flip out of dry-run / raise the emit cap

See the tail of `seed_diagnosis_dormant_agents.sql` for the exact `jsonb_set`
updates (DB-live, no image roll needed).

## Trigger a manual sweep

The agent is `diagnosis-dormant-agents`, manual-trigger, one `sweep` step. Use
the same manual dispatch path the owner uses for `diagnosis-silent-check` /
`diagnosis-triage` (a `start_orchestration` against that agent type). It is a
task-mode workflow with a 120 s timeout; expect the report note within one run.

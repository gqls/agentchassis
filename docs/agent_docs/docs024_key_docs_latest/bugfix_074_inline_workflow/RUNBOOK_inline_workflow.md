# RUNBOOK — bugs_open/074, inline workflow silently ignored

Every command that had to be got right, with its gotcha attached. Change it HERE, not in
scrollback.

---

## Find every task carrying the trap (and the near-misses)

```sql
SELECT name, enabled, target_agent_type,
       jsonb_typeof(input_data->'config'->'workflow') AS inline_wf,
       (input_data ? 'action') AS has_action, (input_data ? 'input_data') AS has_inner_payload
FROM scheduled_tasks
WHERE input_data ?| array['action','config','input_data']
ORDER BY enabled DESC, name;
```

**Gotcha:** `?|` needs `array[...]`, not a comma list. Rows with `action`/`input_data` but no
workflow are the `bugs_closed/054` family — different bug, still legal, leave them alone.

## Prove a scheduled task's work actually happened

Never the timestamps. Both advance on a no-op, which is the whole bug.

```sql
-- did any run ever carry the action the task exists to run?
SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_evidence_base%';

-- what did the run actually report?
SELECT jsonb_pretty(collected_data->'refresh_result')
FROM orchestration_states
WHERE workflow_plan::text LIKE '%refresh_evidence_base%'
ORDER BY created_at DESC LIMIT 1;
```

**Gotcha:** `final_result` is empty on these runs — the payload is in
`collected_data->'<output_field>'`, the name the step's `output_field` declares. Reading
`final_result` and finding NULL says nothing about whether the work happened.

## Force a scheduled task now

```sql
UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = '<task>';
```

Fires within one 30s tick. Since `bugs_closed/030` gave cron its own lane, publish→run start on
`system.agent.scheduled.requests` is ~1s, not the old ~18 min — a missing orchestration row after
a minute is a real absence, not queue latency.

## Verify the constraint (migration 217) — the failing branch first

```sql
-- 1) MUST be refused: 23514, naming scheduled_tasks_no_inline_workflow
INSERT INTO scheduled_tasks (name, target_agent_type, target_topic, input_data)
VALUES ('_bug074_probe','generic','system.agent.scheduled.requests',
        '{"config":{"agent_type":"generic","workflow":{"start_step":"x","steps":{}}}}'::jsonb);

-- 2) POSITIVE CONTROL — MUST be accepted, or check 1 passed for the wrong reason
INSERT INTO scheduled_tasks (name, target_agent_type, target_topic, input_data, enabled)
VALUES ('_bug074_probe','generic','system.agent.scheduled.requests',
        '{"batch_size":20,"config":{"agent_type":"generic"}}'::jsonb, false);
DELETE FROM scheduled_tasks WHERE name = '_bug074_probe';

-- 3) nothing survives that violates it
SELECT name FROM scheduled_tasks WHERE input_data->'config' ? 'workflow';
```

**Gotcha:** insert the control with `enabled = false`. The column defaults to `true`, and a live
probe row on `system.agent.scheduled.requests` fires a real dispatch within 30s.

## Applying a migration without taking other threads' with it

```bash
./scripts/migration/run-migrations.sh                      # dry run + probe, per session
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/217_*.sql
./scripts/migration/run-migrations.sh --record-only 217_scheduled_tasks_reject_inline_workflow.sql \
  --note "applied by hand <date>; <what you checked>"
```

**Gotcha:** `--apply` runs **every** pending file, including other sessions' (13 were pending on
07-26). Hand-apply yours, then `--record-only`.

**Gotcha:** the runner's idempotency lint reads the file's **comment text**. A pasteable
`INSERT INTO <table>` in a header comment raises a false "unguarded insert" warning on a migration
that contains no INSERT at all. Keep pasteable probes in this RUNBOOK; describe them in the header.

## Turn the evidence sweep's writes on (and prove the flip landed)

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,refresh_evidence,config,dry_run}', 'false'::jsonb),
    updated_at = now()
WHERE type = 'evidence-freshness' AND deleted_at IS NULL;

SELECT default_config->'workflow'->'steps'->'refresh_evidence'->'config'->>'dry_run'
FROM agent_definitions WHERE type = 'evidence-freshness' AND deleted_at IS NULL;  -- must read: false
```

**Gotcha:** a dry run left on is indistinguishable from a healthy pass — green status, both
timestamps, a full report. Assert the flag, not the run.

**Gotcha:** `jsonb_set` will not create a missing intermediate path. It works here because
`config` already exists with `dry_run` in it; seeding `{"config": {}}` and expecting the path to
appear is a silent no-op.

## Watch the sweep's artefacts, not its status

```sql
SELECT s.domain, ss.created_by, ss.pinned, left(ss.notes,110)
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE ss.aspect='evidence_base' AND ss.is_current AND ss.created_by='evidence-refresher';

SELECT swi.item_type, swi.status, s.domain, left(swi.summary,80)
FROM site_work_items swi LEFT JOIN sites s ON s.id = swi.site_id
WHERE swi.item_type='stale_evidence' ORDER BY swi.created_at DESC;
```

**Gotcha:** the sweep **supersedes** the spec (`is_current=false` on the old row, INSERT a new
one — `refresh_evidence_base_action.go:669-693`). Any thread holding a `site_specs.id` for an
evidence base must re-SELECT the current row before writing, or its edit lands on a dead revision.

## After the next kafka-scheduler build — verify the WARN against the pod

```bash
kubectl -n ai-persona-system exec <kafka-scheduler-pod> -- \
  sh -c 'strings /app/kafka-scheduler | grep -c "cannot deliver"'        # the string the change CREATED
kubectl -n ai-persona-system exec <kafka-scheduler-pod> -- \
  sh -c 'strings /app/kafka-scheduler | grep -c "Pre-query found no rows"'  # positive control, pre-existing
```

**Gotcha:** grep a string the change *creates*, never one it merely uses — and pair it with a
control, so "0 hits" tells you the binary is stale rather than that the grep was wrong.

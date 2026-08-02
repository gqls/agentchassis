# RUNBOOK — vigilant designer + offer analyser

Commands that were hard to get right, with their gotchas. Update HERE when one changes.

## DB access

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## The improvement loop, hand-fired (the manual-trigger mode this programme runs in)

The per-site sweep script (read its blast-radius header first):

```
./run_improvement_sweep_once.sh <site-domain-or-id>     # repo root; check the header before first use
```

Gotchas already known:
- `scheduled_tasks.last_triggered_at` / `last_completed_at` keep advancing while nothing runs
  (fire-and-forget stamp). Measure liveness at `orchestration_states` (newest run for the
  agent), never at the task row.
- A wedged head orchestration freezes a dispatch lane until a pod roll (dispatch-queue lane).
- No orchestration dispatch within ~300s of a chassis pod (re)start — silently dropped.

## Watching a finding travel (the drain proof)

```sql
-- born
SELECT id, item_type, status, handler_agent, item_key, created_at
FROM site_work_items WHERE site_id='<id>' AND item_type='<type>' ORDER BY created_at DESC LIMIT 5;
-- promoted (only improvement-loop.triage_findings may do this — migration 286, single owner)
-- claimed/complete: watch status + claimed_by; a FAILED step can show COMPLETED with error NULL — read __step_error
```

## Verifying a deploy (image roll)

```
kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<symbol you ADDED>"'
# every replica; plus a NEGATIVE control (a string the change REMOVED, expect 0); grep -ic for case traps
```

Image before config, always: a check name in a checks array before the binary carries it is a
FATAL run (149 B4). An unregistered action name in a workflow is inert at best.

## Migrations (sql_for_agents)

- Dry-run per session and after every roll; `--apply` takes EVERY pending file — scope the dir.
- Take a snapshot before UPDATE-ing an agent definition (`bak_ad_<agent>_<date>` pattern —
  see `bak_ad_designdiscovery_20260727` precedent).
- Seed key is `start_step`, never `initial_step` (VIZ-012 lesson); VERIFY blocks read `start_step`.
- Next free migration number: check `ls docs/agent_docs/sql_for_agents/ | sort -V | tail` at
  write time — concurrent sessions take numbers hourly; 289 was in use on 2026-08-02.

## Council submissions (platform/ internal/ pkg/ changes)

```
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
# budget ~30 min (dispatch queues behind the fleet); find your run by payload, not printed id;
# commit with Council-Submitted: <corr> if committing before the verdict
```

## Single-owner audit (must stay clean after any loop/workflow edit)

```
./scripts/audit-single-owner-actions.sh
```

# RUNBOOK — scheduler group starvation (bugs_open/048)

DB access:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## Observe the starvation (before the fix)

```sql
SELECT name, concurrency_group, max_concurrent, fire_message, interval_seconds,
       last_triggered_at, now()-last_triggered_at AS since_last
FROM scheduled_tasks WHERE enabled = true ORDER BY last_triggered_at NULLS FIRST;
```
The four `maintenance` rows stuck at 2026-05-02; the eight healthy tasks in other
groups advancing normally.

Why `feasibility-recheck` returns no rows (steady state), so it never stamps:
```sql
SELECT count(*) FROM site_work_items WHERE status='blocked';   -- 0
```

Backlog `stale-work-item-reaper` should have cleared:
```sql
SELECT count(*) FROM site_work_items
WHERE status='triaged' AND pipeline='build'
  AND created_at < now() - interval '48 hours' AND claimed_at IS NULL;   -- was 1
```

Inspect the pre-queries (note the CTE `E'\\\\s+'` escaping under `psql -c`):
```sql
SELECT name, fire_message, left(regexp_replace(pre_query, E'\\s+', ' ', 'g'), 240)
FROM scheduled_tasks WHERE concurrency_group IN ('maintenance','thunder-lifecycle')
ORDER BY last_triggered_at;
```

## Build + roll the scheduler (its own binary — NOT a chassis build)

Fresh unique tag; do NOT edit makefile line 16 (shared, another thread's WIP). Pass
the tag on the command line:
```bash
make quick-scheduler-update IMAGE_TAG=v1.0.1146
```
`quick-scheduler-update` = build-kafka-scheduler (git archive HEAD) → docker push →
deploy-kafka-scheduler (seds the kafka-scheduler kustomization newTag, kubectl apply)
→ rollout restart. **Commit the code fix BEFORE building** — the build takes
committed HEAD and prints (and skips) any uncommitted change.

Tail: `make logs-scheduler`  (or `kubectl logs -f -n ai-persona-system -l app=kafka-scheduler`).

## Verify the deploy landed

```bash
kubectl -n ai-persona-system get pods -l app=kafka-scheduler \
  -o jsonpath='{range .items[*]}{.metadata.name}{"  "}{.spec.containers[0].image}{"\n"}{end}'
# expect ...kafka-scheduler:v1.0.1146
```
Pod-grep the new binary for a string the fix CREATED (discriminating, not one it
merely uses — the memory landmine):
```bash
POD=$(kubectl -n ai-persona-system get pods -l app=kafka-scheduler -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  'strings /app/kafka-scheduler | grep -c "task ran with nothing to do"'   # expect >= 1
```
Positive control (a string present before AND after): `grep -c "Triggered task"`.

## Verify the FIX behaviourally (the real proof)

The scheduler runs a tick immediately on startup, then every 30s. Within a minute or
two of the restart, all four `maintenance` tasks must leave May:
```sql
SELECT name, last_triggered_at, now()-last_triggered_at AS since_last
FROM scheduled_tasks WHERE concurrency_group='maintenance'
ORDER BY last_triggered_at;
```
Expected: all four `last_triggered_at` within the last few minutes, and — critically
— they keep advancing on their own intervals (`feasibility-recheck` every 600s even
though `blocked` stays 0). Confirm the reaper backlog drains to 0.

Also confirm `thunder-reaper` (group of one) now records liveness:
```sql
SELECT name, last_triggered_at FROM scheduled_tasks WHERE name='thunder-reaper';
```

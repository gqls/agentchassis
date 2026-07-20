# RUNBOOK — bugfix 003 (spawn loses child response)

Commands that were hard to get right, with their gotchas. Update HERE when one
changes.

## Live scale checks

Reaper kill counts by step (the bug's headline metric):
```sql
SELECT current_step, count(*) FROM orchestration_states
WHERE error LIKE 'reaper:%' AND created_at > now() - interval '2 days'
GROUP BY current_step ORDER BY count(*) DESC;
```

EXECUTING_STEP zombies (should stay at 0 now F1 is live; >4h is the reaper
threshold):
```sql
SELECT count(*) FROM orchestration_states
WHERE status='EXECUTING_STEP' AND last_activity < now() - interval '4 hours';
```

What F1 reaped (fixed prefix is the triage grouping key; step name follows):
```sql
SELECT left(error,70), count(*) FROM orchestration_states
WHERE error LIKE 'reaper: stale EXECUTING_STEP%' GROUP BY 1 ORDER BY 2 DESC;
```

## Dial-error rate (bug 040's metric)

```bash
kubectl -n ai-persona-system logs --since=12h --tail=3000 <pod> \
  | grep 'i/o timeout' | grep -o 'dial tcp [0-9.]*:9092' | sort | uniq -c
```
Gotcha: `for p in $(kubectl get pods -o name | head 6)` — `head 6` is a file
arg, you want `head -6`. Also the static chassis pod legitimately shows 0;
sample the spawned `agent-*` Job pods.

## Healthy-stint audit query (how the F1 threshold was chosen)

```sql
WITH t AS (
  SELECT orchestration_id, changed_at, old_status, new_status,
         lag(changed_at) OVER (PARTITION BY orchestration_id ORDER BY changed_at) AS prev_at,
         lag(new_status) OVER (PARTITION BY orchestration_id ORDER BY changed_at) AS prev_new
  FROM orchestration_state_audit)
SELECT count(*) FILTER (WHERE changed_at - prev_at > interval '3 hours'),
       max(EXTRACT(EPOCH FROM (changed_at - prev_at))/3600)
FROM t
WHERE old_status='EXECUTING_STEP' AND new_status NOT IN ('FAILED')
  AND prev_new='EXECUTING_STEP';
```
Gotcha: audit history starts 2026-05-28 — say "in 7.5 weeks of history", not
"never".

## Reaper config

Live row (config is live immediately; the seed file is only the mirror):
```sql
SELECT pre_query FROM scheduled_tasks WHERE name='stale-orchestration-reaper';
```
- Backup of the pre-F1 pre_query: scratchpad `f1/reaper_pre_query_BACKUP_2026-07-20.sql`
  (session-local); the seed file's earlier section also preserves the old text.
- `scheduled_tasks` has NO `is_active` column — it's `enabled`.
- Gotcha: `'x' || current_step` NULLs the whole string when current_step is
  NULL — always COALESCE inside reaper error strings.
- Mirror edits into `docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql`
  (append a dated section; seeds run in order) or the next re-seed clobbers
  the live row.

## Diagnosis-loop verdict fetch (for the filed structural claim)

Correlation `d971e8c2-0c41-4251-b46f-705b471f5dc1`, item_key
`needs_diagnosis:workflow-step-timeouts-on-awaited-child`:
```sql
SELECT kind, iteration, length(body) FROM diagnosis_artifacts
WHERE correlation_id='d971e8c2-0c41-4251-b46f-705b471f5dc1' ORDER BY created_at;
```
Gotcha: the diagnose-agent orchestration for this run wedged at step `route`
(EXECUTING_STEP, itself a 003-class casualty). If it never produced a verdict,
the intake item must be closed by hand (the 090 printout has the UPDATE) or it
blocks re-filing on the same slug.

## Deploy verification (when F2/F3/F4 images roll)

Discriminating pod-grep — grep a literal the change CREATES plus a positive
control, never a generic string the change merely uses:
```bash
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "<literal-created-by-change>"'
```

## Liveness restart test (bugs_open/003 F4) — how to run it properly

Manifest: `scratchpad/liveness_test_pod.yaml` pattern — a standalone Pod on the
chassis image with `KAFKA_BROKERS=10.255.255.1:9092` (unroutable TEST-NET),
production probe shape, and shortened windows
(`KAFKA_UNHEALTHY_AFTER_SECONDS=60`, `KAFKA_HEALTH_PROBE_INTERVAL_SECONDS=10`).

**Two gotchas that cost the first run:**
1. **Set `KAFKA_TOPIC` explicitly.** A custom `AGENT_TYPE` is NOT enough:
   `personae-prod-config` supplies `KAFKA_TOPIC`, and `main.go` prefers the env
   var over the agent-type-derived default, so the pod consumes
   `system.agent.generic.*`. With a fresh consumer group and the reader's
   `StartOffset: FirstOffset`, it then **replays the topic from the beginning**
   (11-day-old messages, in the observed run). A distinct consumer group means
   its own copies — it does not steal from the real group — but expect the
   replay and check afterwards for writes:
   ```sql
   SELECT count(*) FROM orchestration_states WHERE processing_node='<pod>';
   SELECT count(*) FROM processed_messages  WHERE processed_by='<pod>';
   ```
2. **A pass is a RESTART, not a JSON body.** Watch `restartCount` 0→1 AND
   `kubectl describe pod` events for
   `Liveness probe failed ... statuscode: 503`. If the container dies without
   that event, it died of something else (agent init) — that is not a pass.

```bash
kubectl apply -f <manifest>
# poll: restarts + endpoint together
kubectl -n ai-persona-system get pod chassis-liveness-test \
  -o jsonpath='{.status.containerStatuses[0].restartCount}'
kubectl -n ai-persona-system exec chassis-liveness-test -- \
  wget -qO- --timeout=4 localhost:8080/health
kubectl -n ai-persona-system describe pod chassis-liveness-test | grep -A12 '^Events:'
kubectl delete pod chassis-liveness-test -n ai-persona-system   # always clean up
```

**Prove the pod is really wedged before trusting a negative result** — the
first run's `{"status":"ok"}` was the check being wrong, not the pod being fine:
```bash
kubectl -n ai-persona-system exec <pod> -- wget -T 4 -qO- 10.255.255.1:9092   # must time out
kubectl -n ai-persona-system logs <pod> | grep -c 'i/o timeout'               # must be > 0
```

# RUNBOOK — chassis replica scaling

Commands that were hard to get right, with their gotchas. DB access is the
standard `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U
clients_user -d clients_db`.

## R1 — exposure to the §8.1 ownership drop, right now

Who is stamped on the in-flight orchestrations (the population at risk when
the pod is replaced):

```sql
SELECT status, processing_node, count(*), min(last_activity) AS oldest
FROM orchestration_states
WHERE status IN ('AWAITING_RESPONSES','EXECUTING_STEP','RUNNING')
GROUP BY 1, 2 ORDER BY 3 DESC;
```

**Gotcha:** at a quiet moment this returns ~1 row, and R2's grep then means
nothing. Run this BEFORE a deliberate roll to know whether the roll can even
exhibit the drop.

## R2 — did the ownership drop fire across a roll

```bash
kubectl -n ai-persona-system logs deploy/agent-chassis --since=2h \
  | grep -c "owned by different pod"
kubectl -n ai-persona-system get pods -l app=agent-chassis   # record pod AGE with the count
```

**Gotchas:** `logs deploy/…` reads the CURRENT pod only — the pre-roll pod's
log is gone once it is replaced, so capture during/immediately after the roll.
Always record pod age next to the count: a 0 from a young pod over an idle
window (R1 empty) is not evidence of absence. This pairing is exactly how a 0
was nearly misread on 2026-07-20.

## R3 — orchestration volume by day (sizing baseline)

```sql
SELECT date_trunc('day', created_at) AS day, count(*)
FROM orchestration_states
WHERE created_at > now() - interval '7 days'
GROUP BY 1 ORDER BY 1;
```

2026-07-20 baseline: ~1.9k–3.9k/day at 11 deployed sites.

## R4 — dispatch queue depth and drain rate

Use the **dispatch_queue_serialisation** workstream's RUNBOOK (R7 there), not
ad-hoc sampling: the drain is a sawtooth, and two point readings produced two
wrong rates in opposite directions (WRONG_CALLS 7 & 8). Their runbook holds
the current method; do not copy a stale variant into here.

## R5 — status of this workstream's two filed diagnosis claims

```sql
-- the work items
SELECT left(summary, 60), status, created_at
FROM site_work_items
WHERE item_type = 'needs_diagnosis'
  AND (summary LIKE '%ProcessResponse%' OR summary LIKE '%processRequests%');

-- the runs, by correlation (uuid column: cast before LIKE)
SELECT correlation_id, status, current_step
FROM orchestration_states
WHERE correlation_id::text LIKE '2d02d62a%'
   OR correlation_id::text LIKE '78470372%';

-- verdict artifacts, once a run completes
SELECT iteration, kind, left(body, 200)
FROM diagnosis_artifacts
WHERE correlation_id = '2d02d62a-7d96-41f0-a82b-e1ebd7ef5d6b'
ORDER BY iteration;
```

**Gotchas:** `correlation_id` is uuid-typed — bare `LIKE` fails
(`operator does not exist: uuid ~~ unknown`); cast `::text`. A missing
orchestration row for ~30+ min after filing is QUEUE LATENCY, not a drop
(`bugs_open/030`) — check `LAG` on `generic-requests-group` before concluding
anything, and never re-file on absence alone.

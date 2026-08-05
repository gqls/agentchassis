# RUNBOOK — bugfix 196

Every command here was needed at least once and has its gotcha attached.

## Production probe (the shapes that are actually recorded)

Gotcha: a whole-document `collected_data::text LIKE` counts rows containing the
bytes ANYWHERE (workflow definitions mention `"status": "failed"`), not steps
recording a failure. Probe the two shapes `applyResponseToState` writes:

```sql
-- direct-store shape (non-agent steps)
SELECT count(*) FROM orchestration_states, jsonb_each(collected_data) kv
WHERE jsonb_typeof(kv.value)='object'
  AND kv.value->>'status'='failed' AND kv.value ? 'error';
-- .response-wrapped shape (call_agent / spawn_agent steps)
SELECT count(*) FROM orchestration_states, jsonb_each(collected_data) kv
WHERE jsonb_typeof(kv.value)='object'
  AND kv.value->'response'->>'status'='failed' AND kv.value->'response' ? 'error';
```

Bound the claim: terminal rows are reaped ~24h, and the `output_mapping` path
(coordinator.go:2647) ERASES the blob (extracts mapped fields only) — a zero is a
lower bound, never "never happens". Sample matches before quoting any non-zero.

## Conditions census (objection 2's full form — no filter)

```sql
SELECT DISTINCT jsonb_path_query(default_config, 'strict $.**.condition') #>> '{}'
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
ORDER BY 1;
```

Gotcha: `count(DISTINCT jsonb_path_query(...))` fails — "aggregate function calls
cannot contain set-returning function calls"; wrap in a subquery or just read the
list. Escape `$` as `\$` inside a double-quoted shell heredoc.

## Council round

Submission JSON validated against the strict types BEFORE firing (risks is a
STRING; operation ∈ modify|add|remove|config_change; ≤8 edits):
corr `d1a63089-af5b-41a2-bea1-62259aa5db52` → APPROVED round 1 (11:57Z 08-05).

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='d1a63089-af5b-41a2-bea1-62259aa5db52' AND kind='council_report';
-- full report with per-seat objections: same table, column is body (TEXT), not content
```

## Deploy verification (objection 3 — at the pod, never git/tag/status)

```bash
# positive control — a string the change ADDS (expect >=1 on EVERY replica):
kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | while read p; do
  echo "$p: $(kubectl -n ai-persona-system exec ${p#pod/} -- sh -c \
    'strings /app/agent-chassis | grep -c sendWorkflowResponseWithStatus')"
done
# negative control — the OLD sendErrorResponse produced no such symbol, so instead
# use a string the change REMOVES. The fix removes nothing greppable (the old
# constant "complete" stays for the success path), so the negative control is the
# TAG + a behavioural induction (below), per the 153 lesson that a roll is not
# evidence. Do not skip the induction.
```

## The induction (the bug file's own settlement measurement)

Baseline (pre-roll, optional but strongest): dispatch a parent with one
`call_agent` step at an agent that fails non-permanently; parent completes with
the blob as step data. Post-roll: same dispatch; parent must route to error
handling (`error` column set by failWorkflow with the child's message, or
error_step taken). Read the PARENT:

```sql
SELECT status, current_step, error, collected_data->'<step>' AS step_data
FROM orchestration_states WHERE orchestration_id='<parent>';
```

Gotcha: no orchestration dispatch within ~300s of a chassis pod (re)start — the
spawn is silently dropped. And terminal rows reap in ~24h: read the row the same
day.

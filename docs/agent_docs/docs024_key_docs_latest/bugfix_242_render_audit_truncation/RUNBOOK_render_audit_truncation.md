# RUNBOOK — bugfix 242

## Enumerate the artefact shape on live runs (step key AND output field — the bug only enumerated one)

```sql
SELECT orchestration_id, created_at,
       (SELECT array_agg(k) FROM jsonb_object_keys(collected_data->'audit') k) AS audit_keys,
       (SELECT array_agg(k) FROM jsonb_object_keys(collected_data->'render_audit') k) AS render_audit_keys
FROM orchestration_states
WHERE owner_agent_type='render-audit-agent'
ORDER BY created_at DESC LIMIT 8;
```
Gotcha: `jsonb_object_keys` in the select list needs the scalar-subquery wrapper or it
multiplies rows. Retention on COMPLETED rows is ~24h, AWAITING_RESPONSES ~4h (bug 236's
contribution block) — query the same day or you are measuring retention, not behaviour.

## Read the rotation's workflow config (live row, never the seed)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'audit')
FROM agent_definitions
WHERE type='render-audit-agent' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## After the fix is live: prove it on a run whose cap bit

```sql
SELECT collected_data->'render_audit'->'response'->'summary' AS summary
FROM orchestration_states
WHERE owner_agent_type='render-audit-agent'
ORDER BY created_at DESC LIMIT 3;
-- expect pages_total alongside pages, truncated:true when they differ

SELECT occurred_at, error_message, context
FROM agent_error_log
WHERE error_code='RENDER_AUDIT_TRUNCATED'
ORDER BY occurred_at DESC LIMIT 5;
```
Grade only on a site whose page count exceeds the configured cap; a small site cannot
distinguish fix from no-fix (bug §7).

## Deploy provenance (which commit is the pod running)

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
# startup line — scrolls out of range on busy services; empty means "not in range", not "unstamped"
git merge-base --is-ancestor <your-commit> <the stamp>   # "did my fix ship?" as a query
```
The adapter half needs the BROWSER-RUNNER image (render-audit-adapter is a second
Deployment of the same image with REQUESTS_TOPIC overridden) — check that service's own
stamp, per-service not per-fleet (`bugs_open/249`).

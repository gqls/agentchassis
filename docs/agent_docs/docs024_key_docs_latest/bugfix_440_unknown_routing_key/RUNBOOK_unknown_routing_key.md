# RUNBOOK — bugfix 440

## The census (re-run before quoting; every figure dates fast)
```sql
SELECT COALESCE(spec->>'reason','(none)') AS reason, count(*),
       min(created_at)::date, max(created_at)::date
FROM site_work_items WHERE item_type='page_rerender'
GROUP BY 1 ORDER BY 2 DESC;
```

## Counting warning EMISSIONS (not string presence) — the trap that bit on day one
A text-LIKE over `collected_data` matches council payloads QUOTING the string. Exclude the
quoting population and read one member:
```sql
SELECT orchestration_id, current_step FROM orchestration_states
WHERE updated_at >= '<since>'
  AND collected_data::text LIKE '%not in the sections-rerender vocabulary%'
  AND collected_data->'input_data'->>'fix_correlation_id' IS NULL;
```

## Is the warning capability in the running binary
```bash
POD=$(kubectl -n ai-persona-system get pod -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- grep -aq "not in the sections-rerender vocabulary" /proc/1/exe && echo LIVE
```

## The live gate's condition (schema first; the declaration auditor holds it daily)
```sql
SELECT default_config->'workflow'->'steps'->'check_rerender_mode'
FROM agent_definitions WHERE type='page-rerender' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

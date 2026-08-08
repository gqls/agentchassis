# RUNBOOK — bugfix 210

## Count refusals (the permanent signal this fix adds; 0 until the roll)
```sql
SELECT count(*), max(occurred_at) FROM agent_error_log
WHERE error_code = 'DEPLOY_STAMP_REFUSED_ON_SKIP';
-- column is occurred_at, NOT created_at (cost this lane two error round-trips)
```

## See parked pages
```sql
SELECT s.domain, w.item_key, w.status, w.created_at, w.spec->>'skip_reason'
FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.item_type='page_build_failed' ORDER BY w.created_at DESC;
```

## Who holds a page's slot (all three producers share the key namespace)
```sql
SELECT item_type, status, created_at::date, summary FROM site_work_items
WHERE site_id='<site>' AND item_key='needs_page:<name>' ORDER BY created_at DESC;
```

## Live-window check for the bug shape (WEAK — last loop iteration only)
```sql
SELECT orchestration_id, current_step, status,
       left(collected_data->'assembled_page'->>'skip_reason', 90)
FROM orchestration_states
WHERE collected_data->'assembled_page'->>'skipped' = 'true'
ORDER BY updated_at DESC LIMIT 12;
-- PK is orchestration_id; there is no `id` column. Gotcha: a mid-loop skip
-- followed by a later page's success leaves NO trace here.
```

## Prove the fix on a pod after the roll (positive + negative, every replica)
```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | while read p; do
  echo "$p:"
  kubectl -n ai-persona-system exec "$p" -- sh -c \
    'strings /app/agent-chassis | grep -c "DEPLOY_STAMP_REFUSED_ON_SKIP"; strings /app/agent-chassis | grep -c "refusing to stamp deployed — page was skipped by the owned-page guard"'
done
# first number ≥1 (added string), second stays ≥1 (unchanged 208 guard = positive control
# for the grep pipeline, not for this change — the ADDED string is the real evidence)
```

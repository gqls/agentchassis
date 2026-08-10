# RUNBOOK — bugfix 234

Commands that were hard to get right, with their gotchas. Update HERE when one changes.

## The carrier census — ALL DEPTHS, or the number is wrong

Top-level `->'workflow'->'steps'` walks miss loop sub-workflows (the 08-09 landmine; 356's
`commit_from` had 3 of 6 carriers nested). Use the recursive walk:

```sql
WITH all_steps AS (
  SELECT ad.type, jsonb_path_query(ad.default_config,
         'strict $.**?(@."action" == "create_work_item")') AS step
  FROM agent_definitions ad
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL)
SELECT type, step->'config'->'spec', step->>'description'
FROM all_steps WHERE step->'config' ? 'spec';
-- expected BEFORE migration: 3 rows; AFTER: 0 rows
```

Gotcha (cost me a minute on 2026-08-09): grouping that census by `(type, key)` and then
counting ROWS undercounts — improvement-loop carries `spec` twice, one row in the group. Read
the count column, not the row count.

Escaping: inside a `bash -c`/heredoc to psql, `\$` the jsonpath's `$` or the shell eats it.

## Damage / proof-at-a-filed-row

```sql
SELECT item_key, spec, created_at, status FROM site_work_items
WHERE created_by='improvement-loop' AND item_key LIKE 'improvement_rerender%'
ORDER BY created_at DESC LIMIT 3;
```
- BEFORE the migration: every spec `{}` (16/16 as of 2026-08-09, first row 2026-08-01).
- The fix is proven only by a row **filed after** the migration carrying
  `{"refresh_site_components": true}`. A definition that LOOKS right is exactly this bug.
- Natural rate ~1.8 rows/day. The currently-triaged `improvement_rerender_dartsonline.com`
  row (2026-08-09 13:19Z) predates the fix and stays empty — do not hand-edit it.

## Pod-grep (code half, post-roll)

```bash
for p in $(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name); do
  kubectl exec -n ai-persona-system ${p#pod/} -- sh -c \
    'strings /app/agent-chassis | grep -c "bugs_open/234"; strings /app/agent-chassis | grep -c "zzz_no_such_string_234"'
done
# expect: N>0 then 0 (the second is the pipeline control), EVERY replica
```
The `StrictConfig` bool itself cannot be strings-proven. Live proof = canary: seed a
throwaway active agent whose `create_work_item` step carries a bogus key, watch
ValidateWorkflow reject it in the pod log, delete the canary. (Unit test pins it in CI
either way.)

## Migration apply

Per migration-runner practice: dry-run the runner first, this session, and scope the dir on
`--apply` — it takes EVERY pending file otherwise. This lane applied by hand + `--record-only`.
Re-check the next free number immediately before writing the file: three numbers were claimed
by other sessions while 356 was being written.

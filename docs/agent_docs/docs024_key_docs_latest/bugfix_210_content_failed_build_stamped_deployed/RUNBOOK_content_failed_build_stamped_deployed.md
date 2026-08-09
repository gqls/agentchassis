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

## Cancel-as-mute audit (guardian objection 6 — run before dartsonline is next replanned)
```sql
-- items whose mute the loadOpenPageItems 'cancelled' alignment releases;
-- re-mark intentional mutes as wont_fix (closed under BOTH old and new rules)
SELECT s.domain, w.id, replace(w.item_key,'needs_page:','') AS page, p.build_status
FROM site_work_items w
JOIN sites s ON s.id=w.site_id
LEFT JOIN pages p ON p.site_id=w.site_id AND p.name=replace(w.item_key,'needs_page:','')
WHERE w.item_type='needs_page' AND w.status='cancelled'
  AND COALESCE(p.build_status,'') NOT IN ('deployed')
  AND NOT EXISTS (SELECT 1 FROM site_work_items w2 WHERE w2.site_id=w.site_id
                  AND w2.item_key=w.item_key AND w2.status NOT IN
                  ('complete','verified','rejected','wont_fix','failed','cancelled'));
```

## Which workflows can reach the guard? (the door census — 2026-08-09)

Guard cover follows `output_field = 'assembled_page'`, **not** the presence of an
`update_page_status` step: `upstreamAssemblySkipped` reads that one key. Two queries.

**Gotcha:** `jsonb_each(jsonb_path_query(...))` fails with *"set-returning functions must appear
at top level of FROM"*. Chain them as separate `LATERAL`s (or wrap the inner one in a CTE).

```sql
-- 1. every route INTO assemble_page, fleet-wide. Expect 3, all check_review_approved.
SELECT ad.type, e.path, e.step->>'action',
       COALESCE(e.step->'config'->>'then_step', e.step->>'next_step') AS goes_to,
       e.step->'config'->>'else_step' AS diverts_to
FROM agent_definitions ad,
  LATERAL jsonb_path_query(ad.default_config,'$.**.steps') AS steps,
  LATERAL jsonb_each(steps) AS e(path, step)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND (e.step->>'next_step'='assemble_page' OR e.step->'config'->>'then_step'='assemble_page'
       OR e.step->'config'->>'else_step'='assemble_page');

-- 2. and confirm nothing STARTS there (a loop entry point is not an edge)
SELECT type, jsonb_path_query_first(default_config,'$.**.sub_workflow')->>'start_step' AS loop_start,
       default_config->'workflow'->>'start_step' AS wf_start
FROM agent_definitions WHERE type IN ('page-rebuild','pageflow-builder','site-work-orchestrator')
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Do NOT try to measure historical skips from `orchestration_states`

It reads clean and answers nothing. A `loop` sub-workflow's per-iteration state is **not**
retained on the parent row — `page-rebuild`'s `collected_data` holds pre-loop steps only
(`get_pages_to_rebuild`, `load_rebuild_context`, `select_style_collection`, `site_record`,
`pages_to_build`). So `collected_data->'assembled_page'` exists on **1 row of 4,457** and any
filter over it returns 0 regardless of production. Check the key's existence first — it is the
denominator that invalidates the count:

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE collected_data ? 'assembled_page')                AS key_present,
       count(*) FILTER (WHERE collected_data->'assembled_page'->>'skipped'='true') AS looks_like_an_answer
FROM orchestration_states;   -- key_present ≈ 0 ⇒ the third column means nothing
```

The `agent_error_log` counter (§ watch) is the only instrument, forward-only from the roll.

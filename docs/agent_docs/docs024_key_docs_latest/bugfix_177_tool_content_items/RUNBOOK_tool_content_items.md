# RUNBOOK — bugfix 177

DB shell:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## State of the class (re-run before quoting any number)

```sql
-- the class, by status (healthy end-state after the fix + sweep:
-- 0 needs_human_review, N wont_fix)
SELECT status, count(*) FROM site_work_items
WHERE item_key LIKE 'tool_content:%' GROUP BY 1;

-- the control class from the SAME emitter files (guide pages declare
-- sections, so these complete — if THESE start dying, the defect is not 177's)
SELECT status, count(*) FROM site_work_items
WHERE item_key LIKE 'tool_guide:%' GROUP BY 1;

-- anything still blocked on a tool_content row (must be 0 after the sweep)
SELECT left(w.id::text,8), w.status FROM site_work_items w
WHERE w.status NOT IN ('complete','verified','cancelled','rejected','wont_fix')
  AND w.depends_on && ARRAY(
    SELECT id FROM site_work_items WHERE item_key LIKE 'tool_content:%');
```

## The guard's decision inputs, for one page

```sql
-- what the handler (and the guard) will resolve, in priority order
SELECT sps.component_name FROM site_plan_sections sps
JOIN site_plans sp ON sp.id=sps.plan_id
WHERE sp.site_id='<site>' AND sp.is_current AND sps.page_name='<page>'
ORDER BY sps.ordering;                                   -- source 1

SELECT jsonb_path_query(data, '$.pages[*] ? (@.name == "<page>").sections')
FROM site_specs WHERE site_id='<site>' AND aspect='site_plan' AND is_current;  -- source 2

SELECT sections FROM pages WHERE site_id='<site>' AND name='<page>';  -- source 3
```
Gotcha: the guard takes the FIRST NON-EMPTY source, exactly like
`load_page_sections_from_spec_action.go` — do not union the three when
reasoning about what it will do. Sibling synthesis (source 4) is deliberately
NOT mirrored: it requires plan membership.

## Sweep (one-time, idempotent — re-runnable, the WHERE clauses self-empty)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/297_sweep_tool_content_zombies_and_release_dependents.sql
```
Gotcha: the file pipes on stdin because psql in the pod cannot see the repo
path; `-f -` keeps line numbers in error output where a heredoc loses them.

## Council

Submitted 2026-08-03 ~11:55 BST.
`SUBMISSION_CORR=982507b0-2e18-4457-a354-85a809012bbd`

```sql
-- find the run (payload, not printed id; ~30 min queue latency is NORMAL)
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id'
      = '982507b0-2e18-4457-a354-85a809012bbd';
-- read the verdict
SELECT body FROM doc_notes WHERE categories ? 'council-gate'
ORDER BY created_at DESC LIMIT 1;
```

## Verify the fix (post image roll)

Pod-proof first (a roll is not evidence the fix shipped):
```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | while read p; do
  kubectl -n ai-persona-system exec $p -- sh -c \
    'strings /app/agent-chassis | grep -c "raiseToolContentItem"; strings /app/agent-chassis | grep -c "Failed to create tool content work item"'
done
# expect: positive count for the ADDED symbol, 0 for the REMOVED string, every replica
```
(The removed string is the old call sites' Warn message — it must be absent;
the helper's messages are new. If the helper kept that exact wording, pick
another removed string from the diff before trusting this.)

Behavioural (the bug file's own verify): generate a tool on a site with no
plan entry for it → **no** `tool_content:%` row appears; the `tool_guide:%`
row DOES appear (positive control). Then the deploy path: fork a library tool
→ its `tool_content:%` row appears (the guard resolves its 3 declared prose
sections).
```sql
SELECT left(item_key,60), status, created_at FROM site_work_items
WHERE item_key LIKE 'tool_content:%' OR item_key LIKE 'tool_guide:%'
ORDER BY created_at DESC LIMIT 6;
```

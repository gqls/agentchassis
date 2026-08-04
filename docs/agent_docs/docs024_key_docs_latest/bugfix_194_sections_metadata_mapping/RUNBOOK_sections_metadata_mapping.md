# RUNBOOK — `bugs_open/194` lane

Every command that was hard to get right, with its gotcha attached. Fix it HERE, not in
scrollback.

## R1 — census every `save_page_sections` caller

**Gotcha:** the step is nested in a loop `sub_workflow` in four of the six callers, so the
usual top-level `jsonb_each(default_config->'workflow'->'steps')` finds only two of them and
reads as "the fleet is fine". Use `jsonb_path_query(..., '$.**.steps')`.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT ad.type, s.key, s.value->'config'->>'sections_metadata_field', s.value->'config'->>'html_field'
FROM agent_definitions ad,
LATERAL jsonb_path_query(ad.default_config, '\$.**.steps') AS steps,
LATERAL jsonb_each(steps) AS s(key,value)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' = 'save_page_sections'
ORDER BY ad.type;"
```

**Second gotcha (LANDMINE, 016b):** `default_config::text LIKE '%save_page_sections%'` is NOT
a test for this step — `council-gate` and `fix-proposer` both "contain" the string in prompt
text and neither has the step. Match on `value->>'action'`.

## R2 — a caller's step graph (what data is actually in scope at the save)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT s.key, s.value->>'action', COALESCE(s.value->>'output_field','')
FROM agent_definitions ad,
LATERAL jsonb_path_query(ad.default_config, '\$.**.steps') AS steps,
LATERAL jsonb_each(steps) AS s(key,value)
WHERE ad.type='<agent>' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;"
```

Read the `output_field` of the writer step: that prefix plus `.response.<key>` is what a
config path must name. `page_content` → `page_content.response.sections_metadata`.

## R3 — is an agent dormant? Do NOT ask `orchestration_states`

**Gotcha:** terminal rows are reaped ~daily. `min(created_at)` over the whole table says
weeks only because unreaped non-terminal statuses set the floor; bound per status and you see
`COMPLETED` goes back one day. Use `agent_run_stats` (no reaper, spans from 2026-07-26):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT agent_type, run_count, first_ran_at::date, last_ran_at FROM agent_run_stats
WHERE agent_type IN ('pageflow-builder','site-work-orchestrator','tool-recreation-handler',
                     'page-build-handler','page-rebuild','page-rerender')
ORDER BY last_ran_at DESC NULLS LAST;"
```

**Sanity check before trusting an absence:** confirm the table tracks agents of the same
SHAPE as the one you are calling dormant (orchestrators, not just leaves) —
`SELECT agent_type, run_count FROM agent_run_stats ORDER BY run_count DESC LIMIT 25;`

## R4 — the state of a page's `content_data` (the bug's own verification query)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT slot_name, length(rendered_html), length(content_data::text), updated_at
FROM page_components WHERE page_id='<uuid>' ORDER BY position;"
```

**Gotcha:** `length(content_data::text)` = 0 is impossible; NULL prints as empty in `-tA`, so
an empty column IS the NULL. Add `content_data IS NULL` explicitly if the output is going
into a claim.

## R5 — offline test of the action

```bash
timeout 400 go test ./platform/orchestration/actions/ -run 'TestSavePageSections' -v 2>&1 | tail -30
gofmt -l platform/orchestration/actions/
```

## R6 — prove it shipped (never trust the tag or a green roll)

```bash
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name | head -1)
kubectl exec -n ai-persona-system ${POD#pod/} -- sh -c 'strings /app/agent-chassis | grep -c "<a string my change ADDED>"'   # expect >0
kubectl exec -n ai-persona-system ${POD#pod/} -- sh -c 'strings /app/agent-chassis | grep -c "<a string my change REMOVED>"' # expect 0
```

**Gotcha:** the positive control alone proves the *pipeline*, never your spelling — run the
negative one in the same exec, on every replica. And `logs deploy/X` reads one pod of N.

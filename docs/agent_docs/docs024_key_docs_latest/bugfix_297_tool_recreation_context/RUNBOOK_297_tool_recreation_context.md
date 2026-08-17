# RUNBOOK — bugs_open/297

Every command that was hard to get right, with its gotcha attached.

## Read the live step (the whole config trap: prompt and cap sit at different depths)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A <<'EOF'
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'load_related_context')
FROM agent_definitions
WHERE type='tool-recreation-handler' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
EOF
```

Gotcha: `orchestration_states.owner_agent_type` returns **0** for this agent and is the wrong
instrument for "does it run" — use `llm_call_log` (290 calls) and `site_work_items.handler_agent`.

## The census (population per site, the query's own predicate)

```sql
WITH pop AS (SELECT site_id, COUNT(*)-1 AS population FROM pages GROUP BY site_id)
SELECT COUNT(*), COUNT(*) FILTER (WHERE population > 10),
       percentile_cont(0.5) WITHIN GROUP (ORDER BY population), MAX(population)
FROM pop;
```

## Rendered-payload arithmetic (matches the template `- name (type): title\n`)

```sql
SELECT SUM(4 + length(p.name) + 4 + COALESCE(length(p.page_type),0)
           + COALESCE(length(p.title),0) + 1)
FROM pages p WHERE p.site_id = '<site>';
```

## Fan-out check (why the LATERAL exists)

```sql
SELECT page_id, COUNT(*) FROM research_results
WHERE result_type='adoption_page' GROUP BY page_id HAVING COUNT(*) > 1;
-- 2026-08-17: one hit, page 0747e2fc… = 'index' of site 00ff3af5…, nav_order 1
```

## Apply (own file only, then record — never a hand INSERT)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/453_tool_recreation_whole_site_context.sql
./scripts/migration/run-migrations.sh --record-only \
  docs/agent_docs/sql_for_agents/453_tool_recreation_whole_site_context.sql \
  --note "bugs_open/297: applied by hand 2026-08-17, verified post-state (107-row worst site, fan-out de-duplicated)"
```

Gotcha: `--apply` takes EVERY pending file in the dir (other threads'). Snapshot lands in
`agent_definitions_backup` (NOT an `is_snapshot` row) — check `snapshot_taken_at`, not `created_at`.

## Post-apply verify

```sql
-- the query text changed and carries no multi-row LIMIT
SELECT default_config#>>'{workflow,steps,load_related_context,config,query}'
FROM agent_definitions WHERE type='tool-recreation-handler' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- run it for the worst site: expect 107 rows (was 10)
```

Disconfirming pair: any page with `nav_order` position > 10 on a >10-page site — structurally
impossible in the old result, present in the new one. End-to-end (`llm_call_log.prompt_rendered`)
confirms on the next real recreation run.

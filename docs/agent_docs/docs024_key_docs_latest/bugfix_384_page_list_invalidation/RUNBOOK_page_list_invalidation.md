# RUNBOOK — bugs_open/384 page-list consumer invalidation

Every command here was got wrong at least once before it was got right; the gotcha is attached.

## DB access
```bash
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -X -q"
$PSQL <<'SQL'
\pset format aligned
...
SQL
```
- `-X -q` keeps psqlrc noise out. Heredoc + `<<<` in the same command clashes — write HTML to a scratch file first.
- The shell cwd DRIFTS after a `cd` in an earlier call — use absolute paths or `cd /home/ant/projects/agentchassis;` first.

## Schema facts that cost a query each
- `pages` has **no `slug`** — key on `pages.url` (per site: `/index.html` exists once per site, so ALWAYS join `sites`) or `pages.name`.
- `content_components.input_schema->'fields'` is an **OBJECT** (name → spec), not an array: `jsonb_each(...)`, never `jsonb_array_elements`.
- `agent_definitions` has no `name` column; `orchestration_states` keys the agent by `owner_agent_type`.

## Who consumes a query source (the consumer set — the seam's own query)
```sql
SELECT f.value->>'source' AS source, cc.name, f.key AS field,
       (SELECT count(*) FROM page_components pc WHERE pc.component_id=cc.id AND pc.build_status<>'removed') AS live_instances
FROM content_components cc, jsonb_each(coalesce(cc.input_schema->'fields','{}'::jsonb)) f
WHERE f.value->>'source' LIKE 'query.%' ORDER BY 1,2;
```
`[MEASURED 2026-08-24]` 43 fields / 25 components. Page-IMAGE sources (splice `pageImageJoins`): `pages_where_type:*`, `blog_posts`, `pages_under_section:*`. `section_index_for` does NOT.

## Pair census: (card asset, stored entry) — the disconfirmable measurement
```sql
WITH qf AS (SELECT cc.id component_id, cc.name component, f.key field, f.value->>'source' source
            FROM content_components cc, jsonb_each(coalesce(cc.input_schema->'fields','{}'::jsonb)) f
            WHERE f.value->>'source' LIKE 'query.%' AND f.value->>'type'='array'),
ent AS (SELECT p.site_id, p.url listing_url, qf.component, qf.source, pc.updated_at array_written_at,
               e.value->>'url' entry_url, coalesce(e.value->>'image','') entry_image
        FROM page_components pc JOIN qf ON qf.component_id=pc.component_id JOIN pages p ON p.id=pc.page_id
        CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(pc.content_data->qf.field)='array' THEN pc.content_data->qf.field ELSE '[]'::jsonb END) e
        WHERE pc.build_status<>'removed' AND p.status='active')
SELECT ent.source, ent.component, count(*) pairs_with_card,
       count(*) FILTER (WHERE position(ca.url in ent.entry_image)>0) current,
       count(*) FILTER (WHERE position(ca.url in ent.entry_image)=0) stale,
       count(*) FILTER (WHERE position(ca.url in ent.entry_image)=0 AND ent.array_written_at > ca.created_at) stale_though_written_after_card
FROM ent JOIN pages tp ON tp.url=ent.entry_url AND tp.site_id=ent.site_id
JOIN assets ca ON ca.site_id=tp.site_id AND ca.entity_type='page' AND ca.entity_id=tp.id AND ca.purpose='card' AND ca.status='active'
GROUP BY 1,2 ORDER BY 5 DESC;
```
- Gotcha: a first cut keyed `empty image` over ALL sources and showed news/directory arrays as "20/20 empty" — those entries have no `image` key at all. Join the card, don't count empties.
- Gotcha: verify the URL join on one site before believing a zero (leopardess `/blog.html` = genuinely no cards, checked entry-by-entry).

## Which re-render mode a producer gets
`page-rerender`'s `check_rerender_mode` condition (live, read from `agent_definitions`): reason ∈ {image_landed, section_data_resolved, cta_links_stale, template_changed, literal_markdown} → `rerender_page_sections` (re-resolves `query.*`); anything else → assemble (re-ships stored arrays). Dump it:
```bash
$PSQL -At -c "SELECT default_config::text FROM agent_definitions WHERE type='page-rerender' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL" > /path/agentdef.json
```

## Existing producers of `section_data_resolved` (precedent shapes)
```sql
SELECT DISTINCT ON (created_by) created_by, item_type, handler_agent, item_key, spec::text
FROM site_work_items WHERE spec->>'reason'='section_data_resolved' ORDER BY created_by, created_at DESC;
```
`render_news_section` → `page_rerender`/`page-rerender`, key `page_rerender_<name>_<site>_section_data_resolved` (THE shape to copy). `render_directory`/`reconcile_section_data` → `needs_page`/`page-build-handler` (LLM chain; not ours).

## Risk meters for the handler that receives the new items
```sql
-- escalation rate of reasoned re-renders (baseline 1/25 on 2026-08-24)
SELECT collected_data->'input_data'->'spec'->>'reason' reason, status, coalesce(collected_data->'rerender_sections'->>'escalated','(n/a)') escalated, count(*)
FROM orchestration_states WHERE owner_agent_type='page-rerender' AND created_at > now()-interval '14 days' GROUP BY 1,2,3;
-- owned pages FAIL save_sections on this path → excluded at the lookup
SELECT left(coalesce(error,'(null)'),90), count(*) FROM orchestration_states WHERE owner_agent_type='page-rerender' AND status='FAILED' AND created_at > now()-interval '14 days' GROUP BY 1;
```

## Ownership checks (re-run at EVERY phase boundary — before first code, before each commit, before a council round)
```bash
git log --oneline --since='90 minutes ago' -- bugs_open/384* platform/orchestration/actions/derive_card_asset_action.go platform/orchestration/actions/flag_page_image_rebuild_action.go platform/orchestration/actions/queryresolve
CUT=$(date -u -d '30 minutes ago' +%Y-%m-%dT%H:%M:%SZ); find ~/.claude/projects/-home-ant-projects-agentchassis/ -maxdepth 1 -name '*.jsonl' -newermt "$CUT" | grep -v <own-session-id> | xargs grep -c "PageListConsumerPages\|requestPageListReresolve\|derive_card_asset_action" | grep -v ':0$'
```
Peer sessions are addressable by name via ListAgents/SendMessage (`bugs_open/326`, `bugs_open/357`, `bugs_open/352`, `bugs_open/333 [cb419e]`). The filing lane (`dartsonline_traffic`) has no named session — coordinate through `bugs_open/384` itself.

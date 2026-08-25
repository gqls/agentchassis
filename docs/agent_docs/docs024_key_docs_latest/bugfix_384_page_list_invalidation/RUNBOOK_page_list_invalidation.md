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

## Phase 2 — enabling the sweep (ONLY after the roll)
1. Prove the binary registers the check (capability list, not `strings`):
   ```sql
   SELECT service, built_from, capabilities ? 'page_list_stale' FROM service_binary_capabilities ORDER BY recorded_at DESC LIMIT 3;  -- schema: \d first
   ```
   and `git merge-base --is-ancestor <phase-2 commit> <build provenance sha>` per CLAUDE.md.
2. Apply by hand: `docs/agent_docs/sql_for_agents/603_enable_page_list_stale_HOLD.sql` (snapshot_agent first, DO/RAISE verify inside). Rollback file beside it.
3. First sweep proof (demand control = the 4 sites with stale tool-cta entries, 14 pairs on 2026-08-24 — re-run the pair census first, it may have moved):
   ```sql
   SELECT s.domain, w.status, w.spec->'stale' FROM site_work_items w JOIN sites s ON s.id=w.site_id
    WHERE w.item_type='page_rerender' AND w.spec->>'check'='page_list_stale' ORDER BY w.created_at DESC;
   ```
   Disconfirming result: the completeness agent visits a stale site (`site_discovery_rotation`) and files nothing; or files against a page whose stored array matches a fresh resolve (re-run the comparison by hand before calling it wrong).
4. The per-run summary is in the discovery run's findings (`"summary":true` with stale/current/unknown) — `unknown > 0` on a site means a source did not resolve; that is not "current".

## Post-roll acceptance (the protocol that PASSED 2026-08-25)
```bash
./docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/scripts/induce_card_landing.sh dartsonline.com barrel-shapes
```
⚠ **That script's kcat route FAILS** — `asset-deployer` dispatched to `system.agent.generic.requests` lands on a pod with no S3 client (`derive_card_asset: storage client not available`). Use the **work-item route** instead, which is production's own path:
```sql
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority,
                             handler_agent, status, created_by, item_key, batch_id)
VALUES ('<site>','image-build-handler','build','needs_content_image','low','<why>',
        '{"mode":"content_card","check":"<your_tag>","entity_type":"page","entity_id":"<page uuid>","page_name":"<page>","purpose":"card"}'::jsonb,
        65,'asset-deployer','triaged','<your_tag>','content_image:<page>', gen_random_uuid())
ON CONFLICT DO NOTHING;
```
Then wait for `build-dispatch-loop` — it runs **per site** (`load_items` takes `input_data.site_id`), roughly every 90 minutes per site, so check the rotation rather than assuming it stalled:
```sql
SELECT s.domain, count(*), max(o.created_at)::timestamp(0) FROM orchestration_states o
  LEFT JOIN sites s ON s.id::text = o.collected_data->'input_data'->>'site_id'
 WHERE o.owner_agent_type='build-dispatch-loop' AND o.created_at > now()-interval '3 hours' GROUP BY 1 ORDER BY 3 DESC;
```
The assertion (expect exactly N, none on an `owned` page):
```sql
SELECT p.name, COALESCE(p.rebuild_policy,'generic') AS policy, w.status, w.spec->>'reason', w.created_by, w.item_key
  FROM site_work_items w JOIN pages p ON p.id=w.page_id
 WHERE w.site_id='<site>' AND w.item_type='page_rerender' AND w.spec->>'cause'='card_landed:<page>' ORDER BY p.name;
```
And the causation leg. ⚠ **Do NOT require `pages.deployed_at` to advance** — corrected 2026-08-25 by running it: on a listing whose array is already current the re-resolve produces byte-identical HTML, the deploy is a no-op, and `deployed_at` legitimately does not move (measured: `index` re-rendered 4 of 4 sections, 0 carried, deploy step visited, `deployed_at` unchanged). The causation signals that DO discriminate are (a) the item row itself carries `spec.cause`, (b) the run carries it too, and (c) `page_components.updated_at` advances when the array is rewritten:
```sql
SELECT coalesce(collected_data->'input_data'->'spec'->>'page_name','?'), status,
       coalesce(collected_data->'rerender_sections'->>'escalated','n/a'), current_step
  FROM orchestration_states WHERE owner_agent_type='page-rerender'
   AND collected_data->'input_data'->'spec'->>'cause'='card_landed:<page>';
```
⚠ **The served page is not the measurement on dartsonline** — it serves 12/12 because the filer hand-repaired it on 2026-08-24. The ROWS are.

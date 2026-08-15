# RUNBOOK — bugs_open/281 tool audit by instance

All SQL via `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.
`<webdesign>` = `(SELECT id FROM sites WHERE domain='webdesign.co.uk')`.

## Census — who the widened check sees (before: 4; after: 66 on webdesign)

```sql
-- the widened population, exactly as check_tool_health now selects it
SELECT cc.component_level,
       CASE WHEN cc.component_level='tool' THEN cc.function ELSE regexp_replace(p.name,'^tool-','') END AS subject_key,
       p.name
FROM content_components cc JOIN page_components pc ON pc.component_id=cc.id JOIN pages p ON pc.page_id=p.id
WHERE p.site_id=<webdesign> AND cc.is_active AND p.status='active'
  AND (cc.component_level='tool'
       OR (p.page_type='tool'
           AND NOT EXISTS (SELECT 1 FROM page_components pc_t JOIN content_components cc_t ON cc_t.id=pc_t.component_id
                           WHERE pc_t.page_id=p.id AND cc_t.component_level='tool' AND cc_t.is_active)
           AND (SELECT count(*) FROM page_components pc_n JOIN content_components cc_n ON cc_n.id=pc_n.component_id
                WHERE pc_n.page_id=p.id AND cc_n.is_active)=1))
ORDER BY p.name;
```
Gotcha: expect **66**, not 67 — `tool-ab-test-calculator`'s page carries BOTH a fork and a
ported instance; clause (a) counts the fork, clause (b) rightly skips the page (2 components).

## First-sweep item census (after the roll + a design-discovery run on webdesign)

```sql
SELECT item_type, status, count(*), count(DISTINCT item_key) AS distinct_keys
FROM site_work_items
WHERE site_id=<webdesign> AND created_at > '<sweep-start>'
  AND item_type IN ('improve_tool','audit_tool','ported_tool_fix')
GROUP BY 1,2 ORDER BY 1,2;
-- expect: ported_tool_fix rows == distinct_keys; improve_tool only for the 4 forks;
-- audit_tool <= 12 per run (cap), fork keys byte-identical to pre-change rows.
```
Gotcha (LANDMINES): the check NAME `tool_health` is not an item_type — query by the
`ItemType:` literals in the source (`improve_tool`, `ported_tool_fix`, `audit_tool`).

## Mind Map detectability (the motivating case)

```sql
-- >3 bare hex on colour properties in the instance's own markup = hardcoded_colors fires
SELECT count(*) FROM (
  SELECT regexp_matches(pc.rendered_html,'(background|color|border)[a-z-]*\s*:\s*#[0-9a-fA-F]{3,8}','gi')
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id=<webdesign> AND p.name='tool-mind-map') m;
SELECT item_key, status, spec->>'issue' FROM site_work_items
WHERE site_id=<webdesign> AND item_key='ported_tool_fix:tool_health:mind-map:'||<webdesign>::text;
```

## Negative controls (each with a DEMAND control so a zero is evidence)

```sql
-- 1. non-tool ported pages (/learn/ prose ports) must draw nothing from the tool checks
SELECT count(*) FROM site_work_items w JOIN pages p ON p.id=w.page_id
WHERE w.site_id=<webdesign> AND w.item_type IN ('ported_tool_fix','audit_tool','improve_tool')
  AND w.created_at > '<sweep-start>' AND p.page_type <> 'tool';           -- expect 0
-- demand control: the same pages DO exist and DO carry the ported-page component
SELECT count(*) FROM pages p JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id=<webdesign> AND p.page_type<>'tool' AND pc.component_id='a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef'; -- expect >0
-- 2. the 4 multi-section fleet tool pages (idea.uk 3, leopardessconsulting 1): nothing
SELECT s.domain, p.name FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.page_type='tool' AND p.status='active'
  AND (SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE pc.page_id=p.id AND cc.is_active) > 1
  AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE pc.page_id=p.id AND cc.component_level='tool' AND cc.is_active);
```

## Instance-pin proof (drive one ported audit_tool item)

After 425 is applied and a ported `audit_tool` item has run:
```sql
SELECT os.collected_data->'tool_data'->>'page_id' AS loaded_page,
       w.spec->>'page_id' AS item_page,
       os.collected_data->'tool_data'->>'component_level' AS level
FROM orchestration_states os JOIN site_work_items w
  ON os.collected_data->'input_data'->>'work_item_id' = w.id::text
WHERE w.item_type='audit_tool' AND w.spec->>'component_id'='a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef'
ORDER BY os.created_at DESC LIMIT 3;   -- loaded_page must equal item_page
```

## Seeds

Dry-run (aborted transaction) — the files start `ROLLBACK; BEGIN;` so the runner refuses
to probe them; do it by hand:
```bash
sed '$ s/^COMMIT;$/ROLLBACK;/' docs/agent_docs/sql_for_agents/425_tool_auditor_ported_instances.sql > /tmp/425_DRY.sql
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < /tmp/425_DRY.sql
```
Round-trip proof (apply body + rollback body in one aborted txn, compare to `_before`) — the
exact script is in NOTES 2026-08-15. **Gotcha it caught:** `to_jsonb('literal')` on an untyped
string fails "could not determine polymorphic type"; the forward seeds only worked because
`||` concatenation types the literal. Cast `::text` in the rollbacks.

Apply: run the file itself (it INSERTs its own `schema_migrations` row — do NOT also
`--record-only`).

## Live latent hazard (found 2026-08-15, NOT repaired by this lane)

```sql
SELECT length(html_template), left(html_template,80), updated_at FROM content_components WHERE id='a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef';
SELECT version_number, created_at, changed_by, length(html_template) FROM component_versions WHERE component_id='a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef' ORDER BY 1;
-- propagation check WITH a control: pc.updated_at all == the write instant means the bulk
-- build_status flip, not a re-render; confirm rendered_html content, never a class-name LIKE alone
SELECT s.domain, pc.build_status, count(*), min(pc.updated_at), max(pc.updated_at)
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.component_id='a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef' GROUP BY 1,2;
```
Restore = seed 208's passthrough (`component_versions` v1) — the owning lane's call
(webdesign_couk / adoption-pipeline).

## Did the Go ship? (deploy verification — council debug_historian seat asked for the recipe)

Per CLAUDE.md 2026-08-11: ask the SERVICE, do not `strings` the binary and do not grep a symbol.
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=3000 | grep -m1 'build provenance'
# → the commit the running chassis was built from; then:
git merge-base --is-ancestor 25f92a967 <that sha> && echo "281 Go is in the running image"
```
Startup line scrolls on a busy chassis; if absent from `--tail`, probe the binary with a KNOWN
value AND a control: `kubectl exec <pod> -- grep -aq "<sha>" /proc/1/exe` (present) plus a sha
that must be absent. Per SERVICE, never per fleet (bugs_open/249).

## The two load-bearing absence claims, attached as queries (council prior_art seat asked)

```sql
-- "0 tool PLANs, so no auto-fixer" (D1). Expect 0 / ~89 on 2026-08-15.
SELECT count(*) FILTER (WHERE categories ? 'acceptance_criteria') AS tool_plans,
       count(*) FILTER (WHERE categories ? 'needs_criteria')       AS needs_criteria
FROM doc_notes WHERE subject_type='tool';
-- "update_component_html has ONE live consumer" (D4). Text search over the whole
-- default_config JSON, so nested sub_workflow steps are covered. Expect tool-improver only.
SELECT type FROM agent_definitions
WHERE default_config::text LIKE '%update_component_html%'
  AND is_active AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- Writers of html_template in code (fence coverage): expect the 6 files named in
-- component_write_guard.go's fence section.
--   grep -rn "UPDATE content_components" --include=*.go platform/ internal/ pkg/ | grep -v _test
```

## Seed safety, answered (council guardian + debug_historian seats)

- Double-active-row landmine: both seeds' pre-flight `RAISE` unless exactly ONE active,
  non-snapshot, non-deleted row is in the un-migrated shape; the post-condition asserts exactly
  one row is fully migrated. Two active rows → count 2 → abort before any write.
- Needle-gate discipline: `snapshot_agent` backup first; every UPDATE's WHERE is gated on the
  pre-state literal (params array / start_step / slot path), so a re-apply is a 0-row no-op and
  the pre-flight refuses it anyway; the prompt needle is counted (=1) before the replace.

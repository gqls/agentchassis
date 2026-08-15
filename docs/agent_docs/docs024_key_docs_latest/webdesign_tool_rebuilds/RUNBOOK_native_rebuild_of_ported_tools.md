# RUNBOOK — native rebuild of ported tools (webdesign.co.uk)

DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
Site: `6b49db8e-d447-4467-8277-4f3018af9897`. Shared ported component: `a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef`.

## Watch the pilot / any build

```sql
SELECT item_type, status, claimed_at, completed_at, error
FROM site_work_items
WHERE site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND item_type='add_tool'
ORDER BY created_at DESC LIMIT 3;
```
Gotcha: the generator run appears in `orchestration_states` as `owner_agent_type='tool-generator'`
with a `build-dispatch-loop` parent; the deploy may be a second item (`tool-deployer`).

## After the build completes — retire the ported slot (per page, guarded)

```sql
UPDATE page_components pc SET build_status='removed', updated_at=now()
FROM pages p
WHERE pc.page_id=p.id AND p.site_id='6b49db8e-d447-4467-8277-4f3018af9897'
  AND p.name='<page_name>' AND pc.slot_name='ported-page'
  AND pc.build_status IN ('deployed','pending')
RETURNING pc.id, pc.build_status;
```
Must return exactly 1 row. `removed` is the assembly-excluded tombstone
(`rerender_single_page_action.go:843`); do NOT delete the row (artefact archive trigger keeps
history on UPDATE, and the pre-state stays recoverable).

Then re-render the page (single-page path; the queue's `page_rerender` item type also works):
name the page, verify the item completes, then grade at the artefact below.

## Grade at the artefact (per tool, before calling it replaced)

```bash
curl -s https://webdesign.co.uk/<tool-url> -o /tmp/t.html
grep -c 'ported-page-section' /tmp/t.html   # must be 0
grep -c '{{\.' /tmp/t.html                  # must be 0
grep -c '<script' /tmp/t.html               # must be ≥1 (tool is interactive)
```
And in the DB: the page has exactly one non-removed slot, `component_level='tool'`, deployed.

## File the NEXT replacement (serial — only when no add_tool is open on the site)

```sql
-- refuses itself via the dedup key if one is still open
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, page_id, affected_url,
   priority, handler_agent, status, created_by, item_key, pipeline)
SELECT p.site_id, 'owner-request', 'add_tool', 'medium',
  'REPLACEMENT (owner-directed, bugs_open/281 arc): rebuild the ported ' || p.title ||
  ' as a native framework tool at the same URL. After deploy, retire the ported-page slot and re-render.',
  jsonb_build_object('name', split_part(p.title,' | ',1), 'function', p.name,
                     'priority', 1, 'complexity', 'simple',
                     'description', '<WRITE THIS from the live tool''s actual behaviour — open the page>'),
  p.id, p.url, 60, 'tool-generator', 'triaged',
  'webdesign-tool-rebuilds', 'add_tool_novel_webdesign.co.uk', 'build'
FROM pages p WHERE p.site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND p.name='<page_name>'
ON CONFLICT DO NOTHING;
```
Gotchas learned the hard way:
- `function` = the page name (keeps the `tool-` prefix); check the fleet-wide unique claim
  first: `SELECT 1 FROM content_components WHERE function='<f>' AND is_active;` must be empty.
- The `description` is the GENERATOR'S functional brief — never put process/replacement notes
  in it (they can end up rendered into the page). Process notes go in the item summary.
- Rich apps (mind-map, meme-generator, micro-cms, pasteboard, logic-architect…): do NOT file —
  PLAN §3.

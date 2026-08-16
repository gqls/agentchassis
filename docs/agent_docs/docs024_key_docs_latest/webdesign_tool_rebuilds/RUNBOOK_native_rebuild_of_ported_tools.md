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

## Is the adopt path live? (bugs_open/286 — do this BEFORE refiling the pilot)

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -aq "88897190e" /proc/1/exe && echo FIX_PRESENT || echo FIX_ABSENT
kubectl -n ai-persona-system exec $POD -- grep -aq "5e075a6f9" /proc/1/exe && echo control_1303_present   # must also be true after a roll that includes it
kubectl -n ai-persona-system exec $POD -- grep -aq "deadbeefcafe0000" /proc/1/exe && echo CONTROL_BROKEN || echo control_absent_ok
```
Only when FIX_PRESENT: rename `sql_for_agents/435_tool_generator_adopt_existing_page_HOLD.sql` (drop `_HOLD`,
fix its two filename literals), apply, then confirm
`SELECT default_config#>'{workflow,steps,save_tool,config,adopt_existing_page}' FROM agent_definitions WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;` → `true`.
Then refile the pilot (INSERT above). Grade the generator run: `create_result.page_adopted = true`,
`page_components` gains ONE row on the EXISTING page id at position 2, `pages` gains NO row.

## Visible-text check on a tool slot (a 13 KB slot can be a shell)

```sql
SELECT id, slot_name, build_status, length(rendered_html),
  length(regexp_replace(regexp_replace(regexp_replace(regexp_replace(rendered_html,'<style[^>]*>.*?</style>','','gis'),'<script[^>]*>.*?</script>','','gis'),'<[^>]+>','','g'),'\s|&[a-z#0-9]+;','','g')) AS visible_chars,
  (rendered_html LIKE '%{{.%') AS raw_tag
FROM page_components WHERE page_id='<page>';
```
`visible_chars = 0` ⇒ hollow; and census the TEMPLATE's `{{.` fields before trusting any re-render.

## Guard on "open items on this page" — the status set that matters

`status IN ('triaged','approved','claimed','pending')` — NOT `NOT IN (complete,cancelled,rejected)`:
`unresolved`/`failed` are dead to the dispatcher and there are dozens per page from the pre-434 era.

## The ab-test revert (done 2026-08-16 ~10:05Z; recipe for any fork that turns out unfit)

Status flips only, html untouched, one txn with pre-state asserts: ported slot `removed`→`deployed`,
fork slot `approved`→`removed`, then an assemble-only `page_rerender` (spec `{domain,page_id,page_name,filename}`,
no `reason`). Grade at the served page: `grep -c '{{\.'` = 0, control string present, ONE tool.

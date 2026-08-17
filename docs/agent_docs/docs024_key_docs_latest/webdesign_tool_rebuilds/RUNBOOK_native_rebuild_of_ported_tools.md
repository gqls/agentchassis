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
grep -c 'class="ported-page"' /tmp/t.html   # must be 0 after replacement — CORRECTED 2026-08-16: 'ported-page-section' was never in the served markup (0 before AND after; fingerprint taken FROM the artefact)
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
- `function` = the page name (keeps the `tool-` prefix). **CHECK THE FLEET-WIDE UNIQUE INDEX AND
  STOP IF IT IS NOT EMPTY — CORRECTED 2026-08-17 after this line's imprecise version cost a build.**
  The gate is `idx_cc_tool_function_unique`, and it is NOT the generator's `already_exists` probe:
  `CREATE UNIQUE INDEX ... ON content_components (function) WHERE component_level='tool' AND
  forked_from IS NULL AND is_active` — **fleet-wide, no `site_id`, forks exempt.** So ask it in its
  own terms, or a library template with no placement on your site sails past every other check and
  the build dies at `save_tool` with SQLSTATE 23505:
  ```sql
  SELECT id, name, created_from FROM content_components
  WHERE function='<f>' AND component_level='tool' AND forked_from IS NULL AND is_active;
  ```
  **Must return 0 rows. A non-empty result is a STOP, not an input to a judgement** — the older
  wording ("must be empty") was right and was talked past by reasoning about the probe instead.
  Known blocked on webdesign.co.uk (4 of 62): `tool-ab-test-calculator`, `tool-bg-remover`,
  `tool-meme-generator`, `tool-prompt-architect`. Do not file these until the library-template
  question is decided — the unblock has fleet-wide blast radius.
- The `description` is the GENERATOR'S functional brief — never put process/replacement notes
  in it (they can end up rendered into the page). Process notes go in the item summary.
- Rich apps (mind-map, meme-generator, micro-cms, pasteboard, logic-architect…): do NOT file —
  PLAN §3.

## Is the adopt path live? — DONE 2026-08-16 (v1.0.1304, stamp `5de6cddbe`); recipe kept because the first attempt was WRONG

The binary carries ONLY its own build sha — grepping YOUR commit (or the previous roll's sha) returns
absent on every roll that isn't built from exactly that commit. So: find the STAMP, then ask git.
```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system logs $POD --tail=400 | grep -m1 'build provenance'          # startup line; scrolls fast on chassis
# if it has scrolled: probe each commit in the build window (pod startTime minus ~40 min):
for s in $(git log --format=%h --since='<start>' --until='<pod start>'); do
  kubectl -n ai-persona-system exec $POD -- grep -aq "$s" /proc/1/exe 2>/dev/null && echo "$s = STAMP"; done
git merge-base --is-ancestor <your-commit> <STAMP> && echo LIVE || echo NOT_LIVE
kubectl -n ai-persona-system exec $POD -- grep -aq "0123456789abcdef0123" /proc/1/exe && echo CONTROL_BROKEN || echo control_ok
```
Seed 435 is APPLIED (15:15Z). Flag check:
`SELECT default_config#>'{workflow,steps,save_tool,config,adopt_existing_page}' FROM agent_definitions WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;` → `true`.
Grade a generator run: `create_result.page_adopted = true`, `page_components` gains ONE row on the
EXISTING page id at position 2, `pages` gains NO row.

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

## Grade the served page — use the `index.html` URL, not the directory (added 2026-08-16, adopt-route pilot)

`https://webdesign.co.uk/tools/aspect-ratio/` is a **404** (3,001 B); the served artefact is
`https://webdesign.co.uk/tools/aspect-ratio/index.html` (200). Take the URL from `pages.url`.
This matters because the 404 body scores `{{\.` = 0 and `class="ported-page"` = 0 — **a perfect
pass on a page that could never have contained either.** Assert the status code first:

```bash
curl -s -o /tmp/t.html -w 'http=%{http_code} bytes=%{size_download}\n' "https://webdesign.co.uk$(psql_url)"
# require http=200 before reading any count below
```

## The generator files its OWN rerender — do not file a second one (adopt route, 2026-08-16)

A completed adopt run leaves `page_rerender` queued with key
`page_rerender_<page_name>_<site_id>_assemble` and spec `{domain, page_id, filename, page_name}` —
already the assemble-only shape. It also files `needs_content_page` (`tool_content:<page_name>:<site>`)
and `nav_drift` (`nav_rebuild:<site>`), and it creates a **guide page** (`<page_name>-guide` under
`/guides/`) plus cross-links. So on this route:

- **Retire the ported slot BEFORE that rerender is claimed.** If it renders first the page serves
  BOTH tools — the ab-test shape. Check the backlog ahead of it:
  `SELECT count(*) FROM site_work_items WHERE status IN ('triaged','approved','pending') AND created_at < '<the rerender item's created_at>';`
- "`pages` gained no row" is a check on the **tool** page only. The guide page IS a new `pages` row
  and is correct output — do not read it as an adopt failure.

## Record the revert handle per tool — the ported ROW, not an archive row

A `build_status` retire fires **no** archive trigger (`…archive_upd` is `AFTER UPDATE OF rendered_html`
and requires the html to differ; the twin is `AFTER DELETE`). `page_component_history` stays empty for
the page. Capture this instead, immediately before the retire:

```sql
SELECT id, slot_name, build_status, position, length(rendered_html) AS len, md5(rendered_html) AS md5
FROM page_components WHERE page_id='<page>' AND slot_name='ported-page';
```
Paste the row into NOTES. Revert = flip that id back to its recorded `build_status` and re-render;
the md5 is how a later session proves the bytes it restored are the bytes that were served.

## Scope the batch correctly — two traps in one query (added 2026-08-16)

```sql
-- the 62 remaining ported TOOLS, with the external-script class derived, not proxied
SELECT p.name, p.url, length(pc.rendered_html) AS bytes,
       (SELECT count(*) FROM regexp_matches(pc.rendered_html,'<script[^>]+src=','gi')) AS ext_scripts
FROM pages p JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id='6b49db8e-d447-4467-8277-4f3018af9897'
  AND pc.slot_name='ported-page' AND pc.build_status='deployed'
  AND p.name LIKE 'tool-%'          -- NOT p.url LIKE '/tools/%'
ORDER BY ext_scripts, bytes;
```

1. **`p.url LIKE '/tools/%'` returns 64, not 63** — `tools-index` (`/tools/index.html`) is a ported
   page and is the listing, not a tool. Filter on `p.name LIKE 'tool-%'`.
2. **`content_data ? 'repair'` is NOT the external-script class, even though both count 13.**
   The intersection is **4**. `repair` is residue from the `webdesign_tools_repair` lane; the class
   the PLAN and TL-032 mean is `<script src=` in the ported html. Using one for the other mis-scopes
   9 tools each way. Derive it with the `regexp_matches` expression above.

Also: `ported-page` slots exist on 97 pages, of which only 64 are under `/tools/` (32 `/learn/`,
1 `/about/`). A site-wide `slot_name='ported-page'` census is NOT a tool census.

## Before REFILING a tool that already has a native component — deactivate it first

`create_tool_component_action.go` ~197–217 probes `content_components → page_components → pages` on
`cc.function` + `component_level='tool'` + `p.site_id` + **`cc.is_active=true`, with no
`build_status` filter**. A component whose placement is `removed` (a withdrawn fork) still matches,
so the generator returns `{already_exists:true}` and writes nothing — the run "succeeds" having done
nothing at all. Deactivate the component, then file:

```sql
UPDATE content_components SET is_active=false, updated_at=now()
WHERE id='<component>' AND function='<tool-function>' AND is_active;   -- expect 1 row
```
The `removed` placement row stays as history. Known case: the ab-test fork
`cd60486c-f5e1-4d80-9676-0d65024f0372`.

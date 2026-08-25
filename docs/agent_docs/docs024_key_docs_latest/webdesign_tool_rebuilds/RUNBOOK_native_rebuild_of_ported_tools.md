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

**Wrap it in a transaction with `DO`/`RAISE` asserts, and beware what an ABORT looks like
(2026-08-19, hit twice in one session).** `min(id)` on a uuid column raises `function min(uuid) does
not exist` — cast (`min(id::text)`) or use `string_agg(id::text, ',')`. When that raise happened in a
post-`UPDATE` assert, psql had ALREADY printed the `RETURNING` row and `UPDATE 1`, then rolled the
whole transaction back: **`UPDATE 1` in the output of an aborted transaction is not a write.** Re-read
the row's `build_status` after any transaction that errored anywhere, and only then believe it.

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
                     'description', '<WRITE THIS from the live tool''s actual behaviour — open the page>',
                     -- ADDED 2026-08-24 (staged_component_build lane) — see the gotcha below.
                     -- WITHOUT THIS KEY THE TOOL GETS NO CROSS-MENTIONS AND NOTHING SAYS SO.
                     'related_pages', jsonb_build_array('<page-name-1>','<page-name-2>')),
  p.id, p.url, 60, 'tool-generator', 'triaged',
  'webdesign-tool-rebuilds', 'add_tool_novel_webdesign.co.uk', 'build'
FROM pages p WHERE p.site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND p.name='<page_name>'
ON CONFLICT DO NOTHING;
```
Gotchas learned the hard way:
- **`related_pages` — NEW 2026-08-24, and omitting it silently costs the tool its cross-mentions.**
  A cross-mention is the one-sentence, in-context reference to the new tool that gets woven into a
  related article (live example, dartsonline `barrel-shapes`: *"…the tungsten percentage vs barrel
  diameter visualiser lets you compare percentages against weight…"*). One is emitted per page you
  name here, and **only** for pages you name here.
  **Every hand-filed `add_tool` on this estate has omitted the key — 0 of 58 since 08-17 carried it,
  against 11 of 11 from `tool-suggester`** (measured 2026-08-24). Until migration 516 (live
  2026-08-21 16:55Z) the omission was masked: the resolver substituted another tool's list, which is
  why nine webdesign.co.uk tools all pointed at the same two articles (`bugs_open/330`). 516 removed
  the substitution — correctly — so the omission now means **no cross-mentions at all**, recorded
  only as an `info` row: `agent_error_log.error_code='tool_crosslink_not_emitted:no_related_pages'`.
  The build succeeds, the page deploys, the tool works. Nothing complains.
  **Name 1–3 EXISTING, ACTIVE, NON-TOOL page names** (`pages.name`, not titles, not URLs — a name
  that does not resolve is skipped silently, and a `tool-`-prefixed one is refused by design):
  ```sql
  SELECT name FROM pages
   WHERE site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND status='active'
     AND name NOT LIKE 'tool-%' ORDER BY name;   -- 37 candidates as of 2026-08-24
  ```
  Pick by topic, not by convenience — the writer is told to place the mention "where it's most
  contextually relevant", so an unrelated page produces a worse sentence, not a neutral one.
  `tool-bayesian-rank` → `learn-algorithms-bayesian-theory` is the shape.
  **⚠ CORRECTED 2026-08-25 — the query below counts ROWS, which is not the artefact, and on this
  site it has never once been able to fail.** Measured that day: `tool_crosslink:%` on site
  `6b49db8e` is **41 `wont_fix` · 15 `deferred` · **22** `failed` · 2 `unresolved` · ZERO `complete`**
  across **80** rows since 2026-08-05 (live UNION archive) — **no cross-mention has ever been written on webdesign.co.uk.**
  Every row this lane has filed since 08-24 is parked the same second it is created, with
  `OWNED_PAGE_GUARD: page-build-handler declares refuse_owned_page and page <id> is
  rebuild_policy=owned` and a spec key `not_dispatchable`. That is `bugs_open/333`'s door working as
  designed on `rebuild_policy='owned'` pages, not a fault — but it means **a filing delivers a
  targeted PARKED FINDING, not a mention.** Keep carrying the key (the finding is filed against the
  right pages and is the raw material when the owned-page route lands); just never report a mention
  from a row.
  **Verify after the build. Ask for the status, and then look at the article:**
  ```sql
  -- expect one row per page you named AND read the status: 'complete' is delivered,
  -- 'deferred' is parked behind the owned-page door and nothing further will happen.
  SELECT status, error, spec->>'page_name' FROM site_work_items
   WHERE item_key LIKE 'tool_crosslink:<function>%';
  ```
  **If you then grep the target article for the tool's name, DATE THE SLOT — the grep passes on copy
  that predates your filing by a week.** Every one of these ported `/learn/` articles already
  carries a hand-authored CTA for its tool (`learn-data-unicode-gremlins` has *"Use our Regex-powered
  cleaner… Launch Text Sanitizer →"*, written 2026-08-15). A name-match is only YOUR mention if the
  slot was written after you filed:
  ```sql
  SELECT pc.updated_at   -- must be LATER than your add_tool's completed_at
  FROM pages p JOIN page_components pc ON pc.page_id=p.id
  WHERE p.site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND p.name='<the named page>';
  ```
  This is the interim. The owner ruled on 2026-08-24 that the system should ASK when the key is
  absent rather than rely on it being remembered — tracked by the `staged_component_build` lane.
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
  ~~**Must return 0 rows. A non-empty result is a STOP, not an input to a judgement**~~ —
  **SUPERSEDED 2026-08-19, and only for the LIBRARY-CLAIM case, which is the only case this index
  can report.** RFC_036 §9.3 is live (`create_tool_component_action.go:249-285`, commit `e24bc9c0f`,
  chassis `v1.0.1316`): when a library entry claims the function the generator now sets
  `forked_from` on the new row, so the partial index no longer fires and `save_tool` completes.
  **PROVEN on demand 2026-08-19 20:57Z** with `tool-ab-test-calculator` (NOTES 20:58Z).
  So the gate INVERTS rather than disappearing — assert the claim's IDENTITY, not its absence:
  ```sql
  -- expect exactly ONE row, and pin its bytes so a misbehaving fork is detectable afterwards
  SELECT id, name, md5(html_template) FROM content_components
  WHERE function='<f>' AND component_level='tool' AND forked_from IS NULL AND is_active;
  ```
  and re-read that md5 AFTER the build (it must be unchanged — the library row must never be written).
  Two things still hold: **verify the running chassis actually carries `e24bc9c0f`** before relying on
  this (`git merge-base --is-ancestor e24bc9c0f <the pod's stamp>`, plus a positive and a negative
  literal probe on BOTH replicas — the tag alone is not the evidence), and **a tool whose LOCAL fork
  is still `is_active` short-circuits at the `already_exists` probe instead**, so the fork branch is
  never reached (deactivate it first — see "Before REFILING a tool that already has a native
  component"). Remaining blocked on webdesign.co.uk: **1**, `tool-meme-generator`, and only because it
  is a Phase B rich app that goes last by the owner's instruction — not because the platform refuses it.
- **Measure the MARGIN on the target page's queued `page_rerender` before you file (added 2026-08-19,
  CORRECTED 2026-08-20 — the first version of this bullet said "refuse if one is open", which would
  have blocked every filing for three hours the very next morning when a 121-item site sweep queued one
  rerender per page. A guard that fires on an ordinary day is an outage, not a guard).** What matters is
  whether that assemble can fire inside the build-plus-retire window, so count what is ahead of it and
  refuse only on a thin margin:
  ```sql
  SELECT count(*) FROM site_work_items
  WHERE site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND item_type='page_rerender'
    AND status IN ('triaged','approved','pending')
    AND created_at < (SELECT min(swi.created_at) FROM site_work_items swi JOIN pages p ON p.id=swi.page_id
                       WHERE p.name='<page_name>' AND p.site_id='6b49db8e-d447-4467-8277-4f3018af9897'
                         AND swi.item_type='page_rerender' AND swi.status IN ('triaged','approved','claimed','pending'));
  ```
  At the measured drain rate of ~0.7/min, 20 items ahead is ~25 minutes of headroom; under that, wait.
  The reason to look at all is not the one further down this file ("don't file a second rerender"): an assemble that is ALREADY QUEUED can be claimed inside the 60-to-100 seconds between
  the build completing and your retire, and then the page publishes BOTH tools. Measured that day: the
  ab-test page's rerender had been filed by the `rerender-pages` sweep **21 minutes before** the build,
  and one of the 117 rerenders that sweep queued for this site was sitting on the very page being
  rebuilt. It also **dedupes the generator's own rerender away** (one open `page_rerender` per page by
  `item_key`), so after the retire you must confirm a pending rerender still exists and file an
  assemble-only one if it does not.
  ```sql
  SELECT id, status, created_by, created_at FROM site_work_items
  WHERE page_id='<page>' AND item_type='page_rerender' AND status IN ('triaged','approved','claimed','pending');
  ```
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
  [ -n "$s" ] || continue      # an EMPTY pattern matches every byte and prints "= STAMP" for nothing (WRONG_CALLS 2026-08-19 late)
  kubectl -n ai-persona-system exec $POD -- grep -aq "$s" /proc/1/exe 2>/dev/null && echo "$s = STAMP"; done
# and before you merge-base against anything you copied from a NOTES line: `git cat-file -t <it>` must say commit —
# a 9-hex IMAGE DIGEST prefix (e.g. 2d0d3defc for v1.0.1316) looks exactly like a short sha and is not one
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

## Grade the RUN, not the work item — a failed build reports `complete` with `error` NULL

Measured 2026-08-17 (rebuild #2): item `complete`, `error` empty, **nothing built**. The generator
orchestration ended at `current_step='complete_error'` with `orchestration_states.error` NULL and the
real message one level down. Ask for all three signals at once:

```sql
SELECT current_step,
       collected_data#>>'{create_result,page_adopted}'   AS adopted,      -- want: true
       collected_data#>>'{create_result,already_exists}' AS short_circuit,-- want: NULL
       collected_data#>>'{__step_error,failed_step}'     AS failed_step,  -- want: NULL
       collected_data#>>'{__step_error,message}'         AS why
FROM orchestration_states
WHERE owner_agent_type='tool-generator' AND created_at > '<the item''s claimed_at>'
ORDER BY created_at DESC LIMIT 1;
```
`agent_error_log` (`occurred_at`, not `created_at`) carries the same failure independently — filter
`agent_type='tool-generator'` over the window. **A NULL `create_result` prints as a blank line under
`psql -At`, which reads exactly like "no run happened" — it means the run died before `save_tool`.**

## The retire race is against YOUR ATTENTION, not the queue (added 2026-08-17, after losing it)

The generator queues its own rerender the moment the build completes. Measured margins between build
completion and that rerender being claimed: **~45 min, ~2 min, ~26 min, ~96 min.** There is no floor.
If the build lands while nobody is looking, the page serves BOTH tools until someone notices.

- **Do not file a rebuild you cannot attend.** The window opens when the `add_tool` item completes,
  and the build itself takes under a minute once claimed.
- **"Attend" means the SESSION'S TURN STAYS ALIVE, polling, from filing until the retire lands —
  a background watcher is NOT attendance (lost again 2026-08-20, WRONG_CALLS).** A watcher's FIRING
  is on time; its DELIVERY to the session waits for the next interaction and is unbounded — measured
  six hours, during which a sweep rerender assembled the oklch page with both slots live and the
  public page served two stacked tools all afternoon. Same day, same mechanism, smaller lags: 50 min
  (saved only by an unrelated queue freeze) and ~2 min (worked). If the turn must end, do not file.
- **If you lose it, it is repairable and self-healing** — no data is lost. Retire the ported slot as
  normal; any queued `page_rerender` for that page then assembles it correctly. Check for one before
  filing a new one:
  `SELECT id,status FROM site_work_items WHERE page_id='<page>' AND item_type='page_rerender' AND status IN ('triaged','approved','claimed','pending');`
- **Confirm the damage at the served page rather than assuming it** — a double-tool page is obvious
  by size (the SVG page went ~12 KB to 19,415 B) and by `class="ported-page"` being 1 when it should
  be 0, with both tools' control ids present.

## ALWAYS cache-bust the served-page grade (added 2026-08-18, after a false FAIL)

```bash
curl -s -o /tmp/t.html -w 'http=%{http_code} bytes=%{size_download}\n' \
  "https://webdesign.co.uk/<url>?cb=$(date +%s)"
curl -sI "https://webdesign.co.uk/<url>" | grep -i 'last-modified'   # compare to the rerender's completed_at
```
`cache-control: public, max-age=3600` on these pages, and **the recipe warms that cache itself** when
you fetch the live tool to write the brief — roughly 40 minutes before the rerender lands. The
asymmetry: a **pass** through a stale cache is impossible (a stale copy serves the OLD page, so it
cannot show the new ids), but a **fail** through one is exactly what staleness looks like. Never
report a served-page failure from an un-busted fetch.

## ✅ LIVE 2026-08-21 — re-fixes need ONE FILING (`"replace_existing": true` in the spec); the three hand steps below this section are RETIRED for re-fixes (proven end-to-end on tool-oklch-picker, `bugs_closed/331` §16; the ported→native FIRST replacement is unchanged). Two rules learned live: the brief must NAME the copy to keep (the non-hollow gate refuses a regeneration that drops >50% visible text — that is it working); and grade the RUN (`create_result.regenerated='true'`, same component_id), never the item. Original section as written 08-19:

`bugs_open/331` (fix built, council `7a82c943` submitted, inert until the next chassis roll AND
`sql_for_agents/496_tool_generator_replace_existing_HOLD.sql` is applied): `create_tool_component`
gains a per-ITEM `replace_existing` input. For a RE-FIX of a tool that is already native on the site,
put `"replace_existing": true` in the `add_tool` item's `spec` and file — **no deactivate, no rename,
no retire race**: the generator regenerates the existing component IN PLACE (same `component_id`), rewrites
the live slot's `rendered_html` in the same transaction (`page_component_history` archives the old bytes
— the revert handle the md5 step used to stand in for), and the page never holds two tool slots.
Grade the RUN with the same query as today: `page_adopted='true'`, no `already_exists`, **plus
`create_result.regenerated='true'`**; the component id will be the OLD id. The PORTED→native first
replacement (adopt route) is unchanged — the retire step still applies there.
**Until roll + seed, every section above stands as written.** How to know: the chassis stamp's ancestry
must include the TL-047 commit, and `SELECT default_config#>>'{workflow,steps,save_tool,config,replace_existing!}'
FROM agent_definitions WHERE type='tool-generator' AND is_active AND NOT COALESCE(is_snapshot,false)` must
return `input_data.spec.replace_existing`.

## ⚠ A retire can be UNDONE by a section edit (added 2026-08-22 — four slots resurrected, `bugs_open/360`)

The assemble race above is not the only way a retired slot returns. `check_literal_markdown` scans
tombstones (no build_status filter) and its section-editor `rendered_html_transform` route writes
`build_status='approved'` unconditionally — on 2026-08-21 it un-retired FOUR of this lane's
tombstones and the sweep published four double-tool pages for ~19 h. Until 360's fix ships:

- **At the END of every retire's attendance window, re-read the tombstone:**
  ```sql
  SELECT build_status, updated_at FROM page_components WHERE id='<retired row id>';
  ```
  Must still be `removed`; an `updated_at` newer than your retire is a resurrection in progress —
  re-retire (same guarded UPDATE; the bytes may legitimately differ now, the markdown fix is
  content repair) and confirm a rerender follows.
- **Periodically (and before citing any old serve-grade): count at the page** — exactly one
  non-`removed` slot per rebuilt page. Whole-batch check:
  ```sql
  SELECT p.name FROM pages p
  JOIN page_components pp ON pp.page_id=p.id AND pp.slot_name='ported-page' AND pp.build_status<>'removed'
  JOIN page_components nt ON nt.page_id=p.id AND nt.slot_name<>'ported-page'
  WHERE p.site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND p.name LIKE 'tool-%';
  -- any row = a resurrected tombstone on a rebuilt page
  ```
- An open `literal_markdown` or `section_edit` item naming a rebuilt tool page is the early
  warning — check it BEFORE it completes.
- Phase C addendum: the same applies to the external S3 asset's retirement — nothing edits those
  back today, but the slot half of a Phase C retire is exactly this trap.

## Phase C — the external-asset half of a retire (added 2026-08-22, learned on #28 blueprint-compiler)

The slot half is the normal recipe. The FILE half currently has NO mechanism (`bugs_open/365`):
`retract_asset_files` refuses anything outside `/assets/` by design ("pages, feeds and chrome are
page-retraction's or nobody's"); page-retraction owns pages; webdesignport is import-only. Per tool:

1. The serve-grade MUST carry `src="<sidecar>"` as a negative (0 on the new page) — that is the
   half that protects visitors.
2. Dispatch the DRY-RUN retraction anyway and record the refusal — it is the evidence the file is
   orphaned-but-present (recipe: staged_component_build/scripts/RETRACT_gaswholesalers_logo_jpg.sh
   shape, `input_data.paths=["/tools/<slug>/script.js"]`; READ THE LIVE STEP CONFIG for dry_run
   first, per the 08-22 LANDMINE). Find the run BY PAYLOAD (`collected_data->'input_data'->>'paths'
   LIKE '%<slug>%'`) — the printed correlation is not how rows are found, and check your CLOCK
   against the kcat pod-name epoch before calling a row "old".
3. Append the orphan path to the list in NOTES; cleanup is ONE batch when 365's candidate ships.
⚠ `/tools/assets/webdesign-couk-header.js` is shared by EVERY ported page — retired with the LAST
one, never per-tool.

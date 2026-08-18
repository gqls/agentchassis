# RUNBOOK — ai-agent-orchestration.com improvement

Commands that were hard to get right, with the gotcha attached. Site id
`2a8ebf9c-20a2-4c39-b191-840b012371da`.

---

## R1 — Measure contrast at the artefact (the only instrument that sees it)

```bash
timeout 300 python3 scripts/render_audit.py --json out.json \
  https://ai-agent-orchestration.com/index.html \
  https://ai-agent-orchestration.com/about.html \
  https://ai-agent-orchestration.com/pricing.html \
  https://ai-agent-orchestration.com/services.html
```

**Gotchas.**
- It needs a Chromium; it finds one itself (`$CHROME`, the Playwright cache, `/usr/bin/chromium`).
  **`import playwright` is NOT available in this environment** — the script shells out to
  `chrome --headless=new --dump-dom` instead. Do not write a probe that imports playwright; it
  fails with `ModuleNotFoundError`. Reuse this script's own technique (below, R2).
- **Filter `overImage` before quoting a total.** 47 raw findings were 44 firm + 3 approximations;
  the adapter itself calls an over-image backdrop unknown, and a firm/approximate mix quoted as
  one number overstates the defect.
- `images` in the JSON is *confirmed-broken after a serial re-check*, and `verify_broken` **skips
  any image with an empty `src`**. So a page of empty `<img>` tags reports `broken=0`. An
  `images_reported=5, broken=0` therefore means "5 images failed in-browser and none survived
  re-checking" — which on `index` is 5 empty srcs, not 5 healthy images. Read both numbers.

## R2 — Ask the browser for a computed token (never grep the stylesheet)

Reuses `render_audit.py`'s injection technique: fetch the page, add `<base href>` so relative
assets still resolve against the live origin, inject a probe that appends `<pre id="AUDIT_RESULT">`,
then read it back out of `--dump-dom`. Full script:
`scratchpad/probe2.py` pattern in `NOTES`; the load-bearing part is

```python
sys.path.insert(0, os.path.abspath("scripts")); import render_audit as ra
chrome = ra.find_chrome(); page = ra.fetch(url)
inj = re.sub(r"<head([^>]*)>", r"<head\1><base href='%s'>" % base, page, count=1)
inj = inj.replace("</body>", "<script>%s</script></body>" % PROBE)
subprocess.run([chrome,"--headless=new","--disable-gpu","--no-sandbox",
                "--virtual-time-budget=10000","--dump-dom","file://"+path], ...)
```

**Why it matters here:** the served stylesheet contains `h3, .site-footer h4 { color: #ffffff; }`
— white, legible, and **not the winning declaration**. The component's own embedded `<style>`
block in `rendered_html` overrides it. Reading the site stylesheet would have produced a
confidently wrong answer.

## R3 — Extract a component's embedded CSS rule

The rule lives inside `page_components.rendered_html`, not in any stylesheet file.

```sql
SELECT p.name, pc.slot_name,
       regexp_replace(substring(pc.rendered_html from '[^{};]*background[^;]*(#fff|255, *255, *255)[^;]*'),
                      '\s+',' ','g') AS decl
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='<site>' AND pc.rendered_html ~* 'background[^;]*(#fff|255, *255, *255)';
```

**Gotcha.** A naive `substring(... from 'h3[^{]*\{[^}]*color[^}]*\}')` matches the page's **prose**
(`<h3>Agile Orchestration Architecture</h3>` … ) long before it matches CSS, and returns a
paragraph of marketing copy that looks like a failed query. Anchor on `<style>` or on a
declaration-shaped pattern, and be aware `rendered_html` is multi-line — a bash line-oriented
split over psql output silently drops every component.

## R4 — Census the `<img>` srcs for a site

```sql
WITH srcs AS (
  SELECT p.name, pc.slot_name,
         (regexp_matches(pc.rendered_html,'<img[^>]*src="([^"]*)"','g'))[1] AS src
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='<site>')
SELECT name, slot_name, CASE WHEN src='' THEN '(EMPTY)' ELSE src END FROM srcs ORDER BY 1,2;
```

Confirm a 404 over HTTP rather than from the work item, which only compares against the `assets`
table: `curl -s -o /dev/null -w '%{http_code}' <url>`.

## R5 — Check whether a handler actually repairs, before routing work at it

Two steps, both cheap, and skipping either is how work gets routed at a handler that removes the
thing you wanted built.

```sql
-- 1. what does it DO?  (an action list is enough to spot a triage-only handler)
SELECT jsonb_pretty(jsonb_path_query_array(default_config,'$.workflow.steps.*.action'))
FROM agent_definitions WHERE type='<handler>'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 2. where has it run, and what did the ARTEFACT look like afterwards?
SELECT s.domain, w.status, count(*) FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.item_type='<type>' AND w.handler_agent<>'' GROUP BY 1,2;
```

Then go and look at that site's artefact (R4). `image-source-unsatisfiable-handler` reads as the
obvious fix for a missing image and its only precedent site now has **zero** `<img>` tags.

## R6 — Fleet palette outlier check

```sql
SELECT s.domain,
       ss.data->'palette'->'reference_values'->>'primary' AS primary,
       ss.data->'palette'->'reference_values'->>'surface' AS surface,
       (ss.data->'palette'->'reference_values'->>'primary')
     = (ss.data->'palette'->'reference_values'->>'surface') AS degenerate
FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE ss.is_current AND ss.aspect='design_intent'
  AND ss.data->'palette'->'reference_values' ? 'primary'
ORDER BY degenerate DESC NULLS LAST;
```

**Gotcha.** The table is `site_specs` with columns `aspect` / `data` / `is_current` — **not**
`spec_type` / `content`. `\d site_specs` first.

## R7 — Is a page repairable at all?

```sql
SELECT p.name, count(*) comps, count(*) FILTER (WHERE pc.content_data IS NULL) nulls,
       min(pc.updated_at)::date oldest, max(pc.updated_at)::date newest
FROM pages p JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id='<site>' GROUP BY 1 ORDER BY oldest;
```

A page with `nulls = comps` **cannot be re-rendered** — `rerender_page_sections` rebuilds a
section from `content_data` and there is none. It must be rebuilt through the framework.
NEVER restore from `page_component_history` (its `component_id` is NULLed by
`ON DELETE SET NULL`; see `bugs_closed/194` §4).

## R8 — Propagate a component-TEMPLATE change to the live pages

A `content_components.html_template` edit is inert until each placement re-renders. Getting this
wrong is silent: the wrong route reports **success** and ships the old bytes.

```sql
-- Page-scoped rerenders for every page carrying a changed component, ON ONE SITE.
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             priority, handler_agent, status, created_by, spec)
SELECT DISTINCT p.site_id, 'side_effect', 'build', 'page_rerender', 'low',
       'Rerender page after template fix: ' || p.name,
       80, 'page-rerender', 'triaged', '<who-you-are>',
       jsonb_build_object('reason','template_changed','page_id',p.id::text,
                          'page_name',p.name,'domain',s.domain)
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.component_id IN (<the component ids>)
  AND p.site_id = '<site>'
  AND p.rebuild_policy IS DISTINCT FROM 'owned'
  AND NOT EXISTS (SELECT 1 FROM site_work_items w
                  WHERE w.site_id=p.site_id AND w.item_type='page_rerender'
                    AND w.spec->>'page_id'=p.id::text
                    AND w.spec->>'reason'='template_changed'
                    AND w.status IN ('detected','triaged','claimed'));
```

**Gotchas, all four of which cost something if missed.**

- **`spec.reason` is what selects the code path.** `page-rerender.check_rerender_mode` routes to
  `rerender_sections` (regenerates from `content_data` + the template) ONLY for
  `image_landed | section_data_resolved | cta_links_stale | template_changed`. Every other value,
  **including none**, falls to `render_page`, which assembles the STORED `rendered_html` — old CSS,
  green status. Read the live condition before trusting this list; it has grown:
  ```sql
  SELECT default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
  FROM agent_definitions WHERE type='page-rerender'
    AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  ```
- **Scope to the pages that carry the component; never fire a site-wide `needs_rerender`.** On this
  site `privacy` and `terms` hold **permanently locked** `generic-text-block` components, and a
  rerender aimed at a locked positionally-named section **duplicates** it (`bugs_open/189`). Count
  `page_components` for the page before and after if you ever do hit one.
- **Copy the shape from the LIVE agent row, not from the migration that introduced it.**
  `460_template_changed_rerender_reason.sql` puts `p.filename` in the spec and **`pages` has no such
  column** — that is what `461_fix_460_…` exists to fix. The live query is the corrected one:
  ```sql
  SELECT default_config #>> '{workflow,steps,create_rerender,config,query}'
  FROM agent_definitions WHERE type='component-template-fixer'
    AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  ```
- **`p.rebuild_policy IS DISTINCT FROM 'owned'`** — an `owned` page refuses with
  `save_page_sections: overwrite: REFUSED` rather than silently doing nothing, but filtering keeps
  the queue honest.

Then verify at the artefact, **by colour pair** (R1). A pair present in the after-set and absent
from the before-set is a regression you introduced — the check migration 456 lacked, which is how
its `.stats-cta` regression survived a net 44→33 improvement.

## R9 — Apply ONE migration without sweeping other lanes'

`run-migrations.sh --apply` takes **every** pending file; on 2026-08-18 that was 17 files from at
least six lanes. Apply yours alone, then record it:

```bash
# 0. rehearse: the real file, guards and all, with the final COMMIT swapped for ROLLBACK
sed 's/^COMMIT;$/ROLLBACK;/' <file>.sql > /tmp/rehearsal.sql
kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < /tmp/rehearsal.sql
# 1. apply (same invocation the runner itself uses)
kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < <file>.sql
# 2. record
./scripts/migration/run-migrations.sh --record-only <file>.sql --note "<what you checked>"
```

**Gotchas.**
- **A dry-run `SELECT` proves your ANCHORS match; only the rehearsal proves your GUARDS pass.** Mine
  passed the anchor dry run and would have failed at COMMIT on a guard I had scoped too widely.
- **"Pending" ≠ "unapplied".** The listing shows files applied by hand and never recorded. 460 was
  listed pending while `template_changed` was already live in `page-rerender`. **Check the live row,
  not the ledger**, before concluding a mechanism does not exist.
- **`--no-probe` for a fast listing.** The default probe executes every pending file inside a doomed
  transaction, which took longer than a 240 s timeout here.
- **Next number = highest in the directory + 1, and numbers still collide** (457 and 458 each name
  two unrelated migrations). Re-check immediately before writing: 470 appeared during this session.

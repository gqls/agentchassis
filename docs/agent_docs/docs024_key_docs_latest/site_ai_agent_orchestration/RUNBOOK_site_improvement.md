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

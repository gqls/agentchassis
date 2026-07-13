# Component-to-Asset Pipeline — Two Architectural Concerns

Context: 2026-05-19 debug session on gaswholesalers.com revealed two
gaps in how content_components reference external static assets. Both
manifested as the same bug (latest-news section on index.html silent
because `/tools/assets/latest-news.js` doesn't exist) but stem from
distinct architectural patterns worth addressing separately.

## Concern 1: External script paths aren't enforced or auto-produced

### What we have

Several `content_components` rows have templates like:

```html
<section data-component="latest-news" class="latest-news-section">
  <div class="container">
    <h2 class="section-heading">{{.headline}}</h2>
    <div class="news-grid" id="news-container">
      <noscript>...</noscript>
    </div>
  </div>
</section>
<script src="/tools/assets/latest-news.js"></script>
```

The template names an external script. The component definition also
has `js_content` (2174 chars for latest-news, 3092 for news-listing)
which presumably IS the script body intended to live at that URL.

### What's missing

Nothing in the rerender pipeline writes `js_content` out to the file
the template references. Build evidence:

```bash
$ git show origin/master:gaswholesalers.com/tools/assets/latest-news.js
fatal: path '...' does not exist in 'origin/master'

$ git ls-tree origin/master gaswholesalers.com/tools/assets/
(only the js/ subdirectory exists, holding snippets.js)
```

So the browser loads index.html, hits the `<script>` tag, requests
`/tools/assets/latest-news.js`, gets a 404, the news-grid stays empty,
no errors visible above the fold.

### The bigger problem

A new developer creates an `industry-news` component with:

```html
<script src="/tools/assets/industry-news.js"></script>
```

Adds the `js_content`, commits, the platform deploys, the section
appears on a page — but silently broken. No error in the rerender
pipeline. No warning at commit time. The component LOOKED right in
preview because the headline and subheadline are static HTML; only the
data-loaded content is missing. Easy to ship without noticing.

### Stronger pattern — two options

**Option A: fail loudly at component creation time.** When a
content_components row is inserted or updated and its `html_template`
contains a `<script src="/tools/assets/{filename}">`, validate that
`js_content` is non-empty AND that the pipeline that writes
`js_content` to git includes this component. If either check fails,
reject the write with a clear error.

**Option B: co-locate JS in the rendered output.** Stop using external
scripts for component-specific JS. Inline the `js_content` directly
into the rendered HTML at the point where the template currently has
`<script src=...>`. This eliminates the asset-file-must-exist coupling
entirely.

Trade-offs:
- Option A keeps the file-based caching benefits but requires a real
  asset-deploy step. The step needs to know which sites need which
  files. Probably keyed off page_components → content_components.
- Option B duplicates the same JS across every page that uses the
  component (no caching), but it's bulletproof. The size cost is
  small — ~3KB per news component.

For latest-news and news-listing specifically, option B is preferable.
The JS is small, the inlining makes the component self-contained, and
it matches the working pattern already in use elsewhere (see
news.html's full listing which inlines its fetch logic directly).

For larger JS payloads — e.g. a charting library or a complex
interactive tool — option A becomes more attractive because inlining
would bloat every page.

A reasonable rule:
- `js_content` < 5KB → inline at render time (option B)
- `js_content` >= 5KB → must have a corresponding deploy step (option A
  enforcement)

### What to do for current state

The two existing affected components — `latest-news` (id
`77dafa26-...`) and `news-listing` (id `11d4dc21-...`) — should have
their templates updated to remove the `<script src="..."></script>`
line and have the renderer inject the `js_content` inline. That fixes
the immediate bug AND doesn't depend on a deploy step that may not
exist for all paths.

## Concern 2: Component data file paths are convention, not enforced

### What we have

The `latest-news` JS fetches:
```javascript
fetch("/data/latest-news.json")
```

The `news-listing` JS fetches:
```javascript
fetch("/data/news-archive.json")
```

These paths are hardcoded in the JS. There is no declaration on the
component saying "this component depends on file X being present at
path Y." The expectation is purely by convention: whoever creates the
component knows that something elsewhere produces files at those paths.

### What's there for declaring it

`content_components.data_sources` exists as a `text[]` column. Today
it's lightly used. It could declare the dependency:

```sql
UPDATE content_components
SET data_sources = ARRAY['/data/latest-news.json']
WHERE function = 'latest-news';
```

But nothing reads `data_sources` to verify the files exist at deploy
time. So having the declaration without enforcement is no better than
having no declaration.

### The bigger problem

A developer creates a new component that fetches `/data/insurance-news.json`.
The content-feed pipeline currently produces `/data/news-archive.json`
and `/data/latest-news.json` but not insurance-news. The component
appears on a page, the section renders empty in the browser. No error
in the rerender pipeline. Same silent failure as concern 1.

### Stronger pattern

Three things, in order of effort:

1. **Use `data_sources` consistently.** Every component that fetches
   external JSON declares it. Backfill the existing components.

2. **Validate at deploy time.** When a page-rerender claims a work item
   and assembles a page, before deploying, check that every referenced
   `data_sources` file exists for the site (either in
   `content_feed_items` query that produces the JSON, or as a stub in
   git). If a file is missing AND no pipeline step will produce it,
   block the deploy or commit a stub with a clear "no data yet" payload.

3. **Auto-stub missing files.** When asset-rendering runs and finds a
   declared data source that has no producer, write a stub:
   ```json
   {
     "items": [],
     "items_total": 0,
     "updated_at": null,
     "note": "data source not yet producing — placeholder generated by asset pipeline"
   }
   ```
   The page then renders cleanly with a "no items yet" state instead of
   silent failure.

### Less ambitious version

If full enforcement is too much, even step 1 alone (backfill
`data_sources` accurately) gives us a queryable surface for monitoring:

```sql
-- What data files do active components on this site need?
SELECT DISTINCT unnest(cc.data_sources) AS required_file
FROM content_components cc
JOIN page_components pc ON pc.component_id = cc.id
JOIN pages p ON pc.page_id = p.id
WHERE p.site_id = '<site_id>'
  AND cc.is_active = true
  AND cc.data_sources IS NOT NULL
  AND cardinality(cc.data_sources) > 0;
```

Run that query against git's actual file list — any required_file
missing from git is a known gap. Surfaces silent breakage without
requiring renderer changes.

## How these two relate

Both stem from the same root pattern: **a component template
implicitly assumes the existence of an external file (script OR data),
but the pipeline doesn't guarantee that file exists.**

Concern 1 is the JS side. Concern 2 is the data side. Both manifest as
empty sections in the browser with no server-side error.

The same fixes apply at conceptual level:
- Declare the dependency in the component definition
- Either produce the file deterministically OR inline it OR stub it

## Where in code this matters

The renderer that turns a `content_components` row + `page_components`
data into rendered HTML is `rerender_single_page` action — that's
where inlining of `js_content` should happen if we go option-B in
concern 1.

The pipeline that produces `/data/*.json` files is the content-feed
flow (`content-feed-orchestrator`, `feed-ingester`, `feed-triage`).
That's where the auto-stub for declared `data_sources` would live in
concern 2.

The asset deploy step that would write `js_content` to
`/tools/assets/<function>.js` does NOT currently exist in the
rerender-pages or webdesign-agent workflows. Its absence is the
proximate cause of today's bug. It's worth confirming whether such a
step exists in any other workflow and got missed for gaswholesalers, or
whether it was never built. Quick check:

```bash
grep -rn "tools/assets" /path/to/chassis/source  # find references
```

```sql
-- Check if any agent_definition references tools/assets in its workflow
SELECT name, type
FROM agent_definitions
WHERE default_config::text LIKE '%tools/assets%';
```

If those return nothing, the step was never built. If they return a
workflow, we just need to wire it into rerender-pages.

## Forward-looking architectural fit

The user's framing for the future is: multiple news components on a
site, each potentially filtering differently, each styled differently.
The current architecture supports this well IF concerns 1 and 2 are
addressed:

| Concern | Where it lives | Per-component? |
|---|---|---|
| Structural HTML | `content_components.html_template` | Yes — already |
| Component styling | `css_snippets` keyed on function | Yes — already |
| Component behavior (JS) | `content_components.js_content` | Yes — but needs to actually be served |
| Data source URL | hardcoded in JS today, should be in `data_sources` | Should be |
| Filter logic | inline JS today, could parameterize via content_brief | Should be |

The architectural pieces are there. The wiring between them is what
needs filling in.

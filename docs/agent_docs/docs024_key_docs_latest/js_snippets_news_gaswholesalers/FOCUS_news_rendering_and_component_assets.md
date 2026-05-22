# FOCUS — News rendering and the component-asset pipeline

How news (and any data-driven component) renders and deploys: the three-layer
model, the `files_field` fix that made component JS actually reach git, the
unenforced external-asset coupling that silently breaks sections, the
rerender-pages workflow findings, and the JS-snippet renderer deliverable.

> Consolidated from `006_news_feed_pipeline_addendum_rendering.md`,
> `component_asset_pipeline_concerns.md`, `rerender_pages_workflow_findings.md`,
> the fix-plan half of `findings_and_plan_news_visual.md`, and
> `guidelines_compliance_check_1_.md`. The two debugging-guide entries that
> were in the news-visual doc (css-snippets-missing; migration-vs-stale-render)
> are in `016_debugging_guide_addenda.md`.

---

## News rendering & asset deployment architecture

_Discovered and fixed 2026-05-19/20 during the gaswholesalers.com news rollout. Destined for `006_news_feed_pipeline_v2.md`._

## Rendering layer — three separate concerns

News rendering splits cleanly into three layers that are produced and
deployed independently. Understanding the separation is important because
each layer has a different owner and a different deployment path.

| Layer | What | Where stored | Where deployed | Owner |
|---|---|---|---|---|
| **Data** | The news items themselves (title, summary, url, date, topics) | `content_feed_items` table | `/data/*.json` in git | `render_news_section` action (content-feed pipeline) |
| **Behaviour** | JS that fetches the data and renders cards/list into the DOM | `content_components.js_content` | `/tools/assets/{function}.js` in git | `rerender_single_page` action (page-rerender) |
| **Structure + style** | The HTML shell (section, container, headings) and its CSS | `content_components.html_template` + `css_snippets` | inline in each page's HTML | `rerender_single_page` + `render_css_from_spec` |

The three layers connect at runtime in the browser:

```
page HTML (structure)
  └─ <script src="/tools/assets/latest-news.js"> (behaviour)
       └─ fetch("/data/latest-news.json") (data)
            └─ renders cards into the .news-grid container
```

### The two news components and their pairings

There are two distinct news components, each a separate `content_components`
row with its own function name, template, JS, and data file:

| Component function | Used on | Renders | JS file | Fetches |
|---|---|---|---|---|
| `latest-news` | index.html (homepage) | Card grid, curated recent few | `/tools/assets/latest-news.js` | `/data/latest-news.json` (top 6) |
| `news-listing` | news.html (news-index page) | Full vertical list, archive | `/tools/assets/news-listing.js` | `/data/news-archive.json` (20, source-interleaved) |

These are NOT duplicates. They are two different views of the news, with
different render logic (compact cards vs long-form list), different styling
(`.news-card` family vs `.news-list-item` family), and different data
breadth (curated few vs full archive).

### Designed for multiple views per site

This separation is what allows a site to host several news views, each
filtering or styling differently. To add a new view (e.g. an
`insurance-news` section showing only insurance-tagged items in a
magazine layout):

1. Create a new `content_components` row: `function = 'insurance-news'`,
   with its own `html_template` and `js_content`.
2. The `js_content` fetches whatever data file it should (e.g.
   `/data/insurance-news.json`).
3. Add CSS for the new class family via `css_snippets`.
4. Ensure something produces the data file (a filtered render query, or
   a new content-feed render target).

The pipeline auto-deploys `/tools/assets/insurance-news.js` whenever a
page using that component is rerendered. No workflow or code change is
needed — the deployment mechanism is generic over component function name.

index.html could then host `latest-news` + `insurance-news` +
`pricing-news` as three independent sections, each separately fed and
styled.

### How component JS gets deployed (the mechanism)

The `rerender_single_page` action (the `render_page` step of the
`page-rerender` workflow) does two things:

1. Assembles the page HTML from its `page_components` and the current
   site components (header/footer/head) + CSS.
2. Calls `collectJSAssets(pageID)` which queries:
   ```sql
   SELECT DISTINCT cc.function, cc.js_content
   FROM page_components pc
   JOIN content_components cc ON pc.component_id = cc.id
   WHERE pc.page_id = $1
     AND cc.js_content IS NOT NULL
     AND cc.js_content != ''
   ```
   For each row it adds `tools/assets/{function}.js → js_content` to a
   `files` map.

The action returns both `html` (the page) and `files` (a map of
`{path → content}` containing the HTML plus every component JS asset).

The `deploy_page` step (`git_commit` action) then commits the entire
`files` map in one commit. A page with a `latest-news` component produces
a 2-file commit: `/index.html` + `/tools/assets/latest-news.js`.

**Config requirement (this was the bug — see resolved issue below):** the
`deploy_page` step MUST use `files_field: "rendered_page.files"`, NOT
`content_field: "rendered_page.html"`. The latter commits only the HTML
and silently drops every JS asset.

---

## Resolved: component JS assets not deployed (files_field)

**Symptom:** index.html's latest-news section rendered its heading and
subheadline (static HTML) but showed no news cards. Browser console:
`Loading failed for the <script> with source ".../tools/assets/latest-news.js"`.
The JS file 404'd — it had never been committed to git, for any site, ever.

**Root cause:** The `rerender_single_page` action correctly computed the
`files` map including `tools/assets/{function}.js` entries. But the
`page-rerender` workflow's `deploy_page` step was configured with
`content_field: "rendered_page.html"`, which extracts only the HTML
string. The `git_commit` action's `extractFilesForGit` has three methods:

1. `files_field` → multi-file map (what we needed)
2. `files` → direct config files (legacy)
3. `content_field` → single file (what was configured)

With `content_field`, method 3 won and committed HTML only. The JS assets
in the `files` map were computed, returned, then discarded.

**Fix:** Updated the `page-rerender` agent_definition's `deploy_page`
config:

```sql
-- Before
"config": {
  "repo_name": "sites",
  "domain_field": "rendered_page.domain",
  "content_field": "rendered_page.html",      -- HTML only
  "commit_message": "Rerender: {{.filename}}",
  "filename_field": "rendered_page.filename"
}

-- After
"config": {
  "repo_name": "sites",
  "domain_field": "rendered_page.domain",
  "files_field": "rendered_page.files",        -- HTML + all component JS
  "commit_message": "Rerender: {{.filename}}"
}
```

Applied via `jsonb_set` on `default_config`. No code change — the action
already supported `files_field`; the workflow config was pointing at the
wrong extraction method.

**Why it took effect immediately:** `loadAgentDefinition` reads from the
DB on every spawn (no in-memory cache), and the query is
`ORDER BY version DESC`. The next page-rerender pod picked up the new
config.

**Verification:** Single-page rerenders of index and news both returned
`files_count: 2`:
- index → `["/index.html", "/tools/assets/latest-news.js"]`
- news  → `["/news.html", "/tools/assets/news-listing.js"]`

Both JS files landed in git; both data files (`latest-news.json`,
`news-archive.json`) already existed; the live pages now render news.

**Scope of the fix:** structural, not per-page. Every component on every
site with non-empty `js_content` now deploys its JS asset on rerender.
This unblocks the entire "co-located component JS" pattern, not just news.

---

## Known gaps (added to TODO)

### Multi-file commit message shows empty filename

`buildCommitMessage` only populates `{{.filename}}` when the commit has
exactly one file (`resolvedFilename` is set only for `len(filesMap) == 1`).
For 2-file news commits the message renders as `Rerender: ` (trailing
space, no name). Cosmetic. Fix: when `fileCount > 1`, build a message like
`Rerender: index.html (+1 asset)` using the primary HTML filename plus an
asset count.

### Data file paths are convention, not enforced

Each component's `js_content` hardcodes the data file it fetches
(`latest-news.js` → `/data/latest-news.json`). Nothing links the
component to its required data file declaratively, and nothing verifies
the file exists at deploy time. The `content_components.data_sources`
text[] column exists for this but is not consistently populated or
checked. Risk: a new component fetching `/data/insurance-news.json` will
silently render empty if no pipeline step produces that file. See
`component_asset_pipeline_concerns.md` for the fuller analysis and
proposed enforcement (declare in `data_sources`, validate at deploy, or
auto-stub missing files).

### External-script-reference pattern is fragile

A component template with `<script src="/tools/assets/X.js">` only works
if the matching `js_content` exists AND a page using the component gets
rerendered after the `files_field` fix. Before the fix, the reference
pointed at a non-existent file with no error surfaced anywhere in the
pipeline. Now the deployment works, but the pattern still depends on
remembering to populate `js_content`. Stronger long-term pattern: inline
small `js_content` (<5KB) directly into rendered HTML, reserve external
files for large payloads. See `component_asset_pipeline_concerns.md`.

---

# Component-to-asset pipeline — two architectural concerns

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

---

# rerender-pages workflow findings

Context: investigating how to cleanly rerender gaswholesalers.com after CSS
update, specifically ensuring news.html and index.html pick up the new
news component CSS classes.

## Finding 1: `rebuild_blog_listing` does NOT handle news-index pages

The `rebuild_blog_listing` step in the `rerender-pages` workflow v6 calls
`RebuildBlogListingAction`, which uses `findBlogPage` to locate the page
to rebuild. `findBlogPage` has two strategies:

```go
// Strategy 1: canonical page_type
SELECT id, name FROM pages
WHERE site_id = $1 AND page_type = 'blog-index'

// Strategy 2: page named 'blog' created as content type
SELECT id, name FROM pages
WHERE site_id = $1 AND name = 'blog' AND page_type = 'content'
```

Neither matches `page_type = 'news-index'` or `name = 'news'`.

**Consequence for sites with news (not blog):** the step is a silent no-op.
It logs `"No blog page found, skipping"` and returns. The news page is
still re-rendered as a regular `page_rerender` work item later in the
flow, so news visuals do get updated. But there is no equivalent
news-listing rebuild that would refresh the data shape if news feed
content changed.

**For future work:** the news-listing component renders client-side from
data files in `gaswholesalers.com/data/`, so a server-side rebuild may
not be needed. But if it ever is — e.g. to bake the latest items into
HTML for SEO — we'd need a parallel `rebuild_news_listing` action with
a `findNewsPage` helper that mirrors `findBlogPage` for news-index page
type and the `news-listing` component slot priority.

**Verification query:**
```sql
SELECT name, page_type FROM pages
WHERE site_id = '<site_id>'
  AND (name = 'news' OR name = 'blog'
       OR page_type IN ('blog-index', 'news-index'));
```

If results show `news-index` page type, this site is in the gap.

## Finding 2: JS snippets refresh is coupled to site components refresh

The workflow has three refresh phases tied behind one flag:

```
refresh_site_components=true →
    render_site_components →   (re-renders header, footer, head)
    render_js_snippets →        (builds snippets.js)
    deploy_js_snippets →        (commits snippets.js to git)
    rebuild_blog_listing → ...

refresh_site_components=false →
    rebuild_blog_listing → ... (skips all three of the above)
```

Conceptually these are independent:
- Site components are header/footer/head HTML stored in content_components
- JS snippets are a single concatenated file in the sites repo
- CSS is handled entirely by the webdesign-agent (not by rerender-pages)

A workflow consumer might want to refresh just JS (e.g. after editing
the js_snippets table) without touching site components. Currently they
have to pass `refresh_site_components=true` and accept the header/footer
re-rendering as a side effect.

**Future improvement (note for backlog):** split the single flag into
three:
```
refresh_site_components: bool
refresh_js_snippets:     bool
refresh_blog_listing:    bool   (currently always runs)
```

This is a workflow JSON edit in `agent_definitions` plus a small step
re-ordering. Low effort, modest value, do it next time we touch
rerender-pages versions.

## Finding 3: Migration C concern about rerender-pages was a false alarm

In an earlier session I noted that the rerender-pages workflow used
`input_data.site_id` directly rather than going through an
`ensure_site_record` step like other workflows. I'd flagged this as a
patch-needed item.

On closer inspection of v6:

| Step | site_id source | Issue? |
|---|---|---|
| check_refresh_components | conditional on refresh flag | none — doesn't need site_id |
| render_site_components | `input_fields: ["site_id", "domain"]` | extracted from input_data — fine |
| render_js_snippets | `site_id_field: "input_data.site_id"` | fine |
| deploy_js_snippets | `domain_field: "site_record.domain"` | needs site_record but only runs when refresh=true; verified working in earlier runs |
| rebuild_blog_listing | `site_id: "input_data.site_id"` | fine |
| get_pages | `input_fields: ["site_id", "domain"]` | fine |
| create_rerender_items | `site_id: "rerender_pages.site_id"` | derived from get_pages output |

The original patch I remembered was for a different workflow (the
single-page rerender invoked when content_components were modified by
update_site_content). That workflow had a real field-path bug that was
fixed. rerender-pages v6 is correctly set up for its current usage and
does not need this patch.

Note: this clarification matters because seeing "rerender-pages doesn't
use ensure_site_record" might trigger an instinct to patch it — don't,
the workflow works as designed.

## Trigger recommendation for clean full rerender

```json
{
  "action": "orchestrate",
  "config": {"agent_type": "rerender-pages"},
  "input_data": {
    "site_id": "<uuid>",
    "domain": "<domain>",
    "refresh_site_components": true
  }
}
```

Setting `refresh_site_components: true` produces the most complete
refresh:
- Header/footer/head re-rendered fresh from content_components
- snippets.js rebuilt and committed to git
- Each page re-assembled with current site components and current
  page_components
- One git commit per page

Set to `false` only if you specifically want to skip the site-level
refresh (e.g. you JUST refreshed it and don't want the second pass).

---

# The gaswholesalers news-visual fix plan (surgical + proper)

_The diagnostic entries that preceded this plan are in the debugging-guide addenda; what remains here is the concrete two-fix plan and its cross-site notes._

## Plan for gaswholesalers news visual

Two independent fixes needed. Each has a quick (surgical) path and a
proper path. Doing the surgical paths today makes the visual correct;
the proper paths fix the underlying gaps for future sites.

### Fix 1 — News CSS not in deployed styles.css

**Surgical (today):** Append the news css_snippet content to
`gaswholesalers.com/assets/css/styles.css` directly via git commit.
The CSS is already in the css_snippets rows; we just bypass the
renderer and write it.

```sql
-- Pull the news CSS out for git append
SELECT css_content
FROM css_snippets
WHERE name IN ('Latest News Grid', 'News Listing Page')
ORDER BY name;
```

Copy each row's css_content into a file, git commit, push:

```bash
cd ~/projects/sites
{
  echo ""
  echo "/* === News component styles (manually appended pending all_component_functions fix) === */"
  # paste Latest News Grid css_content
  # paste News Listing Page css_content
} >> gaswholesalers.com/assets/css/styles.css
git add gaswholesalers.com/assets/css/styles.css
git commit -m "news: append news section CSS manually (pending all_component_functions fix)"
git push
```

Survives until next webdesign-agent run on gaswholesalers, when styles.css
gets regenerated. By then ideally fix 1B has landed and the regeneration
includes news CSS itself.

**Proper (separate session):**

1. Find where `load_site_context` populates `site_context.all_component_functions`.
2. Verify it joins through `page_components` (not just `site_components`).
3. Run diagnostic query 4 from the debugging-guide entry above on a few
   recent webdesign-agent runs to see what's missing.
4. Fix the load_site_context query.
5. Re-run webdesign-agent on affected sites.

### Fix 2 — News IIFE still inline on deployed pages

**Surgical (today):** UPDATE `page_components.rendered_html` for the two
gaswholesalers pages (index for latest-news, news for news-listing) to
swap the inline IIFE for a `<script src>` tag, then page-rerender.

```sql
BEGIN;

-- gaswholesalers index page — latest-news component
UPDATE page_components pc
SET rendered_html =
      substring(pc.rendered_html FROM 1 FOR position('<script>' IN pc.rendered_html) - 1)
   || '<script src="/tools/assets/latest-news.js"></script>'
   || substring(pc.rendered_html FROM position('</script>' IN pc.rendered_html) + length('</script>')),
    updated_at = NOW()
FROM pages p
JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id = p.id
  AND p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'index'
  AND cc.function = 'latest-news'
  AND pc.rendered_html LIKE '%<script>%(function%'
  AND pc.rendered_html NOT LIKE '%<script src="/tools/assets/latest-news.js"></script>%';

-- gaswholesalers news page — news-listing component
UPDATE page_components pc
SET rendered_html =
      substring(pc.rendered_html FROM 1 FOR position('<script>' IN pc.rendered_html) - 1)
   || '<script src="/tools/assets/news-listing.js"></script>'
   || substring(pc.rendered_html FROM position('</script>' IN pc.rendered_html) + length('</script>')),
    updated_at = NOW()
FROM pages p
JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id = p.id
  AND p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'news'
  AND cc.function = 'news-listing'
  AND pc.rendered_html LIKE '%<script>%(function%'
  AND pc.rendered_html NOT LIKE '%<script src="/tools/assets/news-listing.js"></script>%';

-- Verify both rows updated
SELECT p.name AS page, cc.function,
       pc.rendered_html LIKE '%<script src="/tools/assets/%' AS has_script_src,
       pc.rendered_html LIKE '%<script>%(function%'           AS still_has_inline_iife
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND cc.function IN ('latest-news', 'news-listing')
ORDER BY p.name;

COMMIT;
```

Then trigger page-rerender for index and news (same kcat pattern as
before).

**Proper (separate session):**

Build a migration helper action that, when a `content_components.html_template`
column changes in a way that should propagate (script tag substitution,
class name changes, etc.), enqueues page-rerender items for every page
that uses that component. The migration framework should track which
columns are "deploy-affecting" and the system should self-heal.

### Order of operations for today

1. Run the diagnostic queries from both debugging-guide entries to
   confirm the diagnosis on gaswholesalers's actual data.
2. Apply Fix 1 surgical (CSS git append).
3. Apply Fix 2 surgical (page_components UPDATE).
4. Trigger page-rerender for `index` and `news` on gaswholesalers.
5. Open gaswholesalers.com — news cards should have the new visual
   design, dates expand to "2 days ago".

Total work: ~15 minutes. After that the news section is done for this
site. The two "proper" fixes (Fix 1B and Fix 2B) are work for separate
sessions — both useful, neither blocking other sites.

### What about other sites?

Other sites with `latest-news` or `news-listing` components have the same
two issues. The css_snippets fix means their next webdesign-agent run
WILL pick up the snippets — IF their `all_component_functions` includes
those function names. The contract-003 IIFE staleness applies to every
site whose news pages haven't been re-rendered by content-writer since
migration B (which is all of them).

If you want to roll the fix across sites, the same surgical approach
works site-by-site. A small loop over `sites WHERE id IN (...)` could
do the page_components UPDATE in bulk. The CSS append is per-site git
commit but scriptable.

Suggest letting the next natural webdesign-agent run cycle handle other
sites' CSS, and only batch-apply the page_components UPDATE if a specific
site is identified as needing it.

---

# The JS-snippet renderer deliverable + compliance review

_From `guidelines_compliance_check_1_.md`: the `render_js_snippets_for_site` action and `site-asset-renderer` agent that implement the missing JS-snippet deploy step (the gap named in the component-asset concerns above), walked against guideline docs 001/002/003._

Walking the deliverables through the three guideline documents
(001 development guide, 002 system architecture, 003 contracts).
This is for human review before applying — flags real concerns,
explains why apparent issues aren't problems.

## Inventory under review

| Deliverable | What it is |
|---|---|
| `render_js_snippets_for_site_action.go` | New Go action |
| Registry entry (one line) | Action registration |
| `migration_a_js_snippets_is_active.sql` | Schema migration |
| `migration_b_news_redesign_proper.sql` | Data + new agent definition |
| `migration_c_wire_snippet_renderer.sql` | Workflow wiring for 7 existing agents |

---

## 001 — Development Guide

### STEP ZERO: does this already exist?

**Searches performed before proposing the new code:**

| Search | Result |
|---|---|
| `grep -rn "js_snippet" /mnt/project/` | Only in three docs (002, 003, HANDOFF_2026-04-17). Zero Go code matches. |
| `grep -rn "loadJSSnippet\|injectJSSnippet\|selectJSSnippet"` | Zero matches. |
| `grep -rn "scroll-reveal\|smooth-scroll\|counter-animate"` (the snippet names) | Zero matches in any code file. |
| `grep -in "snippet" production_agent-chassis-full_context.txt` | Only matches were `loadComponentCSSSnippets` (the working CSS path) and web-search snippets (unrelated). |
| `agent_definitions` for similar agents | Closest matches: `webdesign-agent` (does CSS deploy with LLM), `css-patch-agent` (LLM patch). Neither renders JS snippets. |
| Action registry for similar handlers | `render_css_from_spec` is the closest sibling. No JS equivalent. |

Conclusion: the snippet loader is genuinely missing. New action and agent are justified.

### Reuse before creating

- `RenderJSSnippetsForSiteAction` deliberately mirrors `RenderCSSFromSpecAction`'s
  shape (load context → query snippets by `applies_to && components` → concatenate
  → return files map for git_commit). Same pattern, same conventions.
- Uses the existing `git_commit` action for deployment (not a new commit handler).
- Uses the existing `ensure_site_record` action to resolve site_id/domain in the
  agent's first workflow step (not a new resolver).

### Field path resolution: canonical helpers

Verified: the action only uses canonical `datahelpers` functions —
`ExtractNestedField`, `ExtractNestedFieldString`. No new path-resolver
function was added. The one helper that *is* new (`extractComponentFunctionsList`)
handles a different concern: type-switching a JSON value (which may be
`[]interface{}` or `[]string`) into a `[]string`. There's no canonical
helper for this in `datahelpers` — verified by grepping for `[]string`
return types in the package signatures. If a canonical helper for this
gets added later, the local helper is one-line to delete.

### Actions: don't split into wrapper + core

Verified: `render_js_snippets_for_site_action.go` exports only
`RenderJSSnippetsForSiteAction`. All helpers (`loadJSSnippetsForSite`,
`buildJSSnippetsBundle`, `extractComponentFunctionsList`,
`loadSiteComponentFunctionsForJS`, `emptyJSSnippetsBundle`) are private
(lowercase). No `RenderJSSnippets()` exported function for callers to use
directly.

### Workflows simple, complexity in Go

`site-asset-renderer` workflow has 4 steps:
`load_site → render_js_snippets → deploy_js_snippets → complete`.
No conditionals, no loops, no LLM. All complexity (snippet matching,
bundling, fallback to DB component query) is in the Go action.

### Spawn before call

Pattern B in migration C (webdesign-agent calls site-asset-renderer):
- `spawn_asset_renderer` step (action: `spawn_agent`) precedes
  `call_asset_renderer` step (action: `call_agent`).
- Uses `target_role` (not `agent_type`) for the call lookup — per guide,
  `target_role` is preferred because it scans all collected_data keys
  regardless of `output_field` naming.

### Agents own their domain

`site-asset-renderer`'s input contract is `{required: ["site_id"], optional: ["domain"]}`.
The first step (`ensure_site_record`) loads domain from the sites table if
absent. The agent doesn't need anything pre-computed by callers.

### Every pod-running agent needs a parent (wrapper-orchestrator pattern)

Question: does `site-asset-renderer` need a wrapper-orchestrator?

The guide's table classifies this as "Orchestrator invoked via the generic
entry point" → "runs briefly in-chassis; if it does more than trivial
coordination, it should spawn a child and delegate".

The site-asset-renderer workflow does:
- `ensure_site_record`: one DB query (~ms)
- `render_js_snippets_for_site`: one DB query + JSON marshal (~ms)
- `git_commit`: kafka send + adapter response (seconds, mostly async wait)

Total in-chassis CPU/memory hold: subsecond. The git_commit is async wait,
not CPU/memory pressure. This is "trivial coordination" by the guide's
classification — no wrapper needed.

When invoked from `webdesign-agent` (migration C pattern B), it gets a
dedicated Job pod via `spawn_agent` anyway. The "no wrapper" question
only applies to the manual-trigger path.

### Map fields individually, not as the whole input_data blob

Verified in migration C pattern B's `call_asset_renderer.input_mapping`:
```json
"input_mapping": {
  "site_id": "site_context.site_id",
  "domain":  "site_context.domain"
}
```
Not `"input_data": "input_data"`. Two named fields.

### Logging

All `params.Logger.Info(...)` — no `Debug` calls.

### Logger calls have keyed fields

```go
params.Logger.Info("RenderJSSnippetsForSiteAction: Complete",
    zap.String("site_id", siteIDStr),
    zap.String("domain", domain),
    zap.Int("component_count", len(components)),
    zap.Int("snippet_count", len(snippets)),
    zap.Strings("snippet_names", names),
    zap.Int("js_length", len(js)))
```
All structured. Matches the pattern in existing actions.

---

## 002 — System Architecture

### Site-level vs page-level assets

The architecture doc says (in "JavaScript Management"):
- `js_snippets` table: pre-built reusable JS for standard interactivity.
  Selected during planning, injected via head component. **Site-level,
  not per-page.**
- Custom JS for tools: self-contained, per-page.
- `content_components.js_content`: per-component JS.

The new pipeline matches this: snippets.js is **per site**, written to
`assets/js/snippets.js`. Per-component JS continues to live at
`/tools/assets/{function}.js`. Per-page tool JS isn't affected.

### Loading mechanism

The doc said snippets are "injected via head component". The previous
state had no mechanism. The new state has it: head template includes
`<script src="/assets/js/snippets.js"></script>` before `</head>`.

The doc said "selected during planning" — currently this happens at
render time (when the action runs) by matching `applies_to` against
the site's component functions. This is functionally equivalent: any
site whose components match the snippet's `applies_to` gets it. No
explicit planning step is needed. If a more deliberate "include snippet X
in this site's bundle" step is wanted later, it'd live in
`site-design-planner` or similar.

### Source of truth

- `css_snippets` table → `assets/css/styles.css` per site (via render_css_from_spec).
- `js_snippets` table → `assets/js/snippets.js` per site (via render_js_snippets_for_site).
- Same pattern. Both are per-site files derived from a global table.

---

## 003 — Contracts and Standards

### Component JS contract (the one we're respecting)

Doc 003 says:
- No inline `<script>` blocks in `html_template`.
- Component-specific JS lives in `content_components.js_content`,
  served at `/tools/assets/{function}.js`.
- Shared utilities (across many components) live in `js_snippets`,
  loaded via head's snippet-loading mechanism.

Migration B implements all three for the news components:
1. Inline `<script>(IIFE)</script>` extracted from `html_template`
   into `js_content` (contract 003 split).
2. `html_template` now has `<script src="/tools/assets/{function}.js"></script>`.
3. `formatNewsDate` placed in `js_snippets` (shared between latest-news
   and news-listing).

### Component CSS scoping

The new CSS in migration B sections 1 and 2 is scoped:
- `.latest-news-section .*` for the homepage component.
- `.news-listing-section .*` for the listing page.
- Shared element styles (`.news-card`, `.news-list-item`) are tied to
  parent class names.
- No bare element rules (`h2 { ... }`).

### Dark section variables

The new CSS uses `var(--color-*, fallback)` throughout. Fallbacks are
included so the components render reasonably even on sites without the
theme variables defined.

### Idempotency

All three migrations are re-runnable:
- A: `ADD COLUMN IF NOT EXISTS`.
- B: every UPDATE has a `NOT LIKE` or similar guard; INSERT has `ON CONFLICT`.
- C: each DO block checks for the new step's presence before modifying.

---

## Genuine concerns to flag

### 1. Image version on the new agent definition

The `site-asset-renderer` row in migration B has `version = 'v1.0.1012'`.
This is the previous chassis tag. After the new Go action is added and
the chassis is rebuilt, the new tag (say `v1.0.1013` or similar) should
be set on this row. Easy to update once the build is done.

### 2. The new agent's `applies_to` field

I set `["website"]` to match the pattern used by other site-touching
agents. Verify this matches your routing logic — some agents use this
for filtering.

### 3. Action registry entry — one-line code change

The registry entry in `registry_go.txt` around line 704 needs an
addition. This is NOT in a migration; it's a Go code change. See
`deployment_news_proper.md`. Failing to add it means the workflow
will error at the `render_js_snippets_for_site` step with "unknown
action".

### 4. Migration C touches 7 existing agent_definitions rows

If anything else is currently mid-flight on these agents (a running
workflow execution), the workflow modification doesn't affect in-flight
runs (they keep their loaded config) but applies to subsequent runs.
Worth pausing the scheduler briefly if you have running orchestrations.

### 5. The 6 dormant js_snippets rows stay dormant

The 9 existing snippets stay `is_active = false`. They won't load on
any site after this migration. If you want any to activate (e.g.
`smooth-scroll` globally), just `UPDATE js_snippets SET is_active = true
WHERE name = '<name>'` and run site-asset-renderer per site.

---

## Test plan after applying

| Step | Expected |
|---|---|
| Migration A applied | js_snippets has `is_active` column, 9 rows = false |
| Migration B applied | News content_components have `<script src>` and js_content; head templates have snippets.js loader; site-asset-renderer agent exists |
| Migration C applied | 7 workflows show the new steps in verification SELECTs |
| Chassis rebuilt with new action + registry entry | Pod images bump; rolling restart |
| Trigger site-asset-renderer on gaswholesalers | git commit appears for `assets/js/snippets.js` containing formatNewsDate |
| Direct UPDATE site_components.rendered_html for head slot | Stored head HTML now has the script tag |
| page-rerender on index + news | Pages now reference /tools/assets/latest-news.js and /tools/assets/news-listing.js; collectJSAssets ships those files |
| Browser load gaswholesalers index | `snippets.js` loaded first; news component IIFE runs; `formatNewsDate` resolves dates |

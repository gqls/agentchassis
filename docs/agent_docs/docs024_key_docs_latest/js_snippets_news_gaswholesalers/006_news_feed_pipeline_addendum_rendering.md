# Addendum to 006 — News Rendering & Asset Deployment Architecture

Insert this into `006_news_feed_pipeline_v2.md`. Suggested placement: replace
the brief "Content writer" subsection (currently ~line 325-327) with the
"Rendering layer" section below, and add the "Resolved: component JS assets
not deployed" entry to the resolved-issues area.

Discovered and fixed 2026-05-19/20 during gaswholesalers.com news rollout.

---

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

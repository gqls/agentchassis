# CSS and JS Mechanisms — Focus Document

Compiled 2026-05-12 from a deep read of the codebase and live database state, prompted by the news section redesign task on gaswholesalers.com. Distinguishes what is built and working, what is partially built, and what is declared in the contracts but not implemented.

The point of this document is to make the next decision deliberate: where should new styles and new shared utility JS live, given what actually exists today?

---

## 1. CSS — fully built path

### 1.1 Storage layout

| Table | Purpose | Column of interest |
|---|---|---|
| `css_themes` | Theme-wide CSS templates (Go templates with palette/typography/layout substitutions) | `css_content`, `css_template`, `palette_id`, `layout_id`, `typography_set_id` |
| `palettes`, `layouts`, `typography_sets` | Decomposed theme pieces; FK'd into `css_themes` | n/a |
| `css_snippets` | Section-level CSS keyed by `applies_to` | `css_content`, `applies_to` (jsonb), `semantic_tags` |
| `content_components.html_template` | Section-level CSS embedded inline as `<style>` inside the component's HTML | (legacy pattern for components like `hero`, `features`, `services`) |

### 1.2 css_snippets schema (verified live)

```
id            uuid
name          varchar(100) UNIQUE
description   text
css_content   text NOT NULL
semantic_tags jsonb default '[]'
applies_to    jsonb default '[]'
created_at    timestamp
```

Current rows of interest:

| name | applies_to | css_len |
|---|---|---|
| `input-modern` | `["form","input","newsletter"]` | 393 |
| `Latest News Grid` | `["latest-news"]` | 2061 |
| `News Listing Page` | `["news-listing"]` | 2337 |

`applies_to` is matched against the site's component function list (`site_context.all_component_functions`) using the JSONB overlap operator `&&`. The row's `name` is a display name; `applies_to` is the join key.

### 1.3 Assembly pipeline

The end-to-end flow producing `assets/css/styles.css`:

```
webdesign-agent (orchestrator)
  ↓
analyze_design  (LLM) → produces design_spec
  ↓
update_site     stores design_spec on the site
  ↓
render_css_from_spec   (the assembly step)
  │
  ├─ loadThemeComposition  → palette + layout + typography rows via css_themes FKs
  ├─ extractCSSComponents  → site_context.all_component_functions
  ├─ queryDarkSectionsForCSS → which sections need dark variants
  ├─ buildCSSsectionStyles → per-section style entries
  ├─ loadComponentCSSSnippets → SELECT * FROM css_snippets WHERE applies_to && $1::jsonb
  └─ renders layout's css_template with palette/typography/token helpers, then concatenates everything
  ↓
deploy_css      git_commit to assets/css/styles.css on the site's git repo
  ↓
b2 sync         GitHub Action mirrors the file to the CDN
```

`render_css_from_spec_action.go` is the file. It is deterministic — no LLM. The LLM step is `analyze_design`, one earlier.

### 1.4 Alternative CSS deployment path

`css-patch-agent` exists for targeted fixes. Its workflow:

```
plan_css_fix (LLM, generates a patch from audit finding + current CSS)
  ↓
deploy_css   (git_commit to assets/css/styles.css)
```

This bypasses `css_snippets` entirely — it patches the deployed file directly. Useful for one-off fixes (the "audit finding" pattern: a single CSS bug surfaces, patch it). NOT the right home for section-design changes that should apply to every site using that section.

### 1.5 What this means for the news redesign

The two existing css_snippets rows (`Latest News Grid` and `News Listing Page`) are the canonical homes. Steps:

1. UPDATE the `css_content` column on both rows with the new CSS.
2. For propagation, two options:
   - **Wait** for the next webdesign-agent run on any affected site. Eventually consistent.
   - **One-off git commit** of the new CSS appended to `gaswholesalers.com/assets/css/styles.css`. Cheap, takes effect immediately. Will be overwritten on next webdesign run, but by then the snippet row contains the same content so the overwrite is a no-op visually.
   - Triggering webdesign-agent explicitly is not recommended — `analyze_design` re-runs the LLM and can shift the palette/typography unrelatedly to this change.

---

## 2. JS — three different paths

The system has three JS paths, not one. They serve different purposes and have different levels of completeness.

### 2.1 Path A — Component-specific JS via `content_components.js_content`

**Status: fully built and in use.**

The canonical home for per-component behaviour: the hero's IntersectionObserver, a tool component's calculator logic, an interactive component's event handlers.

| Table.column | Purpose |
|---|---|
| `content_components.js_content` | Raw JS body (no `<script>` tags) |
| `content_components.html_template` | Includes `<script src="/tools/assets/{function}.js"></script>` after `</section>` |

The flow:

```
LLM generates component (may have inline <script> blocks)
  ↓
store_generated_component_action.go calls separateInlineJS()
  → extracts <script>…</script> bodies (only the kind with no attributes)
  → puts them in js_content
  → inserts <script src="/tools/assets/{function}.js"></script> after </section>
  ↓
Stored in content_components
  ↓
page-content-writer renders html_template → rendered_html (has the <script src> tag)
  ↓
page-rerender:
  RerenderSinglePageAction → collectJSAssets()
  → SELECT cc.function, cc.js_content FROM page_components pc JOIN content_components cc
    WHERE pc.page_id = ? AND cc.js_content IS NOT NULL AND cc.js_content != ''
  → returns files map: { "page.html": html, "tools/assets/{function}.js": js, … }
  ↓
git_commit writes all files in a single multi-file commit
```

Conventions:
- Asset path is `/tools/assets/{function}.js` — `function` from content_components, not section_type.
- One JS file per component. Functions are globally unique among active components.
- Validated by `separateInlineJS()` regex; safe against `<script src="…">` (won't extract those).

### 2.2 Path B — Shared JS utilities via `js_snippets` table

**Status: declared in contracts, table populated, BUT NO LOADER IS WIRED UP.**

This is the one to be careful about. The contracts doc and architecture doc both describe a path where small reusable behaviours (scroll reveals, smooth scroll, lazy-load images, date formatters) live in a shared table and get injected onto pages that need them. The intended split:

| Path | Storage | Scoping |
|---|---|---|
| A (component-specific) | `content_components.js_content` | 1:1 with a component |
| B (shared utilities) | `js_snippets` table | Many components via `applies_to` |

The `js_snippets` schema (verified live):

```
id            uuid
name          varchar(100) UNIQUE
description   text
js_content    text NOT NULL
semantic_tags jsonb default '[]'
applies_to    jsonb default '[]'
dependencies  jsonb default '[]'
created_at    timestamp
```

The table has 9 rows in production right now, including `scroll-reveal`, `smooth-scroll`, `counter-animate`, `typing-effect`, `lazy-load-images`. The `applies_to` shape uses generic section/element types (`["section","card","feature"]`, `["navigation","link"]`, `["stats","numbers","social-proof"]`).

**The loader does not exist.** Verified by:

1. Both head components in DB (`head-seo-standard` and `Document Head`) have **no** reference to `js_snippet`, `js_snippets`, or `{{range .js_snippets}}` in their template.
2. `RenderHead` in `component_library.go` calls `GetComponentByFunction(db, "head")` and renders the template via `RenderTemplate` — there is no js_snippets query or injection in that path.
3. None of the deployed pages on gaswholesalers contain script tags or fetches pointing at any js_snippet content.

So a row in `js_snippets` today is dead weight — it sits there but nothing references it. The "loaded via head component's snippet-loading mechanism" claim in contracts/architecture is **aspirational**, not implemented.

Two ways forward if Path B is the goal:

- **Build the loader.** Either (a) add a `loadJSSnippetsForSite()` helper alongside `loadComponentCSSSnippets()` and have `RenderHead` call it, with the result passed to the head template as `{{.js_snippets}}` to iterate; or (b) extract the snippet content into a single concatenated `/assets/js/snippets.js` file and have the head template emit a single `<script src>` tag for it. (a) is more flexible; (b) is closer to how CSS works today.
- **Don't build it yet, but reserve the namespace.** Insert future-intended rows into `js_snippets` so we can come back to it without inventing names later. Inline the actual JS where it's needed today (Path A or inline IIFE) with a comment pointing at the snippet row.

### 2.3 Path C — Inline `<script>` in rendered_html (legacy, anti-pattern)

**Status: works, but explicitly called out in contracts as something to migrate AWAY from.**

The current news component is here. The rendered_html for the `latest-news` page_component on gaswholesalers' index page is:

```html
<section data-component="latest-news" …>
  …
</section>
<script>
(function() {
  fetch("/data/latest-news.json")…
})();
</script>
```

That inline `<script>` violates contract 003 ("Do NOT put inline `<script>` blocks in `html_template`"). The `separateInlineJS()` step is supposed to extract these — but only at component-creation time. Once the component exists with inline scripts in its `html_template` row, every render carries them through into `rendered_html`.

This is what the HANDOFF doc from 2026-04-17 was about: separating inline JS from existing components after the fact. Some have been migrated (the contract lists vonc.com's `provocation-feed` and `archetype-combinations` as having `<script src>` tags vs the older `provocation-card`, `lobby-grid`, `archetype-grid` still with inline JS). The news components haven't been migrated yet.

### 2.4 Path D — html-assembler agent's `inject_js` flag (existence unclear)

The `html-assembler` agent (used by `landing-page-builder`, not the day-to-day rerender flow) has a config flag:

```json
{
  "action": "assemble_full_page",
  "config": {
    "inject_js": true,
    "inject_css": true,
    …
  }
}
```

But the `assemble_full_page` action's Go code visible in source only implements `inject_head`, `inject_header`, `inject_footer`. The `inject_js` flag is defined but I couldn't find a code path that reads it. Either it's implemented in a file outside what's been indexed, or it's a leftover config from a planned feature.

Not relevant to gaswholesalers (which uses the page-rerender path, not html-assembler), but worth knowing about — don't rely on it for shared JS until verified.

---

## 3. Decision matrix for new JS

For any new piece of JS, the path to choose is:

| Type of JS | Right path | State |
|---|---|---|
| Logic that's part of one specific component, tightly coupled to its DOM | A (`content_components.js_content`) | Built |
| Small utility used by multiple components (e.g. relative-time formatter) | B (`js_snippets`) ideally — but loader needs building | Loader missing |
| Browser-wide effect that runs once and modifies the page (scroll reveals, lazy-load) | B (`js_snippets`) ideally — same problem | Loader missing |
| Tool-specific embedded interactivity (calculators, games) | A (`content_components.js_content` with `function` matching the tool) | Built |

For Path B uses today, two pragmatic choices until the loader is built:

- **Promote to Path A** by hosting the utility inside the JS of the component that needs it. Loses generality but ships now.
- **Pre-emptively insert into js_snippets and reference inline**: insert the row with the right `applies_to`, AND duplicate the function body inline at every call site for now. When the loader lands, the inline duplicates can be removed in one pass.

---

## 4. Specific implications for the news redesign

Concrete decisions for the work in flight:

### 4.1 CSS

Path forward is unambiguous:

```sql
UPDATE css_snippets
SET css_content = $$<new CSS from news-section.css>$$
WHERE name = 'Latest News Grid';

UPDATE css_snippets
SET css_content = $$<news-listing portion of news-section.css>$$
WHERE name = 'News Listing Page';
```

The new CSS I drafted earlier is already structured to fit — it scopes everything to `.latest-news-section` / `.news-listing-section` and uses `var(--color-*)` custom properties that match the global theme.

For deployment to gaswholesalers, the pragmatic option:

```bash
# One-off: append the new news block to gaswholesalers.com/assets/css/styles.css
# This appears at the end of the cascade, so it overrides whatever is currently
# there. Survives until the next webdesign-agent run, by which time the
# css_snippets row will produce identical CSS so the overwrite is invisible.
```

No agent triggers required, no LLM runs, no risk of palette drift.

### 4.2 JS

The `formatNewsDate` function is a textbook Path B candidate — it's a presentation-layer utility used by both `latest-news` and `news-listing`, and could be reused by anything else that displays relative dates in future.

Recommended approach, in priority order:

1. **Insert a js_snippets row for the function**:
   ```sql
   INSERT INTO js_snippets (name, description, js_content, applies_to, semantic_tags)
   VALUES (
     'news-date-formatter',
     'Expands abbreviated relative-time strings ("2d ago" → "2 days ago") in news listings.',
     $$function formatNewsDate(s) { … }$$,
     '["latest-news","news-listing"]',
     '["news","presentation","time"]'
   );
   ```
   This claims the name and applies_to even though nothing loads it yet.

2. **Define the same function inline inside the news IIFE for now.** Tactical duplication — when the loader lands, the inline copy gets removed.

3. **Update the news IIFE call sites** in two places:
   - `content_components.html_template` for `latest-news` and `news-listing` (canonical, picked up by future rebuilds)
   - `page_components.rendered_html` for gaswholesalers' index page and news.html (immediate effect, picked up by next rerender)

   Both change `item.date` to `formatNewsDate(item.date)`.

4. **Track the loader work** as a separate follow-up — when there's appetite to build it, `RenderHead` gets a `loadJSSnippetsForSite()` call and the head template gets a `{{range .js_snippets}}<script>{{.js_content}}</script>{{end}}` block, mirroring how CSS already works.

This satisfies the user's "any solutions we do now shouldn't obstruct the longer term design" — the snippet row is in place when the infrastructure catches up.

### 4.3 Larger contract violation to flag

The news component's inline `<script>` in `html_template` is contract 003 non-compliant. Properly fixing this means:

- Run something analogous to `separateInlineJS()` on the existing `latest-news` and `news-listing` content_components rows
- The IIFE moves to `js_content`
- The template gets `<script src="/tools/assets/latest-news.js"></script>` instead
- `collectJSAssets()` picks it up on next rerender, the JS deploys as a separate file
- `formatNewsDate` could live as a regular function inside that JS file (private to the component) OR be loaded via js_snippets (shared) once the loader exists

This is a bigger refactor and not strictly needed for the immediate redesign. Worth tracking as the "fix the news component to comply with contract 003" task.

---

## 5. Reference — files and rows touched

For the next person debugging or extending this:

| What | Where |
|---|---|
| CSS snippet rows for news | `css_snippets` rows `Latest News Grid`, `News Listing Page` |
| CSS theme rendering | `platform/orchestration/actions/render_css_from_spec_action.go` + `render_css_composition_helpers.go` |
| webdesign-agent workflow | agent definition in `agents` table, type `webdesign-agent` |
| css-patch-agent workflow | agent definition in `agents` table, type `css-patch-agent` |
| Per-component JS | `content_components.js_content` |
| Inline-JS extraction | `platform/orchestration/actions/store_generated_component_action.go` → `separateInlineJS()` |
| Per-page JS asset collection | `platform/orchestration/actions/rerender_single_page_action.go` → `collectJSAssets()` |
| Shared JS snippet library | `js_snippets` table (9 rows, no loader) |
| Head component templates | `content_components` rows `head-seo-standard` and `Document Head` (neither references js_snippets) |
| Head rendering | `component_library.go` → `RenderHead()` |
| News pipeline overview | `006_news_feed_pipeline_v2.md` |
| Component JS contract | `003_contracts_and_standards_v8.md` § JS Content Separation |
| JS separation handoff | `HANDOFF_2026-04-17_component_rendering_js_separation_quality.md` |

---

## 6. One-line summary

CSS has a coherent end-to-end pipeline (css_snippets → render_css_from_spec → styles.css). JS has two coherent pipelines (per-component js_content; the legacy inline IIFE which works but violates contract) and one half-built one (js_snippets as a shared library with no loader). Until the js_snippets loader is built, treat that table as a registry of intentions rather than runtime infrastructure.

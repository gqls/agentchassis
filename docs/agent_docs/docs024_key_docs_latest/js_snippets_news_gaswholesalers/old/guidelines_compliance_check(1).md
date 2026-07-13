# Guidelines compliance review

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

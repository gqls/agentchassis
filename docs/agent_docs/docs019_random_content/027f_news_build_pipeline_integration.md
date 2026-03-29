# 026f — News Section Build Pipeline Integration

How the latest-news section integrates with the site build pipeline. This covers all agents involved, what each needs to know, and the minimal changes required.

---

## Design Principle

The news section flows through the standard build pipeline like any other section. It is NOT a special case with custom plumbing. The only difference: its content comes from a database query (`content_feed_items`) rather than LLM generation. Periodic refresh replaces the one-time write.

```
Standard section lifecycle:
  classifier → planner → webdesign → content writer → render → deploy

News section lifecycle (same pipeline):
  classifier → evaluate_news_feed → planner → webdesign → content writer → render → deploy
  + periodic: render_news_section fills items from DB every 6 hours
```

---

## Agent-by-Agent Changes

### 1. Classifier enrichment (`evaluate_news_feed`)

**Status:** Built. No changes needed.

Writes `content_features.news_feed` to classification spec:
```json
{
    "content_features": {
        "news_feed": {
            "recommended": true,
            "reason": "Energy markets have daily price movements...",
            "vertical_keywords": ["wholesale gas prices", "energy market"],
            "source_types": ["rss", "news_search", "api_news"]
        }
    }
}
```

This is available to every downstream agent that reads the classification spec.

### 2. Planner

**Status:** Needs prompt addition.

The planner reads `classification.content_features.news_feed.recommended`. If true, it includes `latest-news` in the homepage section list.

**Prompt addition** (add to the planner's system prompt or section guidance):

```
## News Feed Section

If the classification spec includes content_features.news_feed.recommended = true,
include a "latest-news" section on the homepage.

Placement: after the main content sections (services, social-proof, about),
before the final call-to-action. Typically position 4-6 on a 6-7 section homepage.

The latest-news component is data-driven — its items are populated from the
news feed database, not by the content writer. The content writer will generate
only the headline (e.g. "Latest Energy Market News") and optionally a subheadline.

Section spec:
{
    "component": "latest-news",
    "headline_guidance": "Use the site's vertical and audience to write an engaging headline",
    "render_mode": "data-driven"
}

Do NOT include latest-news if the classification spec does not recommend it.
```

This is a text addition to the planner agent definition's prompt template. No Go code changes.

### 3. Webdesign Agent

**Status:** Mostly automatic. Needs CSS snippet (026e SQL) + small Go patch.

**How it already works:** `LoadSiteForDesignAction` scans all page_components and builds `all_component_functions`. If `latest-news` is in the list, it appears in the component list passed to the CSS generation step.

**What needs adding:**
1. Run `026e_latest_news_css.sql` — creates a `css_snippet` for the news grid layout
2. Apply the `render_css_snippets_patch.go` — makes `render_css_from_spec` automatically append CSS from `css_snippets` for any component in the site's list

After this patch, the webdesign agent needs no further per-component changes. Adding a new component with custom CSS? Insert a `css_snippets` row with `category = 'component'` and it's automatically included.

**The `analyze_design` LLM step** receives the component list. It can make layout decisions about the news section (sidebar vs full-width, card size, emphasis). But the default CSS from the snippet is sufficient — responsive card grid, 3 columns on desktop, 1 on mobile.

**Future enhancement:** The `analyze_design` prompt could include news layout as a decision: "The site has a latest-news section. Should it be: (a) full-width card grid, (b) sidebar alongside content, (c) compact list? Consider the site's design style and the news frequency." The output would set CSS variables that the snippet responds to. Not needed now.

### 4. Content Writer

**Status:** No changes needed.

When the content writer encounters `latest-news` in the page sections, it:
1. Loads the component template from `content_components`
2. Renders it with its generated content data

Since `news_items` will be empty at build time, the template renders the placeholder:
```html
<p class="news-empty">News updates coming soon.</p>
```

The content writer generates the headline ("Latest Energy Market News") and optionally a subheadline. This is normal LLM content generation — no special handling.

After the build completes, `render_news_section` overwrites the page_component with actual items from the database. The placeholder is only visible briefly (or not at all if feed ingestion runs before the first user visits).

### 5. `render_news_section` (post-build + periodic)

**Status:** Built. No changes needed.

Uses the position already set by the planner/content-writer. Doesn't need to figure out where to insert — it finds the existing `page_component` row by matching `component_id` + `function = 'latest-news'` and updates the HTML.

For the initial build flow, this runs as part of the `content-feed-orchestrator` cycle. For the first cycle, it fills items from any existing `content_feed_items`. For subsequent cycles, it updates with the latest relevant items.

### 6. `plan_sections` Action

**Status:** No changes needed.

The `plan_sections` source resolver already handles `query.*` sources:

```go
case "query":
    // Query sources are resolved at render time, not at planning time
    return nil, true
```

The `latest-news` component's input_schema has `"source": "query.content_feed_items"` for the items field. `plan_sections` sees this and marks the section as "ready" (data source available). The section passes the readiness check and proceeds to the content writer.

### 7. Discovery Checks (existing sites)

**Status:** Built. No changes needed.

For sites that were built before news was added:
- `missing_news_sources` → detects spec says news but no sources configured
- `missing_news_section` → detects items exist but no component on homepage
- `stale_news_section` → detects component exists but items are old
- `all_sources_erroring` → detects all sources have errors

The improvement loop creates work items. The handlers add the section, configure sources, or trigger re-renders as needed.

---

## Summary of Changes Required

| Component | Change | Type | Size |
|-----------|--------|------|------|
| Planner agent definition | Add news section guidance to prompt template | SQL (prompt text) | Small |
| CSS snippets | `026e_latest_news_css.sql` — news grid CSS | SQL | Small |
| `render_css_from_spec` | Append component CSS snippets automatically | Go patch (~15 lines) | Small |
| Nothing else | Classifier, content writer, plan_sections, webdesign LLM step, render pipeline all work as-is | — | — |

**Total: 1 SQL file + 1 small Go patch + 1 prompt update.**

Everything else is already built or works without changes.

---

## Build-Time Flow (New Site)

```
1. intake-orchestrator receives domain
2. domain-research-classifier runs research + classification
3. evaluate_news_feed reads classification spec
   → writes content_features.news_feed.recommended: true (for energy vertical)
4. planner reads classification + design_intent
   → sees news_feed.recommended: true
   → includes latest-news at position 5 (before CTA) in homepage site_plan
5. webdesign agent runs:
   → LoadSiteForDesignAction finds latest-news in component list
   → analyze_design picks colors, fonts (knows about news section)
   → render_css_from_spec renders theme CSS
   → loadComponentCSSSnippets appends .news-grid/.news-card CSS
   → CSS deployed to site
6. content writer processes homepage sections:
   → hero: LLM generates content
   → services: LLM generates content
   → social-proof: LLM generates content
   → latest-news: LLM generates headline, template shows "News updates coming soon"
   → call-to-action: LLM generates content
7. page assembled and deployed
8. content-feed-orchestrator runs (same cycle or next):
   → dispatches feed-ingesters (fills content_feed_items)
   → render_news_section overwrites latest-news with actual items
9. feed-triage runs:
   → scores items against site spec
   → marks relevant / rejected
10. render_news_section runs again (next cycle):
    → now only shows status='relevant' items
```

---

## Discovery Flow (Existing Site)

```
1. evaluate_news_feed runs for existing site
   → enriches classification spec with news_feed.recommended: true
2. improvement loop runs:
   → missing_news_sources check: "spec says news but no sources"
     → work item created → configure-feed-sources (future: auto)
                           currently: manual SQL to add sources
3. sources added, feed-ingesters run, items ingested
4. improvement loop runs again:
   → missing_news_section check: "items exist but no latest-news component"
     → work item created → handler adds latest-news page_component to homepage
5. webdesign agent re-runs (triggered or next audit):
   → sees latest-news in component list
   → CSS updated to include news styles
6. render_news_section fills actual items
7. stale_news_section check monitors freshness ongoing
```

---

## Why Not Change the Planner's Code?

The planner decides section composition via LLM prompt. Adding "include latest-news when spec says to" is a prompt-level decision, not a code-level one. The planner's Go code (section resolution, component matching) doesn't need changes because `latest-news` is a standard component in `content_components`.

This follows the system's design: **planner decides intent (via LLM), plan_sections validates feasibility (via Go), content writer implements (via LLM + template)**. News fits this flow without special cases.

## Why Not Special-Case the Content Writer?

The content writer renders whatever sections it's given. For `latest-news`, it renders the template with the generated headline and empty items. The template's `{{if not .news_items}}` clause handles the empty state gracefully. No "skip this section" logic needed.

`render_news_section` then overwrites with real data. This two-render approach is simpler than adding content-writer awareness of "data-driven" sections. It also means the initial deploy has a valid page (with placeholder text) rather than a broken layout with a missing section.

## Why CSS Snippets Instead of Template Editing?

The `css_themes.css_content` Go template is a large, carefully structured document. Editing it for each new component is fragile and creates merge conflicts. The `css_snippets` approach is additive:

- New component needs CSS? → `INSERT INTO css_snippets`
- `render_css_from_spec` queries matching snippets and appends them
- No template editing, no recompilation, no risk to existing CSS

This also means non-developers can add component CSS via SQL, and the webdesign agent's existing design decisions (colours, fonts) are applied via CSS variables in the snippets.

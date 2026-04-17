# Handoff: Nav Tiers, Empty Sections, Component Validation, Footer Quick Links

**Date:** 2026-04-17  
**Continues from:** `103_blog_nav_handoff-2026-04-12.md`  
**Transcript:** `/mnt/transcripts/2026-04-15-16-08-38-blog-nav-handoff-debugging-session.txt`

---

## What Was Deployed This Session

### 1. Empty Section Filter — `rerender_single_page_action.go`

**Problem:** Pages rendered with empty sections — hero with `<h1></h1>`, CTA with no heading, FAQ with empty `<summary>` tags. These produced blank space on the live site.

**Fix:** `getPageSections` now calls `sectionHasVisibleContent()` on each section before including it. Strips `<style>`, `<script>`, HTML tags, and entities, then checks if >10 chars of text remain. Sections that fail are skipped and logged.

**Signature change:** `getPageSections` now takes `logger *zap.Logger` parameter. The call site in `assemblePage` passes `logger`.

**Precompiled regexps:** `reStyleBlocks`, `reScriptBlocks`, `reHTMLTags`, `reHTMLEntities`, `reWhitespace` — package-level vars.

### 2. Tiered Nav Priority — `populate_nav_tables_action.go`

**Problem:** `classifyPagesForNav` used a binary `in_header → primary` approach. With 21 `in_header=true` pages and `maxHeaderItems=8`, the truncation was arbitrary (by nav_order). Blog was excluded, password-entropy tool was included.

**Fix:** Three-tier priority system:

| Tier | Pages | Rationale |
|------|-------|-----------|
| 1 | index, services, about, contact | Core pages every site needs |
| 2 | blog, case-studies, use-cases, pricing, how-we-work, portfolio, products, solutions, industries + page_type blog-index/entity-directory | Content hubs and conversion pages |
| 3 | Everything else (faq, approach, careers, guides, insights) | Secondary pages |

Pages sorted by tier ascending, then `nav_order` ascending within tier. Overflow goes to utility (footer).

**New function:** `navPriorityTier(nameLower, pageType string) int`

**Struct change:** `pageNavInfo` now includes `PageType string`. The `loadPagesForNav` query selects `COALESCE(page_type, 'content')`.

**Import added:** `"sort"`

### 3. Child Page URL Exclusion — `populate_nav_tables_action.go`

**Problem:** Individual tool pages (`/tools/tool-ai-readiness-quiz.html`) appeared in utility nav. Their `page_type` was `content` not `tool`, so `neverPrimaryTypes` didn't catch them.

**Fix:** `isChildPageURL` function checks URL prefixes: `/tools/`, `/blog/`, `/guides/`, `/articles/`, `/case-studies/`, `/news/`, `/resources/`, `/insights/`. Child pages are skipped entirely from all nav groups.

**Note:** Confirmed that the nav_drift rebuild at 14:12:57 ran BEFORE this code was deployed. A fresh `nav_drift_v2:` item was created to trigger a rebuild with the new code. Verify tool pages are excluded after it completes.

### 4. Nav Label Fixes — `populate_nav_tables_action.go` + `component_library.go`

**Problem:** "AI For Your Type Of Business" stored in nav → rendered as "Ai For" (2 words). Two-stage truncation: `navSimplifyLabel` fell back to URL-derived label ignoring `page.NavLabel`, then `simplifyNavLabelForRender` truncated to first 2 words.

**Fix (populate_nav_tables):** `navLabelForPage` now trusts `page.NavLabel` directly when ≤30 chars. Only simplifies with `navSimplifyLabel` if nav_label is unreasonably long or missing.

**Fix (component_library):** `simplifyNavLabelForRender` no longer truncates to 2 words. Only strips brand suffixes (`|`, ` - `) and maps index/home to "Home". Stored labels are trusted.

### 5. Footer Quick Links — `render_site_components_action.go` + `component_library.go`

**Problem:** Footer "Quick Links" used `{{.nav_items_html}}` which was built from primary nav only. Utility items (FAQ, Approach, etc.) didn't appear in footer.

**Fix (render_site_components):** 
- New `quickLinksItems` loaded from primary + utility groups (no legal)
- New `quickLinksHTML` built via `buildNavItemsHTML(quickLinksItems)`
- `FooterNavItems` set on `RenderContext` struct
- `"quick_links_html"` added to ContentData map

**Fix (component_library):** Both `RenderTemplate` functions now substitute `{{.quick_links_html}}` / `{{quick_links_html}}` using `ctx.FooterNavItems`.

**Footer template updated (SQL):**
```sql
UPDATE content_components
SET html_template = REPLACE(html_template, '{{.nav_items_html}}', '{{.quick_links_html}}')
WHERE function IN ('footer', 'site-footer') AND is_active = true
  AND html_template LIKE '%Quick Links%{{.nav_items_html}}%';
-- Then fixed remaining conditional:
UPDATE content_components
SET html_template = REPLACE(html_template, '{{if .nav_items_html}}{{.quick_links_html}}', '{{if .quick_links_html}}{{.quick_links_html}}')
WHERE id = '09034086-a581-4bba-a5b4-760d863bb2df';
```

### 6. Component Validation Gates — `store_generated_component_action.go`

**Problem:** Component-creator LLM generated 9.5KB of CSS for article-body, hit token limit, stored truncated CSS-only template with `input_schema: {}`. Every blog post using this component rendered empty.

**Fix:** Three validation checks before INSERT:
1. Template must contain `<section` or `<div` — rejects CSS-only
2. `<style>` opens must equal `</style>` closes — rejects truncated output
3. `input_schema` must not be `{}` or empty — rejects fieldless components

### 7. Empty Schema Deferral — `plan_sections_action.go`

**Problem:** Components with `input_schema: {}` were marked `status: "ready"` with no LLM fields. Content writer had nothing to generate, render output raw template.

**Fix:** When a component has empty schema AND its function name contains "article", "content", "body", "text", or "blog", mark as `deferred` instead of `ready`. Non-content components (decorative, separators) still pass through for backward compat.

---

## Data Fixes Applied

### Article-body component — STILL NEEDS APPLYING

The broken component still exists with 9.5KB of truncated CSS:
```sql
-- id: 5835b2e1-50d7-4f20-8a9c-8da4d270ae3d
-- Replacement template provided in the conversation
UPDATE content_components
SET html_template = '<section class="article-body-section" data-component="article-body">
  <div class="container">
    <div class="article-body__content">
      {{.content}}
    </div>
  </div>
</section>
<style>
.article-body-section { padding: 3rem 2rem; background: var(--color-background, #fff); color: var(--color-text, #333); }
.article-body-section .container { max-width: 800px; margin: 0 auto; }
.article-body-section .article-body__content { font-size: 1.0625rem; line-height: 1.75; }
.article-body-section .article-body__content h2 { font-size: 1.75rem; font-weight: 700; color: var(--color-heading, #111); margin: 2.5rem 0 1rem; }
.article-body-section .article-body__content h3 { font-size: 1.375rem; font-weight: 600; color: var(--color-heading, #111); margin: 2rem 0 0.75rem; }
.article-body-section .article-body__content p { margin: 0 0 1.25rem; }
.article-body-section .article-body__content ul,
.article-body-section .article-body__content ol { margin: 0 0 1.25rem 1.5rem; }
.article-body-section .article-body__content li { margin-bottom: 0.5rem; }
.article-body-section .article-body__content a { color: var(--color-primary, #1e40af); text-decoration: underline; }
.article-body-section .article-body__content blockquote { border-left: 4px solid var(--color-primary, #1e40af); margin: 1.5rem 0; padding: 0.75rem 1.25rem; background: var(--color-surface, #f5f5f5); font-style: italic; }
@media (max-width: 768px) { .article-body-section { padding: 2rem 1.5rem; } }
</style>',
    input_schema = '{"fields": {"content": {"type": "text", "source": "llm", "required": true, "llm_guidance": "Write the full article body as HTML. Use h2 for main sections, h3 for subsections, p for paragraphs, ul/ol for lists, blockquote for callouts. Write substantive, useful content — not filler."}}}'::jsonb,
    updated_at = NOW()
WHERE function = 'article-body' AND is_active = true;
```

### Blog post rebuild items
```
rebuild_content_chatgpt-has-your-data-does-that-matter_1368e337-... → triaged
rebuild_content_what-is-rag-and-do-small-businesses-need-it_1368e337-... → triaged (was needs_human_review, reset)
```
These will succeed once the article-body template is fixed.

### Password-entropy removed from header
```sql
UPDATE pages SET in_header = false
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND name = 'password-entropy';
```

### Blog nav_order set
```sql
UPDATE pages SET nav_order = 5
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND name = 'blog';
```

---

## Files Delivered

| File | Changes |
|------|---------|
| `rerender_single_page_action.go` | `sectionHasVisibleContent`, `getPageSections` accepts logger, precompiled regexps |
| `populate_nav_tables_action.go` | Tiered priority, `isChildPageURL`, `PageType` in struct/query, `navLabelForPage` trusts nav_label, overflow to utility |
| `render_site_components_action.go` | `quickLinksItems`, `quickLinksHTML`, `FooterNavItems` on RenderContext, `quick_links_html` in ContentData |
| `component_library.go` | `quick_links_html` substitution in both RenderTemplate paths, `simplifyNavLabelForRender` no longer truncates |
| `store_generated_component_action.go` | Template validation: HTML structure, closed style tags, non-empty schema |
| `plan_sections_action.go` | Empty schema deferral for content-type components |

---

## Pending / Next Session

### Verify after deployment
- [ ] Run `nav_drift_v2:` rebuild → tool child pages should disappear from utility
- [ ] Labels in footer should show full nav_label (not truncated "Ai For")
- [ ] Blog post pages should rebuild with content after article-body template fix
- [ ] FAQ empty section should be hidden by `sectionHasVisibleContent`

### Footer subcategorisation
The Quick Links column now shows ~18 items in a flat list. Future improvement: split into categories (Company, Resources, Tools) with separate template columns. Requires new variables (`company_links_html`, `resources_links_html`) and a new footer template layout.

### Content quality — copy and images
**Copy:** LLM defaults to "pain point → solution" negative framing. Fix via `content_direction` in site_specs:
```json
{
  "avoid_phrases": ["no jargon", "no hype", "without the", "not a"],
  "voice": "direct, confident, specific — state what you do, not what you don't do",
  "guidelines": "Limit to 4-5 good points expressed well. Don't promise too much. Don't assume to know too much. Determine real problems a small AI model training company can handle."
}
```
Needs a focused session on the content writer prompt + content rewrite pass.

**Images:** System has `image_prompts` and `asset-deployer` agent. Gap: hero components use `background-image` CSS but URLs point to missing files. `undeployed_asset` work items exist in the queue.

### FAQ component
Same pattern as article-body — likely empty `input_schema`. Check and fix:
```sql
SELECT function, LENGTH(html_template), input_schema::text
FROM content_components
WHERE function LIKE '%faq%' AND is_active = true;
```

### Improvement loop
Currently off. When re-enabled, it will create `content_rewrite` items for empty heroes, empty CTAs, and other content gaps. The empty section filter prevents these from rendering until content exists.

### Other sites
Nav tier + child page exclusion will apply to all sites once `nav_drift_v2:` items are processed. Check robot-hands.com, gaswholesalers.com, leopardessconsulting.co.uk, ai-agent-orchestration.com for similar improvements.

---

## Key Site IDs
- finetuning.uk: `1368e337-dd1d-4799-bbb3-8221a1b79bcc`
- Blog page: `66a6103c-1961-45f3-a0d2-9228ad6a9188`
- Article-body component: `5835b2e1-50d7-4f20-8a9c-8da4d270ae3d`

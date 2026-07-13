# 010 — Section Editor Architecture

Design decisions and architecture for the section-editor agent, which performs granular edits to individual page sections without re-running the full content generation pipeline.

## The Problem

After a site is built and deployed, we need to make targeted changes — fix a headline, rewrite a case study, swap a component template — without rebuilding the entire page through the content-writer pipeline. We also need these edits to persist through future re-renders (nav updates, theme changes, page rebuilds).

## The Source-of-Truth Principle

Every page section has two representations:

- **`content_data`** (in `page_components`) — structured JSON produced by the content writer. Contains fields like `headline`, `subheadline`, `features[]`, `testimonials[]`, etc.
- **`rendered_html`** (in `page_components`) — final HTML produced by rendering the component template with content_data + site context (colors, navigation, company name).

If we only patch `rendered_html`, the edit drifts from `content_data`. The next time anything triggers a re-render — nav-updater updating navigation, theme change, page-rebuild — the render pipeline works from `content_data` + template, and the HTML patch is lost.

**Decision: content_data is always the source of truth.** Every edit updates content_data first, then re-renders the template. Edits survive all future re-renders because they're stored in the structured data that feeds the render.

## Why Not HTML Patching?

We considered a lightweight `html_patch` edit type (find-and-replace in rendered HTML) for quick fixes like typos. We decided against it because:

- Edits vanish on the next re-render (nav-updater, theme change, page-rebuild all re-render from content_data)
- content_data and rendered_html drift apart, making debugging harder
- The render-from-DB path is fast enough (single DB queries, no LLM calls) that the speed benefit of HTML patching doesn't justify the fragility

## Edit Types

Two edit types, both updating content_data then re-rendering:

### content_edit

Updates section content. Two modes:

- **field_updates** (merge): provide a JSON object of fields to update. Merged into existing content_data. Use for changing a headline, updating a phone number, fixing a testimonial attribution.

  ```json
  {
    "edit_type": "content_edit",
    "field_updates": {
      "headline": "Strategic Consulting for Growth"
    }
  }
  ```

- **content_data** (replace): provide a complete new content_data object. Old content_data is discarded. Use for rewriting a whole case study section, replacing all testimonials, restructuring features.

  ```json
  {
    "edit_type": "content_edit",
    "content_data": {
      "section_title": "How We Help",
      "cases": [
        {"title": "Digital Transformation", "description": "..."}
      ]
    }
  }
  ```

### component_swap

Replaces the component template while keeping existing content_data. The new template is rendered with the existing content and site context. Template field names may differ — `RenderTemplate` handles missing fields gracefully (empty strings via `<no value>` cleanup).

```json
{
  "edit_type": "component_swap",
  "new_component_function": "testimonials-grid"
}
```

This also updates `component_id` and `slot_name` in `page_components`, keeping the DB consistent.

## The buildRenderContextFromDB Function

This is the key function that makes section editing work without the full pipeline. During a normal site build, `RenderContext` is populated from `collected_data` — accumulated output from the strategist, content writer, and design agents. That data doesn't exist when doing a standalone edit.

`buildRenderContextFromDB` constructs the same `RenderContext` entirely from database state:

1. `loadSiteDataFull()` → company name, domain, email, phone, logo text, logo URL
2. `GetStyleCollectionForSite()` → color palette (primary, secondary, accent)
3. `getThemeByID()` → CSS theme content
4. `GetNavItems()` → header and footer navigation items
5. Page metadata → title, description, current page (for active nav state)
6. Section content_data → merged on top as `RenderContext.ContentData`

The `RenderContext.ContentData` map is what templates access via `{{.headline}}`, `{{.features}}`, etc. Site-level fields (company name, colors, nav) are set as both struct fields and ContentData entries, matching how `RenderSiteComponentsAction` builds its context.

## Component Architecture

Each section on a page maps to a `page_components` row:

```
page_components
├── id (uuid)
├── page_id → pages
├── component_id → content_components (the template)
├── position (sort order)
├── slot_name (e.g. "hero", "case-studies-list")
├── rendered_html (the baked HTML)
├── content_data (jsonb — source of truth)
└── build_status
```

The component template lives in `content_components.html_template`. It's a Go template with placeholders like `{{.headline}}`, `{{range .features}}`, etc. The `input_schema` field documents what fields the template expects, though rendering is flexible (missing fields become empty strings).

## Agent Flow

The section-editor agent is self-contained with its own workflow. It's triggered via the standard spawn+call pattern through the generic agent:

```
CLI → generic agent (thin launcher)
       ├─ spawn_agent: section-editor
       └─ call_agent: section-editor
            │
            └─ section-editor (self-contained workflow)
                ├─ ensure_site_record     Load site from domain
                ├─ spawn_deployer         Create deployer instance
                ├─ load_edit_context      Load target page_component,
                │                         component template, page info
                ├─ apply_section_edit     Update content_data, re-render
                │                         from template + DB context,
                │                         reassemble full page HTML
                ├─ git_commit             Commit page to git repo
                ├─ update_page_status     Mark page as deployed
                ├─ trigger_deploy         Call deployer for Cloudflare
                └─ complete
```

### Actions

Two new actions, both using the `ActionInputSpec` pattern:

**`load_edit_context`** — Data gathering. Takes `site_id` + target identifier (`page_component_id` or `page_name` + `slot_name`). Loads the `page_components` row, its linked component template and input_schema from `content_components`, and page metadata. Returns everything the edit step needs.

**`apply_section_edit`** — Performs the edit. Reads `edit_context` from collected_data, applies the content_edit or component_swap, calls `buildRenderContextFromDB` to get full site context, calls `RenderTemplate` to produce new HTML, updates `page_components`, then calls `assemblePage()` (from `rerender_single_page_action.go`) to rebuild the full page with head/header/footer.

### Reused Existing Code

No new DB query patterns or rendering logic was created. The section-editor reuses:

- `assemblePage()` + `getPageInfo()` — from `rerender_single_page_action.go`
- `getSiteComponents()`, `getPageSections()` — page assembly queries
- `loadSiteDataFull()` — from `render_site_components_action.go`
- `GetStyleCollectionForSite()` — style collection loading
- `getThemeByID()` — CSS theme loading
- `GetNavItems()` — navigation queries
- `GetComponentWithFallback()` — component lookup with normalization
- `NormalizeComponentFunction()` — naming contract enforcement
- `RenderTemplate()` — Go template rendering
- `ensure_site_record` — site record loading action
- `git_commit` — git adapter integration
- `update_page_status` — page status management
- `deployer-agent` — Cloudflare deployment

## Target Identification

The section to edit can be identified two ways:

- **By UUID**: `page_component_id` — direct reference to the `page_components` row
- **By name**: `page_name` + `slot_name` — looked up via JOIN on `pages.name` and `page_components.slot_name`

The `slot_name` is normalized via `NormalizeComponentFunction()` before lookup, so `social_proof` and `social-proof` both resolve correctly per the naming contract (008).

## Future Extensions

**Section rewrite via LLM**: The `content_edit` with full `content_data` replacement handles cases where the user provides the new content directly. For cases where the user wants the LLM to rewrite ("make this section more persuasive"), we'd add a workflow branch that spawns the content-writer agent for a single section, then feeds the result back into `apply_section_edit`. The `content_direction` column on the `pages` table would support this.

**Bulk edits**: Multiple sections on the same page could be edited in sequence (loop over an array of edit instructions), with a single `assemblePage` + deploy at the end instead of per-section.

**Logo and asset changes**: File-level changes (uploading a new logo) go through the `deploy_image_asset` action + deployer, bypassing the section-editor entirely. If the logo path changes (not just the file), a nav-updater run re-renders site_components and all pages.


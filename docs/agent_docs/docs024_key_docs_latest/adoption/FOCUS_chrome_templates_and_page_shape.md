# FOCUS — Header/Footer Templates and Page-Shape Canonicalisation

**Date:** 2026-05-06
**Status:** Architecture decided, implementation pending. Two related fixes captured here so neither is lost.
**Symptom:** Live gamesdesign.co.uk has duplicate "Games" entries in header nav and a `/tools/tools/index.html` link in the footer "Our Services" column. Both visible in the deployed HTML.
**Root causes (two distinct):** (1) Header/footer component templates are LLM-generated with hardcoded `<li>` items rather than template variables. (2) Adoption and the planner emit the same logical page in two different shapes (`games` vs `games-index`, `tools` vs `tools-index`), both of which end up deployed.

---

## What this is for

To capture the architectural decisions for fixing the visible navigation pollution on deployed sites, with structural enforcement rather than relying on LLM prompts.

This is Phase 2-style cleanup work that should happen AFTER the directory-builder agent (focus doc `FOCUS_directory_builder_and_list_components.md`). The two are independent — directory-builder addresses list section data, this addresses chrome (header/footer) templates and page-table shape.

---

## Fix 1 — Header/footer templates must be variable-driven, not hardcoded

### What's happening today

Inspecting active footer templates in `content_components`:

```sql
SELECT COUNT(*) FILTER (WHERE html_template LIKE '%{{.nav_items_html}}%')
FROM content_components
WHERE function IN ('footer', 'site-footer', 'footer-4-column', ...)
  AND is_active = true;
-- result: 0 of 2 active footer templates
```

Zero footer templates use any of `{{.nav_items_html}}`, `{{.quick_links_html}}`, or `{{.services_html}}`. The same is true of header templates. The chassis's `RenderSiteComponentsAction` computes these variables and passes them into the render context, but **no template consumes them**. The deployed HTML shows hardcoded `<li><a href="/games.html">Games</a></li>` etc.

This is a contract violation. Doc 003 line 191 explicitly says: *"Use `{{range .nav_items}}` or `{{.nav_items_html}}`, NOT hardcoded links."*

The component-creator prompt either doesn't enforce this constraint, or the LLM ignores it for header/footer components. Across multiple sites and component generations, the consistent failure mode is hardcoded HTML.

### Why this matters

Hardcoded HTML freezes the site's nav at the moment the component was generated. When a new page is added, removed, renamed, or its `in_header`/`in_footer` flag changes, the deployed nav doesn't update — the rendered HTML in `site_components.rendered_html` is stale until the component is regenerated.

It also means the chassis's nav-classification work (`populate_nav_tables`, `classifyPagesForNav`'s child-URL filter, the dedup logic) is bypassed entirely. The chassis correctly computes that `/games/index.html` is a child URL and should be skipped from primary nav — but the component template doesn't read that result, so the LLM-baked `<li>` stays put.

### Why prompt-led won't work

The user's directive is right: *"this shouldn't just be llm prompt led, it should be as algorithmic as possible and enforced."* Three reasons:

1. **LLMs forget instructions.** Telling component-creator's prompt "use `{{.nav_items_html}}` for header nav" is a soft constraint. The LLM ignores it under pressure (long prompts, ambiguous user intent, need to fit other constraints).
2. **Detection is straightforward.** A header/footer template that hardcodes `<a href="...">` to internal pages is structurally detectable: it has `<a href="/...html">` elements outside any `{{range}}` or `{{if}}` block.
3. **The contract is explicit.** Doc 003 has the rule. The system should enforce it, not request it politely.

### Proposed fix

A pre-store validation gate in `store_generated_component_action.go` (sibling to existing gates: substantive-template-no-placeholders, `<no value>` artifacts) that rejects header/footer components with hardcoded internal links.

**Detection algorithm:**

For components with `function IN ('site-header', 'site-footer', 'header-*', 'footer-*')` (or whatever set covers headers and footers — the component's `function` value in conjunction with `category='chrome'` is one place to look):

1. Parse the html_template.
2. Find all `<a href="...">` elements.
3. For each, classify:
   - **External link** (`http://`, `https://`, `mailto:`, `tel:`): allowed.
   - **Anchor only** (`#section`): allowed.
   - **Template variable** (`{{ ... }}` somewhere inside the href): allowed.
   - **Inside a `{{range}}` block**: allowed (LLM produced an iteration, that's fine).
   - **Hardcoded internal link** (`/foo.html`, `/foo/`, etc., not inside a range): **REJECTED**.
4. Hardcoded internal links must be reachable through one of the standard nav variables. The validation checks that the template references at least one of: `{{.nav_items_html}}`, `{{.quick_links_html}}`, `{{.services_html}}`, `{{range .nav_items}}`, `{{range .footer_links}}`, etc. If the template has hardcoded internal links AND none of these variable references, it fails.

**Failure mode:** the work item fails with a descriptive error logged to `agent_error_log` (matches existing gate behaviour). Component-creator retries with the prompt's existing self-check telling the LLM "your previous output had hardcoded internal links — use `{{.nav_items_html}}` instead." The retry budget eventually exhausts if the LLM can't comply, in which case the item goes to triaged/HITL.

**Migration to fix existing 2 active footer templates and the headers:**

Migration `038_chrome_template_repair.sql` — same shape as migration 036. Detect hardcoded chrome templates platform-wide, emit `needs_component_regeneration` work items. The new pre-store gate ensures regenerations comply.

### What to update in component-creator's prompt

Even though the gate enforces, the prompt should still teach:

1. Add an explicit example showing `<nav><ul>{{.nav_items_html}}</ul></nav>` for headers and `<ul>{{.quick_links_html}}</ul>` for footer "Quick Links" columns.
2. Add a constraint statement: *"Header and footer templates MUST use template variables for navigation links. Hardcoded links to internal pages will be rejected."*
3. The template-variable list available at render time should be enumerated in the component spec (it currently isn't — component-creator gets the section_type and a free-form description, but not a list of available render-context variables).

### Available render-context variables

For reference, from `RenderSiteComponentsAction` (line 58791 of bundled chassis source):

| Variable | Source | Use |
|---|---|---|
| `nav_items_html` | `classifyPagesForNav` primary group, pre-rendered `<li>` | Header navs |
| `quick_links_html` | Primary + Utility groups, pre-rendered `<li>` | Footer "Quick Links" or similar |
| `services_html` | `buildServicesHTML` (currently broken, see below) | Footer "Our Services" |
| `categories` (range) | Primary nav as `{name, slug, url, label}` array | Templates that want structured iteration |
| `footerLinks` (range) | Primary + Utility + Legal as `{name, slug, url, label}` | Footer with sectioned navs |
| `companyLinks` (range) | Subset of nav: about, contact, careers | Footer "Company" column |
| `legalLinks` (range) | Privacy, Terms (hardcoded for now) | Footer legal section |

### `buildServicesHTML` is structurally broken — drop it

Separate from the template-variable issue: the `buildServicesHTML` function itself doesn't deduplicate by label or filter child URLs (`/tools/...`, `/games/...`). Its query is a parallel implementation of `classifyPagesForNav` that's missing both filters. The output we observed on gamesdesign — `Tools, Guides, Games, Games, Tools` — is its query result verbatim.

Once Fix 1 lands and chrome templates use `{{.quick_links_html}}` rather than `{{.services_html}}`, the `services_html` codepath has no consumers. Drop `buildServicesHTML` and the `services_html` render-context field. Use `quick_links_html` for footer secondary nav. If a template needs a distinct "services" list later, build it on top of `GetNavItems` with proper classification.

---

## Fix 2 — Adoption and planner must produce the same page shape

> Related: the adoption-side framing of this divergence (Variant C clone intent,
> the ranked fidelity problems) lives in `FOCUS_adoption_fidelity_and_variants.md`.
> This section is the chrome/page-table mechanism; that doc is the adoption goal.

### What's happening today

Two snapshots from gamesdesign's pages table:

```
games           | content          | /games.html             | in_header=t | in_footer=t
games-index     | entity_directory | /games/index.html       | in_header=t | in_footer=t

tool-tools      | tool             | /tools/tools/index.html | in_header=t | in_footer=t
tools-index     | entity_directory | /tools/index.html       | in_header=t | in_footer=t
```

Both rows in each pair represent the same logical page (the "games hub", the "tools hub"). Both rows are deployed. Both are linked from header and footer (when chrome templates regenerate against this state, the LLM sees both and links to both).

### Where the divergence happens

**Adoption** (in `apply_adoption_plan_action.go` lines 52169-52176):

```go
switch pageType {
case "blog-post":
    pageURL = "/blog/" + pageName + ".html"
case "tool":
    pageURL = "/tools/" + pageName + ".html"
default:
    pageURL = "/" + pageName + ".html"
}
```

Adoption produces flat URLs. A "games" hub page comes through with `pageType="content"`, name=`games`, URL=`/games.html`. Stored under name `games`.

**Planner** (via the canonicaliser, post-Phase 1):

The canonicaliser sees the planner's emitted spec (which might be `page_type=entity_directory`, name=`games`) and applies the section-index family rule: name becomes `games-index`, URL becomes `/games/index.html`, page_type retained.

The two rows have different `name` values. The `ON CONFLICT (site_id, name) DO UPDATE` clause in adoption's INSERT can't match them because the names diverge. Two rows persist.

Same for tools: adoption emitted `tool-tools` (apparently — possibly with `pageType="tool"` and a slug-merging quirk). Planner emitted `tools-index`. Both deployed.

### Why this is the right place to fix it

You're right that fixing it earlier in the flow is better than detecting and pruning duplicates later:

1. **Single source of truth.** If adoption and planner agree on canonical naming for the same logical page, there's no duplicate to detect.
2. **No race conditions.** Detection-and-prune operates on a state that already has divergence; it's fragile to ordering and re-runs.
3. **No dead pages in storage.** Even if the renderer/nav-builder filters out duplicates, both rows are still in the pages table consuming row IDs and showing up in admin UIs and reconciler logic.

### Proposed shape: canonical page-name vocabulary

**Both adoption and planner emit through the same canonicaliser.** The canonicaliser already exists (Phase 1, in `page_canonical.go` per the HANDOFF). It's currently called by the planner via `write_site_plan`. Adoption needs to call it too.

The canonical vocabulary (from the HANDOFF):

| Logical concept | Canonical name | Canonical URL | Canonical page_type |
|---|---|---|---|
| Site root | `index` | `/index.html` | `index` |
| Top-level content page | `<slug>` | `/<slug>.html` | `content` |
| Section index hub | `<section>-index` | `/<section>/index.html` | `section_index` (or `entity_directory` / `blog_index` retained) |
| Tool detail page | `tool-<slug>` | `/tools/<slug>/index.html` | `tool` |
| Blog post | `<slug>` (under blog) | `/blog/<slug>.html` | `blog_post` |
| Entity page | `<slug>` (under entities) | `/entities/<slug>.html` (or under parent_section) | `entity_page` |

Adoption today doesn't go through this. It computes its own URL based on `page_type` only. Result: adoption-shape and planner-shape diverge whenever:
- An adopted page becomes a section-index hub at plan time (`games` → `games-index`).
- An adopted page is reclassified at plan time (a generic "tools" content page becomes `tool-tools` with page_type=tool, instead of `tools-index` entity_directory).

### What needs to change

**Concretely:**

1. **`apply_adoption_plan_action.go`** stops computing URLs locally. Instead, it calls `CanonicalisePage(name, page_type, parent_section)` and uses the returned canonical name + URL.
2. **The canonicaliser is updated** if needed to handle the cases adoption sees that the planner doesn't (currently it covers section-index family, blog posts, tools, entity pages — sufficient for the cases we've seen).
3. **For pages where adoption and planner might disagree on `page_type`** (the `games` case): the planner should be authoritative because it has more context. The reconciler already does this — it emits `needs_page` items based on the plan, and pages diverge from the plan get marked stale. The remaining issue is **the orphan adopted page that the planner doesn't include is never removed**.

   This needs an additional reconciler pass: pages on a site with NO entry in `site_plan_pages` (after the planner has run) should be soft-deleted or marked for removal. The reconciler currently only marks pages stale based on plan diff; it doesn't prune orphans. That's a follow-up.

### Order of operations

1. **Update adoption to call the canonicaliser.** Pure additive change. After this, fresh adoptions produce canonical-shape pages; the divergence at adoption time disappears for new sites.
2. **Add a reconciler pruning step.** Pages on an adopted site with no entry in the latest `site_plan_pages` get marked `archived` (or similar terminal status) so the rebuild ignores them and they don't appear in nav. This handles the case where adoption emits a page the planner later removes/renames.
3. **Run a one-off cleanup migration** to mark existing orphans. SQL: find pages where the (site_id, name) tuple has no matching row in the current site_plan for that site, mark them archived.

Steps 1-2 are code changes. Step 3 is a migration that runs once per existing site and is idempotent.

---

## Implementation order across both fixes

These are independent fixes — they can land separately.

**Fix 1 (templates) is small and tactical:**
- Add validation gate to `store_generated_component_action.go` (one function).
- Update component-creator prompt with the explicit constraint and examples.
- Migration `038_chrome_template_repair.sql` to flag existing hardcoded templates and emit regeneration items.
- Estimated work: 1 session.

**Fix 2 (page-shape) is larger:**
- Audit canonicaliser for completeness (does it handle every page_type adoption emits?).
- Modify `apply_adoption_plan_action.go` to call the canonicaliser.
- Build reconciler pruning logic.
- One-off cleanup migration for existing orphan pages.
- Test on gamesdesign (re-adopt cleanly with new code).
- Estimated work: 1-2 sessions.

**Order suggestion:** Fix 1 first. It's smaller and immediately improves visible chrome on every regenerated component. Fix 2 second.

Both come AFTER the directory-builder work (focus doc `FOCUS_directory_builder_and_list_components.md`).

---

## What this fix does NOT do

- Does NOT change the nav-classification logic in `classifyPagesForNav` — that's correct as it stands.
- Does NOT change `populate_nav_tables` to dedup — once chrome templates use `{{.nav_items_html}}`, the nav-classification's existing dedup work reaches the rendered HTML.
- Does NOT touch the directory-builder work — that's about list section data, which is orthogonal.
- Does NOT add new page_types or change existing ones.

---

## See also

- `FOCUS_directory_builder_and_list_components.md` — the directory-builder agent (A/B priority, blocks this work)
- Phase 1 deployment state + the visible site-quality issues (dead BEM, nav dup, fabricated lists) are recorded in the directory-builder doc's Phase-1 history section and in `FOCUS_adoption_fidelity_and_variants.md`
- `003_contracts_and_standards_v7.md` line 191 — the explicit "use template variables, NOT hardcoded links" rule
- Doc 002 — page_canonical Phase 1 work
- `apply_adoption_plan_action.go` lines 52169-52176 — where adoption's URL divergence originates
- `render_site_components_action.go` — render context construction (works correctly today, just not consumed by templates)

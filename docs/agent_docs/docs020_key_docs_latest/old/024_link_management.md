# 024 — Link Management

Focus document for the link subsystem. Covers what exists, what's planned, what's missing, and the design direction for making links a first-class entity.

---

## Design Principle

Links are first-class objects. Every link on a site — navigation, embedded content, external, affiliate, CTA — lives in `link_registry` with provenance, classification, and validation state. When pages are deleted, moved, or rewritten, the link registry is the authority on what breaks and what needs updating.

---

## Current State

### Schema: `link_registry`

Exists and is indexed. Full schema:

| Column | Type | Purpose |
|--------|------|---------|
| `id` | uuid PK | |
| `source_component_instance_id` | uuid FK → page_components | Which component contains this link |
| `source_page_id` | uuid FK → pages (CASCADE) | Which page contains this link |
| `source_site_id` | uuid FK → sites (CASCADE) | Which site |
| `target_url` | varchar(1000) | The href value |
| `target_page_id` | uuid FK → pages | Resolved internal target (nullable) |
| `target_site_id` | uuid FK → sites | Resolved target site (nullable) |
| `scope` | varchar(50) | `internal`, `page`, `external` |
| `link_type` | varchar(50) | `navigation`, `content`, `semantic` |
| `anchor_text` | varchar(500) | Link text |
| `rel_attr` | varchar(100) | rel attribute value |
| `affiliate_provider` | varchar(100) | Affiliate network name |
| `affiliate_tag` | varchar(255) | Affiliate tracking tag |
| `affiliate_product_id` | uuid FK → affiliate_products | |
| `requires_disclosure` | boolean | Whether affiliate disclosure needed |
| `status` | varchar(50) | `active`, `broken`, `removed` |
| `last_validated_at` | timestamptz | Last HTTP check |
| `validation_result` | varchar(50) | `ok`, `404`, `timeout`, etc. |

Indexes: source_page, source_site, source_component, target_page, target_site, scope, link_type, broken links, affiliate.

FK cascades: source_page and source_site CASCADE on delete (links removed when page/site deleted). target_page and target_site do NOT cascade — broken link detection catches these.

### Schema: `site_nav_groups` and `site_nav_items`

Navigation is stored separately from the link registry. These tables are the authority for structured navigation.

```
site_nav_groups: id, site_id, group_key, group_label, group_type, position
site_nav_items:  id, site_id, group_id, label, url, page_id, item_type, position, status
```

**Group types:** `primary`, `legal`, `utility`, `content`, `subsection`, `external`.

Navigation and the link registry are related but separate concerns. Nav tables define *structure* (what appears in header/footer and in what order). The link registry tracks *instances* (every `<a>` tag in rendered HTML, including nav links).

---

## Go Code: What Exists

### `extract_and_sync_links` action

File: `site_actions.go` (within ExtractAndSyncLinksAction)

Called after page HTML is generated. Parses HTML with goquery, extracts all `<a href>` tags, classifies them, and syncs to `link_registry` via delete-and-reinsert per page.

**Classification logic:**

| Function | Returns | Logic |
|----------|---------|-------|
| `classifyLinkScope(href)` | `internal`, `page`, `external` | `#` → internal, `/` → page, `http` → external |
| `classifyLinkTypeFromContext(selection)` | `navigation`, `content`, `semantic` | Walks up DOM: nav/header/footer → navigation, cta/button → content, related → semantic, default → content |

**Current sync approach:** Delete all links for `source_page_id`, then insert fresh. Simple but loses `last_validated_at` and validation history on each page rebuild.

### `InjectLinkConstraints`

File: `link_constraints.go`

Prepends an "INTERNAL LINKS" section to content writer prompts listing all valid pages. Prevents the LLM from inventing links to pages that don't exist. Configurable via `LinkConstraintsConfig.Enabled` and `MaxInternalLinksPerSection`.

Sources checked for available pages (priority order): `db_sync.pages`, `render_context.available_pages`.

### `validateInternalLinks` (within validate_page_content.go)

Post-generation check. Scans rendered HTML for `href` values, skips external/anchor/mailto/asset links, checks remaining against `pages` table. Missing targets are **warnings** (not blockers) because the target page may be planned but not yet built.

### `populate_nav_tables` action

File: `populate_nav_tables_action.go`

Reads pages for a site, classifies them into nav groups (primary, legal, utility), populates `site_nav_groups` and `site_nav_items`. Called by nav-updater and during initial build.

**Classification:** Pages with `in_header=true` go to primary group. Privacy/terms/cookies pages go to legal group. Remaining `in_footer=true` pages go to utility group.

### `fix_nav_link_templates` action

Fixes broken template references in header/footer. Replaces patterns like `href="#{{.slug}}"` with `href="{{.url}}"`. Used by `nav-link-fixer` agent.

### `addToolToNav`

In `deploy_tool_action.go` and `create_tool_component_action.go`. Adds tool pages to a "Tools" nav group. Creates group if it doesn't exist.

### Navigation rendering

`GetNavItems()` and `GetNavigationStructure()` read from nav tables with fallback to pages table. Used by `render_site_components` to populate header/footer templates with `{{.nav_items_html}}`.

`extractSitemapInfo()` has a 4-tier priority for resolving navigation in content writer prompts: db_sync → link_data → page_plan → sitemap.

---

## Agent Architecture: What Exists

### Agents that interact with links

| Agent | How it touches links |
|-------|---------------------|
| `nav-updater` | Rebuilds nav tables from pages, re-renders header/footer, deploys |
| `nav-link-fixer` | Fixes broken `#slug` links in header/footer templates |
| `page-build-handler` → `page-content-writer` | Content writer receives link constraints, generates HTML with links, `extract_and_sync_links` runs post-build |
| `rerender-pages` | Re-assembles pages — header/footer nav links update as side effect |
| `tool-deployer` / `tool-generator` | Adds tool pages to nav via `addToolToNav` |
| `content-gap-planner` | Creates `content_rewrite` items that may result in new internal links |
| `internal-linker` (new, from this session) | Finds pages to link to orphaned sub-pages |

### Discovery checks that detect link issues

| Check | File | Detects | Handler |
|-------|------|---------|---------|
| `broken_nav_links` | `check_broken_nav_links.go` | Nav using `#slug` instead of page URLs | `nav-link-fixer` |
| `orphan_pages` | `check_orphan_pages.go` | Pages with no inbound links from nav or content | `content-gap-planner` (orphans), `internal-linker` (sub-pages) |
| `validateInternalLinks` | `validate_page_content.go` | Href to non-existent pages (post-build) | Warning only |

### Planned but not implemented (from 002_system_architecture)

**Links Agent Family** — owner: `links-orchestrator` (algorithmic, no LLM)

Sub-agents: `link-crawler`, `link-validator`, `link-registry-sync`, `redirect-manager`, `affiliate-link-manager` (phase 2).

**Does:** Extract links from HTML, classify types, resolve internals, HTTP HEAD checks, detect broken/orphaned, generate redirects, track link counts.

**Does not:** Decide where to place links (content writer), navigate structure (nav agent), SEO strategy (SEO agent).

---

## Link Lifecycle: Where Links Are Created

### 1. Initial build (planner → content writer)

Pages planned by site-planner. Content writer generates HTML with `<a>` tags. Link constraints prompt injection limits links to known pages. After content generation, `extract_and_sync_links` populates link_registry.

**Gap:** Links created during initial build are not systematically resolved against `target_page_id`. The `target_page_id` column stays NULL for internal links.

### 2. Navigation (populate_nav_tables → render_site_components)

Nav tables populated from pages. Header/footer templates rendered with `{{.nav_items_html}}`. These nav links also appear in `link_registry` via `extract_and_sync_links` as `link_type: navigation`.

**Gap:** Nav links exist in both `site_nav_items` AND `link_registry` — dual source of truth. No reconciliation between them.

### 3. Tool deployment (tool-deployer / tool-generator)

`addToolToNav` adds pages to nav groups. Cross-link items create `content_rewrite` work items that weave tool references (with links) into content pages.

### 4. Improvement loop (auditors → fix agents)

Content quality auditor may flag missing links. Content-gap-planner creates pages or `content_rewrite` items. `internal-linker` creates specific link placement instructions.

### 5. Adoption (future)

Adoption agent crawls existing site, extracts all links. Need to: classify which are navigation vs content vs external, resolve internal targets, flag broken ones, decide which to keep.

---

## Link Types Taxonomy

| Type | Scope | Where found | Management |
|------|-------|-------------|------------|
| **Header nav** | page | `<nav>` in header component | `site_nav_items` + rendered in header template |
| **Footer nav** | page | `<footer>` component | `site_nav_items` + rendered in footer template |
| **Subsection nav** | page | Feature sections (blog listing, tool index) | Page-specific, rendered by section components |
| **Content internal** | page | Body text `<a>` tags linking to other pages | Content writer decides placement |
| **Anchor** | internal | `#section-id` within same page | Component templates |
| **CTA** | page | Buttons/banners linking to contact/services | Content writer + template |
| **External** | external | Links to third-party sites | Need `rel="noopener"`, possibly external indicator |
| **Affiliate** | external | Product/service referral links | `affiliate_provider`, `affiliate_tag`, disclosure required |
| **Social** | external | Social media profile links | Usually in footer, template-driven |
| **Blog/news cross-ref** | page | Blog posts linking to related pages | Article writer, cross-linking pipeline |
| **Tool cross-link** | page | Content pages referencing tool pages | Tool cross-link pipeline (working) |

---

## Gaps and Issues

### 1. `target_page_id` never resolved

Internal links in `link_registry` have `target_url` set (e.g. `/about.html`) but `target_page_id` is NULL. This means:
- Can't detect broken links when a page is deleted (would need URL matching)
- Can't count inbound links per page efficiently
- Can't do impact analysis before deleting a page

**Fix needed:** Post-extraction step that resolves `target_url` → `target_page_id` by matching against `pages.url` for same-site links.

### 2. Delete-and-reinsert loses history

`syncLinksToDB` deletes all links for a page then reinserts. This resets `last_validated_at`, `validation_result`, and loses the link `id` (breaking any external references).

**Fix needed:** Upsert pattern — match on `(source_page_id, target_url)`, update anchor_text/link_type if changed, insert new, soft-delete missing.

### 3. No external link validation

No agent runs HTTP HEAD checks on external links. The `validation_result` and `last_validated_at` columns exist but are never written.

**Fix needed:** Discovery check or scheduled task that validates external links (scope='external', status='active') periodically. Low frequency — weekly per link.

### 4. No page-deletion impact check

When pages are deleted (by out-of-date reapers or manual action), nothing checks `link_registry` for inbound links that will break. FK cascade only removes links *from* the deleted page, not links *to* it.

**Fix needed:** Pre-deletion hook or post-deletion discovery check that finds `link_registry` rows where `target_page_id = <deleted page>` and creates fix work items.

### 5. No external link UX component

External links should visually indicate they leave the site. No standard component or CSS class exists for this.

**Fix needed:** Component contract addition — external links get a class (`ext-link`) and optional icon. Content writer prompt should instruct `rel="noopener noreferrer"` and `target="_blank"` for external links.

### 6. No adoption link extraction

The adoption agent crawls pages but doesn't extract or classify links from the crawled content. Link structure of the original site is lost.

**Fix needed:** During adoption, run link extraction on crawled HTML. Store in `link_registry` with `source='adoption'`. Classify nav vs content vs external. Use to inform the strategist about the site's link topology.

### 7. Nav-registry reconciliation

Nav links exist in both `site_nav_items` (structural authority) and `link_registry` (instance tracking). No process reconciles them. If a nav link changes in `site_nav_items` but the header isn't re-rendered, the registry is stale.

**Fix needed:** After `populate_nav_tables` or `render_site_components`, sync navigation links to registry. Or: treat nav links in the registry as derived (re-extracted on each render) and accept they refresh on rebuild.

### 8. Mobile-specific link handling

Some links may behave differently on mobile (e.g. tel: links, simplified nav). No system support for this currently.

---

## Proposed Agent Architecture

### Phase 1: Foundation (link registry as source of truth)

1. **Fix `extract_and_sync_links`** — resolve `target_page_id`, upsert instead of delete-reinsert
2. **Add `broken_link_check` discovery check** — find registry entries where `target_page_id` references a deleted/inactive page, or where external `validation_result` is stale
3. **Add external link validation scheduled task** — HTTP HEAD on external links, update `validation_result`

### Phase 2: Link-aware agents

4. **`link-validator` agent** — handles `broken_internal_link` work items. Decides: remove link, update target, create redirect
5. **`redirect-manager` agent** — handles `redirect_needed` work items. Creates redirect rules when pages move
6. **Pre-deletion impact check** — before page deletion, query link_registry for inbound links, create `broken_link` work items

### Phase 3: Adoption and intelligence

7. **Adoption link extraction** — extract and classify links during site adoption
8. **Link analytics** — track link density, external/internal ratio, orphan pages via registry queries
9. **Affiliate link management** — dedicated handling for affiliate links with disclosure requirements

---

## Related Tables

| Table | Relationship to links |
|-------|----------------------|
| `link_registry` | Primary link storage |
| `site_nav_groups` | Navigation structure groups |
| `site_nav_items` | Navigation structure items |
| `pages` | Source and target of internal links |
| `page_components` | Source component that contains links |
| `sites` | Site-level link scope |
| `affiliate_products` | Affiliate link targets |
| `content_components` | Templates that generate links (nav templates use `{{.nav_items_html}}`) |
| `site_components` | Rendered header/footer containing nav links |

---

## Work Item Types (link-related)

| item_type | handler_agent | status |
|-----------|---------------|--------|
| `broken_nav_links` | `nav-link-fixer` | Working |
| `needs_internal_links` | `internal-linker` | New (this session) |
| `nav_drift` | `nav-updater` | Working (timeout issues) |
| `broken_internal_link` | TBD (`link-validator`) | Not implemented |
| `redirect_needed` | TBD (`redirect-manager`) | Not implemented |
| `stale_external_link` | TBD (`link-validator`) | Not implemented |

---

## Content Writer Link Instructions

The content writer receives link guidance through multiple channels:

1. **`InjectLinkConstraints`** — "ONLY link to pages from this list" (prevents hallucinated links)
2. **`rewrite_guidance`** — cross-link instructions from tool pipeline or internal-linker ("Add a contextual link to /tools/roi-calculator.html")
3. **`content_direction`** — page-level spec that may include link instructions

The content writer does NOT receive the full link registry. It receives a curated list of valid pages and optional guidance about where specific links should go.

### External link handling in prompts

Currently no specific instruction to the content writer about external links. Need to add:
- Use `rel="noopener noreferrer"` and `target="_blank"` for external links
- Add CSS class `ext-link` for styling
- Don't fabricate external URLs — only use URLs from research results or site spec

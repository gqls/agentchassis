# 009 — Improvement Loop (v4)

Post-build quality improvement cycle. Runs discovery agents (algorithmic), audit agents (LLM-based), triages findings, dispatches fixes, and rerenders.

**v4 changes (2026-03-28):** Orphan page detection, blog listing rebuild in rerender pipeline, content writer no longer injects chrome (prevents double header/footer), link constraints subdirectory awareness.

**v3 changes (2026-03-25):** Audit pass guard (max 3 passes per site), structured findings with acceptance criteria (max 5 per auditor), section locking exclusion in data queries.

---

## Flow

```
improvement-loop (orchestrator)
  │
  ├── 1. ensure_site_record
  │
  ├── 1b. load_pass_count → check_audit_pass_limit
  │     └── if pass_count >= 3 → notify_scheduler_clean → complete_clean
  │         (site has reached max audit passes — no agents spawned)
  │
  ├── 2. quality-discovery-agent (algorithmic)
  │     └── broken_nav_links, placeholder_contact, generic_theme
  │
  ├── 3. design-discovery-agent (algorithmic)
  │     └── hardcoded_section_colors, forced_text_colors,
  │         undeployed_assets, missing_css, validate_component_standards
  │
  ├── 4. completeness-discovery-agent (algorithmic)
  │     └── empty_sections (excludes blog pages)
  │     └── empty_blog (routes to blog-content-planner)
  │     └── orphan_pages (v4: detects unreachable deployed pages)
  │
  ├── 5. design-audit-agent (LLM-based orchestrator)
  │     ├── visual-design-auditor
  │     │     algorithmic checks → one LLM call (TOP 5 findings) → write_audit_findings
  │     │     checks: colour consistency, spacing, typography, dark sections, responsive
  │     │     data queries EXCLUDE locked components (locked_at IS NULL)
  │     └── content-quality-auditor
  │           load brief + page samples (excluding locked) → one LLM call (TOP 5) → write_audit_findings
  │           checks: tone alignment, content gaps, CTA effectiveness, differentiation
  │
  ├── 6. site-review-agent (LLM-based orchestrator)
  │     ├── content-quality-auditor (reused)
  │     └── strategic alignment review (own LLM call, TOP 5 findings)
  │           checks: purpose alignment, page structure, dream spec gaps, conversion path
  │
  ├── 7. triage_detected_items → promote detected → triaged
  │
  ├── 7b. increment_audit_pass (tracks pass count in sites.settings)
  │
  ├── 8. check_has_findings → if none, complete_clean
  │
  ├── 9. insert needs_rerender (priority 99, runs after all fixes)
  │
  └── 10. build-dispatch-loop → processes all triaged items → rerender
```

**Ordering is deliberate:** Structural checks (2-4) run first — they're cheap, fast, and fix issues that would confuse the LLM audits. LLM audits (5-6) run after and see corrected HTML. Triage (7) catches everything. Dispatch (10) processes items by priority.

---

## Triage Drain Controls

Three mechanisms prevent unbounded audit-fix-reaudit cycles:

### 1. Finding Cap (migration 084)

Each LLM auditor reports UP TO 5 findings per pass. Prompts include:
- "Report ONLY the TOP 5 most impactful issues"
- "Do NOT report issues already caught by algorithmic checks"
- Each finding requires: `current_value`, `acceptance_test`, `suggestion`, `max_fix_attempts`

Estimated per-domain findings: max 15 per pass (5 visual + 5 content + 5 strategic) vs 50+ previously.

### 2. Section Locking (migration 086)

Components that pass verification get `locked_at` set. Locked components are excluded from audit data-loading queries:
- `visual-design-auditor.load_design_context`: `AND pc.locked_at IS NULL` on index samples
- `content-quality-auditor.load_page_content`: `AND pc.locked_at IS NULL` on page content
- `content-quality-auditor.check_empty_pages`: doesn't count pages with only locked components as empty
- All discovery agents already filter `AND pc.locked_at IS NULL`

Locked sections are invisible to auditors — they can't be reported on or generate work items.

Unlock is always manual: `UPDATE page_components SET locked_at = NULL, locked_by = NULL WHERE id = ?`.

### 3. Audit Pass Cap (migrations 086 + 087)

Sites track `audit_pass_count` in `sites.settings.maintenance_profile`:

```
Pass 1 (initial audit): top 5 findings per auditor → fix → verify → increment pass
Pass 2 (re-audit): top 5 findings → fix → verify → increment pass
Pass 3 (final): top 5 findings → fix or accept → increment pass
Pass 4+: improvement-loop loads pass count, sees >= 3, exits immediately
```

Functions:
- `get_audit_pass_count(site_id)` — reads from `sites.settings`
- `increment_audit_pass(site_id)` — called by improvement-loop after triage
- `reset_audit_passes(site_id)` — manual reset (e.g. after major site redesign)

Operator view: `SELECT * FROM site_locking_progress` shows per-site component counts, locked counts, pass counts, and overall status.

### Combined Effect

```
Before (per domain):  ~8+ audit passes, ~50+ findings per pass = ~88K+ tokens
After (per domain):   3 audit passes, ~15 findings per pass + verification = ~30K tokens
Reduction: ~65-70%
```

---

## Structured Findings Format (migration 084)

Each LLM finding now includes:

```json
{
  "category": "colour",
  "severity": "medium",
  "description": "Hero uses hardcoded #1a1a2e instead of CSS variable",
  "current_value": "background: #1a1a2e in .hero-centered inline style",
  "suggestion": "Replace with var(--color-primary) or var(--section-bg-dark)",
  "acceptance_test": "Hero section background uses a CSS variable, not a hardcoded hex value",
  "affected_component": "hero-centered",
  "page": "index",
  "max_fix_attempts": 2
}
```

The `acceptance_test` enables a cheap verification call after fixing:
```
"The acceptance test is: 'Hero section background uses a CSS variable, not a hardcoded hex value.'
The fixed HTML is: [rendered_html snippet]
Does it pass? YES or NO with brief explanation."
```

This is a tiny LLM call — could run on Mistral on CPU. Pass → item complete, lock section. Fail → one more attempt. After `max_fix_attempts` → `needs_human_review`.

The Go `auditFinding` struct captures these fields and writes them to `site_work_items.spec` JSONB.

---

## Discovery Agents (Algorithmic)

These run Go discovery checks registered via `init()`. Each check queries the database, finds issues, and creates `site_work_items` with `status: detected`. No LLM calls.

### quality-discovery-agent

| Check | Detects | Handler |
|-------|---------|---------|
| `broken_nav_links` | Nav using anchor links (#slug) instead of page URLs | `nav-link-fixer` |
| `placeholder_contact` | Generic contact details from templates | `page-content-writer` |
| `generic_theme` | Default theme with no customisation | `webdesign-agent` |

### design-discovery-agent

| Check | Detects | Handler |
|-------|---------|---------|
| `hardcoded_section_colors` | Inline hex colours that should use CSS variables | `color-variable-fixer` |
| `forced_text_colors` | Child text elements with hardcoded `color: #hex` | `color-variable-fixer` |
| `undeployed_assets` | Images in assets table not committed to git | `asset-deployer` |
| `missing_css` | Site with no `/assets/css/styles.css` deployed | `webdesign-agent` |
| `validate_component_standards` | Unlinked site components, slot mismatches, missing metadata | `site-component-linker` / `component-template-fixer` |

### completeness-discovery-agent

| Check | Detects | Handler |
|-------|---------|---------|
| `empty_sections` | Page sections with null/empty/near-empty rendered HTML (excludes blog pages) | `page-build-handler` |
| `empty_blog` | Blog page exists but no blog-post pages | `blog-content-planner` |
| `orphan_pages` | Deployed pages with no inbound links from nav, header/footer, or other pages | `rerender-pages` (blog) / `content-gap-planner` (content) |

### orphan_pages check detail (v4)

Finds deployed pages that are unreachable — not linked from `site_nav_items`, not referenced in any `site_components` rendered_html (header/footer), and not linked from any `page_components` rendered_html on other pages.

Exclusions: index/home (always reachable as /), blog-index (in nav), tool pages (may have external entry points).

Work item routing:
- Blog post orphans → single `orphan_blog_posts` work item → `rerender-pages` (which runs `rebuild_blog_listing` to regenerate the listing from current posts)
- Content page orphans → individual `orphan_page` work items → `content-gap-planner` (decides whether to add nav entry or internal links)

---

## Blog Listing Rebuild (v4)

The `rebuild_blog_listing` action runs as part of the `rerender-pages` workflow, before `get_pages`:

```
rerender-pages workflow:
  check_refresh_components
    → (true) render_site_components → rebuild_blog_listing
    → (false) rebuild_blog_listing
  rebuild_blog_listing → get_pages → check_pages_exist → create_rerender_items → complete
```

The action:
1. Finds `blog-index` pages for the site (skips if none exist)
2. Queries deployed `blog-post` pages ordered by creation date
3. Loads the blog listing template from `content_components` (tries `blog-listing`, falls back to `article_grid`, then a minimal default)
4. Renders the template with post data using `RenderTemplateWithMap`
5. Upserts the `blog-listing` page_component

No LLM, no hardcoded styles. The template comes from the component library — different style collections can have different listing layouts. The site's CSS provides all styling through class names and variables.

Trigger points: every rerender (keeps listing current), and after blog post publishing (dispatch loop creates `needs_rerender` after blog posts are built).

---

## Content Writer Chrome Fix (v4)

The `page-content-writer` agent definition had `inject_header: true, inject_footer: true, inject_head: true` in its `compile_page` step config. This caused the content writer to embed header/footer/head into each page's HTML, which `save_page_sections` then stored as page_components. The rerender then added site_components header/footer on top → double header/footer on every page.

**Fix:** Set all three to `false`. Site chrome is now only injected by the rerender/assembly step, never by the content writer.

**Cleanup:** Removed baked-in header/footer page_components across all sites. These were identifiable by `slot_name LIKE 'header-%'` or `slot_name LIKE 'footer-%'` or matching `content_components WHERE component_level = 'site'`.

---

## Link Constraints Subdirectory Fix (v4)

`link_constraints.go` builds page URLs from page names when the `url` field is empty. The fallback assumed all pages live at `/{name}.html`, which is wrong for blog posts (`/blog/{name}.html`) and tools (`/tools/{name}.html`).

Fix: the fallback now checks `page_type` and maps to the correct subdirectory:
- `blog-post` → `/blog/{name}.html`
- `tool` → `/tools/{name}.html`
- default → `/{name}.html`

Low immediate risk (the `url` field is populated in the DB), but prevents future issues when page data flows through without the URL field.

---

## Audit Agents (LLM-Based)

These load site context and specs (excluding locked components), run algorithmic pre-checks, then make one LLM call per group for subjective assessment. Each call produces UP TO 5 structured findings with acceptance criteria. Findings are written via `write_audit_findings` action with dedup and blocked-item filtering.

### design-audit-agent (orchestrator)

Spawns two group auditors:

**visual-design-auditor:**
- Loads: style collection, CSS theme, colour palette, rendered HTML samples (locked components excluded)
- Algorithmic checks: unlinked components count, slot mismatches, nav stacked, dark sections missing contract
- LLM checks (TOP 5): colour consistency, spacing regularity, typography hierarchy, responsive issues
- Prompt instructs: "Do NOT report issues already caught by algorithmic checks"
- Findings routed to: `webdesign-agent`, `color-variable-fixer`, `component-template-fixer`

**content-quality-auditor:**
- Loads: site specs (identity, content_direction), page content samples (locked excluded), empty pages list
- LLM checks (TOP 5): tone alignment with spec, content gaps, CTA effectiveness, differentiation, audience targeting
- Findings routed to: `page-build-handler`, `component-template-fixer`

### site-review-agent (orchestrator)

Spawns content-quality-auditor (reused), then runs its own strategic review:
- Loads: site specs, dream spec (if exists), deployed page count, content audit results
- LLM checks (TOP 5): overall purpose alignment, page structure gaps, biggest improvements, conversion path
- Findings routed to: `page-build-handler`, `site-component-linker`

---

## How Audit Agents Read Site Specs

Audit agents compare rendered state against declared intent from `site_specs`:

```
Classifier wrote:  tone = "conversational, direct, anti-corporate"
Content writer produced:  "We synergize cutting-edge solutions..."
Audit finds:  tone mismatch → work item: content_rewrite
```

The content-quality-auditor's `load_brief` step queries `site_specs` for identity, content_direction, and falls back to `sites.content_data` for older sites. The visual-design-auditor loads `design_intent` from specs and compares against rendered CSS and HTML.

Audit agents enforce the spec — they don't make design or content decisions.

---

## Finding Dedup and Blocked Item Filtering

`write_audit_findings` action prevents duplicate and blocked findings:

1. **Bulk preload:** Loads all blocked item keys for the site, skips matching findings
2. **Broader match:** Checks if a blocked item exists with same item_type + page
3. **Dedup:** Checks if a pending item (detected/triaged/claimed/blocked) already exists with same item_key

Item keys: `{audit_source}_{item_type}_{page}_{site_id}`

---

## Fix Agents (Handlers)

Called by the dispatch loop via spawn→call. Each receives raw identifiers and loads its own context.

| Handler | Fixes | How |
|---------|-------|-----|
| `color-variable-fixer` | `hardcoded_section_colors`, `forced_text_colors` | Replaces hex with CSS variables, injects --section-* contract for dark sections, WCAG contrast check |
| `webdesign-agent` | `needs_design_review`, `missing_css`, `generic_theme` | Regenerates CSS from design spec |
| `page-build-handler` | `needs_content_page`, `content_rewrite`, `tone_shift` | Wraps page-content-writer with persistence |
| `component-template-fixer` | `cta_improvement`, `spacing_fix`, `responsive_fix` | CSS injection, element modification |
| `site-component-linker` | `unlinked_site_component`, `header_footer_fix` | Links site_components to style collection templates |
| `asset-deployer` | `undeployed_asset` | Downloads from S3, optimises, commits to git |
| `blog-content-planner` | `needs_blog_posts` | Plans blog posts via LLM, creates pages + work items |
| `rerender-pages` | `needs_rerender`, `orphan_blog_posts` | Rebuilds blog listing, re-assembles pages from stored components, deploys |
| `content-gap-planner` | `orphan_page` | Decides whether to add nav entry or create internal links to unreachable page |

---

## Triggering

| Trigger | How |
|---------|-----|
| Manual | `./trigger-audit.sh improvement-loop <site_id> <domain>` |
| Post-build | Side-effect after last build item completes (insert work item for improvement-loop) |
| Scheduled | `improvement-sweep` scheduled task (every 600s, finds least-recently-built site) |
| Per-agent | Individual auditors can be triggered directly for targeted checks |

---

## Configuration

Per-site audit config in `sites.settings.maintenance_profile`:

```json
{
    "audit_pass_count": 0,
    "audit": {
        "visual_design": { "enabled": true, "every": "7d" },
        "content_quality": { "enabled": true, "every": "7d" },
        "strategic_review": { "enabled": true, "every": "30d" }
    }
}
```

`audit_pass_count` is managed by `increment_audit_pass()` / `reset_audit_passes()`. Currently all audits run every cycle up to the 3-pass cap. Per-site config filtering is a future enhancement.

---

## Timeout

The improvement loop timeout is 1800 seconds (30 minutes). Typical run:
- Pass count check: < 1 second (exits immediately if >= 3)
- Algorithmic checks: 10-30 seconds total
- Design audit (2 LLM calls, 5 findings each): 30-60 seconds
- Site review (2 LLM calls, 5 findings each): 30-60 seconds
- Dispatch + fixes: varies by finding count (much reduced with cap)
- Rerender (includes blog listing rebuild): 30-60 seconds

Total: 3-8 minutes for a typical site with findings. < 1 second for sites at pass limit.

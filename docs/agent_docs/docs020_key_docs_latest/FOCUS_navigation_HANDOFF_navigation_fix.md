# Handoff: Navigation Duplication Fix & Site Design Planner

Date: 2026-04-11

---

## Context

Sites (ai-agent-orchestration.com, finetuning.uk, and potentially others) are rendering duplicate headers and footers. This investigation traced the root cause, applied data fixes, and identified a broader architectural gap in how navigation and site-level layout decisions are made.

---

## What Was Found

### The Duplicate Header/Footer Problem

**Symptom:** Pages render two headers (with different nav items, different CTAs, one with logo, one without) and two footers.

**Root cause:** The `pages.sections` column listed site-level component names (`header-professional-dark`, `footer-standard`, `site-header`, `site-footer`) alongside content sections (`hero`, `features`, `call-to-action`). When any agent rebuilt a page (content_rewrite, page_rerender, etc.), it rendered header/footer as regular `page_components` rows. Then `InjectHeader`/`InjectFooter` added a *second* header/footer from `site_components` during page assembly.

**How they got there:** The site planner or adoption flow saw headers/footers in the original site structure and included them in the page plan. This is reasonable behaviour — the planner sees what's on the page and records it. The problem is downstream: nothing filtered them out before treating them as content sections to render.

**Why it appeared recently:** Discovery agents started creating `content_rewrite` work items, which triggered `page-build-handler`. This rebuilt content pages (index, about, services, contact, case-studies) and re-saved header/footer as `page_components`, producing the duplication. Blog posts and tool pages were unaffected because their `pages.sections` never included header/footer names.

### Data Showing the Problem

Affected pages before fix (all had `header-professional-dark` and `footer-standard` in `pages.sections`):
- ai-agent-orchestration.com: index, services, about, contact, case-studies
- finetuning.uk: index, services, about, contact, faq, use-cases, approach

The `page_components` table had rows with `slot_name` like `header-professional-dark` and `footer-standard`, all with NULL `component_id` (came through HTML regex fallback, not structured metadata path).

### Other Nav Problems Identified (Not Yet Fixed)

| # | Problem | Root cause |
|---|---------|------------|
| 3 | Tool pages listed individually in primary nav | `addToolToNav` adds to primary group, no "Tools" grouping |
| 4 | Tool labels too long ("LLM Provider Cost Comparison Calculator") | No label shortening for nav context |
| 5 | Truncated to meaningless "AI Agent" | `rerenderSimplifyNavLabel` takes first word when >20 chars |
| 8 | max_header_items: 8 too generous | Hardcoded default, not design-driven |
| 9 | Hover colour #0f3460 invisible on #1a1a2e background | Hardcoded in header-professional-dark template |
| 10 | Responsive CSS injected 4× by component-template-fixer | Idempotency check is case-sensitive ("responsive fix" vs "Responsive fix") |
| 11 | First footer has `background: ;` (empty) | Primary colour not reaching footer template's inline CSS |
| 12 | Logo missing from injected header | `logo_url` not a direct field on RenderContext, flows through ContentData merge which is unreliable |
| 13 | Placeholder email "agents@contactforsales.com" | Data issue |
| 14 | Footer "Our Services" column has junk entries | `buildServicesHTML` query returning wrong pages |

---

## What Was Fixed (Data)

### Fix 1: Stripped header/footer from pages.sections

```sql
UPDATE pages
SET sections = (
    SELECT jsonb_agg(elem)
    FROM jsonb_array_elements_text(sections) AS elem
    WHERE elem NOT LIKE '%header%'
      AND elem NOT LIKE '%footer%'
      AND elem NOT LIKE 'site-header'
      AND elem NOT LIKE 'site-footer'
)
WHERE sections::text LIKE '%header%'
   OR sections::text LIKE '%footer%';
-- Updated 12 rows
```

### Fix 2: Deleted dirty page_components rows

```sql
DELETE FROM page_components
WHERE id IN (
    SELECT pc.id
    FROM page_components pc
    JOIN pages p ON pc.page_id = p.id
    JOIN sites s ON s.id = p.site_id
    WHERE (pc.slot_name LIKE '%header%' OR pc.slot_name LIKE '%footer%'
           OR pc.slot_name = 'site-header' OR pc.slot_name = 'site-footer')
      AND s.domain IN ('ai-agent-orchestration.com', 'finetuning.uk')
);
-- Deleted 24 rows
```

### Fix 3: Snapshots taken

Snapshots of all sites were taken before any fixes, labelled "Pre nav-fix snapshot - good content before haiku downgrade". These can be used to recover content quality if the haiku model switch degrades output.

---

## What Still Needs Doing (Code)

### 1. Plan_sections filter (prevents recurrence)

In `PlanSectionsAction` (used by `page-build-handler`), filter out section names that match site-level component patterns before passing to the content writer.

```go
func filterSiteLevelSections(sections []string) []string {
    filtered := make([]string, 0, len(sections))
    for _, s := range sections {
        lower := strings.ToLower(s)
        if strings.Contains(lower, "header") || 
           strings.Contains(lower, "footer") ||
           lower == "site-header" || 
           lower == "site-footer" {
            continue
        }
        filtered = append(filtered, s)
    }
    return filtered
}
```

File: `platform/orchestration/actions/` — wherever `PlanSectionsAction` is defined.

### 2. InjectHeader/InjectFooter skip-if-present guard

Instead of always stripping and re-injecting, check if the assembled HTML already contains a site-level header before injecting. This handles edge cases where a legitimate design includes nav in section content.

```go
// In InjectHeader, before injecting:
if strings.Contains(html, `class="site-header`) {
    logger.Info("InjectHeader: Page already contains site-header, skipping")
    return html
}
```

Same pattern for `InjectFooter` with `class="site-footer"`.

File: `platform/orchestration/actions/component_library.go`

### 3. Component-template-fixer idempotency (problem 10)

The responsive CSS check uses `strings.Contains(html, "responsive fix")` but the injected comment has uppercase R: `/* Responsive fix — ...*/`. Fix: use `strings.Contains(strings.ToLower(html), "responsive fix")`.

File: `platform/orchestration/actions/` — `fixInjectResponsiveCSS` function.

### 4. Trigger rerender of affected pages

After Go fixes are deployed, trigger rerenders for ai-agent-orchestration.com and finetuning.uk to deploy clean single-header pages:

```sql
INSERT INTO site_work_items (site_id, item_type, status, handler_agent, summary, priority, pipeline, source)
SELECT s.id, 'needs_rerender', 'triaged', 'rerender-site-pages', 
       'Rerender all pages - post nav-fix cleanup', 5, 'build', 'manual'
FROM sites s WHERE s.domain IN ('ai-agent-orchestration.com', 'finetuning.uk');
```

---

## Architectural Gap: Site Design Planner (Option B)

### The problem

No agent makes site-level design decisions about navigation, layout, or component relationships. The current system assumes every site has a horizontal top header, full-width stacked sections, hamburger mobile menu. These are hardcoded in templates and Go code, not design decisions.

### What nobody decides today

- **Navigation architecture** — horizontal top, vertical sidebar, split nav, megamenu, transparent overlay, bottom tab bar
- **Nav content strategy** — which pages should be prominent, CTA wording, tool grouping, label optimisation for audience
- **Page layout patterns** — full-width stacked vs sidebar vs two-column vs asymmetric
- **Header/hero relationship** — separate vs merged (some designs overlay nav on hero)
- **Mobile navigation** — hamburger, bottom tab, slide-out, simplified items
- **Footer strategy** — minimal vs comprehensive, column count, what goes where
- **Logo treatment** — size, position, light/dark variants, text vs image vs combined

### Proposed: site-design-planner agent

A new agent that runs after the classifier and before page planning. It reads `design_intent`, `identity`, and competitor research, then outputs two new `site_specs` aspects:

**`navigation` aspect:**
```json
{
    "architecture": "horizontal-top",
    "primary_items": ["Home", "Services", "Case Studies", "About", "Blog", "Contact"],
    "tools_strategy": "grouped_under_tools_page",
    "cta": {"label": "Discuss Your Architecture", "url": "/contact.html"},
    "mobile": "hamburger-slide",
    "max_visible_items": 6,
    "sticky": true,
    "logo_position": "left",
    "legal_in_footer_only": true
}
```

**`layout` aspect:**
```json
{
    "default_page_layout": "full-width-stacked",
    "header_style": "dark-professional",
    "footer_style": "4-column",
    "hero_nav_merged": false,
    "sidebar_pages": [],
    "page_overrides": {
        "docs": {"layout": "sidebar-left", "nav": "sidebar-toc"}
    }
}
```

### What reads these specs

| Consumer | What it reads | What changes |
|---|---|---|
| `populate_nav_tables` | `navigation.primary_items`, `tools_strategy`, `max_visible_items` | Stops being purely mechanical, respects design intent |
| `render_site_components` | `layout.header_style`, `layout.footer_style` | Template selection driven by spec, not just style collection |
| Header template rendering | `navigation.cta`, `navigation.logo_position`, `navigation.sticky` | CTA, logo, sticky behaviour come from spec |
| `AssembleMultipageSiteAction` | `layout.default_page_layout`, `layout.page_overrides` | Can produce sidebar layouts, not just full-width |
| Mobile CSS injection | `navigation.mobile` | Different mobile patterns per design |
| `InjectHeader` / `InjectFooter` | `layout.hero_nav_merged`, `layout.page_overrides` | Knows when to skip injection for specific pages |

### Implementation sequence

1. Define the `navigation` and `layout` spec schemas
2. Create `site-design-planner` agent definition with LLM prompt
3. Wire it into the build pipeline after classifier, before page planning
4. Update `populate_nav_tables` to read `navigation` spec (with fallback to current mechanical behaviour)
5. Update header/footer template selection to read `layout` spec
6. Add `navigation` spec to classifier's output for adopted sites (reads existing nav structure)
7. Add discovery check: nav spec vs actual nav drift

### Where it fits in the build flow

```
intake-orchestrator
  → site-classifier (writes identity, design_intent, classification)
  → site-design-planner (NEW — writes navigation, layout)
  → chief-strategist / build-site-planner (writes site_plan with pages)
  → page-content-writer (renders sections)
  → assembly + deploy
```

---

## Key Files

| Area | File/Location |
|---|---|
| InjectHeader / InjectFooter | `platform/orchestration/actions/component_library.go` |
| rerenderStripWrapper | `platform/orchestration/actions/` (rerender actions file) |
| rerenderSinglePage | Same file |
| PlanSectionsAction | `platform/orchestration/actions/` |
| SavePageSectionsAction | `platform/orchestration/actions/multipage_actions.go` (or similar) |
| PopulateNavTablesAction | `platform/orchestration/actions/` (nav actions file) |
| fixInjectResponsiveCSS | `platform/orchestration/actions/` (component template fixer) |
| GetNavItems | `platform/orchestration/actions/component_library.go` |
| page-build-handler workflow | `agent_definitions` table, type = 'page-build-handler' |
| page-content-writer workflow | `agent_definitions` table, type = 'page-content-writer' |
| Header template | `content_components` table, name = 'header-professional-dark' |

## Diagnostic Queries

### Check for header/footer in pages.sections (should return 0 rows after fix)
```sql
SELECT s.domain, p.name, p.sections
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.sections::text LIKE '%header%' OR p.sections::text LIKE '%footer%';
```

### Check for header/footer in page_components (should return 0 rows after fix)
```sql
SELECT s.domain, p.name, pc.slot_name, pc.updated_at
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
JOIN sites s ON s.id = p.site_id
WHERE pc.slot_name LIKE '%header%' OR pc.slot_name LIKE '%footer%'
   OR pc.slot_name = 'site-header' OR pc.slot_name = 'site-footer';
```

### Check component-template-fixer duplicate CSS
```sql
SELECT s.domain,
    (LENGTH(sc.rendered_html) - LENGTH(REPLACE(sc.rendered_html, 'Responsive fix', ''))) 
        / LENGTH('Responsive fix') as responsive_fix_count
FROM site_components sc
JOIN sites s ON s.id = sc.site_id
WHERE sc.slot_name = 'header'
  AND sc.rendered_html LIKE '%Responsive fix%';
```

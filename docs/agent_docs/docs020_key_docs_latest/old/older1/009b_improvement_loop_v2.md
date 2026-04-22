# 009 — Improvement Loop

Post-build quality improvement cycle. Runs discovery agents (algorithmic), audit agents (LLM-based), triages findings, dispatches fixes, and rerenders.

---

## Flow

```
improvement-loop (orchestrator)
  │
  ├── 1. ensure_site_record
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
  │
  ├── 5. design-audit-agent (LLM-based orchestrator)
  │     ├── visual-design-auditor
  │     │     algorithmic checks → one LLM call → write_audit_findings
  │     │     checks: colour consistency, spacing, typography, dark sections, responsive
  │     └── content-quality-auditor
  │           load brief + page samples → one LLM call → write_audit_findings
  │           checks: tone alignment, content gaps, CTA effectiveness, differentiation
  │
  ├── 6. site-review-agent (LLM-based orchestrator)
  │     ├── content-quality-auditor (reused)
  │     └── strategic alignment review (own LLM call)
  │           checks: purpose alignment, page structure, dream spec gaps, conversion path
  │
  ├── 7. triage_detected_items → promote detected → triaged
  │
  ├── 8. check_has_findings → if none, complete_clean
  │
  ├── 9. insert needs_rerender (priority 99, runs after all fixes)
  │
  └── 10. build-dispatch-loop → processes all triaged items → rerender
```

**Ordering is deliberate:** Structural checks (2-4) run first — they're cheap, fast, and fix issues that would confuse the LLM audits. LLM audits (5-6) run after and see corrected HTML. Triage (7) catches everything. Dispatch (10) processes items by priority.

---

## Discovery Agents (Algorithmic)

These run Go discovery checks registered via `init()`. Each check queries the database, finds issues, and creates `site_work_items` with `status: detected`. No LLM calls.

### quality-discovery-agent

| Check | Detects | Handler |
|-------|---------|---------|
| `broken_nav_links` | Nav items pointing to non-existent pages | `nav-link-fixer` |
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

---

## Audit Agents (LLM-Based)

These load site context and specs, run algorithmic pre-checks, then make one LLM call per group for subjective assessment. Findings are written via `write_audit_findings` action with dedup and blocked-item filtering.

### design-audit-agent (orchestrator)

Spawns two group auditors:

**visual-design-auditor:**
- Loads: style collection, CSS theme, colour palette, rendered HTML samples
- Algorithmic checks: unlinked components count, slot mismatches, nav stacked, dark sections missing contract
- LLM checks: colour consistency, spacing regularity, typography hierarchy, responsive issues
- Findings routed to: `webdesign-agent`, `color-variable-fixer`, `component-template-fixer`

**content-quality-auditor:**
- Loads: site specs (identity, content_direction), page content samples, empty pages list
- LLM checks: tone alignment with spec, content gaps, CTA effectiveness, differentiation, audience targeting
- Findings routed to: `page-build-handler`, `component-template-fixer`

### site-review-agent (orchestrator)

Spawns content-quality-auditor (reused), then runs its own strategic review:
- Loads: site specs, dream spec (if exists), deployed page count, content audit results
- LLM checks: overall purpose alignment, page structure gaps, biggest improvements, conversion path
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
| `rerender-pages` | `needs_rerender` | Re-assembles pages from stored components, deploys |

---

## Colour Fix Detail

```
hardcoded_section_colors:
  discovery: countHardcodedColorComponents() counts page_components with inline hex
  fix: fix_hardcoded_colors replaces hex with CSS variables
  both templates (permanent) and rendered HTML (immediate) are fixed

forced_text_colors:
  discovery: findForcedTextColors() parses <style> blocks for child text rules
  only flags: h1-h6, p, li, span, td, th, dt, dd, blockquote, figcaption with color: #hex
  skips: container rules, link rules, dark section container declarations
  fix: removes child text color declarations
  safety: loads site palette, calculates resulting text colour, WCAG AA check (4.5:1)
  for dark sections missing --section-* contract: injects contract first, then removes colours
```

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

Per-site audit config in `sites.settings.maintenance_profile.audit`:

```json
{
    "audit": {
        "visual_design": { "enabled": true, "every": "7d" },
        "content_quality": { "enabled": true, "every": "7d" },
        "strategic_review": { "enabled": true, "every": "30d" }
    }
}
```

Currently all audits run every cycle. Per-site config filtering is a future enhancement.

---

## Timeout

The improvement loop timeout is 1800 seconds (30 minutes). Typical run:
- Algorithmic checks: 10-30 seconds total
- Design audit (2 LLM calls): 30-60 seconds
- Site review (2 LLM calls): 30-60 seconds
- Dispatch + fixes: varies by finding count
- Rerender: 30-60 seconds

Total: 3-8 minutes for a typical site with a few findings.


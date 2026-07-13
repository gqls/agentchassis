# 004 — Improvement Loop

Post-build quality improvement cycle. Runs discovery agents (algorithmic), audit agents (LLM-based), triages findings, dispatches fixes, and rerenders.

**v4 changes (2026-03-31):** Orphan page detection, blog listing rebuild in rerender pipeline, content writer chrome fix, lock types with expiry, audit pass auto-reset, HITL direction integration, adopted site handling.

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
  │     └── orphan_pages (detects unreachable deployed pages)
  │
  ├── 5. design-audit-agent (LLM-based orchestrator)
  │     ├── visual-design-auditor
  │     │     algorithmic checks → one LLM call (TOP 5 findings) → write_audit_findings
  │     │     checks: colour consistency, spacing, typography, dark sections, responsive
  │     │     data queries EXCLUDE locked components (locked_at IS NULL or expired)
  │     └── content-quality-auditor
  │           load brief + direction + page samples (excluding locked) → one LLM call (TOP 5) → write_audit_findings
  │           checks: tone alignment, content gaps, CTA effectiveness, differentiation
  │           RESPECTS direction spec must_have features
  │
  ├── 6. site-review-agent (LLM-based orchestrator)
  │     ├── content-quality-auditor (reused)
  │     └── strategic alignment review (own LLM call, TOP 5 findings)
  │           checks: purpose alignment, page structure, dream spec gaps, conversion path
  │           READS direction spec for human-requested features
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

---

## Adopted Sites and the Improvement Loop

Sites created by the adoption agent enter the improvement loop the same way as any other site. The key difference is the strategist's first run.

### First improvement pass on an adopted site

The identity spec contains `adopted_from` and `adopted_at` fields. When the strategist runs for the first time on an adopted site:

1. It reads the identity spec (factual: "this company does X, the tone is Y")
2. It reads the existing page structure (adopted pages)
3. It writes an aspirational `content_direction` spec that goes BEYOND the adopted state
4. The improvement loop then has room to push toward the aspiration

The strategist must not be constrained by the status quo. If the adopted site has 5 pages, the strategist might recommend 12. The prompt for the strategist explicitly instructs: "This site was adopted from an existing domain. The current state is a baseline, not a goal. Consider what's missing, what competitors do better, what opportunities the current site doesn't exploit."

### Adopted content in audits

Content built with `mode: "recreate"` (preserving original text) is treated normally by auditors. The auditors don't know or care that content was adopted — they evaluate it against the identity spec and content_direction like any other content. If adopted content is poor quality, the auditor flags it for improvement. This is intentional — adoption is a starting point, and the improvement loop pushes the site forward.

---

## Triage Drain Controls

### 1. Finding Cap (migration 084)

Each LLM auditor reports UP TO 5 findings per pass. Prompts include:
- "Report ONLY the TOP 5 most impactful issues"
- "Do NOT report issues already caught by algorithmic checks"
- "Do NOT flag features listed as must_have in the direction spec"
- Each finding requires: `current_value`, `acceptance_test`, `suggestion`, `max_fix_attempts`

### 2. Section Locking

Components that pass verification get locked. Locked components are excluded from audit data-loading queries.

**Lock types (v4):**

| Lock type | Behaviour | Use case |
|-----------|-----------|----------|
| `permanent` | Never expires, manual unlock only | Brand elements, legal disclaimers, human-crafted content |
| `timed` | Expires after N days (configurable, default 90) | HITL-requested content that should eventually re-enter improvement cycle |
| `review` | Creates HITL review item on expiry | Content needing human approval before agents touch it again |

Implementation: `lock_type` and `lock_expires_at` columns on `page_components` and `site_components`.

Discovery check queries expand from `AND locked_at IS NULL` to:
```sql
AND (locked_at IS NULL OR (lock_expires_at IS NOT NULL AND lock_expires_at < NOW()))
```

When a `review` lock expires, a discovery check creates a `needs_lock_review` work item → HITL decides: re-lock, release, or update. If no human response within configurable period, system can auto-release (per-site config).

### 3. Audit Pass Cap

Sites track `audit_pass_count` in `sites.settings.maintenance_profile`.

```
Pass 1: top 5 findings per auditor → fix → verify → increment
Pass 2: top 5 findings → fix → verify → increment
Pass 3: top 5 findings → fix or accept → increment
Pass 4+: improvement-loop exits immediately
```

**Auto-reset (v4):**

| Trigger | Mechanism |
|---------|-----------|
| Time-based (default) | Improvement sweep checks `last_audit_reset_at`, resets after 60 days |
| Direction change | Human updates direction spec → auto-reset |
| Major rebuild | N pages rebuilt in one cycle → auto-reset |
| Manual | Human clicks "re-audit" in dashboard |

Combined with timed lock expiry, this creates a natural rhythm:

```
Build → audit × 3 → cap reached → site quiet
  ... 60 days ...
  → pass counter resets, expired locks release
  → improvement loop runs fresh
  → finds new issues (content aged, design dated, new patterns available)
  → audit × 3 → quiet again
```

Sites breathe and improve rather than constant churn or permanent stasis. HITL involvement optional.

### Combined Effect

```
Before (per domain):  ~8+ audit passes, ~50+ findings per pass = ~88K+ tokens
After (per domain):   3 audit passes, ~15 findings per pass + verification = ~30K tokens
                      Auto-reset every 60d for ongoing improvement
```

---

## Human Direction Integration (v4)

### How direction reaches the improvement loop

The `direction` spec (pinned, human-managed) influences the improvement loop at multiple points:

**Content-quality-auditor:** Loads direction spec alongside identity and content_direction. Features marked `must_have` or `should_have` are not flagged for removal. The prompt includes: "The following features were specifically requested and should not be removed: [list]."

**Strategic review:** Reads direction spec. Evaluates whether the site is moving toward the requested goals. May suggest additions aligned with direction but does not suggest removing requested features.

**Discovery checks:** Not affected — these are algorithmic and check structural issues, not strategic alignment.

### Direction spec lifetime

Direction persists until the human changes it. The strategist can suggest refinements (adding `nice_to_have` items) but cannot remove `must_have` items. Direction changes trigger an audit pass reset so the improvement loop can re-evaluate against new goals.

---

## Discovery Agents (Algorithmic)

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

---

## Blog Listing Rebuild

The `rebuild_blog_listing` action runs as part of the `rerender-pages` workflow, before `get_pages`:

```
rerender-pages workflow:
  check_refresh_components
    → (true) render_site_components → rebuild_blog_listing
    → (false) rebuild_blog_listing
  rebuild_blog_listing → get_pages → check_pages_exist → create_rerender_items → complete
```

Queries deployed blog-post pages, loads template from content_components (`blog-listing` or `article_grid` fallback), renders with post data, upserts page_component. No LLM, no hardcoded styles. Different style collections can have different listing layouts.

---

## Audit Agents (LLM-Based)

Load site context, specs, and direction (excluding locked/unexpired components). One LLM call per group, UP TO 5 structured findings with acceptance criteria.

### design-audit-agent (orchestrator)

**visual-design-auditor:**
- Loads: style collection, CSS theme, colour palette, rendered HTML samples (locked/unexpired excluded)
- LLM checks (TOP 5): colour consistency, spacing, typography, responsive
- Respects direction spec — doesn't flag requested visual features

**content-quality-auditor:**
- Loads: site specs (identity, content_direction, direction), page samples (locked/unexpired excluded)
- LLM checks (TOP 5): tone alignment, content gaps, CTA effectiveness, differentiation
- Respects direction must_have features

### site-review-agent (orchestrator)

- Strategic review (TOP 5): purpose alignment, page structure, conversion path
- Reads direction spec — evaluates progress toward human-requested goals

---

## Fix Agents (Handlers)

| Handler | Fixes | How |
|---------|-------|-----|
| `color-variable-fixer` | `hardcoded_section_colors`, `forced_text_colors` | Replaces hex with CSS variables |
| `webdesign-agent` | `needs_design_review`, `missing_css`, `generic_theme` | Regenerates CSS from design spec |
| `page-build-handler` | `needs_content_page`, `content_rewrite`, `tone_shift` | Wraps page-content-writer with persistence |
| `component-template-fixer` | `cta_improvement`, `spacing_fix`, `responsive_fix` | CSS injection, element modification |
| `site-component-linker` | `unlinked_site_component`, `header_footer_fix` | Links site_components to style collection templates |
| `asset-deployer` | `undeployed_asset` | Downloads from S3, optimises, commits to git |
| `blog-content-planner` | `needs_blog_posts` | Plans blog posts, creates pages + work items |
| `rerender-pages` | `needs_rerender`, `orphan_blog_posts` | Rebuilds blog listing, re-assembles pages, deploys |
| `content-gap-planner` | `orphan_page` | Adds nav entry or internal links to unreachable page |

---

## Triggering

| Trigger | How |
|---------|-----|
| Manual | `./trigger-audit.sh improvement-loop <site_id> <domain>` |
| Post-build | Side-effect after last build item completes |
| Post-adoption | Dispatch loop processes adoption work items → rerender → improvement loop |
| Scheduled | `improvement-sweep` every 600s, finds least-recently-built site |
| Auto-reset | Sweep resets pass counter on sites older than 60 days since last reset |
| Direction change | Human updates direction spec → pass counter resets → next sweep picks it up |

---

## Timeout

1800 seconds (30 minutes). Typical run: 3-8 minutes with findings. < 1 second for sites at pass limit.

---

## Entry Points and Interaction (from 009_build_pipeline_vs_improvement_loop)

Three entry points exist:

**build-pipeline-trigger** (scheduled, every 120s): finds sites with triaged items, fires build-dispatch-loop per site. Only processes existing triaged items — does NOT create new items. This is the worker.

**improvement-loop** (scheduled via improvement-sweep, every 600s): finds least-recently-audited site with queue depth < 20. Runs ALL discovery + audit agents, creates items, triages them, then calls build-dispatch-loop internally.

**Manual trigger** (`./trigger-audit.sh`): fires improvement-loop or individual agents for a specific site.

### How they interact

The build-pipeline-trigger and improvement-loop don't conflict:

| | build-pipeline-trigger | improvement-loop |
|---|---|---|
| Creates items? | No | Yes |
| Runs audits? | No | Yes |
| Calls dispatch? | Yes — directly | Yes — internally at step 10 |
| Concurrency risk | Claims atomically | Sweep pre-query checks queue depth |

If both fire simultaneously, `claim_work_item` uses atomic UPDATE — only one succeeds per item.

### Gap: manual discovery without full loop

If you manually trigger just a discovery agent (not the full improvement-loop), items stay as `detected` until the next improvement-loop triage step promotes them. The build-pipeline-trigger ignores `detected` items. Either run the full improvement-loop or manually promote: `UPDATE site_work_items SET status = 'triaged' WHERE status = 'detected'`.

---

## Component Standards Validation (from 002_component_standards_validation)

The `validate_component_standards` discovery check audits all components used by a site. Runs as part of the improvement loop's `run_discovery_checks` action.

### Sub-checks

| Check | Detects | Handler |
|---|---|---|
| `checkUnlinkedSiteComponents` | header/footer/head with NULL component_id | `site-component-linker` |
| `checkMissingDataComponent` | rendered_html without `data-component` attribute | `component-template-fixer` |
| `checkSlotNameMismatch` | slot_name != data-component attribute | `component-template-fixer` |
| `checkMissingSiteMetadata` | empty company_name, tagline, logo_url, email | `site-metadata-fixer` |
| `checkMissingAssetRefs` | logo_url set but header has no `<img>` tag | `site-component-linker` |
| `checkNavLayout` | nav `<ul>` with no flex CSS | `component-template-fixer` |
| `checkUnwantedElements` | search icon in header when no search feature | `component-template-fixer` |
| `checkDarkSectionContract` | dark sections missing CSS variables | existing check |
| `checkEmptyPageSections` | deployed pages with no rendered components | `page-content-writer` |

### Ordering

`validate_component_standards` runs BEFORE design checks (hardcoded_colors, forced_text_colors) because those operate on rendered HTML which may be wrong if components aren't linked. Fix structural issues first.

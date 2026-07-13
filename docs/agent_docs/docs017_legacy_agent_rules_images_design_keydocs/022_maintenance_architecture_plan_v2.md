# Site Maintenance Architecture — Plan

## Overview

Maintenance is fundamentally different from build. During build, work flows in one direction: plan → generate → review → deploy. During maintenance, problems arrive from everywhere — the world changes, content ages, links break, competitors evolve, regulations update, clients request changes — and each needs a different detection method, urgency level, and resolution path.

The maintenance system uses three layers, loosely coupled through shared database tables:

```
Discovery Agents → maintenance_findings table → Triage → Fix Agents
```

Discovery agents find problems. Triage assesses severity and resolution path. Fix agents execute changes. Fix agents are purpose-built for narrow, targeted changes — they are separate from the build-phase agents because build agents are designed for "create from nothing" (broad scope, full brief) while fix agents are designed for "change this specific thing because of this specific reason" (narrow scope, finding-driven, minimal blast radius).

Not every discovery type has an automated fix agent. Some findings need human decisions.

---

## Spawn Chain

The maintenance system follows the same spawning pattern as the vet-batch-processor → vet-practice-verifier pipeline. Agents are dynamically spawned, not persistently running.

```
K8s CronJob (every 8 hours)
  │  publishes Kafka message to agent-chassis process topic
  │
  └─► agent-chassis (persistent, always listening)
        │  spawns maintenance-batch-scheduler as K8s job
        │
        └─► maintenance-batch-scheduler
              │  1. populate_maintenance_queue  → query sites, insert due tasks
              │  2. load_maintenance_batch      → claim N tasks (FOR UPDATE SKIP LOCKED)
              │  3. group by site
              │  4. for each site group:
              │
              └─► site-maintenance-orchestrator (one per site, K8s job)
                    │  1. process_pending_fixes   → fix items from previous cycles
                    │  2. verify_previous_fixes   → check previous fixes worked
                    │  3. run_discovery            → spawn discovery agents for due domains
                    │  4. triage_new_findings      → enrich and classify new findings
                    │  5. update_last_run          → update cadence timestamps
                    │
                    ├─► content-discovery-agent    → scans pages, writes findings
                    ├─► links-discovery-agent      → checks links, writes findings
                    ├─► seo-discovery-agent        → validates SEO, writes findings
                    ├─► compliance-discovery-agent  → checks legal/regulatory, writes findings
                    ├─► structural-discovery-agent  → analyses structure, writes findings
                    │
                    ├─► section-rewriter           → rewrites stale content sections
                    ├─► redirect-manager           → fixes redirect chains
                    ├─► sitemap-regenerator         → rebuilds sitemap.xml
                    └─► (other fix agents as needed)


K8s CronJob (daily, separate)
  └─► maintenance-catch-all
        → cleans up stale findings
        → sends HITL reminders
        → reclassifies unclaimed findings
        → detects cross-site patterns
```

Each site gets its own orchestrator instance. The orchestrator reads that site's maintenance profile from `sites.settings` and spawns only the discovery agents that site's profile requires. Different sites get different combinations of agents — a finance site might need regulatory compliance checking while a brochure site only needs content freshness and link validation.

The batch scheduler controls concurrency through `batch_size`, limiting how many site orchestrators run simultaneously. This prevents flooding the cluster when thousands of sites are due.

---

## Precedent: The Vet Batch Pattern

The maintenance spawn chain directly mirrors the existing vet data collection pipeline:

```
Vet pattern:
  vet-batch-processor (orchestrator)
    1. load_business_batch  → claim N pending tasks from collection_tasks
                               (FOR UPDATE SKIP LOCKED — safe concurrent claiming)
    2. check_batch          → any work?
    3. spawn_verifier       → spawn vet-practice-verifier agent
    4. process_batch        → loop: call verifier per business
    5. complete

Maintenance pattern:
  maintenance-batch-scheduler (orchestrator)
    1. populate_maintenance_queue → query sites, insert due tasks into maintenance_tasks
    2. load_maintenance_batch    → claim N pending tasks
                                    (FOR UPDATE SKIP LOCKED)
    3. check_batch               → any work?
    4. group by site             → combine tasks for same site
    5. process_batch             → loop: spawn site-maintenance-orchestrator per site
    6. complete
```

Key difference: the vet processor calls the verifier once per business. The maintenance scheduler groups tasks by site and calls the site orchestrator once per site, passing all the domains that are due. If site X has both content and links due, it gets one orchestrator invocation with both domains — maintaining the "site is king" principle.

---

## Database Tables

### maintenance_tasks — Work Queue

Same pattern as `business_intel.collection_tasks`. The batch scheduler populates this from site profiles, then claims items for processing.

```sql
CREATE TABLE maintenance_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    task_type TEXT NOT NULL,              -- 'discovery', 'triage', 'fix'
    domain TEXT NOT NULL,                 -- 'content', 'links', 'seo', 'compliance', 'structural', 'design'
    status TEXT NOT NULL DEFAULT 'pending',  -- pending, in_progress, completed, failed
    priority INTEGER DEFAULT 100,
    scheduled_for TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    orchestration_id UUID,               -- which agent run claimed this
    result JSONB,                        -- summary of what happened
    error TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_mt_pending ON maintenance_tasks(status, priority, scheduled_for)
    WHERE status = 'pending';
CREATE INDEX idx_mt_site ON maintenance_tasks(site_id, domain);
```

The scheduler populates this by querying sites with due maintenance:

```sql
-- Example: find sites where content maintenance is due
SELECT s.id as site_id, 'content' as domain
FROM sites s
WHERE s.status = 'active'
  AND s.settings->'maintenance_profile'->'content'->>'enabled' = 'true'
  AND (
    s.settings->'maintenance_profile'->'content'->>'last_run_at' IS NULL
    OR (s.settings->'maintenance_profile'->'content'->>'last_run_at')::timestamptz
       + (s.settings->'maintenance_profile'->'content'->>'every')::interval < NOW()
  )
```

### maintenance_findings — Coordination Point

Every discovery agent writes here. Triage reads and enriches. Fix agents read triaged items and update status after acting.

```sql
CREATE TABLE maintenance_findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),

    -- What was found
    domain TEXT NOT NULL,                    -- 'content', 'links', 'seo', 'legal', 'design',
                                             -- 'structural', 'compliance', 'technical'
    finding_type TEXT NOT NULL,              -- 'stale_date_reference', 'broken_external_link',
                                             -- 'missing_competitor_feature', etc.
    severity TEXT NOT NULL DEFAULT 'low',    -- 'info', 'low', 'medium', 'high', 'urgent'
    summary TEXT NOT NULL,                   -- human-readable description
    detail JSONB DEFAULT '{}',              -- finding-specific structured data

    -- What it affects
    page_id UUID REFERENCES pages(id),       -- NULL for site-wide findings
    component_id UUID,                       -- specific page_component if applicable
    entity_id UUID,                          -- specific entity if applicable
    affected_url TEXT,                       -- the URL or link in question

    -- Triage enrichment (filled by triage step)
    impact JSONB,                            -- inbound links, nav membership, SEO value, traffic
    resolution_path TEXT,                    -- 'auto_fix', 'suggest', 'flag', 'monitor', 'ignore'
    suggested_action TEXT,                   -- 'rewrite_section', 'remove_page', 'add_redirect',
                                             -- 'update_disclaimer', etc.
    suggested_detail JSONB,                  -- action-specific data (e.g. proposed rewrite, redirect target)
    priority INTEGER,                        -- computed from severity + impact
    fix_agent TEXT,                          -- which fix agent should handle this
                                             -- (set during triage routing)

    -- Lifecycle
    status TEXT NOT NULL DEFAULT 'detected', -- 'detected', 'triaged', 'approved', 'in_progress',
                                             -- 'fixed', 'fixed_pending_verify', 'verified',
                                             -- 'rejected', 'wont_fix'
    discovered_by TEXT NOT NULL,             -- agent type that found it
    fixed_by TEXT,                           -- agent type that fixed it
    approved_by TEXT,                        -- 'auto' or user identifier

    -- Cross-domain references
    parent_finding_id UUID REFERENCES maintenance_findings(id),
                                             -- if created as side-effect of fixing another finding
    related_finding_ids UUID[],              -- findings that should be considered together

    -- Deduplication
    finding_key TEXT,                        -- deterministic key for dedup
                                             -- e.g. 'stale_date:page_id:component_id'

    created_at TIMESTAMPTZ DEFAULT NOW(),
    triaged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,

    UNIQUE(site_id, finding_key)             -- prevents duplicate findings
);

CREATE INDEX idx_mf_site_status ON maintenance_findings(site_id, status);
CREATE INDEX idx_mf_domain ON maintenance_findings(domain, status);
CREATE INDEX idx_mf_page ON maintenance_findings(page_id);
CREATE INDEX idx_mf_priority ON maintenance_findings(priority DESC NULLS LAST)
    WHERE status = 'triaged';
CREATE INDEX idx_mf_fix_agent ON maintenance_findings(fix_agent, status)
    WHERE status IN ('triaged', 'approved');
```

---

## Agent Definitions

### maintenance-batch-scheduler

The top-level orchestrator. Triggered by CronJob → agent-chassis. Populates the work queue, claims a batch, spawns site orchestrators.

```
Workflow:
  1. populate_maintenance_queue   → query sites with due maintenance, insert into maintenance_tasks
  2. load_maintenance_batch       → claim N pending tasks (FOR UPDATE SKIP LOCKED)
                                     batch_size controls concurrency
  3. check_batch                  → conditional: any work? if not → complete_empty
  4. spawn_site_maintainer        → spawn site-maintenance-orchestrator agent
  5. process_batch                → loop through claimed tasks, grouped by site_id
     └── per site: call_agent(site-maintenance-orchestrator)
                   with { site_id, domains: ['content', 'links', ...] }
  6. complete

Input:  { batch_size: 10 }  (configurable)
Output: { tasks_claimed, sites_processed, summary }
```

### site-maintenance-orchestrator

The per-site agent. Receives a site_id and list of domains due for maintenance. Runs the full find → triage → fix → verify cycle.

```
Workflow:
  1. load_site_context           → load site record, maintenance profile,
                                    current maintenance_findings summary

  2. process_pending_fixes       → query maintenance_findings for this site where:
                                    status = 'triaged' AND resolution_path = 'auto_fix'
                                    OR status = 'approved'
                                  → for each: route to appropriate fix agent
                                    (see Fix Agents section below)
                                  → update finding status to 'in_progress' then
                                    'fixed_pending_verify'

  3. verify_previous_fixes       → query findings with status = 'fixed_pending_verify'
                                  → re-scan the specific page/component to confirm fix
                                  → update to 'verified' or 'fix_failed'

  4. run_discovery               → for each domain in input domains list:
                                    check site's maintenance profile for which
                                    sub-agents are enabled
                                    spawn the appropriate discovery agent
                                    collect results

  5. triage_new_findings         → query findings with status = 'detected' for this site
                                  → cross-reference impact (read link_registry,
                                    site_nav_items, pages tables)
                                  → classify resolution_path
                                  → assign fix_agent
                                  → score priority
                                  → update findings to 'triaged'

  6. update_last_run             → update sites.settings maintenance_profile
                                    last_run_at for each domain processed

  7. complete                    → return summary: findings created, fixes applied,
                                    fixes verified, items awaiting human review

Input:  { site_id, domains: ['content', 'links'] }
Output: { findings_created, fixes_applied, fixes_verified, pending_human_review }
```

The order is deliberate: fix first, then verify, then discover, then triage. Each cycle processes the backlog before adding to it. This prevents the queue from growing unboundedly if fixing is slower than discovering.

---

## Discovery Agents

Each discovery agent is spawned by the site-maintenance-orchestrator as a K8s job. They have narrow focus — scan a specific aspect of the site and write findings to maintenance_findings. They do not triage or fix.

### content-discovery-agent

```
Workflow:
  1. load_pages             → load all deployed pages + page_components for this site
  2. scan_date_refs         → regex scan for time-sensitive patterns
                               ("in 2024", "last year", "since 2023", specific dates)
                               write findings of type 'stale_date_reference'
  3. scan_entity_drift      → compare page_components content against site_entities data
                               where entity data has changed but page content hasn't
                               write findings of type 'entity_data_drift'
  4. scan_thin_content      → check section word counts against thresholds
                               write findings of type 'thin_content'
  5. scan_statistics        → pattern match for statistical claims that may be outdated
                               ("over 500 clients", "97% satisfaction rate")
                               write findings of type 'stale_statistic'
  6. complete               → return { findings_count, pages_scanned }

Input:  { site_id, enabled_checks: ['date_refs', 'entity_drift', 'thin_content'] }
Output: { findings_count, pages_scanned }
```

Which sub-checks run is controlled by the site's maintenance profile `agents` config.

### links-discovery-agent

```
Workflow:
  1. load_link_registry     → load all links for this site from link_registry
  2. check_internal         → verify internal link targets exist in pages table
                               and have build_status = 'deployed'
                               write findings of type 'broken_internal_link'
  3. check_external         → HTTP HEAD on external links (rate-limited, batched)
                               write findings of type 'broken_external_link'
                               for 4xx/5xx responses
  4. check_orphans          → find pages with zero inbound internal links
                               write findings of type 'orphan_page'
  5. check_redirect_chains  → find redirect chains > 1 hop in redirect table
                               write findings of type 'redirect_chain'
  6. complete               → return { findings_count, links_checked }

Input:  { site_id, enabled_checks: [...], external_batch_size: 50 }
Output: { findings_count, links_checked, external_links_checked }
```

### seo-discovery-agent

```
Workflow:
  1. check_sitemap_sync     → compare sitemap XML against pages table
                               write findings for pages missing from sitemap
                               or sitemap entries for non-existent pages
  2. validate_schema        → check JSON-LD / structured data on each page
                               validate against schema.org
                               write findings of type 'invalid_schema'
  3. check_meta_freshness   → compare meta descriptions against current page content
                               (LLM similarity check if enabled)
                               write findings of type 'stale_meta_description'
  4. complete

Input:  { site_id, enabled_checks: [...] }
Output: { findings_count }
```

### compliance-discovery-agent

```
Workflow:
  1. check_disclaimers      → scan pages for required disclaimer text
                               based on site's industry + jurisdiction config
                               write findings of type 'missing_disclaimer'
  2. check_legal_versions   → compare deployed legal page content hashes
                               against current template hashes
                               write findings of type 'outdated_legal_template'
  3. check_tool_compliance  → for sites with tools/calculators:
                               check tool dependency config (regulatory body,
                               last_checked date, data dependencies)
                               write findings of type 'tool_compliance_risk'
  4. check_regulatory       → check regulatory body RSS/feeds for changes
                               that affect this site's industry/jurisdiction
                               write findings of type 'regulatory_change'
  5. complete

Input:  { site_id, enabled_checks: [...], regulatory_bodies: ['FCA', 'ICO'],
          jurisdictions: ['UK'] }
Output: { findings_count }
```

### structural-discovery-agent

```
Workflow:
  1. check_nav_complexity   → analyse site_nav_items counts and structure
                               flag if primary nav > threshold items
                               flag redundant or poorly grouped items
                               write findings of type 'nav_too_complex'
  2. detect_redundant       → LLM similarity comparison across pages
                               write findings of type 'redundant_content'
  3. competitor_analysis    → crawl competitor sites (if configured)
                               compare structural features
                               write findings of type 'missing_competitor_feature'
  4. complete

Input:  { site_id, enabled_checks: [...], competitors: ['competitor1.com'] }
Output: { findings_count }
```

### Discovery Agent Summary

| Domain | Agent | Checks | Method | Priority |
|--------|-------|--------|--------|----------|
| Content | `content-discovery-agent` | date refs, entity drift, thin content, stale stats | Regex + data comparison | Phase 0 |
| Links | `links-discovery-agent` | internal links, external links, orphans, redirect chains | DB queries + HTTP HEAD | Phase 0 |
| SEO | `seo-discovery-agent` | sitemap sync, schema, meta freshness | DB comparison + validation | Phase 0 |
| Compliance | `compliance-discovery-agent` | disclaimers, legal templates, tool compliance, regulatory | Template comparison + feeds | Phase 1 |
| Structural | `structural-discovery-agent` | nav complexity, redundant content, competitor gaps | DB analysis + LLM + crawl | Phase 3 |

---

## Triage

Triage runs as a step within the site-maintenance-orchestrator (step 5), not as a separate agent. It reads `detected` findings for this site and enriches them.

### What triage does

**1. Deduplication** — the `finding_key` unique constraint handles exact duplicates at insert time. Triage also checks for near-duplicates (same page, similar finding type, different cycle).

**2. Impact assessment** — cross-references other domains' tables to understand blast radius:

```sql
-- For a content finding about page X:

-- How many inbound internal links?
SELECT COUNT(*) FROM link_registry
WHERE target_page_id = $page_id AND link_type = 'internal';

-- Is it in navigation?
SELECT group_id, label FROM site_nav_items
WHERE page_id = $page_id AND status = 'active';

-- Page visibility
SELECT build_status, last_deployed_at FROM pages WHERE id = $page_id;

-- Entity references
SELECT entity_type, entity_key FROM site_entities WHERE page_id = $page_id;

-- Other findings about this page
SELECT finding_type, severity FROM maintenance_findings
WHERE page_id = $page_id AND status NOT IN ('resolved', 'ignored');
```

This is database reads only. The triage step reads other domains' tables but does not write to them.

**3. Resolution classification:**

| Path | Meaning | Examples |
|------|---------|---------|
| `auto_fix` | Safe to fix without human approval | Redirect chain shortening, sitemap regeneration, image format conversion |
| `suggest` | Propose a specific fix, await approval | "Rewrite this paragraph — here's the proposed text", "Add this disclaimer" |
| `flag` | Needs human discussion, no specific fix proposed | "Competitor has a blog — should we?", "This page's purpose is unclear" |
| `monitor` | Not actionable yet, track for next cycle | "Statistic is 8 months old — flag again at 12 months" |
| `ignore` | Valid finding but not worth acting on | "Date reference in a case study — dates are expected there" |

**4. Routing** — triage assigns the `fix_agent` field based on finding type and suggested action. This is important because the discovery agent that found the problem often isn't the agent that should fix it. Example: the links-discovery-agent finds "broken external link" but the fix might be "rewrite the paragraph" → routes to `section-rewriter`, not to `redirect-manager`.

| Finding Type | Suggested Action | Routes To (fix_agent) |
|-------------|-----------------|----------------------|
| `stale_date_reference` | `rewrite_section` | `section-rewriter` |
| `entity_data_drift` | `rewrite_section` | `section-rewriter` |
| `thin_content` | `rewrite_section` | `section-rewriter` |
| `broken_internal_link` | `update_link` or `add_redirect` | `redirect-manager` |
| `broken_external_link` | `rewrite_section` | `section-rewriter` |
| `broken_external_link` | `remove_link` | `redirect-manager` |
| `redirect_chain` | `shorten_chain` | `redirect-manager` |
| `orphan_page` | `add_nav_item` or `add_internal_links` | `nav-updater` or `section-rewriter` |
| `sitemap_out_of_sync` | `regenerate_sitemap` | `sitemap-regenerator` |
| `invalid_schema` | `regenerate_schema` | `schema-fixer` |
| `stale_meta_description` | `regenerate_meta` | `schema-fixer` |
| `missing_disclaimer` | `inject_disclaimer` | `legal-updater` |
| `outdated_legal_template` | `regenerate_legal_page` | `legal-updater` |
| `nav_too_complex` | `suggest_reorganisation` | — (human decision) |
| `missing_competitor_feature` | `suggest_new_page` | — (human decision) |
| `redundant_content` | `suggest_merge` | — (human decision) |
| `image_not_optimised` | `optimise_image` | `image-optimiser` |

**5. Priority scoring** — combines severity (from discovery) with impact (from cross-referencing):

```
priority = severity_weight × impact_multiplier

severity_weight:
  info=1, low=2, medium=4, high=8, urgent=16

impact_multiplier (cumulative):
  in_primary_nav = ×3
  has_inbound_links = ×2
  has_entity_references = ×2
  is_homepage = ×5
  recently_deployed = ×0.5
```

A medium-severity finding on the homepage with inbound links is higher priority than a high-severity finding on an orphan page.

### What triage does NOT do

- Does not fix anything
- Does not call other agents
- Does not write to other domains' tables
- Only reads other domains' tables and writes back to maintenance_findings

---

## Fix Agents

Fix agents are purpose-built for narrow, targeted changes. They are separate from the build-phase agents. Build agents are designed for "create from nothing" — they expect a full brief, generate entire pages, and deploy whole sites. Fix agents are designed for "change this specific thing because of this specific reason" — they receive a finding, understand why the change is needed, make the minimal change, and redeploy.

Each fix agent follows a common pattern:

```
1. load_finding_context    → load finding, affected page, affected component, site record
2. [agent-specific fix logic]
3. mark_finding_fixed      → update finding status to 'fixed_pending_verify'
4. check_side_effects      → if fix affects other domains, write cross-domain findings
5. complete
```

### section-rewriter

The primary fix agent. Handles any finding where the fix is rewriting content in a specific page section.

```
Workflow:
  1. load_finding_context      → finding, page, page_component, site record
  2. load_component_template   → content_component template for this section
  3. load_current_content      → current rendered HTML of this section
  4. generate_rewrite          → LLM prompt, tightly scoped:
                                  "Current content: [section HTML]
                                   Problem: [finding.summary]
                                   Reason: [finding.detail]
                                   Brand voice: [site brief excerpt]
                                   Task: Rewrite ONLY the affected text.
                                   Preserve structure, headings, valid links.
                                   Return updated content JSON matching the
                                   component's input_schema."
  5. render_section            → render_component with new content
  6. reassemble_page           → load all page_components for this page,
                                  replace the updated section,
                                  re-assemble full page HTML
                                  (header + sections + footer)
  7. commit_page               → git_commit the updated page
  8. update_finding            → status = 'fixed_pending_verify'
  9. check_side_effects        → did rewrite change any links?
                                  → write cross-domain findings if so
                                  (type: 'link_changed', domain: 'links')
  10. complete

Input:  { site_id, finding_id }
Output: { rewritten: true, component_id, page_name, commit_sha }
```

Step 6 (reassemble_page) works from stored page_components. We already store rendered sections in page_components via the `save_page_sections` action. The section-rewriter updates one page_component's rendered HTML, then calls an action that re-assembles the full page from all stored sections + header + footer. Same assembly logic as build, driven from stored sections rather than freshly generated ones.

**Handles finding types:** `stale_date_reference`, `entity_data_drift`, `thin_content`, `stale_statistic`, `broken_external_link` (when fix is rewording), `stale_meta_description` (content portion).

### redirect-manager

Algorithmic, no LLM needed.

```
Workflow:
  1. load_finding_context       → finding, affected links
  2. apply_fix:
     - redirect_chain:          → shorten A→B→C to A→C in redirect table
     - broken_internal_link:    → if redirect exists, update link_registry to new target
                                  if no redirect, flag for section-rewriter
     - broken_external_link:    → if resolution is 'remove_link', update link_registry
  3. update_affected_pages      → if links were changed in link_registry,
                                   identify pages containing those links
                                   (may trigger section-rewriter for link text changes)
  4. update_finding             → status = 'fixed_pending_verify'
  5. complete

Input:  { site_id, finding_id }
Output: { redirects_shortened, links_updated }
```

**Handles finding types:** `redirect_chain`, `broken_internal_link` (redirect cases), `broken_external_link` (removal cases).

### sitemap-regenerator

Algorithmic, no LLM needed.

```
Workflow:
  1. load_site_pages     → query all deployed pages for this site
  2. generate_sitemap    → build sitemap.xml from page records
  3. commit_sitemap      → git_commit sitemap.xml
  4. update_finding      → status = 'fixed_pending_verify'
  5. complete

Input:  { site_id, finding_id }
Output: { pages_in_sitemap, commit_sha }
```

**Handles finding types:** `sitemap_out_of_sync`.

### nav-updater

Handles navigation structure changes.

```
Workflow:
  1. load_finding_context      → finding, site nav structure
  2. apply_fix:
     - nav_item_orphaned:      → remove nav item from site_nav_items
     - orphan_page (add nav):  → add page to appropriate nav group
  3. re_render_nav_components  → re-render header and footer components
                                  with updated nav data
  4. redeploy_affected_pages   → pages that include nav need fresh HTML
                                  (or if nav is loaded dynamically, just
                                   the nav component file)
  5. update_finding            → status = 'fixed_pending_verify'
  6. complete

Input:  { site_id, finding_id }
Output: { nav_items_changed, pages_redeployed }
```

**Handles finding types:** `nav_item_orphaned`, `orphan_page` (when fix is adding to nav), `nav_too_complex` (when approved reorganisation).

### legal-updater

Handles legal page updates and disclaimer injection.

```
Workflow:
  1. load_finding_context       → finding, current legal page content
  2. load_current_template      → latest version of the legal template
                                   for this site's jurisdiction/industry
  3. generate_legal_content     → LLM generates legal page scoped to this site
                                   using current template + site details
  4. render_and_commit          → render page, commit to git
  5. update_finding             → status = 'fixed_pending_verify'
  6. complete

Input:  { site_id, finding_id }
Output: { page_updated, template_version, commit_sha }
```

**Handles finding types:** `missing_disclaimer`, `outdated_legal_template`.

### schema-fixer

Handles structured data / JSON-LD fixes.

```
Workflow:
  1. load_finding_context    → finding, page, current schema markup
  2. generate_schema         → build correct JSON-LD from page content
                                and page type (article, local business,
                                product, FAQ, etc.)
  3. inject_and_commit       → update page head with correct schema,
                                commit to git
  4. update_finding          → status = 'fixed_pending_verify'
  5. complete

Input:  { site_id, finding_id }
Output: { schema_type, commit_sha }
```

**Handles finding types:** `invalid_schema`, `stale_meta_description` (meta tag portion).

### image-optimiser

Handles image format conversion, resizing, and alt text.

```
Workflow:
  1. load_finding_context   → finding, image asset record
  2. apply_fix:
     - format conversion:   → convert to WebP/AVIF, commit optimised version
     - oversized:           → resize to appropriate dimensions, commit
     - missing_alt_text:    → LLM generates alt text from image context,
                              update page_component HTML, reassemble page
  3. update_finding         → status = 'fixed_pending_verify'
  4. complete

Input:  { site_id, finding_id }
Output: { images_optimised, bytes_saved }
```

**Handles finding types:** `image_not_optimised`, `missing_alt_text`, `oversized_image`.

### css-patcher

Handles narrow CSS fixes without full stylesheet regeneration.

```
Workflow:
  1. load_finding_context   → finding, current stylesheet
  2. apply_patch            → modify specific CSS variables, add missing
                               rules, fix media queries
  3. commit_stylesheet      → git_commit updated styles.css
  4. update_finding         → status = 'fixed_pending_verify'
  5. complete

Input:  { site_id, finding_id }
Output: { changes_applied, commit_sha }
```

**Handles finding types:** `css_variable_drift`, `missing_responsive_rule`.

---

## Cross-Domain Coordination

Domains are independent — each discovery agent writes to its own domain, each fix agent handles its own domain's fixes. Coordination happens through two mechanisms:

### 1. Impact assessment reads (during triage)

Triage reads other domains' tables to assess blast radius. This is read-only — triage never writes to other domains' tables.

```
Content triage for "suggest page removal":
  → reads link_registry (how many inbound links?)
  → reads site_nav_items (is page in nav?)
  → reads pages (build status, deploy status)
  → enriches finding with impact data
  → impact data changes resolution_path
     (zero links + not in nav = 'suggest',
      3 inbound links + in primary nav = 'flag')
```

### 2. Side-effect findings (during fix)

When a fix agent makes a change that affects other domains, it writes new findings with `parent_finding_id` pointing to the original. These get picked up by the appropriate domain on the next cycle.

```
section-rewriter removes a paragraph that contained a link:
  → new finding: domain='links', type='link_removed',
    parent_finding_id=original_finding
  
content-fixer removes a page (after human approval):
  → new finding: domain='links', type='inbound_links_broken', severity='high'
  → new finding: domain='navigation', type='nav_item_orphaned', severity='high'
  → new finding: domain='seo', type='sitemap_entry_removed', severity='medium'

These get triaged and fixed in their respective domains on the next cycle.
```

No agent calls another agent for coordination. They communicate through the findings table. Each picks up relevant findings on their own schedule.

---

## Catch-All Agent

Runs on a separate, less frequent cron (daily). Prevents entropy — the system accumulating unactioned findings that nobody cleans up.

```
maintenance-catch-all:
  1. find_stale_detected       → findings in 'detected' for > 48 hours
                                  (triage never ran — maybe domain orchestrator
                                   isn't enabled for this site)
                                → attempt basic triage, or reclassify domain,
                                  or mark for HITL

  2. find_stale_triaged        → findings in 'triaged' for > 7 days
                                  where resolution_path NOT 'auto_fix'
                                  (waiting for human action)
                                → send HITL reminder notification

  3. find_stale_suggestions    → findings with resolution_path = 'suggest'
                                  for > 14 days (human hasn't responded)
                                → escalate severity, or send reminder,
                                  or auto-close if low severity

  4. find_failed_fixes         → findings in 'fix_failed' status
                                → reclassify (try different fix agent,
                                  or escalate to HITL)

  5. cross_site_patterns       → check if same finding_type + affected_url
                                  appears across multiple sites
                                → create cross-site finding or notify
                                  for batch-fix
                                  Example: same external URL broken across
                                  50 sites → fix once, apply to all

  6. release_stuck_tasks       → maintenance_tasks in 'in_progress' for > 2 hours
                                  (worker crashed)
                                → reset to 'pending' for retry

  7. cleanup                   → archive resolved findings older than 90 days
                                → prune completed maintenance_tasks older than 30 days

  8. complete

Input:  {}
Output: { stale_detected, reminders_sent, failed_reclassified,
          cross_site_patterns, stuck_released, archived }
```

The catch-all also handles findings that no domain orchestrator claims — maybe the domain was set wrong during discovery, or the site's profile doesn't have that domain enabled. The catch-all can reclassify or escalate to HITL with a message like "these findings have been sitting for 2 weeks with no owner."

---

## Per-Site Configuration

Stored in `sites.settings` as a `maintenance_profile` JSON object. Controls which domains and sub-checks run, at what cadence, with what parameters.

```json
{
  "maintenance_profile": {
    "content": {
      "enabled": true,
      "every": "7d",
      "last_run_at": "2026-02-10T08:00:00Z",
      "agents": {
        "date_reference_scanner": true,
        "entity_drift_detector": true,
        "statistics_checker": false,
        "thin_content_detector": false,
        "content_gap_finder": false
      },
      "staleness_threshold_days": 180,
      "auto_fix_enabled": false
    },
    "links": {
      "enabled": true,
      "every": "8h",
      "last_run_at": "2026-02-12T06:00:00Z",
      "agents": {
        "internal_link_checker": true,
        "external_link_checker": true,
        "orphan_page_detector": true,
        "redirect_chain_detector": true
      },
      "auto_fix_redirects": true,
      "external_check_batch_size": 50
    },
    "seo": {
      "enabled": true,
      "every": "7d",
      "last_run_at": "2026-02-08T08:00:00Z",
      "agents": {
        "sitemap_sync_checker": true,
        "schema_validator": true,
        "meta_freshness_checker": false
      },
      "auto_fix_sitemap": true
    },
    "compliance": {
      "enabled": true,
      "every": "30d",
      "last_run_at": "2026-01-15T08:00:00Z",
      "agents": {
        "disclaimer_presence_checker": true,
        "legal_template_version_checker": true,
        "regulatory_change_monitor": false,
        "tool_compliance_checker": false
      },
      "regulatory_bodies": ["FCA", "ICO"],
      "jurisdictions": ["UK"]
    },
    "structural": {
      "enabled": false,
      "every": "30d",
      "agents": {
        "nav_complexity_checker": false,
        "redundant_content_detector": false,
        "competitor_structure_analyser": false
      },
      "competitors": []
    },
    "budget": {
      "llm_calls_per_cycle": 20,
      "max_auto_fixes_per_cycle": 5
    }
  }
}
```

The maintenance-batch-scheduler reads this to determine what's due. The site-maintenance-orchestrator reads it to know which discovery agents to spawn and what parameters to pass.

Different sites get different profiles:
- A simple brochure site: content + links, weekly/daily cadence
- A finance site with calculators: all domains including compliance with FCA monitoring
- A news site: content at higher frequency, feed lifecycle integration
- A site that just launched: links only for the first few months, then gradually enable more

---

## The Adopt / Research Connection

The adopt flow is closely related to competitive research. The same crawl → decompose → extract pipeline serves two purposes:

| Mode | Purpose | Output | Creates Site Record? |
|------|---------|--------|---------------------|
| **Adopt** | Import existing site into our system | Operational tables (sites, pages, page_components, nav, links) | Yes |
| **Research** | Understand a site to inform decisions | Intelligence tables (research_results, competitive_analysis) | No |

### Adopt pipeline

```
1. Crawl         → discover all pages, assets, links, nav structure
2. Classify      → what industry, what content strategy, what site type
3. Decompose     → break each page into sections, map to our component templates
4. Extract       → pull content into structured data for our templates
5. Store         → create site, pages, page_components, nav items, link registry
6. Assess        → run discovery agents (first maintenance cycle)
7. Baseline      → initial findings become the maintenance backlog
```

Steps 1-5 are adoption. Steps 6-7 are where adopt meets maintenance — the first maintenance cycle produces the initial findings backlog.

### Research pipeline

Same crawl/decompose/extract but output goes to intelligence tables:
- Structural patterns → what pages competitors have
- Design patterns → colour palettes, typography, component layouts → feeds theme library
- Content patterns → how this industry communicates, what topics they cover
- Compliance patterns → what disclaimers competitors include
- Tool inventory → what interactive tools/calculators competitors offer

This intelligence feeds the growth advisor (structural-discovery-agent's competitor analysis) and informs build-phase decisions for new sites in the same industry.

### Decompose/extract and the component library

When decomposing an adopted or researched site, each section gets mapped to our component library. Sections that don't match any component are opportunities to add new components. Every adoption that encounters a new pattern is an opportunity to grow the library.

Similarly, the CSS/design analysis can extract themes — colour palettes, typography, spacing — and store them in the theme library. Adopt three finance sites and you build up "what finance sites actually look like."

---

## Human Interaction

For now, human interaction uses the existing HITL message pattern. No new frontend needed initially.

### HITL notifications for maintenance

```
Finding with resolution_path = 'suggest':
  → HITL message: "We found that [page X] has a date reference from 2024.
     Here's a proposed rewrite: [proposed text].
     Approve / Reject / Modify"

Finding with resolution_path = 'flag':
  → HITL message: "Your competitors have a FAQ section and your site doesn't.
     Should we add one? This would require a new page and nav item."

Catch-all reminder:
  → HITL message: "These 5 maintenance findings have been waiting for
     your review for 2 weeks: [summary list]"
```

Human responses update finding status:
- Approve → status = 'approved', fix agent picks it up next cycle
- Reject → status = 'rejected'
- Won't fix → status = 'wont_fix'
- Ignore → status = 'ignored' (won't be re-detected due to finding_key dedup)

---

## Stale Finding Lifecycle

Findings that persist across cycles follow this lifecycle:

```
Detected → Triaged → [waiting for action]
                        │
                        ├── auto_fix: picked up next fix cycle
                        ├── suggest: HITL notification sent, wait for response
                        │     └── no response after 7 days: catch-all sends reminder
                        │     └── no response after 14 days: catch-all may auto-close
                        │         if severity is low, or escalate if high
                        ├── flag: HITL notification sent, wait for discussion
                        │     └── same escalation pattern as suggest
                        ├── monitor: re-evaluated each cycle
                        │     └── if condition worsens: severity escalated, reclassified
                        │     └── if condition resolves: auto-closed
                        └── ignore: persists, prevents re-detection via finding_key
```

The `finding_key` unique constraint prevents the same finding from being re-created each cycle. A stale date reference on page X, component Y has key `stale_date:pageX:componentY`. If it already exists (any status), the discovery agent's INSERT is silently skipped via `ON CONFLICT DO NOTHING`.

If a human marks a finding as `wont_fix` or `ignored`, the finding_key remains, so it won't be re-detected. If they mark it `rejected` (meaning "your proposed fix is wrong, try again"), the finding could be reclassified or a new finding created with a different approach.

---

## Multi-Site Patterns

The catch-all agent handles cross-site coordination. During its `cross_site_patterns` step:

```sql
-- Find the same broken URL across multiple sites
SELECT affected_url, COUNT(DISTINCT site_id) as site_count,
       array_agg(DISTINCT site_id) as affected_sites
FROM maintenance_findings
WHERE finding_type = 'broken_external_link'
  AND status NOT IN ('resolved', 'ignored', 'wont_fix')
GROUP BY affected_url
HAVING COUNT(DISTINCT site_id) > 1
ORDER BY site_count DESC;
```

When the same problem appears across multiple sites, the catch-all can:
- Find the fix once (e.g. new URL for the moved resource)
- Apply it across all affected sites in one pass
- Or create a batch fix task that the next scheduler run processes

The triage step within each site orchestrator can also check for this before writing a finding — "has this URL already been flagged by another site?" — to enrich the finding with cross-site context.

---

## Budget Management

Each site's maintenance profile includes a budget section:

```json
{
  "budget": {
    "llm_calls_per_cycle": 20,
    "max_auto_fixes_per_cycle": 5
  }
}
```

Discovery agents that need LLM calls check the budget before running. If exhausted, they skip and log that they were skipped. The catch-all can include budget status in HITL notifications.

This creates a natural tiering model:
- Basic tier: algorithmic-only maintenance (link checking, date scanning, sitemap sync)
- Standard tier: adds LLM-assisted discovery (entity drift, meta freshness)
- Premium tier: adds LLM-assisted fixes (section rewrites) and competitive analysis

---

## Implementation Phases

### Phase 0 — Foundation

The tables, the scheduler, and the simplest algorithmic discovery agents. Human reviews findings manually via HITL.

```
1. maintenance_findings table
2. maintenance_tasks table
3. maintenance-batch-scheduler agent definition + workflow
4. site-maintenance-orchestrator agent definition + workflow
   (discovery + triage steps only — no fix step yet)
5. content-discovery-agent (date_reference_scanner check only)
6. links-discovery-agent (internal_link_checker check only)
7. seo-discovery-agent (sitemap_sync_checker check only)
8. K8s CronJob manifest for the heartbeat trigger
9. HITL notifications for detected findings
```

No triage automation, no auto-fix. Humans review findings and act manually. This validates the model — are the discovery agents finding real problems? Are the findings useful?

### Phase 1 — Triage and Simple Fixes

Add triage logic to the site orchestrator. Add the first fix agents for deterministic, algorithmic fixes.

```
10. Triage step with impact cross-referencing
11. Priority scoring
12. redirect-manager fix agent (shorten chains, update links)
13. sitemap-regenerator fix agent
14. nav-updater fix agent (orphaned items)
15. links-discovery-agent: add external_link_checker, orphan_page_detector
16. compliance-discovery-agent: disclaimer_presence_checker
17. maintenance-catch-all agent (daily cron)
```

### Phase 2 — LLM-Assisted Discovery and Fixes

Add the section-rewriter and LLM-based discovery checks.

```
18. section-rewriter fix agent (the big one)
19. content-discovery-agent: add entity_drift_detector, statistics_checker
20. seo-discovery-agent: add meta_freshness_checker (LLM similarity)
21. legal-updater fix agent
22. schema-fixer fix agent
23. compliance-discovery-agent: add legal_template_version_checker
24. image-optimiser fix agent
```

### Phase 3 — Strategic and Competitive

Research-heavy agents, competitive analysis, structural suggestions.

```
25. structural-discovery-agent: nav_complexity_checker
26. structural-discovery-agent: competitor_structure_analyser (crawl + LLM)
27. content-discovery-agent: content_gap_finder
28. Adopt/research pipeline
29. compliance-discovery-agent: regulatory_change_monitor, tool_compliance_checker
30. css-patcher fix agent
```

### Phase 4 — Analytics and Advanced Automation

Requires analytics integration (Google Analytics API or similar). Likely several months out.

```
31. Analytics data source integration
32. Traffic-weighted priority scoring
33. keyword-drift-detector (SEO)
34. Auto-fix confidence thresholds (auto-approve section rewrites above N% confidence)
35. Cross-site pattern learning (findings common across industry)
36. Automated A/B testing of fixes
```

---

## Resolved Decisions

1. **Maintenance profile location:** `sites.settings` — simple, already exists, queryable via JSONB operators.

2. **Human interaction:** HITL messages via existing pattern. No new frontend initially.

3. **Stale findings:** Catch-all agent handles — reclassifies, sends HITL reminders, auto-closes low-severity items after timeout. Humans can set to `wont_fix` or `ignored`.

4. **Multi-site patterns:** Triage agent within each site can check for cross-site occurrences. Catch-all agent detects patterns across sites and can batch-fix.

5. **Fix agents:** Purpose-built, separate from build-phase agents. Build agents do "create from nothing" (broad scope). Fix agents do "change this specific thing" (narrow scope, finding-driven). They share some underlying actions (render_component, git_commit) but have their own workflows.

6. **Discovery agents:** Separately spawned as K8s jobs by the site-maintenance-orchestrator. Gives cleaner logs, failure isolation, and independent scaling.

7. **Spawn pattern:** Follows vet-batch-processor model. CronJob → agent-chassis → maintenance-batch-scheduler → site-maintenance-orchestrator (per site) → discovery/fix agents.

8. **Site independence:** Each site gets its own orchestrator instance with its own profile. Sites are fully independent — no shared state between site maintenance runs except the cross-site pattern detection in the catch-all.

---

## Open Questions

1. **Batch size tuning:** How many site orchestrators should run simultaneously? Depends on cluster capacity and spot instance behaviour. Start small (5-10), tune based on observation.

2. **Analytics integration timing:** Several discovery agents benefit from traffic data. Google Analytics API integration is a separate project — needed for Phase 4 but not before.

3. **Tool/calculator maintenance detail:** Tools that rely on regulatory data (mortgage calculators, tax calculators) need a structured dependency model. The compliance-discovery-agent checks for changes, but the data model for tool dependencies needs further design.

4. **Adopt pipeline detail:** The decompose/extract steps need detailed design — how do we map arbitrary HTML sections to our component templates? What happens when no template matches? This is a separate design exercise.

5. **Cost model:** Per-site billing based on maintenance tier? Flat fee with budget caps? Needs product/business input.

6. **Cadence staggering:** With 2000 sites, we don't want all weekly checks due on Monday. Should the scheduler stagger "week start" so ~1/7th of sites are due each day? Or let the batch_size naturally spread the load across CronJob invocations?
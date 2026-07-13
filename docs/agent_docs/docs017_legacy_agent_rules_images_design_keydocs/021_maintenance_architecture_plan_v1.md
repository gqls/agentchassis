# Site Maintenance Architecture — Plan

## The Model

Three layers, loosely coupled through a shared findings table:

```
Discovery Agents → maintenance_findings table → Triage → Fix Agents (reuse build agents)
```

Discovery agents find problems. Triage assesses severity and resolution path. Fix agents execute changes — mostly reusing agents from the build phase. Not every discovery type has an automated fix. Some findings need human decisions.

---

## The Findings Table

The coordination point for the entire maintenance system. Every discovery agent writes here. Triage reads and enriches. Fix agents read triaged items and update status after acting.

```sql
CREATE TABLE maintenance_findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    
    -- What was found
    domain TEXT NOT NULL,                    -- "content", "links", "seo", "legal", "design", "structural", "compliance", "technical"
    finding_type TEXT NOT NULL,              -- "stale_date_reference", "broken_external_link", "missing_competitor_feature", etc.
    severity TEXT NOT NULL DEFAULT 'low',    -- "info", "low", "medium", "high", "urgent"
    summary TEXT NOT NULL,                   -- human-readable description
    detail JSONB DEFAULT '{}',              -- finding-specific structured data
    
    -- What it affects
    page_id UUID REFERENCES pages(id),       -- NULL for site-wide findings
    component_id UUID,                       -- specific page_component if applicable
    entity_id UUID,                          -- specific entity if applicable
    affected_url TEXT,                       -- the URL or link in question
    
    -- Triage enrichment (filled by triage step)
    impact JSONB,                            -- inbound links, nav membership, SEO value, traffic etc.
    resolution_path TEXT,                    -- "auto_fix", "suggest", "flag", "monitor", "ignore"
    suggested_action TEXT,                   -- "rewrite_section", "remove_page", "add_redirect", "update_disclaimer", etc.
    suggested_detail JSONB,                  -- action-specific data (e.g. proposed rewrite, redirect target)
    priority INTEGER,                        -- computed from severity + impact
    
    -- Lifecycle
    status TEXT NOT NULL DEFAULT 'detected', -- "detected", "triaged", "approved", "in_progress", "fixed", "verified", "rejected", "wont_fix"
    discovered_by TEXT NOT NULL,             -- agent type that found it
    fixed_by TEXT,                           -- agent type that fixed it
    approved_by TEXT,                        -- "auto" or user identifier
    
    -- Cross-domain references
    parent_finding_id UUID REFERENCES maintenance_findings(id),  -- if this was created as a side-effect of fixing another finding
    related_finding_ids UUID[],              -- findings that should be considered together
    
    -- Deduplication
    finding_key TEXT,                        -- deterministic key for dedup (e.g. "stale_date:page_id:component_id")
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    triaged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    
    UNIQUE(site_id, finding_key)             -- prevents duplicate findings
);

CREATE INDEX idx_mf_site_status ON maintenance_findings(site_id, status);
CREATE INDEX idx_mf_domain ON maintenance_findings(domain, status);
CREATE INDEX idx_mf_page ON maintenance_findings(page_id);
CREATE INDEX idx_mf_priority ON maintenance_findings(priority DESC NULLS LAST) WHERE status = 'triaged';
```

---

## Discovery Agents

Each discovery agent has a narrow focus, a detection method, and a cadence. Grouped by domain.

### Content Discovery

| Agent | Finds | Method | Cadence | Severity |
|-------|-------|--------|---------|----------|
| `date-reference-scanner` | Pages with time-sensitive text ("in 2024", "last year", specific dates) | Regex + pattern matching on rendered HTML | Weekly | Low-Medium |
| `statistics-checker` | Outdated statistics, figures that should be refreshed | Pattern matching + optional LLM assessment | Monthly | Medium |
| `entity-drift-detector` | Entity data changed but page content hasn't (price changed, team member left) | Compare page_components content against site_entities data | Daily (if entities change) | Medium-High |
| `thin-content-detector` | Sections with placeholder or insufficient content | Word count + LLM quality assessment | Monthly | Low |
| `voice-consistency-checker` | Content that sounds different from rest of site (different era, different writer) | LLM comparison against brand spec | Monthly | Low |
| `content-gap-finder` | Topics competitors cover that we don't | Crawl competitors + LLM comparison | Monthly | Info |

Implementation priority: `date-reference-scanner` first (algorithmic, high value). Then `entity-drift-detector` (data-driven, automated). The LLM-heavy ones later.

### Link Discovery

| Agent | Finds | Method | Cadence |
|-------|-------|--------|---------|
| `internal-link-checker` | Broken internal links (target page missing or not deployed) | Query link_registry against pages table | Daily |
| `external-link-checker` | Broken external links (HTTP HEAD returns 4xx/5xx) | HTTP HEAD requests, rate-limited | Weekly |
| `orphan-page-detector` | Pages with no inbound internal links | Query link_registry for pages with zero inbound | Weekly |
| `redirect-chain-detector` | Redirect chains (A→B→C) | Follow redirects in redirect table | Weekly |
| `missing-crosslink-finder` | Related pages that should link to each other but don't | Entity relationships vs actual links | Monthly |
| `anchor-text-auditor` | Empty anchors, generic "click here" text | Parse link_registry anchor text | Monthly |

Implementation priority: `internal-link-checker` and `external-link-checker` first — these are the links-orchestrator's existing design. `orphan-page-detector` next.

### SEO Discovery

| Agent | Finds | Method | Cadence |
|-------|-------|--------|---------|
| `meta-freshness-checker` | Meta descriptions that don't match current page content | Compare meta against rendered content (LLM similarity) | Weekly |
| `schema-validator` | Invalid or outdated structured data / JSON-LD | Schema validation against schema.org | Weekly |
| `sitemap-sync-checker` | Pages in sitemap that don't exist, or deployed pages missing from sitemap | Compare sitemap XML against pages table | Daily |
| `keyword-drift-detector` | Pages that used to rank but content has shifted away from target keywords | Needs analytics integration | Monthly |

Implementation priority: `sitemap-sync-checker` (algorithmic, fast). `schema-validator` next.

### Legal / Compliance Discovery

| Agent | Finds | Method | Cadence |
|-------|-------|--------|---------|
| `legal-template-version-checker` | Deployed legal pages using outdated templates | Compare deployed content hash against current template hash | Monthly |
| `disclaimer-presence-checker` | Pages that should have disclaimers but don't (per legal_rules config) | Scan page content for disclaimer text, check against rules | Weekly |
| `regulatory-change-monitor` | Regulations that affect this site have been updated | Check regulatory body RSS/feeds/pages for changes | Monthly |
| `tool-compliance-checker` | Tools/calculators relying on standards that may have changed | Check tool dependency config against regulatory sources | Monthly |

Implementation priority: `disclaimer-presence-checker` first (algorithmic, uses existing legal_rules). `legal-template-version-checker` next.

### Structural Discovery

| Agent | Finds | Method | Cadence |
|-------|-------|--------|---------|
| `competitor-structure-analyser` | Structural features competitors have that we lack (FAQ, blog, resources, tools) | Crawl + classify competitor pages | Monthly |
| `nav-complexity-checker` | Nav grown too large, redundant items, poor grouping | Analyse nav table item counts and structure | Monthly |
| `redundant-content-detector` | Multiple pages covering similar ground | LLM similarity comparison across pages | Monthly |
| `page-purpose-auditor` | Pages that exist but don't serve a clear purpose | LLM assessment of page content vs site goals | Quarterly |

Implementation priority: `nav-complexity-checker` first (algorithmic, uses existing nav tables). The rest are research-heavy and come later.

### Design / Technical Discovery

| Agent | Finds | Method | Cadence |
|-------|-------|--------|---------|
| `component-version-checker` | Pages using outdated component versions when library has newer ones | Compare page_components against content_components versions | Monthly |
| `image-optimisation-scanner` | Images not in modern formats, oversized, missing alt text | Scan assets table and rendered HTML | Monthly |
| `performance-baseline` | Page load time regression, asset size growth | Lighthouse or similar automated testing | Monthly |
| `accessibility-checker` | WCAG violations | Automated accessibility testing (axe-core or similar) | Monthly |
| `mobile-rendering-checker` | Pages that render poorly on mobile | Automated viewport testing | Monthly |

Implementation priority: `image-optimisation-scanner` first (algorithmic, uses assets table). `component-version-checker` next.

---

## Triage

Triage is a step that runs within each domain orchestrator, not a separate agent. It reads "detected" findings for its domain and enriches them.

### What triage does

1. **Deduplication** — is this finding already recorded from a previous cycle? The `finding_key` unique constraint handles exact duplicates. Triage also checks for near-duplicates (same page, similar finding type).

2. **Impact assessment** — cross-references other agents' tables:

```
For a content finding about page X:
  - link_registry: how many inbound links? (removal breaks these)
  - site_nav_items: is it in navigation? which groups? (removal needs nav update)
  - pages: what's the build_status, last_deployed? (how visible is this?)
  - site_entities: any entity references? (entity page vs editorial page)
  - maintenance_findings: any other findings about this page? (cluster related issues)
```

3. **Resolution classification:**

| Path | Meaning | Examples |
|------|---------|---------|
| `auto_fix` | Safe to fix without human approval | Redirect chain shortening, sitemap regeneration, image format conversion |
| `suggest` | Propose a specific fix, await approval | "Rewrite this paragraph — here's the proposed text", "Add this disclaimer" |
| `flag` | Needs human discussion, no specific fix proposed | "Competitor has a blog section — should we?", "This page's purpose is unclear" |
| `monitor` | Not actionable yet, but track it | "This statistic is 8 months old — flag again at 12 months" |
| `ignore` | Finding is valid but not worth acting on | "This page has a date from 2024 but it's a case study — dates are expected" |

4. **Priority scoring** — combines severity (from discovery) with impact (from cross-referencing):

```
priority = severity_weight × impact_multiplier

severity_weight:
  info=1, low=2, medium=4, high=8, urgent=16

impact_multiplier:
  in_primary_nav=3, has_inbound_links=2, has_entity=2, 
  is_homepage=5, high_traffic=3, recent_deploy=0.5
```

Numbers are illustrative — the point is that a medium-severity finding on the homepage is higher priority than a high-severity finding on an orphan page nobody visits.

### What triage doesn't do

- Doesn't fix anything
- Doesn't call other agents
- Doesn't write to other agents' tables
- Only reads other agents' tables and writes back to maintenance_findings

---

## Fix Agents

Fixes reuse existing build-phase agents where possible. The maintenance system invokes them with a narrower scope.

### Mapping findings to fix agents

| Finding Type | Fix Agent | Scope | Auto-fixable? |
|-------------|-----------|-------|---------------|
| `stale_date_reference` | page-content-writer | Rewrite specific section | No — needs LLM judgment on replacement text |
| `broken_internal_link` | links-orchestrator | Update or remove link | Sometimes — if redirect exists, auto-fix. Otherwise suggest. |
| `broken_external_link` | page-content-writer | Rewrite sentence/paragraph with alternative source | No — needs LLM to find alternative |
| `orphan_page` | nav-agent (populate_nav) | Add to appropriate nav group | No — suggest to human |
| `redirect_chain` | redirect-manager | Shorten chain A→B→C to A→C | Yes |
| `sitemap_out_of_sync` | deployer-agent | Regenerate and redeploy sitemap | Yes |
| `outdated_legal_template` | legal-content-agent | Regenerate from current template | Suggest — legal changes need review |
| `missing_disclaimer` | legal-content-agent | Inject disclaimer into page | Suggest |
| `invalid_schema` | seo-content-agent | Regenerate structured data | Yes (if schema is deterministic) |
| `meta_description_stale` | seo-content-agent | Regenerate meta from current content | Suggest |
| `image_not_optimised` | image-generator / asset pipeline | Convert format, resize | Yes |
| `outdated_component_version` | page-rerender | Re-render page with new component | Suggest — visual change needs review |
| `entity_data_changed` | page-content-writer + page-rerender | Update content then re-render | Sometimes — price change is auto, structural change is suggest |
| `nav_too_complex` | nav-agent | Suggest reorganisation | No — structural decision |
| `competitor_has_feature` | site-planner | Propose new page/section | No — strategic decision |

### Cross-domain side effects

When a fix agent acts, it may create findings in other domains. The fix agent writes these as new `maintenance_findings` with `parent_finding_id` set to the original finding:

```
Content fixer removes a page (approved by human):
  → new finding: domain="links", type="inbound_links_broken", severity="high"
  → new finding: domain="navigation", type="nav_item_orphaned", severity="high"  
  → new finding: domain="seo", type="sitemap_entry_removed", severity="medium"

These get picked up by their respective domain orchestrators on the next heartbeat.
```

This is how coordination happens without agents calling each other.

---

## Domain Orchestrators

Each domain orchestrator runs on heartbeat and handles its own find → triage → fix cycle. Each can run independently.

### Content Maintenance Orchestrator

```
Heartbeat invocation:
  1. Run enabled content discovery agents for this site
     (date-reference-scanner, entity-drift-detector, etc.)
  2. Triage new "detected" findings in content domain
     - Cross-reference links, nav, SEO tables for impact
     - Classify resolution path
     - Score priority
  3. Execute auto-fixable items (if any exist for content domain)
  4. Verify previous fixes (re-scan fixed pages)
  5. Pick up cross-domain findings directed at content
     (e.g. links agent found broken external link → content needs rewriting)
```

### Links Maintenance Orchestrator

```
Heartbeat invocation:
  1. Run link discovery agents
     (internal-link-checker, external-link-checker, orphan-page-detector)
  2. Triage new link findings
  3. Auto-fix: redirect chains, internal link updates where redirect exists
  4. Verify previous fixes
  5. Pick up cross-domain findings (page removed → create redirects)
```

### SEO Maintenance Orchestrator

```
Heartbeat invocation:
  1. Run SEO discovery agents
     (sitemap-sync, schema-validator, meta-freshness)
  2. Triage
  3. Auto-fix: sitemap regeneration, schema fixes
  4. Verify
```

### Compliance Maintenance Orchestrator

```
Heartbeat invocation:
  1. Run compliance discovery agents
     (disclaimer-checker, legal-template-version, regulatory-change-monitor, tool-compliance)
  2. Triage — almost never auto-fix, mostly suggest or flag
  3. Pick up cross-domain findings that have compliance implications
```

### Structural / Growth Orchestrator

```
Heartbeat invocation (less frequent — monthly):
  1. Run structural discovery agents
     (competitor-analysis, nav-complexity, redundant-content, page-purpose)
  2. Triage — mostly flag for human discussion
  3. No auto-fix — all findings need strategic decisions
```

---

## Per-Site Configuration

Each site has a maintenance profile that controls which orchestrators run and at what cadence:

```json
{
  "maintenance_profile": {
    "content": {
      "enabled": true,
      "cadence": "weekly",
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
      "cadence": "daily",
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
      "cadence": "weekly",
      "agents": {
        "sitemap_sync_checker": true,
        "schema_validator": true,
        "meta_freshness_checker": false
      },
      "auto_fix_sitemap": true
    },
    "compliance": {
      "enabled": true,
      "cadence": "monthly",
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
      "enabled": false
    }
  }
}
```

Stored on `sites.settings` or a dedicated `site_maintenance_config` table. The heartbeat scheduler reads this to decide what to invoke for each site.

---

## The Adopt / Research Connection

The same crawl → decompose → extract pipeline serves two purposes:

| Mode | Purpose | Output destination | Creates site record? |
|------|---------|-------------------|---------------------|
| **Adopt** | Import site into our system for ongoing maintenance | Operational tables (sites, pages, page_components, nav, links) | Yes |
| **Research** | Understand a site to inform decisions about other sites | Intelligence tables (research_results, competitive_analysis) | No |

In adopt mode, the first maintenance cycle after import produces the initial findings backlog — "here's everything that needs attention on this site."

In research mode, the analysis feeds into:
- Growth advisor: "competitor has X, we don't"
- Theme library: extract their design patterns
- Component library: discover new layout patterns
- Content patterns: how this industry communicates
- Compliance: what disclaimers do competitors use

Both modes use the same discovery agents for the initial assessment. The difference is whether the output creates a manageable site or reference data.

---

## Implementation Phases

### Phase 0 — Foundation (build first)

Just the table and the simplest discovery agents:

```
1. maintenance_findings table
2. internal-link-checker (algorithmic, reuses link_registry)
3. date-reference-scanner (algorithmic, regex on page HTML)
4. sitemap-sync-checker (algorithmic, compare tables)
5. Human-facing findings dashboard (read-only — just show what's found)
```

No triage automation, no auto-fix. Humans review findings and act manually. This validates the model before investing in automation.

### Phase 1 — Triage and simple fixes

```
6. Triage step with impact cross-referencing
7. Auto-fix for redirect chains
8. Auto-fix for sitemap regeneration
9. Priority scoring
10. external-link-checker (HTTP HEAD, rate-limited)
11. disclaimer-presence-checker
12. orphan-page-detector
```

### Phase 2 — LLM-assisted discovery and fixes

```
13. entity-drift-detector (compare entity data to page content)
14. meta-freshness-checker (LLM similarity)
15. Content rewriting via page-content-writer (scoped to single section)
16. legal-template-version-checker
17. image-optimisation-scanner + auto-fix
18. component-version-checker
```

### Phase 3 — Strategic and competitive

```
19. competitor-structure-analyser (crawl + LLM)
20. content-gap-finder
21. nav-complexity-checker
22. regulatory-change-monitor
23. tool-compliance-checker
24. Adopt/research pipeline feeding competitive intelligence
```

### Phase 4 — Full automation

```
25. Auto-fix content rewrites (with confidence thresholds)
26. Auto-suggest structural changes
27. Automated A/B testing of changes
28. Analytics-driven priority adjustment
29. Cross-site pattern learning (findings common across sites in same industry)
```

---

## Open Questions

1. **Where does the maintenance profile live?** `sites.settings` (simple, already exists) vs a dedicated table (queryable across sites, easier to set defaults per industry).

2. **How does the human interact?** Dashboard showing findings with approve/reject buttons? Email digest? In-app notifications? All of these eventually, but what's first?

3. **How do we handle findings that persist across cycles?** A stale date reference will be re-detected every week. The `finding_key` dedup prevents duplicates, but the finding sits in "detected" or "triaged" indefinitely. Do findings escalate in severity over time? Do they eventually auto-close?

4. **Analytics integration** — several discovery agents would benefit from knowing which pages get traffic. This needs an analytics data source (Google Analytics API, Cloudflare analytics, or similar). When does this become available?

5. **Cost management** — LLM-heavy discovery agents (content-gap-finder, meta-freshness-checker, voice-consistency-checker) cost money to run. Per-site billing? Included in subscription tiers? Cost caps per heartbeat cycle?

6. **Multi-site patterns** — if the same broken external link appears across 50 sites (common resource moved), should we detect this once and apply across all? This is a cross-site coordination problem that single-site orchestrators don't handle.
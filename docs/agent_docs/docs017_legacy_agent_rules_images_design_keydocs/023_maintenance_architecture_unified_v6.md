# 023 — Unified Build & Maintenance Architecture

## Core Idea

Build and maintenance are the same process. A new site is a set of findings that need fixing. An existing site generates findings through discovery. Both go through the same queue, processed by the same orchestrator, handled by the same fix agents.

A site starts minimal and improves incrementally. Users never see a broken site — each work item leaves the site in a working state. The system adds pages, tools, entity directories, and improvements over time through the same mechanism that later maintains them.

---

## Naming and Coexistence

The existing `pageflow-builder`, `site-classifier`, `site-planner`, and `intake-orchestrator` continue to work unchanged. They are the production pipeline for content sites and must not be disrupted.

The new system uses separate agent definitions:

| Existing (keeps working) | New (parallel system) |
|--------------------------|----------------------|
| `intake-orchestrator` | `intake-orchestrator-v2` |
| `site-classifier` | `site-classifier-v2` |
| `site-planner` | `site-planner-v2` |
| `pageflow-builder` | `site-work-orchestrator` |
| `maintenance-triage` + `page-rebuild` | `maintenance-batch-scheduler` + `site-work-orchestrator` |

The v2 intake orchestrator routes to either the existing pageflow-builder (for straightforward content sites where the existing pipeline works well) or the new site-work-orchestrator (for mixed sites, tool-heavy sites, entity-driven sites, or sites where incremental building is preferred).

Over time, as the new system proves reliable, more sites can be routed through it. The existing pipeline doesn't need to be retired — it's a valid fast path for simple content sites. The new system handles the cases the old one can't.

Similarly, the existing `maintenance_queue` table and `maintenance-triage` agent continue working. The new `site_work_items` table runs alongside. Migration happens gradually — new sites use the new table, existing sites stay on the old system until individually migrated.

---

Replaces the separate concepts of "build plan pages" and "maintenance findings." Every piece of work — building a new page, fixing stale content, adding a tool, publishing a news article — is a work item.

```sql
CREATE TABLE site_work_items (
                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                 site_id UUID NOT NULL REFERENCES sites(id),

    -- What needs doing
                                 source TEXT NOT NULL,                    -- 'planner', 'discovery', 'content_feed',
    -- 'manual', 'improvement', 'side_effect'
                                 domain TEXT NOT NULL,                    -- 'build', 'content', 'links', 'seo',
    -- 'compliance', 'structural', 'design',
    -- 'navigation', 'entity', 'tools'
                                 item_type TEXT NOT NULL,                 -- 'needs_content_page', 'needs_tool_page',
    -- 'stale_date_reference', 'broken_link',
    -- 'publish_article', 'entity_data_drift',
    -- 'needs_entity_setup', etc.
                                 severity TEXT NOT NULL DEFAULT 'medium', -- 'info', 'low', 'medium', 'high', 'urgent'
                                 summary TEXT NOT NULL,                   -- human-readable description
                                 spec JSONB DEFAULT '{}',                -- work-specific structured data
    -- (page_spec, tool_spec, finding detail,
    --  entity config, article brief, etc.)

    -- What it affects
                                 page_id UUID REFERENCES pages(id),       -- NULL for site-wide items
                                 component_id UUID,                       -- specific page_component if applicable
                                 entity_id UUID,                          -- specific entity if applicable
                                 affected_url TEXT,                       -- the URL or resource in question

    -- Triage enrichment
                                 impact JSONB,                            -- inbound links, nav membership, traffic
                                 resolution_path TEXT,                    -- 'auto_fix', 'suggest', 'flag',
    -- 'monitor', 'ignore'
                                 suggested_action TEXT,                   -- 'build_page', 'rewrite_section',
    -- 'add_redirect', 'build_tool', etc.
                                 priority INTEGER DEFAULT 100,            -- computed from severity + impact
    -- lower = higher priority
                                 handler_agent TEXT,                      -- which agent processes this
    -- (set during triage or by planner)

    -- Lifecycle
                                 status TEXT NOT NULL DEFAULT 'detected', -- 'detected', 'triaged', 'approved',
    -- 'claimed', 'in_progress',
    -- 'complete', 'pending_verify',
    -- 'verified', 'failed',
    -- 'rejected', 'wont_fix'
                                 created_by TEXT NOT NULL,                -- agent type or 'manual'
                                 handled_by TEXT,                         -- agent type that processed it
                                 approved_by TEXT,                        -- 'auto' or user identifier
                                 claimed_by TEXT,                         -- job ID that claimed this

    -- Dependencies and relationships
                                 depends_on UUID[],                       -- items that must complete first
                                 parent_item_id UUID REFERENCES site_work_items(id),
    -- if created as side-effect of another item
                                 related_item_ids UUID[],                 -- items to consider together
                                 batch_id UUID,                           -- groups items from same planning/discovery run

    -- Tracking
                                 attempt_count INTEGER DEFAULT 0,
                                 max_attempts INTEGER DEFAULT 3,
                                 result JSONB DEFAULT '{}',              -- what the handler produced
    -- includes commit_sha for git tracking
                                 error TEXT,

    -- Deduplication
                                 item_key TEXT,                           -- deterministic key for dedup
    -- e.g. 'stale_date:page_id:component_id'
    -- or 'needs_page:about'

                                 created_at TIMESTAMPTZ DEFAULT NOW(),
                                 triaged_at TIMESTAMPTZ,
                                 claimed_at TIMESTAMPTZ,
                                 completed_at TIMESTAMPTZ,

                                 UNIQUE(site_id, item_key)               -- prevents duplicate items
);

CREATE INDEX idx_swi_pending ON site_work_items(site_id, status, priority)
    WHERE status IN ('pending', 'triaged', 'approved');
CREATE INDEX idx_swi_claimed ON site_work_items(status, claimed_by)
    WHERE status = 'claimed';
CREATE INDEX idx_swi_handler ON site_work_items(handler_agent, status)
    WHERE status IN ('triaged', 'approved');
CREATE INDEX idx_swi_batch ON site_work_items(batch_id);
CREATE INDEX idx_swi_page ON site_work_items(page_id)
    WHERE page_id IS NOT NULL;
CREATE INDEX idx_swi_deps ON site_work_items USING GIN(depends_on)
    WHERE depends_on IS NOT NULL;
```

### Archival

Completed items are archived after 90 days by the catch-all agent:

```sql
-- Move to archive
INSERT INTO site_work_items_archive
SELECT * FROM site_work_items
WHERE status IN ('verified', 'rejected', 'wont_fix')
  AND completed_at < NOW() - INTERVAL '90 days';

DELETE FROM site_work_items
WHERE status IN ('verified', 'rejected', 'wont_fix')
  AND completed_at < NOW() - INTERVAL '90 days';
```

The active table stays small. Queries always filter on `site_id + status`, which the composite index handles well even at scale.

---

## Content Feed Items: Separate Table

News and article content has a different lifecycle from work items. Feed items go through ingestion, filtering, deduplication, and relevance scoring before they become publishable. Time-sensitive content (finance news, event updates) has different cadence requirements per site. The preprocessing pipeline is its own concern.

```sql
CREATE TABLE content_feed_items (
                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                    site_id UUID REFERENCES sites(id),       -- NULL if not yet assigned to a site
                                    source_id UUID,                          -- references content_sources config

    -- Source data
                                    external_id TEXT,                        -- ID from the source (RSS guid, API id)
                                    source_url TEXT,                         -- original article URL
                                    source_title TEXT,
                                    source_summary TEXT,
                                    source_content TEXT,                     -- full text if available
                                    source_published_at TIMESTAMPTZ,

    -- Processing
                                    relevance_score FLOAT,                   -- 0.0-1.0, per-site relevance
                                    topics JSONB DEFAULT '[]',               -- extracted topics/tags
                                    entity_ids UUID[],                       -- cross-referenced entities
                                    duplicate_of UUID REFERENCES content_feed_items(id),

    -- Publication
                                    status TEXT DEFAULT 'ingested',          -- 'ingested', 'filtered', 'relevant',
    -- 'queued', 'published', 'rejected',
    -- 'expired', 'duplicate'
                                    work_item_id UUID REFERENCES site_work_items(id),
    -- link to the work item when queued
                                    published_page_id UUID REFERENCES pages(id),

                                    created_at TIMESTAMPTZ DEFAULT NOW(),
                                    processed_at TIMESTAMPTZ,
                                    published_at TIMESTAMPTZ,
                                    expires_at TIMESTAMPTZ                   -- for time-sensitive content
);

CREATE INDEX idx_cfi_status ON content_feed_items(site_id, status);
CREATE INDEX idx_cfi_source ON content_feed_items(source_id, external_id);
CREATE INDEX idx_cfi_relevance ON content_feed_items(site_id, relevance_score DESC)
    WHERE status = 'relevant';
```

### Feed Pipeline

```
content_sources (config table)
  → content-feed-orchestrator (discovery agent, runs on schedule)
    → ingest from sources (RSS, API, scrape)
    → filter (relevance scoring per site)
    → deduplicate (same story from multiple sources)
    → for items worth publishing:
      → write work item to site_work_items
        (source: 'content_feed', handler_agent: 'article-writer')
      → link feed item to work item via work_item_id
    → expire stale items
```

The feed orchestrator runs at its own cadence per site, controlled by the maintenance profile:

```json
{
  "content_feed": {
    "enabled": true,
    "every": "4h",
    "max_articles_per_cycle": 3,
    "relevance_threshold": 0.6,
    "sources": [
      { "type": "rss", "url": "...", "topics": ["boxing", "fight news"] }
    ],
    "time_sensitivity": "normal"
  }
}
```

For finance sites, `"time_sensitivity": "high"` and `"every": "1h"` with immediate processing triggers. For blog-style content sites, `"every": "24h"` and `"max_articles_per_cycle": 1`.

### News Display is a Design Concern, Not a Feed Concern

The feed pipeline handles content supply — what to publish, when, from which sources. But the visual presentation of news — article card layouts, news grid components, featured article styling, date formatting, category tags, author blocks, article detail templates — is part of the site's design system.

When the classifier identifies `page_types: ["news"]` and the planner creates the site structure, it plans news display components as part of the site architecture:

- News landing page with `article-grid` or `news-feed` component
- Article detail template with `article-layout` component
- Sidebar widgets like `latest-articles` or `related-articles`
- Category/tag index pages

The webdesign-agent styles these components as part of the CSS theme, just like it styles hero sections and service grids. News typography, card spacing, date formatting, category tag colours — all part of the style collection.

The article-writer agent then fills in these pre-designed templates when publishing. It doesn't make layout decisions — it provides content (headline, body, image, category, date) and the existing component template + CSS theme handle the presentation.

This means news design work items belong to the `build` or `design` domain:

```
needs_content_page      | news/index    | domain: build  | handler: page-content-writer
needs_article_template  | article       | domain: build  | handler: page-content-writer
design_news_components  | site          | domain: design | handler: webdesign-agent
```

While news content work items belong to the `content_feed` domain:

```
publish_article         | news/boxing-preview | domain: content_feed | handler: article-writer
```

Improving how news looks on the site (better card design, mobile article layout, adding a "trending" section) goes through the normal design/build work items. These are design iterations, not feed operations. The feed pipeline is unaware of them.

---

## The Site Lifecycle

```
Phase 1: Initial Build
  1. Classify + brief (human involved)
  2. Planner writes work items (source: 'planner')
  3. Site work orchestrator processes items → fix agents build pages
  4. Individual git commits per page → each goes live via GitHub Action → S3
  5. Basic site is live

Phase 2+: Incremental Improvement
  6. Planner or improvement agent writes more items
     (additional pages, tools, entity setup)
  7. Orchestrator processes → site grows
  8. Each batch leaves site in working state

Ongoing: Maintenance
  9. Discovery agents scan → write items (source: 'discovery')
  10. Triage enriches items
  11. Fix agents process → site stays fresh
  12. Feed orchestrator writes article items (source: 'content_feed')
  13. Article writers publish → site gains content
  14. Goto 9
```

Steps 3, 7, and 11 are the same code.

---

## The Orchestrator

### site-work-orchestrator

Replaces both the build orchestrator and the site-maintenance-orchestrator. One agent handles all work for a site, regardless of whether the work originated from a planner, a discovery agent, a feed pipeline, or a human.

```
Workflow:
  1. load_site_context        → site record, maintenance profile,
                                 current work items summary

  2. process_pending_items    → query site_work_items for this site:
                                 status IN ('triaged', 'approved')
                                 AND dependencies satisfied
                                 ORDER BY priority
                               → for each: route to handler_agent
                               → each handler: do work → git commit → mark complete

  3. verify_previous_work     → items with status = 'pending_verify'
                               → re-check the specific change
                               → mark 'verified' or 'failed'

  4. run_discovery            → for each domain in the maintenance profile
                                 that is due:
                                 spawn discovery agent → collect findings
                               → (skipped on first build — nothing to discover
                                  on an empty/new site)

  5. triage_new_items         → items with status = 'detected'
                               → cross-reference impact
                               → classify resolution_path
                               → assign handler_agent
                               → score priority
                               → update to 'triaged'

  6. update_last_run          → update maintenance profile timestamps

  7. complete

Input:  { site_id, domains: ['build', 'content', 'links', ...] }
Output: { items_processed, items_verified, items_discovered, items_triaged }
```

Order is deliberate: process existing work first, verify previous work, then discover new work, then triage. Each cycle clears the backlog before adding to it.

On initial build, domains=['build']. Discovery is skipped (new site). The only items to process are the planner's build items. After the site exists, domains expand to include maintenance domains.

### Batch Scheduler

Same pattern as the maintenance doc. CronJob triggers the batch scheduler, which claims sites that are due for work and spawns one site-work-orchestrator per site.

```
K8s CronJob (every 8 hours)
  └─► agent-chassis → maintenance-batch-scheduler
        1. populate_work_queue  → query sites with due maintenance OR
                                   pending build items
        2. claim_batch          → FOR UPDATE SKIP LOCKED, batch_size sites
        3. process_batch        → loop: spawn site-work-orchestrator per site
        4. complete
```

For initial builds, the intake-orchestrator can trigger the site-work-orchestrator directly (no need to wait for the cron). For maintenance, the cron handles scheduling.

---

## The Planner as Discovery Agent

The site-planner changes from "return a page list" to "write work items." It becomes a specialised discovery agent that discovers what a site needs based on the brief and classification.

### Planner Output

Instead of returning:
```json
{
  "pages": [
    {"name": "index", "sections": ["hero", "features", "call_to_action"]},
    {"name": "about", "sections": ["about-hero", "about-content"]}
  ]
}
```

The planner writes to `site_work_items`:

```
For a simple brochure site (leopardessconsulting.co.uk):

  needs_content_page  | index    | priority 10 | handler: page-content-writer
  needs_content_page  | about    | priority 10 | handler: page-content-writer
  needs_content_page  | services | priority 10 | handler: page-content-writer
  needs_content_page  | contact  | priority 10 | handler: page-content-writer
  needs_logo          | site     | priority 5  | handler: image-generator
  needs_hero_image    | index    | priority 5  | handler: image-generator
  needs_design        | site     | priority 8  | handler: webdesign-agent
  needs_deploy        | site     | priority 20 | handler: deployer-agent
```

For wykefarm (mixed site):

```
  -- Tier 1: core content pages
  needs_content_page  | index      | priority 10 | handler: page-content-writer
  needs_content_page  | philosophy | priority 10 | handler: page-content-writer
  needs_content_page  | wildlife   | priority 10 | handler: page-content-writer
  needs_content_page  | contact    | priority 10 | handler: page-content-writer

  -- Tier 1: commerce
  needs_content_page  | shop       | priority 10 | handler: page-content-writer
  needs_content_page  | produce/meat      | priority 12 | handler: page-content-writer
  needs_content_page  | produce/happy-cows| priority 12 | handler: page-content-writer

  -- Tier 1: site infrastructure
  needs_logo          | site       | priority 5  | handler: image-generator
  needs_design        | site       | priority 8  | handler: webdesign-agent

  -- Tier 2: tools (can be built after core site is live)
  needs_tool_page     | tools/carbon    | priority 30 | handler: tool-builder
  needs_tool_page     | tools/feed      | priority 30 | handler: tool-builder
  needs_tool_page     | tools/freezer   | priority 30 | handler: tool-builder
  ... (12 tools total)
  needs_directory     | tools           | priority 35 | handler: directory-builder
                                                         depends_on: [all tool items]

  -- Tier 3: entity directory
  needs_entity_setup  | species         | priority 50 | handler: entity-data-agent
  needs_entity_pages  | species/*       | priority 55 | handler: entity-page-builder
                                                         depends_on: [entity_setup]
  needs_directory     | species         | priority 58 | handler: directory-builder
                                                         depends_on: [entity_pages]
```

Priority numbers control ordering. The orchestrator processes items lowest-priority-number-first, respecting dependencies. Core content pages and infrastructure (logo, design) complete first. Tools come next. Entity directory comes last.

The planner writes these as a Go action — one DB transaction inserting all items for the site with a shared `batch_id`. This replaces the current planner's JSON output.

### What the planner receives

The classifier's multi-dimensional output:

```json
{
  "build_approach": "page-assembled",
  "primary_archetype": "content-commerce-tools",
  "page_types": ["content", "tool", "commerce", "entity-directory"],
  "entity_types": ["species"],
  "detected_industry": "agriculture/farm-direct",
  "hosting_trajectory": "static_now_api_later"
}
```

Plus the reviewed brief (questionnaire answers, company info, etc.) and the available components/styles from the DB.

The planner uses this to decide what items to write — how many content pages, what tools, whether to set up entities. The LLM call still happens (to decide page structure, component selection, tool specs), but the output goes to the DB instead of being returned as JSON.

---

## The Classifier

### Updated Output

```json
{
  "build_approach": "page-assembled | application | hybrid",
  "primary_archetype": "descriptive label",
  "page_types": ["content", "tool", "commerce", "entity-directory", "news"],
  "entity_types": [],
  "hosting_trajectory": "static_only | static_now_api_later | needs_server",
  "confidence": 0.85,
  "reasoning": "brief explanation",
  "recommended_builder": "pageflow-builder",
  "detected_industry": "industry name",
  "detected_signals": ["signal1", "signal2"]
}
```

`recommended_builder` still exists for backward compatibility but becomes less important — the planner decides what agents handle what, not the classifier. For application-type sites (`build_approach: "application"`), the recommended_builder would be `app-builder` which has a different planning approach.

### HITL Confirmation

The human reviews the classification and can adjust:

- Build approach (page-assembled / application / hybrid)
- Page types present (add/remove from list)
- Entity types (add/remove)
- Recommended builder
- Industry detection

This is a config change to the intake-orchestrator's `hitl_confirm_type` step.

---

## Handler Agents (Fix Agents)

All handler agents follow the same contract:

```
Input:  work item (from site_work_items row)
        site context (site record, style collection, nav data, brief)
Output: result JSONB (includes commit_sha if page was deployed)
        status: 'complete' or 'failed'
```

The orchestrator passes the work item's `spec` field as the primary input. Each handler extracts what it needs.

### Existing agents, adapted as handlers

| Agent | Handles item_type | Notes |
|-------|------------------|-------|
| `page-content-writer` | `needs_content_page` | Existing. Receives page spec from item.spec. Outputs HTML. |
| `image-generator` | `needs_logo`, `needs_hero_image`, `needs_image` | Existing. Receives prompt from item.spec. |
| `webdesign-agent` | `needs_design`, `design_refresh` | Existing. Generates/updates CSS theme. |
| `deployer-agent` | `needs_deploy` | Existing. Triggers Cloudflare. |
| `content-reviewer` | (called within page build) | Existing. Review step before commit. |

### New agents from maintenance doc, also serve build

| Agent | Handles item_type | Phase |
|-------|------------------|-------|
| `section-rewriter` | `stale_date_reference`, `entity_data_drift`, `thin_content`, `stale_statistic` | Maintenance Phase 2 |
| `redirect-manager` | `redirect_chain`, `broken_internal_link` | Maintenance Phase 1 |
| `sitemap-regenerator` | `sitemap_out_of_sync`, `needs_sitemap` | Maintenance Phase 1 |
| `nav-updater` | `nav_item_orphaned`, `nav_update_needed` | Maintenance Phase 1 |
| `legal-updater` | `missing_disclaimer`, `outdated_legal_template` | Maintenance Phase 2 |
| `schema-fixer` | `invalid_schema`, `stale_meta_description` | Maintenance Phase 2 |
| `image-optimiser` | `image_not_optimised`, `missing_alt_text` | Maintenance Phase 2 |
| `css-patcher` | `css_variable_drift`, `missing_responsive_rule` | Maintenance Phase 2 |

### New agents for extended site types

| Agent | Handles item_type | Phase |
|-------|------------------|-------|
| `tool-builder` | `needs_tool_page`, `tool_improvement` | Build Phase 2 |
| `entity-page-builder` | `needs_entity_pages`, `entity_page_refresh` | Build Phase 3 |
| `directory-builder` | `needs_directory` | Build Phase 2 |
| `article-writer` | `publish_article` | Feed Phase 1 |
| `app-builder` | `needs_app_shell`, `needs_app_module` | Build Phase 4 |

All handlers output HTML that goes through the same post-processing: assemble (inject head/header/footer) → git commit → save sections. The orchestrator handles this uniformly after each handler completes.

---

## Git Commit Strategy

Individual commits per work item. Each commit is atomic and reversible.

```
abc1234 - Build index.html (planner: needs_content_page) [item: uuid]
def5678 - Build about.html (planner: needs_content_page) [item: uuid]
ghi9012 - Update styles.css (planner: needs_design) [item: uuid]
jkl3456 - Build tools/carbon-calc.html (planner: needs_tool_page) [item: uuid]
mno7890 - Fix hero date ref on index.html (discovery: stale_date_reference) [item: uuid]
pqr1234 - Publish fight-preview.html (content_feed: publish_article) [item: uuid]
```

Each commit message includes: what changed, why (source + item_type), and the work item ID for traceability.

The work item's `result` field stores `{"commit_sha": "abc1234"}`. If a commit needs reverting, the table provides the full audit trail of what was changed and why.

### Deploy: Commit IS Deploy

The deploy mechanism is GitHub Actions. Each git commit triggers an Action that writes the changed files to S3 (Backblaze B2). Cloudflare is DNS only — it points the domain at the S3 bucket. There is no separate deploy step.

This means every individual commit goes live automatically. There is no batching, no deploy trigger, no gap between commit and live. The git history and the live site are always in sync.

This reinforces the individual-commit-per-work-item approach — each commit is a small, safe change that goes live immediately. If something goes wrong, the specific commit can be reverted in git, the Action writes the reverted state to S3, and the site returns to the previous state. The work item is then marked as `failed` for retry or escalation.

No `deployer-agent` or `trigger_site_deploy` step is needed in the new system. The git commit action is the final step per work item.

---

## Post-Processing: The Assembly Step

After any handler produces HTML for a page, the orchestrator runs post-processing:

```
handler returns page HTML
  → assemble_page (inject head, header, footer from stored site chrome)
  → git_commit (individual commit with item reference)
  → save_page_sections (store rendered sections in page_components)
  → update work item status → 'complete'
  → update page build_status → 'deployed'
```

This uses the existing `assemble_page`, `git_commit`, and `save_page_sections` actions. Every handler agent's output goes through the same pipeline. The handler doesn't need to know about headers, footers, or git — it just produces the page body content.

For site-wide items (design updates, nav changes) the post-processing is different:
- Design changes → commit CSS → may trigger re-render of all pages
- Nav changes → update nav tables → re-render header/footer → re-assemble affected pages

The orchestrator knows which post-processing path to use based on the item's `domain` field.

---

## Discovery Agents

Unchanged from the maintenance doc. Each runs on its own schedule, writes items to `site_work_items` with `source: 'discovery'` and `status: 'detected'`.

| Agent | Domain | Checks | Schedule |
|-------|--------|--------|----------|
| `content-discovery-agent` | content | date refs, entity drift, thin content, stale stats | weekly |
| `links-discovery-agent` | links | internal links, external links, orphans, redirects | 8 hours |
| `seo-discovery-agent` | seo | sitemap sync, schema, meta freshness | weekly |
| `compliance-discovery-agent` | compliance | disclaimers, legal templates, tool compliance | monthly |
| `structural-discovery-agent` | structural | nav complexity, redundant content, competitors | monthly |
| `entity-data-agent` | entity | entity source changes, data drift | per entity source config |
| `content-feed-orchestrator` | content_feed | new articles, topic opportunities | per site feed config |

Discovery agents write items. They do not fix anything. They do not call other agents.

---

## Triage

Runs as a step within the site-work-orchestrator (step 5). Reads items with `status: 'detected'` and enriches them. Same logic as the maintenance doc:

1. Deduplication (item_key unique constraint + near-duplicate check)
2. Impact assessment (cross-reference links, nav, pages tables — read only)
3. Resolution classification (auto_fix / suggest / flag / monitor / ignore)
4. Handler routing (assign handler_agent based on item_type + suggested_action)
5. Priority scoring (severity × impact multiplier)

For build items from the planner, triage is pre-done — the planner sets handler_agent, priority, and resolution_path directly. Items arrive as 'triaged' not 'detected'.

---

## Cross-Domain Coordination

Same as maintenance doc. Agents communicate through the work items table, not by calling each other.

When a fix agent makes a change that affects other domains:
```
section-rewriter removes a paragraph containing a link
  → new item: domain='links', item_type='link_removed',
    parent_item_id=original_item, source='side_effect'

page-content-writer creates a new page
  → new item: domain='navigation', item_type='nav_update_needed',
    source='side_effect'
  → new item: domain='seo', item_type='needs_sitemap_update',
    source='side_effect'
```

Side-effect items get triaged and processed in the next cycle. No agent calls another agent for coordination.

---

## Cross-Site Operations

### Same change across multiple sites

When the same change needs to apply to many sites (e.g., update contact details, apply new legal template, fix a broken external URL), work items are created per site:

```sql
INSERT INTO site_work_items (site_id, source, domain, item_type, summary, spec, handler_agent)
SELECT s.id, 'manual', 'content', 'contact_page_update',
  'Update contact phone number',
  '{"new_data": {"phone": "+44..."}, "target_section": "contact-info"}'::jsonb,
  'section-rewriter'
FROM sites s
WHERE s.status = 'active'
  AND EXISTS (SELECT 1 FROM pages WHERE site_id = s.id AND name = 'contact');
```

Each site's orchestrator processes independently. The section-rewriter respects each site's design and tone when making the change.

### Cross-site pattern detection

The catch-all agent (daily cron) detects patterns:

```sql
SELECT item_type, affected_url, COUNT(DISTINCT site_id) as site_count
FROM site_work_items
WHERE status NOT IN ('verified', 'rejected', 'wont_fix')
GROUP BY item_type, affected_url
HAVING COUNT(DISTINCT site_id) > 1;
```

Same broken URL across 50 sites → fix once, apply to all via batch INSERT.

---

## Entity Data: Setup vs Sync

### Initial Setup (work items)

Entity types are set up as a dependency chain of work items:

```
1. needs_entity_source_config  → configure API/manual source
   handler: entity-data-agent
   priority: 50

2. needs_entity_data_fetch     → initial data population
   handler: entity-data-agent
   depends_on: [1]
   priority: 52

3. needs_entity_pages          → render pages from entity data
   handler: entity-page-builder
   depends_on: [2]
   priority: 55

4. needs_directory             → build index/listing page
   handler: directory-builder
   depends_on: [3]
   priority: 58
```

### Ongoing Sync (discovery agent)

After setup, the `entity-data-agent` runs as a discovery agent on a schedule. It checks configured sources for changes and writes work items when data drifts:

```
entity-data-agent (discovery mode):
  1. load entity sources for this site
  2. for each source: fetch current data, compare with stored entities
  3. for changed entities: write work item
     (item_type: 'entity_data_drift', handler: 'entity-page-builder')
  4. for new entities: write work item
     (item_type: 'needs_entity_page', handler: 'entity-page-builder')
  5. for removed entities: write work item
     (item_type: 'entity_removed', handler: nav-updater or manual)
```

### Real-time data

Fast-changing data (ticket prices, availability, live scores) is NOT handled through the work queue. Instead:

- Entity pages include client-side JS that fetches from a data API at view time
- The page structure (layout, static content) is built through the queue
- Dynamic data is served separately, outside the queue system
- This matches the hosting trajectory: static now, API-ready structure

---

## Intake Orchestrator: Updated Flow

```
Workflow:
  1. fetch_available_builders     → discover builder agents
  2. spawn_classifier             → spawn site-classifier
  3. call_classifier              → classify (multi-dimensional output)
  4. hitl_confirm_classification  → human reviews build_approach, page_types,
                                    entity_types, builder
  5. fetch_questionnaire          → get briefing questions for this site type
  6. spawn_briefer                → spawn briefing agent
  7. call_briefer                 → collect brief data
  8. hitl_review_brief            → human reviews brief answers
  9. ensure_site_record           → create/update site record
  10. call_planner                → planner writes work items to DB
  11. trigger_orchestrator        → spawn site-work-orchestrator
                                    with domains: ['build']
  12. complete
```

The key change: step 10 writes work items instead of returning a plan, and step 11 triggers the unified orchestrator instead of spawning a specific builder.

After initial build completes, the site enters the maintenance cycle. The batch scheduler's cron picks it up based on its maintenance profile.

---

## Per-Site Configuration

Unchanged from maintenance doc. Stored in `sites.settings.maintenance_profile`. Controls which domains run, at what cadence, with what parameters.

Extended with build-related config:

```json
{
  "maintenance_profile": {
    "build": {
      "current_tier": 2,
      "auto_advance_tiers": true,
      "planner_model": "claude-sonnet-4-5"
    },
    "content": {
      "enabled": true,
      "every": "7d",
      "agents": { "date_reference_scanner": true, ... },
      "auto_fix_enabled": false
    },
    "links": { ... },
    "seo": { ... },
    "compliance": { ... },
    "content_feed": {
      "enabled": false,
      "every": "24h",
      "max_articles_per_cycle": 3,
      "relevance_threshold": 0.6,
      "time_sensitivity": "normal"
    },
    "entity": {
      "enabled": false,
      "types": [],
      "sources": []
    },
    "budget": {
      "llm_calls_per_cycle": 20,
      "max_auto_fixes_per_cycle": 5
    }
  }
}
```

---

## Handling Design Consistency

A concern: when fix agents modify individual sections, the site's design must stay consistent. The approach:

1. **Site chrome (header, footer, head) is stored in the DB** via `page_components` site-level slots. Every page assembly reads from these stored components. No agent regenerates chrome unless explicitly told to.

2. **CSS theme is a single file** (`/assets/css/styles.css`) committed to git. Components use CSS variables from this theme. Changing a section's content doesn't change its styling — the CSS comes from the theme, not from the section HTML.

3. **Section reassembly** loads ALL sections for the page from `page_components`, replaces only the changed section, and re-assembles with current chrome. The page looks the same except for the intended change.

4. **Design changes are explicit work items** (`needs_design`, `design_refresh`). The webdesign-agent handles these. No other agent modifies CSS. If a section-rewriter's change looks wrong with the current design, that's a new finding (domain: 'design') for the next cycle.

5. **The render pipeline** (render_component → save to page_components → assemble_page) is shared infrastructure. Build and maintenance use the same functions. Fixing the assembly bugs (content_data contamination, head/header regex) fixes them for all paths.

---

## JavaScript Management

Sites use JavaScript from two sources:

### 1. JS Snippets Library (`js_snippets` table)

Pre-built, reusable JS that provides standard interactivity — mobile nav toggle, smooth scrolling, scroll spy, form validation, etc. Stored in `js_snippets` with semantic function names, categories, triggers, and dependencies.

These are selected during planning/build based on what the site needs:
- A site with a hamburger nav gets `nav-mobile-toggle`
- A landing page gets `nav-smooth-scroll`
- A site with forms gets `form-validation`

JS snippets are injected via the `<head>` component or inline `<script>` tags during page assembly. The `head-seo-standard` component (or site-specific head component) references which snippets a site uses. This is configured at the site level, not per-page.

Work items that affect JS:
- `needs_js_setup` — initial selection of JS snippets for a new site (handler: planner or webdesign-agent)
- `js_snippet_outdated` — a snippet in the library was updated, sites using it need re-deployment (cross-site pattern detection by catch-all)

### 2. Custom JS for Tools and Interactive Pages

Tool pages (calculators, converters) generate custom JS as part of their build. The `tool-builder` agent produces HTML + CSS + JS together as a self-contained page. This custom JS lives inline in the page HTML or as a page-specific script file committed to git.

For tool improvements over multiple iterations:
- First build: basic calculator JS
- Improvement item `tool_mobile_ux`: update the JS for better touch handling
- Improvement item `tool_accessibility`: add ARIA labels and keyboard nav
- Each improvement produces a new commit with updated page content

The tool's JS is part of the page content, managed through `page_components` content_data like any other section content. The section-rewriter or a specialised tool-improver agent can modify it.

### JS and the Component System

The `content_components` table defines component templates. Some components reference JS snippets they need (e.g., an accordion component needs accordion toggle JS). When a page uses that component, the assembly step ensures the required JS is included.

This is existing infrastructure — the component templates already declare their JS dependencies. No new mechanism needed. The work items model just needs to respect this: when a new component is used on a page, the assembly step handles JS inclusion automatically.

---

## Implementation Phases

### Phase 0 — Foundation (unified table + basic build)

```
1. site_work_items table (replacing both build plan and maintenance_findings)
2. content_feed_items table
3. site_work_items_archive table
4. Planner action: write work items to DB instead of returning JSON
5. site-work-orchestrator agent definition
6. Intake orchestrator: updated flow (trigger orchestrator instead of builder)
7. Existing page-content-writer adapted to accept work item input
8. Assembly bug fixes (content_data contamination, head/header regex, CSS loss)
```

Outcome: can build simple content sites through the queue. Existing sites continue working. No maintenance yet.

### Phase 1 — Maintenance Discovery + Simple Fixes

```
9. maintenance-batch-scheduler (cron-triggered)
10. content-discovery-agent (date refs only)
11. links-discovery-agent (internal links only)
12. seo-discovery-agent (sitemap sync only)
13. Triage step in orchestrator
14. redirect-manager fix agent
15. sitemap-regenerator fix agent
16. nav-updater fix agent
17. maintenance-catch-all (daily cron)
18. K8s CronJob manifests
```

Outcome: sites get basic maintenance. Findings written, triaged, simple fixes automated.

### Phase 2 — Tools + LLM Fixes

```
19. tool-builder agent (generates calculators/interactive tools)
20. directory-builder agent (generates index/listing pages)
21. section-rewriter fix agent
22. legal-updater, schema-fixer, image-optimiser
23. LLM-based discovery checks (entity drift, meta freshness, statistics)
24. compliance-discovery-agent
```

Outcome: can build mixed sites (content + tools). LLM-assisted maintenance.

### Phase 3 — Entities + Feeds

```
25. entity-data-agent (setup + sync modes)
26. entity-page-builder agent
27. content_sources configuration
28. content-feed-orchestrator
29. article-writer agent
30. Entity source integrations (boxing APIs, then football, then finance)
```

Outcome: can build entity-driven sites. News/article publishing pipeline.

### Phase 4 — Application Sites + Advanced

```
31. app-builder agent (blueprint-based)
32. Blueprint tables (site_blueprints, blueprint_reference_files)
33. structural-discovery-agent (competitor analysis)
34. Adopt/research pipeline
35. css-patcher fix agent
36. Analytics integration
```

Outcome: can build application-type sites. Competitive intelligence. Full maintenance.

---

## Key Design Principles

- Build and maintenance are the same queue, same orchestrator, same agents
- Every piece of work is a work item in `site_work_items`
- Each work item produces one git commit (atomic, reversible, traceable)
- Deploy is batched (one per orchestrator run), commits are individual
- Handler agents are narrow and focused — one concern each
- Agents communicate through the work items table, not by calling each other
- Content feeds have their own preprocessing table, bridge to work items when ready to publish
- Entity setup is work items with dependencies; entity sync is a discovery agent
- Real-time data is served outside the queue (client-side JS + API)
- Discovery agents find problems. Triage classifies. Fix agents execute. Clean separation.
- The site is always in a working state between work items
- Cross-site operations are batches of per-site work items
- Archival keeps the active table small
- Per-site configuration controls what runs, when, and at what budget

---

## Relationship to Existing Code

### What stays
- `page-content-writer` — adapted to accept work item input
- `image-generator` — unchanged, called as handler
- `webdesign-agent` — unchanged, called as handler
- `deployer-agent` — unchanged, called as handler
- `assemble_page` action — unchanged (after bug fixes)
- `git_commit` action — unchanged
- `save_page_sections` action — unchanged
- `populate_nav_tables` action — unchanged
- `render_site_components` action — unchanged
- CSS themes, style collections, content_components, page_components — all unchanged
- Site record, pages table, nav tables, link_registry — all unchanged

### What changes
- Site classifier — prompt change for multi-dimensional output (new agent: `site-classifier-v2`)
- HITL confirmation — richer fields for classification review
- Site planner — new agent (`site-planner-v2`) that writes to DB instead of returning JSON
- Intake orchestrator — new agent (`intake-orchestrator-v2`) routes to either pageflow-builder or site-work-orchestrator
- `maintenance_queue` — eventually superseded by `site_work_items` (existing `maintenance-triage` + `page-rebuild` keep working until migrated)

### What's new
- `site_work_items` table (runs alongside existing `maintenance_queue`)
- `site_work_items_archive` table
- `content_feed_items` table
- `site-work-orchestrator` agent (the new unified build/maintenance orchestrator)
- `intake-orchestrator-v2` agent
- `site-classifier-v2` agent
- `site-planner-v2` agent (writes work items to DB)
- `maintenance-batch-scheduler` agent (replaces maintenance-triage over time)
- `maintenance-catch-all` agent
- Planner DB-write action (Go code to INSERT work items)
- New handler agents as listed in phases above

---

## Open Questions — with current positions

1. **Existing `maintenance_queue` table.** It exists and is in use by `maintenance-triage` + `page-rebuild`. Columns: `site_id, task_type, priority, reason, payload, requested_by, status, claimed_by, claimed_at, retry_count, max_retries`. The new `site_work_items` table runs alongside it. No migration needed initially — new sites use new table, existing maintenance pipeline keeps working.

2. **Planner complexity for large sites.** Two-pass approach: first pass outputs site structure (page list with types and priorities), second pass generates detail per page just-in-time (tool specs, entity configs generated by the handler agent when processing the work item, not by the planner up front). Keeps each step simple and frequent rather than bulky.

3. **Work item granularity.** Per-page for build, per-section for maintenance. Both coexist in the same table. A build item says "create about page with these sections." A maintenance item says "rewrite the hero section on the about page." Different granularity, same table, same orchestrator.

4. **Handler agent spawning.** Dynamic, independent agents — spawned per orchestrator run as needed, not persistent. Each has its own K8s job, own logs, own responsibility. If the number of simultaneous agents becomes a concern, address it then. The system handles many concurrent jobs well.

5. **Verification and approval.** Uses the existing HITL loop. For new pages/tools: the item is deployed to the site but not linked in navigation (deployed but unlisted). HITL notification goes to the user for approval. Default behaviour is auto-approve after a timeout. When the human gets round to reviewing and wants changes, those become new work items in the queue. For maintenance fixes: similar pattern — fix is deployed, finding marked `pending_verify`, next cycle verifies or HITL reviews.

6. **Budget accounting.** Will become increasingly important, especially with looping behaviour that could call LLMs excessively. Needs careful design — probably a shared counter in the orchestration state checked before each LLM call, with hard limits per cycle per site. Deferred to implementation but flagged as a design concern that needs attention before scaling.

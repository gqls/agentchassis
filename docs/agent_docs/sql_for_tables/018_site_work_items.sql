-- 023_site_work_items.sql
-- Unified build and maintenance work queue
-- Runs alongside existing maintenance_queue (no migration needed)

-- =============================================================================
-- MAIN TABLE: site_work_items
-- =============================================================================
-- Every piece of work — building a new page, fixing stale content, adding a
-- tool, publishing a news article — is a work item. Processed by
-- site-work-orchestrator via handler agents.

CREATE TABLE IF NOT EXISTS site_work_items (
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
-- 'publish_article', 'entity_data_drift', etc.
    severity TEXT NOT NULL DEFAULT 'medium', -- 'info', 'low', 'medium', 'high', 'urgent'
    summary TEXT NOT NULL,                   -- human-readable description
    spec JSONB DEFAULT '{}',                 -- work-specific structured data

-- What it affects
    page_id UUID,                            -- NULL for site-wide items
    component_id UUID,                       -- specific page_component if applicable
    entity_id UUID,                          -- specific entity if applicable
    affected_url TEXT,                       -- the URL or resource in question

-- Triage enrichment
    impact JSONB,                            -- inbound links, nav membership, traffic
    resolution_path TEXT,                    -- 'auto_fix', 'suggest', 'flag',
-- 'monitor', 'ignore'
    suggested_action TEXT,                   -- 'build_page', 'rewrite_section', etc.
    priority INTEGER DEFAULT 100,            -- lower number = higher priority
-- which agent type processes this. NOT NULL DEFAULT '' since migration 217
-- (bugs_closed/078): nullable, one hand-written INSERT that omitted this column
-- livelocked the fleet build dispatcher — find_dispatchable_site counted the
-- row while LoadWorkItemsAction dropped it on a failed scan, so the trigger
-- re-picked the same site every 120s forever. '' is the safe spelling of "no
-- handler" (flag-only / human-review items) and was already the majority one.
    handler_agent TEXT NOT NULL DEFAULT '',

-- Lifecycle
    status TEXT NOT NULL DEFAULT 'detected', -- 'detected', 'triaged', 'approved',
-- 'claimed', 'in_progress',
-- 'complete', 'pending_verify',
-- 'verified', 'failed',
-- 'rejected', 'wont_fix'
    created_by TEXT NOT NULL,                -- agent type or 'manual'
    handled_by TEXT,                         -- agent instance that processed it
    approved_by TEXT,                        -- 'auto' or user identifier
    claimed_by TEXT,                         -- orchestration_id that claimed this

-- Dependencies and relationships
    depends_on UUID[],                       -- items that must complete first
    parent_item_id UUID REFERENCES site_work_items(id),
    related_item_ids UUID[],                 -- items to consider together
    batch_id UUID,                           -- groups items from same planning/discovery run

-- Tracking
    attempt_count INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    result JSONB DEFAULT '{}',               -- what the handler produced (includes commit_sha)
    error TEXT,

    -- Deduplication
    item_key TEXT,                            -- deterministic key for dedup
-- e.g. 'needs_page:about' or 'stale_date:page_id:comp_id'

    created_at TIMESTAMPTZ DEFAULT NOW(),
    triaged_at TIMESTAMPTZ,
    claimed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
    );

-- Unique constraint for deduplication (only one active item per key per site)
CREATE UNIQUE INDEX IF NOT EXISTS idx_swi_dedup
    ON site_work_items(site_id, item_key)
    WHERE item_key IS NOT NULL
    AND status NOT IN ('complete', 'verified', 'rejected', 'wont_fix', 'failed');

-- Primary query: orchestrator loading pending items for a site
CREATE INDEX IF NOT EXISTS idx_swi_site_pending
    ON site_work_items(site_id, priority)
    WHERE status IN ('triaged', 'approved');

-- Claimed items (for timeout/stuck detection)
CREATE INDEX IF NOT EXISTS idx_swi_claimed
    ON site_work_items(status, claimed_at)
    WHERE status = 'claimed';

-- Handler routing
CREATE INDEX IF NOT EXISTS idx_swi_handler
    ON site_work_items(handler_agent, status)
    WHERE status IN ('triaged', 'approved');

-- Batch grouping
CREATE INDEX IF NOT EXISTS idx_swi_batch
    ON site_work_items(batch_id)
    WHERE batch_id IS NOT NULL;

-- Page-specific lookups
CREATE INDEX IF NOT EXISTS idx_swi_page
    ON site_work_items(page_id)
    WHERE page_id IS NOT NULL;

-- Dependency resolution (GIN index on UUID array)
CREATE INDEX IF NOT EXISTS idx_swi_deps
    ON site_work_items USING GIN(depends_on)
    WHERE depends_on IS NOT NULL;

-- Archival candidates
CREATE INDEX IF NOT EXISTS idx_swi_completed
    ON site_work_items(completed_at)
    WHERE status IN ('complete', 'verified', 'rejected', 'wont_fix');

-- Status overview per site
CREATE INDEX IF NOT EXISTS idx_swi_site_status
    ON site_work_items(site_id, status);


-- =============================================================================
-- ARCHIVE TABLE: site_work_items_archive
-- =============================================================================
-- Same structure, stores completed items after 90 days.
-- Catch-all agent moves items here periodically.

CREATE TABLE IF NOT EXISTS site_work_items_archive (
                                                       LIKE site_work_items INCLUDING ALL
);

-- Archive only needs site+date lookups
CREATE INDEX IF NOT EXISTS idx_swia_site
    ON site_work_items_archive(site_id, completed_at DESC);
CREATE INDEX IF NOT EXISTS idx_swia_batch
    ON site_work_items_archive(batch_id)
    WHERE batch_id IS NOT NULL;


-- =============================================================================
-- CONTENT FEED TABLE: content_feed_items
-- =============================================================================
-- Separate lifecycle from work items. Feed items go through ingestion,
-- filtering, deduplication before becoming publishable.
-- When ready to publish, a work item is created in site_work_items.

CREATE TABLE IF NOT EXISTS content_feed_items (
                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id),       -- NULL if not yet assigned

-- Source data
    source_id UUID,                          -- references content_sources config
    external_id TEXT,                         -- ID from source (RSS guid, API id)
    source_url TEXT,
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
    published_page_id UUID,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
    );

CREATE INDEX IF NOT EXISTS idx_cfi_site_status
    ON content_feed_items(site_id, status);
CREATE INDEX IF NOT EXISTS idx_cfi_source
    ON content_feed_items(source_id, external_id);
CREATE INDEX IF NOT EXISTS idx_cfi_relevance
    ON content_feed_items(site_id, relevance_score DESC)
    WHERE status = 'relevant';
CREATE INDEX IF NOT EXISTS idx_cfi_dedup
    ON content_feed_items(source_url)
    WHERE status NOT IN ('duplicate', 'expired', 'rejected');


--

-- ============================================================================
-- Site Quality Fixes — Immediate Data Patches
-- Run in order. Each section is independent.
-- ============================================================================

-- ============================================================================
-- ISSUE 1A: Populate site metadata from content_data
-- These fields feed into loadSiteDataFull → render context → header/footer/head templates
-- ============================================================================

-- Finetuning.uk
UPDATE sites SET
                 company_name = 'FineTuning',
                 name = 'FineTuning',
                 tagline = 'AI for the Rest of Us',
                 logo_url = '/assets/images/logo.png',
                 logo_text = 'FineTuning'
WHERE id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND (company_name IS NULL OR company_name = '');

-- Gaswholesalers.com (company_name already set)
UPDATE sites SET
                 tagline = 'Wholesale Gas Supply Solutions',
                 logo_url = '/assets/images/logo.png',
                 logo_text = 'Gas Wholesalers'
WHERE id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND (logo_url IS NULL OR logo_url = '');

-- ============================================================================
-- ISSUE 5: Fix mismatched slot_names on gaswholesalers.com contact page
-- The page_components have wrong slot_names relative to their actual content.
-- The work items reference 'contact-info' which matches the HTML data-component
-- but not the DB slot_name.
-- Fix: align slot_names with the HTML data-component attribute values.
-- ============================================================================

-- Component 1: hero-contact (slot_name is NULL, should be 'hero-contact')
UPDATE page_components SET slot_name = 'hero-contact'
WHERE id = 'd7c96e1b-a549-4435-8bac-682d8d8dc678'
  AND slot_name IS NULL;

-- Component 2: slot_name is 'contact-hero' but HTML says 'contact-form'
UPDATE page_components SET slot_name = 'contact-form'
WHERE id = '4b3bafce-dbf8-49c9-9b4a-84a0bc00a7aa'
  AND slot_name = 'contact-hero';

-- Component 3: slot_name is 'contact-form' but HTML says 'contact-info'
UPDATE page_components SET slot_name = 'contact-info'
WHERE id = 'fc5c9dac-2058-45ec-8c26-94de257aa31c'
  AND slot_name = 'contact-form';

-- ============================================================================
-- ISSUE 5b: Fail/remove the orphaned content_edit work items if they still
-- reference components that don't exist after the slot_name fix.
-- After the fix above, 'contact-info' should now exist.
-- Reset the failed items so they can be retried.
-- ============================================================================

-- Reset content_edit items that failed due to missing component
-- (they were marked 'failed' or 'triaged' with attempt_count > 0)
UPDATE site_work_items
SET status = 'triaged',
    attempt_count = 0,
    error = NULL
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND item_type = 'content_edit'
  AND status IN ('failed', 'claimed')
  AND spec->>'slot_name' = 'contact-info';

-- ============================================================================
-- ISSUE: Clean up stuck claimed items from previous runs
-- ============================================================================

UPDATE site_work_items
SET status = 'triaged', claimed_at = NULL, claimed_by = NULL
WHERE status = 'claimed'
  AND claimed_at < NOW() - INTERVAL '10 minutes';

-- ============================================================================
-- Verify
-- ============================================================================

-- Check site metadata
SELECT domain, company_name, tagline, logo_url
FROM sites
WHERE id IN ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', '5fe15466-4e2e-4ff2-981e-98c1b7074002');

-- Check contact page slot_names
SELECT pc.slot_name, pc.id,
       substring(pc.rendered_html from 'data-component="([^"]*)"') as data_component
FROM page_components pc
         JOIN pages p ON pc.page_id = p.id
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'contact'
ORDER BY pc.position;

-- Check remaining triaged items
SELECT wi.item_type, wi.status, s.domain, wi.spec->>'slot_name' as slot
FROM site_work_items wi
    JOIN sites s ON s.id = wi.site_id
WHERE wi.status = 'triaged' AND wi.domain = 'build'
ORDER BY s.domain, wi.priority;

----

-- increase throughput

-- Blanket reset for items that failed before today's fixes
UPDATE site_work_items
SET attempt_count = 0
WHERE status = 'triaged'
  AND domain = 'build'
  AND attempt_count >= max_attempts;

-- Also increase max_attempts for content items (they had legit bugs, not real failures)
UPDATE site_work_items
SET max_attempts = 5
WHERE domain = 'build'
  AND item_type IN ('content_rewrite', 'needs_content_planning', 'tone_shift', 'needs_content_page');

---

-- timeouts

-- 1. Increase dispatch loop idle timeout — it waits for slow handlers
UPDATE agent_definitions
SET idle_timeout_seconds = 600
WHERE type = 'build-dispatch-loop' AND deleted_at IS NULL;

-- 2. Reset the current stale claims so the pipeline can continue
UPDATE site_work_items
SET status = 'triaged', claimed_by = NULL, claimed_at = NULL
WHERE status = 'claimed' AND domain = 'build';

-- 3. Force the trigger to fire on next tick
UPDATE scheduled_tasks
SET last_completed_at = NOW()
WHERE name = 'build-pipeline-trigger';

-- 4. Fix the claimed-item-timeout not firing
-- It hasn't run since March 9. Check if it's stuck behind the no-refire guard:
-- fire_message=false should bypass the guard. But last_triggered_at isn't updating
-- which means the pre_query is returning no rows every tick.
-- Force a trigger to verify:
UPDATE scheduled_tasks
SET last_triggered_at = NOW() - INTERVAL '5 minutes'
WHERE name = 'claimed-item-timeout';

-- 5. Mark the 3 exhausted leopardess items as failed — they'll never succeed
-- (webdesign-agent doesn't handle needs_design_review properly)
UPDATE site_work_items
SET status = 'failed', error = 'webdesign-agent cannot handle needs_design_review'
WHERE status = 'triaged'
  AND domain = 'build'
  AND attempt_count >= max_attempts;

--

-- ============================================================================
-- Find all deployed sections containing placeholder text across all sites
-- ============================================================================

-- 1. Check the scope
SELECT s.domain, p.name as page, pc.position,
    LEFT(pc.rendered_html, 200) as preview
FROM page_components pc
    JOIN pages p ON pc.page_id = p.id
    JOIN sites s ON p.site_id = s.id
WHERE pc.rendered_html IS NOT NULL
  AND (
    pc.rendered_html ILIKE '%NEEDS HUMAN REVIEW%'
   OR pc.rendered_html ILIKE '%NEEDS_HUMAN_REVIEW%'
   OR pc.rendered_html ILIKE '%Lorem ipsum%'
   OR pc.rendered_html ILIKE '%[INSERT %'
   OR pc.rendered_html ILIKE '%[YOUR %'
   OR pc.rendered_html ILIKE '%<no value>%'
    )
ORDER BY s.domain, p.name, pc.position;

-- 2. Suppress affected sections (replace with hidden comment)
UPDATE page_components pc
SET rendered_html = '<!-- Section hidden: contains placeholder content (needs human review) -->',
    build_status = 'pending'
    FROM pages p, sites s
WHERE pc.page_id = p.id
  AND p.site_id = s.id
  AND pc.rendered_html IS NOT NULL
  AND (
    pc.rendered_html ILIKE '%NEEDS HUMAN REVIEW%'
   OR pc.rendered_html ILIKE '%NEEDS_HUMAN_REVIEW%'
   OR pc.rendered_html ILIKE '%Lorem ipsum%'
    );

-- 3. Create needs_human_review work items for each affected page
-- (one per page, not per section — human reviews the whole page)
INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
)
SELECT DISTINCT ON (s.id, p.name)
    s.id,
    'validation',
    'build',
    'placeholder_content',
    'medium',
    FORMAT('Page %s has placeholder content that needs real data', p.name),
    jsonb_build_object(
    'page_name', p.name,
    'page_id', p.id,
    'domain', s.domain,
    'missing_data', 'Team member names, titles, photos, bios or other section-specific data',
    'fix_guidance', 'Provide real data for this section. It is hidden on the live site until resolved.'
    ),
    40,
    'human-review',
    'needs_human_review',
    'validation',
    FORMAT('placeholder_%s_%s', p.name, s.id)
FROM page_components pc
    JOIN pages p ON pc.page_id = p.id
    JOIN sites s ON p.site_id = s.id
WHERE pc.rendered_html = '<!-- Section hidden: contains placeholder content (needs human review) -->'
ON CONFLICT DO NOTHING;

-- 4. Trigger rerender for affected pages
INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
)
SELECT DISTINCT ON (s.id)
    s.id,
    'validation',
    'build',
    'needs_rerender',
    'medium',
    FORMAT('Re-render %s pages after placeholder content removal', s.domain),
    jsonb_build_object(
    'refresh_site_components', true,
    'reason', 'Placeholder content suppressed — pages need reassembly'
    ),
    50,
    'rerender-pages',
    'triaged',
    'validation',
    FORMAT('rerender_placeholder_fix_%s', s.id)
FROM page_components pc
    JOIN pages p ON pc.page_id = p.id
    JOIN sites s ON p.site_id = s.id
WHERE pc.rendered_html = '<!-- Section hidden: contains placeholder content (needs human review) -->'
ON CONFLICT DO NOTHING;

---

ALTER TABLE site_work_items
    ADD COLUMN processing_tier TEXT NOT NULL DEFAULT 'standard';

-- Values:
-- 'standard'  — process immediately using whatever ai_service is configured (Claude or CPU Ollama)
-- 'batch_gpu' — hold until GPU batch starts, then process via GPU Ollama

ALTER TABLE site_work_items DROP COLUMN domain;

-- ============================================================================
-- 028_news_feed_handler_routing_fixes.sql
--
-- Fixes three issues:
-- 1. Existing work items with wrong handler_agent values
-- 2. Ensures content-feed-refresh scheduled task exists
-- 3. Render query improvement: prefer scored items over raw ingested
-- ============================================================================

-- ============================================================================
-- 1. Fix existing work items with wrong handler_agent
-- ============================================================================

-- missing_news_sources: was page-build-handler, should be content-feed-orchestrator
-- The content-feed-orchestrator's first step (seed_content_sources) creates
-- content_sources rows from the classification spec.
UPDATE site_work_items
SET handler_agent = 'content-feed-orchestrator',
    updated_at = NOW()
WHERE item_type = 'missing_news_sources'
  AND handler_agent = 'page-build-handler'
  AND status IN ('detected', 'triaged', 'approved');

-- missing_news_section: was page-build-handler, should be content-gap-planner
-- The gap planner reads the spec, decides how to add latest-news to the
-- homepage, and creates targeted work items for page-build-handler.
UPDATE site_work_items
SET handler_agent = 'content-gap-planner',
    updated_at = NOW()
WHERE item_type = 'missing_news_section'
  AND handler_agent = 'page-build-handler'
  AND status IN ('detected', 'triaged', 'approved');

-- stale_news_section: was rerender-pages, should be content-feed-orchestrator
-- Stale news means we need fresh items from sources, not a page rerender.
UPDATE site_work_items
SET handler_agent = 'content-feed-orchestrator',
    updated_at = NOW()
WHERE item_type = 'stale_news_section'
  AND handler_agent = 'rerender-pages'
  AND status IN ('detected', 'triaged', 'approved');

-- all_sources_erroring: was "" (empty), should be content-feed-orchestrator
-- Empty handler_agent would crash the dispatch loop's spawn_handler step.
UPDATE site_work_items
SET handler_agent = 'content-feed-orchestrator',
    updated_at = NOW()
WHERE item_type = 'all_sources_erroring'
  AND (handler_agent IS NULL OR handler_agent = '')
  AND status IN ('detected', 'triaged', 'approved');

-- ============================================================================
-- 2. Ensure content-feed-refresh scheduled task exists
-- ============================================================================
-- The content-feed-trigger agent updates this task's last_completed_at on
-- each run. The scheduler checks interval_seconds against last_triggered_at.

INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type,
    target_topic, enabled, fire_message, timeout_seconds,
    concurrency_group, max_concurrent, pre_query
) VALUES (
             'content-feed-refresh',
             'Triggers content-feed-trigger agent every 6 hours to refresh news feeds for all recommended sites',
             21600,          -- 6 hours = 21600 seconds
             'content-feed-trigger',
             'system.agent.generic.requests',
             true,
             true,           -- fire Kafka message to trigger agent
             600,            -- 10 minute timeout
             'content-feed', -- concurrency group
             1,              -- max 1 concurrent
             NULL            -- no gating query — the agent itself checks which sites need refresh
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              enabled = EXCLUDED.enabled,
                              updated_at = NOW();

-- ============================================================================
-- 3. Verify: show current news-related work items and their handlers
-- ============================================================================
-- Run this after the above to confirm the fixes took effect:
--
-- SELECT item_type, handler_agent, status, COUNT(*)
-- FROM site_work_items
-- WHERE item_type IN (
--     'missing_news_sources', 'missing_news_section',
--     'stale_news_section', 'all_sources_erroring'
-- )
-- GROUP BY item_type, handler_agent, status
-- ORDER BY item_type, status;

-- ============================================================================
-- 4. Verify: scheduled tasks for news
-- ============================================================================
-- SELECT name, enabled, interval_seconds, target_agent_type,
--        last_triggered_at, last_completed_at
-- FROM scheduled_tasks
-- WHERE name = 'content-feed-refresh';

--

-- ============================================================================
-- 028_news_feed_handler_routing_fixes.sql
--
-- Fixes three issues:
-- 1. Existing work items with wrong handler_agent values
-- 2. Ensures content-feed-refresh scheduled task exists
-- 3. Render query improvement: prefer scored items over raw ingested
-- ============================================================================

-- ============================================================================
-- 1. Fix existing work items with wrong handler_agent
-- ============================================================================

-- missing_news_sources: was page-build-handler, should be content-feed-orchestrator
-- The content-feed-orchestrator's first step (seed_content_sources) creates
-- content_sources rows from the classification spec.
UPDATE site_work_items
SET handler_agent = 'content-feed-orchestrator',
    updated_at = NOW()
WHERE item_type = 'missing_news_sources'
  AND handler_agent = 'page-build-handler'
  AND status IN ('detected', 'triaged', 'approved');

-- missing_news_section: was page-build-handler, should be content-gap-planner
-- The gap planner reads the spec, decides how to add latest-news to the
-- homepage, and creates targeted work items for page-build-handler.
UPDATE site_work_items
SET handler_agent = 'content-gap-planner',
    updated_at = NOW()
WHERE item_type = 'missing_news_section'
  AND handler_agent = 'page-build-handler'
  AND status IN ('detected', 'triaged', 'approved');

-- stale_news_section: was rerender-pages, should be content-feed-orchestrator
-- Stale news means we need fresh items from sources, not a page rerender.
UPDATE site_work_items
SET handler_agent = 'content-feed-orchestrator',
    updated_at = NOW()
WHERE item_type = 'stale_news_section'
  AND handler_agent = 'rerender-pages'
  AND status IN ('detected', 'triaged', 'approved');

-- all_sources_erroring: was "" (empty), should be content-feed-orchestrator
-- Empty handler_agent would crash the dispatch loop's spawn_handler step.
UPDATE site_work_items
SET handler_agent = 'content-feed-orchestrator',
    updated_at = NOW()
WHERE item_type = 'all_sources_erroring'
  AND (handler_agent IS NULL OR handler_agent = '')
  AND status IN ('detected', 'triaged', 'approved');

-- ============================================================================
-- 2. Ensure content-feed-refresh scheduled task exists
-- ============================================================================
-- The content-feed-trigger agent updates this task's last_completed_at on
-- each run. The scheduler checks interval_seconds against last_triggered_at.

INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type,
    target_topic, enabled, fire_message, timeout_seconds,
    concurrency_group, max_concurrent, pre_query
) VALUES (
             'content-feed-refresh',
             'Triggers content-feed-trigger agent every 6 hours to refresh news feeds for all recommended sites',
             21600,          -- 6 hours = 21600 seconds
             'content-feed-trigger',
             'system.agent.generic.requests',
             true,
             true,           -- fire Kafka message to trigger agent
             600,            -- 10 minute timeout
             'content-feed', -- concurrency group
             1,              -- max 1 concurrent
             NULL            -- no gating query — the agent itself checks which sites need refresh
         )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              enabled = EXCLUDED.enabled,
                              updated_at = NOW();

-- ============================================================================
-- 3. Verify: show current news-related work items and their handlers
-- ============================================================================
-- Run this after the above to confirm the fixes took effect:
--
-- SELECT item_type, handler_agent, status, COUNT(*)
-- FROM site_work_items
-- WHERE item_type IN (
--     'missing_news_sources', 'missing_news_section',
--     'stale_news_section', 'all_sources_erroring'
-- )
-- GROUP BY item_type, handler_agent, status
-- ORDER BY item_type, status;

-- ============================================================================
-- 4. Verify: scheduled tasks for news
-- ============================================================================
-- SELECT name, enabled, interval_seconds, target_agent_type,
--        last_triggered_at, last_completed_at
-- FROM scheduled_tasks
-- WHERE name = 'content-feed-refresh';

-- ============================================================================
-- 075a: Team data for all three sites
-- ============================================================================

-- ai-agent-orchestration.com
UPDATE site_specs
SET data = jsonb_set(data, '{team}', '[
    {
        "name": "Peter Grenfell",
        "title": "Principal Consultant",
        "bio": "Over 30 years of engineering experience across startups, scale-ups, and enterprise systems. Built and operated worldsoccernews.com — one of the world''s busiest sports sites, holding top search engine placement for highly competitive terms for over a decade. Former senior engineer at Bumble. Strong financial systems background. Designed this multi-agent orchestration framework from the conviction that many specialised AI agents collaborating through robust infrastructure outperform any single-model approach. The framework now coordinates 30+ agent types on Kubernetes and Kafka, managing thousands of concurrent agent instances in production across websites, data pipelines, and research systems.",
        "photo": ""
    },
    {
        "name": "Archivist",
        "title": "Head of Research",
        "bio": "AI managing agent. Leads a team of specialist research agents that investigate before anything gets built or published. Depending on the task, Archivist coordinates deep web research, source credibility verification, structured data extraction from industry registries, Companies House cross-referencing, and competitive analysis. Whether the job is grounding a content page in verified facts or building a dataset of 10,000 businesses with financial enrichment, Archivist''s team ensures the output is sourced, attributed, and accurate.",
        "photo": ""
    },
    {
        "name": "Sentinel",
        "title": "Head of Quality Assurance",
        "bio": "AI managing agent. Coordinates a team of specialist auditors that continuously monitor everything we deploy — websites, data pipelines, interactive tools, news feeds, and content accuracy. Each auditor focuses on a different dimension: structural integrity, content quality, design consistency, factual accuracy, and data pipeline health. Findings are classified by severity and routed to the right specialist for resolution. Full audit trails on every finding and fix.",
        "photo": ""
    },
    {
        "name": "Quartermaster",
        "title": "Head of Operations",
        "bio": "AI managing agent. Runs the ongoing improvement of every deployed system. Quartermaster coordinates maintenance cycles across a team of specialist agents — each responsible for a different concern: content freshness, tool health, navigation structure, news feed reliability, data quality, and deployment integrity. Every improvement cycle is observable and every change can be routed through human review before it takes effect. Systems get better between review cycles, with routine maintenance handled autonomously.",
        "photo": ""
    }
]'::jsonb)
WHERE site_id = (SELECT id FROM sites WHERE domain = 'ai-agent-orchestration.com')
  AND aspect = 'identity' AND is_current = true;

-- finetuning.uk
UPDATE site_specs
SET data = jsonb_set(data, '{team}', '[
    {
        "name": "Peter Grenfell",
        "title": "Principal Consultant",
        "bio": "30+ years building software — from running one of the world''s busiest sports websites to senior engineering at Bumble to financial systems. Built this AI framework because practical AI should be accessible to every business, not just those with engineering teams. The system now runs dozens of specialised AI agents that handle research, content, design, data collection, and ongoing maintenance — so business owners can focus on their business.",
        "photo": ""
    },
    {
        "name": "Archivist",
        "title": "Head of Research",
        "bio": "AI managing agent. Leads a team of research specialists that do the homework before anything gets written or published. Depending on what''s needed, Archivist''s team researches your industry from multiple sources, checks facts against official records, verifies business data, and builds a clear picture of what your customers need to know. Your content and data are grounded in real, verified information — not generic filler.",
        "photo": ""
    },
    {
        "name": "Sentinel",
        "title": "Head of Quality",
        "bio": "AI managing agent. Runs a team of specialist checkers that continuously review everything we build and maintain — your website, your content, your tools, and any data we collect on your behalf. Each checker focuses on something different: is the content accurate, do the tools work, is the design consistent, is the data clean. When something needs attention, Sentinel routes it to the right place. You can review changes before they go live, or let routine fixes happen automatically.",
        "photo": ""
    },
    {
        "name": "Quartermaster",
        "title": "Head of Maintenance",
        "bio": "AI managing agent. Keeps everything we''ve built improving over time instead of slowly going stale. Quartermaster coordinates a team of maintenance specialists — monitoring content freshness, tool health, data quality, and site structure. Each maintenance cycle leaves your systems measurably better than before. You decide what gets your approval and what''s handled automatically.",
        "photo": ""
    }
]'::jsonb)
WHERE site_id = (SELECT id FROM sites WHERE domain = 'finetuning.uk')
  AND aspect = 'identity' AND is_current = true;

-- leopardessconsulting.co.uk
UPDATE site_specs
SET data = jsonb_set(data, '{team}', '[
    {
        "name": "Peter Grenfell",
        "title": "Principal Consultant",
        "bio": "Over 30 years of engineering experience. Built and operated worldsoccernews.com with sustained top search placement for competitive terms across a decade. Former senior engineer at Bumble. Strong financial systems background. Designed this multi-agent orchestration framework on Kubernetes and Kafka from first principles — workflow-driven state machines with Postgres persistence, coordinating 30+ specialised agent types across websites, data pipelines, news aggregation, and research systems. The architecture: each named agent here is itself an orchestrator managing teams of specialist agents, not a single prompt, and not just prompting.",
        "photo": ""
    },
    {
        "name": "Archivist",
        "title": "Head of Research",
        "bio": "AI managing agent. Orchestrates a team of specialist research agents across multiple data acquisition strategies: deep web search with source attribution, structured extraction from industry registries, Companies House verification cascades with financial enrichment, news aggregation with credibility scoring, and competitive analysis. The research layer runs before any content generation or data publication — every output is sourced and verifiable. The same research infrastructure that grounds website content also powers standalone data collection pipelines.",
        "photo": ""
    },
    {
        "name": "Sentinel",
        "title": "Head of Quality Assurance",
        "bio": "AI managing agent. Coordinates a team of specialist auditors across every deployed system — content quality, structural integrity, tool functionality, CSS consistency, data pipeline health, and deployment status. Each auditor runs independently with its own detection logic and severity classification. Findings route to specialist fix agents or escalate to human review based on confidence. Full audit trails with processing history on every finding, fix, and verification.",
        "photo": ""
    },
    {
        "name": "Quartermaster",
        "title": "Head of Operations",
        "bio": "AI managing agent. Runs continuous improvement cycles across all deployed infrastructure. Coordinates specialist agents for content freshness, tool health, navigation integrity, news feed reliability, data quality monitoring, and deployment verification. Each cycle is fully observable — work items track every finding through triage, fix, and verification. Human review gates configurable per workflow stage. The system improves autonomously between human review cycles.",
        "photo": ""
    }
]'::jsonb)
WHERE site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk')
  AND aspect = 'identity' AND is_current = true;

-- ============================================================================
-- 075b: Remove pricing sections, replace with engagement process
-- ============================================================================

-- Remove 'pricing' from sections arrays and replace with a second generic-text-block
-- Actually, just removing pricing is cleaner — the page still has hero, text block,
-- features, faq, and CTA which is a good engagement page structure.

UPDATE pages SET
    sections = (
        SELECT jsonb_agg(sec)
        FROM jsonb_array_elements_text(sections) sec
        WHERE sec != 'pricing'
    )::jsonb,
    build_status = 'needs_rebuild'
WHERE site_id IN (SELECT id FROM sites WHERE domain IN ('ai-agent-orchestration.com', 'finetuning.uk', 'leopardessconsulting.co.uk'))
  AND name IN ('pricing', 'how-we-work', 'engagement-model')
  AND status = 'active';

-- ============================================================================
-- 075c: Clear pricing needs_section_data items
-- ============================================================================

UPDATE site_work_items
SET status = 'wont_fix', completed_at = NOW(),
    result = '{"reason": "pricing section removed — replaced with engagement process content"}'::jsonb
WHERE status = 'needs_human_review'
  AND item_type = 'needs_section_data'
  AND spec->>'section_name' = 'pricing';

-- ============================================================================
-- 075d: Clear leadership-team needs_section_data items (data now supplied)
-- ============================================================================

UPDATE site_work_items
SET status = 'wont_fix', completed_at = NOW(),
    result = '{"reason": "team data now in site_specs.identity.team"}'::jsonb
WHERE status = 'needs_human_review'
  AND item_type = 'needs_section_data'
  AND spec->>'section_name' = 'leadership-team'
  AND site_id IN (SELECT id FROM sites WHERE domain IN ('ai-agent-orchestration.com', 'finetuning.uk', 'leopardessconsulting.co.uk'));

-- ============================================================================
-- 075e: Also rebuild about pages (for leadership-team sections)
-- ============================================================================

UPDATE pages SET build_status = 'needs_rebuild'
WHERE site_id IN (SELECT id FROM sites WHERE domain IN ('ai-agent-orchestration.com', 'finetuning.uk', 'leopardessconsulting.co.uk'))
  AND name = 'about'
  AND status = 'active'
  AND build_status = 'deployed';

-- ============================================================================
-- Verification
-- ============================================================================

-- Check pricing removed from sections
SELECT s.domain, p.name, p.sections::text, p.build_status
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.name IN ('pricing', 'how-we-work', 'engagement-model')
  AND s.domain IN ('ai-agent-orchestration.com', 'finetuning.uk', 'leopardessconsulting.co.uk');

-- Check team data landed
SELECT s.domain, jsonb_array_length(ss.data->'team') as team_size
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE ss.aspect = 'identity' AND ss.is_current = true
  AND s.domain IN ('ai-agent-orchestration.com', 'finetuning.uk', 'leopardessconsulting.co.uk');

-- Check remaining needs_human_review for these sites
SELECT spec->>'section_name', status, count(*)
FROM site_work_items
WHERE item_type = 'needs_section_data'
  AND site_id IN (SELECT id FROM sites WHERE domain IN ('ai-agent-orchestration.com', 'finetuning.uk', 'leopardessconsulting.co.uk'))
GROUP BY 1, 2;

-- ============================================================================
-- FIX 3: Reset failed improve_tool and audit_tool items
--
-- These failed with "query param path 'input_data.component_id' resolved to nil"
-- because the dispatch loop didn't flatten component_id from spec.
-- Fix 1 corrects the dispatch loop. This resets the items so they re-dispatch.
-- ============================================================================

-- 3a. Clear parent references to these failed items
UPDATE site_work_items SET parent_item_id = NULL
WHERE parent_item_id IN (
    SELECT id FROM site_work_items
    WHERE status = 'failed'
      AND pipeline = 'build'
      AND item_type IN ('improve_tool', 'audit_tool')
      AND error LIKE '%resolved to nil%'
);

-- 3b. Delete failed items where an active duplicate already exists
--     (the dedup index would block the reset)
DELETE FROM site_work_items
WHERE status = 'failed'
  AND pipeline = 'build'
  AND item_type IN ('improve_tool', 'audit_tool')
  AND error LIKE '%resolved to nil%'
  AND item_key IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM site_work_items live
    WHERE live.site_id = site_work_items.site_id
      AND live.item_key = site_work_items.item_key
      AND live.id != site_work_items.id
        AND live.status NOT IN ('complete', 'verified', 'rejected', 'wont_fix', 'failed')
);

-- 3c. Reset remaining failed items to triaged
UPDATE site_work_items
SET status = 'triaged',
    attempt_count = 0,
    error = NULL,
    claimed_by = NULL,
    claimed_at = NULL,
    updated_at = NOW()
WHERE status = 'failed'
  AND pipeline = 'build'
  AND item_type IN ('improve_tool', 'audit_tool')
  AND error LIKE '%resolved to nil%';

---

timeout fix

-- migration_claimed_item_timeout_evidence_check.sql
--
-- Fix the false-positive completion bug in the claimed-item-timeout
-- scheduled task. The previous evidence check for `page_rerender` and
-- `needs_rerender` looked for ANY page on the site updated since the
-- claim — so unrelated rerenders on the same site would falsely mark a
-- stuck item complete. Also used updated_at (bumped on any UPDATE)
-- rather than deployed_at (the actual deploy signal).
--
-- Confirmed false positive: gaswholesalers fuel-industry-insights on
-- 2026-05-12 — claimed at 19:28, auto-completed at 19:43, actual git
-- commit at 20:30 via a separate path. The 19:43 auto-complete had no
-- evidence the targeted page had actually been rendered.
--
-- Changes:
--   page_rerender    : was site-level + updated_at  → page_id-specific + deployed_at
--   needs_content_page : was page_name + updated_at → page_id-specific + deployed_at
--   needs_rerender   : dropped from evidence branch — page_id is NULL on
--                      most rows (it's a site-level orchestrator that fans
--                      out), and this branch was the main false-positive
--                      source. Lets the item fall through to `reset` where
--                      it'll retry; needs_rerender is fast and idempotent
--                      so retry is cheap.
--   needs_design     : unchanged. Site-level by nature. Comment added
--                      to flag that this branch still has narrow false-
--                      positive potential (only checks `head` slot, uses
--                      updated_at).
--
-- Verification before applying:
--   SELECT name, LEFT(pre_query, 100) AS preview
--   FROM scheduled_tasks WHERE name = 'claimed-item-timeout';

BEGIN;

-- Snapshot the current pre_query so we can rollback by hand if needed.
-- (Optional — pg_dump or pg_dumpall is the real backup mechanism.)
SELECT name, pre_query, updated_at
FROM scheduled_tasks
WHERE name = 'claimed-item-timeout';

UPDATE scheduled_tasks
SET pre_query = $PQ$
    WITH completed_by_evidence AS (
    -- Items where the handler's work is provably done on the specific
    -- targeted artifact (not just "something on the same site changed").
    -- See debugging guide section 9: "claimed-item-timeout evidence
    -- check produces false-positive completions" for the prior bug.
    UPDATE site_work_items wi
    SET status = 'complete',
        completed_at = NOW(),
        error = 'Auto-completed: work verified done despite lost response'
    WHERE wi.status = 'claimed'
      AND wi.claimed_at < NOW() - INTERVAL '15 minutes'
      AND (
        -- Content pages: the specific page was deployed after claim.
        (wi.item_type = 'needs_content_page' AND wi.page_id IS NOT NULL AND EXISTS (
            SELECT 1 FROM pages p
            WHERE p.id = wi.page_id
              AND p.build_status = 'deployed'
              AND p.deployed_at IS NOT NULL
              AND p.deployed_at > wi.claimed_at
        ))
        OR
        -- Page rerenders: the specific page was deployed after claim.
        -- Note: NOT 'needs_rerender' — that's a site-level orchestrator
        -- with page_id NULL, can't have per-page evidence. Let it
        -- fall through to reset.
        (wi.item_type = 'page_rerender' AND wi.page_id IS NOT NULL AND EXISTS (
            SELECT 1 FROM pages p
            WHERE p.id = wi.page_id
              AND p.build_status = 'deployed'
              AND p.deployed_at IS NOT NULL
              AND p.deployed_at > wi.claimed_at
        ))
        OR
        -- Design items: site-level by nature. CAVEAT — this branch still
        -- has narrow false-positive potential because (a) it only checks
        -- the head slot, not header/footer, and (b) it uses updated_at
        -- rather than a deploy-specific timestamp (site_components has
        -- no deployed_at column). Acceptable for now; needs_design items
        -- are rare and the impact is bounded to design-only work.
        (wi.item_type = 'needs_design' AND EXISTS (
            SELECT 1 FROM site_components sc
            WHERE sc.site_id = wi.site_id
              AND sc.slot_name = 'head'
              AND sc.updated_at > wi.claimed_at
        ))
      )
    RETURNING id, item_type, handler_agent, status
),
reset AS (
    -- Remaining stuck items: no evidence of completion, reset for retry.
    -- needs_rerender items always land here now (previously could be
    -- false-positive auto-completed); that's intended — retry is cheap.
    UPDATE site_work_items
    SET status = CASE
            WHEN attempt_count + 1 >= max_attempts THEN 'failed'
            ELSE 'triaged'
        END,
        claimed_by = NULL,
        claimed_at = NULL,
        attempt_count = attempt_count + 1,
        error = CASE
            WHEN attempt_count + 1 >= max_attempts THEN 'Claim timed out (attempts exhausted)'
            ELSE 'Claim timed out — handler pod likely died'
        END
    WHERE status = 'claimed'
      AND claimed_at < NOW() - INTERVAL '40 minutes'
      AND id NOT IN (SELECT id FROM completed_by_evidence)
    RETURNING id, item_type, handler_agent, status
)
SELECT
    (SELECT COUNT(*) FROM completed_by_evidence) as auto_completed,
    (SELECT COUNT(*) FROM reset) as reset_count
    $PQ$,
    updated_at = NOW()
    WHERE name = 'claimed-item-timeout';

-- Verify the new pre_query took.
SELECT name, enabled, interval_seconds, updated_at,
    LEFT(pre_query, 200) AS preview
FROM scheduled_tasks
WHERE name = 'claimed-item-timeout';

COMMIT;

-- Rollback (paste the snapshot from before applying — keep that output):
-- UPDATE scheduled_tasks SET pre_query = $$<original pre_query text>$$
-- WHERE name = 'claimed-item-timeout';
-- migration_claimed_item_timeout_evidence_v2.sql
--
-- SUPERSEDES migration_claimed_item_timeout_evidence_check.sql.
-- Apply THIS one. It is the same change as the prior migration (page-specific
-- deploy evidence instead of the old loose site-level updated_at check) PLUS a
-- strengthening of the needs_content_page branch. Idempotent: it sets the
-- claimed-item-timeout task's pre_query in one shot, so it is safe to apply
-- whether or not the prior migration was already applied.
--
-- Why the false-positive completion bug exists: the claimed-item-timeout
-- scheduled task's pre_query marks a stuck `claimed` item `complete` when it
-- believes the work is done. The original check looked for ANY page on the
-- site updated since the claim (site-level + updated_at), so unrelated work on
-- the same site falsely marked a stuck item complete.
--   Confirmed false positive: gaswholesalers fuel-industry-insights, 2026-05-12
--   — claimed 19:28, auto-completed 19:43, actual git commit 20:30 via a
--   separate path. The 19:43 auto-complete had no evidence the targeted page
--   had actually been rendered.
--   Confirmed again: gamesdesign.co.uk homepage (`index`), 2026-06-04 — its
--   `needs_content_page` was auto-completed "verified done despite lost
--   response" with ZERO page_components and no deployed file. Root cause of the
--   missing root index.html.
--
-- Changes vs the prior migration:
--   needs_content_page : was page_id-specific + pages.deployed_at  →  now
--                        page_id-specific + page_components existence + recency.
--                        Rationale: a needs_content_page produces page_components
--                        (content), NOT a deploy — deploy is a separate
--                        page_rerender item in this pipeline (we observed
--                        needs_page:index and page_rerender_index as distinct
--                        items). So the deploy flag was the WRONG (downstream)
--                        signal: it under-credits content that was written but
--                        not yet deployed (forcing a needless re-generation),
--                        and over-credits a page that is `deployed` from prior
--                        content while THIS claim's content was never written.
--                        Worse, build_status='deployed' is a separately-set,
--                        derived flag that can be true with zero components (the
--                        homepage). Checking page_components directly keys the
--                        evidence off the ground-truth artifact and decouples
--                        this branch from the untrustworthy flag. The
--                        `pc.updated_at > wi.claimed_at` guard ensures the
--                        components are from THIS claim's run, not stale rows
--                        from a prior plan/generation.
--   page_rerender      : UNCHANGED — page_id-specific + pages.deployed_at. A
--                        page_rerender's job IS to deploy, so deployed_at is the
--                        correct artifact for it. (Trustworthiness of the
--                        `deployed` flag is hardened separately by the
--                        UpdatePageStatusAction guard — "Option B" — which
--                        refuses to mark a 0-component page deployed.)
--   needs_rerender     : still dropped from the evidence branch — page_id is
--                        NULL on most rows (site-level orchestrator that fans
--                        out), so it cannot have per-page evidence. Falls
--                        through to `reset` and retries; needs_rerender is fast
--                        and idempotent so retry is cheap.
--   needs_design       : UNCHANGED. Site-level by nature. CAVEAT retained: this
--                        branch still has narrow false-positive potential (only
--                        checks the `head` slot, uses updated_at — site_components
--                        has no deployed_at). Acceptable; needs_design is rare.
--   reset branch       : UNCHANGED — this is the stale-claim reset (the Lever-C
--                        / Mode-3 behaviour): >40 min, attempt_count+1,
--                        fail-if-exhausted, clear claim fields. Do not duplicate
--                        it with a separate watchdog.
--
-- Schema confirmed before writing: page_components(page_id, component_id,
-- rendered_html, updated_at) all exist. Error strings and the returned column
-- names (auto_completed, reset_count) are unchanged so nothing that greps logs
-- or consumes the task output breaks.
--
-- Verification before applying:
--   SELECT name, LEFT(pre_query, 120) AS preview
--   FROM scheduled_tasks WHERE name = 'claimed-item-timeout';

BEGIN;

-- Snapshot the current pre_query so we can rollback by hand if needed.
-- (Optional — pg_dump / pg_dumpall is the real backup mechanism.)
SELECT name, pre_query, updated_at
FROM scheduled_tasks
WHERE name = 'claimed-item-timeout';

UPDATE scheduled_tasks
SET pre_query = $PQ$
    WITH completed_by_evidence AS (
    -- Items where the handler's work is provably done on the specific
    -- targeted artifact (not just "something on the same site changed").
    -- See debugging guide section 9: "claimed-item-timeout evidence
    -- check produces false-positive completions" for the prior bug.
    UPDATE site_work_items wi
    SET status = 'complete',
        completed_at = NOW(),
        error = 'Auto-completed: work verified done despite lost response'
    WHERE wi.status = 'claimed'
      AND wi.claimed_at < NOW() - INTERVAL '15 minutes'
      AND (
        -- Content pages: the page's OWN content artifact (page_components)
        -- was written by THIS claim's handler. needs_content_page produces
        -- components, not a deploy (deploy is a separate page_rerender item),
        -- so we check the artifact directly rather than build_status='deployed'
        -- (a downstream, separately-set flag that can be true with zero
        -- components — see the gamesdesign homepage, 2026-06-04). The
        -- updated_at > claimed_at guard ensures the components are from THIS
        -- claim's run, not stale rows from a prior plan/generation.
        (wi.item_type = 'needs_content_page' AND wi.page_id IS NOT NULL AND EXISTS (
            SELECT 1 FROM page_components pc
            WHERE pc.page_id = wi.page_id
              AND pc.component_id IS NOT NULL
              AND pc.rendered_html IS NOT NULL
              AND pc.rendered_html <> ''
              AND pc.updated_at > wi.claimed_at
        ))
        OR
        -- Page rerenders: the specific page was deployed after claim.
        -- Note: NOT 'needs_rerender' — that's a site-level orchestrator
        -- with page_id NULL, can't have per-page evidence. Let it
        -- fall through to reset. A page_rerender's job IS to deploy, so
        -- deployed_at is the correct artifact here (the deployed flag is
        -- hardened separately by the UpdatePageStatusAction guard).
        (wi.item_type = 'page_rerender' AND wi.page_id IS NOT NULL AND EXISTS (
            SELECT 1 FROM pages p
            WHERE p.id = wi.page_id
              AND p.build_status = 'deployed'
              AND p.deployed_at IS NOT NULL
              AND p.deployed_at > wi.claimed_at
        ))
        OR
        -- Design items: site-level by nature. CAVEAT — this branch still
        -- has narrow false-positive potential because (a) it only checks
        -- the head slot, not header/footer, and (b) it uses updated_at
        -- rather than a deploy-specific timestamp (site_components has
        -- no deployed_at column). Acceptable for now; needs_design items
        -- are rare and the impact is bounded to design-only work.
        (wi.item_type = 'needs_design' AND EXISTS (
            SELECT 1 FROM site_components sc
            WHERE sc.site_id = wi.site_id
              AND sc.slot_name = 'head'
              AND sc.updated_at > wi.claimed_at
        ))
      )
    RETURNING id, item_type, handler_agent, status
),
reset AS (
    -- Remaining stuck items: no evidence of completion, reset for retry.
    -- needs_rerender items always land here now (previously could be
    -- false-positive auto-completed); that's intended — retry is cheap.
    UPDATE site_work_items
    SET status = CASE
            WHEN attempt_count + 1 >= max_attempts THEN 'failed'
            ELSE 'triaged'
        END,
        claimed_by = NULL,
        claimed_at = NULL,
        attempt_count = attempt_count + 1,
        error = CASE
            WHEN attempt_count + 1 >= max_attempts THEN 'Claim timed out (attempts exhausted)'
            ELSE 'Claim timed out — handler pod likely died'
        END
    WHERE status = 'claimed'
      AND claimed_at < NOW() - INTERVAL '40 minutes'
      AND id NOT IN (SELECT id FROM completed_by_evidence)
    RETURNING id, item_type, handler_agent, status
)
SELECT
    (SELECT COUNT(*) FROM completed_by_evidence) as auto_completed,
    (SELECT COUNT(*) FROM reset) as reset_count
    $PQ$,
    updated_at = NOW()
    WHERE name = 'claimed-item-timeout';

-- Verify the new pre_query took.
SELECT name, enabled, interval_seconds, updated_at,
    LEFT(pre_query, 200) AS preview
FROM scheduled_tasks
WHERE name = 'claimed-item-timeout';

COMMIT;

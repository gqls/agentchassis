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
    handler_agent TEXT,                      -- which agent type processes this

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
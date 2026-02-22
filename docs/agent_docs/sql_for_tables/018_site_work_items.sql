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
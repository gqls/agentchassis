-- ============================================================================
-- 026: content_sources table
-- Foundation for the news/content feed pipeline.
-- Each row defines one source of content for a site.
-- ============================================================================

-- The table
CREATE TABLE IF NOT EXISTS content_sources (
                                               id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

    -- What kind of source this is
    -- news_search: web search with search_type=news (uses existing web search adapter)
    -- rss:         RSS/Atom feed URL
    -- api_news:    LLM-based news API (xAI/Grok, Perplexity, etc.)
    -- scrape:      targeted site scraping via firecrawl
    -- api_data:    structured data API (BoE rates, ticket prices, etc.)
    source_type     TEXT NOT NULL,

    -- Human-readable name, e.g. "BoxingScene RSS", "Boxing news search"
    name            TEXT NOT NULL,

    -- Which entity type this source provides (NULL for general news/content)
    -- Links to site_entities.entity_type when source feeds entity data
    entity_type     TEXT,

    -- Type-specific configuration (see examples below)
    config          JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Scheduling
    fetch_interval  INTERVAL NOT NULL DEFAULT '6 hours',
    last_fetched_at TIMESTAMPTZ,
    next_fetch_at   TIMESTAMPTZ,

    -- Status tracking
    is_active       BOOLEAN NOT NULL DEFAULT true,
    error_count     INT NOT NULL DEFAULT 0,
    last_error      TEXT,
    last_error_at   TIMESTAMPTZ,

    -- Metadata
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cs_site       ON content_sources (site_id) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_cs_due        ON content_sources (next_fetch_at) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_cs_type       ON content_sources (source_type);
CREATE INDEX IF NOT EXISTS idx_cs_entity     ON content_sources (site_id, entity_type) WHERE entity_type IS NOT NULL;

-- Dedup: one source name per site (needed for seed function ON CONFLICT)
CREATE UNIQUE INDEX IF NOT EXISTS idx_cs_site_name ON content_sources (site_id, name);

-- Add the missing FK from content_feed_items.source_id -> content_sources.id
-- Check if it already exists first (idempotent)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'content_feed_items_source_id_fkey'
    ) THEN
ALTER TABLE content_feed_items
    ADD CONSTRAINT content_feed_items_source_id_fkey
        FOREIGN KEY (source_id) REFERENCES content_sources(id);
END IF;
END $$;

-- ============================================================================
-- Config examples by source_type
-- ============================================================================

-- news_search: uses existing web_search action with search_type=news
-- {
--     "queries": ["boxing news", "boxing fight results"],
--     "num_results": 10,
--     "provider": "firecrawl"        -- optional, uses default if empty
-- }

-- rss: RSS/Atom feed URL
-- {
--     "feed_url": "https://www.boxingscene.com/feed",
--     "max_items": 20                -- per fetch
-- }

-- api_news: LLM-based news service (Grok/xAI)
-- {
--     "provider": "xai",
--     "model": "grok-3",
--     "prompt_template": "What are the latest boxing news stories from the last {{.hours}} hours? Include fight announcements, results, and major developments. Return as JSON array with fields: title, summary, source_url, published_at",
--     "hours_lookback": 12,
--     "max_items": 10
-- }

-- scrape: targeted site scraping
-- {
--     "urls": [
--         "https://www.boxrec.com/en/news",
--         "https://www.ringtv.com/category/news/"
--     ],
--     "scrape_config": {
--         "only_main_content": true,
--         "extract_links": true
--     },
--     "content_selector": "article"  -- CSS selector hint for extraction
-- }

-- api_data: structured data APIs (for future: BoE rates, ticket prices)
-- {
--     "endpoint": "https://api.bankofengland.co.uk/...",
--     "method": "GET",
--     "headers": {},
--     "response_path": "result.data",
--     "data_type": "mortgage_rate",
--     "transform_template": "Current base rate: {{.value}}% as of {{.date}}"
-- }


-- ============================================================================
-- Seed sources for boxing/events vertical
-- (will be parameterised by site_id when wired into the pipeline)
-- ============================================================================

-- Helper function to seed boxing sources for a given site
-- Usage: SELECT seed_boxing_sources('site-uuid-here');
CREATE OR REPLACE FUNCTION seed_boxing_sources(p_site_id UUID)
RETURNS void AS $$
BEGIN
    -- 1. News search sources
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
VALUES
    (p_site_id, 'news_search', 'Boxing news - general',
     '{"queries": ["boxing news today", "boxing fight results", "boxing upcoming fights"], "num_results": 10}'::jsonb,
     '4 hours'::interval),

    (p_site_id, 'news_search', 'Boxing news - major events',
     '{"queries": ["boxing championship fight", "boxing title fight announcement", "boxing PPV results"], "num_results": 5}'::jsonb,
     '2 hours'::interval)
    ON CONFLICT (site_id, name) DO NOTHING;

-- 2. RSS feeds
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
VALUES
    (p_site_id, 'rss', 'BoxingScene RSS',
     '{"feed_url": "https://www.boxingscene.com/feed", "max_items": 15}'::jsonb,
     '2 hours'::interval),

    (p_site_id, 'rss', 'ESPN Boxing RSS',
     '{"feed_url": "https://www.espn.com/espn/rss/boxing/news", "max_items": 10}'::jsonb,
     '3 hours'::interval),

    (p_site_id, 'rss', 'BBC Sport Boxing RSS',
     '{"feed_url": "https://feeds.bbci.co.uk/sport/boxing/rss.xml", "max_items": 10}'::jsonb,
     '3 hours'::interval)
    ON CONFLICT (site_id, name) DO NOTHING;

-- 3. Grok/xAI news
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
VALUES
    (p_site_id, 'api_news', 'Grok boxing news',
     '{"provider": "xai", "model": "grok-3", "prompt_template": "What are the most significant boxing news stories from the last {{.hours}} hours? Include fight announcements, results, rankings changes, and injury updates. For each story provide: title, a 2-3 sentence summary, source attribution, and approximate time. Return as a JSON array with fields: title, summary, source_url (if known), source_name, published_at (ISO format).", "hours_lookback": 12, "max_items": 10}'::jsonb,
     '6 hours'::interval)
    ON CONFLICT (site_id, name) DO NOTHING;

-- 4. Scrape targets
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
VALUES
    (p_site_id, 'scrape', 'BoxRec news',
     '{"urls": ["https://www.boxrec.com/en/news"], "scrape_config": {"only_main_content": true}, "max_items": 10}'::jsonb,
     '6 hours'::interval),

    (p_site_id, 'scrape', 'Ring Magazine news',
     '{"urls": ["https://www.ringtv.com/category/news/"], "scrape_config": {"only_main_content": true}, "max_items": 10}'::jsonb,
     '6 hours'::interval)
    ON CONFLICT (site_id, name) DO NOTHING;

-- Set initial next_fetch_at for all new sources
UPDATE content_sources
SET next_fetch_at = now()
WHERE site_id = p_site_id
  AND next_fetch_at IS NULL;

END;
$$ LANGUAGE plpgsql;

   ---
   --
   -- ============================================================================
-- 026: content_sources table
-- Foundation for the news/content feed pipeline.
-- Each row defines one source of content for a site.
-- ============================================================================

-- The table
CREATE TABLE IF NOT EXISTS content_sources (
                                               id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

    -- What kind of source this is
    -- news_search: web search with search_type=news (uses existing web search adapter)
    -- rss:         RSS/Atom feed URL
    -- api_news:    LLM-based news API (xAI/Grok, Perplexity, etc.)
    -- scrape:      targeted site scraping via firecrawl
    -- api_data:    structured data API (BoE rates, ticket prices, etc.)
    source_type     TEXT NOT NULL,

    -- Human-readable name, e.g. "BoxingScene RSS", "Boxing news search"
    name            TEXT NOT NULL,

    -- Which entity type this source provides (NULL for general news/content)
    -- Links to site_entities.entity_type when source feeds entity data
    entity_type     TEXT,

    -- Type-specific configuration (see examples below)
    config          JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Scheduling
    fetch_interval  INTERVAL NOT NULL DEFAULT '6 hours',
    last_fetched_at TIMESTAMPTZ,
    next_fetch_at   TIMESTAMPTZ,

    -- Status tracking
    is_active       BOOLEAN NOT NULL DEFAULT true,
    error_count     INT NOT NULL DEFAULT 0,
    last_error      TEXT,
    last_error_at   TIMESTAMPTZ,

    -- Metadata
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cs_site       ON content_sources (site_id) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_cs_due        ON content_sources (next_fetch_at) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_cs_type       ON content_sources (source_type);
CREATE INDEX IF NOT EXISTS idx_cs_entity     ON content_sources (site_id, entity_type) WHERE entity_type IS NOT NULL;

-- Dedup: one source name per site (needed for seed function ON CONFLICT)
CREATE UNIQUE INDEX IF NOT EXISTS idx_cs_site_name ON content_sources (site_id, name);

-- Add the missing FK from content_feed_items.source_id -> content_sources.id
-- Check if it already exists first (idempotent)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'content_feed_items_source_id_fkey'
    ) THEN
ALTER TABLE content_feed_items
    ADD CONSTRAINT content_feed_items_source_id_fkey
        FOREIGN KEY (source_id) REFERENCES content_sources(id);
END IF;
END $$;

-- ============================================================================
-- Config examples by source_type
-- ============================================================================

-- news_search: uses existing web_search action with search_type=news
-- {
--     "queries": ["boxing news", "boxing fight results"],
--     "num_results": 10,
--     "provider": "firecrawl"        -- optional, uses default if empty
-- }

-- rss: RSS/Atom feed URL
-- {
--     "feed_url": "https://www.boxingscene.com/feed",
--     "max_items": 20                -- per fetch
-- }

-- api_news: LLM-based news service (Grok/xAI)
-- {
--     "provider": "xai",
--     "model": "grok-3",
--     "prompt_template": "What are the latest boxing news stories from the last {{.hours}} hours? Include fight announcements, results, and major developments. Return as JSON array with fields: title, summary, source_url, published_at",
--     "hours_lookback": 12,
--     "max_items": 10
-- }

-- scrape: targeted site scraping
-- {
--     "url": "https://www.boxrec.com/en/news",
--     "scrape_config": {
--         "only_main_content": true
--     },
--     "max_items": 10
-- }

-- api_data: structured data APIs (for future: BoE rates, ticket prices)
-- {
--     "endpoint": "https://api.bankofengland.co.uk/...",
--     "method": "GET",
--     "headers": {},
--     "response_path": "result.data",
--     "data_type": "mortgage_rate",
--     "transform_template": "Current base rate: {{.value}}% as of {{.date}}"
-- }


-- ============================================================================
-- Seed sources for boxing/events vertical
-- (will be parameterised by site_id when wired into the pipeline)
-- ============================================================================

-- Helper function to seed boxing sources for a given site
-- Usage: SELECT seed_boxing_sources('site-uuid-here');
CREATE OR REPLACE FUNCTION seed_boxing_sources(p_site_id UUID)
RETURNS void AS $$
BEGIN
    -- 1. News search sources
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
VALUES
    (p_site_id, 'news_search', 'Boxing news - general',
     '{"queries": ["boxing news today", "boxing fight results", "boxing upcoming fights"], "num_results": 10}'::jsonb,
     '4 hours'::interval),

    (p_site_id, 'news_search', 'Boxing news - major events',
     '{"queries": ["boxing championship fight", "boxing title fight announcement", "boxing PPV results"], "num_results": 5}'::jsonb,
     '2 hours'::interval)
    ON CONFLICT (site_id, name) DO NOTHING;

-- 2. RSS feeds
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
VALUES
    (p_site_id, 'rss', 'BoxingScene RSS',
     '{"feed_url": "https://www.boxingscene.com/feed", "max_items": 15}'::jsonb,
     '2 hours'::interval),

    (p_site_id, 'rss', 'ESPN Boxing RSS',
     '{"feed_url": "https://www.espn.com/espn/rss/boxing/news", "max_items": 10}'::jsonb,
     '3 hours'::interval),

    (p_site_id, 'rss', 'BBC Sport Boxing RSS',
     '{"feed_url": "https://feeds.bbci.co.uk/sport/boxing/rss.xml", "max_items": 10}'::jsonb,
     '3 hours'::interval)
    ON CONFLICT (site_id, name) DO NOTHING;

-- 3. Grok/xAI news
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
VALUES
    (p_site_id, 'api_news', 'Grok boxing news',
     '{"provider": "xai", "model": "grok-3", "prompt_template": "What are the most significant boxing news stories from the last {{.hours}} hours? Include fight announcements, results, rankings changes, and injury updates. For each story provide: title, a 2-3 sentence summary, source attribution, and approximate time. Return as a JSON array with fields: title, summary, source_url (if known), source_name, published_at (ISO format).", "hours_lookback": 12, "max_items": 10}'::jsonb,
     '6 hours'::interval)
    ON CONFLICT (site_id, name) DO NOTHING;

-- 4. Scrape targets
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
VALUES
    (p_site_id, 'scrape', 'BoxRec news',
     '{"url": "https://www.boxrec.com/en/news", "scrape_config": {"only_main_content": true}, "max_items": 10}'::jsonb,
     '6 hours'::interval),

    (p_site_id, 'scrape', 'Ring Magazine news',
     '{"url": "https://www.ringtv.com/category/news/", "scrape_config": {"only_main_content": true}, "max_items": 10}'::jsonb,
     '6 hours'::interval)
    ON CONFLICT (site_id, name) DO NOTHING;

-- Set initial next_fetch_at for all new sources
UPDATE content_sources
SET next_fetch_at = now()
WHERE site_id = p_site_id
  AND next_fetch_at IS NULL;

END;
$$ LANGUAGE plpgsql;
https://claude.ai/chat/1ac4f28a-bee4-4bca-a9eb-aa3f0ca041a2

     -- ===========================================================================
-- MIGRATION: Component-Based Website Architecture
-- File: 041_component_architecture_schema.sql
-- ===========================================================================
-- This migration adds support for:
--   - Component-driven page building
--   - Research with source tracking
--   - Asset provenance tracking
--   - Product and affiliate content
--   - Build status tracking per page/section
-- ===========================================================================

BEGIN;

-- ===========================================================================
-- PART 1: EXTEND sites TABLE
-- ===========================================================================
-- Note: github_repo, github_branch, style_collection_id already exist

-- Content store for questionnaire/brief answers
ALTER TABLE sites ADD COLUMN IF NOT EXISTS
    content_data JSONB DEFAULT '{}'::jsonb;
COMMENT ON COLUMN sites.content_data IS 
    'Structured content from brief: company_name, services[], about_us, contact info, etc.';

-- Brand asset references (logo, favicon, og_image URLs)
ALTER TABLE sites ADD COLUMN IF NOT EXISTS
    brand_assets JSONB DEFAULT '{}'::jsonb;
COMMENT ON COLUMN sites.brand_assets IS 
    'References to brand assets: {logo: {primary: {asset_id, url}}, favicon: {...}}';

-- Default components for all pages (head, header, footer)
ALTER TABLE sites ADD COLUMN IF NOT EXISTS
    default_components JSONB DEFAULT '{}'::jsonb;
COMMENT ON COLUMN sites.default_components IS 
    'Default component names: {head: "head-seo-standard", header: "header-pro", footer: "footer-4col"}';

-- Deployment configuration (cloudflare, dns, etc.)
ALTER TABLE sites ADD COLUMN IF NOT EXISTS
    deploy_config JSONB DEFAULT '{}'::jsonb;
COMMENT ON COLUMN sites.deploy_config IS 
    'Deployment settings: {provider: "cloudflare", project_name: "...", dns: {...}}';

-- Build status
ALTER TABLE sites ADD COLUMN IF NOT EXISTS
    build_status TEXT DEFAULT 'pending';
COMMENT ON COLUMN sites.build_status IS 
    'Overall site build status: pending, planning, building, deployed, failed';

-- Index for content queries
CREATE INDEX IF NOT EXISTS idx_sites_content ON sites USING gin(content_data);
CREATE INDEX IF NOT EXISTS idx_sites_build_status ON sites(build_status);


-- ===========================================================================
-- PART 2: EXTEND content_components TABLE
-- ===========================================================================

-- Rendering approach
ALTER TABLE content_components ADD COLUMN IF NOT EXISTS
    render_mode TEXT DEFAULT 'template';
COMMENT ON COLUMN content_components.render_mode IS 
    'How to render: template (direct), agent (spawn agent), composite (has children)';

-- Agent type if render_mode = 'agent'
ALTER TABLE content_components ADD COLUMN IF NOT EXISTS
    agent_type TEXT;
COMMENT ON COLUMN content_components.agent_type IS 
    'Agent to spawn if render_mode=agent, e.g. section-writer, research-agent';

-- Agent workflow override
ALTER TABLE content_components ADD COLUMN IF NOT EXISTS
    agent_workflow TEXT;
COMMENT ON COLUMN content_components.agent_workflow IS 
    'Optional specific workflow name for the agent';

-- Data sources - where to pull data from for rendering
ALTER TABLE content_components ADD COLUMN IF NOT EXISTS
    data_sources TEXT[];
COMMENT ON COLUMN content_components.data_sources IS 
    'Dot paths for data: ["site.content_data.services", "brief.company_name"]';

-- Child components for composite render_mode
ALTER TABLE content_components ADD COLUMN IF NOT EXISTS
    child_components JSONB;
COMMENT ON COLUMN content_components.child_components IS 
    'Child component names for composite: ["hero", "services-grid", "cta"]';

-- Component level for hierarchy queries
ALTER TABLE content_components ADD COLUMN IF NOT EXISTS
    component_level TEXT DEFAULT 'section';
COMMENT ON COLUMN content_components.component_level IS 
    'Hierarchy level: site, page, section, element, head, header, footer';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_components_render_mode ON content_components(render_mode);
CREATE INDEX IF NOT EXISTS idx_components_level ON content_components(component_level);
CREATE INDEX IF NOT EXISTS idx_components_agent_type ON content_components(agent_type)
    WHERE agent_type IS NOT NULL;


-- ===========================================================================
-- PART 3: EXTEND pages TABLE
-- ===========================================================================

-- Build status tracking
ALTER TABLE pages ADD COLUMN IF NOT EXISTS
    build_status TEXT DEFAULT 'pending';
COMMENT ON COLUMN pages.build_status IS 
    'Build status: pending, planning, building, reviewing, approved, deployed, failed';

-- Git commit reference when deployed
ALTER TABLE pages ADD COLUMN IF NOT EXISTS
    deploy_commit TEXT;
COMMENT ON COLUMN pages.deploy_commit IS 
    'Git commit SHA when page was deployed';

-- When page was deployed
ALTER TABLE pages ADD COLUMN IF NOT EXISTS
    deployed_at TIMESTAMPTZ;

-- Index
CREATE INDEX IF NOT EXISTS idx_pages_build_status ON pages(build_status);


-- ===========================================================================
-- PART 4: EXTEND page_components TABLE
-- ===========================================================================

-- Build status for individual sections
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    build_status TEXT DEFAULT 'pending';
COMMENT ON COLUMN page_components.build_status IS 
    'Section status: pending, writing, reviewing, approved, deployed';

-- Review tracking
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    reviewed_at TIMESTAMPTZ;

ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    reviewed_by TEXT;
COMMENT ON COLUMN page_components.reviewed_by IS 
    'Who reviewed: hitl, eval-agent, auto-approved';

-- Git commit reference
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    deploy_commit TEXT;

-- Link to research used for this content
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    research_id UUID;
-- FK added after research_results table created

-- Whether sources are displayed on the page
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    sources_displayed BOOLEAN DEFAULT false;

-- Index
CREATE INDEX IF NOT EXISTS idx_page_components_status ON page_components(build_status);


-- ===========================================================================
-- PART 5: CREATE research_results TABLE
-- ===========================================================================

CREATE TABLE IF NOT EXISTS research_results (
                                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- What this research is for (all optional, at least one should be set)
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    component_instance_id UUID REFERENCES page_components(id) ON DELETE SET NULL,

    -- The research query/topic
    query TEXT NOT NULL,
    topic TEXT,  -- More descriptive topic name

-- Research findings
    findings JSONB,  -- Structured findings from synthesis
    summary TEXT,    -- Plain text summary

-- Sources with full attribution
    sources JSONB NOT NULL DEFAULT '[]'::jsonb,
    -- Format: [{
    --   url: "https://...",
    --   title: "Page Title",
    --   domain: "example.com",
    --   accessed_at: "2024-01-15T...",
    --   quotes: ["relevant quote 1", "quote 2"],
    --   relevance_score: 0.85
    -- }]

    -- Tracking
    researched_by TEXT,  -- agent_id that performed research
    research_agent_type TEXT DEFAULT 'research-agent',

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT now(),
    expires_at TIMESTAMPTZ  -- When research should be refreshed
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_research_site ON research_results(site_id);
CREATE INDEX IF NOT EXISTS idx_research_page ON research_results(page_id);
CREATE INDEX IF NOT EXISTS idx_research_component ON research_results(component_instance_id);
CREATE INDEX IF NOT EXISTS idx_research_created ON research_results(created_at DESC);

-- Now add FK from page_components
ALTER TABLE page_components
    ADD CONSTRAINT page_components_research_id_fkey
        FOREIGN KEY (research_id) REFERENCES research_results(id) ON DELETE SET NULL;


-- ===========================================================================
-- PART 6: CREATE assets TABLE
-- ===========================================================================

CREATE TABLE IF NOT EXISTS assets (
                                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Ownership
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,

    -- What is this asset
    asset_type TEXT NOT NULL,  -- 'image', 'video', 'document', 'logo', 'favicon'
    purpose TEXT,  -- 'hero', 'product_image', 'og_image', 'headshot', 'background'
    name TEXT,     -- Human-readable name

-- Storage location
    url TEXT NOT NULL,
    storage_provider TEXT,  -- 'cloudflare-r2', 's3', 'github', 'external'
    storage_path TEXT,      -- Path within provider

-- File metadata
    filename TEXT,
    mime_type TEXT,
    file_size INTEGER,      -- Bytes
    dimensions JSONB,       -- {width: 1200, height: 630} for images/video
    duration INTEGER,       -- Seconds for video/audio

-- Origin tracking (provenance)
    origin_type TEXT NOT NULL DEFAULT 'uploaded',
    -- Values: 'generated', 'uploaded', 'scraped', 'stock', 'affiliate', 'derived'

    origin_url TEXT,        -- Original source URL if scraped/stock
    origin_prompt TEXT,     -- AI generation prompt
    origin_model TEXT,      -- AI model used (dall-e-3, midjourney, etc.)
    origin_asset_id UUID REFERENCES assets(id),  -- Parent if derived

-- Alterations history
    alterations JSONB DEFAULT '[]'::jsonb,
    -- Format: [{type: 'resize', params: {width: 800}, at: '2024-...', by: 'agent-id'}]

    -- Attribution and licensing
    attribution TEXT,       -- Required attribution text
    license TEXT,           -- License type: 'owned', 'cc-by', 'stock-license', etc.
    license_url TEXT,

    -- Status
    status TEXT DEFAULT 'active',  -- 'active', 'archived', 'deleted'

-- Timestamps
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_assets_site ON assets(site_id);
CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(asset_type);
CREATE INDEX IF NOT EXISTS idx_assets_purpose ON assets(purpose);
CREATE INDEX IF NOT EXISTS idx_assets_origin ON assets(origin_type);
CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_assets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_assets_updated_at ON assets;
CREATE TRIGGER trg_assets_updated_at
    BEFORE UPDATE ON assets
    FOR EACH ROW
    EXECUTE FUNCTION update_assets_updated_at();


-- ===========================================================================
-- PART 7: CREATE products TABLE
-- ===========================================================================

CREATE TABLE IF NOT EXISTS products (
                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Ownership
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,

    -- Identity
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    sku TEXT,

    -- Descriptions
    short_description TEXT,   -- For listings/cards
    description TEXT,         -- Full description
    features JSONB,           -- [{name: "...", value: "..."}]
    specifications JSONB,     -- Technical specs

-- Pricing
    price DECIMAL(10,2),
    compare_at_price DECIMAL(10,2),  -- Original/strikethrough price
    currency TEXT DEFAULT 'GBP',
    price_display TEXT,       -- Custom display: "From £99" or "Contact us"

-- Categorization
    category TEXT,
    subcategory TEXT,
    tags TEXT[],

    -- SEO
    meta_title TEXT,
    meta_description TEXT,

    -- Additional content
    content_data JSONB DEFAULT '{}'::jsonb,

    -- Status
    status TEXT DEFAULT 'draft',  -- 'draft', 'active', 'archived'
    published_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    -- Constraints
    UNIQUE(site_id, slug)
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_products_site ON products(site_id);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_tags ON products USING gin(tags);

-- Trigger
DROP TRIGGER IF EXISTS trg_products_updated_at ON products;
CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_assets_updated_at();


-- ===========================================================================
-- PART 8: CREATE product_assets TABLE (junction)
-- ===========================================================================

CREATE TABLE IF NOT EXISTS product_assets (
                                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,

    -- Ordering and classification
    position INTEGER DEFAULT 0,
    is_primary BOOLEAN DEFAULT false,
    asset_role TEXT,  -- 'gallery', 'thumbnail', 'zoom', 'video', 'document'

-- Timestamps
    created_at TIMESTAMPTZ DEFAULT now(),

    -- Constraints
    UNIQUE(product_id, asset_id)
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_product_assets_product ON product_assets(product_id);
CREATE INDEX IF NOT EXISTS idx_product_assets_asset ON product_assets(asset_id);


-- ===========================================================================
-- PART 9: CREATE affiliate_programs TABLE
-- ===========================================================================

CREATE TABLE IF NOT EXISTS affiliate_programs (
                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Program identity
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    network TEXT,  -- 'amazon', 'awin', 'cj', 'shareasale', 'direct'

-- Tracking configuration
    affiliate_id TEXT,        -- Our affiliate ID with this network
    tracking_params JSONB,    -- How to build affiliate URLs
-- Example: {param_name: "tag", format: "{{affiliate_id}}-21"}

-- Terms
    commission_type TEXT,     -- 'percentage', 'fixed', 'tiered'
    commission_rate TEXT,     -- "5%" or "£2 per sale"
    cookie_duration TEXT,     -- "24 hours", "30 days"

-- API access (if available)
    api_endpoint TEXT,
    api_credentials_ref TEXT,  -- Reference to secrets store

-- Status
    status TEXT DEFAULT 'active',  -- 'active', 'paused', 'terminated'

-- Timestamps
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
    );

-- Index
CREATE INDEX IF NOT EXISTS idx_affiliate_programs_network ON affiliate_programs(network);
CREATE INDEX IF NOT EXISTS idx_affiliate_programs_status ON affiliate_programs(status);


-- ===========================================================================
-- PART 10: CREATE affiliate_products TABLE
-- ===========================================================================

CREATE TABLE IF NOT EXISTS affiliate_products (
                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Relationships
    program_id UUID NOT NULL REFERENCES affiliate_programs(id) ON DELETE CASCADE,
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,  -- Which site uses this

-- External reference
    external_id TEXT NOT NULL,    -- ASIN, product ID, SKU
    external_url TEXT NOT NULL,   -- Direct link to product
    affiliate_url TEXT,           -- Link with our tracking

-- Cached content from affiliate network (may be stale)
    cached_name TEXT,
    cached_description TEXT,
    cached_price TEXT,
    cached_currency TEXT,
    cached_image_url TEXT,
    cached_availability TEXT,     -- 'in_stock', 'out_of_stock', 'unknown'
    cached_at TIMESTAMPTZ,

    -- Our custom content (overrides cached)
    custom_name TEXT,
    custom_description TEXT,
    custom_short_description TEXT,
    custom_pros JSONB,            -- ["Pro 1", "Pro 2"]
    custom_cons JSONB,            -- ["Con 1"]
    custom_verdict TEXT,
    custom_rating DECIMAL(2,1),   -- Our rating: 4.5

-- Custom image (our own, not cached)
    custom_image_id UUID REFERENCES assets(id),

    -- Categorization for our site
    category TEXT,
    tags TEXT[],

    -- Content status
    content_status TEXT DEFAULT 'cached',
    -- 'cached' (only affiliate data), 'enhanced' (we've added custom), 'reviewed' (fully reviewed)

    -- Status
    status TEXT DEFAULT 'active',  -- 'active', 'unavailable', 'archived'
    last_checked_at TIMESTAMPTZ,   -- When we last verified availability

-- Timestamps
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    -- Constraints
    UNIQUE(program_id, external_id)
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_affiliate_products_program ON affiliate_products(program_id);
CREATE INDEX IF NOT EXISTS idx_affiliate_products_site ON affiliate_products(site_id);
CREATE INDEX IF NOT EXISTS idx_affiliate_products_external ON affiliate_products(external_id);
CREATE INDEX IF NOT EXISTS idx_affiliate_products_status ON affiliate_products(status);
CREATE INDEX IF NOT EXISTS idx_affiliate_products_content ON affiliate_products(content_status);


-- ===========================================================================
-- PART 11: EXTEND link_registry TABLE
-- ===========================================================================

-- Link to affiliate product if this is an affiliate link
ALTER TABLE link_registry ADD COLUMN IF NOT EXISTS
    affiliate_product_id UUID REFERENCES affiliate_products(id) ON DELETE SET NULL;

-- Whether this link requires disclosure (FTC/ASA compliance)
ALTER TABLE link_registry ADD COLUMN IF NOT EXISTS
    requires_disclosure BOOLEAN DEFAULT false;

-- Index
CREATE INDEX IF NOT EXISTS idx_link_registry_affiliate
    ON link_registry(affiliate_product_id)
    WHERE affiliate_product_id IS NOT NULL;


-- ===========================================================================
-- PART 12: USEFUL VIEWS
-- ===========================================================================

-- View: Pages with build status summary
CREATE OR REPLACE VIEW v_page_build_status AS
SELECT
    p.id AS page_id,
    p.site_id,
    p.name AS page_name,
    p.title AS page_title,
    p.build_status AS page_status,
    p.deployed_at,
    s.domain,
    COUNT(pc.id) AS total_sections,
    COUNT(CASE WHEN pc.build_status = 'deployed' THEN 1 END) AS deployed_sections,
    COUNT(CASE WHEN pc.build_status = 'approved' THEN 1 END) AS approved_sections,
    COUNT(CASE WHEN pc.build_status = 'reviewing' THEN 1 END) AS reviewing_sections,
    COUNT(CASE WHEN pc.build_status = 'pending' THEN 1 END) AS pending_sections
FROM pages p
         JOIN sites s ON p.site_id = s.id
         LEFT JOIN page_components pc ON pc.page_id = p.id
GROUP BY p.id, p.site_id, p.name, p.title, p.build_status, p.deployed_at, s.domain;

-- View: Research with source count
CREATE OR REPLACE VIEW v_research_summary AS
SELECT
    r.id,
    r.site_id,
    r.page_id,
    r.query,
    r.topic,
    r.summary,
    jsonb_array_length(r.sources) AS source_count,
    r.researched_by,
    r.created_at,
    s.domain,
    p.name AS page_name
FROM research_results r
         LEFT JOIN sites s ON r.site_id = s.id
         LEFT JOIN pages p ON r.page_id = p.id;

-- View: Assets by site with counts
CREATE OR REPLACE VIEW v_site_assets AS
SELECT
    s.id AS site_id,
    s.domain,
    a.asset_type,
    a.origin_type,
    COUNT(*) AS asset_count
FROM sites s
         LEFT JOIN assets a ON a.site_id = s.id
WHERE a.status = 'active'
GROUP BY s.id, s.domain, a.asset_type, a.origin_type;


-- ===========================================================================
-- PART 13: COMMENTS FOR DOCUMENTATION
-- ===========================================================================

COMMENT ON TABLE research_results IS 
    'Stores research findings with full source attribution for LLM-generated content';

COMMENT ON TABLE assets IS 
    'All images, videos, documents with full provenance tracking';

COMMENT ON TABLE products IS 
    'Product catalog for e-commerce or product-focused sites';

COMMENT ON TABLE product_assets IS 
    'Links products to their images and documents';

COMMENT ON TABLE affiliate_programs IS 
    'Configuration for affiliate networks (Amazon, Awin, etc.)';

COMMENT ON TABLE affiliate_products IS 
    'Cached affiliate product data with our custom content overlays';


COMMIT;

-- ===========================================================================
-- VERIFICATION QUERIES (run after migration)
-- ===========================================================================

-- Check sites columns
-- SELECT column_name, data_type FROM information_schema.columns 
-- WHERE table_name = 'sites' ORDER BY ordinal_position;

-- Check new tables exist
-- SELECT table_name FROM information_schema.tables 
-- WHERE table_schema = 'public' 
-- AND table_name IN ('research_results', 'assets', 'products', 'product_assets', 'affiliate_programs', 'affiliate_products');

-- Check content_components columns
-- SELECT column_name, data_type FROM information_schema.columns 
-- WHERE table_name = 'content_components' AND column_name IN ('render_mode', 'agent_type', 'data_sources');
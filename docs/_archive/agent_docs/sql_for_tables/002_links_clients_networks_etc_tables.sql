-- ============================================================================
-- LINK MANAGEMENT INTEGRATION - MVP Migration
-- ============================================================================
-- This builds on existing tables (content_components, relationships)
-- Adds minimal tables needed for multipage-website-builder to work
-- Designed for patch-only updates from the start
-- ============================================================================

-- ============================================================================
-- 1. CLIENT & NETWORK HIERARCHY
-- ============================================================================

-- Clients (links to auth-service)
CREATE TABLE IF NOT EXISTS clients (
                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(255) UNIQUE, -- ID from auth-service
    name VARCHAR(255) NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

-- Networks belong to clients
CREATE TABLE IF NOT EXISTS networks (
                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID REFERENCES clients(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    settings JSONB DEFAULT '{}', -- network-wide config (affiliates, etc.)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(client_id, slug)
    );

CREATE INDEX IF NOT EXISTS idx_networks_client ON networks(client_id);

-- Sites belong to networks
CREATE TABLE IF NOT EXISTS sites (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    network_id UUID REFERENCES networks(id) ON DELETE CASCADE,
    domain VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    brand_dna JSONB DEFAULT '{}', -- visual identity, voice parameters, invariants
    github_repo VARCHAR(500),
    github_branch VARCHAR(100) DEFAULT 'main',
    settings JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active', -- active, paused, archived
    last_built_at TIMESTAMP WITH TIME ZONE,
    last_deployed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_sites_network ON sites(network_id);
CREATE INDEX IF NOT EXISTS idx_sites_domain ON sites(domain);
CREATE INDEX IF NOT EXISTS idx_sites_status ON sites(status);

-- ============================================================================
-- 2. FLOWS (Multi-Track User Journeys)
-- ============================================================================

CREATE TABLE IF NOT EXISTS site_flows (
                                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    audience_segment VARCHAR(255),
    narrative_arc JSONB, -- {"stages": [{"name": "awareness", "voice": {...}}]}
    entry_points TEXT[],
    success_metric TEXT,
    voice_parameters JSONB DEFAULT '{}', -- base voice for this flow
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(site_id, slug)
    );

CREATE INDEX IF NOT EXISTS idx_flows_site ON site_flows(site_id);

-- ============================================================================
-- 3. PAGES
-- ============================================================================

CREATE TABLE IF NOT EXISTS pages (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL, -- slug: "about", "services/consulting"
    url VARCHAR(500) NOT NULL, -- "/about.html", "/services/consulting.html"
    title VARCHAR(500),
    page_type VARCHAR(50), -- index, content, product, legal, landing
    status VARCHAR(50) DEFAULT 'active', -- active, draft, archived, redirected
    content_hash VARCHAR(64), -- MD5 of full page HTML for change detection
    meta_description TEXT,
    topics TEXT[], -- extracted topics for semantic queries
    nav_label VARCHAR(255), -- display name in navigation (if different from title)
    nav_order INTEGER DEFAULT 100, -- sort order in navigation
    in_header BOOLEAN DEFAULT true,
    in_footer BOOLEAN DEFAULT true,
    last_built_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE, -- for campaign pages
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(site_id, name)
    );

CREATE INDEX IF NOT EXISTS idx_pages_site ON pages(site_id);
CREATE INDEX IF NOT EXISTS idx_pages_status ON pages(status);
CREATE INDEX IF NOT EXISTS idx_pages_type ON pages(page_type);
CREATE INDEX IF NOT EXISTS idx_pages_nav ON pages(site_id, in_header, nav_order)
    WHERE status = 'active';

-- Pages can belong to multiple flows
CREATE TABLE IF NOT EXISTS flow_pages (
                                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flow_id UUID REFERENCES site_flows(id) ON DELETE CASCADE,
    page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    stage_in_narrative VARCHAR(100), -- "awareness", "consideration", "conversion"
    sequence_order INTEGER,
    context_overrides JSONB DEFAULT '{}', -- voice_formality, urgency overrides

    UNIQUE(flow_id, page_id)
    );

CREATE INDEX IF NOT EXISTS idx_flow_pages_flow ON flow_pages(flow_id);
CREATE INDEX IF NOT EXISTS idx_flow_pages_page ON flow_pages(page_id);

-- ============================================================================
-- 4. PAGE COMPONENTS (Instances of content_components on pages)
-- ============================================================================

CREATE TABLE IF NOT EXISTS page_components (
                                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    component_id UUID REFERENCES content_components(id), -- template used
    position INTEGER NOT NULL, -- order on page
    slot_name VARCHAR(100), -- if nested: which slot in parent
    parent_instance_id UUID REFERENCES page_components(id), -- for nesting

-- Rendered content (source of truth for this instance)
    rendered_html TEXT,
    content_data JSONB, -- data passed to template
    content_hash VARCHAR(64), -- MD5 of rendered_html for change detection

-- Semantic addressing for future editing
    data_path VARCHAR(500), -- "page.section[2].grid.slot[left]"
    data_uuid UUID DEFAULT gen_random_uuid(), -- unique ID for element targeting

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_page_components_page ON page_components(page_id);
CREATE INDEX IF NOT EXISTS idx_page_components_template ON page_components(component_id);
CREATE INDEX IF NOT EXISTS idx_page_components_parent ON page_components(parent_instance_id);

-- ============================================================================
-- 5. LINK REGISTRY (Index of all links, extracted from components)
-- ============================================================================

CREATE TABLE IF NOT EXISTS link_registry (
                                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Source (where the link lives)
    source_component_instance_id UUID REFERENCES page_components(id) ON DELETE CASCADE,
    source_page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
    source_site_id UUID REFERENCES sites(id) ON DELETE CASCADE,

    -- Target
    target_url VARCHAR(1000) NOT NULL,
    target_page_id UUID REFERENCES pages(id), -- if internal, resolved
    target_site_id UUID REFERENCES sites(id), -- if cross-site internal

-- Classification
    scope VARCHAR(50) NOT NULL, -- internal, page, site, network, external
    link_type VARCHAR(50) NOT NULL, -- navigation, content, semantic, affiliate, reference

-- Metadata
    anchor_text VARCHAR(500),
    rel_attr VARCHAR(100), -- nofollow, sponsored, ugc

-- For affiliates (future)
    affiliate_provider VARCHAR(100),
    affiliate_tag VARCHAR(255),

    -- Status & health
    status VARCHAR(50) DEFAULT 'active',
    last_validated_at TIMESTAMP WITH TIME ZONE,
                                                                         validation_result VARCHAR(50), -- ok, broken, timeout, redirect

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_links_source_component ON link_registry(source_component_instance_id);
CREATE INDEX IF NOT EXISTS idx_links_source_page ON link_registry(source_page_id);
CREATE INDEX IF NOT EXISTS idx_links_source_site ON link_registry(source_site_id);
CREATE INDEX IF NOT EXISTS idx_links_target_page ON link_registry(target_page_id);
CREATE INDEX IF NOT EXISTS idx_links_target_site ON link_registry(target_site_id);
CREATE INDEX IF NOT EXISTS idx_links_type ON link_registry(link_type);
CREATE INDEX IF NOT EXISTS idx_links_scope ON link_registry(scope);
CREATE INDEX IF NOT EXISTS idx_links_broken ON link_registry(validation_result)
    WHERE validation_result IS NOT NULL AND validation_result != 'ok';

-- ============================================================================
-- 6. NAVIGATION STRUCTURES (Pre-computed, cached)
-- ============================================================================

CREATE TABLE IF NOT EXISTS navigation_structures (
                                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    nav_type VARCHAR(50) NOT NULL, -- header, footer, mobile, sidebar

-- The structure
    structure JSONB NOT NULL,
    /*
    {
      "items": [
        {"page_id": "uuid", "label": "Home", "url": "/index.html", "children": []},
        {"page_id": "uuid", "label": "Services", "url": "/services.html", "children": [
          {"page_id": "uuid", "label": "Consulting", "url": "/services/consulting.html"}
        ]}
      ]
    }
    */

    -- Versioning
    version INTEGER DEFAULT 1,
    is_current BOOLEAN DEFAULT true,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(site_id, nav_type, version)
    );

CREATE INDEX IF NOT EXISTS idx_nav_site ON navigation_structures(site_id);
CREATE INDEX IF NOT EXISTS idx_nav_current ON navigation_structures(site_id, nav_type)
    WHERE is_current = true;

-- ============================================================================
-- 7. REDIRECTS
-- ============================================================================

CREATE TABLE IF NOT EXISTS redirects (
                                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,
    from_url VARCHAR(500) NOT NULL,
    to_url VARCHAR(500) NOT NULL,
    redirect_type INTEGER DEFAULT 301, -- 301, 302, 307, 410
    reason VARCHAR(255),
    source_page_id UUID REFERENCES pages(id), -- original page if known
    hit_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

                                          UNIQUE(site_id, from_url)
    );

CREATE INDEX IF NOT EXISTS idx_redirects_site ON redirects(site_id);
CREATE INDEX IF NOT EXISTS idx_redirects_from ON redirects(from_url);

-- ============================================================================
-- 8. HELPER FUNCTIONS
-- ============================================================================

-- Build navigation structure from pages
CREATE OR REPLACE FUNCTION build_navigation_for_site(p_site_id UUID, p_nav_type VARCHAR(50))
RETURNS JSONB AS $$
DECLARE
v_structure JSONB;
BEGIN
SELECT jsonb_build_object(
               'items', COALESCE(jsonb_agg(
                                         jsonb_build_object(
                                                 'page_id', id,
                                                 'label', COALESCE(nav_label, title, name),
                                                 'url', url,
                                                 'children', '[]'::jsonb
                                         ) ORDER BY nav_order, name
                                 ), '[]'::jsonb)
       )
INTO v_structure
FROM pages
WHERE site_id = p_site_id
  AND status = 'active'
  AND CASE
          WHEN p_nav_type = 'header' THEN in_header
          WHEN p_nav_type = 'footer' THEN in_footer
          ELSE true
    END;

RETURN v_structure;
END;
$$ LANGUAGE plpgsql;

-- Get or create current navigation
CREATE OR REPLACE FUNCTION get_current_navigation(p_site_id UUID, p_nav_type VARCHAR(50))
RETURNS JSONB AS $$
DECLARE
v_structure JSONB;
BEGIN
    -- Try to get cached
SELECT structure INTO v_structure
FROM navigation_structures
WHERE site_id = p_site_id
  AND nav_type = p_nav_type
  AND is_current = true;

-- If not found, build and cache
IF v_structure IS NULL THEN
        v_structure := build_navigation_for_site(p_site_id, p_nav_type);

INSERT INTO navigation_structures (site_id, nav_type, structure, is_current)
VALUES (p_site_id, p_nav_type, v_structure, true)
    ON CONFLICT (site_id, nav_type, version)
        DO UPDATE SET structure = EXCLUDED.structure, is_current = true;
END IF;

RETURN v_structure;
END;
$$ LANGUAGE plpgsql;

-- Invalidate navigation cache when pages change
CREATE OR REPLACE FUNCTION invalidate_navigation_cache()
RETURNS TRIGGER AS $$
BEGIN
UPDATE navigation_structures
SET is_current = false
WHERE site_id = COALESCE(NEW.site_id, OLD.site_id);

RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger to invalidate nav cache on page changes
DROP TRIGGER IF EXISTS trg_invalidate_nav_on_page_change ON pages;
CREATE TRIGGER trg_invalidate_nav_on_page_change
    AFTER INSERT OR UPDATE OR DELETE ON pages
    FOR EACH ROW
    EXECUTE FUNCTION invalidate_navigation_cache();

-- ============================================================================
-- 9. SEMANTIC RELATIONSHIPS (using existing relationships table)
-- ============================================================================
-- The relationships table already exists and is perfect for:
-- - pillar_to_cluster
-- - cluster_to_pillar
-- - related_content
-- - cross_site_reference
-- - next_in_series
-- - See integrated_link_architecture.md for usage examples

-- Add index for page relationships if not exists
CREATE INDEX IF NOT EXISTS idx_relationships_pages
    ON relationships(source_entity_id, target_entity_id)
    WHERE source_entity_type = 'page' AND target_entity_type = 'page';

-- ============================================================================
-- 10. VIEWS FOR COMMON QUERIES
-- ============================================================================

-- All links for a site with resolution
CREATE OR REPLACE VIEW v_site_links AS
SELECT
    lr.*,
    sp.name as source_page_name,
    sp.url as source_page_url,
    tp.name as target_page_name,
    tp.url as target_page_url,
    cc.function as component_function
FROM link_registry lr
         LEFT JOIN pages sp ON lr.source_page_id = sp.id
         LEFT JOIN pages tp ON lr.target_page_id = tp.id
         LEFT JOIN page_components pc ON lr.source_component_instance_id = pc.id
         LEFT JOIN content_components cc ON pc.component_id = cc.id;

-- Navigation-ready page list
CREATE OR REPLACE VIEW v_navigation_pages AS
SELECT
    p.id,
    p.site_id,
    COALESCE(p.nav_label, p.title, p.name) as label,
    p.url,
    p.nav_order,
    p.in_header,
    p.in_footer,
    p.page_type
FROM pages p
WHERE p.status = 'active'
ORDER BY p.site_id, p.nav_order, p.name;

-- ============================================================================
-- 11. DEFAULT DATA
-- ============================================================================

-- Create a default client for testing (can be removed in production)
INSERT INTO clients (id, external_id, name)
VALUES ('00000000-0000-0000-0000-000000000001', 'default', 'Default Client')
    ON CONFLICT (external_id) DO NOTHING;

-- Create a default network
INSERT INTO networks (id, client_id, name, slug)
VALUES (
           '00000000-0000-0000-0000-000000000002',
           '00000000-0000-0000-0000-000000000001',
           'Default Network',
           'default'
       )
    ON CONFLICT (client_id, slug) DO NOTHING;

-- ============================================================================
-- VERIFICATION
-- ============================================================================

SELECT 'Tables created:' as status;
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN (
                     'clients', 'networks', 'sites', 'site_flows', 'pages',
                     'flow_pages', 'page_components', 'link_registry',
                     'navigation_structures', 'redirects'
    )
ORDER BY table_name;

SELECT 'Functions created:' as status;
SELECT routine_name FROM information_schema.routines
WHERE routine_schema = 'public'
  AND routine_name IN (
                       'build_navigation_for_site',
                       'get_current_navigation',
                       'invalidate_navigation_cache'
    );
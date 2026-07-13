-- ============================================================
-- 2. Site areas (major sections of a site with different styling)
-- ============================================================
CREATE TABLE IF NOT EXISTS site_areas (
                                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,               -- 'ai-solutions', 'resources'
    display_name VARCHAR(255),                -- 'AI Solutions'
    url_prefix VARCHAR(100),                  -- '/ai', '/resources'
    description TEXT,

    -- Area-level defaults
    nav_style VARCHAR(50) DEFAULT 'inherit',  -- 'tabs', 'sidebar', 'hamburger', 'inherit'
    theme_overrides JSONB DEFAULT '{}',       -- accent colors, etc.

    sort_order INTEGER DEFAULT 0,
    is_default BOOLEAN DEFAULT false,         -- main/default area
    status VARCHAR(50) DEFAULT 'active',

    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    UNIQUE(site_id, name)
    );

CREATE INDEX idx_site_areas_site ON site_areas(site_id);
CREATE INDEX idx_site_areas_prefix ON site_areas(url_prefix);

COMMENT ON TABLE site_areas IS 'Major sections of a site that may have different navigation/styling';
COMMENT ON COLUMN site_areas.url_prefix IS 'URL path prefix for pages in this area';
COMMENT ON COLUMN site_areas.is_default IS 'True for the main/default area of the site';

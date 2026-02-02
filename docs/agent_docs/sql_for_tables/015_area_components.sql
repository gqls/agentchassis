-- ============================================================
-- 3. Area-level components (override site defaults for an area)
-- ============================================================
CREATE TABLE IF NOT EXISTS area_components (
                                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    area_id UUID NOT NULL REFERENCES site_areas(id) ON DELETE CASCADE,
    slot_name VARCHAR(100) NOT NULL,          -- 'header', 'footer', 'nav'
    component_id UUID REFERENCES content_components(id),
    rendered_html TEXT,
    content_data JSONB DEFAULT '{}',
    build_status TEXT DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(area_id, slot_name)
    );

CREATE INDEX idx_area_components_area ON area_components(area_id);

COMMENT ON TABLE area_components IS 'Area-specific components that override site-level defaults';

-- ============================================================================
-- CONTENT ITEMS - Reusable text content separate from layout
-- ============================================================================
--
-- This table stores the "what to say" separately from "how to show it"
--
-- Patterns followed from existing tables:
--   - site_id scoping like assets (NULL = library/shared)
--   - origin tracking like assets (generated, written, imported)
--   - content_data JSONB like page_components
--   - relationships table for complex entity linking
--
-- Usage:
--   - page_components.content_item_id references this instead of inline content_data
--   - Same headline can be used in hero AND footer (reference, not duplicate)
--   - Library content (site_id=NULL) available to multiple clients
--   - Industry-specific content collections
-- ============================================================================

CREATE TABLE IF NOT EXISTS content_items (
                                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Ownership scope (NULL = library/shared content)
    site_id UUID REFERENCES sites(id) ON DELETE CASCADE,

    -- Classification
    content_type VARCHAR(100) NOT NULL,     -- 'headline', 'tagline', 'service_description', 'bio', 'testimonial', 'cta', 'paragraph'
    content_key VARCHAR(255),               -- semantic key: 'hero.headline', 'services.consulting.description'

-- The actual content
    content_data JSONB NOT NULL DEFAULT '{}',  -- flexible structure per content_type
-- Examples:
--   headline: {"text": "Transform Your Business", "subtext": "With AI-powered solutions"}
--   service_description: {"title": "Consulting", "summary": "...", "details": "...", "benefits": [...]}
--   testimonial: {"quote": "...", "author": "...", "role": "...", "company": "..."}
--   bio: {"name": "...", "title": "...", "summary": "...", "full_bio": "..."}

-- For text search and display
    plain_text TEXT,                        -- searchable text extracted from content_data

-- Library/sharing
    is_library BOOLEAN DEFAULT false,       -- available for reuse across sites
    library_tags TEXT[],                    -- ['consulting', 'saas', 'professional-services']
    industry_vertical VARCHAR(100),         -- 'technology', 'healthcare', 'finance'

-- Origin tracking (like assets)
    origin_type TEXT NOT NULL DEFAULT 'generated',  -- 'generated', 'written', 'imported', 'edited'
    origin_agent TEXT,                      -- 'content-writer', 'site-planner'
    origin_research_id UUID REFERENCES research_results(id) ON DELETE SET NULL,
    origin_content_id UUID REFERENCES content_items(id),  -- if derived from another

-- Workflow
    status VARCHAR(50) DEFAULT 'draft',     -- 'draft', 'pending_review', 'approved', 'published', 'archived'
    version INTEGER DEFAULT 1,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    approved_at TIMESTAMP WITH TIME ZONE,
                                          approved_by VARCHAR(100)
    );

-- Indexes
CREATE INDEX idx_content_items_site ON content_items(site_id);
CREATE INDEX idx_content_items_type ON content_items(content_type);
CREATE INDEX idx_content_items_key ON content_items(site_id, content_key);
CREATE INDEX idx_content_items_library ON content_items(is_library, industry_vertical) WHERE is_library = true;
CREATE INDEX idx_content_items_status ON content_items(status);
CREATE INDEX idx_content_items_search ON content_items USING gin(to_tsvector('english', plain_text));

-- Unique constraint: one content_key per site (but allows multiple library items with same key)
CREATE UNIQUE INDEX idx_content_items_unique_key ON content_items(site_id, content_key)
    WHERE site_id IS NOT NULL AND content_key IS NOT NULL;

-- ============================================================================
-- Update page_components to reference content_items
-- ============================================================================

-- Add foreign key to content_items (content_data becomes optional/override)
ALTER TABLE page_components
    ADD COLUMN IF NOT EXISTS content_item_id UUID REFERENCES content_items(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_page_components_content ON page_components(content_item_id);

-- Add comment explaining the relationship
COMMENT ON COLUMN page_components.content_item_id IS 'Reference to reusable content. If set, content_data acts as override/customization';
COMMENT ON COLUMN page_components.content_data IS 'Inline content OR overrides for content_item. Merged with content_item.content_data if both exist';

-- ============================================================================
-- Helper function: Get effective content for a page_component
-- ============================================================================

CREATE OR REPLACE FUNCTION get_component_content(p_component_id UUID)
RETURNS JSONB AS $$
DECLARE
v_content_item_id UUID;
    v_content_data JSONB;
    v_item_data JSONB;
BEGIN
    -- Get component's content references
SELECT content_item_id, content_data
INTO v_content_item_id, v_content_data
FROM page_components
WHERE id = p_component_id;

-- If no content_item, return inline content_data
IF v_content_item_id IS NULL THEN
        RETURN COALESCE(v_content_data, '{}'::jsonb);
END IF;

    -- Get content item data
SELECT content_data INTO v_item_data
FROM content_items
WHERE id = v_content_item_id;

-- Merge: content_data overrides content_item (shallow merge)
RETURN COALESCE(v_item_data, '{}'::jsonb) || COALESCE(v_content_data, '{}'::jsonb);
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- View: Content items with usage stats
-- ============================================================================

CREATE OR REPLACE VIEW v_content_usage AS
SELECT
    ci.*,
    COUNT(DISTINCT pc.id) as component_usage_count,
    COUNT(DISTINCT pc.page_id) as page_usage_count,
    COUNT(DISTINCT p.site_id) as site_usage_count,
    array_agg(DISTINCT p.site_id) FILTER (WHERE p.site_id IS NOT NULL) as used_by_sites
FROM content_items ci
         LEFT JOIN page_components pc ON pc.content_item_id = ci.id
         LEFT JOIN pages p ON pc.page_id = p.id
GROUP BY ci.id;

-- ============================================================================
-- Example content types and their structures
-- ============================================================================

COMMENT ON TABLE content_items IS '
Content Types and Structures:

headline:
  {"text": "...", "subtext": "...", "style": "bold|subtle"}

tagline:
  {"text": "...", "context": "hero|footer|navigation"}

service_description:
  {"title": "...", "summary": "...", "details": "...", "benefits": ["...", "..."], "icon": "..."}

testimonial:
  {"quote": "...", "author": "...", "role": "...", "company": "...", "image_asset_id": "uuid"}

bio:
  {"name": "...", "title": "...", "summary": "...", "full_bio": "...", "image_asset_id": "uuid", "social": {...}}

cta:
  {"text": "...", "action_text": "...", "urgency": "low|medium|high"}

paragraph:
  {"heading": "...", "body": "...", "style": "default|highlight|callout"}

faq:
  {"question": "...", "answer": "...", "category": "..."}

feature:
  {"title": "...", "description": "...", "icon": "...", "link": "..."}
';
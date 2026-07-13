-- Migration: Create site_nav_groups and site_nav_items tables
-- These replace the scattered pages-table queries for navigation
-- and the navigation_structures cache table.
--
-- Existing sites without nav table entries are handled by fallback
-- logic in Go that queries the pages table directly.

-- Nav groups: categorised collections of navigation items per site
CREATE TABLE IF NOT EXISTS site_nav_groups (
                                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    group_key TEXT NOT NULL,              -- "primary", "legal", "utility", "content"
    group_label TEXT NOT NULL DEFAULT '', -- "Main Navigation", "Legal", etc.
    group_type TEXT NOT NULL,             -- semantic type for consumer queries
    parent_group_id UUID REFERENCES site_nav_groups(id),
    position INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(site_id, group_key)
    );

-- Nav items: individual navigation entries within groups
CREATE TABLE IF NOT EXISTS site_nav_items (
                                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES site_nav_groups(id) ON DELETE CASCADE,
    parent_item_id UUID REFERENCES site_nav_items(id),
    label TEXT NOT NULL,
    url TEXT NOT NULL,
    page_id UUID REFERENCES pages(id) ON DELETE SET NULL,
    item_type TEXT NOT NULL DEFAULT 'page_link',  -- "page_link", "external_link", "anchor", "section_header"
    position INTEGER DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_nav_groups_site ON site_nav_groups(site_id);
CREATE INDEX IF NOT EXISTS idx_nav_items_site ON site_nav_items(site_id);
CREATE INDEX IF NOT EXISTS idx_nav_items_group ON site_nav_items(group_id);
CREATE INDEX IF NOT EXISTS idx_nav_items_page ON site_nav_items(page_id);
CREATE INDEX IF NOT EXISTS idx_nav_items_site_status ON site_nav_items(site_id, status);
-- Migration: Add site_components, site_areas, area_components tables
-- This enables clean separation of site-level vs page-level components
-- and supports site areas (sections) with their own styling

-- ============================================================
-- 1. Site-level components (header, footer, head - shared across pages)
-- ============================================================
CREATE TABLE IF NOT EXISTS site_components (
                                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    slot_name VARCHAR(100) NOT NULL,          -- 'header', 'footer', 'head'
    component_id UUID REFERENCES content_components(id),
    rendered_html TEXT,
    content_data JSONB DEFAULT '{}',          -- data used to render (for re-rendering)
    build_status TEXT DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(site_id, slot_name)
    );

CREATE INDEX idx_site_components_site ON site_components(site_id);
CREATE INDEX idx_site_components_status ON site_components(build_status);

COMMENT ON TABLE site_components IS 'Site-wide components shared across all pages (header, footer, head)';
COMMENT ON COLUMN site_components.slot_name IS 'Component slot: header, footer, head';
COMMENT ON COLUMN site_components.content_data IS 'Render context data for re-rendering';


====

-- populate site components
-- Migration: Populate site_components from existing data
-- Run this after 41_site_components_migration.sql

-- ============================================================
-- For each site, we need to render and store header/footer
-- This uses the content_components templates + site data
-- ============================================================

-- First, let's see what we're working with
-- SELECT id, domain, name, company_name, email, phone FROM sites;

-- ============================================================
-- Insert header components for each site
-- Uses header-professional-dark as default (or site's configured header)
-- ============================================================

INSERT INTO site_components (site_id, slot_name, component_id, content_data, build_status)
SELECT
    s.id as site_id,
    'header' as slot_name,
    cc.id as component_id,
    jsonb_build_object(
            'company_name', COALESCE(s.company_name, s.name, s.domain),
            'logo_text', COALESCE(s.logo_text, s.company_name, s.name),
            'domain', s.domain,
            'tagline', s.tagline
    ) as content_data,
    'pending' as build_status  -- Will be rendered by build process
FROM sites s
         CROSS JOIN LATERAL (
    SELECT id FROM content_components
    WHERE function = 'header'
    ORDER BY
        CASE WHEN name = COALESCE(s.content_data->>'header', 'header-professional-dark') THEN 0 ELSE 1 END,
        name
        LIMIT 1
) cc
ON CONFLICT (site_id, slot_name) DO UPDATE SET
    component_id = EXCLUDED.component_id,
                                        content_data = EXCLUDED.content_data,
                                        updated_at = now();

-- ============================================================
-- Insert footer components for each site
-- ============================================================

INSERT INTO site_components (site_id, slot_name, component_id, content_data, build_status)
SELECT
    s.id as site_id,
    'footer' as slot_name,
    cc.id as component_id,
    jsonb_build_object(
            'company_name', COALESCE(s.company_name, s.name, s.domain),
            'domain', s.domain,
            'email', s.email,
            'phone', s.phone,
            'tagline', s.tagline,
            'year', EXTRACT(YEAR FROM now())::text
    ) as content_data,
    'pending' as build_status
FROM sites s
         CROSS JOIN LATERAL (
    SELECT id FROM content_components
    WHERE function = 'footer'
    ORDER BY
        CASE WHEN name = COALESCE(s.content_data->>'footer', 'footer-standard') THEN 0 ELSE 1 END,
        name
        LIMIT 1
) cc
ON CONFLICT (site_id, slot_name) DO UPDATE SET
    component_id = EXCLUDED.component_id,
                                        content_data = EXCLUDED.content_data,
                                        updated_at = now();

-- ============================================================
-- Insert head template for each site
-- ============================================================

INSERT INTO site_components (site_id, slot_name, component_id, content_data, build_status)
SELECT
    s.id as site_id,
    'head' as slot_name,
    cc.id as component_id,
    jsonb_build_object(
            'company_name', COALESCE(s.company_name, s.name, s.domain),
            'domain', s.domain,
            'tagline', s.tagline
    ) as content_data,
    'pending' as build_status
FROM sites s
         CROSS JOIN LATERAL (
    SELECT id FROM content_components
    WHERE function = 'head' OR name LIKE 'head%'
    ORDER BY name
        LIMIT 1
) cc
WHERE cc.id IS NOT NULL
ON CONFLICT (site_id, slot_name) DO UPDATE SET
    component_id = EXCLUDED.component_id,
                                        content_data = EXCLUDED.content_data,
                                        updated_at = now();

-- ============================================================
-- Verify the migration
-- ============================================================

SELECT
    s.domain,
    sc.slot_name,
    cc.name as component_name,
    sc.build_status,
    CASE WHEN sc.rendered_html IS NOT NULL THEN 'yes' ELSE 'no' END as has_html
FROM sites s
         LEFT JOIN site_components sc ON sc.site_id = s.id
         LEFT JOIN content_components cc ON sc.component_id = cc.id
ORDER BY s.domain, sc.slot_name;


---

-- Migration: Populate site_components.rendered_html
-- This does basic template substitution using SQL REPLACE
-- Handles both {{var}} and {{.var}} syntax, standardizes on {{.var}}
-- For production, use the render_site_components Go action

-- ============================================================
-- Helper function to replace both {{var}} and {{.var}} syntax
-- ============================================================
CREATE OR REPLACE FUNCTION replace_template_var(
    template TEXT,
    var_name TEXT,
    var_value TEXT
) RETURNS TEXT AS $$
BEGIN
    -- Replace both syntaxes
RETURN REPLACE(
        REPLACE(template, '{{.' || var_name || '}}', COALESCE(var_value, '')),
        '{{' || var_name || '}}', COALESCE(var_value, '')
       );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ============================================================
-- First, let's see what we're working with
-- ============================================================

-- Check templates and their syntax
SELECT name, function,
       LENGTH(html_template) as template_length,
       CASE
           WHEN html_template LIKE '%{{.%' THEN 'Go {{.var}}'
           WHEN html_template LIKE '%{{%' THEN 'Simple {{var}}'
           ELSE 'No variables'
           END as syntax_style
FROM content_components
WHERE function IN ('header', 'footer', 'head')
ORDER BY function, name;

-- Check site data
SELECT id, domain, company_name, tagline, email, phone, logo_text
FROM sites;

-- ============================================================
-- Render header components
-- Variables: company_name, logo_text, domain, tagline, nav_items_html
-- ============================================================

WITH site_nav AS (
    SELECT
        p.site_id,
        string_agg(
                format('<li><a href="%s">%s</a></li>',
                       COALESCE(p.url, '/' || p.name || '.html'),
                       COALESCE(p.nav_label, p.title, initcap(replace(p.name, '-', ' ')))
                ),
                E'\n                    '
            ORDER BY COALESCE(p.nav_order, 99),
                CASE p.name
                    WHEN 'index' THEN 1
                    WHEN 'services' THEN 2
                    WHEN 'about' THEN 3
                    WHEN 'contact' THEN 10
                    ELSE 5
                    END
        ) as nav_html
    FROM pages p
    WHERE p.status IN ('deployed', 'active')
      AND p.deleted_at IS NULL
      AND p.name NOT IN ('privacy', 'terms', 'cookies', '404', 'sitemap')
    GROUP BY p.site_id
)
UPDATE site_components sc
SET
    rendered_html = replace_template_var(
            replace_template_var(
                    replace_template_var(
                            replace_template_var(
                                    replace_template_var(
                                            cc.html_template,
                                            'company_name', COALESCE(s.company_name, s.name, s.domain)
                                    ),
                                    'logo_text', COALESCE(s.logo_text, s.company_name, s.name)
                            ),
                            'domain', s.domain
                    ),
                    'tagline', s.tagline
            ),
            'nav_items_html', COALESCE(sn.nav_html, '')
                    ),
    build_status = 'rendered',
    updated_at = now()
    FROM sites s
JOIN content_components cc ON sc.component_id = cc.id
    LEFT JOIN site_nav sn ON sn.site_id = s.id
WHERE sc.site_id = s.id
  AND sc.slot_name = 'header';

-- ============================================================
-- Render footer components
-- Variables: company_name, email, phone, tagline, year, footer_nav_html
-- ============================================================

-- Migration: Populate site_components.rendered_html
-- Uses simpler SQL without complex JOINs

-- ============================================================
-- Helper function to replace both {{var}} and {{.var}} syntax
-- ============================================================
CREATE OR REPLACE FUNCTION replace_template_var(
    template TEXT,
    var_name TEXT,
    var_value TEXT
) RETURNS TEXT AS $$
BEGIN
RETURN REPLACE(
        REPLACE(template, '{{.' || var_name || '}}', COALESCE(var_value, '')),
        '{{' || var_name || '}}', COALESCE(var_value, '')
       );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ============================================================
-- Helper function to get nav HTML for a site
-- ============================================================
CREATE OR REPLACE FUNCTION get_site_nav_html(p_site_id UUID)
RETURNS TEXT AS $$
SELECT COALESCE(
               string_agg(
                       format('<li><a href="%s">%s</a></li>',
                              COALESCE(url, '/' || name || '.html'),
                              COALESCE(nav_label, title, initcap(replace(name, '-', ' ')))
                       ),
                       E'\n                    '
            ORDER BY COALESCE(nav_order, 99),
                       CASE name
                           WHEN 'index' THEN 1
                           WHEN 'services' THEN 2
                           WHEN 'about' THEN 3
                           WHEN 'contact' THEN 10
                           ELSE 5
                           END
               ),
               ''
       )
FROM pages
WHERE site_id = p_site_id
  AND status IN ('deployed', 'active')
  AND name NOT IN ('privacy', 'terms', 'cookies', '404', 'sitemap');
$$ LANGUAGE SQL STABLE;

-- ============================================================
-- Helper function to get footer nav HTML for a site
-- ============================================================
CREATE OR REPLACE FUNCTION get_footer_nav_html(p_site_id UUID)
RETURNS TEXT AS $$
SELECT COALESCE(
               string_agg(
                       format('<li><a href="%s">%s</a></li>',
                              COALESCE(url, '/' || name || '.html'),
                              COALESCE(nav_label, title, initcap(replace(name, '-', ' ')))
                       ),
                       E'\n                '
            ORDER BY COALESCE(nav_order, 99)
               ),
               ''
       )
FROM pages
WHERE site_id = p_site_id
  AND status IN ('deployed', 'active')
  AND name NOT IN ('index', '404', 'sitemap');
$$ LANGUAGE SQL STABLE;

-- ============================================================
-- Check current state
-- ============================================================
SELECT
    s.domain,
    sc.slot_name,
    cc.name as component_name,
    sc.build_status,
    LENGTH(sc.rendered_html) as html_length
FROM sites s
         LEFT JOIN site_components sc ON sc.site_id = s.id
         LEFT JOIN content_components cc ON sc.component_id = cc.id
ORDER BY s.domain, sc.slot_name;

-- ============================================================
-- Render HEADER components
-- ============================================================
UPDATE site_components sc
SET
    rendered_html = replace_template_var(
            replace_template_var(
                    replace_template_var(
                            replace_template_var(
                                    replace_template_var(
                                            cc.html_template,
                                            'company_name', COALESCE(s.company_name, s.name, s.domain)
                                    ),
                                    'logo_text', COALESCE(s.logo_text, s.company_name, s.name)
                            ),
                            'domain', s.domain
                    ),
                    'tagline', COALESCE(s.tagline, '')
            ),
            'nav_items_html', get_site_nav_html(s.id)
                    ),
    build_status = 'rendered',
    updated_at = now()
    FROM sites s, content_components cc
WHERE sc.site_id = s.id
  AND sc.component_id = cc.id
  AND sc.slot_name = 'header';

-- ============================================================
-- Render FOOTER components
-- ============================================================
UPDATE site_components sc
SET
    rendered_html = replace_template_var(
            replace_template_var(
                    replace_template_var(
                            replace_template_var(
                                    replace_template_var(
                                            replace_template_var(
                                                    cc.html_template,
                                                    'company_name', COALESCE(s.company_name, s.name, s.domain)
                                            ),
                                            'email', COALESCE(s.email, '')
                                    ),
                                    'phone', COALESCE(s.phone, '')
                            ),
                            'tagline', COALESCE(s.tagline, '')
                    ),
                    'year', EXTRACT(YEAR FROM now())::text
            ),
            'footer_nav_html', get_footer_nav_html(s.id)
                    ),
    build_status = 'rendered',
    updated_at = now()
    FROM sites s, content_components cc
WHERE sc.site_id = s.id
  AND sc.component_id = cc.id
  AND sc.slot_name = 'footer';

-- ============================================================
-- Render HEAD components
-- ============================================================
UPDATE site_components sc
SET
    rendered_html = replace_template_var(
            replace_template_var(
                    replace_template_var(
                            cc.html_template,
                            'domain', s.domain
                    ),
                    'company_name', COALESCE(s.company_name, s.name, s.domain)
            ),
            'tagline', COALESCE(s.tagline, '')
                    ),
    build_status = 'rendered',
    updated_at = now()
    FROM sites s, content_components cc
WHERE sc.site_id = s.id
  AND sc.component_id = cc.id
  AND sc.slot_name = 'head';

-- ============================================================
-- Additional variable replacements for all slots
-- ============================================================
UPDATE site_components sc
SET
    rendered_html = replace_template_var(
            replace_template_var(
                    replace_template_var(
                            replace_template_var(
                                    sc.rendered_html,
                                    'contact_email', COALESCE(s.email, '')
                            ),
                            'logo_url', COALESCE(s.logo_url, '')
                    ),
                    'cta_text', 'Get Started'
            ),
            'cta_url', '/contact.html'
                    )
    FROM sites s
WHERE sc.site_id = s.id
  AND sc.rendered_html IS NOT NULL;

-- ============================================================
-- Verify results
-- ============================================================
SELECT
    s.domain,
    sc.slot_name,
    sc.build_status,
    LENGTH(sc.rendered_html) as html_length,
    LEFT(sc.rendered_html, 150) as preview
FROM sites s
    JOIN site_components sc ON sc.site_id = s.id
ORDER BY s.domain, sc.slot_name;

-- ============================================================
-- Check for unreplaced template variables
-- ============================================================
SELECT
    s.domain,
    sc.slot_name,
    (regexp_matches(sc.rendered_html, '\{\{\.?[a-z_]+\}\}', 'g'))[1] as unreplaced_var
FROM sites s
    JOIN site_components sc ON sc.site_id = s.id
WHERE sc.rendered_html ~ '\{\{\.?[a-z_]+\}\}'
ORDER BY s.domain, sc.slot_name;

-- Add more variable replacements
UPDATE site_components sc
SET rendered_html = replace_template_var(
        replace_template_var(
                replace_template_var(
                        sc.rendered_html,
                        'brand_name', COALESCE(s.company_name, s.name, s.domain)
                ),
                'copyright', '© ' || EXTRACT(YEAR FROM now())::text || ' ' || COALESCE(s.company_name, s.name)
        ),
        'theme_css', ''
                    )
    FROM sites s
WHERE sc.site_id = s.id;


------

-- Phase 7a: Add lock columns to site_components for site-wide component governance.
ALTER TABLE site_components ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE site_components ADD COLUMN IF NOT EXISTS locked_by TEXT;
CREATE INDEX IF NOT EXISTS idx_site_components_locked
    ON site_components (locked_at) WHERE locked_at IS NOT NULL;

-- Phase 7: Site-level lock — prevents all automated agent activity.
ALTER TABLE sites ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE sites ADD COLUMN IF NOT EXISTS locked_by TEXT;


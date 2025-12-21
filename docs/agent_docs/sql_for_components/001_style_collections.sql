-- ===========================================================================
-- STYLE COLLECTIONS - Site-Specific Component Bundles
-- ===========================================================================
--
-- Concept: A "style collection" groups together the header, footer, theme,
-- and design tokens that define a site's visual identity. Sites are linked
-- to a style collection, ensuring consistency.
--
-- Benefits:
-- - Tested, working components (no LLM randomness)
-- - Easy to update all sites using a collection
-- - Can offer multiple styles: "professional-dark", "minimal-light", etc.
-- - Site-specific overrides (colors, logo) while keeping structure
--
-- ===========================================================================

-- Style collections that bundle components together
CREATE TABLE IF NOT EXISTS style_collections (
                                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,

    -- Component references
    header_component_id UUID REFERENCES content_components(id),
    header_home_component_id UUID REFERENCES content_components(id),  -- Optional variant for home page
    footer_component_id UUID REFERENCES content_components(id),
    css_theme_id UUID REFERENCES css_themes(id),

    -- Design tokens (can override theme)
    color_palette JSONB DEFAULT '{
        "primary": "#1a1a2e",
        "secondary": "#2d2d44",
        "accent": "#16a085",
        "text": "#333333",
        "text_light": "#666666",
        "background": "#ffffff",
        "background_alt": "#f8f9fa"
    }'::jsonb,

    typography JSONB DEFAULT '{
        "font_family": "-apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif",
        "heading_font": "inherit",
        "base_size": "16px",
        "line_height": "1.6"
    }'::jsonb,

    -- Metadata
    category TEXT,  -- 'professional', 'creative', 'minimal', 'corporate'
    industry_tags TEXT[],  -- ['consulting', 'tech', 'finance']
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
    );

-- Link sites to their style collection
ALTER TABLE sites
    ADD COLUMN IF NOT EXISTS style_collection_id UUID REFERENCES style_collections(id);

-- Site-specific style overrides (colors, logo, etc.)
ALTER TABLE sites
    ADD COLUMN IF NOT EXISTS style_overrides JSONB DEFAULT '{}'::jsonb;

-- ===========================================================================
-- HEADER COMPONENTS
-- ===========================================================================
-- Insert header component templates into content_components

-- Professional Dark Header (default)
INSERT INTO content_components (
    name,
    display_name,
    category,
    function,
    description,
    html_template,
    input_schema,
    semantic_tags
) VALUES (
             'header-professional-dark',
             'Professional Dark Header',
             'header',
             'site-header',
             'Clean, professional header with dark background. Sticky, responsive, mobile menu.',
             '<!-- Professional Dark Header -->
         <header class="site-header">
             <div class="header-container">
                 <a href="/index.html" class="logo">
                     <span class="logo-text">{{logo_text}}</span>
                     {{#if logo_accent}}<span class="logo-accent">{{logo_accent}}</span>{{/if}}
                 </a>
                 <button class="mobile-menu-toggle" aria-label="Toggle menu">
                     <span></span><span></span><span></span>
                 </button>
                 <nav class="main-nav" id="main-nav">
                     <ul>
                         {{#each nav_items}}
                         <li><a href="{{this.url}}"{{#if this.is_active}} class="active"{{/if}}>{{this.label}}</a></li>
                         {{/each}}
                     </ul>
                 </nav>
             </div>
         </header>
         <style>
         .site-header {
             background: {{primary_color}};
             padding: 1rem 0;
             position: sticky;
             top: 0;
             z-index: 1000;
             box-shadow: 0 2px 10px rgba(0,0,0,0.1);
         }
         .header-container {
             max-width: 1200px;
             margin: 0 auto;
             padding: 0 2rem;
             display: flex;
             align-items: center;
             justify-content: space-between;
         }
         .logo {
             text-decoration: none;
             font-size: 1.5rem;
             font-weight: 700;
             color: white;
         }
         .logo-accent { color: {{accent_color}}; }
         .main-nav ul {
             display: flex;
             list-style: none;
             margin: 0;
             padding: 0;
             gap: 2rem;
         }
         .main-nav a {
             color: rgba(255,255,255,0.9);
             text-decoration: none;
             font-weight: 500;
             padding: 0.5rem 0;
             transition: color 0.2s;
         }
         .main-nav a:hover, .main-nav a.active { color: {{accent_color}}; }
         .mobile-menu-toggle {
             display: none;
             background: none;
             border: none;
             cursor: pointer;
             padding: 0.5rem;
         }
         .mobile-menu-toggle span {
             display: block;
             width: 24px;
             height: 2px;
             background: white;
             margin: 5px 0;
             transition: 0.3s;
         }
         @media (max-width: 768px) {
             .mobile-menu-toggle { display: block; }
             .main-nav {
                 position: absolute;
                 top: 100%;
                 left: 0;
                 right: 0;
                 background: {{primary_color}};
                 padding: 1rem;
                 display: none;
             }
             .main-nav.active { display: block; }
             .main-nav ul { flex-direction: column; gap: 0; }
             .main-nav a {
                 display: block;
                 padding: 0.75rem 0;
                 border-bottom: 1px solid rgba(255,255,255,0.1);
             }
         }
         </style>
         <script>
         document.addEventListener("DOMContentLoaded", function() {
             var toggle = document.querySelector(".mobile-menu-toggle");
             var nav = document.querySelector(".main-nav");
             if (toggle && nav) {
                 toggle.addEventListener("click", function() {
                     nav.classList.toggle("active");
                 });
             }
         });
         </script>',
             '{
                 "type": "object",
                 "required": ["logo_text", "nav_items"],
                 "properties": {
                     "logo_text": {
                         "type": "string",
                         "description": "Main logo text"
                     },
                     "logo_accent": {
                         "type": "string",
                         "description": "Optional accent text after logo"
                     },
                     "primary_color": {
                         "type": "string",
                         "default": "#1a1a2e",
                         "description": "Header background color"
                     },
                     "accent_color": {
                         "type": "string",
                         "default": "#16a085",
                         "description": "Accent color for hover/active states"
                     },
                     "nav_items": {
                         "type": "array",
                         "description": "Navigation items",
                         "items": {
                             "type": "object",
                             "properties": {
                                 "label": {"type": "string"},
                                 "url": {"type": "string"},
                                 "is_active": {"type": "boolean", "default": false}
                             }
                         }
                     }
                 }
             }'::jsonb,
             '["header", "navigation", "professional", "dark", "sticky"]'::jsonb
         ) ON CONFLICT (name) DO UPDATE SET
    html_template = EXCLUDED.html_template,
    input_schema = EXCLUDED.input_schema,
    updated_at = now();

-- Minimal Light Header
INSERT INTO content_components (
    name,
    display_name,
    category,
    function,
    description,
    html_template,
    input_schema,
    semantic_tags
) VALUES (
             'header-minimal-light',
             'Minimal Light Header',
             'header',
             'site-header',
             'Clean, minimal header with light background. Centered logo option.',
             '<!-- Minimal Light Header -->
         <header class="site-header site-header--light">
             <div class="header-container">
                 <a href="/index.html" class="logo">{{logo_text}}</a>
                 <nav class="main-nav">
                     <ul>
                         {{#each nav_items}}
                         <li><a href="{{this.url}}"{{#if this.is_active}} class="active"{{/if}}>{{this.label}}</a></li>
                         {{/each}}
                     </ul>
                 </nav>
                 <button class="mobile-menu-toggle" aria-label="Toggle menu">☰</button>
             </div>
         </header>
         <style>
         .site-header--light {
             background: {{background_color}};
             padding: 1.25rem 0;
             border-bottom: 1px solid #eee;
         }
         .site-header--light .header-container {
             max-width: 1200px;
             margin: 0 auto;
             padding: 0 2rem;
             display: flex;
             align-items: center;
             justify-content: space-between;
         }
         .site-header--light .logo {
             text-decoration: none;
             font-size: 1.25rem;
             font-weight: 600;
             color: {{primary_color}};
             letter-spacing: -0.02em;
         }
         .site-header--light .main-nav ul {
             display: flex;
             list-style: none;
             margin: 0;
             padding: 0;
             gap: 2.5rem;
         }
         .site-header--light .main-nav a {
             color: {{text_color}};
             text-decoration: none;
             font-size: 0.95rem;
             transition: color 0.2s;
         }
         .site-header--light .main-nav a:hover,
         .site-header--light .main-nav a.active {
             color: {{accent_color}};
         }
         .site-header--light .mobile-menu-toggle {
             display: none;
             background: none;
             border: none;
             font-size: 1.5rem;
             cursor: pointer;
             color: {{primary_color}};
         }
         @media (max-width: 768px) {
             .site-header--light .mobile-menu-toggle { display: block; }
             .site-header--light .main-nav { display: none; }
             .site-header--light .main-nav.active {
                 display: block;
                 position: absolute;
                 top: 100%;
                 left: 0;
                 right: 0;
                 background: white;
                 padding: 1rem;
                 box-shadow: 0 4px 12px rgba(0,0,0,0.1);
             }
             .site-header--light .main-nav ul { flex-direction: column; gap: 0; }
             .site-header--light .main-nav a { display: block; padding: 0.75rem 0; }
         }
         </style>',
             '{
                 "type": "object",
                 "required": ["logo_text", "nav_items"],
                 "properties": {
                     "logo_text": {"type": "string"},
                     "primary_color": {"type": "string", "default": "#1a1a2e"},
                     "accent_color": {"type": "string", "default": "#2563eb"},
                     "text_color": {"type": "string", "default": "#4a5568"},
                     "background_color": {"type": "string", "default": "#ffffff"},
                     "nav_items": {
                         "type": "array",
                         "items": {
                             "type": "object",
                             "properties": {
                                 "label": {"type": "string"},
                                 "url": {"type": "string"},
                                 "is_active": {"type": "boolean"}
                             }
                         }
                     }
                 }
             }'::jsonb,
             '["header", "navigation", "minimal", "light", "clean"]'::jsonb
         ) ON CONFLICT (name) DO UPDATE SET
    html_template = EXCLUDED.html_template,
    input_schema = EXCLUDED.input_schema,
    updated_at = now();

-- ===========================================================================
-- DEFAULT STYLE COLLECTION
-- ===========================================================================

INSERT INTO style_collections (
    name,
    display_name,
    description,
    category,
    industry_tags,
    color_palette
) VALUES (
             'professional-dark',
             'Professional Dark',
             'Professional style with dark header, clean typography. Suitable for consulting, tech, finance.',
             'professional',
             ARRAY['consulting', 'tech', 'finance', 'b2b'],
             '{
                 "primary": "#1a1a2e",
                 "secondary": "#2d2d44",
                 "accent": "#16a085",
                 "text": "#333333",
                 "text_light": "#666666",
                 "background": "#ffffff",
                 "background_alt": "#f8f9fa"
             }'::jsonb
         ) ON CONFLICT (name) DO NOTHING;

-- Link the header component to the collection
UPDATE style_collections
SET header_component_id = (SELECT id FROM content_components WHERE name = 'header-professional-dark')
WHERE name = 'professional-dark';

-- ===========================================================================
-- USAGE EXAMPLE - How to get header for a site
-- ===========================================================================
/*
-- 1. Get the site's style collection
SELECT
    s.domain,
    sc.name as collection_name,
    cc.html_template as header_template,
    cc.input_schema as header_schema,
    COALESCE(s.style_overrides->'color_palette', sc.color_palette) as colors
FROM sites s
JOIN style_collections sc ON s.style_collection_id = sc.id
JOIN content_components cc ON sc.header_component_id = cc.id
WHERE s.domain = 'leopardessconsulting.co.uk';

-- 2. Or get a specific header component
SELECT
    html_template,
    input_schema
FROM content_components
WHERE name = 'header-professional-dark';
*/

-- ===========================================================================
-- INDEXES
-- ===========================================================================

CREATE INDEX IF NOT EXISTS idx_style_collections_category ON style_collections(category);
CREATE INDEX IF NOT EXISTS idx_style_collections_industry ON style_collections USING GIN(industry_tags);
CREATE INDEX IF NOT EXISTS idx_content_components_category ON content_components(category);
CREATE INDEX IF NOT EXISTS idx_content_components_function ON content_components(function);
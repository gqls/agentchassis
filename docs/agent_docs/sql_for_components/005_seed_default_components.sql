https://claude.ai/chat/1ac4f28a-bee4-4bca-a9eb-aa3f0ca041a2

-- ===========================================================================
-- SEED: Default Head, Header, Footer Components
-- File: 042_seed_default_components.sql
-- ===========================================================================
-- Creates standard components for page assembly
-- ===========================================================================

BEGIN;

-- ===========================================================================
-- HEAD COMPONENTS
-- ===========================================================================

INSERT INTO content_components (
    name,
    display_name,
    function,
    category,
    component_level,
    render_mode,
    description,
    html_template,
    input_schema,
    data_sources
) VALUES (
             'head-seo-standard',
             'Standard SEO Head',
             'head',
             'structure',
             'head',
             'template',
             'Standard HTML head with SEO, Open Graph, and structured data support',
             '<!DOCTYPE html>
         <html lang="{{lang}}">
         <head>
             <meta charset="UTF-8">
             <meta name="viewport" content="width=device-width, initial-scale=1.0">
             <title>{{title}}</title>
             <meta name="description" content="{{description}}">
             {{#if keywords}}<meta name="keywords" content="{{keywords}}">{{/if}}

             <!-- Open Graph -->
             <meta property="og:title" content="{{og_title}}">
             <meta property="og:description" content="{{og_description}}">
             <meta property="og:type" content="{{og_type}}">
             <meta property="og:url" content="{{canonical_url}}">
             {{#if og_image}}<meta property="og:image" content="{{og_image}}">{{/if}}
             <meta property="og:site_name" content="{{site_name}}">

             <!-- Twitter Card -->
             <meta name="twitter:card" content="summary_large_image">
             <meta name="twitter:title" content="{{og_title}}">
             <meta name="twitter:description" content="{{og_description}}">
             {{#if og_image}}<meta name="twitter:image" content="{{og_image}}">{{/if}}

             <!-- Canonical -->
             <link rel="canonical" href="{{canonical_url}}">

             <!-- Favicon -->
             <link rel="icon" type="image/x-icon" href="{{favicon_url}}">
             <link rel="apple-touch-icon" href="{{apple_touch_icon_url}}">

             <!-- Theme Color -->
             <meta name="theme-color" content="{{primary_color}}">

             <!-- Fonts -->
             <link rel="preconnect" href="https://fonts.googleapis.com">
             <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
             {{#if font_url}}<link href="{{font_url}}" rel="stylesheet">{{/if}}

             <!-- Theme CSS -->
             <style>
                 :root {
                     --primary: {{primary_color}};
                     --secondary: {{secondary_color}};
                     --accent: {{accent_color}};
                     --text-dark: {{text_color}};
                     --text-light: #666666;
                     --background: {{background_color}};
                     --border: #e0e0e0;
                 }
                 {{theme_css}}
             </style>

             {{#if structured_data}}
             <!-- JSON-LD Structured Data -->
             <script type="application/ld+json">
             {{structured_data}}
             </script>
             {{/if}}

             {{#if analytics_id}}
             <!-- Analytics -->
             <script async src="https://www.googletagmanager.com/gtag/js?id={{analytics_id}}"></script>
             <script>
                 window.dataLayer = window.dataLayer || [];
                 function gtag(){dataLayer.push(arguments);}
                 gtag(''js'', new Date());
                 gtag(''config'', ''{{analytics_id}}'');
             </script>
             {{/if}}
         </head>
         <body>',
             '{
                 "required": ["title", "description", "canonical_url"],
                 "optional": ["keywords", "og_image", "analytics_id", "structured_data", "font_url"],
                 "defaults": {
                     "lang": "en",
                     "og_type": "website",
                     "primary_color": "#1a1a2e",
                     "secondary_color": "#2d2d44",
                     "accent_color": "#16a085",
                     "text_color": "#333333",
                     "background_color": "#ffffff",
                     "favicon_url": "/favicon.ico",
                     "apple_touch_icon_url": "/apple-touch-icon.png"
                 }
             }'::jsonb,
             ARRAY['site.content_data.company_name', 'page.title', 'page.meta_description', 'site.brand_assets']
         ) ON CONFLICT (name) DO UPDATE SET
    html_template = EXCLUDED.html_template,
    input_schema = EXCLUDED.input_schema,
    updated_at = now();


-- ===========================================================================
-- HEADER COMPONENTS
-- ===========================================================================

-- Professional Dark Header
INSERT INTO content_components (
    name,
    display_name,
    function,
    category,
    component_level,
    render_mode,
    description,
    html_template,
    input_schema,
    data_sources,
    semantic_tags
) VALUES (
             'header-professional-dark',
             'Professional Dark Header',
             'site-header',
             'navigation',
             'header',
             'template',
             'Dark professional header with logo, navigation, and mobile menu',
             '<!-- HEADER SOURCE: component-db:header-professional-dark -->
         <header class="site-header site-header--dark">
             <div class="header-container">
                 <a href="/index.html" class="logo">
                     {{#if logo_url}}<img src="{{logo_url}}" alt="{{company_name}}" class="logo-image">{{/if}}
                     <span class="logo-text">{{logo_text}}</span>
                 </a>
                 <button class="mobile-menu-toggle" aria-label="Toggle menu" aria-expanded="false">
                     <span></span><span></span><span></span>
                 </button>
                 <nav class="main-nav" id="main-nav" role="navigation">
                     <ul>
                         {{nav_items_html}}
                     </ul>
                 </nav>
                 {{#if cta_text}}
                 <a href="{{cta_url}}" class="header-cta">{{cta_text}}</a>
                 {{/if}}
             </div>
         </header>
         <style>
         .site-header--dark {
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
             gap: 2rem;
         }
         .logo {
             display: flex;
             align-items: center;
             gap: 0.75rem;
             text-decoration: none;
         }
         .logo-image {
             height: 40px;
             width: auto;
         }
         .logo-text {
             font-size: 1.25rem;
             font-weight: 700;
             color: #fff;
         }
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
         .main-nav a:hover,
         .main-nav a.active {
             color: {{accent_color}};
         }
         .header-cta {
             background: {{accent_color}};
             color: #fff;
             padding: 0.6rem 1.25rem;
             border-radius: 4px;
             text-decoration: none;
             font-weight: 500;
             transition: opacity 0.2s;
         }
         .header-cta:hover {
             opacity: 0.9;
         }
         .mobile-menu-toggle {
             display: none;
             background: none;
             border: none;
             cursor: pointer;
             padding: 0.5rem;
             flex-direction: column;
             gap: 5px;
         }
         .mobile-menu-toggle span {
             display: block;
             width: 24px;
             height: 2px;
             background: #fff;
             transition: transform 0.3s;
         }
         @media (max-width: 768px) {
             .mobile-menu-toggle { display: flex; }
             .main-nav {
                 position: absolute;
                 top: 100%;
                 left: 0;
                 right: 0;
                 background: {{primary_color}};
                 padding: 1rem 2rem;
                 display: none;
                 box-shadow: 0 4px 10px rgba(0,0,0,0.1);
             }
             .main-nav.active { display: block; }
             .main-nav ul {
                 flex-direction: column;
                 gap: 0;
             }
             .main-nav a {
                 display: block;
                 padding: 0.75rem 0;
                 border-bottom: 1px solid rgba(255,255,255,0.1);
             }
             .header-cta { display: none; }
         }
         </style>
         <script>
         document.addEventListener("DOMContentLoaded", function() {
             var toggle = document.querySelector(".mobile-menu-toggle");
             var nav = document.querySelector(".main-nav");
             if (toggle && nav) {
                 toggle.addEventListener("click", function() {
                     var expanded = toggle.getAttribute("aria-expanded") === "true";
                     toggle.setAttribute("aria-expanded", !expanded);
                     nav.classList.toggle("active");
                 });
             }
         });
         </script>',
             '{
                 "required": ["logo_text", "nav_items"],
                 "optional": ["logo_url", "cta_text", "cta_url"],
                 "defaults": {
                     "primary_color": "#1a1a2e",
                     "accent_color": "#16a085",
                     "cta_url": "/contact.html"
                 }
             }'::jsonb,
             ARRAY['site.content_data.company_name', 'site.brand_assets.logo', 'db_sync.navigation'],
             '["professional", "corporate", "consulting", "dark"]'::jsonb
         ) ON CONFLICT (name) DO UPDATE SET
    html_template = EXCLUDED.html_template,
    input_schema = EXCLUDED.input_schema,
    updated_at = now();


-- Minimal Light Header
INSERT INTO content_components (
    name,
    display_name,
    function,
    category,
    component_level,
    render_mode,
    description,
    html_template,
    input_schema,
    data_sources,
    semantic_tags
) VALUES (
             'header-minimal-light',
             'Minimal Light Header',
             'site-header',
             'navigation',
             'header',
             'template',
             'Clean, minimal light header for creative and design sites',
             '<!-- HEADER SOURCE: component-db:header-minimal-light -->
         <header class="site-header site-header--light">
             <div class="header-container">
                 <a href="/index.html" class="logo">
                     {{#if logo_url}}<img src="{{logo_url}}" alt="{{company_name}}" class="logo-image">{{/if}}
                     <span class="logo-text">{{logo_text}}</span>
                 </a>
                 <button class="mobile-menu-toggle" aria-label="Toggle menu">
                     <span></span><span></span><span></span>
                 </button>
                 <nav class="main-nav" id="main-nav">
                     <ul>
                         {{nav_items_html}}
                     </ul>
                 </nav>
             </div>
         </header>
         <style>
         .site-header--light {
             background: #ffffff;
             padding: 1.25rem 0;
             position: sticky;
             top: 0;
             z-index: 1000;
             border-bottom: 1px solid #f0f0f0;
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
             display: flex;
             align-items: center;
             gap: 0.75rem;
             text-decoration: none;
         }
         .logo-image { height: 36px; width: auto; }
         .logo-text {
             font-size: 1.25rem;
             font-weight: 600;
             color: {{primary_color}};
             letter-spacing: -0.02em;
         }
         .main-nav ul {
             display: flex;
             list-style: none;
             margin: 0;
             padding: 0;
             gap: 2.5rem;
         }
         .main-nav a {
             color: #666;
             text-decoration: none;
             font-weight: 400;
             font-size: 0.95rem;
             transition: color 0.2s;
         }
         .main-nav a:hover,
         .main-nav a.active {
             color: {{primary_color}};
         }
         .mobile-menu-toggle {
             display: none;
             background: none;
             border: none;
             cursor: pointer;
             padding: 0.5rem;
             flex-direction: column;
             gap: 5px;
         }
         .mobile-menu-toggle span {
             display: block;
             width: 24px;
             height: 2px;
             background: {{primary_color}};
         }
         @media (max-width: 768px) {
             .mobile-menu-toggle { display: flex; }
             .main-nav {
                 position: absolute;
                 top: 100%;
                 left: 0;
                 right: 0;
                 background: #fff;
                 padding: 1rem 2rem;
                 display: none;
                 border-bottom: 1px solid #f0f0f0;
             }
             .main-nav.active { display: block; }
             .main-nav ul { flex-direction: column; gap: 0; }
             .main-nav a {
                 display: block;
                 padding: 0.75rem 0;
                 border-bottom: 1px solid #f5f5f5;
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
                 "required": ["logo_text", "nav_items"],
                 "optional": ["logo_url"],
                 "defaults": {
                     "primary_color": "#1a1a2e"
                 }
             }'::jsonb,
             ARRAY['site.content_data.company_name', 'site.brand_assets.logo', 'db_sync.navigation'],
             '["minimal", "light", "creative", "design", "agency"]'::jsonb
         ) ON CONFLICT (name) DO UPDATE SET
    html_template = EXCLUDED.html_template,
    input_schema = EXCLUDED.input_schema,
    updated_at = now();


-- ===========================================================================
-- FOOTER COMPONENTS
-- ===========================================================================

-- 4-Column Footer
INSERT INTO content_components (
    name,
    display_name,
    function,
    category,
    component_level,
    render_mode,
    description,
    html_template,
    input_schema,
    data_sources,
    semantic_tags
) VALUES (
             'footer-4-column',
             '4-Column Footer',
             'site-footer',
             'navigation',
             'footer',
             'template',
             'Standard 4-column footer with brand, links, services, and contact',
             '<!-- FOOTER SOURCE: component-db:footer-4-column -->
         <footer class="site-footer">
             <div class="footer-container">
                 <div class="footer-brand">
                     <h3>{{company_name}}</h3>
                     <p>{{tagline}}</p>
                     {{#if social_links}}
                     <div class="social-links">
                         {{social_links_html}}
                     </div>
                     {{/if}}
                 </div>
                 <div class="footer-links">
                     <h4>Quick Links</h4>
                     <ul>
                         {{nav_items_html}}
                     </ul>
                 </div>
                 <div class="footer-services">
                     <h4>Services</h4>
                     <ul>
                         {{#each services}}
                         <li><a href="/services.html#{{this.slug}}">{{this.name}}</a></li>
                         {{/each}}
                     </ul>
                 </div>
                 <div class="footer-contact">
                     <h4>Contact</h4>
                     {{#if contact_email}}<p><a href="mailto:{{contact_email}}">{{contact_email}}</a></p>{{/if}}
                     {{#if contact_phone}}<p><a href="tel:{{contact_phone}}">{{contact_phone}}</a></p>{{/if}}
                     {{#if address}}<p>{{address}}</p>{{/if}}
                 </div>
             </div>
             <div class="footer-bottom">
                 <div class="footer-bottom-container">
                     <p>&copy; {{year}} {{company_name}}. All rights reserved.</p>
                     <div class="footer-legal">
                         <a href="/privacy.html">Privacy Policy</a>
                         <a href="/terms.html">Terms of Service</a>
                     </div>
                 </div>
             </div>
         </footer>
         <style>
         .site-footer {
             background: {{primary_color}};
             color: rgba(255,255,255,0.9);
             padding: 4rem 0 0;
             margin-top: auto;
         }
         .footer-container {
             max-width: 1200px;
             margin: 0 auto;
             padding: 0 2rem;
             display: grid;
             grid-template-columns: 2fr 1fr 1fr 1fr;
             gap: 3rem;
         }
         .footer-brand h3 {
             color: #fff;
             margin: 0 0 0.75rem;
             font-size: 1.25rem;
         }
         .footer-brand p {
             color: rgba(255,255,255,0.7);
             margin: 0 0 1.5rem;
             line-height: 1.6;
         }
         .social-links {
             display: flex;
             gap: 1rem;
         }
         .social-links a {
             color: rgba(255,255,255,0.7);
             transition: color 0.2s;
         }
         .social-links a:hover {
             color: {{accent_color}};
         }
         .footer-links h4,
         .footer-services h4,
         .footer-contact h4 {
             color: #fff;
             margin: 0 0 1rem;
             font-size: 1rem;
             font-weight: 600;
         }
         .footer-links ul,
         .footer-services ul {
             list-style: none;
             padding: 0;
             margin: 0;
         }
         .footer-links li,
         .footer-services li {
             margin-bottom: 0.5rem;
         }
         .footer-links a,
         .footer-services a,
         .footer-contact a {
             color: rgba(255,255,255,0.7);
             text-decoration: none;
             transition: color 0.2s;
         }
         .footer-links a:hover,
         .footer-services a:hover,
         .footer-contact a:hover {
             color: #fff;
         }
         .footer-contact p {
             margin: 0 0 0.5rem;
             color: rgba(255,255,255,0.7);
         }
         .footer-bottom {
             margin-top: 3rem;
             padding: 1.5rem 0;
             border-top: 1px solid rgba(255,255,255,0.1);
         }
         .footer-bottom-container {
             max-width: 1200px;
             margin: 0 auto;
             padding: 0 2rem;
             display: flex;
             justify-content: space-between;
             align-items: center;
         }
         .footer-bottom p {
             margin: 0;
             color: rgba(255,255,255,0.5);
             font-size: 0.9rem;
         }
         .footer-legal {
             display: flex;
             gap: 2rem;
         }
         .footer-legal a {
             color: rgba(255,255,255,0.5);
             text-decoration: none;
             font-size: 0.9rem;
         }
         .footer-legal a:hover {
             color: rgba(255,255,255,0.8);
         }
         @media (max-width: 768px) {
             .footer-container {
                 grid-template-columns: 1fr;
                 gap: 2rem;
             }
             .footer-bottom-container {
                 flex-direction: column;
                 gap: 1rem;
                 text-align: center;
             }
         }
         </style>',
             '{
                 "required": ["company_name"],
                 "optional": ["tagline", "contact_email", "contact_phone", "address", "services", "social_links"],
                 "defaults": {
                     "primary_color": "#1a1a2e",
                     "accent_color": "#16a085"
                 }
             }'::jsonb,
             ARRAY['site.content_data.company_name', 'site.content_data.tagline', 'site.content_data.services', 'site.content_data.contact_email'],
             '["professional", "corporate", "4-column"]'::jsonb
         ) ON CONFLICT (name) DO UPDATE SET
    html_template = EXCLUDED.html_template,
    input_schema = EXCLUDED.input_schema,
    updated_at = now();


-- Simple Footer
INSERT INTO content_components (
    name,
    display_name,
    function,
    category,
    component_level,
    render_mode,
    description,
    html_template,
    input_schema,
    data_sources,
    semantic_tags
) VALUES (
             'footer-simple',
             'Simple Footer',
             'site-footer',
             'navigation',
             'footer',
             'template',
             'Minimal single-line footer for simple sites',
             '<!-- FOOTER SOURCE: component-db:footer-simple -->
         <footer class="site-footer site-footer--simple">
             <div class="footer-container">
                 <p>&copy; {{year}} {{company_name}}. All rights reserved.</p>
                 <nav class="footer-nav">
                     <a href="/privacy.html">Privacy</a>
                     <a href="/terms.html">Terms</a>
                     {{#if contact_email}}<a href="mailto:{{contact_email}}">Contact</a>{{/if}}
                 </nav>
             </div>
         </footer>
         <style>
         .site-footer--simple {
             background: #f8f9fa;
             padding: 2rem 0;
             margin-top: auto;
             border-top: 1px solid #e9ecef;
         }
         .site-footer--simple .footer-container {
             max-width: 1200px;
             margin: 0 auto;
             padding: 0 2rem;
             display: flex;
             justify-content: space-between;
             align-items: center;
         }
         .site-footer--simple p {
             margin: 0;
             color: #6c757d;
             font-size: 0.9rem;
         }
         .footer-nav {
             display: flex;
             gap: 2rem;
         }
         .footer-nav a {
             color: #6c757d;
             text-decoration: none;
             font-size: 0.9rem;
         }
         .footer-nav a:hover {
             color: {{primary_color}};
         }
         @media (max-width: 768px) {
             .site-footer--simple .footer-container {
                 flex-direction: column;
                 gap: 1rem;
                 text-align: center;
             }
         }
         </style>',
             '{
                 "required": ["company_name"],
                 "optional": ["contact_email"],
                 "defaults": {
                     "primary_color": "#1a1a2e"
                 }
             }'::jsonb,
             ARRAY['site.content_data.company_name', 'site.content_data.contact_email'],
             '["minimal", "simple", "light"]'::jsonb
         ) ON CONFLICT (name) DO UPDATE SET
    html_template = EXCLUDED.html_template,
    input_schema = EXCLUDED.input_schema,
    updated_at = now();


-- ===========================================================================
-- CLOSING BODY TAG COMPONENT
-- ===========================================================================

INSERT INTO content_components (
    name,
    display_name,
    function,
    category,
    component_level,
    render_mode,
    description,
    html_template
) VALUES (
             'body-close',
             'Body Close',
             'body-close',
             'structure',
             'element',
             'template',
             'Closing body and html tags',
             '</body>
         </html>'
         ) ON CONFLICT (name) DO UPDATE SET
    html_template = EXCLUDED.html_template,
    updated_at = now();


-- ===========================================================================
-- UPDATE style_collections WITH COMPONENT REFERENCES
-- ===========================================================================

-- Update professional-dark to use our new components
UPDATE style_collections SET
                             header_component_id = (SELECT id FROM content_components WHERE name = 'header-professional-dark'),
                             footer_component_id = (SELECT id FROM content_components WHERE name = 'footer-4-column')
WHERE name = 'professional-dark';

-- Update minimal-light
UPDATE style_collections SET
                             header_component_id = (SELECT id FROM content_components WHERE name = 'header-minimal-light'),
                             footer_component_id = (SELECT id FROM content_components WHERE name = 'footer-simple')
WHERE name = 'minimal-light';


COMMIT;

-- ===========================================================================
-- VERIFICATION
-- ===========================================================================
-- SELECT name, function, component_level, render_mode FROM content_components
-- WHERE component_level IN ('head', 'header', 'footer') ORDER BY component_level, name;
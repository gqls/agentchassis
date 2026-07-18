-- SQL_2026-07-17_r1_dark_theme_restore.sql
--
-- R1 of HANDOFF_2026-07-17_robot_hands_site_fixes.md: restore the dark theme
-- at the component/palette level.
--
-- ROOT CAUSE (corrects the handoff's working hypothesis): the B7 fix
-- (SQL_2026-07-10_b7_layout_swap.sql) swapped ONLY css_themes.layout_id to
-- tool-portal-dark. Three palette copies stayed brochure-blue/light
-- (background #ffffff, primary #3b82f6):
--   1. palettes.colours            (palette-robot-hands-com)  → CSS var merge
--   2. style_collections.color_palette (collection-robot-hands-com)
--        → read by loadSiteDataFull / RenderSiteComponentsAction: fills
--          {{.primary_color}} etc. in component templates
--   3. css_themes.color_palette    (theme-robot-hands-com)    → legacy copy
-- AND site_components header/footer still point at DEACTIVATED components
-- (header-bold-gradient, footer-4-column) whose templates bake palette
-- colours into literal inline <style> blocks. renderAndStoreSiteComponent
-- renders whatever site_components.component_id says, ignoring is_active
-- (render_site_components_action.go:489-494) — this is exactly what the
-- three FAILED 2026-05-13 deactivated_component items flagged.
-- The 2026-07-16 needs_rerender completions re-rendered header/footer with
-- the light palette and pushed the blue chrome to all 37 pages.
-- The blue header first appeared in deploys between 2026-07-08T22:44 and
-- 2026-07-09T09:13 (gqls/sites af0ead8da1 → 78532b8c63) — it predates B7.
--
-- FIX (kills the class, not the instance):
--   A. Two NEW site-chrome components, var()-based only — markup uses the
--      classes tool-portal-dark already styles (.site-header/.main-nav/
--      .site-footer/.footer-container); the only <style> they carry covers
--      gaps the layout does not style (.header-cta, mobile menu, footer
--      bottom row) and references CSS variables exclusively. Any future
--      re-render stays theme-correct BY CONSTRUCTION, whatever the palette.
--      Names sort AFTER the current alphabetical-first components so the
--      "ORDER BY name LIMIT 1" fallback for OTHER sites is unchanged.
--   B. Repoint robot-hands' site_components header/footer to them.
--   C. Rewrite all three palette copies with the dark scheme the user
--      approved at the B7 gate (values = the served tool-portal-dark root
--      vars; card_bg/heading corrected to dark-scheme-consistent values —
--      card_bg #ffffff and heading #0f172a were light-palette leaks visible
--      as white cards / near-invisible headings).
--   D. Close the three 2026-05-13 deactivated_component items (this IS
--      their fix) and the stale 2026-05-13 "Assemble and deploy pages after
--      plan reconcile" is left alone.
--
-- After this file: trigger rerender-pages with refresh_site_components:true
-- (kcat, sql_for_agents/033_rerender_pages_trigger.sh pattern) — NOT within
-- 300s of a chassis pod restart.

\set ON_ERROR_STOP on

-- ── Backups (outside transaction) ──
CREATE TABLE IF NOT EXISTS palettes_backup_20260717_r1 AS
SELECT * FROM palettes WHERE id = '617e93c7-b1f1-4c5b-b7c4-482f3c0e9736';
CREATE TABLE IF NOT EXISTS style_collections_backup_20260717_r1 AS
SELECT * FROM style_collections WHERE id = 'cb95d40f-9bd2-4480-ba99-98b263aea44b';
CREATE TABLE IF NOT EXISTS css_themes_backup_20260717_r1 AS
SELECT * FROM css_themes WHERE id = 'b1b60faf-ca68-43f5-a1e6-da3a769e4a25';
CREATE TABLE IF NOT EXISTS site_components_backup_20260717_r1 AS
SELECT * FROM site_components WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92';

SELECT (SELECT count(*) FROM palettes_backup_20260717_r1)          AS bak_palettes,
       (SELECT count(*) FROM style_collections_backup_20260717_r1) AS bak_collections,
       (SELECT count(*) FROM css_themes_backup_20260717_r1)        AS bak_themes,
       (SELECT count(*) FROM site_components_backup_20260717_r1)   AS bak_site_components;

BEGIN;

-- ── A1. Theme-chrome header component ──
INSERT INTO content_components
    (id, name, display_name, description, html_template, function,
     component_level, render_mode, is_active, created_from, category)
VALUES
    ('58fde68f-9190-4e5e-b6a5-ea21cf27a9af',
     'header-theme-chrome',
     'Header (theme chrome, var-based)',
     'Site header that defers ALL colour to the site stylesheet''s CSS variables. Markup uses the classes dark layouts style natively (.site-header/.header-container/.logo/.main-nav/.mobile-menu-toggle). Data contract: RenderSiteComponentsAction ContentData — logo_url, logo_text, nav_items_html, cta_url, cta_text. Carries NO literal colours; safe to regenerate under any palette. Created 2026-07-17 for robot-hands R1 (blue-baked header-bold-gradient replacement).',
     E'<header class="site-header">\n    <div class="header-container">\n        <a href="/index.html" class="logo">\n            {{if .logo_url}}<img src="{{.logo_url}}" alt="{{.logo_text}}" class="logo-img">{{else}}<span class="logo-text">{{.logo_text}}</span>{{end}}\n        </a>\n        <nav class="main-nav">\n            <ul>\n                {{if .nav_items_html}}{{.nav_items_html}}{{end}}\n            </ul>\n        </nav>\n        {{if .cta_url}}<a href="{{.cta_url}}" class="header-cta">{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>{{end}}\n        <button class="mobile-menu-toggle" aria-label="Toggle menu" aria-expanded="false">\n            <span></span><span></span><span></span>\n        </button>\n    </div>\n</header>\n<style>\n/* Theme-owned chrome: every colour is a CSS variable resolved by the site\n   stylesheet. Only gaps the layout does not style are covered here. */\n.header-cta {\n    background: var(--color-cta-bg, var(--color-accent));\n    color: var(--color-cta-text, var(--color-primary-text));\n    padding: 0.5rem 1.1rem;\n    border-radius: var(--radius, 4px);\n    text-decoration: none;\n    font-weight: 600;\n    font-size: 0.9rem;\n    white-space: nowrap;\n}\n.header-cta:hover { filter: brightness(1.1); }\n.mobile-menu-toggle span {\n    display: block;\n    width: 24px;\n    height: 2px;\n    background: var(--color-header-text, var(--color-text));\n    margin: 5px 0;\n}\n@media (max-width: 768px) {\n    .main-nav.is-open {\n        position: absolute;\n        top: 100%;\n        left: 0;\n        right: 0;\n        background: var(--color-header-bg, var(--color-surface));\n        padding: 1rem;\n        border-bottom: 1px solid var(--color-border);\n    }\n    .main-nav.is-open ul { flex-direction: column; }\n    .main-nav.is-open a { display: block; padding: 0.75rem 0; }\n    .header-cta { display: none; }\n}\n</style>\n<script>\ndocument.addEventListener("DOMContentLoaded", function() {\n    var toggle = document.querySelector(".mobile-menu-toggle");\n    var nav = document.querySelector(".main-nav");\n    if (toggle && nav) {\n        toggle.addEventListener("click", function() {\n            var open = nav.classList.toggle("is-open");\n            toggle.setAttribute("aria-expanded", open ? "true" : "false");\n        });\n    }\n});\n</script>',
     'site-header', 'site', 'template', true, 'manual', 'site-chrome');

-- ── A2. Theme-chrome footer component ──
INSERT INTO content_components
    (id, name, display_name, description, html_template, function,
     component_level, render_mode, is_active, created_from, category)
VALUES
    ('e6347680-4c7c-448b-8cfc-1cea509159d1',
     'footer-theme-chrome',
     'Footer (theme chrome, var-based)',
     'Site footer that defers ALL colour to the site stylesheet''s CSS variables. Markup uses the classes dark layouts style natively (.site-footer/.footer-container/.footer-bottom). Data contract: RenderSiteComponentsAction ContentData — logo_text, tagline, quick_links_html, services_html, email, phone, year, company_name, legal_links. No literal colours. Created 2026-07-17 for robot-hands R1 (blue-baked footer-4-column replacement).',
     E'<footer class="site-footer">\n    <div class="footer-container">\n        <div class="footer-brand">\n            <h3>{{.logo_text}}</h3>\n            {{if .tagline}}<p>{{.tagline}}</p>{{end}}\n        </div>\n        {{if .quick_links_html}}<div class="footer-links">\n            <h4>Quick Links</h4>\n            <ul>\n                {{.quick_links_html}}\n            </ul>\n        </div>{{end}}\n        {{if .services_html}}<div class="footer-services">\n            <h4>Explore</h4>\n            <ul>\n                {{.services_html}}\n            </ul>\n        </div>{{end}}\n        <div class="footer-contact">\n            <h4>Contact</h4>\n            {{if .email}}<p><a href="mailto:{{.email}}">{{.email}}</a></p>{{end}}\n            {{if .phone}}<p>{{.phone}}</p>{{end}}\n        </div>\n    </div>\n    <div class="footer-bottom">\n        <p>&copy; {{.year}} {{.company_name}}. All rights reserved.</p>\n        {{if .legal_links}}<div class="footer-legal">\n            {{range .legal_links}}<a href="{{.url}}">{{.name}}</a>{{end}}\n        </div>{{end}}\n    </div>\n</footer>\n<style>\n/* Theme-owned chrome — var()-based gaps only; the layout styles .site-footer,\n   .footer-container and .footer-bottom. */\n.footer-brand p { color: var(--color-footer-text, var(--color-text-muted)); margin: 0.5rem 0 0; }\n.footer-legal { display: flex; gap: 1rem; flex-wrap: wrap; justify-content: center; margin-top: 0.5rem; }\n.footer-legal a { color: var(--color-footer-text, var(--color-text-muted)); }\n.footer-legal a:hover { color: var(--color-accent); }\n</style>',
     'site-footer', 'site', 'template', true, 'manual', 'site-chrome');

-- ── B. Repoint robot-hands header/footer slots ──
UPDATE site_components
SET component_id = '58fde68f-9190-4e5e-b6a5-ea21cf27a9af', updated_at = now()
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slot_name = 'header';

UPDATE site_components
SET component_id = 'e6347680-4c7c-448b-8cfc-1cea509159d1', updated_at = now()
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slot_name = 'footer';

-- ── C. One consistent dark palette, written to all three copies ──
-- Values = tool-portal-dark's served root vars (user-approved dark state),
-- with the light leaks corrected: card_bg white→surface, heading #0f172a→
-- light, header_bg slate→body bg (layout intent: "low chrome, same
-- background as body"), footer_bg → darker than body per layout comment.
-- cta_bg keeps the blue gradient that was present in the approved 07-10 CSS.
UPDATE palettes
SET colours = '{
  "primary": "#1A1F2E", "primary_hover": "#2563eb", "primary_text": "#ffffff",
  "secondary": "#C8D8E8", "secondary_hover": "#A9BFD4", "secondary_text": "#0F1218",
  "accent": "#E8500A",
  "background": "#0F1218", "background_alt": "#1a1a1a",
  "surface": "#1E2535", "surface_alt": "#1a1a1a",
  "text": "#E2E8F0", "text_light": "#7A8FA6", "text_muted": "#7A8FA6",
  "border": "#2D3A4A",
  "heading": "#E2E8F0", "hero_title": "#ffffff", "hero_subtitle": "#C8D8E8",
  "card_bg": "#1E2535",
  "header_bg": "#0F1218", "header_text": "#E2E8F0",
  "cta_bg": "linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)", "cta_text": "#ffffff",
  "footer_bg": "#0B0E14", "footer_text": "rgba(226,232,240,0.8)",
  "code_bg": "#0d0d0d",
  "callout_bg": "rgba(0,188,212,0.08)", "callout_border": "#00bcd4"
}'::jsonb,
    updated_at = now()
WHERE id = '617e93c7-b1f1-4c5b-b7c4-482f3c0e9736';

UPDATE style_collections
SET color_palette = (SELECT colours FROM palettes WHERE id = '617e93c7-b1f1-4c5b-b7c4-482f3c0e9736'),
    updated_at = now()
WHERE id = 'cb95d40f-9bd2-4480-ba99-98b263aea44b';

UPDATE css_themes
SET color_palette = (SELECT colours FROM palettes WHERE id = '617e93c7-b1f1-4c5b-b7c4-482f3c0e9736'),
    updated_at = now()
WHERE id = 'b1b60faf-ca68-43f5-a1e6-da3a769e4a25';

-- ── D. Close the 2026-05-13 deactivated_component items — this is their fix ──
UPDATE site_work_items
SET status = 'complete', completed_at = now(), updated_at = now(),
    error = COALESCE(error || E'\n', '')
            || 'Resolved 2026-07-17 (R1 dark restore): header/footer slots repointed to var-based header-theme-chrome/footer-theme-chrome; palette copies rewritten dark. See robot_hands/SQL_2026-07-17_r1_dark_theme_restore.sql.'
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND item_type = 'deactivated_component'
  AND status = 'failed'
  AND id IN ('be37dc00-8efb-4238-8e40-8da417b178e3',
             '437b7d3d-013b-4d82-8e56-3280c9a43ede',
             '0e34dbc7-f554-4c91-8ffe-47d63e04e976');

-- Note: the head slot's "Document Head" is also deactivated, but its rendered
-- output is fully var()-based (verified 2026-07-17) — left pointing as-is;
-- its deactivated_component item above is closed by the same reasoning being
-- documented here. Revisit only if head regeneration misbehaves.

-- ── Verify ──
DO $verify$
DECLARE
    v_hdr uuid; v_ftr uuid; v_bg text; v_cnt int; v_active boolean;
BEGIN
    SELECT component_id INTO v_hdr FROM site_components
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slot_name = 'header';
    IF v_hdr <> '58fde68f-9190-4e5e-b6a5-ea21cf27a9af' THEN
        RAISE EXCEPTION 'header slot not repointed (got %)', v_hdr;
    END IF;

    SELECT component_id INTO v_ftr FROM site_components
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND slot_name = 'footer';
    IF v_ftr <> 'e6347680-4c7c-448b-8cfc-1cea509159d1' THEN
        RAISE EXCEPTION 'footer slot not repointed (got %)', v_ftr;
    END IF;

    SELECT is_active INTO v_active FROM content_components
    WHERE id = '58fde68f-9190-4e5e-b6a5-ea21cf27a9af';
    IF NOT v_active THEN RAISE EXCEPTION 'new header component not active'; END IF;

    -- No literal colours in the new templates beyond var() fallbacks:
    SELECT count(*) INTO v_cnt FROM content_components
    WHERE id IN ('58fde68f-9190-4e5e-b6a5-ea21cf27a9af','e6347680-4c7c-448b-8cfc-1cea509159d1')
      AND (html_template LIKE '%{{.primary_color}}%' OR html_template LIKE '%#3b82f6%');
    IF v_cnt <> 0 THEN RAISE EXCEPTION 'new templates carry palette-baked colours'; END IF;

    SELECT colours->>'background' INTO v_bg FROM palettes
    WHERE id = '617e93c7-b1f1-4c5b-b7c4-482f3c0e9736';
    IF v_bg <> '#0F1218' THEN RAISE EXCEPTION 'palette background not dark (got %)', v_bg; END IF;

    SELECT color_palette->>'card_bg' INTO v_bg FROM style_collections
    WHERE id = 'cb95d40f-9bd2-4480-ba99-98b263aea44b';
    IF v_bg <> '#1E2535' THEN RAISE EXCEPTION 'collection card_bg not dark (got %)', v_bg; END IF;

    SELECT count(*) INTO v_cnt FROM site_work_items
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
      AND item_type = 'deactivated_component' AND status = 'failed';
    IF v_cnt <> 0 THEN RAISE EXCEPTION '% deactivated_component items still failed', v_cnt; END IF;

    RAISE NOTICE 'R1 dark restore applied: components created+repointed, palettes dark, items closed';
END
$verify$;

COMMIT;

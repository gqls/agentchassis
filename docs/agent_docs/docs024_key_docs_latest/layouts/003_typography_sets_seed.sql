-- =====================================================================
-- MIGRATION 025 — Phase 3 Step 1: Typography Sets Seed
-- =====================================================================
-- Seeds the 6 typography sets described in the migration plan section 8.
-- Each set is a `fonts` JSONB (body/heading/mono) + `scale` JSONB
-- (base_size, line_height, h1_ratio). Layouts reference these via
-- the `{{typo "key" "fallback"}}` template helper.
--
-- Idempotent: ON CONFLICT (name) DO UPDATE.
-- =====================================================================

BEGIN;

-- ── 1. sans-modern ──────────────────────────────────────────────────
-- Clean, neutral, the default for most brochure/tool/storefront layouts.
INSERT INTO typography_sets (
    name, display_name, description, fonts, scale, category, industry_tags, origin, is_active
) VALUES (
    'sans-modern',
    'Sans — Modern',
    'Clean, neutral, reads well at any size. The default for most brochure, technical, commerce, and tool layouts. Inter with system fallbacks.',
    '{
        "font_family": "''Inter'', -apple-system, BlinkMacSystemFont, ''Segoe UI'', Roboto, sans-serif",
        "heading_font": "inherit",
        "mono_font": "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    }'::jsonb,
    '{
        "base_size": "16px",
        "line_height": "1.6",
        "h1_ratio": "2.5",
        "h2_ratio": "1.875",
        "h3_ratio": "1.375",
        "h4_ratio": "1.125"
    }'::jsonb,
    'sans',
    ARRAY['general', 'brochure', 'tech', 'commerce']::text[],
    'seed',
    true
)
ON CONFLICT (name) DO UPDATE SET
    display_name  = EXCLUDED.display_name,
    description   = EXCLUDED.description,
    fonts         = EXCLUDED.fonts,
    scale         = EXCLUDED.scale,
    category      = EXCLUDED.category,
    industry_tags = EXCLUDED.industry_tags,
    is_active     = EXCLUDED.is_active,
    updated_at    = NOW();

-- ── 2. serif-editorial ──────────────────────────────────────────────
-- Warm, reading-first, magazine feel.
INSERT INTO typography_sets (
    name, display_name, description, fonts, scale, category, industry_tags, origin, is_active
) VALUES (
    'serif-editorial',
    'Serif — Editorial',
    'Warm, reading-first, magazine-grade. Merriweather headings over Lato body. Pairs with soft-editorial, magazine-grid, industry-hub layouts.',
    '{
        "font_family": "''Lato'', Georgia, ''Times New Roman'', serif",
        "heading_font": "''Merriweather'', Georgia, ''Times New Roman'', serif",
        "mono_font": "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    }'::jsonb,
    '{
        "base_size": "17px",
        "line_height": "1.75",
        "h1_ratio": "2.5",
        "h2_ratio": "1.875",
        "h3_ratio": "1.375",
        "h4_ratio": "1.125"
    }'::jsonb,
    'serif',
    ARRAY['editorial', 'magazine', 'long-form', 'wellness', 'lifestyle']::text[],
    'seed',
    true
)
ON CONFLICT (name) DO UPDATE SET
    display_name  = EXCLUDED.display_name,
    description   = EXCLUDED.description,
    fonts         = EXCLUDED.fonts,
    scale         = EXCLUDED.scale,
    category      = EXCLUDED.category,
    industry_tags = EXCLUDED.industry_tags,
    is_active     = EXCLUDED.is_active,
    updated_at    = NOW();

-- ── 3. display-bold ─────────────────────────────────────────────────
-- High-impact, condensed, uppercase-friendly. For high-energy /
-- brochure-bold layouts.
INSERT INTO typography_sets (
    name, display_name, description, fonts, scale, category, industry_tags, origin, is_active
) VALUES (
    'display-bold',
    'Display — Bold',
    'High-impact display headings, uppercase-friendly. Archivo Black over Inter body. Pairs with high-energy and brochure-bold layouts.',
    '{
        "font_family": "''Inter'', -apple-system, BlinkMacSystemFont, ''Segoe UI'', sans-serif",
        "heading_font": "''Archivo Black'', ''Arial Black'', Impact, sans-serif",
        "mono_font": "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    }'::jsonb,
    '{
        "base_size": "16px",
        "line_height": "1.5",
        "h1_ratio": "3.5",
        "h2_ratio": "2.5",
        "h3_ratio": "1.5",
        "h4_ratio": "1.125"
    }'::jsonb,
    'display',
    ARRAY['high-energy', 'combat-sports', 'fitness', 'conversion', 'events']::text[],
    'seed',
    true
)
ON CONFLICT (name) DO UPDATE SET
    display_name  = EXCLUDED.display_name,
    description   = EXCLUDED.description,
    fonts         = EXCLUDED.fonts,
    scale         = EXCLUDED.scale,
    category      = EXCLUDED.category,
    industry_tags = EXCLUDED.industry_tags,
    is_active     = EXCLUDED.is_active,
    updated_at    = NOW();

-- ── 4. mono-technical ───────────────────────────────────────────────
-- Code-friendly, docs, utilitarian. IBM Plex Mono for code areas,
-- system-ui for body text.
INSERT INTO typography_sets (
    name, display_name, description, fonts, scale, category, industry_tags, origin, is_active
) VALUES (
    'mono-technical',
    'Mono — Technical',
    'Code-friendly, docs, utilitarian. IBM Plex Mono for code areas, system-ui for body. Default for docs-sidebar. Pairs well with tool-first-landing for developer tools.',
    '{
        "font_family": "system-ui, -apple-system, BlinkMacSystemFont, ''Segoe UI'', Roboto, sans-serif",
        "heading_font": "inherit",
        "mono_font": "''IBM Plex Mono'', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    }'::jsonb,
    '{
        "base_size": "15px",
        "line_height": "1.75",
        "h1_ratio": "2",
        "h2_ratio": "1.5",
        "h3_ratio": "1.25",
        "h4_ratio": "1.0625"
    }'::jsonb,
    'mono',
    ARRAY['developer-docs', 'api-reference', 'knowledge-base', 'technical-guide']::text[],
    'seed',
    true
)
ON CONFLICT (name) DO UPDATE SET
    display_name  = EXCLUDED.display_name,
    description   = EXCLUDED.description,
    fonts         = EXCLUDED.fonts,
    scale         = EXCLUDED.scale,
    category      = EXCLUDED.category,
    industry_tags = EXCLUDED.industry_tags,
    is_active     = EXCLUDED.is_active,
    updated_at    = NOW();

-- ── 5. serif-classical ──────────────────────────────────────────────
-- Formal, elegant, luxury feel. Cormorant Garamond.
-- Used by premium-elegant theme override.
INSERT INTO typography_sets (
    name, display_name, description, fonts, scale, category, industry_tags, origin, is_active
) VALUES (
    'serif-classical',
    'Serif — Classical',
    'Formal, elegant, luxury feel. Cormorant Garamond with Georgia fallbacks. Pairs with premium-elegant theme on technical-precise layout, or with portfolio-kinetic for editorial design portfolios.',
    '{
        "font_family": "''Cormorant Garamond'', Georgia, ''Times New Roman'', serif",
        "heading_font": "inherit",
        "mono_font": "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    }'::jsonb,
    '{
        "base_size": "17px",
        "line_height": "1.7",
        "h1_ratio": "3",
        "h2_ratio": "2.25",
        "h3_ratio": "1.5",
        "h4_ratio": "1.125"
    }'::jsonb,
    'serif',
    ARRAY['luxury', 'premium', 'editorial', 'fashion', 'fine-dining']::text[],
    'seed',
    true
)
ON CONFLICT (name) DO UPDATE SET
    display_name  = EXCLUDED.display_name,
    description   = EXCLUDED.description,
    fonts         = EXCLUDED.fonts,
    scale         = EXCLUDED.scale,
    category      = EXCLUDED.category,
    industry_tags = EXCLUDED.industry_tags,
    is_active     = EXCLUDED.is_active,
    updated_at    = NOW();

-- ── 6. sans-friendly ────────────────────────────────────────────────
-- Rounded, approachable, conversational. Nunito.
INSERT INTO typography_sets (
    name, display_name, description, fonts, scale, category, industry_tags, origin, is_active
) VALUES (
    'sans-friendly',
    'Sans — Friendly',
    'Rounded, approachable, conversational. Nunito with Segoe UI fallbacks. Pairs with soft-editorial for friendly-small-business themes (bakery, warm-friendly) as an alternative to serif-editorial.',
    '{
        "font_family": "''Nunito'', ''Segoe UI'', -apple-system, BlinkMacSystemFont, sans-serif",
        "heading_font": "inherit",
        "mono_font": "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    }'::jsonb,
    '{
        "base_size": "16px",
        "line_height": "1.65",
        "h1_ratio": "2.25",
        "h2_ratio": "1.75",
        "h3_ratio": "1.25",
        "h4_ratio": "1.0625"
    }'::jsonb,
    'sans',
    ARRAY['friendly', 'small-business', 'consumer', 'wellness']::text[],
    'seed',
    true
)
ON CONFLICT (name) DO UPDATE SET
    display_name  = EXCLUDED.display_name,
    description   = EXCLUDED.description,
    fonts         = EXCLUDED.fonts,
    scale         = EXCLUDED.scale,
    category      = EXCLUDED.category,
    industry_tags = EXCLUDED.industry_tags,
    is_active     = EXCLUDED.is_active,
    updated_at    = NOW();

-- Verification
DO $verify$
DECLARE
    expected_count int := 6;
    actual_count   int;
BEGIN
    SELECT COUNT(*) INTO actual_count FROM typography_sets WHERE origin = 'seed';
    IF actual_count < expected_count THEN
        RAISE EXCEPTION 'Expected at least % typography_sets, found %', expected_count, actual_count;
    END IF;
    RAISE NOTICE 'typography_sets seed complete: % rows with origin=seed', actual_count;
END
$verify$;

COMMIT;

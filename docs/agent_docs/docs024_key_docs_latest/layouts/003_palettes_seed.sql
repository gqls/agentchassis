-- =====================================================================
-- MIGRATION 025 — Phase 3 Step 2: Palette Extraction
-- =====================================================================
-- Extracts the 13 palettes from css_themes.css_content (raw CSS text)
-- into the `palettes` table. Uses a helper function to parse
--   --color-KEY: VALUE;
-- declarations out of each theme's :root block.
--
-- Notes:
--   - Diagnostic run (Phase 3 preflight) confirmed:
--       * color_palette JSONB is NULL for all 14 css_themes rows
--       * 13 rows have palette data in css_content (CSS text)
--       * 1 row (standard-brochure) has no palette of its own —
--         it's the layout carrier. It gets mapped to the `default`
--         palette in Phase 3 Step 3 (the theme-mapping UPDATE).
--   - Key naming: CSS `--color-primary-hover` becomes JSONB key
--     `primary_hover` (strip `--color-` prefix, kebab→snake). This
--     matches the {{palette "primary_hover" "..."}} template helpers
--     in the Phase 1 layout files.
--   - Idempotent: INSERT ... ON CONFLICT (name) DO UPDATE.
--
-- Runs after:
--   - Phase 2 migration (palettes table exists)
--   - 003_typography_sets_seed.sql (typography_sets populated —
--     independent but logical order)
--   - 003_layouts_seed_driver.sql (layouts populated — same)
-- =====================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ── 1. Helper function: parse CSS :root block into a JSONB map ──
--
-- Input: raw CSS text that includes a `:root { ... }` block.
-- Output: JSONB object mapping colour-variable-name → value string,
--         with `--color-` prefix stripped and kebab→snake case.
--
-- Only `--color-*` variables are extracted. Non-colour declarations
-- (e.g. --border-radius, --shadow) are structure tokens, not palette,
-- and deliberately excluded.
--
-- Marked IMMUTABLE because the function is deterministic and has no
-- side effects; this lets the planner reuse results and supports
-- functional indexes if ever needed.
CREATE OR REPLACE FUNCTION _extract_css_palette(css_text text)
RETURNS jsonb AS $func$
DECLARE
    result       jsonb := '{}'::jsonb;
    root_content text;
    match_rec    record;
    color_key    text;
    color_val    text;
BEGIN
    -- Guard: null or empty input → empty palette
    IF css_text IS NULL OR length(css_text) = 0 THEN
        RETURN '{}'::jsonb;
    END IF;

    -- Extract the content inside the first `:root { ... }` block.
    -- Non-greedy match to first closing brace (CSS :root typically
    -- contains only declarations, no nested rules).
    root_content := substring(css_text FROM ':root\s*\{([^}]*)\}');

    IF root_content IS NULL OR length(trim(root_content)) = 0 THEN
        RETURN '{}'::jsonb;
    END IF;

    -- Iterate over all `--color-KEY: VALUE;` declarations.
    -- regexp_matches with 'g' flag returns one row per match.
    FOR match_rec IN
        SELECT m[1] AS k, m[2] AS v
        FROM regexp_matches(
            root_content,
            '--color-([a-zA-Z0-9_-]+)\s*:\s*([^;]+);',
            'g'
        ) AS m
    LOOP
        color_key := replace(match_rec.k, '-', '_');
        color_val := trim(match_rec.v);

        -- Skip if we somehow ended up with an empty value
        IF length(color_val) > 0 THEN
            result := result || jsonb_build_object(color_key, color_val);
        END IF;
    END LOOP;

    RETURN result;
END;
$func$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION _extract_css_palette(text) IS
    'Phase 3 step 2 helper: parses --color-* CSS variables out of a '
    ':root block into a JSONB palette map. Strips --color- prefix and '
    'converts kebab-case to snake_case. Non-colour variables excluded.';


-- ── 2. Quick sanity check on extraction BEFORE the INSERT ──
--
-- Run the extractor against every theme's css_content and report
-- the key count. If any theme returns 0 keys, something is wrong
-- (empty css_content, malformed :root, unexpected syntax) and we
-- abort before touching palettes.
DO $sanity$
DECLARE
    r          record;
    zero_keys  int := 0;
    total      int := 0;
BEGIN
    FOR r IN
        SELECT
            name,
            (SELECT count(*)
             FROM jsonb_object_keys(_extract_css_palette(
                 COALESCE(NULLIF(css_content, ''), css_template)))
            ) AS key_count
        FROM css_themes
        WHERE name != 'standard-brochure'  -- handled separately in step 3
        ORDER BY name
    LOOP
        total := total + 1;
        RAISE NOTICE 'Extracted palette: % → % keys', r.name, r.key_count;
        IF r.key_count = 0 THEN
            zero_keys := zero_keys + 1;
        END IF;
    END LOOP;

    IF zero_keys > 0 THEN
        RAISE EXCEPTION
            'Aborting: % of % themes produced an empty palette via '
            'extraction. Check css_content formatting before seeding.',
            zero_keys, total;
    END IF;

    RAISE NOTICE 'Sanity OK: % themes extracted with non-empty palettes', total;
END
$sanity$;


-- ── 3. Insert palettes ──
--
-- One palette per theme (13 rows). name mirrors the theme name so
-- the mapping UPDATE in step 3 can use a simple name join.
-- `standard-brochure` excluded — it will share `default` palette.
INSERT INTO palettes (
    name,
    display_name,
    description,
    colours,
    category,
    industry_tags,
    origin,
    is_active
)
SELECT
    t.name,
    t.display_name,
    CASE
        WHEN t.description IS NOT NULL
            THEN 'Palette extracted from theme "' || t.name || '": ' || t.description
        ELSE  'Palette extracted from theme "' || t.name || '" (Phase 3 migration)'
    END AS description,
    _extract_css_palette(
        COALESCE(NULLIF(t.css_content, ''), t.css_template)
    ) AS colours,
    t.category,
    -- No industry_tags on source themes (the column on css_themes is
    -- semantic_tags, which is a slightly different axis). Leave empty;
    -- curators can tag palettes later.
    ARRAY[]::text[] AS industry_tags,
    'seed' AS origin,
    true AS is_active
FROM css_themes t
WHERE t.name != 'standard-brochure'
ON CONFLICT (name) DO UPDATE SET
    display_name  = EXCLUDED.display_name,
    description   = EXCLUDED.description,
    colours       = EXCLUDED.colours,
    category      = EXCLUDED.category,
    is_active     = EXCLUDED.is_active,
    updated_at    = NOW();


-- ── 4. Verification ──
DO $verify$
DECLARE
    expected int := 13;
    actual   int;
    empties  int;
BEGIN
    SELECT COUNT(*) INTO actual FROM palettes WHERE origin = 'seed';

    SELECT COUNT(*) INTO empties
    FROM palettes
    WHERE origin = 'seed' AND colours = '{}'::jsonb;

    IF actual < expected THEN
        RAISE EXCEPTION 'Expected at least % seed palettes, found %',
            expected, actual;
    END IF;

    IF empties > 0 THEN
        RAISE EXCEPTION 'Found % palettes with empty colours — extraction failed',
            empties;
    END IF;

    RAISE NOTICE 'palettes seed complete: % rows with non-empty colours', actual;
END
$verify$;


-- ── 5. Report: one-line summary per palette ──
SELECT
    name,
    (SELECT COUNT(*) FROM jsonb_object_keys(colours)) AS colour_count,
    colours->>'primary'    AS "primary",
    colours->>'background' AS background
FROM palettes
WHERE origin = 'seed'
ORDER BY name;

COMMIT;

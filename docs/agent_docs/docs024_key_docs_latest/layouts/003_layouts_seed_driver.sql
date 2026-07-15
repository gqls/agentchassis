-- =====================================================================
-- MIGRATION 025 — Phase 3 Layouts Seed Driver
-- =====================================================================
-- Loads all 15 layout seeds inside ONE transaction. If any single
-- layout fails (syntax error, constraint violation, etc.) the whole
-- transaction rolls back — you don't end up with a half-loaded
-- layouts table and a confused psql session.
--
-- Each layout SQL is idempotent (ON CONFLICT DO UPDATE), so re-running
-- this driver is safe and will refresh any updated CSS templates.
--
-- Prerequisites:
--   - Phase 2 migration has run (palettes/layouts/typography_sets
--     tables exist; layouts table has all required columns)
--
-- Usage:
--   psql -d <database> -f 003_layouts_seed_driver.sql
--
--   Or from inside psql:
--   \i 003_layouts_seed_driver.sql
-- =====================================================================

\set ON_ERROR_STOP on
\echo 'Starting layouts seed — all 15 layouts in one transaction'

BEGIN;

-- Each layout SQL uses INSERT ... ON CONFLICT (name) DO UPDATE, so
-- re-running this driver is idempotent. No destructive DELETE needed;
-- existing seed rows get refreshed in place, any adopted-origin
-- (forked) rows are left untouched.

\echo '→ 01/15  brochure-formal'
\ir layouts/layout_01_brochure-formal.sql
\echo '→ 02/15  brochure-bold'
\ir layouts/layout_02_brochure-bold.sql
\echo '→ 03/15  portfolio-kinetic'
\ir layouts/layout_03_portfolio-kinetic.sql
\echo '→ 04/15  magazine-grid'
\ir layouts/layout_04_magazine-grid.sql
\echo '→ 05/15  utility-tool'
\ir layouts/layout_05_utility-tool.sql
\echo '→ 06/15  media-grid'
\ir layouts/layout_06_media-grid.sql
\echo '→ 07/15  docs-sidebar'
\ir layouts/layout_07_docs-sidebar.sql
\echo '→ 08/15  soft-editorial'
\ir layouts/layout_08_soft-editorial.sql
\echo '→ 09/15  technical-precise'
\ir layouts/layout_09_technical-precise.sql
\echo '→ 10/15  high-energy'
\ir layouts/layout_10_high-energy.sql
\echo '→ 11/15  comparison-aggregator'
\ir layouts/layout_11_comparison-aggregator.sql
\echo '→ 12/15  affiliate-hub'
\ir layouts/layout_12_affiliate-hub.sql
\echo '→ 13/15  ecommerce-storefront'
\ir layouts/layout_13_ecommerce-storefront.sql
\echo '→ 14/15  tool-first-landing'
\ir layouts/layout_14_tool-first-landing.sql
\echo '→ 15/15  industry-hub'
\ir layouts/layout_15_industry-hub.sql

-- Verification
DO $verify$
DECLARE
    expected_count int := 15;
    actual_count   int;
    missing        text[];
BEGIN
    SELECT COUNT(*), array_agg(name ORDER BY name)
    INTO actual_count, missing
    FROM layouts WHERE origin = 'seed';

    IF actual_count < expected_count THEN
        RAISE EXCEPTION 'Expected % seed layouts, found %: %',
            expected_count, actual_count, missing;
    END IF;

    RAISE NOTICE 'layouts seed complete: % rows loaded', actual_count;
END
$verify$;

COMMIT;

\echo 'Done. All 15 layouts loaded.'

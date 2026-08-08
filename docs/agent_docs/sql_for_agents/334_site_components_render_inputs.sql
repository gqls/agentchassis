-- ============================================================================
-- 334_site_components_render_inputs.sql
--
-- One nullable column: site_components.render_inputs jsonb — the chrome
-- render-provenance stamp (bugs_open/117, lane
-- docs024_key_docs_latest/bugfix_117_chrome_staleness_reference/).
--
-- WHAT IT IS. Site chrome (head/header/footer) is pre-rendered once into
-- site_components.rendered_html and served verbatim by assemblePage; nothing
-- on the page-render path regenerates it, so a change to anything chrome is
-- built from (template, nav, identity, specs, style, brand assets) is inert
-- until a site-chrome rebuild runs. render_inputs records, at render time, a
-- jsonb of NAMED digests over every store the render consumed — written by
-- render_site_components in the SAME UPDATE that stores rendered_html. The
-- stale_site_components discovery check recomputes the SAME shared SQL
-- expression (datahelpers/chrome_render_inputs.go — single source, both
-- callers embed it) and fires `needs_rerender` when the two differ.
--
-- WHY NOT content_data: content_data_envelope_guard.go:115 documents
-- "site_components has NO automated content_data writer at all" as a
-- structural property bug 190's guard relies on. A stamp key there would
-- silently invalidate another mechanism's stated invariant; a dedicated
-- column keeps provenance out of content.
--
-- NULL means "no provenance recorded" — true for every existing row, and the
-- check deliberately treats it as stale (a one-time, bounded baseline drain:
-- 19 sites fire once each on the first post-roll discovery pass, which also
-- catches oufe.com/footer, the measured false negative). Backfilling the
-- current state as a baseline was REJECTED: it would declare known-stale
-- chrome fresh.
--
-- ORDERING: apply this BEFORE rolling the chassis image that writes the
-- column (the render store UPDATE names it; on the old schema that UPDATE
-- errors and chrome stops updating fleet-wide). The column is inert until
-- that image rolls — nothing on the current image reads or writes it.
-- ============================================================================

BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'site_components'
    ) THEN
        RAISE EXCEPTION '334: no site_components table — wrong database?';
    END IF;
END $$;

ALTER TABLE site_components ADD COLUMN IF NOT EXISTS render_inputs jsonb;

COMMENT ON COLUMN site_components.render_inputs IS
    'Render-provenance stamp (bugs_open/117): jsonb of named digests over every store this chrome was rendered from, written by render_site_components in the same UPDATE as rendered_html. Recomputed and compared by the stale_site_components discovery check (shared expression: datahelpers/chrome_render_inputs.go). NULL = no provenance recorded, which the check treats as stale.';

-- Verify by asking the catalog, DO/RAISE so a miss stops the COMMIT
-- (ON_ERROR_STOP ignores a non-empty SELECT result).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'site_components'
          AND column_name = 'render_inputs' AND data_type = 'jsonb'
    ) THEN
        RAISE EXCEPTION '334: site_components.render_inputs jsonb did not land';
    END IF;
    RAISE NOTICE '334: site_components.render_inputs jsonb present';
END $$;

COMMIT;

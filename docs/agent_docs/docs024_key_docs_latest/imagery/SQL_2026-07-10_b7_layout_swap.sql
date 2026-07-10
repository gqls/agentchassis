-- SQL_2026-07-10_b7_layout_swap.sql
--
-- B7 completion, step 2 (after SQL_2026-07-10_b7_layout_fix.sql added the
-- classification tags). The runtime re-compose path is deliberately absent:
-- install_site_composition refuses when style_collection_id exists, and
-- fork_theme_from_site's install mode was removed 2026-04-19. The sanctioned
-- precedent for changing an EXISTING site's composition is the 025 migration
-- pattern — a targeted FK update on the site's own css_themes row.
--
-- robot-hands' theme (theme-robot-hands-com, origin=adopted,
-- source_site_id=self) is site-specific, so swapping its layout_id affects
-- only this site. tool-portal-dark declares no default header/footer
-- components (NULL per 025 convention) and the site's style_collection has
-- none either — no relinking needed. CSS re-renders from the FKs on the next
-- webdesign-agent run; pages re-render via rerender-pages.

\set ON_ERROR_STOP on

-- ── Backup (outside transaction) ──
CREATE TABLE IF NOT EXISTS css_themes_backup_20260710_b7_layout_swap AS
SELECT * FROM css_themes WHERE id = 'b1b60faf-ca68-43f5-a1e6-da3a769e4a25';

SELECT count(*) AS backup_rows FROM css_themes_backup_20260710_b7_layout_swap;

BEGIN;

UPDATE css_themes
SET layout_id = (SELECT id FROM layouts WHERE name = 'tool-portal-dark' AND is_active = true),
    updated_at = now()
WHERE id = 'b1b60faf-ca68-43f5-a1e6-da3a769e4a25'
  AND layout_id = (SELECT id FROM layouts WHERE name = 'brochure-formal');

-- Close the failed re-compose item with the outcome.
UPDATE site_work_items
SET status = 'wont_fix',
    error = COALESCE(error || E'\n', '')
            || 'Closed 2026-07-10: runtime re-compose is unsupported by design; layout switched via targeted css_themes.layout_id migration (025 pattern) to tool-portal-dark. CSS + page re-render triggered separately.',
    updated_at = now()
WHERE item_key = 'needs_composition_b7_fix' AND status = 'failed';

DO $verify$
DECLARE
    v_layout text;
BEGIN
    SELECT l.name INTO v_layout
    FROM css_themes ct JOIN layouts l ON l.id = ct.layout_id
    WHERE ct.id = 'b1b60faf-ca68-43f5-a1e6-da3a769e4a25';
    IF v_layout <> 'tool-portal-dark' THEN
        RAISE EXCEPTION 'layout swap did not land (got %)', v_layout;
    END IF;
    RAISE NOTICE 'robot-hands css_theme now on tool-portal-dark';
END
$verify$;

COMMIT;

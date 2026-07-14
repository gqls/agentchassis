-- 093_vonc_prose_link_404_fixes.sql — vonc.com: fix the two true-404 prose links
-- Created 2026-07-14. Run AFTER 091 (schema source flip) and 092 (wiring), with
-- the chassis image carrying the cta_links_stale recompute live.
--
-- SCOPE — deliberately small. Of vonc's owner-flagged link defects:
--   - the 19 misdirected hero/call-to-action CTAs (-> /contact.html) are NOT
--     hand-fixed here: /contact.html is an excluded destination, so the
--     generic cta_links_stale rerender recompute replaces them with the real
--     interactive targets (primary -> /tools/gauntlet/index.html, secondary
--     -> /tools/quiz/index.html). Dispatch per the runbook, don't hand-edit —
--     "fix the writer, not the row".
--   - the 2 Arena CTAs get the Gauntlet via the same recompute (circular
--     /index.html is replaced); they are retargeted to the Arena in 094 once
--     the Arena page is deployed.
--   - nav (Gauntlet/Quiz missing from site_nav_items) is repaired by the
--     nav_drift -> nav-updater path, now that orphan_pages considers
--     nav-flagged tool pages.
-- What remains for SQL is exactly the two phantom links living in PROSE
-- components (content-block-about, platform-comparison) — outside
-- ctaFieldNames, so the recompute leaves them alone by design:
--   /about.html content-block-about.cta_url:   /how-it-works           -> /archetypes.html
--   /about.html platform-comparison.cta_url:   /how-it-works/the-gauntlet -> /tools/gauntlet/index.html
--
-- The WHEREs match on the phantom VALUE (any vonc component carrying it),
-- not on a hardcoded component id — if the same phantom leaked elsewhere on
-- the site it is fixed too; if a value is already fixed the UPDATE is a no-op.
--
-- LANDMINE: provocation-card / lobby-grid blank content_data is deliberate
-- (runtime-fill shells) — nothing here touches them (cta_url equality only).
--
-- Reversal: _vonc_093_backup_20260714_content.

BEGIN;

CREATE TABLE _vonc_093_backup_20260714_content AS
  SELECT pc.id, pc.page_id, pc.slot_name, pc.content_data
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
    AND pc.content_data->>'cta_url' IN ('/how-it-works', '/how-it-works/the-gauntlet');

UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{cta_url}', '"/archetypes.html"'),
    updated_at = NOW()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND pc.content_data->>'cta_url' = '/how-it-works';

UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{cta_url}', '"/tools/gauntlet/index.html"'),
    updated_at = NOW()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND pc.content_data->>'cta_url' = '/how-it-works/the-gauntlet';

-- ── Verify: the two phantom values are gone from vonc content_data ─────────
DO $$
DECLARE remaining INT; backed_up INT;
BEGIN
  SELECT COUNT(*) INTO backed_up FROM _vonc_093_backup_20260714_content;
  IF backed_up = 0 THEN
    RAISE NOTICE '093: nothing matched — links already fixed (idempotent re-run?)';
  END IF;

  SELECT COUNT(*) INTO remaining
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
    AND pc.content_data->>'cta_url' IN ('/how-it-works', '/how-it-works/the-gauntlet');
  IF remaining <> 0 THEN
    RAISE EXCEPTION 'verify failed: % vonc components still carry a /how-it-works phantom', remaining;
  END IF;
  RAISE NOTICE 'verified: 0 /how-it-works phantoms remain (% rows retargeted)', backed_up;
END $$;

COMMIT;

-- NOTE: /about.html serves stored rendered_html until re-rendered. The runbook
-- dispatches a cta_links_stale rerender for the affected pages (which also
-- re-renders these sections from the updated content_data); alternatively the
-- misdirected_cta discovery items cover the same pages.

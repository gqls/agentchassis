-- ROLLBACK for 570 — restores both testimonial components and their instances.
--
-- Restores is_active and each instance's build_status from the backup 570 took. The
-- rendered_html and content_data were never modified by 570 (it tombstones, it does not
-- delete), so nothing needs restoring there — which is the point of tombstoning rather
-- than deleting.
--
-- ⚠ ROLLING BACK DOES NOT PUT THE BLOCKQUOTES BACK ON THE LIVE PAGES BY ITSELF. If the
-- two pages were rerendered after 570 applied, their deployed HTML no longer contains
-- the sections; restoring the rows makes them assemble again only on the NEXT rerender.
-- Rollback the database, then rerender, in that order — the same two-part shape as the
-- forward migration.
--
-- ⚠ AND NOTE WHAT YOU ARE RESTORING. The owner retired these because no real
-- testimonials exist: every blockquote served was a first-person company statement with
-- an empty author and company, inside a testimonials grid. Roll back only to undo a
-- mistake in 570 itself, not to put that content back without a decision.

BEGIN;

DO $r$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM bak_570_testimonials_20260823;
  IF n < 3 THEN
    RAISE EXCEPTION '570 ROLLBACK ABORT: backup holds % row(s), expected at least 3.', n;
  END IF;
END
$r$;

-- Components first, so the instances below rejoin an active parent.
UPDATE content_components c
SET is_active = b.is_active
FROM (SELECT DISTINCT component_id, name, is_active FROM bak_570_testimonials_20260823) b
WHERE c.id = b.component_id;

UPDATE page_components pc
SET build_status = b.build_status
FROM bak_570_testimonials_20260823 b
WHERE pc.id = b.page_component_id
  AND b.page_component_id IS NOT NULL;

DO $v$
DECLARE n_active int; n_restored int;
BEGIN
  SELECT count(*) INTO n_active FROM content_components
  WHERE is_active AND name IN ('testimonials','social_proof');
  IF n_active <> 2 THEN
    RAISE EXCEPTION '570 ROLLBACK VERIFY FAILED: % component(s) active, expected 2.', n_active;
  END IF;

  SELECT count(*) INTO n_restored
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id
  WHERE cc.name IN ('testimonials','social_proof') AND p.status IN ('active','deployed')
    AND COALESCE(pc.build_status,'pending') <> 'removed';
  IF n_restored <> 3 THEN
    RAISE EXCEPTION '570 ROLLBACK VERIFY FAILED: % live un-tombstoned instance(s), expected 3.', n_restored;
  END IF;

  RAISE NOTICE '570 ROLLBACK OK: 2 components reactivated, 3 instances un-tombstoned. A RERENDER is still needed to put them back on the pages.';
END
$v$;

COMMIT;

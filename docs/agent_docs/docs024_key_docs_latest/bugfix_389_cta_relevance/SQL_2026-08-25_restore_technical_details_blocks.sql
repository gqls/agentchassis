-- Restore two authored text blocks on finetuning.uk/technical-details.html -- THE CANARY.
--
-- WHAT HAPPENED: the same defect as on your-own-model.html, and it happened FIRST, on the
-- page this lane used to validate the whole recipe. The canary's content_rewrite
-- b422751a-3745-474c-87d6-aeff50028546 (archived 2026-08-25 13:05:41.827Z) overwrote the
-- generic-text-block components at positions 2 and 3 with near-copies of position 4.
--
-- Pre-rewrite, the three blocks were distinct (all headed "The model and its licence"):
--   pos 2  rendered 1828 B  "The base model ITSELF is a small open-weight model: one where the maker publishes..."
--   pos 3  rendered 1599 B  "The base model is a small open-weight model, meaning the UNDERLYING WEIGHTS are published..."
--   pos 4  rendered 1712 B  "The base model is a small open-weight model, meaning the COMPANY THAT built it has published..."
-- After it, all three carry position 4's text (1710 / 1712 / 1712 B), so position 4 is intact
-- and only 2 and 3 need restoring.
--
-- ⚠ WHY THIS WAS MISSED FOR SEVEN HOURS, and it is the reason the recipe changed:
-- the canary was signed off with `grep -c '<p'` at 15 before and 15 after. It held at 15/15
-- BECAUSE the replacement preserved the paragraph count exactly -- the blocks it copied in
-- have the same shape as the ones it destroyed. The count control did not merely fail to
-- move enough; it was structurally incapable of moving. That page was then used to validate
-- the control for the other eleven. The check that finds it is distinctness, not volume:
--   count(*) vs count(DISTINCT left(text,80))  ->  6 components, 4 distinct.
--
-- Restores content_data VERBATIM by subquery from the archive written by the offending item
-- -- nothing retyped. Neither block carries a CTA url (keys are content,heading; no
-- password-entropy), so this reverts prose only and leaves the canary's CTA repair intact.
-- rendered_html is deliberately NOT written here: a page_rerender regenerates it.

BEGIN;

UPDATE page_components pc
SET content_data = h.content_data, updated_at = now()
FROM (SELECT content_data FROM page_component_history
      WHERE page_id = 'a32b8822-db49-4e45-88f8-bda06d73de62'
        AND source = 'save_page_sections_overwrite'
        AND source_item_id = 'b422751a-3745-474c-87d6-aeff50028546'
        AND content_data->>'content' LIKE '<p>The base model itself is a small open-weight model%') h
WHERE pc.page_id = 'a32b8822-db49-4e45-88f8-bda06d73de62'
  AND pc.slot_name = 'generic-text-block'
  AND pc.position = 2;

UPDATE page_components pc
SET content_data = h.content_data, updated_at = now()
FROM (SELECT content_data FROM page_component_history
      WHERE page_id = 'a32b8822-db49-4e45-88f8-bda06d73de62'
        AND source = 'save_page_sections_overwrite'
        AND source_item_id = 'b422751a-3745-474c-87d6-aeff50028546'
        AND content_data->>'content' LIKE '<p>The base model is a small open-weight model, meaning the underlying weights%') h
WHERE pc.page_id = 'a32b8822-db49-4e45-88f8-bda06d73de62'
  AND pc.slot_name = 'generic-text-block'
  AND pc.position = 3;

DO $$
DECLARE n_blocks int; n_distinct int; n_null int;
BEGIN
  SELECT count(*), count(DISTINCT left(content_data->>'content', 80))
    INTO n_blocks, n_distinct
    FROM page_components
   WHERE page_id = 'a32b8822-db49-4e45-88f8-bda06d73de62'
     AND slot_name = 'generic-text-block';

  IF n_blocks <> 3 OR n_distinct <> 3 THEN
    RAISE EXCEPTION 'restore failed: % generic-text-blocks, % distinct openings (both must be 3)',
      n_blocks, n_distinct;
  END IF;

  SELECT count(*) INTO n_null FROM page_components
   WHERE page_id = 'a32b8822-db49-4e45-88f8-bda06d73de62' AND content_data IS NULL;
  IF n_null > 0 THEN
    RAISE EXCEPTION 'refusing: % section(s) have content_data IS NULL; a rerender would regenerate the page', n_null;
  END IF;

  RAISE NOTICE 'restore OK: 3 generic-text-blocks, 3 distinct openings, 0 NULL content_data';
END $$;

COMMIT;

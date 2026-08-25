-- Restore two authored text blocks on finetuning.uk/your-own-model.html.
--
-- WHAT HAPPENED: this lane's own content_rewrite work item
-- 10b8b6d2-660c-4696-ae6a-ca20c8823dcf (2026-08-25 19:43:54Z) was commissioned to
-- reword CTA LABELS only. It also regenerated the page body and overwrote the
-- generic-text-block components at positions 2 and 3 with near-copies of the block
-- at position 4, so the page served the same "How it works" text THREE times and two
-- distinct authored sections were lost. Confirmed at the served bytes 2026-08-25 20:2xZ
-- (3 x "You send us examples of how your business writes and answers", 0 x each
-- destroyed opening) and in page_component_history.
--
-- The pre-rewrite content_data is recoverable because the same item archived it. This
-- restores it VERBATIM from that archive -- nothing is retyped. Neither archived block
-- contains a CTA url (has_pe = false, keys = content,heading), so this reverts prose
-- only and leaves the CTA repair in hero/call-to-action untouched.
--
-- rendered_html is deliberately NOT written here: a page_rerender regenerates it from
-- content_data, which keeps the writer set for that column unchanged.

BEGIN;

UPDATE page_components pc
SET content_data = h.content_data, updated_at = now()
FROM (SELECT content_data FROM page_component_history
      WHERE page_id = 'a8909fc1-f1ff-43fe-842c-5ce364b8b182'
        AND source = 'save_page_sections_overwrite'
        AND source_item_id = '10b8b6d2-660c-4696-ae6a-ca20c8823dcf'
        AND content_data->>'content' LIKE '<p>Training a model on your own documents%') h
WHERE pc.page_id = 'a8909fc1-f1ff-43fe-842c-5ce364b8b182'
  AND pc.slot_name = 'generic-text-block'
  AND pc.position = 2;

UPDATE page_components pc
SET content_data = h.content_data, updated_at = now()
FROM (SELECT content_data FROM page_component_history
      WHERE page_id = 'a8909fc1-f1ff-43fe-842c-5ce364b8b182'
        AND source = 'save_page_sections_overwrite'
        AND source_item_id = '10b8b6d2-660c-4696-ae6a-ca20c8823dcf'
        AND content_data->>'content' LIKE '<p>The process runs in three steps%') h
WHERE pc.page_id = 'a8909fc1-f1ff-43fe-842c-5ce364b8b182'
  AND pc.slot_name = 'generic-text-block'
  AND pc.position = 3;

-- Verification that can actually ABORT the commit (a bare SELECT cannot).
DO $$
DECLARE
  n_distinct int;
  n_blocks   int;
  n_null     int;
BEGIN
  SELECT count(*), count(DISTINCT left(content_data->>'content', 80))
    INTO n_blocks, n_distinct
    FROM page_components
   WHERE page_id = 'a8909fc1-f1ff-43fe-842c-5ce364b8b182'
     AND slot_name = 'generic-text-block';

  IF n_blocks <> 3 OR n_distinct <> 3 THEN
    RAISE EXCEPTION 'restore failed: % generic-text-blocks, % distinct openings (both must be 3)',
      n_blocks, n_distinct;
  END IF;

  -- A NULL content_data on ANY section escalates the next rerender to the content
  -- writer, which regenerates the copy -- i.e. it would undo this restore.
  SELECT count(*) INTO n_null
    FROM page_components
   WHERE page_id = 'a8909fc1-f1ff-43fe-842c-5ce364b8b182'
     AND content_data IS NULL;

  IF n_null > 0 THEN
    RAISE EXCEPTION 'refusing: % section(s) have content_data IS NULL; a rerender would regenerate the page', n_null;
  END IF;

  RAISE NOTICE 'restore OK: 3 generic-text-blocks, 3 distinct openings, 0 NULL content_data';
END $$;

COMMIT;

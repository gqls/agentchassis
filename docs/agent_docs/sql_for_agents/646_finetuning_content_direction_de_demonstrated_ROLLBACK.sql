-- ROLLBACK for 646: restore the previous finetuning.uk content_direction, i.e.
-- restore the 7 'rather than' demonstrations and the 3 others. Only reach for this
-- if the de-demonstrated brief measurably harms output; it does not change any
-- instruction's force, only the form the brief teaches by example.
BEGIN;
UPDATE site_specs ss SET is_current = false, superseded_at = now()
  FROM sites s WHERE s.id = ss.site_id AND s.domain = 'finetuning.uk'
   AND ss.aspect = 'content_direction' AND ss.is_current;
UPDATE site_specs ss SET is_current = true, superseded_at = NULL
  FROM sites s WHERE s.id = ss.site_id AND s.domain = 'finetuning.uk'
   AND ss.aspect = 'content_direction'
   AND ss.id = (SELECT ss2.id FROM site_specs ss2 JOIN sites s2 ON s2.id = ss2.site_id
                 WHERE s2.domain='finetuning.uk' AND ss2.aspect='content_direction'
                   AND ss2.superseded_at IS NOT NULL AND ss2.created_by <> 'claude-finetuning-uk-lane'
                 ORDER BY ss2.superseded_at DESC LIMIT 1);
COMMIT;

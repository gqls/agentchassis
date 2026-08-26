-- ROLLBACK for 647: restore the previous finetuning.uk content_direction (the
-- post-646 row, i.e. rather-than cleared but ", not just" restored).
BEGIN;
UPDATE site_specs ss SET is_current = false, superseded_at = now()
  FROM sites s WHERE s.id = ss.site_id AND s.domain='finetuning.uk'
   AND ss.aspect='content_direction' AND ss.is_current;
UPDATE site_specs ss SET is_current = true, superseded_at = NULL
  FROM sites s WHERE s.id = ss.site_id AND s.domain='finetuning.uk' AND ss.aspect='content_direction'
   AND ss.id = (SELECT ss2.id FROM site_specs ss2 JOIN sites s2 ON s2.id=ss2.site_id
                 WHERE s2.domain='finetuning.uk' AND ss2.aspect='content_direction'
                   AND ss2.superseded_at IS NOT NULL ORDER BY ss2.superseded_at DESC LIMIT 1);
COMMIT;

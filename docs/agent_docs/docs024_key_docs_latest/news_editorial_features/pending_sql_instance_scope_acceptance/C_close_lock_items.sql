-- news_editorial_features lane — 2026-08-25
-- Close the two lock_blocked_change items once the change is live and re-locked.
-- 'complete' (not 'wont_fix'): the blocked change was APPLIED, not declined.
-- Run LAST, after B_relock.sql and after the served pages verify.

BEGIN;

UPDATE site_work_items
   SET status = 'complete',
       completed_at = now(),
       handled_by = 'news_editorial_features-lane',
       result = result || jsonb_build_object(
         'disposition', 'accepted',
         'decided_by', 'owner ruling 2026-08-25',
         'note', 'Lock lifted deliberately; the 283/RFC_034 instance-scope conversion was delivered to this instance and the row re-locked. Dry-run diff before the write showed the id attribute as the only delta (evidence-timeseries-ifr / -pdc-calendar -> c-evidence-timeseries).')
 WHERE id IN ('6f48b825-2be7-46f9-8dda-3e63b0bd2469',
              '5e42ea75-dd26-48dc-80da-1b05ecc84522')
   AND status = 'needs_human_review'
RETURNING id, item_key, status, completed_at;

COMMIT;

SELECT id, item_key, status, completed_at, result->>'disposition' AS disposition
  FROM site_work_items
 WHERE id IN ('6f48b825-2be7-46f9-8dda-3e63b0bd2469',
              '5e42ea75-dd26-48dc-80da-1b05ecc84522');

-- 668 RELOCK — close the window 668_terms_publish_unlock.sql opened.
-- Run this the moment the terms render completes. Safe to run repeatedly.
-- Restores the ORIGINAL 2026-07-21 provenance, not now(): that timestamp is the
-- lock's history and any lock-age review reads it.
BEGIN;
UPDATE page_components pc
   SET locked_at = TIMESTAMPTZ '2026-07-21 09:21:45.96136+00',
       locked_by = '182_legal_pages', lock_type = 'permanent', lock_expires_at = NULL
  FROM pages p, sites s
 WHERE p.id = pc.page_id AND s.id = p.site_id AND s.domain='finetuning.uk' AND p.name='terms';

UPDATE pages p SET rebuild_policy = 'owned', updated_at = now()
  FROM sites s WHERE s.id = p.site_id AND s.domain='finetuning.uk' AND p.name='terms';

DO $$
DECLARE writable bool; pol text; lk timestamptz; lb text; lt text;
BEGIN
    SELECT (pc.locked_at IS NULL OR (pc.lock_type='timed' AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at < NOW())),
           COALESCE(p.rebuild_policy,'generic'), pc.locked_at, pc.locked_by, pc.lock_type
      INTO writable, pol, lk, lb, lt
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
     WHERE s.domain='finetuning.uk' AND p.name='terms';
    IF writable THEN RAISE EXCEPTION 'RELOCK: terms is STILL agent-writable'; END IF;
    IF pol <> 'owned' THEN RAISE EXCEPTION 'RELOCK: rebuild_policy is %, want owned', pol; END IF;
    IF lk <> TIMESTAMPTZ '2026-07-21 09:21:45.96136+00' OR lb <> '182_legal_pages' OR lt <> 'permanent' THEN
        RAISE EXCEPTION 'RELOCK: provenance not restored (at=%, by=%, type=%)', lk, lb, lt;
    END IF;
    RAISE NOTICE 'RELOCK OK: terms is owned + permanently locked with its original 2026-07-21 provenance';
END $$;
COMMIT;

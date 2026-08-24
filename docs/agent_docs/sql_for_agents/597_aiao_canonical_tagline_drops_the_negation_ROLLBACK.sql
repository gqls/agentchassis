-- 597 ROLLBACK — restore the define-by-negation tagline.
--
-- ⚠ THIS RE-MANDATES THE SENTENCE THE OWNER OBJECTED TO, and re-arms the gate's
-- exemption of it, so the pages will keep serving it and the gate will keep
-- reporting it as `exempt` rather than repairable. Reverting is a CONTENT
-- decision, not a technical one — it needs the owner, not a session.
--
-- Restores by superseding the corrected rows and re-activating the originals,
-- rather than by editing text back, so the pre-597 bytes are exactly what returns.

BEGIN;

UPDATE site_specs ss SET is_current = false, superseded_at = now()
  FROM sites s
 WHERE s.id = ss.site_id AND s.domain = 'ai-agent-orchestration.com'
   AND ss.is_current AND ss.notes LIKE '%Migration 597.%';

UPDATE site_specs ss SET is_current = true, superseded_at = NULL
  FROM sites s
 WHERE s.id = ss.site_id AND s.domain = 'ai-agent-orchestration.com'
   AND NOT ss.is_current
   AND ss.data::text LIKE '%in days, not months%'
   AND ss.aspect IN ('identity','content_direction','site_plan');

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='ai-agent-orchestration.com' AND ss.is_current
     AND ss.data::text LIKE '%in days, not months%';
  IF n <> 3 THEN
    RAISE EXCEPTION '597 ROLLBACK FAILED: expected 3 current rows carrying the original phrase, got %', n;
  END IF;
  RAISE NOTICE '597 ROLLBACK OK: the negation tagline is mandated again';
END $$;

COMMIT;

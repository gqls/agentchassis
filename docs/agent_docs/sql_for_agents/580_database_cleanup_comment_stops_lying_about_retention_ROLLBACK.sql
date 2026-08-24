-- 580 ROLLBACK — restore the stale arm-1 comment.
--
-- ⚠ This puts back a comment that is FALSE. It describes 14/30-day retention that migration 567
-- removed, in a live config row where an operator has nothing else to check it against. Roll this
-- back only if 580 itself is somehow harmful — which is hard to arrange for a comment — and never
-- as tidy-up.

BEGIN;

DO $do$
DECLARE q text; n int;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  n := (length(q) - length(replace(q, 'Retention is BY FINDING CODE since migration 567', '')))
       / length('Retention is BY FINDING CODE since migration 567');
  IF n <> 1 THEN
    RAISE EXCEPTION '580 ROLLBACK REFUSED: 580''s comment is not present exactly once (found %)', n;
  END IF;
END
$do$;

UPDATE scheduled_tasks
   SET pre_query = regexp_replace(pre_query,
         '-- 1\. Clean agent_error_log\. Retention is BY FINDING CODE since migration 567:.*?--    --finding-codes; see bugs_open/358\.',
         '-- 1. Clean agent_error_log (resolved errors > 14 days, unresolved > 30 days)',
         ''),   -- NO 'n' FLAG. In PostgreSQL 'n' means newline-SENSITIVE matching, which
                -- stops `.` matching a newline — and this pattern spans lines. With 'n' the
                -- replace silently no-ops, the UPDATE reports "UPDATE 1" having changed
                -- nothing, and only the verify block below catches it. Measured 2026-08-24
                -- against the live row: with 'n' NO MATCH, without it MATCHED.
       updated_at = now()
 WHERE name = 'database-cleanup';

DO $do$
DECLARE q text;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  IF q LIKE '%Retention is BY FINDING CODE%' THEN
    RAISE EXCEPTION '580 ROLLBACK: 580''s comment was not fully removed';
  END IF;
  -- the RULE must still be 567's, exactly as the forward migration guaranteed
  IF q NOT LIKE '%INTERVAL ''365 days''%' OR q LIKE '%resolved = true AND occurred_at%' THEN
    RAISE EXCEPTION '580 ROLLBACK: arm 1''s retention rule changed — a comment rollback must not touch behaviour';
  END IF;
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '580 ROLLBACK: pre_query does NOT execute (%)', SQLERRM;
    END IF;
  END;
  RAISE NOTICE '580 ROLLBACK: applied — the stale (and false) comment is back.';
END
$do$;

COMMIT;

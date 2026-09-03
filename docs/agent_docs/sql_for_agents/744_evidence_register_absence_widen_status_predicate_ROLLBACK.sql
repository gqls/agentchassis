-- 744_evidence_register_absence_widen_status_predicate_ROLLBACK.sql
--
-- Narrows `evidence-register-absence` back to `s.status = 'deployed'`.
--
-- ⚠ This re-opens the blind spot 744 closed: an `active` site with no register becomes
-- invisible to the one check built to find invisible sites. Only run it if `active` turns out
-- to mean something other than "live" on this estate — and if so, fix the discovery rotations
-- too, because they all scope on ('active','deployed') and would be wrong in the same way.

BEGIN;

DO $$
DECLARE pq text;
BEGIN
  SELECT pre_query INTO pq FROM scheduled_tasks WHERE name = 'evidence-register-absence';
  IF pq IS NULL THEN
    RAISE EXCEPTION '744 ROLLBACK ABORT: task not found';
  END IF;
  IF pq NOT LIKE '%s.status IN (''active'', ''deployed'')%' THEN
    RAISE EXCEPTION '744 ROLLBACK ABORT: the widened predicate is not present - nothing to revert';
  END IF;
END $$;

UPDATE scheduled_tasks
   SET pre_query = replace(pre_query, 's.status IN (''active'', ''deployed'')', 's.status = ''deployed'''),
       updated_at = now()
 WHERE name = 'evidence-register-absence';

DO $$
DECLARE pq text;
BEGIN
  SELECT pre_query INTO pq FROM scheduled_tasks WHERE name = 'evidence-register-absence';
  IF pq NOT LIKE '%s.status = ''deployed''%' THEN
    RAISE EXCEPTION '744 ROLLBACK VERIFY: narrow predicate not restored';
  END IF;
  RAISE NOTICE '744 ROLLBACK OK: narrowed to deployed only - the active-site blind spot is re-opened, deliberately';
END $$;

COMMIT;

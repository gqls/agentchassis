-- ============================================================================
-- 342_thunder_orphan_scan_ROLLBACK.sql — sidecar, never applied by the runner
-- (SIDECAR_RE excludes UPPERCASE-suffixed files; apply by hand if needed).
--
-- Reverts the three rows 342 seeds. Deliberately does NOT touch:
--   - site_work_items rows (item_type='thunder_orphan') — findings are
--     evidence about the vendor account, not part of the pipeline's wiring;
--     if any exist at rollback time, a human should read them first.
--   - schema_migrations — un-recording is a hand decision (see the runner's
--     record-only notes; a ledger row for a reverted file should have its
--     note amended, not vanish).
-- ============================================================================

ROLLBACK; -- defensive, same as the forward file

BEGIN;

DELETE FROM scheduled_tasks WHERE name = 'thunder-orphan-scan';

-- Soft-disable the definition rather than DELETE if you prefer history:
--   UPDATE agent_definitions SET is_active=false, updated_at=NOW()
--   WHERE type='thunder-orphan-scan';
DELETE FROM agent_definitions WHERE type = 'thunder-orphan-scan';

DELETE FROM doc_notes
WHERE subject_type = 'pipeline' AND subject_key = 'thunder-orphan-scan'
  AND created_by = 'finetuning_uk_service lane';

DO $verify$
BEGIN
    PERFORM 1 FROM scheduled_tasks WHERE name = 'thunder-orphan-scan';
    IF FOUND THEN RAISE EXCEPTION 'scheduled_tasks row survived rollback'; END IF;
    PERFORM 1 FROM agent_definitions WHERE type = 'thunder-orphan-scan';
    IF FOUND THEN RAISE EXCEPTION 'agent_definitions row survived rollback'; END IF;
END
$verify$;

COMMIT;

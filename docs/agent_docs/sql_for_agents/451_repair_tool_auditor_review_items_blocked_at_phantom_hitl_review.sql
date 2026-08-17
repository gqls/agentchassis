-- 451_repair_tool_auditor_review_items_blocked_at_phantom_hitl_review.sql
--
-- bugs_open/291, the repair half (450 is the config half — apply 450 FIRST so the
-- population stops growing). Re-parks every tool-auditor review item that claim
-- flipped to blocked-at-a-phantom-handler back into the HITL parking idiom:
-- status='needs_human_review', handler_agent='' (canonical per migration 217),
-- error cleared. The rows are REAL audit findings ("no labels on the four number
-- inputs", "bare import of policy-generator.js", ...) — the repair is a re-shape,
-- not a cancel, and their (site_id, item_key) dedup slots are deliberately KEPT:
-- the parked row IS the finding, and re-filing a duplicate while a human has not
-- looked would be the dedup index's job to prevent anyway. Their designed exit is
-- the admin confirm endpoint (confirm_work_item_handler.go: spec.check='tool_auditor'
-- → improve_tool follow-up).
--
-- Gates on SHAPE + HOMOGENEITY, not count: rows were still arriving daily until
-- 450 applied, so a pinned count would go stale between writing and applying
-- (measured 5 → 14 across 08-16 → 08-17). attempt_count/claimed_* are asserted
-- zero/NULL rather than cleared — claim's not-registered branch nulls them itself.
--
-- Every repaired row is stamped result.repair_291 so none looks spontaneously
-- fixed and the ROLLBACK can key on the stamp (the 442 mechanism).
--
-- ⚠ An auditor orchestration loaded with the OLD config before 450 can still file
-- blocked rows for as long as it lives. The straggler sweep (same UPDATE,
-- re-runnable) is in
-- docs024_key_docs_latest/bugfix_291_hitl_review_phantom_handler/RUNBOOK_*.md —
-- re-census the next day.
--
-- Data-only: no image dependency, LIVE ON APPLY.
-- Rollback: 451_repair_tool_auditor_review_items_blocked_at_phantom_hitl_review_ROLLBACK.sql

BEGIN;

DO $$
DECLARE
  n int;
  bad int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE status = 'blocked'
     AND error = 'Handler agent not registered: hitl-review';
  IF n = 0 THEN
    RAISE EXCEPTION '451: nothing to repair — already applied, or repaired by another hand; check result ? ''repair_291''';
  END IF;

  -- Homogeneity: every row matching the error must match the measured shape.
  -- If any does not, a DIFFERENT producer has started naming this phantom —
  -- re-derive before widening the repair.
  SELECT count(*) INTO bad FROM site_work_items
   WHERE status = 'blocked'
     AND error = 'Handler agent not registered: hitl-review'
     AND NOT (created_by = 'tool-auditor'
              AND item_type = 'needs_human_review'
              AND attempt_count = 0
              AND claimed_at IS NULL);
  IF bad > 0 THEN
    RAISE EXCEPTION '451: % row(s) carry the error but not the measured shape (created_by=tool-auditor, item_type=needs_human_review, attempt_count=0, claimed_at NULL) — re-derive, do not widen', bad;
  END IF;
END $$;

UPDATE site_work_items
   SET status        = 'needs_human_review',
       handler_agent = '',
       error         = NULL,
       result        = COALESCE(result, '{}'::jsonb) || jsonb_build_object('repair_291',
                         jsonb_build_object(
                           'repaired_at',  now()::text,
                           'from_status',  'blocked',
                           'from_handler', 'hitl-review',
                           'from_error',   'Handler agent not registered: hitl-review',
                           'why',          'born dispatchable at a handler that has never existed; re-parked at the HITL idiom — bugs_open/291')),
       updated_at    = now()
 WHERE status = 'blocked'
   AND error = 'Handler agent not registered: hitl-review'
   AND created_by = 'tool-auditor';

-- Post-verify: predicate empty, every stamped row at the parked shape.
DO $$
DECLARE
  leftover int;
  malformed int;
  repaired int;
BEGIN
  SELECT count(*) INTO leftover FROM site_work_items
   WHERE status = 'blocked'
     AND error = 'Handler agent not registered: hitl-review';
  IF leftover > 0 THEN
    RAISE EXCEPTION '451: % row(s) still at the repair predicate after UPDATE', leftover;
  END IF;

  SELECT count(*) INTO repaired FROM site_work_items WHERE result ? 'repair_291';
  IF repaired = 0 THEN
    RAISE EXCEPTION '451: no rows carry the repair_291 stamp — the UPDATE matched nothing it should have';
  END IF;

  SELECT count(*) INTO malformed FROM site_work_items
   WHERE result ? 'repair_291'
     AND NOT (status = 'needs_human_review' AND handler_agent = '' AND error IS NULL);
  IF malformed > 0 THEN
    RAISE EXCEPTION '451: % stamped row(s) not at the parked shape', malformed;
  END IF;

  RAISE NOTICE '451: % row(s) re-parked at needs_human_review', repaired;
END $$;

COMMIT;

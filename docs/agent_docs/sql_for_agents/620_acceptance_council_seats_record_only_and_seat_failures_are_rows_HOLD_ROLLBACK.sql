-- 620_..._HOLD_ROLLBACK.sql - undo Phase 3: restore improvement-loop from 620's own
-- pre-update snapshot, strip filing_mode from the five model seats, soft-delete the
-- two agents 620 created. NOTE what this reinstates: every model finding dispatches a
-- page rewrite again, and seat failures leave no row.
DO $probe$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type='reader-experience-auditor' AND deleted_at IS NULL) THEN
        RAISE EXCEPTION '620 ROLLBACK: not applied';
    END IF;
END $probe$;
BEGIN;
SELECT snapshot_agent('improvement-loop', '620_..._ROLLBACK: pre-restore');
DO $u$
DECLARE v_snap jsonb; r RECORD;
BEGIN
    -- snapshot_agent() writes to agent_definitions_backup with snapshot_reason (read from
    -- the function body 2026-08-25), not to agent_definitions.is_snapshot.
    SELECT default_config INTO v_snap FROM agent_definitions_backup
     WHERE type='improvement-loop'
       AND snapshot_reason LIKE '620_acceptance_council_seats_record_only_and_seat_failures_are_rows_HOLD: pre-update%'
     ORDER BY snapshot_taken_at DESC LIMIT 1;
    IF v_snap IS NULL THEN RAISE EXCEPTION '620 ROLLBACK: the pre-update snapshot of improvement-loop was not found - restore by hand from the snapshot_agent row'; END IF;
    UPDATE agent_definitions SET default_config = v_snap, updated_at = now()
     WHERE type='improvement-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    FOR r IN SELECT * FROM (VALUES
        ('visual-design-auditor','write_findings'),('content-quality-auditor','write_findings'),
        ('site-review-agent','write_strategic_findings'),('offer-analyser','write_offer_findings'),
        ('brief-fidelity-auditor','write_findings')) AS t(agent, step)
    LOOP
        UPDATE agent_definitions
           SET default_config = default_config #- ARRAY['workflow','steps', r.step, 'config', 'filing_mode'], updated_at = now()
         WHERE type = r.agent AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    END LOOP;
    UPDATE agent_definitions SET is_active = false, deleted_at = now(), updated_at = now()
     WHERE type IN ('reader-experience-auditor','acceptance-discovery-agent') AND deleted_at IS NULL;
    IF (SELECT count(*) FROM jsonb_object_keys((SELECT default_config #> '{workflow,steps}' FROM agent_definitions WHERE type='improvement-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL))) <> 31 THEN
        RAISE EXCEPTION '620 ROLLBACK verify: improvement-loop does not have 31 steps after restore';
    END IF;
    RAISE NOTICE '620 ROLLBACK: done';
END $u$;
COMMIT;

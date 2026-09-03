-- ROLLBACK for 736. Restores migration 466's original seed text for RUNNING.
-- ⚠ Restoring it makes the row name a function that does not exist in any image
-- built from b55f837ef or later. Only run this alongside a revert of that commit.
BEGIN;
UPDATE orchestration_status_vocabulary
   SET written_by = 'StateRepository.ClearExecutingStep (state.go:1428)',
       notes      = 'The inter-step gap during a stuck-orchestration takeover. Milliseconds by construction; bugs_closed/294.',
       updated_at = now()
 WHERE status = 'RUNNING';
COMMIT;

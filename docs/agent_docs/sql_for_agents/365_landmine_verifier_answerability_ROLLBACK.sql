-- 365_landmine_verifier_answerability_ROLLBACK.sql
--
-- Undoes 365: the verifier goes back to run_checks --> verify, the no-evidence
-- branch and the evidence suffix are removed, and the amended verify prompt is
-- restored from the snapshot row 365 took before it wrote.
--
-- WHY IT RESTORES FROM THE SNAPSHOT rather than trimming the appended text: the
-- EVIDENCE RULES block was appended to whatever the live prompt said at the time,
-- and another lane may have amended that prompt since. String-trimming would then
-- silently discard their work. The snapshot is the only copy that is certainly
-- the pre-365 text, and if it is missing this script REFUSES rather than guessing.

BEGIN;

DO $$
DECLARE
  snap jsonb;
  live_ver int;
BEGIN
  SELECT default_config INTO snap FROM agent_definitions
   WHERE type = 'landmine-verifier' AND COALESCE(is_snapshot, false) = true
     AND name LIKE '%pre-365 snapshot%' AND deleted_at IS NULL
   ORDER BY created_at DESC LIMIT 1;

  IF snap IS NULL THEN
    RAISE EXCEPTION 'no pre-365 snapshot row found — refusing to guess the previous prompt text. Recover from git (docs/agent_docs/sql_for_agents/276_landmine_verifier.sql is history, NOT the live row) or hand-write the revert.';
  END IF;

  SELECT version INTO live_ver FROM agent_definitions
   WHERE type = 'landmine-verifier' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  UPDATE agent_definitions
     SET default_config = snap
   WHERE type = 'landmine-verifier' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  -- Assert the revert, in the same transaction, with a RAISE: the four things 365
  -- added must all be gone, and run_checks must again hand off to verify.
  PERFORM 1;
  IF EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE type = 'landmine-verifier' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND (default_config #> '{workflow,steps,gate_evidence}' IS NOT NULL
         OR default_config #> '{workflow,steps,verify_unverifiable}' IS NOT NULL
         OR default_config #>> '{workflow,steps,persist_verdict,config,note_body_suffix_field}' IS NOT NULL
         OR default_config #>> '{workflow,steps,run_checks,next_step}' <> 'verify')
  ) THEN
    RAISE EXCEPTION 'revert incomplete — the snapshot did not restore the pre-365 topology';
  END IF;

  RAISE NOTICE 'bugs_open/223 gate reverted from the pre-365 snapshot (live version %)', live_ver;
END $$;

COMMIT;

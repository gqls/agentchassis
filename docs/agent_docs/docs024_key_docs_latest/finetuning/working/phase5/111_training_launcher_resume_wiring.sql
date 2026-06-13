-- Migration 111 — training-launcher: wire Phase D resume into the launch path.
--
-- WHY: the adapter already implements prepare_resume_url (lists a run's
-- checkpoints in B2, presigns a GET for the latest, replies {found, presigned_url,
-- key, index}); migration 110 + the new dispatch_thunder_prepare_resume_url close
-- the chassis gap. This wires ONE check_resume step into the launcher so a launch
-- auto-resumes from the latest checkpoint when one exists.
--
-- KEY DESIGN: found=false (no checkpoints) is a VALID answer — a fresh start.
-- assemble_upload_manifest only emits a manifest "resume" block when resume_url is
-- non-empty, so wiring check_resume in UNCONDITIONALLY is safe: a first launch
-- finds nothing (fresh), a re-run of the SAME training_run_id (on a fresh box)
-- finds the prior box's checkpoints and resumes. One workflow serves both.
--
-- Chain change: presign_final -> check_resume -> assemble_manifest (was
-- presign_final -> assemble_manifest). Verified post-110 shape:
--   presign_final.next_step = "assemble_manifest"
--   assemble_manifest.config = {final_key, final_url, checkpoint_keys,
--     checkpoint_urls, training_run_id}, next_step write_manifest.
--
-- training_run_id is NOT set in check_resume config: configOrInput reads a config
-- value as a LITERAL (it does not resolve dot-paths), so a config entry
-- "training_run_id":"input_data.training_run_id" would send that literal string.
-- Instead we omit it — configOrInput falls back to input_data.training_run_id and
-- resolves the real value. (expiry_minutes 3000 benefits from the configOrInput
-- numeric-coercion fix; if that chassis isn't deployed yet it degrades to the
-- adapter's 1h GET default, which is fine — the resume URL is used at launch,
-- within a minute.)
--
-- DEPLOY ORDERING: apply only AFTER the chassis carrying
-- DispatchThunderPrepareResumeURLAction + its registry entry is built+deployed and
-- the def image_tag bumped — else the launcher fails check_resume with an
-- unknown-action error. (The adapter side, prepare_resume_url, is already live.)

BEGIN;

-- 0. Guard: exactly one live training-launcher def.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION 'migration 111: expected exactly 1 live training-launcher def, found %', n;
  END IF;
END $$;

-- 0a. Snapshot before mutating (sanctioned helper; revert with revert_agent).
SELECT snapshot_agent('training-launcher', 'pre-migration-111 resume wiring');

-- 1. presign_final now flows into check_resume (was assemble_manifest).
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config, '{workflow,steps,presign_final,next_step}',
    '"check_resume"'::jsonb, true)
WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

-- 2. Add the check_resume step (awaited resume probe), flowing into assemble.
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config, '{workflow,steps,check_resume}',
    '{
        "action": "dispatch_thunder_prepare_resume_url",
        "config": { "expiry_minutes": 3000 },
        "next_step": "assemble_manifest",
        "output_field": "resume_probe",
        "description": "Probe B2 for the run latest checkpoint and presign a GET. training_run_id resolves from input_data via configOrInput fallback. found=false (no checkpoints) means fresh start; assemble_upload_manifest emits a resume block only when resume_url is non-empty, so this is safe on first launches and auto-resumes re-runs of the same training_run_id."
    }'::jsonb, true)
WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

-- 3. assemble_manifest reads the resume probe. These are OPTIONAL inputs on
--    assemble_upload_manifest; it only adds the manifest resume block when
--    resume_url resolves non-empty, so fresh launches are unaffected.
UPDATE agent_definitions
SET default_config = jsonb_set(
    jsonb_set(
      jsonb_set(
        default_config,
        '{workflow,steps,assemble_manifest,config,resume_url}',
        '"resume_probe.presigned_url"'::jsonb, true),
      '{workflow,steps,assemble_manifest,config,resume_key}',
      '"resume_probe.key"'::jsonb, true),
    '{workflow,steps,assemble_manifest,config,resume_index}',
    '"resume_probe.index"'::jsonb, true)
WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

COMMIT;

-- Verify (run after commit):
--   compute_keys -> presign_checkpoints(batch) -> presign_final -> check_resume
--     -> assemble_manifest -> write_manifest -> ssh_exec_launch -> mark_running -> complete
-- SELECT jsonb_pretty(default_config #> '{workflow,steps,check_resume}')          AS check_resume,
--        (default_config #>> '{workflow,steps,presign_final,next_step}')          AS presign_final_next,
--        jsonb_pretty(default_config #> '{workflow,steps,assemble_manifest,config}') AS assemble_config
--   FROM agent_definitions
--  WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

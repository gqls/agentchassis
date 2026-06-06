-- 109a — fix write_manifest /workspace permission
-- Apply to clients_db. Use this because 109 is ALREADY applied; this patches the live row.
-- (The canonical 109 has been corrected too, for a clean re-apply from scratch.)
--
-- Why:
--   write_manifest is the FIRST launcher step to touch the VM filesystem (every step
--   before it is presign/pure — no SSH to the box). It used a NON-sudo
--   `mkdir -p /workspace`. The very next step, ssh_exec_launch, creates /workspace with
--   `sudo mkdir -p /workspace && sudo chown $(id -u):$(id -g) /workspace` — which proves /
--   is not writable by the ssh user on these boxes. So write_manifest's non-sudo mkdir
--   (or, if /workspace happens to pre-exist root-owned, the `> /workspace/...` redirect)
--   fails before the manifest is written, and the launch dies at write_manifest.
--   iter_0 never hit this (the old launcher had no write_manifest step).
--
-- Fix: make write_manifest create /workspace exactly as ssh_exec_launch does
--   (sudo mkdir + sudo chown to the ssh user), then write the manifest as that user
--   (the redirect runs as the user, so /workspace must be user-owned at that point —
--   hence the chown, not just the mkdir). Surgical: one jsonb_set. Idempotent.

-- VERIFY BEFORE (expect the non-sudo `mkdir -p /workspace && echo ...`):
--   SELECT default_config #>> '{workflow,steps,write_manifest,config,command_template}'
--     FROM agent_definitions WHERE type='training-launcher'
--      AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,write_manifest,config,command_template}',
    to_jsonb(
      'sudo mkdir -p /workspace && sudo chown $(id -u):$(id -g) /workspace && echo ''{manifest_b64}'' | base64 -d > /workspace/upload_manifest.json && echo MANIFEST_WRITTEN'::text
    ),
    false
)
WHERE type='training-launcher'
  AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

COMMIT;

-- VERIFY AFTER (expect the `sudo mkdir -p /workspace && sudo chown ...` prefix):
--   SELECT default_config #>> '{workflow,steps,write_manifest,config,command_template}'
--     FROM agent_definitions WHERE type='training-launcher'
--      AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

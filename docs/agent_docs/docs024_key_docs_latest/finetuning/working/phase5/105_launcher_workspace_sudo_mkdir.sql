-- 105_launcher_workspace_sudo_mkdir.sql
-- DB: clients_db  (the live flywheel-C agent_definitions — NOT templates_db)
--
-- Problem (observed live 2026-06-03, orch cd906623 / launcher orch b002b359):
--   The full iter_0 chain ran end-to-end for the first time after 104. The
--   training-launcher's ssh_exec_launch connected, returned exit_code 0 and
--   LAUNCH_PID=193 — but training never started. stderr was:
--       mkdir: cannot create directory '/workspace': Permission denied
--       bash: line 1: /workspace/launch.log: No such file or directory
--
--   The command_template did a *plain* `mkdir -p /workspace`. The `ubuntu` ssh
--   user cannot create a dir at `/`, so the mkdir failed; the curls/run.sh had
--   nowhere to land, and the `&`-backgrounded setsid job died immediately. The
--   exit_code is 0 only because the command's last token is `echo` (the known
--   detached-ssh_exec false-success: a VM-side failure looks like success).
--
-- Why this fix (and why it is NOT a patch):
--   The bundle's OWN setup script already establishes the convention —
--   00_vm_setup.sh L51-52:
--       sudo mkdir -p "${WORKSPACE}"
--       sudo chown "$(id -u):$(id -g)" "${WORKSPACE}"
--   i.e. /workspace is meant to be created with sudo and chowned to the caller.
--   The launcher's command_template diverged (plain mkdir) AND runs the curls
--   before 00_vm_setup.sh executes. This migration makes the command_template
--   mirror the script: create + chown /workspace up front, as the running user.
--   sudo is known-good on these Thunder instances (the whole setup script uses
--   it) and /workspace on the root volume has the 100GB the prior manual run
--   used — so no re-bundle is needed: run.sh and 00_vm_setup.sh keep /workspace
--   (its sudo-mkdir simply becomes idempotent).
--
-- Shape change: in-place edit of one string. The chassis loads the def per
--   orchestrate (no restart needed). No version bump (consistent with 104).
--
-- Target def: training-launcher (active, non-snapshot). Verify the BEFORE/AFTER
--   SELECT shows exactly one row and the expected string.

BEGIN;

SELECT 'training-launcher ssh_exec_launch.command_template BEFORE' AS label,
       default_config #> '{workflow,steps,ssh_exec_launch,config,command_template}' AS value
FROM public.agent_definitions
WHERE type = 'training-launcher'
  AND is_active = true
  AND COALESCE(is_snapshot, false) = false;

UPDATE public.agent_definitions
   SET default_config = jsonb_set(
       default_config,
       '{workflow,steps,ssh_exec_launch,config,command_template}',
       to_jsonb($cmd$sudo mkdir -p /workspace && sudo chown $(id -u):$(id -g) /workspace; setsid bash -c 'curl -fsSL "{scripts_url}" -o /workspace/bundle.tar.gz && tar -xzf /workspace/bundle.tar.gz -C /workspace && curl -fsSL "{dataset_url}" -o /workspace/training_iter0.jsonl && chmod +x /workspace/run.sh && /workspace/run.sh > /workspace/train.log 2>&1' < /dev/null > /workspace/launch.log 2>&1 & echo "LAUNCH_PID=$!"$cmd$::text),
       false
   )
 WHERE type = 'training-launcher'
   AND is_active = true
   AND COALESCE(is_snapshot, false) = false;

SELECT 'training-launcher ssh_exec_launch.command_template AFTER' AS label,
       default_config #> '{workflow,steps,ssh_exec_launch,config,command_template}' AS value
FROM public.agent_definitions
WHERE type = 'training-launcher'
  AND is_active = true
  AND COALESCE(is_snapshot, false) = false;

-- Expect: UPDATE 1, and the AFTER value beginning with
--   sudo mkdir -p /workspace && sudo chown $(id -u):$(id -g) /workspace; setsid ...
-- If satisfied:
COMMIT;
-- else: ROLLBACK;

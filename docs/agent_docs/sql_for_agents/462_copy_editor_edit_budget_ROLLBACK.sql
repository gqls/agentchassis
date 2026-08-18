-- 462_copy_editor_edit_budget_ROLLBACK.sql
--
-- Reverses 462 by restoring the snapshot that 462 took before editing, which is the
-- only faithful way back: the prompt edit is a string replacement, and reversing it by
-- a second replacement would silently no-op if anything else has touched the prompt
-- since. Restoring the snapshot fails loudly instead if there is no snapshot to restore.
--
-- ⚠ This also reverts any OTHER change made to copy-editor after 462. Check first:
--   SELECT updated_at FROM agent_definitions WHERE type='copy-editor'
--     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

BEGIN;

DO $$
DECLARE
  snap jsonb;
  n int;
BEGIN
  -- ⚠ snapshot_agent(type, reason) — the TWO-ARG overload 462 calls — writes to
  -- `agent_definitions_backup` with the reason in `snapshot_reason`. The ONE-ARG
  -- overload is the one that inserts an is_snapshot row into agent_definitions.
  -- Reading the wrong one here would raise "no snapshot found" on a snapshot that
  -- exists, which reads as a missing backup rather than as a wrong query.
  SELECT default_config INTO snap
    FROM agent_definitions_backup
   WHERE type = 'copy-editor'
     AND snapshot_reason LIKE '%462_copy_editor_edit_budget.sql: pre-update%'
   ORDER BY snapshot_taken_at DESC LIMIT 1;

  IF snap IS NULL THEN
    RAISE EXCEPTION 'no 462 pre-update snapshot found — refusing to guess at the prior config';
  END IF;

  UPDATE agent_definitions
     SET default_config = snap, updated_at = now()
   WHERE type = 'copy-editor' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'copy-editor' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND (default_config #>> '{workflow,steps,run_copy_edit,config,ai_service,max_tokens}') = '16000';
  IF n <> 1 THEN
    RAISE EXCEPTION 'rollback did not restore max_tokens=16000 (matched % row(s))', n;
  END IF;

  RAISE NOTICE 'copy-editor restored to its pre-462 config (max_tokens 16000, no edit budget)';
END $$;

COMMIT;

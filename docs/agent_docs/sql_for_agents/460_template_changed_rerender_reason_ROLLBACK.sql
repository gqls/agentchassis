-- 460 ROLLBACK — restore both rows from the snapshots taken by 460's
-- snapshot_agent() calls (agent_definitions_backup, reason '460 pre-image: …').
-- Restore by copying default_config back from the newest matching backup row;
-- verify with the same DO/RAISE shape inverted (condition must NOT contain
-- template_changed afterwards). Written for the operator; not auto-applied.
DO $$
BEGIN
  RAISE NOTICE 'Restore page-rerender and component-template-fixer default_config from agent_definitions_backup rows with reason LIKE ''460 pre-image%%'' (newest each), then re-run the checks in 460 inverted.';
END $$;

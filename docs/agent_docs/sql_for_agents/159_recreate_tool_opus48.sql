-- 159_recreate_tool_opus48.sql
--
-- Upgrade tool-recreation-handler's recreate_tool step from claude-opus-4-6
-- to claude-opus-4-8 (user decision 2026-07-16, following 158's Sonnet 5
-- upgrade of the pipeline's 7 Sonnet steps). recreate_tool is the 64k-token
-- code-generation step that rebuilds a tool from its adopted source — the one
-- step that deliberately runs on the Opus tier.
--
-- Safe without a chassis rebuild: ResolveModelAlias passes unknown ids
-- through unchanged (claude-opus-4-8 is also in model_aliases.go as of the
-- working tree, riding the next image), and the Anthropic client sends no
-- temperature.
--
-- Applied out of band (psql -f + manual ledger row): the migration runner is
-- still blocked at 151_gripper_spec_sheet_component.sql, with 152-156 (other
-- workstreams' files) pending behind it.

BEGIN;

SELECT snapshot_agent('tool-recreation-handler', '159_recreate_tool_opus48: pre-update');

DO $$
DECLARE
  n int;
BEGIN
  UPDATE agent_definitions
  SET default_config = jsonb_set(default_config,
        '{workflow,steps,recreate_tool,config,ai_service,model}',
        '"claude-opus-4-8"'::jsonb)
  WHERE type = 'tool-recreation-handler'
    AND is_active
    AND default_config #>> '{workflow,steps,recreate_tool,config,ai_service,model}' = 'claude-opus-4-6';

  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN
    RAISE EXCEPTION '159: expected exactly 1 recreate_tool step on claude-opus-4-6, updated %', n;
  END IF;
END $$;

-- Convention: workflow-altering migrations leave a pipeline note.
INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## recreate_tool upgraded to claude-opus-4-8
Observed: tool-recreation-handler''s recreate_tool step (64k-token tool rebuild) ran on claude-opus-4-6.
Root cause: n/a — routine model upgrade, user decision 2026-07-16 (completes 158, which moved the pipeline''s 7 Sonnet steps to claude-sonnet-5).
Fix: recreate_tool moved to claude-opus-4-8 (migration 159); agent snapshot taken pre-update.
Verified: post-apply step/model map shows recreate_tool on claude-opus-4-8; max_tokens (64000) and all other step config untouched.
Categories: fix',
  '["fix"]'::jsonb,
  'migration', '159_recreate_tool_opus48'
);

COMMIT;

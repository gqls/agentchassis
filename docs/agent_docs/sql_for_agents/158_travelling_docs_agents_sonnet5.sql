-- 158_travelling_docs_agents_sonnet5.sql
--
-- Upgrade the travelling-docs tool-pipeline agents from claude-sonnet-4-6 to
-- claude-sonnet-5 (user decision 2026-07-16). Exactly these 7 LLM steps:
--
--   tool-generator            generate_tool_html, compose_plan
--   tool-improver             improve_tool, compose_note
--   tool-recreation-handler   analyze_tool, compose_note
--   component-template-fixer  compose_note
--
-- DELIBERATELY UNTOUCHED: tool-recreation-handler.recreate_tool stays on
-- claude-opus-4-6 (64k-token code generation chose the Opus tier on purpose;
-- an upgrade there would be claude-opus-4-8 — a separate decision).
-- tool-acceptance-agent has no LLM steps (pure Go actions).
--
-- Safe without a chassis rebuild: ResolveModelAlias passes unknown ids
-- through unchanged, and the Anthropic client sends no temperature (the
-- Claude 5 family rejects non-default temperature). Precedent: diagnose-agent
-- has been running claude-sonnet-5 through this same chassis in production
-- since 2026-07-16 (fixloop workstream).
--
-- Applied out of band (psql -f + manual ledger row): the migration runner is
-- still blocked at 151_gripper_spec_sheet_component.sql, with 152-156 (other
-- workstreams' files) pending behind it.

BEGIN;

SELECT snapshot_agent('tool-generator',           '158_travelling_docs_agents_sonnet5: pre-update');
SELECT snapshot_agent('tool-improver',            '158_travelling_docs_agents_sonnet5: pre-update');
SELECT snapshot_agent('tool-recreation-handler',  '158_travelling_docs_agents_sonnet5: pre-update');
SELECT snapshot_agent('component-template-fixer', '158_travelling_docs_agents_sonnet5: pre-update');

DO $$
DECLARE
  r record;
  n int := 0;
BEGIN
  FOR r IN
    SELECT a.id, k AS step
    FROM agent_definitions a,
         LATERAL jsonb_object_keys(a.default_config #> '{workflow,steps}') k
    WHERE a.type IN ('tool-generator','tool-improver','tool-recreation-handler','component-template-fixer')
      AND a.is_active
      AND a.default_config #>> ARRAY['workflow','steps',k,'config','ai_service','model'] = 'claude-sonnet-4-6'
  LOOP
    UPDATE agent_definitions
    SET default_config = jsonb_set(default_config,
          ARRAY['workflow','steps',r.step,'config','ai_service','model'],
          '"claude-sonnet-5"'::jsonb)
    WHERE id = r.id;
    n := n + 1;
  END LOOP;

  IF n <> 7 THEN
    RAISE EXCEPTION '158: expected exactly 7 claude-sonnet-4-6 steps across the 4 agents, found %', n;
  END IF;
END $$;

-- Convention: workflow-altering migrations leave a pipeline note.
INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## Tool-pipeline agents upgraded to claude-sonnet-5
Observed: tool-generator, tool-improver, tool-recreation-handler and component-template-fixer ran their LLM steps on claude-sonnet-4-6 (the chassis default generation).
Root cause: n/a — routine model upgrade, user decision 2026-07-16.
Fix: 7 steps moved to claude-sonnet-5 (migration 158); recreate_tool deliberately stays on claude-opus-4-6. Agent snapshots taken pre-update.
Verified: post-apply step/model map shows 7x claude-sonnet-5 + 1x claude-opus-4-6; diagnose-agent precedent already proved sonnet-5 through this chassis in production.
Categories: fix',
  '["fix"]'::jsonb,
  'migration', '158_travelling_docs_agents_sonnet5'
);

COMMIT;

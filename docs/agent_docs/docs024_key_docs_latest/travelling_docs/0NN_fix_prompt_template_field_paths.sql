-- 0NN_fix_prompt_template_field_paths.sql — fix TEMPLATE_FIELD_ERROR in compose_plan
-- and harden the Task-4 note templates against the same class of fault.
-- DRAFT 2026-07-09. Renumber 0NN. DB-only; effective immediately.
--
-- EVIDENCE (agent_error_log, orchestration 036cd3bd, 2026-07-09 11:41:36):
--   step compose_plan failed: execute_llm_prompt: failed to render prompt
--   template: ... at <.generated_html.result>: can't evaluate field result in
--   type interface {}   [error_code TEMPLATE_FIELD_ERROR]
-- The run still COMPLETED at `complete` and the tool was created: the doc
-- steps' config.error_step containment worked exactly as designed.
--
-- THE RULE (proven by this run, not assumed):
--   * execute_llm_prompt with output_format "text"  -> the prompt template
--     receives the BARE STRING. Use {{.generated_html}}. Reaching .result on it
--     errors, because a string has no fields.
--   * execute_llm_prompt with output_format "json"  -> the template receives a
--     map. Use {{.tool_analysis.result | toJSON}} — the form already live and
--     working in tool-recreation-handler.recreate_tool.
--   * ACTION CONFIG field paths are a DIFFERENT resolver and keep the .result
--     suffix: save_tool's html_content = "generated_html.result" resolved fine
--     in the same run (the action hard-fails on empty html, and did not).
--   Corroboration: {{.site_record.domain}} (a map) renders fine in the same
--   template, so map traversal is not the problem.
--
-- CHANGES (four agents; three are pre-emptive corrections of the same class of
-- bug introduced by the Task-4 migration, before they ever fire):
--   1. tool-generator.compose_plan       — THE BLOCKER: {{.generated_html.result}} -> {{.generated_html}}
--   2. tool-recreation-handler.compose_note — nested LLM-JSON key access ({{.tool_analysis.result.tool_name}} etc,
--      keys never verified) -> the proven {{.tool_analysis.result | toJSON}} dump; {{.deploy_result | toJSON}}
--   3. component-template-fixer.compose_note — {{.input_data}} / {{.fix_result}} -> | toJSON
--   4. tool-improver.compose_note           — {{.update_result}} -> | toJSON
-- Replaces are idempotent (a second run is a no-op).

BEGIN;

SELECT snapshot_agent('tool-generator',           '0NN_fix_prompt_template_field_paths.sql: pre-update');
SELECT snapshot_agent('tool-recreation-handler',  '0NN_fix_prompt_template_field_paths.sql: pre-update');
SELECT snapshot_agent('component-template-fixer', '0NN_fix_prompt_template_field_paths.sql: pre-update');
SELECT snapshot_agent('tool-improver',            '0NN_fix_prompt_template_field_paths.sql: pre-update');

-- 1) tool-generator.compose_plan — the blocker.
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config = jsonb_set(
        ad.default_config,
        '{workflow,steps,compose_plan,config,prompt_template}',
        to_jsonb(replace(
            ad.default_config #>> '{workflow,steps,compose_plan,config,prompt_template}',
            '{{.generated_html.result}}',
            '{{.generated_html}}')),
        true),
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

-- 2) tool-recreation-handler.compose_note — rewrite the input block wholesale.
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config = jsonb_set(
        ad.default_config,
        '{workflow,steps,compose_note,config,prompt_template}',
        to_jsonb(
          replace(
            replace(
              replace(
                replace(
                  ad.default_config #>> '{workflow,steps,compose_note,config,prompt_template}',
                  'Tool: {{.tool_analysis.result.tool_name}} ({{.tool_analysis.result.tool_type}})',
                  'Tool analysis (JSON): {{.tool_analysis.result | toJSON}}'),
                'Purpose: {{.tool_analysis.result.purpose}}' || E'\n',
                ''),
              'Deploy result (a Go map dump — read it as data): {{.deploy_result}}',
              'Deploy result (JSON): {{.deploy_result | toJSON}}'),
            '## <short title: recreated <tool name>>',
            '## <short title: recreated the tool named in the analysis JSON>')),
        true),
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

-- 3) component-template-fixer.compose_note
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'component-template-fixer' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config = jsonb_set(
        ad.default_config,
        '{workflow,steps,compose_note,config,prompt_template}',
        to_jsonb(
          replace(
            replace(
              ad.default_config #>> '{workflow,steps,compose_note,config,prompt_template}',
              'Work item input (a Go map dump — read it as data): {{.input_data}}',
              'Work item input (JSON): {{.input_data | toJSON}}'),
            'Fix result: {{.fix_result}}',
            'Fix result (JSON): {{.fix_result | toJSON}}')),
        true),
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

-- 4) tool-improver.compose_note
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-improver' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config = jsonb_set(
        ad.default_config,
        '{workflow,steps,compose_note,config,prompt_template}',
        to_jsonb(replace(
            ad.default_config #>> '{workflow,steps,compose_note,config,prompt_template}',
            'Update result (a Go map dump — read it as data): {{.update_result}}',
            'Update result (JSON): {{.update_result | toJSON}}')),
        true),
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

-- Guards: new forms present, old forms gone.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,compose_plan,config,prompt_template}' LIKE '%{{.generated_html}}%'
      AND default_config #>> '{workflow,steps,compose_plan,config,prompt_template}' NOT LIKE '%generated_html.result%';
    IF n <> 1 THEN RAISE EXCEPTION 'compose_plan template not fixed (found %)', n; END IF;

    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,compose_note,config,prompt_template}' LIKE '%tool_analysis.result | toJSON%'
      AND default_config #>> '{workflow,steps,compose_note,config,prompt_template}' NOT LIKE '%tool_analysis.result.tool_%'
      AND default_config #>> '{workflow,steps,compose_note,config,prompt_template}' NOT LIKE '%tool_analysis.result.purpose%';
    IF n <> 1 THEN RAISE EXCEPTION 'recreation compose_note template not fixed (found %)', n; END IF;

    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'component-template-fixer' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,compose_note,config,prompt_template}' LIKE '%input_data | toJSON%'
      AND default_config #>> '{workflow,steps,compose_note,config,prompt_template}' LIKE '%fix_result | toJSON%';
    IF n <> 1 THEN RAISE EXCEPTION 'fixer compose_note template not fixed (found %)', n; END IF;

    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-improver' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,compose_note,config,prompt_template}' LIKE '%update_result | toJSON%';
    IF n <> 1 THEN RAISE EXCEPTION 'improver compose_note template not fixed (found %)', n; END IF;
END $$;

COMMIT;

-- Verify after apply:
--   SELECT type,
--          (default_config #>> '{workflow,steps,compose_plan,config,prompt_template}') LIKE '%{{.generated_html}}%'   AS plan_fixed,
--          (default_config #>> '{workflow,steps,compose_note,config,prompt_template}') LIKE '%toJSON%'                AS note_json
--   FROM agent_definitions
--   WHERE type IN ('tool-generator','tool-recreation-handler','component-template-fixer','tool-improver')
--     AND deleted_at IS NULL ORDER BY type;
--
-- Re-run the Task-3 proof with a NEW function name — tool-xp-curve-designer now
-- exists and idx_cc_tool_function_unique (function WHERE component_level='tool'
-- AND forked_from IS NULL AND is_active) would reject a second active row:
--   SPEC_FUNCTION=tool-drop-rate-tuner SPEC_NAME="Drop Rate Tuner" \
--   SPEC_DESC="Calculate the chance of at least one drop across N attempts at a given drop rate, with a cumulative-probability table and the attempts needed for 50/90/99% confidence." \
--     ./drafts/085_TRIGGER_toolgen_gamesdesign_v1.sh
--
-- Proof:
--   SELECT subject_key, is_current, body LIKE '%```criteria%' AS has_fence,
--          length(body) AS body_len, source, created_at
--   FROM doc_plans WHERE source='tool-generator' ORDER BY created_at DESC LIMIT 3;
--   SELECT body FROM doc_plans WHERE subject_key='tool-drop-rate-tuner' AND is_current;  -- composer review
--
-- Backfill note: tool-xp-curve-designer (created 2026-07-09) has no PLAN — it
-- predates the working hook. Leave it; a later improve/recreate run, or a small
-- one-off, can write one.
--
-- Rollback: restore the four agents from the snapshots taken at the top, or
-- reverse each replace (they are pure string substitutions).

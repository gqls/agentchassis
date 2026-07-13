-- 0NN_tool_generator_plan_writing.sql — Task 3: automate PLAN-writing at tool creation.
-- DRAFT 2026-07-07. Renumber 0NN. Requires: doc tables live (they are),
-- write_doc_plan + rag_index deployed (they are). DB-only; effective immediately.
--
-- What this does (grounded in the live workflow pasted 2026-07-07):
--   ensure_site_record -> load_brand_context -> generate_tool_html -> save_tool
--   -> [NEW] compose_plan -> [NEW] write_plan -> [NEW] index_plan -> complete
-- The PLAN is written AFTER save_tool succeeds (no component => no PLAN; a
-- failed save still routes to complete_error, untouched). Every new step has
-- config.error_step = "complete": a docs failure can NEVER fail tool creation.
--
-- Deliberate changes beyond the insertions (noted per the house rule):
--   * save_tool.next_step: "complete" -> "compose_plan" (the rewire).
--   * workflow.timeout_seconds: 300 -> 480 (a second Sonnet call is added).
--   * correct-while-touching (001 §16): the THREE inert step-LEVEL error_steps
--     (save_tool, generate_tool_html, load_brand_context) move INTO config
--     with their original targets, and the dead step-level keys are DELETED.
--
-- Standing rule: the snapshot is the FIRST statement inside the transaction —
-- MVCC captures the pre-update state; if anything below fails, snapshot and
-- changes roll back together.

BEGIN;

-- 0) Snapshot (standing rule, 2026-07-06).
SELECT snapshot_agent('tool-generator', '0NN_tool_generator_plan_writing.sql: pre-update');

-- 1) Add the three steps, rewire save_tool, extend the timeout.
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config =
    jsonb_set(
      jsonb_set(
        jsonb_set(
          jsonb_set(
            jsonb_set(ad.default_config,
              '{workflow,steps,compose_plan}', $cp$
{
  "action": "execute_llm_prompt",
  "config": {
    "ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 4000, "api_key_env_var": "ANTHROPIC_API_KEY"},
    "input_fields": ["input_data", "site_record", "generated_html"],
    "output_format": "text",
    "error_step": "complete",
    "prompt_template": "You are writing the travelling PLAN document for a website tool that was just generated.\n\n## Tool\nFunction: {{.input_data.spec.function}}\nName: {{.input_data.spec.name}}\nDescription: {{.input_data.spec.description}}\nSite: {{.site_record.domain}}\n\n## The generated tool HTML (selectors and behaviour come from here)\n{{.generated_html.result}}\n\n## Output\nOutput ONLY a markdown document, no fences around it, starting exactly with:\n# PLAN — {{.input_data.spec.function}}\n\nSections, in this order:\n\n## Aim\nOne short paragraph, product terms.\n\n## Source spec\nThe spec fields above restated in one or two lines.\n\n## Behaviour contract\nThe states, inputs and outputs the code keeps. Read the tool-doc header comment inside the script and the code itself. All logic is client-side; no external calls.\n\n## Acceptance criteria\nA fenced block exactly like:\n```criteria\n{ \"profiles\": [\"desktop\",\"mobile\"],\n  \"checks\": [ ... ] }\n```\nAlways include these five checks verbatim: {\"id\":\"boots\",\"type\":\"selector_exists\",\"selector\":\".tool-container\"}, {\"id\":\"console\",\"type\":\"no_console_errors\"}, {\"id\":\"asset\",\"type\":\"asset_loads\",\"path\":\"/tools/assets/{{.input_data.spec.function}}.js\"}, {\"id\":\"status\",\"type\":\"page_status_ok\"}, {\"id\":\"mobile-fit\",\"type\":\"no_horizontal_overflow\",\"profiles\":[\"mobile\"]}. Add ONE interaction check ONLY if you can copy real ids or classes from the HTML above (fill or click steps plus an expect selector). NEVER invent a selector — if unsure, omit the interaction check entirely. The JSON must be valid; ids lowercase-kebab.\n\n## Delivery mechanism\nOne line: Path 1 — component inline script, extracted to /tools/assets/{{.input_data.spec.function}}.js on rerender.\n\n## Dependencies\nData, assets or components this tool relies on; write None if none.\n\n## Deliberate decisions — do not re-fix\n2 to 4 bullets of intentional v1 choices a later pass must not \"fix\". Derive from the spec and the code. Include: v1 kept simple by design — the improvement loop iterates it against the criteria above.\n\nKeep the whole document under 3000 characters. No preamble, no explanation, no trailing commentary."
  },
  "next_step": "write_plan",
  "description": "Draft the travelling PLAN (body + criteria) from the spec and the generated HTML. Docs must never fail tool creation.",
  "output_field": "doc_plan"
}
$cp$::jsonb, true),
            '{workflow,steps,write_plan}', $wp$
{
  "action": "write_doc_plan",
  "config": {
    "subject_type": "tool",
    "subject_key_field": "input_data.spec.function",
    "plan_body_field": "doc_plan.result",
    "plan_source": "tool-generator",
    "created_by": "tool-generator",
    "error_step": "complete"
  },
  "next_step": "index_plan",
  "description": "Persist the PLAN (supersede pattern). Error routes to complete — never fails creation.",
  "output_field": "doc_plan_write"
}
$wp$::jsonb, true),
          '{workflow,steps,index_plan}', $ip$
{
  "action": "rag_index",
  "config": {
    "content_field": "doc_plan.result",
    "collection": "tool_docs",
    "error_step": "complete"
  },
  "next_step": "complete",
  "description": "Derived retrieval copy into knowledge_base (tool_docs). source_type reads scrape until parameterised — accepted open item.",
  "output_field": "doc_plan_index"
}
$ip$::jsonb, true),
        '{workflow,steps,save_tool,next_step}', '"compose_plan"'::jsonb, true),
      '{workflow,timeout_seconds}', '480'::jsonb, true),
    updated_at = now()
FROM cur
WHERE ad.id = cur.id;

-- 2) Correct-while-touching: config-level error_step with the ORIGINAL targets.
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config =
    jsonb_set(
      jsonb_set(
        jsonb_set(ad.default_config,
          '{workflow,steps,save_tool,config,error_step}',          '"complete_error"'::jsonb, true),
        '{workflow,steps,generate_tool_html,config,error_step}',   '"complete_error"'::jsonb, true),
      '{workflow,steps,load_brand_context,config,error_step}',     '"generate_tool_html"'::jsonb, true),
    updated_at = now()
FROM cur
WHERE ad.id = cur.id;

-- 3) Delete the inert step-LEVEL keys (they were never read; leaving them
--    invites copying the broken shape).
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config =
    ((ad.default_config #- '{workflow,steps,save_tool,error_step}')
                        #- '{workflow,steps,generate_tool_html,error_step}')
                        #- '{workflow,steps,load_brand_context,error_step}',
    updated_at = now()
FROM cur
WHERE ad.id = cur.id;

-- Guard: exactly one row carries the full shape.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type = 'tool-generator' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,compose_plan,action}' = 'execute_llm_prompt'
      AND default_config #>> '{workflow,steps,write_plan,action}'   = 'write_doc_plan'
      AND default_config #>> '{workflow,steps,index_plan,action}'   = 'rag_index'
      AND default_config #>> '{workflow,steps,save_tool,next_step}' = 'compose_plan'
      AND default_config #>> '{workflow,steps,write_plan,config,subject_key_field}' = 'input_data.spec.function'
      AND default_config #>> '{workflow,steps,save_tool,config,error_step}'        = 'complete_error'
      AND default_config #>> '{workflow,steps,generate_tool_html,config,error_step}' = 'complete_error'
      AND default_config #>> '{workflow,steps,load_brand_context,config,error_step}' = 'generate_tool_html'
      AND (default_config #> '{workflow,steps,save_tool,error_step}')          IS NULL
      AND (default_config #> '{workflow,steps,generate_tool_html,error_step}') IS NULL
      AND (default_config #> '{workflow,steps,load_brand_context,error_step}') IS NULL
      AND default_config #>> '{workflow,timeout_seconds}' = '480';
    IF n <> 1 THEN
        RAISE EXCEPTION 'tool-generator plan-writing wiring incomplete (found %)', n;
    END IF;
END $$;

COMMIT;

-- Verify after apply:
--   SELECT default_config #>> '{workflow,steps,save_tool,next_step}'                    AS save_next,
--          default_config #>> '{workflow,steps,write_plan,config,subject_key_field}'    AS subject_key_field,
--          default_config #>> '{workflow,timeout_seconds}'                              AS timeout,
--          (default_config #> '{workflow,steps,save_tool,error_step}') IS NULL          AS step_level_removed
--   FROM agent_definitions
--   WHERE type='tool-generator' AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;
--
-- Proof (after the next tool creation):
--   SELECT subject_key, is_current, body LIKE '%```criteria%' AS has_fence, source, created_at
--   FROM doc_plans WHERE source = 'tool-generator' ORDER BY created_at DESC LIMIT 3;
--
-- Rollback: restore from the snapshot taken at the top (companion function —
-- see \df *agent*), or manually: save_tool.next_step -> "complete"; delete
-- steps compose_plan/write_plan/index_plan; timeout_seconds -> 300. The
-- error_step corrections can be kept — they are strictly corrective.

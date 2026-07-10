-- 0NN_fix_agents_note_writing.sql — Task 4: every fix appends a NOTES entry.
-- DRAFT 2026-07-07. Renumber 0NN. Requires: doc tables + append_doc_note live
-- (they are). DB-only; effective immediately.
--
-- Grounded in the three workflows fetched 2026-07-07 (403-line paste).
-- Insertions (SUCCESS paths only — a note records a change, not an attempt):
--   component-template-fixer: create_rerender -> [compose_note -> append_note] -> complete
--                             AND check_needs_rerender.else_step -> compose_note
--                             (both branches reach the note; a no-op pass is
--                             still a recorded pass). Subject: pipeline/build +
--                             note_site_id (site-scoped template work).
--   tool-improver:            create_rerender_item -> [compose_note -> append_note] -> complete.
--                             Subject: tool / tool_data.function (the REAL
--                             function, straight from the loaded component).
--   tool-recreation-handler:  deploy_page -> [compose_note -> append_note] -> complete.
--                             Subject: tool / input_data.spec.function —
--                             recreation items are page-scoped and may lack it;
--                             append then errors and routes to complete (note
--                             skipped, run unharmed). REFINEMENT ON RECORD:
--                             add function to recreation item specs at creation.
--   complete_not_found / complete_error paths: untouched — no change, no note.
--
-- Note body: LLM-drafted uniform entry ("## <title>" + Observed/Root cause/
-- Fix/Verified/Categories lines). The entry heading carries NO date — the
-- row's created_at is the timestamp (deviation from persist_diagnosis_note,
-- which embeds one; accepted). Machine categories v1 = ["fix"] (operational
-- tag); failure-class tags live in the body Categories line until a parser
-- justifies more.
--
-- Declared deliberate changes:
--   * component-template-fixer timeout 120 -> 240; tool-improver 300 -> 480
--     (each gains a Sonnet call). tool-recreation-handler stays 2400.
--   * correct-while-touching (001 §16): ALL TEN inert step-LEVEL error_steps
--     in tool-recreation-handler move INTO config with their ORIGINAL targets;
--     dead step-level keys are DELETED. (The other two agents carry none.)
--
-- Standing rule: three snapshot_agent calls open the transaction.

BEGIN;

SELECT snapshot_agent('component-template-fixer', '0NN_fix_agents_note_writing.sql: pre-update');
SELECT snapshot_agent('tool-improver',            '0NN_fix_agents_note_writing.sql: pre-update');
SELECT snapshot_agent('tool-recreation-handler',  '0NN_fix_agents_note_writing.sql: pre-update');

-- ── A) component-template-fixer ─────────────────────────────────────────────
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'component-template-fixer' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config =
    jsonb_set(
      jsonb_set(
        jsonb_set(
          jsonb_set(
            jsonb_set(ad.default_config,
              '{workflow,steps,compose_note}', $cn1$
{
  "action": "execute_llm_prompt",
  "config": {
    "ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 1200, "api_key_env_var": "ANTHROPIC_API_KEY"},
    "input_fields": ["input_data", "fix_result", "site_record"],
    "output_format": "text",
    "error_step": "complete",
    "prompt_template": "You are appending a maintenance NOTES entry after a website template fix.\n\nAgent: component-template-fixer\nSite: {{.site_record.domain}}\nWork item input (a Go map dump — read it as data): {{.input_data}}\nFix result: {{.fix_result}}\n\nOutput ONLY the entry, exactly this shape, no preamble:\n## <short title of the fix>\nObserved: <symptom or trigger, one line>\nRoot cause: <one line; write unknown if unknown>\nFix: <what changed, one or two lines>\nVerified: <how success was checked, or pending rerender>\nCategories: <comma-separated tags from: css-variable-mismatch, empty-shell, mode-b-template, broken-template-slots, content-vs-runtime-mismatch, detool-on-rebuild, js-not-extracted, js-bundle-stale, schema-template-drift, acceptance-fail, seam, fix>\n\nUnder 600 characters. If the fix result shows fixed=false, title it no-op fix pass and say nothing changed."
  },
  "next_step": "append_note",
  "description": "Draft the NOTES entry for this fix. Docs must never fail the fix.",
  "output_field": "note_draft"
}
$cn1$::jsonb, true),
            '{workflow,steps,append_note}', $an1$
{
  "action": "append_doc_note",
  "config": {
    "subject_type": "pipeline",
    "subject_key": "build",
    "note_body_field": "note_draft.result",
    "note_categories": ["fix"],
    "note_source": "component-template-fixer",
    "created_by": "component-template-fixer",
    "note_site_id_field": "site_record.site_id",
    "error_step": "complete"
  },
  "next_step": "complete",
  "description": "Append the entry (site-scoped template work files under pipeline/build).",
  "output_field": "note_write"
}
$an1$::jsonb, true),
          '{workflow,steps,create_rerender,next_step}', '"compose_note"'::jsonb, true),
        '{workflow,steps,check_needs_rerender,config,else_step}', '"compose_note"'::jsonb, true),
      '{workflow,timeout_seconds}', '240'::jsonb, true),
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

-- ── B) tool-improver ────────────────────────────────────────────────────────
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-improver' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config =
    jsonb_set(
      jsonb_set(
        jsonb_set(
          jsonb_set(ad.default_config,
            '{workflow,steps,compose_note}', $cn2$
{
  "action": "execute_llm_prompt",
  "config": {
    "ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 1200, "api_key_env_var": "ANTHROPIC_API_KEY"},
    "input_fields": ["input_data", "tool_data", "update_result"],
    "output_format": "text",
    "error_step": "complete",
    "prompt_template": "You are appending a maintenance NOTES entry after an interactive tool fix.\n\nAgent: tool-improver\nTool function: {{.tool_data.function}}\nPage: {{.tool_data.page_name}} ({{.tool_data.page_url}})\nIssue that was reported: {{.input_data.issue}}\nUpdate result (a Go map dump — read it as data): {{.update_result}}\n\nOutput ONLY the entry, exactly this shape, no preamble:\n## <short title of the fix>\nObserved: <the reported issue, one line>\nRoot cause: <one line; write unknown if the fix was applied without confirming cause>\nFix: <what the improved HTML changed, one or two lines>\nVerified: <pending rerender + acceptance criteria, unless something stronger is known>\nCategories: <comma-separated tags from: css-variable-mismatch, empty-shell, mode-b-template, broken-template-slots, content-vs-runtime-mismatch, detool-on-rebuild, js-not-extracted, js-bundle-stale, schema-template-drift, acceptance-fail, seam, fix>\n\nUnder 600 characters."
  },
  "next_step": "append_note",
  "description": "Draft the NOTES entry for this fix. Docs must never fail the fix.",
  "output_field": "note_draft"
}
$cn2$::jsonb, true),
          '{workflow,steps,append_note}', $an2$
{
  "action": "append_doc_note",
  "config": {
    "subject_type": "tool",
    "subject_key_field": "tool_data.function",
    "note_body_field": "note_draft.result",
    "note_categories": ["fix"],
    "note_source": "tool-improver",
    "created_by": "tool-improver",
    "note_site_id_field": "input_data.site_id",
    "error_step": "complete"
  },
  "next_step": "complete",
  "description": "Append the entry to the tool's travelling NOTES.",
  "output_field": "note_write"
}
$an2$::jsonb, true),
        '{workflow,steps,create_rerender_item,next_step}', '"compose_note"'::jsonb, true),
      '{workflow,timeout_seconds}', '480'::jsonb, true),
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

-- ── C) tool-recreation-handler: note tail ───────────────────────────────────
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config =
    jsonb_set(
      jsonb_set(
        jsonb_set(ad.default_config,
          '{workflow,steps,compose_note}', $cn3$
{
  "action": "execute_llm_prompt",
  "config": {
    "ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 1200, "api_key_env_var": "ANTHROPIC_API_KEY"},
    "input_fields": ["input_data", "tool_analysis", "deploy_result", "page_record"],
    "output_format": "text",
    "error_step": "complete",
    "prompt_template": "You are appending a maintenance NOTES entry after an interactive tool was RECREATED from an adopted page.\n\nAgent: tool-recreation-handler\nPage: {{.page_record.name}}\nTool: {{.tool_analysis.result.tool_name}} ({{.tool_analysis.result.tool_type}})\nPurpose: {{.tool_analysis.result.purpose}}\nDeploy result (a Go map dump — read it as data): {{.deploy_result}}\n\nOutput ONLY the entry, exactly this shape, no preamble:\n## <short title: recreated <tool name>>\nObserved: <why recreation ran — adoption or repair, from the inputs>\nRoot cause: <one line; write not-applicable for a fresh adoption>\nFix: <recreated as self-contained HTML/CSS/JS; note anything the spec called out>\nVerified: <completeness + validation passed; deployed via page renderer>\nCategories: <comma-separated tags from: css-variable-mismatch, empty-shell, mode-b-template, broken-template-slots, content-vs-runtime-mismatch, detool-on-rebuild, js-not-extracted, js-bundle-stale, schema-template-drift, acceptance-fail, seam, fix>\n\nUnder 600 characters."
  },
  "next_step": "append_note",
  "description": "Draft the NOTES entry for this recreation. Docs must never fail the run.",
  "output_field": "note_draft"
}
$cn3$::jsonb, true),
        '{workflow,steps,append_note}', $an3$
{
  "action": "append_doc_note",
  "config": {
    "subject_type": "tool",
    "subject_key_field": "input_data.spec.function",
    "note_body_field": "note_draft.result",
    "note_categories": ["fix"],
    "note_source": "tool-recreation-handler",
    "created_by": "tool-recreation-handler",
    "note_site_id_field": "site_record.site_id",
    "error_step": "complete"
  },
  "next_step": "complete",
  "description": "Append the entry. Recreation items may lack spec.function — append then errors and routes to complete (note skipped, run unharmed). Refinement on record: stamp function into recreation item specs at creation.",
  "output_field": "note_write"
}
$an3$::jsonb, true),
      '{workflow,steps,deploy_page,next_step}', '"compose_note"'::jsonb, true),
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

-- ── D) tool-recreation-handler: correct-while-touching — config-level
--       error_step with the ORIGINAL targets (ten steps) ─────────────────────
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config =
  jsonb_set(jsonb_set(jsonb_set(jsonb_set(jsonb_set(
  jsonb_set(jsonb_set(jsonb_set(jsonb_set(jsonb_set(ad.default_config,
    '{workflow,steps,deploy_page,config,error_step}',           '"complete_error"'::jsonb, true),
    '{workflow,steps,analyze_tool,config,error_step}',          '"complete_error"'::jsonb, true),
    '{workflow,steps,recreate_tool,config,error_step}',         '"complete_error"'::jsonb, true),
    '{workflow,steps,save_sections,config,error_step}',         '"complete_error"'::jsonb, true),
    '{workflow,steps,check_completeness,config,error_step}',    '"complete_error"'::jsonb, true),
    '{workflow,steps,validate_tool,config,error_step}',         '"save_sections"'::jsonb, true),
    '{workflow,steps,load_site_specs,config,error_step}',       '"load_existing_content"'::jsonb, true),
    '{workflow,steps,save_training_data,config,error_step}',    '"validate_tool"'::jsonb, true),
    '{workflow,steps,load_related_context,config,error_step}',  '"analyze_tool"'::jsonb, true),
    '{workflow,steps,load_existing_content,config,error_step}', '"load_related_context"'::jsonb, true),
  updated_at = now()
FROM cur WHERE ad.id = cur.id;

-- ── E) tool-recreation-handler: delete the ten dead step-LEVEL keys ─────────
WITH cur AS (
    SELECT id FROM agent_definitions
    WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL
    ORDER BY version DESC LIMIT 1
)
UPDATE agent_definitions ad
SET default_config =
    ad.default_config
      #- '{workflow,steps,deploy_page,error_step}'
      #- '{workflow,steps,analyze_tool,error_step}'
      #- '{workflow,steps,recreate_tool,error_step}'
      #- '{workflow,steps,save_sections,error_step}'
      #- '{workflow,steps,check_completeness,error_step}'
      #- '{workflow,steps,validate_tool,error_step}'
      #- '{workflow,steps,load_site_specs,error_step}'
      #- '{workflow,steps,save_training_data,error_step}'
      #- '{workflow,steps,load_related_context,error_step}'
      #- '{workflow,steps,load_existing_content,error_step}',
    updated_at = now()
FROM cur WHERE ad.id = cur.id;

-- ── Guards ───────────────────────────────────────────────────────────────────
DO $$
DECLARE n int;
BEGIN
    -- A: component-template-fixer
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'component-template-fixer' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,compose_note,action}' = 'execute_llm_prompt'
      AND default_config #>> '{workflow,steps,append_note,action}'  = 'append_doc_note'
      AND default_config #>> '{workflow,steps,create_rerender,next_step}' = 'compose_note'
      AND default_config #>> '{workflow,steps,check_needs_rerender,config,else_step}' = 'compose_note'
      AND default_config #>> '{workflow,steps,append_note,config,subject_key}' = 'build'
      AND default_config #>> '{workflow,timeout_seconds}' = '240'
      AND NOT EXISTS (SELECT 1 FROM jsonb_each(default_config #> '{workflow,steps}') t(k,v) WHERE v ? 'error_step');
    IF n <> 1 THEN RAISE EXCEPTION 'component-template-fixer wiring incomplete (found %)', n; END IF;

    -- B: tool-improver
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-improver' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,compose_note,action}' = 'execute_llm_prompt'
      AND default_config #>> '{workflow,steps,append_note,action}'  = 'append_doc_note'
      AND default_config #>> '{workflow,steps,create_rerender_item,next_step}' = 'compose_note'
      AND default_config #>> '{workflow,steps,append_note,config,subject_key_field}' = 'tool_data.function'
      AND default_config #>> '{workflow,timeout_seconds}' = '480'
      AND NOT EXISTS (SELECT 1 FROM jsonb_each(default_config #> '{workflow,steps}') t(k,v) WHERE v ? 'error_step');
    IF n <> 1 THEN RAISE EXCEPTION 'tool-improver wiring incomplete (found %)', n; END IF;

    -- C/D/E: tool-recreation-handler
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,compose_note,action}' = 'execute_llm_prompt'
      AND default_config #>> '{workflow,steps,append_note,action}'  = 'append_doc_note'
      AND default_config #>> '{workflow,steps,deploy_page,next_step}' = 'compose_note'
      AND default_config #>> '{workflow,steps,deploy_page,config,error_step}' = 'complete_error'
      AND default_config #>> '{workflow,steps,validate_tool,config,error_step}' = 'save_sections'
      AND default_config #>> '{workflow,timeout_seconds}' = '2400'
      AND NOT EXISTS (SELECT 1 FROM jsonb_each(default_config #> '{workflow,steps}') t(k,v) WHERE v ? 'error_step');
    IF n <> 1 THEN RAISE EXCEPTION 'tool-recreation-handler wiring incomplete (found %)', n; END IF;
END $$;

COMMIT;

-- Verify after apply (one row per agent, all true):
--   SELECT type,
--          default_config #>> '{workflow,steps,compose_note,action}' = 'execute_llm_prompt' AS has_compose,
--          default_config #>> '{workflow,steps,append_note,action}'  = 'append_doc_note'   AS has_append,
--          default_config #>> '{workflow,timeout_seconds}'                                  AS timeout,
--          NOT EXISTS (SELECT 1 FROM jsonb_each(default_config #> '{workflow,steps}') t(k,v)
--                      WHERE v ? 'error_step')                                              AS no_step_level_error
--   FROM agent_definitions
--   WHERE type IN ('component-template-fixer','tool-improver','tool-recreation-handler')
--     AND deleted_at IS NULL
--   ORDER BY type;
--
-- Proof (after the next fix of any kind):
--   SELECT subject_type, subject_key, site_id, categories, source,
--          left(body,120) AS head, created_at
--   FROM doc_notes WHERE categories ? 'fix' ORDER BY created_at DESC LIMIT 5;
--
-- Rollback: restore all three from the snapshots taken at the top, or manually:
--   create_rerender.next_step -> "complete"; check_needs_rerender.config.else_step -> "complete";
--   create_rerender_item.next_step -> "complete"; deploy_page.next_step -> "complete";
--   delete the six compose_note/append_note steps; timeouts 240->120, 480->300.
--   The ten error_step corrections can be kept — strictly corrective.

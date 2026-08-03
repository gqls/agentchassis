-- 299_edit_live_channel_for_content_rewrite_writer.sql
--
-- bugs_open/178 — a content_rewrite item (e.g. "add a link to X") regenerates
-- the whole target section from scratch and silently drops most of its prose,
-- because there is no channel that hands page-content-writer a page's own
-- CURRENT stored content to edit. load_existing_content only fires for
-- mode="recreate" (adoption crawl), and even then it sources research_results,
-- never page_components — see load_existing_content_action.go:64-69 and its
-- own doc comment.
--
-- This migration wires the new "edit_live" channel into the one shared
-- page-build-handler workflow every content build already runs:
--
--   1. New step load_current_section_content, inserted between
--      check_has_ready_sections and spawn_content_writer. Read-only.
--      Reuses the "section_plan" output_field plan_sections itself uses, so
--      call_content_writer's existing input_mapping needs no change.
--   2. page-content-writer's per-section prompt gains a new conditional block
--      that only fires when a section carries existing_content_html — i.e.
--      only for an item that opted into edit_live.
--
-- Both changes are structural no-ops for every existing caller: no live item
-- sets spec.mode="edit_live" until the two emitters (create_tool_cross_link_items.go,
-- apply_gap_plan_action.go's applyAddToPage) are updated to set it, in a
-- separate commit. Default OFF per the 2026-08-02 owner ruling on shared-seam
-- authority.
--
-- Run AFTER deploying load_current_section_content_action.go + its registry
-- entry (registry.go). Image first, then this config — an unregistered
-- action name in a live step would error at runtime.
--
-- Verify: bugs_open/178's own test —
--   SELECT length(content_data::text) FROM page_components
--   WHERE page_id=<target> AND slot_name='generic-text-block';
--   -- expect: prior length + ~90 chars, NOT a wholesale replacement.

BEGIN;

-- ============================================================================
-- 1. page-build-handler: insert load_current_section_content
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_current_section_content}',
        '{
            "action": "load_current_section_content",
            "config": {
                "site_id": "site_record.site_id",
                "page_id": "page_record.id",
                "mode": "input_data.spec.mode"
            },
            "next_step": "spawn_content_writer",
            "error_step": "spawn_content_writer",
            "output_field": "section_plan",
            "description": "When spec.mode=edit_live, attach the current rendered_html for each ready section so the writer edits instead of regenerating (bugs_open/178)"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler'
  AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_has_ready_sections,config,then_step}',
        '"load_current_section_content"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'page-build-handler'
  AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify step 1, inside the transaction: a migration verify block of bare
-- SELECTs cannot stop the COMMIT (ON_ERROR_STOP ignores a non-empty result
-- set), so the check that matters is a DO/RAISE, not the SELECT below it.
DO $$
DECLARE
    then_step text;
    new_step_action text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'check_has_ready_sections'->'config'->>'then_step'
    INTO then_step
    FROM agent_definitions
    WHERE type = 'page-build-handler'
      AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    SELECT default_config->'workflow'->'steps'->'load_current_section_content'->>'action'
    INTO new_step_action
    FROM agent_definitions
    WHERE type = 'page-build-handler'
      AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF then_step IS DISTINCT FROM 'load_current_section_content' THEN
        RAISE EXCEPTION 'page-build-handler wiring failed: check_has_ready_sections.then_step = %, want load_current_section_content', then_step;
    END IF;
    IF new_step_action IS DISTINCT FROM 'load_current_section_content' THEN
        RAISE EXCEPTION 'page-build-handler wiring failed: load_current_section_content step missing or wrong action (%)', new_step_action;
    END IF;
END $$;

SELECT
    default_config->'workflow'->'steps'->'check_has_ready_sections'->'config'->>'then_step' AS then_step,
    default_config->'workflow'->'steps'->'load_current_section_content'->>'next_step' AS new_step_next,
    default_config->'workflow'->'steps'->'load_current_section_content'->>'output_field' AS new_step_output_field
FROM agent_definitions
WHERE type = 'page-build-handler' AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
-- Expected: then_step=load_current_section_content, new_step_next=spawn_content_writer,
--           new_step_output_field=section_plan

-- ============================================================================
-- 2. page-content-writer: prompt gains the edit-mode block
-- ============================================================================
-- Anchored on the exact, verified-unique byte sequence immediately preceding
-- the Admin Content Brief block (confirmed occurrences=1 against the live
-- row before writing this migration). The inserted block is gated on
-- {{if .current_section.existing_content_html}}, which is empty for every
-- section unless load_current_section_content (above) populated it — i.e.
-- inert for every page that is not an edit_live content_rewrite.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(
            replace(
                default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
                E'{{end}}\n\n{{if .current_section.component.content_brief}}',
                E'{{end}}\n\n{{if .current_section.existing_content_html}}## EXISTING CONTENT FOR THIS SECTION (Edit Mode, bugs_open/178)\nThis section already exists on the live page. Below is its current rendered content.\nYour task: apply the Rewrite Guidance above by EDITING this content, not by replacing it.\nPreserve the existing prose, structure and information. Change only what the guidance asks for.\nDo NOT discard, shorten, or regenerate unrelated material, and do NOT invent a fresh replacement section.\n\nCurrent section content:\n{{.current_section.existing_content_html}}\n{{end}}\n\n{{if .current_section.component.content_brief}}'
            )
        )
                     ),
    updated_at = NOW()
WHERE type = 'page-content-writer'
  AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify step 2: RAISE, not a bare SELECT, so a silent no-op (anchor text
-- didn't match — e.g. whitespace drift since this was written) aborts the
-- transaction instead of committing a "fix" that never applied.
DO $$
DECLARE
    tpl text;
BEGIN
    SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO tpl
    FROM agent_definitions
    WHERE type = 'page-content-writer'
      AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF tpl IS NULL OR position('existing_content_html' in tpl) = 0 THEN
        RAISE EXCEPTION 'page-content-writer prompt patch failed: existing_content_html marker not found — the replace() anchor did not match, template unchanged';
    END IF;
    IF (length(tpl) - length(replace(tpl, 'existing_content_html', ''))) / length('existing_content_html') <> 2 THEN
        RAISE EXCEPTION 'page-content-writer prompt patch produced an unexpected occurrence count (want 2: the {{if}} guard and the {{.current_section.existing_content_html}} field reference)';
    END IF;
END $$;

SELECT
    position('existing_content_html' in default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}') > 0 AS patch_present
FROM agent_definitions
WHERE type = 'page-content-writer' AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
-- Expected: patch_present = t

COMMIT;

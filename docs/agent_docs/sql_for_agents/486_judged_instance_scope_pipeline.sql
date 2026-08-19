-- 486 — component-template-fixer grows the JUDGED branch (bugs_open/283, RFC_034 shape C;
--       design bugfix_283_component_instance_scope/PLAN_2026-08-18_judged_pipeline.md §3).
--
-- ⚠ _HOLD: ORDERING-CRITICAL. The steps below dispatch fix_type
-- 'scope_component_instance_judged', which exists only in chassis builds carrying the
-- 2026-08-19 code. Apply ONLY after RUNBOOK §1 digest-verifies the roll and
-- `git merge-base --is-ancestor <the 283 judged commit> <pod revision>` says LIVE.
-- Rename away from _HOLD when applying.
--
-- What it adds, per the design:
--   apply_fix (unchanged Go, result now carries the ids-converted template + handler
--   inventory) → check_scope_route: a needs_script_scoping refusal branches to
--   scope_script_llm (sonnet, 32k, NO tolerate_truncation — a capped completion fails the
--   step) → apply_judged_write (gate+write fused in Go; refusals return fixed:false with the
--   failing checks named) → check_judged_result: fixed:true continues to the rerender chain,
--   anything else → judged_refusal (fail_work_item → needs_human_review). create_rerender
--   (generic pages, mig 460/462) now chains into create_section_edit_delivery: one
--   section_edit item per OWNED placement (the sanctioned owned-page path; apply_section_edit
--   binds InstanceID — section_editor_actions.go:850/:948), closing §13.6's "owned pages
--   deliver nothing" gap for every future template fix.
--
-- Verify block PREPARE-compiles the embedded delivery query (the 460→461 lesson).
BEGIN;

-- Double-application guard (council round 7, debug_historian): _HOLD files are not
-- ledger-recorded, so a hand-reapply is possible — and it would take a SECOND
-- snapshot_agent labelled "pre-image" of an already-patched row, poisoning any
-- newest-first rollback. Abort BEFORE the snapshot.
DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg IS NULL THEN
    RAISE EXCEPTION '486: no active component-template-fixer row';
  END IF;
  IF (cfg #> '{workflow,steps}') ? 'check_scope_route' THEN
    RAISE EXCEPTION '486: already applied (check_scope_route present) — a second apply would snapshot the patched row as a "pre-image". Nothing to do.';
  END IF;
END $$;

SELECT snapshot_agent('component-template-fixer', '486 pre-image: judged branch + owned-page section_edit delivery');

-- 1. The six new steps.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps}',
      (default_config #> '{workflow,steps}') || $STEPS${
  "check_scope_route": {
    "action": "conditional",
    "config": {
      "condition": "fix_result.action == needs_script_scoping",
      "then_step": "scope_script_llm",
      "else_step": "check_needs_rerender"
    },
    "description": "Route the judged pool (script declares into global scope, or bindings the mechanical passes could not place) to the LLM branch; every other outcome continues exactly as before"
  },
  "scope_script_llm": {
    "action": "execute_llm_prompt",
    "config": {
      "ai_service": {
        "model": "claude-sonnet-5",
        "provider": "anthropic",
        "max_tokens": 32000,
        "api_key_env_var": "ANTHROPIC_API_KEY"
      },
      "input_fields": ["fix_result"],
      "output_format": "text",
      "prompt_template": "You are completing the per-instance scope conversion of an interactive website component.\n\nThe element ids in this template have ALREADY been renamed mechanically: every id attribute now begins with a render-time placeholder prefix (you can see it in the id attributes; it ends with a hyphen). Your job is ONLY the script surgery. Do not rename ids, do not edit markup except as rule 2 licenses, do not reformat, do not improve copy or CSS.\n\nComponent function: {{.fix_result.function}}\n\n## Template (your input — ids already converted)\n{{.fix_result.converted_template}}\n\n## Inline event-handler attributes found in the markup ({{.fix_result.inline_handler_n}})\n{{.fix_result.inline_handlers}}\n\n## Bindings the mechanical pass could not place\n{{.fix_result.unprefixed_bindings}}\n\n## Rules\n1. Wrap each inline script body in an IIFE: (function () { 'use strict'; ... })(); — if the script opens with a tool-doc comment block, keep that comment at the top of the script element, ABOVE the IIFE, unchanged.\n2. Remove every inline on*= attribute listed above from the markup, and re-create each binding inside the IIFE with addEventListener on the same element, looking it up by its converted id copied exactly as it appears in the markup, placeholder prefix included.\n3. Replace every window.onload assignment with a DOMContentLoaded listener registered inside the IIFE.\n4. For each item under 'Bindings the mechanical pass could not place': rework that code inside the IIFE so the lookup or key resolves to the converted (prefixed) form at runtime. Copy the exact placeholder prefix text from an existing id attribute in the template; never invent, alter or partially spell the placeholder.\n5. Change NOTHING else. Every id attribute, every piece of text, every style rule and every placeholder token must survive byte-for-byte apart from the removed on*= attributes and the script bodies.\n6. Output ONLY the full converted template. No markdown fences. No explanation. Start directly with the first character of the template.",
      "error_step": "judged_refusal"
    },
    "next_step": "apply_judged_write",
    "error_step": "judged_refusal",
    "description": "LLM performs the judged script surgery on the ids-converted template. tolerate_truncation is DELIBERATELY absent: a capped completion fails the step (bugs_open/012 class, first layer).",
    "output_field": "scoped_script"
  },
  "apply_judged_write": {
    "action": "fix_component_template",
    "config": {
      "fix_type": "scope_component_instance_judged",
      "site_id": "site_record.site_id",
      "html_field": "scoped_script.result"
    },
    "next_step": "check_judged_result",
    "error_step": "judged_refusal",
    "description": "Gate the rewrite mechanically (two-instance render fully clean, markup parity outside scripts, id-set parity, no surviving bindings, comparative write guard), then snapshot + write. Refusals write nothing.",
    "output_field": "fix_result"
  },
  "check_judged_result": {
    "action": "conditional",
    "config": {
      "condition": "fix_result.fixed == true",
      "then_step": "check_needs_rerender",
      "else_step": "judged_refusal"
    },
    "description": "A judged refusal must fail the ITEM for a human — it must never complete as a quiet no-op"
  },
  "judged_refusal": {
    "action": "fail_work_item",
    "config": {
      "work_item_id": "input_data.work_item_id",
      "error_message": "judged instance-scope conversion refused: the rewrite failed the gate (or the LLM step failed); the component was left untouched — see fix_result.gate_issues / reason on the orchestration state",
      "status_override": "needs_human_review"
    },
    "next_step": "compose_note",
    "error_step": "compose_note",
    "description": "Fail the item to needs_human_review; the note still gets written (docs never fail the fix)",
    "output_field": "refusal"
  },
  "create_section_edit_delivery": {
    "action": "query_database",
    "config": {
      "query": "INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, priority, handler_agent, status, created_by, spec, item_key) SELECT p.site_id, 'side_effect', 'build', 'section_edit', 'medium', 'Deliver template fix to owned page via section-editor: ' || p.name, 60, 'section-editor', 'triaged', 'component-template-fixer', jsonb_build_object('edit_type','content_edit','field_updates','{}'::jsonb,'page_name',p.name,'slot_name',pc.slot_name,'page_id',p.id::text,'component_id',pc.component_id::text,'reason','template_changed'), 'section_edit_tplfix_' || p.id::text FROM page_components pc JOIN pages p ON p.id = pc.page_id WHERE pc.component_id = $1::uuid AND p.rebuild_policy = 'owned' AND NOT EXISTS (SELECT 1 FROM site_work_items w WHERE w.site_id = p.site_id AND w.item_key = 'section_edit_tplfix_' || p.id::text AND w.status IN ('detected','triaged','claimed'))",
      "params": ["fix_result.component_id"]
    },
    "next_step": "compose_note",
    "error_step": "compose_note",
    "description": "Owned pages take a template fix via the section-editor (apply_section_edit binds InstanceID); generic pages were already covered by create_rerender. Closes 283 §13.6.",
    "output_field": "section_edit_created"
  }
}$STEPS$::jsonb
    ),
    updated_at = now()
WHERE type = 'component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps}') ? 'check_scope_route'
  AND default_config #>> '{workflow,steps,apply_fix,next_step}' = 'check_needs_rerender';

-- 2. Rewire apply_fix and create_rerender (guarded on the pre-image values).
UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(default_config, '{workflow,steps,apply_fix,next_step}', '"check_scope_route"'),
      '{workflow,steps,create_rerender,next_step}', '"create_section_edit_delivery"'),
    updated_at = now()
WHERE type = 'component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND (default_config #> '{workflow,steps}') ? 'check_scope_route'
  AND default_config #>> '{workflow,steps,apply_fix,next_step}' = 'check_needs_rerender'
  AND default_config #>> '{workflow,steps,create_rerender,next_step}' = 'compose_note';

DO $$
DECLARE cfg jsonb; q text; prompt text;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF NOT (cfg #> '{workflow,steps}') ?& array['check_scope_route','scope_script_llm','apply_judged_write','check_judged_result','judged_refusal','create_section_edit_delivery'] THEN
    RAISE EXCEPTION '486: a new step is missing';
  END IF;
  IF cfg #>> '{workflow,steps,apply_fix,next_step}' <> 'check_scope_route' THEN
    RAISE EXCEPTION '486: apply_fix does not route into check_scope_route';
  END IF;
  IF cfg #>> '{workflow,steps,create_rerender,next_step}' <> 'create_section_edit_delivery' THEN
    RAISE EXCEPTION '486: create_rerender does not chain into the owned-page delivery';
  END IF;
  IF (cfg #>> '{workflow,steps,scope_script_llm,config,ai_service,max_tokens}')::int <> 32000 THEN
    RAISE EXCEPTION '486: scope_script_llm max_tokens is not 32000';
  END IF;
  IF (cfg #> '{workflow,steps,scope_script_llm,config}') ? 'tolerate_truncation' THEN
    RAISE EXCEPTION '486: tolerate_truncation must NOT be set on the judged LLM step';
  END IF;
  prompt := cfg #>> '{workflow,steps,scope_script_llm,config,prompt_template}';
  IF prompt IS NULL OR length(prompt) < 500 OR prompt NOT LIKE '%converted_template%' THEN
    RAISE EXCEPTION '486: judged prompt missing or does not reference the converted template';
  END IF;
  q := cfg #>> '{workflow,steps,create_section_edit_delivery,config,query}';
  EXECUTE 'PREPARE chk486 (uuid) AS ' || q;
  DEALLOCATE chk486;
END $$;

COMMIT;

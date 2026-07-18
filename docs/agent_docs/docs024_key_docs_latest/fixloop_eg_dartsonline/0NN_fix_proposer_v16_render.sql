-- v16 (2026-07-18): adds council reviewer review_render_guardian (candidate #7),
-- GATED behind the relevance filter, added SURGICALLY (chained jsonb_set on the
-- live config — never a full-config reapply; preserves the co-edited guardian's
-- code_checks + stability proviso). 
-- Chain: ... -> gate_compliance -> [review_compliance?] -> gate_render -> [review_render_guardian?] -> review_guardian

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: v16 — render-pipeline guardian (gated, surgical)')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='fix-proposer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

UPDATE agent_definitions d
SET default_config =
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
    d.default_config,
    '{workflow,steps,select_panel,config,footprints,render}',
    jsonb_build_array('rerender','render_','assemble','sectionhasvisiblecontent','runtime-fill','rendered_html','page_components','css','styling','content_components')
  ),
    '{workflow,steps,review_render_guardian}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'Council reviewer (gated, candidate #7) -- render-pipeline guardian. Advisory only (no veto). Runs only when select_panel flags the fix as relevant.',
      'output_field', 'review_render_guardian',
      'next_step', 'review_guardian',
      'config', jsonb_build_object(
        'error_step', 'complete_refused',
        'ai_service', jsonb_build_object('model','claude-sonnet-4-6','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',3000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('diagnosis_row','plan_persisted','schema_hint'),
        'output_format', 'json',
        'prompt_template',
'# Council reviewer: RENDER-PIPELINE GUARDIAN' || chr(10) ||
'' || chr(10) ||
'You judge one thing: does this fix respect the rendering/assembly pipeline''s contracts -- the layer where most of this platform''s silent, visually-invisible bugs live? You change nothing; you judge.' || chr(10) ||
'' || chr(10) ||
'## The render pipeline''s load-bearing contracts' || chr(10) ||
'- FAIL LOUD, NOT SILENT: a render path must never silently drop or blank content. The escalate-not-blank guard and the required-field refusal exist for exactly this; a new render path needs the same posture.' || chr(10) ||
'- TWO RERENDER MODES, DIFFERENT SKIP SEMANTICS: scoped mode (spec.reason = image_landed / section_data_resolved) REGENERATES section HTML from content_data but SKIPS pages whose content hash is unchanged -- silently wrong for header/footer/chrome-only changes; assemble mode (page_id, no reason) re-embeds chrome unconditionally. A fix routing chrome changes through scoped mode will silently miss pages.' || chr(10) ||
'- INTENTIONALLY-EMPTY IS NOT EMPTY: the assembler drops sections with <=10 visible chars, EXCEPT those marked data-runtime-fill (interactive shells filled client-side). A fix touching the filter must preserve the exemption; and runtime-fill templates are RENDERED ARTIFACTS -- editing the rendered output instead of the source template is a documented landmine.' || chr(10) ||
'- CSS INHERITANCE: element colours resolve via the two-level var() fallback chain (--section-*, then --color-*). Hardcoding colours or breaking the chain is the platform''s single most important styling rule violated.' || chr(10) ||
'- VALIDATION LAYERS: store-time rejection, plan-time deferral, and render-time filtering are three separate defence layers -- removing one because the others exist is how layers die.' || chr(10) || chr(10) ||
'Judge the plan: (a) does any edit add a render path that can silently drop/blank/skip content instead of failing loud or escalating; (b) does it route chrome/header/footer changes through scoped rerender (hash-skip) instead of assemble mode; (c) does it break the data-runtime-fill exemption or edit rendered artifacts instead of source templates; (d) does it hardcode colours / bypass the var() inheritance chain, or remove a validation layer. If the fix does not touch rendering/assembly/styling, approve.' || chr(10) || chr(10) ||
'Verdicts: approve (contracts intact, or the fix does not touch this area), object (names the specific contract/lesson violated, in objections). You do NOT have a veto -- put a severe concern in objections at "high" severity and trust the router; note a true architecture-level concern explicitly.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle, put it in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "render_guardian", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
      )
    )
  ),
    '{workflow,steps,gate_render}',
    jsonb_build_object(
      'action', 'conditional',
      'description', 'Relevance gate for review_render_guardian; skip = abstain.',
      'config', jsonb_build_object(
        'condition', 'panel.run_render == true',
        'then_step', 'review_render_guardian',
        'else_step', 'review_guardian'
      )
    )
  ),
    '{workflow,steps,review_compliance,next_step}', '"gate_render"'::jsonb
  ),
    '{workflow,steps,gate_compliance,config,else_step}', '"gate_render"'::jsonb
  ),
    '{workflow,steps,council_decide,config,review_fields}',
    (d.default_config #> '{workflow,steps,council_decide,config,review_fields}') || '["review_render_guardian.result"]'::jsonb
  ),
    '{workflow,steps,escalate,config,review_fields}',
    (d.default_config #> '{workflow,steps,escalate,config,review_fields}') || '["review_render_guardian.result"]'::jsonb
  ),
    '{workflow,steps,run_checks,config,check_fields}',
    (d.default_config #> '{workflow,steps,run_checks,config,check_fields}') || '["review_render_guardian.result.checks"]'::jsonb
  ),
  updated_at = now()
WHERE d.type='fix-proposer' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND NOT (d.default_config->'workflow'->'steps' ? 'review_render_guardian');

COMMIT;

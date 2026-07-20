-- SEED R0+R1 (2026-07-20, owner-approved plan
-- PLAN_2026-07-20_direction_reach_and_drift_guard.md): mission reach into the
-- build pipeline, on domain-research-classifier. SURGICAL (chained jsonb_set on
-- the live row; patch-style, never a full-config reapply -- the re-seed clobber
-- landmine). DB config is LIVE IMMEDIATELY on apply.
--
-- R0: the classifier finally SEES the platform mission -- a compact BIZ-001/doc-028
--     digest inserted into classify_and_extract's prompt, before its final
--     return-JSON instruction. Complement, not enforcement.
-- R1: OBSERVE-ONLY mission review after the specs are written:
--     write_design_intent_spec -> review_mission_alignment (LLM, error_step =
--     create_next_item: a failed review can NEVER block a build) ->
--     gate_mission_note (objection_found?) -> append_mission_note (doc_notes,
--     categories ['mission-review']) -> create_next_item.
--     CONSUMER (corrected from the plan): doc_notes + the 101 report script --
--     NOT site_work_items at 'detected': triage_detect_items_action.go:91-103
--     promotes ALL detected items site-wide into the dispatch pipeline, so a
--     detected mission_review row would be swept toward a nonexistent handler
--     (not observe-only). doc_notes cannot be dispatched; the council gate
--     already records verdicts there (reuse before recreate).

BEGIN;

SELECT snapshot_agent('domain-research-classifier', 'pre-update: R0+R1 mission reach (digest + observe-only review)')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='domain-research-classifier' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

UPDATE agent_definitions d
SET default_config =
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
    d.default_config,
    '{workflow,steps,classify_and_extract,config,prompt_template}',
    to_jsonb(replace(
      d.default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}',
      'Return ONLY valid JSON with identity, classification, content_direction, and design_intent keys.',
      '## The platform mission (fixed direction -- check yourself against it)
- Best possible site for THIS domain, end to end, minimal human input. "Best" = most useful to the real visitors this domain can attract AND the best revenue via whatever model genuinely fits the domain.
- THE REVENUE MODEL SHAPES THE SITE, not the other way round. Ad-supported tools/content, affiliate comparison, publication, SaaS, direct services -- each is a legitimate, completely different site. DEFAULTING to a consultancy/services shape when the signal is absent is a FAILURE MODE, not a safe fallback: name the revenue model and argue it from this domain''s evidence.
- You are the strategic brain. Adoption evidence and an operator mission WEIGHT your reasoning; they never replace research or shortcut your full job.
- Do not trim the vision to what can be built today: describe the best version of this site; items beyond current capability are recorded and marked, never silently dropped.

' || 'Return ONLY valid JSON with identity, classification, content_direction, and design_intent keys.'
    ))
  ),
    -- bugs_open/016: in a prompt TEMPLATE an input_fields value is already
    -- unwrapped, so it is '{{.analysis}}' — never '{{.analysis.result}}', which
    -- renders '<no value>' silently for a json-output producer. The '.result'
    -- form stays correct in CONFIG dot-paths: see gate_mission_note's condition
    -- 'mission_review.result.objection_found' below, which reads raw
    -- collected_data and must keep it.
    '{workflow,steps,review_mission_alignment}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'R1 OBSERVE-ONLY mission-alignment review of the just-written classification; records a doc_note on objection; never blocks the build (error routes to create_next_item).',
      'output_field', 'mission_review',
      'next_step', 'gate_mission_note',
      'config', jsonb_build_object(
        'error_step', 'create_next_item',
        'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',8000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('analysis','input_data'),
        'output_format', 'json',
        'prompt_template', '# Mission-alignment review (OBSERVE-ONLY)

You are reviewing a site classification that has JUST been written, against the platform mission. Your judgement is RECORDED, never enforced: the build proceeds unchanged whatever you say. Be strict about the mission but honest about evidence -- a consultancy/services shape is legitimate WHEN the domain''s evidence supports it; the mission bans the unargued DEFAULT, not the shape.

## The mission tests
1. REVENUE MODEL ARGUED, NOT DEFAULTED: does the classification name a revenue model and argue it from this domain''s evidence? A services/consultancy shape with no supporting signal is the classic failure.
2. NO SHAPE-MIXING: does the direction mix shapes (e.g. an ad-supported tools/content site carrying "hire us" / "Start a Project" service CTAs)?
3. NOTHING SILENTLY TRIMMED: does the output read like the BEST version of this site, with beyond-capability items marked rather than absent? (You can only judge from what is present -- flag a suspiciously conservative spec, do not guess.)
4. DIRECTION, NOT DRIFT: does anything in the classification push against the mission (best-site-per-domain, minimal human input, classifier-as-strategic-brain)?

## The classification just produced
{{.analysis}}

## Output -- ONLY this JSON
{"objection_found": true|false, "concerns": [{"test": 1, "problem": "..."}], "note": "MISSION-REVIEW <domain if known>: one readable paragraph a human can act on -- what drifted, which test, what the classifier should have done. Empty string when objection_found is false."}'
      )
    )
  ),
    '{workflow,steps,gate_mission_note}',
    jsonb_build_object(
      'action', 'conditional',
      'description', 'Route to the doc_note append only when the mission review objected; else straight on.',
      'config', jsonb_build_object(
        'condition', 'mission_review.result.objection_found == true',
        'then_step', 'append_mission_note',
        'else_step', 'create_next_item'
      )
    )
  ),
    '{workflow,steps,append_mission_note}',
    jsonb_build_object(
      'action', 'append_doc_note',
      'description', 'Persist the mission-review objection as a doc_note (categories mission-review); consumer = 101_REPORT_mission_review_findings.sh.',
      'next_step', 'create_next_item',
      'config', jsonb_build_object(
        'error_step', 'create_next_item',
        'subject_type', 'pipeline',
        'subject_key', 'mission-review',
        'note_body_field', 'mission_review.result.note',
        'note_categories', jsonb_build_array('mission-review'),
        'created_by', 'classifier-mission-review'
      )
    )
  ),
    '{workflow,steps,write_design_intent_spec,next_step}', '"review_mission_alignment"'::jsonb
  ),
  updated_at = now()
WHERE d.type='domain-research-classifier' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND NOT (d.default_config->'workflow'->'steps' ? 'review_mission_alignment')
  AND (d.default_config #>> '{workflow,steps,write_design_intent_spec,next_step}') = 'create_next_item'
  AND (d.default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}') LIKE '%Return ONLY valid JSON with identity, classification, content_direction, and design_intent keys.%'
  AND (d.default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}') NOT LIKE '%platform mission (fixed direction%';

COMMIT;

-- Verify:
--  1. steps present: SELECT default_config->'workflow'->'steps' ?& array['review_mission_alignment','gate_mission_note','append_mission_note'] ...
--  2. write_design_intent_spec.next_step = review_mission_alignment; its error_step stays create_next_item.
--  3. prompt carries the mission block once.
-- NOTE: write_design_intent_spec.error_step (= create_next_item) is left as-is:
-- a failed design-spec write skips the review -- acceptable for an observe lane.
-- Rollback: restore the snapshot from agent_definitions_backup.

-- 449 ROLLBACK — remove the experience-plan citation wiring from tool-suggester.
--
-- Surgical inverse of 449, scoped by id + pre-state gate (the shape migration
-- 406's council round asked for). It restores the chain, drops the step, drops
-- the input field, and reverses the three prompt splices by the SAME anchors —
-- the added text is removed, the anchors it was attached to are left standing.
--
-- If this file cannot restore the prompt cleanly (because a later migration
-- edited the same template), STOP and restore the snapshot 449 took instead:
--   SELECT revert_agent('tool-suggester');  -- see snapshot_agent note
--     '449_tool_suggester_cites_approved_experience_plans: pre-update'

BEGIN;

DO $$
DECLARE
  n int; cfg jsonb;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '449 rollback: expected exactly 1 live tool-suggester row, found %', n; END IF;

  SELECT default_config->'workflow'->'steps' INTO cfg FROM agent_definitions
   WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF NOT (cfg ? 'load_experience_plans') THEN
    RAISE EXCEPTION '449 rollback: load_experience_plans is not present — 449 is not applied here';
  END IF;
  IF cfg#>>'{load_library_tools,next_step}' <> 'load_experience_plans' THEN
    RAISE EXCEPTION '449 rollback: chain is %, not the state 449 leaves — a later migration has moved it', cfg#>>'{load_library_tools,next_step}';
  END IF;
END $$;

-- 1. chain back to direct, then drop the step
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config, '{workflow,steps,load_library_tools,next_step}', to_jsonb('suggest_tools'::text)) #- '{workflow,steps,load_experience_plans}',
       updated_at = now()
 WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 2. drop the input field (rebuilt without it, order preserved)
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,suggest_tools,config,input_fields}',
         (SELECT COALESCE(jsonb_agg(f ORDER BY ord), '[]'::jsonb)
            FROM jsonb_array_elements_text(default_config#>'{workflow,steps,suggest_tools,config,input_fields}') WITH ORDINALITY AS t(f, ord)
           WHERE f <> 'experience_plans')
       ),
       updated_at = now()
 WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 3. reverse the three prompt splices, by the same anchors
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,suggest_tools,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               replace(
                 default_config#>>'{workflow,steps,suggest_tools,config,prompt_template}',
                 E'\n## Approved Experience Plans (council-reviewed)\n{{if .experience_plans}}{{range .experience_plans}}- {{.subject_key}}: {{.plan_digest}}\n{{end}}{{else}}None on file.{{end}}\n\n## Your Task\n',
                 E'\n## Your Task\n'
               ),
               E'\n6. If an approved EXPERIENCE_PLAN above covers the experience a suggested tool would deliver, set experience_plan to that plan''s subject_key and follow the constraints it states. That plan settled the journey, safety and fallback questions for this experience through council review, so do not re-reason them here. If no approved plan covers the tool, set experience_plan to null.\n\nExamples of GOOD suggestions:',
               E'\nExamples of GOOD suggestions:'
             ),
             E'"complexity": "simple",\n      "experience_plan": "subject_key of an approved plan above, or null",',
             '"complexity": "simple",'
           )
         )
       ),
       updated_at = now()
 WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
  cfg jsonb; prompt text;
BEGIN
  SELECT default_config->'workflow'->'steps' INTO cfg FROM agent_definitions
   WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg ? 'load_experience_plans' THEN RAISE EXCEPTION '449 rollback verify: step still present'; END IF;
  IF cfg#>>'{load_library_tools,next_step}' <> 'suggest_tools' THEN RAISE EXCEPTION '449 rollback verify: chain not restored'; END IF;
  IF cfg#>'{suggest_tools,config,input_fields}' ? 'experience_plans' THEN RAISE EXCEPTION '449 rollback verify: input field still present'; END IF;
  prompt := cfg#>>'{suggest_tools,config,prompt_template}';
  IF position('experience_plan' in prompt) > 0 THEN
    RAISE EXCEPTION '449 rollback verify: prompt still mentions experience_plan — restore the snapshot instead';
  END IF;
  IF position('## Your Task' in prompt) = 0 OR position('Examples of GOOD suggestions:' in prompt) = 0 THEN
    RAISE EXCEPTION '449 rollback verify: an anchor was consumed';
  END IF;
END $$;

COMMIT;

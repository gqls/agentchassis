-- 449 — tool-suggester cites the council-approved EXPERIENCE_PLAN
-- (PLAN_2026-08-11_chat_box_as_framework_capability §3/§6, step 6 — the
--  webdesign_uk_build_service lane; TL-043/VMB-010 are step 5's half)
--
-- WHY. PLAN §3 splits two questions deliberately: "what is a safe, correct
-- site chat experience, in general?" is asked ONCE via experience-planner and
-- answered by a council-reviewed EXPERIENCE_PLAN (honesty seat has a hard
-- veto); "should THIS site get one?" is asked per site by tool-suggester.
-- Today the second question is answered from first principles every run — the
-- suggester cannot see that the first question has already been settled, so
-- the plan's journeys, promise ledger and fail-closed contract inform nothing
-- at suggestion time. Step 6 closes that: the suggester is handed the current
-- approved experience plans and asked to CITE one rather than re-reason it.
--
-- Ordering (PLAN's own condition): step 6 lands only after step 5 proved
-- deployment works, so the suggestion path does not outrun delivery. Step 5's
-- deployer half is TL-043 (council APPROVED, corr 55cda19b, live v1.0.1305 via
-- 6a782274b); its box half is the shared `sitechat` binary, proven 2026-08-16.
--
-- WHAT IT DOES. Three surgical edits to the one live tool-suggester row:
--   1. a new `load_experience_plans` step (query_database, NO params — the
--      build-site-planner.load_styles idiom) spliced between
--      load_library_tools and suggest_tools;
--   2. `experience_plans` added to suggest_tools' input_fields — WITHOUT THIS
--      THE TEMPLATE VARIABLE RENDERS EMPTY AND NOTHING ERRORS (the filed
--      landmine this migration's verify block exists to make impossible);
--   3. the prompt gains an "Approved Experience Plans" section, one rule, and
--      an `experience_plan` key in the suggestion schema.
--
-- The prompt is edited by ANCHORED replace() on the live text, never retyped:
-- the template is 3,471 bytes of another lane's reviewed copy and a retype is
-- how a migration silently drops a paragraph. Each anchor is asserted to occur
-- exactly once BEFORE the update, and the pre-state is asserted too, so a
-- concurrent edit by another session aborts this transaction rather than
-- landing on top of it.
--
-- SCOPE / BLAST RADIUS. Additive and inert by construction:
--   * `experience_plan` on a suggestion has NO automated consumer today. It
--     rides into the add_tool item's spec (spec_data: current_suggestion) where
--     tool-deployer can read it later. Producer set + shape named in the
--     concept register in the same commit (owner ruling 2026-08-02 §1).
--   * With zero current experience plans the new section renders "None on
--     file." and the prompt is behaviourally what it was.
--   * The digest is `left(body, 600)` — BOUNDED ON PURPOSE. The three current
--     plans are 10,075 / 11,152 / 13,971 chars; injecting them whole would add
--     ~35KB to every tool-suggester call fleet-wide for a question that only
--     needs "which experience is already settled, and what is it". The cost of
--     the bound is stated honestly: the suggester sees each plan's opening
--     (title + first journey), not its promise ledger. If a seat needs the
--     ledger, that is a follow-up that should extract a named section, not a
--     bigger left().
--
-- KNOWN, NOT TOUCHED: the live prompt carries a malformed rule ("5. 6. For
-- each suggestion…") and literal backslash-n sequences from an earlier
-- migration. Both are pre-existing defects in another lane's copy; fixing them
-- here would be an unrelated behaviour change riding inside this one.
--
-- Config-only: no image dependency, LIVE ON APPLY.
-- ROLLBACK: 449_..._ROLLBACK.sql, or the snapshot this file takes
-- (snapshot_agent note '449_tool_suggester_cites_approved_experience_plans: pre-update').

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '449_tool_suggester_cites_approved_experience_plans: pre-update');

-- ── PRE-STATE GUARDS ────────────────────────────────────────────────────────
-- Every one of these aborts the whole transaction (snapshot included), so a
-- row another session has already edited is never written over.
DO $$
DECLARE
  n int; cfg jsonb; prompt text; fields jsonb;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '449: expected exactly 1 live tool-suggester row, found %', n; END IF;

  SELECT default_config->'workflow'->'steps' INTO cfg FROM agent_definitions
   WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF cfg ? 'load_experience_plans' THEN
    RAISE EXCEPTION '449: load_experience_plans already exists — already applied, or another session got there first';
  END IF;
  IF cfg#>>'{load_library_tools,next_step}' <> 'suggest_tools' THEN
    RAISE EXCEPTION '449: load_library_tools.next_step is %, expected suggest_tools — the chain has moved', cfg#>>'{load_library_tools,next_step}';
  END IF;

  fields := cfg#>'{suggest_tools,config,input_fields}';
  IF fields IS NULL OR fields ? 'experience_plans' THEN
    RAISE EXCEPTION '449: suggest_tools input_fields missing or already carries experience_plans: %', fields;
  END IF;

  prompt := cfg#>>'{suggest_tools,config,prompt_template}';
  IF prompt IS NULL THEN RAISE EXCEPTION '449: suggest_tools has no prompt_template'; END IF;
  -- Anchors must each occur EXACTLY ONCE, or replace() would edit the wrong
  -- place (or several). Counted by length arithmetic — no regex dialect risk.
  IF (length(prompt) - length(replace(prompt, E'\n## Your Task\n', ''))) / length(E'\n## Your Task\n') <> 1 THEN
    RAISE EXCEPTION '449: anchor "## Your Task" does not occur exactly once';
  END IF;
  IF (length(prompt) - length(replace(prompt, E'\nExamples of GOOD suggestions:', ''))) / length(E'\nExamples of GOOD suggestions:') <> 1 THEN
    RAISE EXCEPTION '449: anchor "Examples of GOOD suggestions:" does not occur exactly once';
  END IF;
  IF (length(prompt) - length(replace(prompt, '"complexity": "simple",', ''))) / length('"complexity": "simple",') <> 1 THEN
    RAISE EXCEPTION '449: anchor "complexity" does not occur exactly once';
  END IF;
END $$;

-- ── 1. the new loader step + rewire the chain ───────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,load_experience_plans}',
           jsonb_build_object(
             'action', 'query_database',
             'config', jsonb_build_object(
               'query', 'SELECT subject_key, left(body, 600) AS plan_digest FROM doc_plans WHERE subject_type = ''experience'' AND is_current ORDER BY subject_key',
               'output_format', 'array'
             ),
             'next_step', 'suggest_tools',
             'description', 'Load council-approved experience plans so suggestions cite them instead of re-reasoning settled safety questions',
             'output_field', 'experience_plans'
           )
         ),
         '{workflow,steps,load_library_tools,next_step}',
         to_jsonb('load_experience_plans'::text)
       ),
       updated_at = now()
 WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── 2. input_fields — the half that makes the template variable non-empty ───
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,suggest_tools,config,input_fields}',
         (default_config#>'{workflow,steps,suggest_tools,config,input_fields}') || to_jsonb('experience_plans'::text)
       ),
       updated_at = now()
 WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── 3. the prompt, by anchored replace ──────────────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,suggest_tools,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               replace(
                 default_config#>>'{workflow,steps,suggest_tools,config,prompt_template}',
                 E'\n## Your Task\n',
                 E'\n## Approved Experience Plans (council-reviewed)\n{{if .experience_plans}}{{range .experience_plans}}- {{.subject_key}}: {{.plan_digest}}\n{{end}}{{else}}None on file.{{end}}\n\n## Your Task\n'
               ),
               E'\nExamples of GOOD suggestions:',
               E'\n6. If an approved EXPERIENCE_PLAN above covers the experience a suggested tool would deliver, set experience_plan to that plan''s subject_key and follow the constraints it states. That plan settled the journey, safety and fallback questions for this experience through council review, so do not re-reason them here. If no approved plan covers the tool, set experience_plan to null.\n\nExamples of GOOD suggestions:'
             ),
             '"complexity": "simple",',
             E'"complexity": "simple",\n      "experience_plan": "subject_key of an approved plan above, or null",'
           )
         )
       ),
       updated_at = now()
 WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── POST-STATE VERIFY (DO/RAISE: a block of SELECTs cannot stop a COMMIT) ───
DO $$
DECLARE
  cfg jsonb; prompt text;
BEGIN
  SELECT default_config->'workflow'->'steps' INTO cfg FROM agent_definitions
   WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF cfg#>>'{load_experience_plans,output_field}' <> 'experience_plans' THEN
    RAISE EXCEPTION '449 verify: load_experience_plans.output_field wrong: %', cfg#>>'{load_experience_plans,output_field}';
  END IF;
  IF cfg#>>'{load_experience_plans,next_step}' <> 'suggest_tools' THEN
    RAISE EXCEPTION '449 verify: load_experience_plans does not hand on to suggest_tools';
  END IF;
  IF cfg#>>'{load_library_tools,next_step}' <> 'load_experience_plans' THEN
    RAISE EXCEPTION '449 verify: the chain was not rewired: %', cfg#>>'{load_library_tools,next_step}';
  END IF;
  IF NOT (cfg#>'{suggest_tools,config,input_fields}' ? 'experience_plans') THEN
    RAISE EXCEPTION '449 verify: input_fields does not carry experience_plans — the template variable would render EMPTY and error nothing';
  END IF;

  prompt := cfg#>>'{suggest_tools,config,prompt_template}';
  IF position('{{range .experience_plans}}' in prompt) = 0 THEN
    RAISE EXCEPTION '449 verify: the plans section did not land in the prompt';
  END IF;
  IF position('"experience_plan": "subject_key' in prompt) = 0 THEN
    RAISE EXCEPTION '449 verify: the output schema did not gain experience_plan';
  END IF;
  IF position('set experience_plan to that plan''s subject_key' in prompt) = 0 THEN
    RAISE EXCEPTION '449 verify: the citation rule did not land in the prompt';
  END IF;
  -- The three anchors' own content must SURVIVE — a replace that ate its
  -- anchor would pass every check above while destroying the prompt.
  IF position('## Your Task' in prompt) = 0
     OR position('Examples of GOOD suggestions:' in prompt) = 0
     OR position('"complexity": "simple",' in prompt) = 0 THEN
    RAISE EXCEPTION '449 verify: an anchor was consumed rather than extended';
  END IF;
END $$;

COMMIT;

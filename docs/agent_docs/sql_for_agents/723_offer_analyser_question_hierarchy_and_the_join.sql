-- 723 — the visitor's QUESTION HIERARCHY, and the join that makes it checkable.
--
-- Owner Decision D, authorised directly 2026-09-02 (this lane declined the same
-- ruling twice while it reached it only by relay).
--
-- WHAT IT ADDS. offer-analyser already derives `lead_with` — a ranked set of
-- benefit statements. This adds `question_hierarchy`: the visitor's likely
-- doubts, in the order they arise, each joined to the lead_with point that
-- answers it or explicitly marked `unanswered`.
--
-- WHY, AND THE MEASUREMENT THAT DECIDED THE SHAPE. The owner rejected a hero
-- reading "No vendor pays us, so the choice is made on fit alone" — true,
-- strongly put, and far down the visitor's list. That is not a specimen defect:
-- [MEASURED 2026-08-31] the share of lead_with points marked `differentiated`
-- by rank runs 100/100/97/61/31/30 across 186 points. **The ordering key is
-- DIFFERENTIATION — "what can we say that competitors cannot" — which is a
-- SELLER'S question.** The owner's is a buyer's question: what will this get me,
-- and how much work is it.
--
-- ⚠ AND THE GAP IS ABSENCE, NOT ORDER, WHICH IS WHY THIS IS A NEW DERIVATION
-- AND NOT A RE-RANK. [MEASURED 2026-08-31] only 19 of 186 points (10%) address
-- effort or practicality at all. Re-ranking cannot surface material that was
-- never derived. A prompt migration that merely re-sorted would produce the same
-- seller-axis list backwards. (⚠ A later re-measure read 30% and is NOT
-- comparable — the regex was widened. A different instrument is not a different
-- result. Do not quote an improvement between those two numbers.)
--
-- ⚠ THE JOIN IS THE DELIVERABLE, NOT THE LIST. A hierarchy with no link to the
-- copy would be the THIRD provenance-stamped artefact nobody reads — this lane
-- has just measured its own `offer_ordering` at 32 sites with ZERO writer
-- consumers, and `lead_with` itself is read by nothing that writes a page.
-- `answered_by` is what turns "we think they ask X first" into a checkable claim
-- about a specific hero.
--
-- ⚠ PRE-REGISTERED ACCEPTANCE CRITERION: the first pass should come back MOSTLY
-- `unanswered` AT THE TOP. That is the finding, not a failure — it is the 10%
-- measurement restated per site. A first pass that answers everything means the
-- model stretched points to cover questions, which is the failure mode.
--
-- BOUNDARY (owner ruling 12, and identical on both lanes): modelling the
-- hierarchy to ORDER content is licensed — he does it himself in the
-- instruction. ASSERTING it in served copy is presumption and stays banned.
-- This is UNSERVED RATIONALE, structured input only, and must never be rendered
-- into a prompt as prose: "most visitors first ask X" sitting in a writer's
-- context window IS the presumption shape and will be copied.
--
-- ⚠ IT GOES THROUGH THE GATE — the whole point, and a second step rather than a
-- Go change. `question_hierarchy` is newly derived model text, which is exactly
-- what was minting at 23% while a migration washed the corpus behind it. A
-- hierarchy written around `repair_ordering_register` would rebuild that defect
-- one field along. The gate takes ONE items_key per invocation, so this chains a
-- SECOND invocation over the new collection:
--     verify_ordering_cardinals
--       -> repair_ordering_register   (lead_with / point)
--       -> repair_hierarchy_register  (question_hierarchy / question)   [NEW]
--       -> write_offer_ordering       (now reads the second gate's output)
-- No Go change, no roll: DB config, live on apply.
--
-- ⚠ THE MERGE TRAP, WHICH THE PROMPT ALREADY WARNS ABOUT AND NOW APPLIES TO THIS
-- KEY TOO. `ordering` is deep-merged over whatever the site already holds, so a
-- key the model omits silently leaves the previous run's value standing and
-- looking current (the `bugs_open/327` class). `question_hierarchy` is therefore
-- REQUIRED on every run, and the prompt says so in the same breath as the others.
--
-- spec_version 1 -> 2.

BEGIN;

SELECT snapshot_agent('offer-analyser', '723_offer_analyser_question_hierarchy_and_the_join.sql: pre-update');

-- 1. Guidance + output contract on the analysis prompt.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,run_offer_analysis,config,prompt}',
      to_jsonb(
        replace(
          replace(
            default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt',
            'OUTPUT. Return ONE JSON object and nothing else',
            $g$THE QUESTION HIERARCHY. Before a visitor weighs anything you can say, they arrive with doubts, in an order. Rank them.

Write each question in the visitor's own terms, as the doubt they actually hold — "How long is this going to take me?", not "Time investment". Rank 1 is the FIRST doubt, not the one most important to us. For most sites that is some form of what will this actually get me and how much work is it to get it; the things we are proudest of — our independence, our method, our credentials — usually sit further down than we would like, and putting them there is the point of this exercise.

Derive each question from a named field of the strategy you were shown and put that field in "from_field". If you cannot ground a question in the premise, do not invent it.

THE JOIN IS THE POINT. For each question set "answered_by" to the rank of the lead_with point that genuinely answers it, or set "unanswered": true and "answered_by": null. Do NOT stretch a point to cover a question it does not answer, and do NOT add a lead_with point merely to close a gap. A hierarchy whose top entries are unanswered is the USEFUL result: it states what this site does not yet tell its visitor, which is the thing nobody can currently see.

This hierarchy ORDERS what we say. It is never itself said. Never write a question into page copy, and never assert what visitors are thinking — "Most visitors arrive with one specific question:" is presumption and is banned.

OUTPUT. Return ONE JSON object and nothing else$g$),
          '"spec_version": 1}, "findings"',
          '"question_hierarchy": [{"rank": 1, "question": "...", "why": "...", "from_field": "satisfaction_condition", "answered_by": 3, "unanswered": false}], "spec_version": 2}, "findings"'
        )
      ))
WHERE type = 'offer-analyser' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 2. The second gate invocation, over the new collection.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,repair_hierarchy_register}',
      jsonb_build_object(
        'action', 'repair_ordering_register',
        'config', jsonb_build_object(
            'object_field',       'ordering_register_checked.object',
            'items_key',          'question_hierarchy',
            'text_key',           'question',
            'record_key',         'hierarchy_register_repairs',
            'differentiated_key', 'differentiated',
            'ai_service', default_config->'workflow'->'steps'->'repair_ordering_register'->'config'->'ai_service'),
        'output_field', 'hierarchy_register_checked',
        'next_step',    'write_offer_ordering',
        'description',  'Decision D. The question hierarchy is newly derived model text and goes through the SAME register gate as lead_with, by a second invocation rather than a Go change — the gate takes one items_key per call. A hierarchy minted around the gate would be the 23%-born-dirty defect wearing a new field name. Its ai_service block is copied from the sibling step deliberately: offer-analyser has NO ROOT ai_service, so a step without one is live, firing and repairing nothing while recording why (migration 682, and the council llm_reliability seat predicted it verbatim).'))
WHERE type = 'offer-analyser' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 3. Chain it in: lead_with gate -> hierarchy gate -> write.
UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(default_config,
        '{workflow,steps,repair_ordering_register,next_step}', '"repair_hierarchy_register"'),
        '{workflow,steps,write_offer_ordering,config,spec_data}', '"hierarchy_register_checked.object"')
WHERE type = 'offer-analyser' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- VERIFY — DO/RAISE, never a bare SELECT: ON_ERROR_STOP does not abort a COMMIT
-- on a non-empty result set (the RFC_006 lesson).
DO $$
DECLARE
    rows_touched integer;
    prompt_txt   text;
    v_next       text;
    v_spec       text;
    v_items      text;
    v_aisvc      jsonb;
BEGIN
    SELECT count(*) INTO rows_touched FROM agent_definitions
     WHERE type='offer-analyser' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    -- ⚠ An agent-config migration keyed on type can hit TWO active rows; four
    -- types carry duplicates today. Assert exactly one before trusting anything.
    IF rows_touched <> 1 THEN
        RAISE EXCEPTION '723: % active offer-analyser rows, expected exactly 1 — aborting', rows_touched;
    END IF;

    SELECT default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt'
      INTO prompt_txt FROM agent_definitions
     WHERE type='offer-analyser' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF position('THE QUESTION HIERARCHY.' in prompt_txt) = 0 THEN
        RAISE EXCEPTION '723: the guidance block did not land — the OUTPUT anchor did not match';
    END IF;
    IF position('"question_hierarchy"' in prompt_txt) = 0 THEN
        RAISE EXCEPTION '723: question_hierarchy is absent from the OUTPUT contract';
    END IF;
    IF position('"spec_version": 2' in prompt_txt) = 0 THEN
        RAISE EXCEPTION '723: spec_version was not bumped to 2';
    END IF;
    -- The merge trap: a key omitted by the model leaves the previous run standing.
    IF position('EVERY key shown in "ordering" must be present on every run' in prompt_txt) = 0 THEN
        RAISE EXCEPTION '723: the required-every-run warning is missing — question_hierarchy would silently go stale';
    END IF;

    SELECT default_config->'workflow'->'steps'->'repair_hierarchy_register'->'config'->>'items_key',
           default_config->'workflow'->'steps'->'repair_hierarchy_register'->'config'->'ai_service',
           default_config->'workflow'->'steps'->'repair_ordering_register'->>'next_step',
           default_config->'workflow'->'steps'->'write_offer_ordering'->'config'->>'spec_data'
      INTO v_items, v_aisvc, v_next, v_spec
      FROM agent_definitions
     WHERE type='offer-analyser' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF v_items IS DISTINCT FROM 'question_hierarchy' THEN
        RAISE EXCEPTION '723: the hierarchy gate step is missing or misconfigured (items_key=%)', v_items;
    END IF;
    -- Without a step-level ai_service the gate is live, fires, and repairs
    -- nothing while recording why. Assert it rather than trusting the copy.
    IF v_aisvc IS NULL OR v_aisvc->>'provider' IS NULL THEN
        RAISE EXCEPTION '723: the hierarchy gate has no resolvable ai_service — it would repair nothing silently';
    END IF;
    IF v_next IS DISTINCT FROM 'repair_hierarchy_register' THEN
        RAISE EXCEPTION '723: the lead_with gate does not chain into the hierarchy gate (next_step=%)', v_next;
    END IF;
    IF v_spec IS DISTINCT FROM 'hierarchy_register_checked.object' THEN
        RAISE EXCEPTION '723: the write step does not read the hierarchy gate output (spec_data=%) — the gate would run and be discarded', v_spec;
    END IF;

    RAISE NOTICE '723 OK: question_hierarchy in the contract at spec_version 2, gated by a second repair_ordering_register invocation, chained into the write';
END $$;

COMMIT;

-- POST-APPLY READING, with the demand control that makes a zero informative.
--
--   SELECT s.domain, jsonb_array_length(ss.data->'question_hierarchy') AS questions,
--          count(*) FILTER (WHERE (q->>'unanswered')::bool) AS unanswered
--     FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
--          LATERAL jsonb_array_elements(ss.data->'question_hierarchy') q
--    WHERE ss.is_current AND ss.aspect='offer_ordering'
--    GROUP BY 1,2 ORDER BY 1;
--
-- ⚠ A ZERO HERE MEANS "no site has been re-analysed since this applied", NOT
-- "the model ignored it". offer_ordering is written only when offer-analyser
-- runs. Confirm the demand side first:
--   SELECT count(*) FROM site_specs WHERE aspect='offer_ordering' AND created_at > '<apply time>';
--
-- ⚠ AND EXPECT THE TOP ENTRIES TO BE `unanswered` ON THE FIRST PASS. That is the
-- pre-registered result, restating the 10%-of-186 measurement per site. A first
-- pass answering everything means points were stretched to cover questions,
-- which is the failure this criterion exists to catch.

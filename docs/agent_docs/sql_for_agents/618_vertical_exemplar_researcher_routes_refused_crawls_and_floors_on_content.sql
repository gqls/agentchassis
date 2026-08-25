-- 618_vertical_exemplar_researcher_routes_refused_crawls_and_floors_on_content.sql
--
-- bugs_open/376, the fix designed in its section 11 and verified there against the
-- code that executes it. Config only, live on apply, no image roll, NO ORDERING
-- CONSTRAINT CLAIMED (owner ruling 2026-07-29): it depends on no unshipped code and
-- nothing depends on it.
--
-- WHAT IS WRONG. `vertical-exemplar-researcher` is the second hop of every greenfield
-- build (classify -> research the vertical -> strategy -> ...). It crawls three
-- exemplar sites in a strict line with NO error_step on any crawl step, so ONE
-- Firecrawl refusal (thespruce.com, refused 4 of 5 draws on garden-tools.uk,
-- 2026-08-23) kills the stage, discards the crawls that succeeded, and never
-- chains `create_next_item` - the build stops with nothing to say why. That is
-- outcome (a). Outcome (b), found on the homegarden.uk canary 2026-08-25 10:52Z:
-- a crawl can return `success: true` and deliver NOTHING (which.co.uk formatted to
-- `source_count: 0, content_quality: none`) and the chain proceeds to synthesise a
-- vertical landscape from two exemplars while every status says three.
--
-- THE FIX, two halves, and why each is shaped the way it is:
--
--   1. crawl_exemplar_N.error_step = format_exemplar_N  (STEP level).
--      A refused crawl is routed to ITS OWN format step, not to the next crawl.
--      [VERIFIED 2026-08-25, format_crawl_for_analysis_action.go] when the format
--      action finds no pages it returns {source_count:0, content_quality:"none",
--      sources:[]} with a NIL error, and ExtractNestedField returns nil for a
--      MISSING crawl_N exactly as for an empty one - so (a) collapses into (b) and
--      the floor below can count it. Skipping format_N would leave formatted_N
--      ABSENT, which `synthesise` interpolates as "<no value>" and the floor cannot
--      count. Step-level placement: routeToErrorStepOrFail (coordinator.go:3916)
--      prefers step.ErrorStep and falls back to config.error_step; both routes are
--      live ([MEASURED 2026-08-25] 3,868 persisted plan steps carry a step-level
--      error_step, 18,710 a config-level one).
--
--   2. A floor between format_exemplar_3 and synthesise, evaluated on CONTENT.
--      `check_exemplar_floor` (conditional_branch) passes when at least TWO of the
--      three formatted_N.source_count are > 0. Counting step SUCCESS would count
--      which.co.uk (success:true, 0 sources); the receipt is not the result.
--
--      NO PARENTHESES IN THE CONDITION, AND THAT IS NOT STYLE [VERIFIED
--      conditional_branch_action.go:144-186 + LANDMINES "conditional_branch ignores
--      parentheses"]: the evaluator splits on " OR " first, then " AND ", and
--      cleanExpressionPart STRIPS brackets. So
--          A AND B OR A AND C OR B AND C
--      parses as (A^B) v (A^C) v (B^C) = "at least two", and adding brackets for
--      readability would silently change the meaning. fail_on_non_numeric:true is
--      load-bearing: without it a source_count that fails to resolve evaluates
--      FALSE and routes to the failure arm silently, so a broken instrument would
--      read as a genuine shortfall.
--
--   3. Below the floor: FAIL LOUDLY, NAMED. `record_exemplar_floor` composes a
--      message carrying the three counts (query_database, so the counts are in
--      collected_data AND in the failure text), then `insufficient_exemplars`
--      (fail_workflow, reason_field) ends the run with a FAILURE verdict. The
--      handler saga marks the `needs_vertical_research` work item failed with that
--      text in `error` - a durable row that names the cause - and
--      create_next_item never runs. Today's failure is loud; this must not trade it
--      for a quiet landscape written from no research, which would be strictly
--      worse than the bug.
--
-- WHAT THIS DOES NOT DO (bug 376 section 11e): it does not exclude refused hosts at
-- selection (candidate 2, a separate change to WHAT IS CHOSEN); it does not touch
-- select_exemplars' missing temperature; and a format_exemplar_N step that itself
-- ERRORS still has no error_step - that is the pre-existing rarer case, unchanged.
--
-- VERIFY AFTER APPLY (all three, or the fix is not proven - section 11e):
--   1. a draw containing a REFUSED host reaches create_next_item, with the refused
--      exemplar's formatted_N.content_quality = 'none';
--   2. a succeeds-but-empty host (which.co.uk/reviews/home-and-garden, 2026-08-25)
--      counts ZERO toward the floor;
--   3. an induced below-floor run (two of three empty) FAILS and the work item's
--      `error` names the counts.
--   Read formatted_N.source_count, never the crawl step's record - a refused crawl's
--   own record reads "success": true (a dispatch receipt).

-- Probe guard: tell the runner when this is already applied.
DO $probe$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'vertical-exemplar-researcher'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #> '{workflow,steps}' ? 'check_exemplar_floor'
    ) THEN
        RAISE EXCEPTION '376/618: already applied - check_exemplar_floor is present';
    END IF;
END $probe$;

BEGIN;

-- README rule: every migration touching agent_definitions opens with a snapshot.
SELECT snapshot_agent('vertical-exemplar-researcher',
  '618_vertical_exemplar_researcher_routes_refused_crawls_and_floors_on_content: pre-update');

-- == DRIFT GUARD + EDIT, one block so the guard and the edit read the same row ==
DO $edit$
DECLARE
    v_id     uuid;
    v_steps  jsonb;
    v_n      int;
    v_name   text;
    v_cond   text := 'formatted_1.source_count > 0 AND formatted_2.source_count > 0 OR formatted_1.source_count > 0 AND formatted_3.source_count > 0 OR formatted_2.source_count > 0 AND formatted_3.source_count > 0';
BEGIN
    SELECT id, default_config #> '{workflow,steps}'
      INTO v_id, v_steps
      FROM agent_definitions
     WHERE type = 'vertical-exemplar-researcher'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     ORDER BY version DESC
     LIMIT 1;

    IF v_id IS NULL THEN
        RAISE EXCEPTION '376/618: no live vertical-exemplar-researcher row found';
    END IF;

    -- The shape section 11 was verified against. Any drift = a shape this file
    -- has not read; stop rather than edit blind.
    FOR v_n IN 1..3 LOOP
        v_name := format('crawl_exemplar_%s', v_n);
        IF v_steps #>> ARRAY[v_name, 'action'] IS DISTINCT FROM 'firecrawl_crawl' THEN
            RAISE EXCEPTION '376/618 drift: %.action is %, expected firecrawl_crawl', v_name, v_steps #>> ARRAY[v_name, 'action'];
        END IF;
        IF v_steps #> ARRAY[v_name] ? 'error_step' OR v_steps #> ARRAY[v_name, 'config'] ? 'error_step' THEN
            RAISE EXCEPTION '376/618 drift: % already carries an error_step (config or step level) - someone has been here', v_name;
        END IF;
        IF v_steps #>> ARRAY[v_name, 'next_step'] IS DISTINCT FROM format('format_exemplar_%s', v_n) THEN
            RAISE EXCEPTION '376/618 drift: %.next_step is %, expected format_exemplar_%', v_name, v_steps #>> ARRAY[v_name, 'next_step'], v_n;
        END IF;
        v_name := format('format_exemplar_%s', v_n);
        IF v_steps #>> ARRAY[v_name, 'action'] IS DISTINCT FROM 'format_crawl_for_analysis'
           OR v_steps #>> ARRAY[v_name, 'output_field'] IS DISTINCT FROM format('formatted_%s', v_n) THEN
            RAISE EXCEPTION '376/618 drift: % is not format_crawl_for_analysis -> formatted_%', v_name, v_n;
        END IF;
    END LOOP;
    IF v_steps #>> '{format_exemplar_3,next_step}' IS DISTINCT FROM 'synthesise' THEN
        RAISE EXCEPTION '376/618 drift: format_exemplar_3.next_step is %, expected synthesise', v_steps #>> '{format_exemplar_3,next_step}';
    END IF;
    IF NOT (v_steps ? 'synthesise') OR v_steps #>> '{synthesise,action}' IS DISTINCT FROM 'execute_llm_prompt' THEN
        RAISE EXCEPTION '376/618 drift: synthesise step missing or not execute_llm_prompt';
    END IF;
    IF v_steps ? 'check_exemplar_floor' OR v_steps ? 'record_exemplar_floor' OR v_steps ? 'insufficient_exemplars' THEN
        RAISE EXCEPTION '376/618 drift: a step this migration adds already exists';
    END IF;
    -- The one property the floor's correctness rests on, asserted on the literal
    -- that will be written rather than on prose about it.
    IF position('(' in v_cond) > 0 OR position(')' in v_cond) > 0 THEN
        RAISE EXCEPTION '376/618: the floor condition must carry NO parentheses (conditional_branch strips them)';
    END IF;

    -- Half one: route each refused crawl to its own format step (STEP level).
    FOR v_n IN 1..3 LOOP
        v_steps := jsonb_set(v_steps,
            ARRAY[format('crawl_exemplar_%s', v_n), 'error_step'],
            to_jsonb(format('format_exemplar_%s', v_n)), true);
        v_steps := jsonb_set(v_steps,
            ARRAY[format('crawl_exemplar_%s', v_n), 'description'],
            to_jsonb(format('Shallow crawl of exemplar %s (front page + direct links). error_step -> its own format step (376/618): a refused crawl formats to source_count 0 and is COUNTED by the floor, not skipped past it', v_n)), true);
    END LOOP;

    -- Half two: the floor sits between format_exemplar_3 and synthesise.
    v_steps := jsonb_set(v_steps, '{format_exemplar_3,next_step}', to_jsonb('check_exemplar_floor'::text), true);

    v_steps := v_steps || jsonb_build_object(
        'check_exemplar_floor', jsonb_build_object(
            'action', 'conditional_branch',
            'config', jsonb_build_object(
                'condition', v_cond,
                'fail_on_non_numeric', true,
                'then_step', 'synthesise',
                'else_step', 'record_exemplar_floor'
            ),
            'description', 'Floor on CONTENT, not step success (376/618): at least two of three exemplars must have formatted to source_count > 0. Bracket-free on purpose - conditional_branch splits OR before AND and strips parentheses, so this reads (1^2) v (1^3) v (2^3). fail_on_non_numeric so a broken count errors instead of routing to the failure arm silently'
        ),
        'record_exemplar_floor', jsonb_build_object(
            'action', 'query_database',
            'config', jsonb_build_object(
                'query', 'SELECT format(''vertical exemplar research floor unmet (bugs_open/376): source_count exemplar_1=%s exemplar_2=%s exemplar_3=%s - at least two of three must be > 0; no landscape written, no strategy item chained'', $1::text, $2::text, $3::text) AS reason',
                'params', jsonb_build_array('formatted_1.source_count', 'formatted_2.source_count', 'formatted_3.source_count'),
                'output_format', 'object'
            ),
            'next_step', 'insufficient_exemplars',
            'description', 'Compose the failure text WITH the three counts so the work item''s error names the cause (376/618). A query_database step rather than a static reason: the counts live in collected_data and orchestration_states is reaped within ~25h',
            'output_field', 'exemplar_floor'
        ),
        'insufficient_exemplars', jsonb_build_object(
            'action', 'fail_workflow',
            'config', jsonb_build_object(
                'reason_field', 'exemplar_floor.reason',
                'reason', 'vertical exemplar research floor unmet (bugs_open/376): fewer than two of three exemplars delivered content'
            ),
            'description', 'End with a FAILURE verdict (376/618): the handler saga marks the needs_vertical_research item failed with the reason above, create_next_item never runs. Loud and named - a landscape synthesised from under-floor research with every step green would be worse than the bug'
        )
    );

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config, '{workflow,steps}', v_steps, false),
           updated_at = now()
     WHERE id = v_id;

    RAISE NOTICE '376/618: edited row % (3 error_steps, floor, record + fail steps)', v_id;
END $edit$;

-- == VERIFY: the post-conditions, each one a RAISE (a SELECT cannot stop a COMMIT) ==
DO $verify$
DECLARE
    v_steps    jsonb;
    v_dangling int;
    v_n        int;
BEGIN
    SELECT default_config #> '{workflow,steps}' INTO v_steps
      FROM agent_definitions
     WHERE type = 'vertical-exemplar-researcher'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     ORDER BY version DESC LIMIT 1;

    FOR v_n IN 1..3 LOOP
        IF v_steps #>> ARRAY[format('crawl_exemplar_%s', v_n), 'error_step'] IS DISTINCT FROM format('format_exemplar_%s', v_n) THEN
            RAISE EXCEPTION '376/618 verify: crawl_exemplar_%.error_step not set', v_n;
        END IF;
    END LOOP;
    IF v_steps #>> '{format_exemplar_3,next_step}' <> 'check_exemplar_floor'
       OR v_steps #>> '{check_exemplar_floor,config,then_step}' <> 'synthesise'
       OR v_steps #>> '{check_exemplar_floor,config,else_step}' <> 'record_exemplar_floor'
       OR v_steps #>> '{record_exemplar_floor,next_step}' <> 'insufficient_exemplars'
       OR (v_steps #>> '{check_exemplar_floor,config,fail_on_non_numeric}')::boolean IS DISTINCT FROM true THEN
        RAISE EXCEPTION '376/618 verify: floor wiring wrong';
    END IF;
    IF position('(' in (v_steps #>> '{check_exemplar_floor,config,condition}')) > 0 THEN
        RAISE EXCEPTION '376/618 verify: parentheses reached the stored condition';
    END IF;
    -- 546's post-condition: every edge must still resolve to a real step.
    SELECT count(*) INTO v_dangling
      FROM (
        SELECT e.v->>'next_step' AS tgt FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'next_step'
        UNION ALL SELECT e.v->>'error_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'error_step'
        UNION ALL SELECT e.v->'config'->>'then_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'then_step'
        UNION ALL SELECT e.v->'config'->>'else_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'else_step'
      ) AS edges
     WHERE tgt IS NOT NULL AND NOT (v_steps ? tgt);
    IF v_dangling > 0 THEN
        RAISE EXCEPTION '376/618 verify: % workflow edge(s) point at a step that does not exist', v_dangling;
    END IF;
    -- synthesise still reads all three formatted fields (the floor counts what it consumes).
    IF NOT (v_steps #> '{synthesise,config,input_fields}' @> '["formatted_1","formatted_2","formatted_3"]'::jsonb) THEN
        RAISE EXCEPTION '376/618 verify: synthesise no longer consumes formatted_1..3';
    END IF;
    RAISE NOTICE '376/618: verified - 3 error_steps, floor wired, 0 dangling edges, 15 steps total = %', (SELECT count(*) FROM jsonb_object_keys(v_steps));
END $verify$;

COMMIT;

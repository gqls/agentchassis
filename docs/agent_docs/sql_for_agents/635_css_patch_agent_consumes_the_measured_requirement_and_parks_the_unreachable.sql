-- 635_css_patch_agent_consumes_the_measured_requirement_and_parks_the_unreachable.sql
--
-- bugs_open/390, commit 3 of 3. The agent stops guessing.
--
-- WHAT THE FIRST TWO COMMITS LEFT. Migration 616 (applied 2026-08-25, council
-- APPROVED round 2, corr ef5f9a0d) corrected a prompt that instructed the losing
-- move, with a blanket !important as the interim lever. Commit ea64845e0
-- (council APPROVED, corr 058b59b6; live on both images since the 2026-08-25
-- ~23:11 roll, stamp 2fb40a96) taught the render audit to record WHICH
-- declaration actually decides a failing element's colour - proven in the page
-- by removing it - and the filer to write three spec fields on every new
-- contrast_failure: `repair_surface` (theme / unreachable / unattributed),
-- `winning_rule`, and `override_requirement` (minimum specificity computed with
-- cascadia, strictness, !important parity, plus an `override_example` the
-- platform has CHECKED with satisfiesRequirement).
--
-- THIS MIGRATION makes css-patch-agent read them:
--
--   1. A `check_repair_surface` gate before the LLM step. `repair_surface =
--      'unreachable'` - the winner is an !important declaration in the element's
--      own style attribute, which no stylesheet rule can outrank - PARKS at
--      needs_human_review (`parked_by = 'css_cascade_unreachable_390'`, the
--      542/546 shape) BEFORE any LLM spend. A rule appended for that finding is
--      inert by construction; authoring it is what bug 390 is.
--   2. The prompt renders the measured requirement when it is present: the
--      winning declaration, the specificity bar and its strictness, and
--      !important ONLY where importance parity demands it - superseding 616's
--      blanket instruction on exactly the rows where we know better. Rows
--      without the fields (legacy, or an unattributed page) keep 616's general
--      guidance unchanged.
--
-- SKEW-SAFE IN EVERY ORDER, so no ordering constraint is claimed (owner ruling
-- 2026-07-29):
--   * a legacy/unattributed row has no `repair_surface`; conditional_branch's
--     compareValues(nil, ...) is false (conditional_branch_action.go:518-521,
--     re-read 2026-08-26), so the gate routes to plan_css_fix - today's path;
--   * the template block is fenced by {{if .input_data.spec.override_requirement}};
--     absent fields render NOTHING. Proven by executing the exact block with Go
--     text/template against four data shapes (full, needs-important,
--     legacy-empty, requirement-without-winner) - no errors, empty on legacy.
--     Fenced on override_requirement rather than winning_rule DELIBERATELY: the
--     filer writes winning_rule for a VERIFIED unreachable winner too, where
--     req is nil, and a template that dereferences a missing sibling errors at
--     execute time and fails the whole LLM step.
--
-- WHY THE PARK CANNOT STRAND THE FINDING. `needs_human_review` is in NEITHER
-- workItemTerminalStatuses (so the row keeps its dedup slot - the finding cannot
-- re-file while parked) NOR workItemClosedStatuses (so WII-016's retraction
-- still closes it the moment a page rebuild removes the inline offender and the
-- next audit stops observing the pairing). The park is self-draining on positive
-- evidence, exactly like the css_no_theme_198 parks.
--
-- WHAT THIS DOES NOT DO:
--   * Nothing for `unattributed` - it routes to the LLM exactly as today, under
--     616's corrected general guidance. Parking on "we do not know" would park
--     the entire legacy backlog.
--   * Nothing about completion semantics (owner decision 2026-08-25: routing
--     first) and nothing about erasure (bugs_open/396).
--
-- DRIFT ANCHOR NOTE FOR THE NEXT MIGRATION: after this applies, the prompt
-- carries 616's text AND the {{if}} block; assert THIS file's shape, not 616's.

-- Probe guard: tell the runner when this is already applied.
DO $probe$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'css-patch-agent'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #> '{workflow,steps}' ? 'check_repair_surface'
    ) THEN
        RAISE EXCEPTION '390/635: already applied - check_repair_surface already exists';
    END IF;
END $probe$;

BEGIN;

SELECT snapshot_agent('css-patch-agent',
  '635_css_patch_agent_consumes_the_measured_requirement: pre-update');

-- Guard and apply in ONE block: every literal declared once and used for both
-- the assertion and the edit (the 616-round-2 lesson - a literal written down
-- twice is a literal that will disagree with itself).
DO $apply$
DECLARE
    v_anchor616 text := 'can never win on position';
    v_heading   text := '## Current Stylesheet (read-only context';
    v_block     text := $tpl$## Audit Finding{{if .input_data.spec.override_requirement}}
## The declaration you must BEAT (measured in the browser, proven by removal)
{{with .input_data.spec.winning_rule}}{{if .selector}}Winner: `{{.selector}}` ({{.surface}}), declaring `{{.decl}}`{{if .important}}, marked !important{{end}}.{{else}}Winner: the element's own style attribute, declaring `{{.decl}}`.{{end}}
{{end}}{{with .input_data.spec.override_requirement}}{{.why}}.
Your selector must reach specificity ({{.min_specificity_text}}){{if .strictly_greater}} and EXCEED it strictly - an equal-specificity rule loses on source order{{end}}.
{{if .needs_important}}The corrected property MUST carry !important to match the winner.{{else}}Do NOT use !important - it is not needed to win here; this measured section supersedes the general !important guidance below.{{end}}
{{end}}{{with .input_data.spec.override_example}}A selector the platform has VERIFIED satisfies this requirement: `{{.}}`
{{end}}{{end}}$tpl$;
    v_rows   int;
    v_steps  jsonb;
    v_prompt text;
BEGIN
    -- Row-count assertion first (616 round 2, debug_historian, HIGH): four agent
    -- types in this estate carry TWO active non-snapshot rows and only the
    -- higher version loads. css-patch-agent is not one of them today; assert
    -- rather than inherit.
    SELECT count(*) INTO v_rows
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '390/635: expected exactly ONE live css-patch-agent row, found %', v_rows;
    END IF;

    SELECT default_config #> '{workflow,steps}' INTO v_steps
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- The edges being rewired, asserted at their current values (542's shape).
    IF v_steps #>> '{check_base_integrity,config,then_step}' <> 'plan_css_fix' THEN
        RAISE EXCEPTION '390/635 drift: check_base_integrity.then_step is %, expected plan_css_fix',
            v_steps #>> '{check_base_integrity,config,then_step}';
    END IF;
    IF NOT (v_steps ? 'complete_refused') THEN
        RAISE EXCEPTION '390/635 drift: complete_refused terminal is missing - the park has nowhere to land';
    END IF;
    IF v_steps #>> '{plan_css_fix,action}' <> 'execute_llm_prompt' THEN
        RAISE EXCEPTION '390/635 drift: plan_css_fix.action is %, expected execute_llm_prompt',
            v_steps #>> '{plan_css_fix,action}';
    END IF;

    v_prompt := v_steps #>> '{plan_css_fix,config,prompt_template}';
    IF v_prompt IS NULL THEN
        RAISE EXCEPTION '390/635 drift: plan_css_fix.config.prompt_template missing';
    END IF;
    -- 616 must still be in place: the {{if}} block's "supersedes the general
    -- guidance" sentence is meaningless against a prompt that has lost it.
    IF position(v_anchor616 in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/635 drift: 616''s correction is not in the prompt (rolled back? superseded?)';
    END IF;
    -- The insertion anchor, present exactly once. Searching twice, no counts.
    IF position('## Audit Finding' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/635 drift: the "## Audit Finding" heading is not in the prompt';
    END IF;
    IF position('## Audit Finding' in substr(v_prompt, position('## Audit Finding' in v_prompt) + 16)) > 0 THEN
        RAISE EXCEPTION '390/635 drift: "## Audit Finding" appears more than once';
    END IF;
    IF position(v_heading in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/635 drift: the stylesheet heading is not in the prompt';
    END IF;
    IF position('{{if .input_data.spec.override_requirement}}' in v_prompt) > 0 THEN
        RAISE EXCEPTION '390/635 drift: the measured-requirement block is already present';
    END IF;

    -- THE EDIT, all three parts against the ONE live row.
    UPDATE agent_definitions
       SET default_config =
           jsonb_set(
             jsonb_set(
               jsonb_set(
                 jsonb_set(
                   default_config,
                   '{workflow,steps,check_repair_surface}',
                   jsonb_build_object(
                     'action', 'conditional_branch',
                     'config', jsonb_build_object(
                       'condition', 'input_data.spec.repair_surface == ''unreachable''',
                       'then_step', 'mark_cascade_unreachable',
                       'else_step', 'plan_css_fix'
                     ),
                     'description', 'bugs_open/390: an !important inline style attribute is the one thing no stylesheet rule can outrank. A legacy or unattributed row has no repair_surface field, compareValues(nil,...) is false, and it routes to plan_css_fix - today''s path, by construction.'
                   ),
                   true
                 ),
                 '{workflow,steps,mark_cascade_unreachable}',
                 jsonb_build_object(
                   'action', 'update_work_item_status',
                   'next_step', 'complete_refused',
                   'config', jsonb_build_object(
                     'status', 'needs_human_review',
                     'error_message', 'css-patch refused (bugs_open/390): the browser proved the winning declaration is an !important rule in the element''s own style attribute, which no stylesheet rule can outrank at any specificity. The repair surface is the component markup (spec.winning_rule names the offender), or a page rebuild that drops the inline style. This row closes itself when the next render audit no longer observes the pairing.',
                     'result_fields', jsonb_build_object('parked_by', 'css_cascade_unreachable_390')
                   ),
                   'description', 'bugs_open/390: park before any LLM spend. needs_human_review keeps the dedup slot (no re-filing) and stays retractable (WII-016 closes it on positive re-measurement), so the park is self-draining.'
                 ),
                 true
               ),
               '{workflow,steps,check_base_integrity,config,then_step}',
               to_jsonb('check_repair_surface'::text),
               false
             ),
             '{workflow,steps,plan_css_fix,config,prompt_template}',
             to_jsonb(replace(v_prompt, '## Audit Finding', v_block)),
             false
           ),
           updated_at = NOW()
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '390/635: UPDATE touched % rows, expected exactly 1', v_rows;
    END IF;
END $apply$;

-- VERIFY: every post-condition RAISEs (a SELECT cannot stop a COMMIT).
DO $verify$
DECLARE
    v_steps    jsonb;
    v_prompt   text;
    v_dangling int;
BEGIN
    SELECT default_config #> '{workflow,steps}' INTO v_steps
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_steps #>> '{check_base_integrity,config,then_step}' <> 'check_repair_surface' THEN
        RAISE EXCEPTION '390/635 verify: check_base_integrity not rewired';
    END IF;
    IF v_steps #>> '{check_repair_surface,config,condition}' <> 'input_data.spec.repair_surface == ''unreachable''' THEN
        RAISE EXCEPTION '390/635 verify: gate condition is not the expected literal - got %',
            v_steps #>> '{check_repair_surface,config,condition}';
    END IF;
    IF v_steps #>> '{check_repair_surface,config,then_step}' <> 'mark_cascade_unreachable'
       OR v_steps #>> '{check_repair_surface,config,else_step}' <> 'plan_css_fix' THEN
        RAISE EXCEPTION '390/635 verify: gate arms miswired';
    END IF;
    IF v_steps #>> '{mark_cascade_unreachable,config,status}' <> 'needs_human_review'
       OR v_steps #>> '{mark_cascade_unreachable,config,result_fields,parked_by}' <> 'css_cascade_unreachable_390'
       OR v_steps #>> '{mark_cascade_unreachable,next_step}' <> 'complete_refused' THEN
        RAISE EXCEPTION '390/635 verify: park step wrong - the stamp must precede the terminal (542''s rule)';
    END IF;

    v_prompt := v_steps #>> '{plan_css_fix,config,prompt_template}';
    IF position('{{if .input_data.spec.override_requirement}}' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/635 verify: the measured-requirement block is absent from the prompt';
    END IF;
    IF position('## Audit Finding' in v_prompt) = 0
       OR position('can never win on position' in v_prompt) = 0
       OR position('"css_added"' in v_prompt) = 0
       OR position('{{.current_css.css_content}}' in v_prompt) = 0 THEN
        RAISE EXCEPTION '390/635 verify: an existing prompt part was damaged by the insert';
    END IF;

    -- 546's post-condition, re-run because this file edits the same row:
    -- every edge in the workflow must resolve to a real step.
    SELECT count(*) INTO v_dangling
      FROM (
        SELECT e.v->>'next_step' AS tgt FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'next_step'
        UNION ALL
        SELECT e.v->>'error_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'error_step'
        UNION ALL
        SELECT e.v->'config'->>'then_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'then_step'
        UNION ALL
        SELECT e.v->'config'->>'else_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'else_step'
      ) AS edges
     WHERE tgt IS NOT NULL AND NOT (v_steps ? tgt);
    IF v_dangling > 0 THEN
        RAISE EXCEPTION '390/635 verify: % workflow edge(s) point at a step that does not exist', v_dangling;
    END IF;

    -- No non-success exit reaches a terminal without stamping first (546).
    IF v_steps #>> '{check_has_css,config,else_step}' <> 'mark_no_css'
       OR v_steps #>> '{check_saved,config,else_step}' <> 'mark_append_refused'
       OR v_steps #>> '{check_base_integrity,config,else_step}' <> 'mark_base_unsafe' THEN
        RAISE EXCEPTION '390/635 verify: 542/546 park wiring disturbed';
    END IF;

    RAISE NOTICE '390/635: verified - gate wired, park stamps before its terminal, prompt block present, 0 dangling edges';
END $verify$;

COMMIT;

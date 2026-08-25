-- 619_improvement_loop_bypasses_the_llm_audit_seats_HOLD.sql
--
-- _HOLD: NOT for the runner. Applied BY HAND, on the owner's word, as Phase 1 of
-- docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/
--   PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md
-- It is held because it changes what the improvement loop DOES fleet-wide the
-- moment the sweep is re-enabled, and the owner asked for a plan, not a change.
--
-- WHAT THE OWNER ASKED (2026-08-25, in substance): "turn off the evolutionary
-- aspect of the improvement loop for now — the bit that keeps rewriting pages
-- that have been judged to be good for the sake of aspirational improvements.
-- It should stay because it's a great thing, but it is causing too many bad /
-- unexpected renders, so switch just that bit off and turn the improvement loop
-- back on."
--
-- WHAT "THAT BIT" IS, measured. The loop has two kinds of seat. The MECHANICAL
-- seats (quality / design / completeness discovery agents) find defects — broken
-- nav, empty sections, missing CSS, 404 assets. The LLM seats — design-audit
-- (visual-design-auditor + content-quality-auditor), site-review-agent,
-- offer-analyser, brief-fidelity-auditor — file OPINIONS about pages that already
-- work, and write_audit_findings routes those opinions to handlers that
-- REGENERATE the page: content_rewrite / needs_content_page / tone_shift ->
-- page-build-handler, needs_content_planning -> content-gap-planner.
-- [MEASURED 2026-08-25, site_work_items UNION archive] from the design-audit
-- source alone, lifetime: 976 content_rewrite, 399 needs_content_page, 964
-- needs_content_planning, 26 tone_shift. bugs_open/238 is what one of those did to
-- a page that was fine. The four LLM seats are the evolutionary aspect; the
-- rewrites are their findings dispatched.
--
-- WHAT THIS DOES. One edge: call_completeness_discovery.next_step moves from
-- spawn_design_audit to record_audit_pass. The loop then runs: enrichment (news
-- feed, directory) -> fingerprint -> the three mechanical seats -> record the
-- audit -> triage -> rerender -> dispatch. The eight LLM-seat steps
-- (spawn/call x design_audit, site_review, offer_analyser, brief_fidelity) stay in
-- the workflow, inert and unreachable — nothing is deleted, and the ROLLBACK is
-- the same edge flipped back. "It should stay": it does.
--
-- WHAT THIS DOES NOT DO, said plainly so nobody reads it as more:
--   * It does not re-enable the sweep. That is a one-row UPDATE on
--     scheduled_tasks the plan lists as the owner's own step AFTER this applies,
--     because the vigilant_designer_offer_analysis lane is holding that switch
--     until ordering is settled, and because 389 (2026-08-19) quotes the owner:
--     "it will be expensive so I am wary of costs".
--   * It does not stop the render-audit rotation (site-render-audit-rotation,
--     LIVE, hourly), whose contrast_failure -> css-patch-agent items are the OTHER
--     live source of bad renders (bugs_open/198, /390, /396: 239 completions in
--     the last 14 days). That is a separate carrier and a separate decision; the
--     plan names it.
--   * It does not make the seats record-only. That needs the Go opt-in
--     `filing_mode: record` on write_audit_findings (Phase 2, same lane, council
--     submission pending), after which the seats can be switched back on as
--     VERDICT rows that dispatch nothing — the site acceptance council of RFC_056.
--   * It does not close the second door. detected-item-promoter (LIVE, every 15
--     min) promotes any 'detected' row whose (item_type, handler) pair has ever
--     completed — [MEASURED 2026-08-25] 26 LLM-audit rows were promoted between
--     08-20 and 08-24 while the sweep was OFF. With the seats bypassed no new such
--     rows are filed (0 exist at 'detected' today), but a hand-fired auditor
--     (the offer-analyser-oneshot-* rows, all disabled) would reopen it.
--
-- NO ORDERING CONSTRAINT IS CLAIMED (owner ruling 2026-07-29): config only, live
-- on apply, depends on no unshipped code.

DO $probe$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'improvement-loop'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #>> '{workflow,steps,call_completeness_discovery,next_step}' = 'record_audit_pass'
    ) THEN
        RAISE EXCEPTION '619: already applied - call_completeness_discovery already bypasses the LLM seats';
    END IF;
END $probe$;

BEGIN;

SELECT snapshot_agent('improvement-loop', '619_improvement_loop_bypasses_the_llm_audit_seats_HOLD: pre-update');

DO $edit$
DECLARE
    v_id    uuid;
    v_steps jsonb;
    v_seat  text;
BEGIN
    SELECT id, default_config #> '{workflow,steps}' INTO v_id, v_steps
      FROM agent_definitions
     WHERE type = 'improvement-loop'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     ORDER BY version DESC LIMIT 1;
    IF v_id IS NULL THEN
        RAISE EXCEPTION '619: no live improvement-loop row';
    END IF;

    -- The shape this was written against [MEASURED 2026-08-25, 31 steps].
    IF v_steps #>> '{call_completeness_discovery,next_step}' IS DISTINCT FROM 'spawn_design_audit' THEN
        RAISE EXCEPTION '619 drift: call_completeness_discovery.next_step is %, expected spawn_design_audit',
            v_steps #>> '{call_completeness_discovery,next_step}';
    END IF;
    IF v_steps #>> '{call_brief_fidelity,next_step}' IS DISTINCT FROM 'record_audit_pass' THEN
        RAISE EXCEPTION '619 drift: the LLM chain no longer ends at record_audit_pass (got %)',
            v_steps #>> '{call_brief_fidelity,next_step}';
    END IF;
    FOREACH v_seat IN ARRAY ARRAY['spawn_design_audit','call_design_audit','spawn_site_review','call_site_review',
                                  'spawn_offer_analyser','call_offer_analyser','spawn_brief_fidelity','call_brief_fidelity',
                                  'record_audit_pass'] LOOP
        IF NOT (v_steps ? v_seat) THEN
            RAISE EXCEPTION '619 drift: step % missing', v_seat;
        END IF;
    END LOOP;

    v_steps := jsonb_set(v_steps, '{call_completeness_discovery,next_step}', to_jsonb('record_audit_pass'::text), true);
    v_steps := jsonb_set(v_steps, '{call_completeness_discovery,description}',
        to_jsonb('Run completeness checks (empty sections). next_step -> record_audit_pass (619, owner 2026-08-25): the four LLM audit seats (design audit, site review, offer analyser, brief fidelity) are BYPASSED, not deleted - their findings dispatch page rewrites, and those are switched off until write_audit_findings can file them record-only (RFC_056 Phase 2). Restore: next_step = spawn_design_audit'::text), true);

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config, '{workflow,steps}', v_steps, false), updated_at = now()
     WHERE id = v_id;
    RAISE NOTICE '619: edited % - LLM seats bypassed, 31 steps kept', v_id;
END $edit$;

DO $verify$
DECLARE
    v_steps    jsonb;
    v_dangling int;
BEGIN
    SELECT default_config #> '{workflow,steps}' INTO v_steps
      FROM agent_definitions
     WHERE type = 'improvement-loop'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     ORDER BY version DESC LIMIT 1;
    IF v_steps #>> '{call_completeness_discovery,next_step}' <> 'record_audit_pass' THEN
        RAISE EXCEPTION '619 verify: edge not moved';
    END IF;
    IF (SELECT count(*) FROM jsonb_object_keys(v_steps)) <> 31 THEN
        RAISE EXCEPTION '619 verify: expected 31 steps (nothing deleted), got %', (SELECT count(*) FROM jsonb_object_keys(v_steps));
    END IF;
    SELECT count(*) INTO v_dangling
      FROM (
        SELECT e.v->>'next_step' AS tgt FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'next_step'
        UNION ALL SELECT e.v->>'error_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'error_step'
        UNION ALL SELECT e.v->'config'->>'error_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'error_step'
        UNION ALL SELECT e.v->'config'->>'then_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'then_step'
        UNION ALL SELECT e.v->'config'->>'else_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'else_step'
      ) AS edges
     WHERE tgt IS NOT NULL AND NOT (v_steps ? tgt);
    IF v_dangling > 0 THEN
        RAISE EXCEPTION '619 verify: % dangling edge(s)', v_dangling;
    END IF;
    RAISE NOTICE '619: verified - mechanical seats -> record_audit_pass, 31 steps, 0 dangling edges';
END $verify$;

COMMIT;

-- 531 — mirror 530 into the council-gate roster BY HAND (099 --apply is
-- suspended: it would revert migration 377 across all 17 seats), and fix a
-- SECOND defect found while reading the two rosters side by side.
--
-- DEFECT 1 — same as 530: the seat asserts the coordinator reads ONLY
-- step.config.error_step and that a step-level error_step is "silently inert".
-- routeToErrorStepOrFail (coordinator.go:3666-3679) checks step-level FIRST and
-- its own comment calls that the preferred location; config-level is the
-- backward-compatibility fallback. Both are read. 530's header carries the
-- full quotation and the provenance.
--
-- DEFECT 2 — the council-gate copy carries a MANGLED SECTION HEADING:
--     "## The author's stated rationale loop's load-bearing disciplines"
-- where fix-proposer reads "## The diagnosis loop's load-bearing disciplines".
-- Cause, read at source: 099_SYNC_gate_roster.py:85 does an UNANCHORED
--     p.replace("## The diagnosis", "## The author's stated rationale")
-- which is meant to swap the diagnosis CONTEXT block for the submitter's
-- rationale, and also hits any other heading that starts with those words. The
-- result is a sentence that parses as nonsense at the top of the list of
-- disciplines the seat exists to defend. The mirror script is being anchored in
-- the same task so it cannot manufacture this again; this migration repairs the
-- live text, which the script fix cannot do.
--
-- 377 SAFETY. Every edit here is AFTER the shared evidence prefix and its
-- <!--CACHE_BREAKPOINT-->, so the marker's byte offset must be UNCHANGED. The
-- DO block asserts exactly that and aborts otherwise — 377's ~68% prefix-cache
-- saving is the thing most easily destroyed by a well-meaning prompt edit.
--
-- ⚠ HONEST NOTE ON THAT GUARD, because a control that cannot fail is not
-- evidence: with THIS file's anchor set the marker-offset check is structurally
-- unfirable. All three anchors are downstream of the marker, and an anchor
-- planted upstream would trip the occurs-exactly-once check first. It is kept
-- because the next author to add an anchor may not check where it sits. The
-- guard that CAN fail here is the fleet-wide prefix-fragmentation check in the
-- verify block: induced 2026-08-21 by prepending text to this seat's prompt,
-- it aborted with "377 shared prefix FRAGMENTED — 2 distinct prefixes across
-- 17 marked seats".
--
-- Pair with 530. Running only one leaves the two rosters stating opposite
-- things about the same coordinator, which is the drift class this gate exists
-- to catch.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('council-gate',
  '531_diagnosis_guardian_error_step_precedence_council_gate.sql: pre-update');

DO $mig$
DECLARE
    v_marker     text := '<!--CACHE_BREAKPOINT-->';
    v_old_bullet text := '- CONFIG-LEVEL error_step: the workflow coordinator reads ONLY step.config.error_step -- a step-level error_step is parsed but silently inert (a real, recurring trap). Any plan adding error routing must place it inside config.';
    v_new_bullet text := '- error_step, BOTH LOCATIONS ARE LIVE: routeToErrorStepOrFail (coordinator.go) checks the STEP-LEVEL error_step FIRST -- the code comment there calls it the preferred location -- and falls back to step.config.error_step for backward compatibility. Neither is inert. Do NOT object to error routing on placement. Object where a plan REMOVES routing that a failure path needs, or points it at a step that swallows the error.';
    v_old_judge  text := '(d) does it place error_step outside config (silently inert), or move loop work/tokens onto shared pods.';
    v_new_judge  text := '(d) does it remove a step''s error routing, or point it at a step that swallows the failure (placement is NOT the test -- step-level and config-level are both honoured), or move loop work/tokens onto shared pods.';
    v_old_head   text := '## The author''s stated rationale loop''s load-bearing disciplines';
    v_new_head   text := '## The diagnosis loop''s load-bearing disciplines';
    v_old text;
    v_new text;
    v_mark_old int;
    v_rows int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_diagnosis_guardian'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL THEN
        RAISE EXCEPTION '531: no live council-gate row, or review_diagnosis_guardian has no prompt_template';
    END IF;

    v_mark_old := position(v_marker in v_old);
    IF v_mark_old = 0 THEN
        RAISE EXCEPTION '531: 377 cache breakpoint absent from this seat — stop and investigate before editing';
    END IF;

    -- Each anchor exactly once.
    IF (length(v_old) - length(replace(v_old, v_old_bullet, ''))) / length(v_old_bullet) <> 1 THEN
        RAISE EXCEPTION '531: the CONFIG-LEVEL error_step bullet does not occur exactly once — refusing to replace blind';
    END IF;
    IF (length(v_old) - length(replace(v_old, v_old_judge, ''))) / length(v_old_judge) <> 1 THEN
        RAISE EXCEPTION '531: judging clause (d) does not occur exactly once — refusing to replace blind';
    END IF;
    IF (length(v_old) - length(replace(v_old, v_old_head, ''))) / length(v_old_head) <> 1 THEN
        RAISE EXCEPTION '531: the mangled disciplines heading does not occur exactly once — refusing to replace blind';
    END IF;

    v_new := replace(replace(replace(v_old, v_old_bullet, v_new_bullet), v_old_judge, v_new_judge), v_old_head, v_new_head);

    -- Reverse-replacement control: back to the EXACT pre-image, or the edit
    -- touched something it should not have.
    IF replace(replace(replace(v_new, v_new_bullet, v_old_bullet), v_new_judge, v_old_judge), v_new_head, v_old_head)
       IS DISTINCT FROM v_old THEN
        RAISE EXCEPTION '531: reverse-replacement control failed — the edit is not confined to the three anchors';
    END IF;

    -- 377: every edit is downstream of the marker, so it must not move.
    IF position(v_marker in v_new) <> v_mark_old THEN
        RAISE EXCEPTION '531: the 377 cache breakpoint MOVED (% -> %) — refusing', v_mark_old, position(v_marker in v_new);
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','review_diagnosis_guardian','config','prompt_template'],
             to_jsonb(v_new), false),
           updated_at = now()
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '531: expected 1 row, updated %', v_rows;
    END IF;
END $mig$;

DO $verify$
DECLARE
    v_live text;
    v_marker text := '<!--CACHE_BREAKPOINT-->';
    v_seats int;
    v_prefixes int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_diagnosis_guardian'->'config'->>'prompt_template'
      INTO v_live
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('reads ONLY step.config.error_step' in v_live) > 0
       OR position('silently inert' in v_live) > 0 THEN
        RAISE EXCEPTION '531 VERIFY: the inverted claim is still live';
    END IF;
    IF position('checks the STEP-LEVEL error_step FIRST' in v_live) = 0
       OR position('placement is NOT the test' in v_live) = 0 THEN
        RAISE EXCEPTION '531 VERIFY: a corrected passage is missing';
    END IF;
    IF position('## The diagnosis loop''s load-bearing disciplines' in v_live) = 0 THEN
        RAISE EXCEPTION '531 VERIFY: the disciplines heading was not repaired';
    END IF;
    IF position('## The author''s stated rationale loop' in v_live) > 0 THEN
        RAISE EXCEPTION '531 VERIFY: the mangled heading is still present';
    END IF;
    IF position('THREE-TIER CITATIONS' in v_live) = 0
       OR position('TOKEN/POD ISOLATION' in v_live) = 0 THEN
        RAISE EXCEPTION '531 VERIFY: a neighbouring discipline was lost — the replace over-reached';
    END IF;
    -- The rationale CONTEXT block (the thing 099's replace legitimately makes)
    -- must survive: repairing the heading must not have taken it with it.
    IF position('{{.input_data.rationale}}' in v_live) = 0 THEN
        RAISE EXCEPTION '531 VERIFY: the submitter-rationale context block is gone';
    END IF;

    -- 377 health check, fleet-wide: 17 marked seats, ONE distinct shared prefix.
    SELECT count(*) INTO v_seats
      FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
     WHERE a.type = 'council-gate' AND a.is_active
       AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
       AND position(v_marker in COALESCE(s.value->'config'->>'prompt_template','')) > 0;
    SELECT count(DISTINCT split_part(s.value->'config'->>'prompt_template', v_marker, 1)) INTO v_prefixes
      FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
     WHERE a.type = 'council-gate' AND a.is_active
       AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
       AND position(v_marker in COALESCE(s.value->'config'->>'prompt_template','')) > 0;
    IF v_prefixes <> 1 THEN
        RAISE EXCEPTION '531 VERIFY: 377 shared prefix FRAGMENTED — % distinct prefixes across % marked seats', v_prefixes, v_seats;
    END IF;

    RAISE NOTICE '531 OK: council-gate.review_diagnosis_guardian corrected (error_step precedence + mangled heading); 377 intact — % marked seats, 1 distinct prefix.', v_seats;
END $verify$;

COMMIT;

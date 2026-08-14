-- 405_rfc022_budget_ruled_n10_gate_clause_without_reverting_377.sql
--
-- Mirrors 404 into council-gate.review_architecture BY HAND — same reason as
-- 383/403: 099_SYNC_gate_roster.py remains UNSAFE to --apply (its transform
-- predates migration 377). Surgical anchored replace in ONE seat, breakpoint
-- guarded. See 404's header for the ruling and the owner's framing correction.
--
-- ROLLBACK: 405_..._ROLLBACK.sql.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('council-gate',
  '405_rfc022_budget_ruled_n10_gate_clause_without_reverting_377.sql: pre-update');

DO $apply$
DECLARE
    v_anchor text := E'The budget N\n'
                  || E'  is not yet ruled, so until it is: if the plan shows the action ALREADY\n'
                  || E'  carries several optional keys, say so as "insufficient" with the count you\n'
                  || E'  actually observed, citing the counter, which can give you the exact figure.';
    v_insert text := E'The budget is\n'
                  || E'  RULED (owner, 2026-08-14): N = 10, on SHARED actions (2 or more live\n'
                  || E'  carriers). If the plan grows such an action''s optional-key set past 10, or\n'
                  || E'  grows one already past it, that ACCUMULATION is architecture-scope: signal\n'
                  || E'  needs_rfc with the count you actually observed, citing the counter\n'
                  || E'  (scripts/audit-optional-key-budget.sh). Sharing itself is estate design, not\n'
                  || E'  the defect (owner, same ruling): what is reviewed is the accumulated\n'
                  || E'  surface, never the reuse.';
    v_marker   text := '<!--CACHE_BREAKPOINT-->';
    v_old      text;
    v_new      text;
    v_rows     int;
    v_mark_old int;
    v_mark_new int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL THEN
        RAISE EXCEPTION '405: no live council-gate row, or review_architecture has no prompt_template';
    END IF;

    IF position('RULED (owner, 2026-08-14)' in v_old) > 0 THEN
        RAISE EXCEPTION '405: already applied — the ruled-budget sentence is present.';
    END IF;

    IF position(v_anchor in v_old) = 0 THEN
        RAISE EXCEPTION '405: 403''s closing sentences not found verbatim — the clause has been edited since. Re-read and re-anchor rather than forcing this.';
    END IF;

    v_mark_old := position(v_marker in v_old);
    IF v_mark_old = 0 THEN
        RAISE EXCEPTION '405: the 377 cache breakpoint is ABSENT from this seat. Something has already reverted 377 — stop and investigate before adding to it.';
    END IF;
    IF v_mark_old >= position(v_anchor in v_old) THEN
        RAISE EXCEPTION '405: the anchor precedes the cache breakpoint; replacing here WOULD change the cached prefix. Refusing.';
    END IF;

    v_new := replace(v_old, v_anchor, v_insert);
    IF v_new = v_old THEN
        RAISE EXCEPTION '405: replace() was a no-op despite the anchor being present';
    END IF;

    v_mark_new := position(v_marker in v_new);
    IF v_mark_new <> v_mark_old THEN
        RAISE EXCEPTION '405: the cache breakpoint MOVED (% -> %) — the cached prefix would be invalidated. Refusing.',
            v_mark_old, v_mark_new;
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','review_architecture','config','prompt_template'],
             to_jsonb(v_new), false),
           updated_at = now()
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '405: expected to update exactly 1 council-gate row, updated %', v_rows;
    END IF;

    RAISE NOTICE '405: ruled-budget sentence applied to council-gate.review_architecture (% -> % chars); cache breakpoint unmoved at %',
        length(v_old), length(v_new), v_mark_new;
END
$apply$;

-- Verify by RAISE, never by SELECT.
DO $verify$
DECLARE
    v_txt               text;
    v_fp                text;
    v_seats_with_marker int;
    v_distinct_prefixes int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_txt
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('RULED (owner, 2026-08-14)' in v_txt) = 0 THEN
        RAISE EXCEPTION '405 VERIFY: ruled-budget sentence absent after update';
    END IF;
    IF position('is not yet ruled' in v_txt) > 0 THEN
        RAISE EXCEPTION '405 VERIFY: the stale not-yet-ruled sentence survived the replace';
    END IF;
    IF position('NARROWED BY OWNER RULING 2026-08-11' in v_txt) = 0 THEN
        RAISE EXCEPTION '405 VERIFY: 383''s narrowing clause was lost';
    END IF;

    -- Cross-roster: 404 runs before this file (numeric order), so fix-proposer
    -- must already carry the ruled sentence or the rosters diverge.
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_fp
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_fp IS NULL OR position('RULED (owner, 2026-08-14)' in v_fp) = 0 THEN
        RAISE EXCEPTION '405 VERIFY: fix-proposer does not carry the ruled sentence — 404 has not been applied; the rosters would diverge';
    END IF;

    -- 377's invariants across ALL seats: marker everywhere, ONE shared prefix.
    SELECT count(*) FILTER (WHERE s.value->'config'->>'prompt_template' LIKE '%<!--CACHE_BREAKPOINT-->%'),
           count(DISTINCT md5(split_part(s.value->'config'->>'prompt_template', '<!--CACHE_BREAKPOINT-->', 1)))
      INTO v_seats_with_marker, v_distinct_prefixes
      FROM agent_definitions a,
           jsonb_each(a.default_config->'workflow'->'steps') s
     WHERE a.type = 'council-gate' AND a.is_active
       AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
       AND s.key LIKE 'review_%';

    IF v_seats_with_marker <> 17 THEN
        RAISE EXCEPTION '405 VERIFY: expected 17 seats carrying the 377 marker, found %', v_seats_with_marker;
    END IF;
    IF v_distinct_prefixes <> 1 THEN
        RAISE EXCEPTION '405 VERIFY: the cached prefix has FRAGMENTED into % distinct values — caching would stop paying', v_distinct_prefixes;
    END IF;

    RAISE NOTICE '405 VERIFY: ruled budget in both rosters; 377 intact — 17 seats marked, 1 shared prefix';
END
$verify$;

COMMIT;

-- DO NOT run 099_SYNC_gate_roster.py --apply after this file. See 383's header.

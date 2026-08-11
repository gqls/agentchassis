-- 383_rfc022_narrow_the_gate_architecture_seat_without_reverting_377.sql
--
-- Mirrors 381's RFC_022 narrowing into council-gate.review_architecture BY HAND,
-- deliberately, instead of running 099_SYNC_gate_roster.py — and this file is
-- mostly here to explain why that is the safe choice today and the mirror is not.
--
-- ⚠ 099_SYNC_gate_roster.py IS CURRENTLY UNSAFE TO --apply. It regenerates every
-- gate seat from fix-proposer through its own transform, and its transform does
-- NOT know about migration 377. 377 hoisted the 172-char shared evidence block
-- (schema + rationale + plan) to the FRONT of all 17 seats and inserted
-- <!--CACHE_BREAKPOINT--> after it, so Anthropic prompt caching — a PREFIX match
-- — can fire at all. Running the mirror would put that block back in the middle
-- of each template and drop the marker, reverting a change that is live and
-- measured at 68% saving in production (council_gate_cost lane, v1.0.1283).
-- Dry run on 2026-08-11 confirms it: `drift (steps that would change)` lists ALL
-- SEVENTEEN review_* steps, and a diff of one untouched seat (review_mission)
-- shows the difference is exactly 377's hoist, not a real roster divergence.
-- CLAUDE.md's "do not hand-patch the gate" exists because two hand-maintained
-- rosters drift; it assumes the mirror is faithful. Today it is not, and running
-- it to obey the letter of that rule would cause the regression the rule exists
-- to prevent. Filed so the mirror can be taught 377 and the rule restored.
--
-- WHY THIS FILE IS SAFE WHERE THE MIRROR IS NOT. It is a surgical insert at a
-- verbatim anchor, in ONE seat, and the anchor sits at char 1103 while the cache
-- breakpoint sits at 174 — so the cached PREFIX is not touched at all, and the
-- other 16 seats are not touched at all. The guards below assert both.
--
-- WHAT IT INSERTS: the identical clause 381 applied to fix-proposer, so the two
-- rosters say the same thing about the same trigger. OWNER RULING 2026-08-11,
-- RFC_022: option (3), a budget on the accumulated optional-key count, with
-- option (1) — an opt-in field, unsafe default OFF, named by no live consumer, is
-- NOT architecture-scope — as the interim. This file is the interim only.
--
-- ROLLBACK: 383_..._ROLLBACK.sql.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('council-gate',
  '383_rfc022_narrow_the_gate_architecture_seat_without_reverting_377.sql: pre-update');

DO $apply$
DECLARE
    v_anchor    text := 'relocate it. Say so via ARCHITECTURE_SIGNAL: needs_rfc. Two live precedents:';
    v_marker    text := '<!--CACHE_BREAKPOINT-->';
    v_insert    text;
    v_old       text;
    v_new       text;
    v_rows      int;
    v_mark_old  int;
    v_mark_new  int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL THEN
        RAISE EXCEPTION '383: no live council-gate row, or review_architecture has no prompt_template';
    END IF;

    IF position(v_anchor in v_old) = 0 THEN
        RAISE EXCEPTION '383: anchor line not found in the gate seat — re-read it and re-anchor rather than forcing this.';
    END IF;

    IF position('NARROWED BY OWNER RULING 2026-08-11' in v_old) > 0 THEN
        RAISE EXCEPTION '383: already applied — the narrowing clause is present.';
    END IF;

    v_mark_old := position(v_marker in v_old);
    IF v_mark_old = 0 THEN
        RAISE EXCEPTION '383: the 377 cache breakpoint is ABSENT from this seat. Something has already reverted 377 — stop and investigate before adding to it.';
    END IF;

    IF v_mark_old >= position(v_anchor in v_old) THEN
        RAISE EXCEPTION '383: the anchor precedes the cache breakpoint; inserting here WOULD change the cached prefix. Refusing.';
    END IF;

    -- Byte-identical to 381's clause, so the two rosters cannot say different
    -- things about the same trigger.
    v_insert := v_anchor || E'\n'
      || E'\n'
      || E'  NARROWED BY OWNER RULING 2026-08-11 (RFC_022) — ONE NARROW EXCEPTION.\n'
      || E'  An OPT-IN FIELD whose unsafe default is OFF, and which NO LIVE CONSUMER\n'
      || E'  NAMES, is NOT architecture-scope. Do not signal needs_rfc on that shape\n'
      || E'  alone. It is the owner''s own prescribed remedy for shipping new authority\n'
      || E'  safely (ruling 2026-08-02 §2), so flagging it penalises exactly the care it\n'
      || E'  was meant to buy, and a signal that fires on compliant work stops\n'
      || E'  discriminating between compliant and careless.\n'
      || E'  ALL THREE conditions must hold: opt-in; the UNSAFE side is the default;\n'
      || E'  and zero live consumers name it. If the author asserts the last one without\n'
      || E'  enumerating the consumers, THAT omission is your objection — not the field.\n'
      || E'  On-by-default, or behaviour changed for an existing caller, or a key another\n'
      || E'  workflow already names: architecture-scope exactly as before.\n'
      || E'  WHAT THIS RULING DELIBERATELY STOPS WATCHING IS ACCUMULATION. Ten\n'
      || E'  individually inert opt-in fields are a shared action nobody understands, and\n'
      || E'  this trigger was the only thing that would have noticed the tenth. The\n'
      || E'  ruling''s destination is a BUDGET on the accumulated optional-key COUNT, but\n'
      || E'  that counter is not built yet. So: if the plan shows the action ALREADY\n'
      || E'  carries several optional keys, say so as "insufficient" with the count you\n'
      || E'  actually observed. That reduced signal is the part still worth having.';

    v_new := replace(v_old, v_anchor, v_insert);

    IF v_new = v_old THEN
        RAISE EXCEPTION '383: replace() was a no-op despite the anchor being present';
    END IF;

    v_mark_new := position(v_marker in v_new);
    IF v_mark_new <> v_mark_old THEN
        RAISE EXCEPTION '383: the cache breakpoint MOVED (% -> %) — the cached prefix would be invalidated. Refusing.',
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
        RAISE EXCEPTION '383: expected to update exactly 1 council-gate row, updated %', v_rows;
    END IF;

    RAISE NOTICE '383: clause applied to council-gate.review_architecture (% -> % chars); cache breakpoint unmoved at %',
        length(v_old), length(v_new), v_mark_new;
END
$apply$;

-- Verify by RAISE, never by SELECT: under ON_ERROR_STOP a non-empty result set
-- does not abort the transaction, so a SELECT-shaped verify block lets a failed
-- migration COMMIT and report success.
DO $verify$
DECLARE
    v_seats_with_marker int;
    v_distinct_prefixes int;
    v_txt               text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_txt
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('NARROWED BY OWNER RULING 2026-08-11' in v_txt) = 0 THEN
        RAISE EXCEPTION '383 VERIFY: narrowing clause absent after update';
    END IF;
    IF position('It is architecture-scope even when it' in v_txt) = 0 THEN
        RAISE EXCEPTION '383 VERIFY: the original trigger text was lost';
    END IF;

    -- 377's invariants must be intact across ALL seats, not just this one: the
    -- marker present everywhere, and the cached prefix still SHARED (one
    -- distinct value). If the prefix fragmented, caching silently stops paying.
    SELECT count(*) FILTER (WHERE s.value->'config'->>'prompt_template' LIKE '%<!--CACHE_BREAKPOINT-->%'),
           count(DISTINCT md5(split_part(s.value->'config'->>'prompt_template', '<!--CACHE_BREAKPOINT-->', 1)))
      INTO v_seats_with_marker, v_distinct_prefixes
      FROM agent_definitions a,
           jsonb_each(a.default_config->'workflow'->'steps') s
     WHERE a.type = 'council-gate' AND a.is_active
       AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
       AND s.key LIKE 'review_%';

    IF v_seats_with_marker <> 17 THEN
        RAISE EXCEPTION '383 VERIFY: expected 17 seats carrying the 377 marker, found %', v_seats_with_marker;
    END IF;
    IF v_distinct_prefixes <> 1 THEN
        RAISE EXCEPTION '383 VERIFY: the cached prefix has FRAGMENTED into % distinct values — caching would stop paying', v_distinct_prefixes;
    END IF;

    RAISE NOTICE '383 VERIFY: clause present; 377 intact — 17 seats marked, 1 shared prefix';
END
$verify$;

COMMIT;

-- DO NOT run 099_SYNC_gate_roster.py --apply after this file. See the header.
-- Check drift with the DRY RUN only, and read any reported drift against 377
-- before believing it is a roster divergence.

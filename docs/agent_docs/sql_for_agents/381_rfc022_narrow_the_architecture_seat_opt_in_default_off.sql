-- 381_rfc022_narrow_the_architecture_seat_opt_in_default_off.sql
--
-- OWNER RULING 2026-08-11, RFC_022: option (3) — a BUDGET on the accumulated
-- optional-key count — WITH OPTION (1) AS THE INTERIM. This file is the interim
-- only. The destination needs a counter over RegisterActionInputSpec
-- declarations per action, which does not exist yet.
--
-- WHY. The owner's ruling of 2026-08-02 §2 prescribes that new authority on a
-- shared seam ships as an OPT-IN FIELD with the unsafe default OFF. The
-- architecture seat's trigger keys on the SHAPE — a new reserved key on a
-- widely-reused shared action — so it fires on exactly that remedy. Measured on
-- bugs_open/223 phase 1 (council 495df717): the seat signalled needs_rfc at
-- MEDIUM on a change that was opt-in, default-OFF, byte-identical when unset,
-- and named by 0 of append_doc_note's 8 live consumers, and it did not misread
-- anything. A signal that fires on compliant work stops discriminating between
-- careful and careless, which is the only property that made it worth having.
--
-- WHAT THIS DELIBERATELY GIVES UP, stated in the prompt itself so the seat
-- knows what it is no longer watching: ACCUMULATION. Ten individually inert
-- opt-in fields are a shared action nobody understands, and this trigger was
-- the only thing that would have noticed the tenth. The prompt therefore keeps
-- a reduced form of the signal — report an observed optional-key count as
-- "insufficient" — so the harm is visible rather than silent while the counter
-- is unbuilt. RFC_022's STATUS block says the same thing and names the exit.
--
-- SCOPE. fix-proposer ONLY. CLAUDE.md forbids hand-patching the gate:
-- 099_SYNC_gate_roster.py mirrors this into council-gate immediately after this
-- file, and the mirror is what keeps the two rosters identical. feature-designer
-- also seats review_architecture but its prompt does NOT carry this anchor (it
-- judges designs, not fixes) — deliberately untouched, no evidence it has the
-- same conflict.
--
-- SAFETY. The UPDATE is anchored on a verbatim line. If that line is absent —
-- because another session has rewritten this prompt — the DO block RAISEs and
-- the transaction aborts rather than silently applying to nothing. A verify
-- block of bare SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty
-- result), so the checks below are RAISE-based on purpose.
--
-- ROLLBACK: 381_..._ROLLBACK.sql, anchored on the inserted text.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('fix-proposer',
  '381_rfc022_narrow_the_architecture_seat_opt_in_default_off.sql: pre-update');

DO $apply$
DECLARE
    v_anchor  text := 'relocate it. Say so via ARCHITECTURE_SIGNAL: needs_rfc. Two live precedents:';
    v_insert  text;
    v_old     text;
    v_new     text;
    v_rows    int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL THEN
        RAISE EXCEPTION '381: no live fix-proposer row, or review_architecture has no prompt_template';
    END IF;

    IF position(v_anchor in v_old) = 0 THEN
        RAISE EXCEPTION '381: anchor line not found — the prompt has been rewritten since 2026-08-11. Re-read it and re-anchor rather than forcing this.';
    END IF;

    -- Idempotence: refuse a second application rather than stacking the clause.
    IF position('NARROWED BY OWNER RULING 2026-08-11' in v_old) > 0 THEN
        RAISE EXCEPTION '381: already applied — the narrowing clause is present.';
    END IF;

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
        RAISE EXCEPTION '381: replace() was a no-op despite the anchor being present';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','review_architecture','config','prompt_template'],
             to_jsonb(v_new), false),
           updated_at = now()
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '381: expected to update exactly 1 fix-proposer row, updated %', v_rows;
    END IF;

    RAISE NOTICE '381: narrowing clause applied to fix-proposer.review_architecture (% -> % chars)',
        length(v_old), length(v_new);
END
$apply$;

-- Verify by RAISE, not by SELECT: a non-empty result set does not abort a
-- transaction under ON_ERROR_STOP, so a SELECT-shaped "verify block" would let a
-- failed migration COMMIT and report success.
DO $verify$
DECLARE
    v_txt text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_txt
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('NARROWED BY OWNER RULING 2026-08-11' in v_txt) = 0 THEN
        RAISE EXCEPTION '381 VERIFY: narrowing clause absent after update';
    END IF;
    -- The clause must not have displaced what it qualifies.
    IF position('It is architecture-scope even when it' in v_txt) = 0 THEN
        RAISE EXCEPTION '381 VERIFY: the original trigger text was lost';
    END IF;
    IF position('ARCHITECTURE_SIGNAL: point_fix|needs_rfc|insufficient' in v_txt) = 0 THEN
        RAISE EXCEPTION '381 VERIFY: the routing-signal contract was lost';
    END IF;
    RAISE NOTICE '381 VERIFY: clause present, trigger text intact, signal contract intact';
END
$verify$;

COMMIT;

-- AFTER THIS FILE, RUN THE MIRROR — the gate is not hand-patched:
--   python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/099_SYNC_gate_roster.py --apply

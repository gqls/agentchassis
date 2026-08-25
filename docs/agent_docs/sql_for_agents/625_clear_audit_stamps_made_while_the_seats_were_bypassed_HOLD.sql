-- 625_clear_audit_stamps_made_while_the_seats_were_bypassed_HOLD.sql
--
-- _HOLD: applied BY HAND, immediately AFTER 624, with the 624 apply timestamp as
-- the window end:  psql -v WINDOW_END='<624 apply timestamp>' -f 625_...sql
--
-- WHY (found by the vigilant_designer_offer_analysis lane within two hours of the
-- 623 switch-on, verified first-hand here). Between migration 623 (2026-08-25
-- 21:18:19Z, the four model seats bypassed) and migration 624 (the acceptance
-- council armed), the improvement loop still ended its audit path at
-- record_audit_pass — so each site it visited was stamped
-- settings.maintenance_profile.last_audit = {fingerprint, at, passes} after only
-- the three MECHANICAL seats ran. The stamp then holds audit_due = false for 14
-- days (or until the fingerprint moves), so when 624 arms the council, every
-- bypass-stamped site reads as "audited at this fingerprint" and THE SEATS FIND
-- NOTHING DUE. A record that says the audit ran, when what ran did not do the
-- work — the same shape as gate-1c's false green, one level up. [MEASURED
-- 2026-08-25 21:4xZ] cookly.uk, at=21:20:23Z, passes=1 — one tick after enable.
--
-- WHAT THIS DOES: for exactly the bypass window (> 21:18:19Z, the 623+enable
-- moment, and < WINDOW_END, the 624 apply), it RECORDS each cleared stamp in a
-- doc_notes row (so nothing is lost and a reader can see what stood), then
-- REMOVES maintenance_profile.last_audit from those sites — the council sees
-- them as unaudited, which is what they are. passes_at_fingerprint goes with it:
-- those passes were counted by a loop running no judges.
--
-- WHAT IT DOES NOT DO: touch stamps from before the window (real, pre-bypass
-- audits with the model seats on) or after it (624-era audits, real again).

\if :{?WINDOW_END}
\else
\set WINDOW_END 'REFUSE'
\endif
DROP TABLE IF EXISTS _625_operator;
CREATE TEMP TABLE _625_operator AS SELECT :'WINDOW_END'::text AS window_end;

DO $probe$
DECLARE v_end text;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type='reader-experience-auditor' AND deleted_at IS NULL) THEN
        RAISE EXCEPTION '625: refuse - 624 is not applied (reader-experience-auditor absent); this file runs AFTER 624, window-ended at its apply time';
    END IF;
    SELECT window_end INTO v_end FROM _625_operator;
    IF v_end = 'REFUSE' THEN
        RAISE EXCEPTION '625: refuse - pass -v WINDOW_END=<the 624 apply timestamp>';
    END IF;
    PERFORM v_end::timestamptz; -- must parse
    IF EXISTS (SELECT 1 FROM doc_notes WHERE source='migration-625') THEN
        RAISE EXCEPTION '625: already applied - the cleared-stamps record exists in doc_notes';
    END IF;
END $probe$;

BEGIN;

DO $clear$
DECLARE
    v_end     timestamptz;
    v_cleared jsonb;
    v_n       int;
BEGIN
    SELECT window_end::timestamptz INTO v_end FROM _625_operator;

    SELECT COALESCE(jsonb_object_agg(s.domain, s.settings#>'{maintenance_profile,last_audit}'), '{}'::jsonb), count(*)
      INTO v_cleared, v_n
      FROM sites s
     WHERE (s.settings#>>'{maintenance_profile,last_audit,at}')::timestamptz > '2026-08-25 21:18:19+00'
       AND (s.settings#>>'{maintenance_profile,last_audit,at}')::timestamptz < v_end;

    INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
    VALUES ('decision', 'improvement-loop',
        'migration 625: cleared ' || v_n || ' bypass-era audit stamp(s) (623 applied 2026-08-25 21:18:19Z -> 624 at ' || v_end::text ||
        ') so the acceptance council sees those sites as unaudited. The stamps that stood, verbatim: ' || v_cleared::text,
        '["decision"]'::jsonb, 'migration-625', 'loanzy_uk_example_site');

    UPDATE sites s
       SET settings = s.settings #- '{maintenance_profile,last_audit}',
           updated_at = now()
     WHERE (s.settings#>>'{maintenance_profile,last_audit,at}')::timestamptz > '2026-08-25 21:18:19+00'
       AND (s.settings#>>'{maintenance_profile,last_audit,at}')::timestamptz < v_end;

    IF EXISTS (SELECT 1 FROM sites s
                WHERE (s.settings#>>'{maintenance_profile,last_audit,at}')::timestamptz > '2026-08-25 21:18:19+00'
                  AND (s.settings#>>'{maintenance_profile,last_audit,at}')::timestamptz < v_end) THEN
        RAISE EXCEPTION '625 verify: a bypass-window stamp survived';
    END IF;
    RAISE NOTICE '625: cleared % bypass-era stamp(s); the verbatim stamps are in the migration-625 doc_notes row', v_n;
END $clear$;

COMMIT;

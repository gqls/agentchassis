-- 690_refuse_untraceable_park_VERIFY.sql
--
-- Proves the park-provenance guard (migrations 690 + 700) is LIVE and DISCRIMINATING. bugs_open/396.
--
-- ⚠ REQUIRES MIGRATION 700. Assertion 5b fails against 690 alone — 690 accepted a handler
-- re-point on an already-deferred row, which is the hole 700 closes.
--
-- Run:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--         psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--         -f - < docs/agent_docs/sql_for_agents/690_refuse_untraceable_park_VERIFY.sql
--
-- Exit 0 = pass. Any assertion RAISEs, which aborts and exits non-zero.
--
-- ⚠ THIS FILE ENDS IN `ROLLBACK`. It creates two synthetic rows, drives five assertions
-- through them and discards everything. It touches no real work item and leaves no litter.
--
-- ⚠ WHY THIS IS NOT A `SELECT ... FROM pg_trigger` CHECK. That a trigger is ATTACHED says
-- nothing about what it does — it is the same mistake this bug's own lane made a week
-- earlier, when the guard it had nominated turned out to be a substring test that returned
-- "HONOURS" on four different spellings including two that switched the rule off entirely
-- (LANDMINES, `sites.locked_at`). So the structural check below is kept only as a
-- precondition, and the real verification INDUCES the refusal.
--
-- ⚠ AND WHY BOTH DIRECTIONS ARE ASSERTED. A guard that refuses EVERYTHING would pass a
-- one-sided "did it refuse?" test while breaking 2,656 live shelf rows and five producers.
-- Assertions 2-5 are the half that makes assertion 1 mean something.

BEGIN;

DO $v$
DECLARE
    v_site   uuid;
    v_id     uuid;
    v_id2    uuid;
    v_fired  boolean := false;
    v_key    text := 'MIGRATION_690_VERIFY_' || gen_random_uuid()::text;
    v_key2   text := 'MIGRATION_690_VERIFY_' || gen_random_uuid()::text;
    v_key3   text := 'MIGRATION_690_VERIFY_' || gen_random_uuid()::text;
    v_key4   text := 'MIGRATION_690_VERIFY_' || gen_random_uuid()::text;
    v_id3    uuid;
BEGIN
    -- PRECONDITION (structural, and NOT sufficient on its own — see the header).
    IF NOT EXISTS (SELECT 1 FROM pg_trigger
                    WHERE tgrelid = 'site_work_items'::regclass
                      AND tgname  = 'trg_site_work_items_park_provenance'
                      AND NOT tgisinternal) THEN
        RAISE EXCEPTION '690 VERIFY: trg_site_work_items_park_provenance is NOT attached — migration 690 is not applied (or was rolled back)';
    END IF;

    SELECT id INTO v_site FROM sites ORDER BY created_at LIMIT 1;
    IF v_site IS NULL THEN
        RAISE EXCEPTION '690 VERIFY: no sites row to hang a synthetic work item on';
    END IF;

    INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                 handler_agent, status, item_key)
    VALUES (v_site, 'migration_690_verify', 'migration_690_verify',
            'synthetic row for migration 690 VERIFY — this transaction is rolled back',
            'migration_690_verify', 'some-named-handler', 'triaged', v_key)
    RETURNING id INTO v_id;

    -- ── 1. THE INDUCED REFUSAL. An untraceable park must be REFUSED. ──
    BEGIN
        UPDATE site_work_items SET status = 'deferred' WHERE id = v_id;
    EXCEPTION WHEN OTHERS THEN
        v_fired := true;
        IF SQLERRM NOT LIKE '%WORK_ITEM_PARK_PROVENANCE_REFUSED%' THEN
            RAISE EXCEPTION '690 VERIFY: the write was refused, but NOT by this guard (%). A mutation that passes may have hit a guard in series — this is not proof.', SQLERRM;
        END IF;
    END;
    IF NOT v_fired THEN
        RAISE EXCEPTION '690 VERIFY FAILED: an untraceable park was ACCEPTED. The trigger is attached and INERT.';
    END IF;

    -- ── 2. parked_by WITHOUT parked_reason must ALSO be refused (both keys required). ──
    v_fired := false;
    BEGIN
        UPDATE site_work_items
           SET status = 'deferred',
               result = COALESCE(result,'{}'::jsonb) || jsonb_build_object('parked_by','verify')
         WHERE id = v_id;
    EXCEPTION WHEN OTHERS THEN
        v_fired := true;
        IF SQLERRM NOT LIKE '%WORK_ITEM_PARK_PROVENANCE_REFUSED%' THEN RAISE; END IF;
    END;
    IF NOT v_fired THEN
        RAISE EXCEPTION '690 VERIFY FAILED: a park carrying parked_by but NO parked_reason was accepted — half a provenance is not provenance';
    END IF;

    -- ── 3. THE DISCRIMINATION CONTROL: a park WITH provenance must be ACCEPTED. ──
    --      In `result` (park_work_items(), migration 621).
    BEGIN
        UPDATE site_work_items
           SET status = 'deferred',
               result = COALESCE(result,'{}'::jsonb) || jsonb_build_object(
                          'parked_by','migration_690_verify','parked_reason','verify run')
         WHERE id = v_id;
        RAISE EXCEPTION 'VERIFY_SENTINEL';
    EXCEPTION WHEN OTHERS THEN
        IF SQLERRM <> 'VERIFY_SENTINEL' THEN
            RAISE EXCEPTION '690 VERIFY FAILED: a park WITH provenance in result was REFUSED — this guard would break park_work_items(). Error: %', SQLERRM;
        END IF;
    END;

    --      And in `spec` (migration 389's location).
    BEGIN
        UPDATE site_work_items
           SET status = 'deferred',
               spec = COALESCE(spec,'{}'::jsonb) || jsonb_build_object(
                        'parked_by','migration_690_verify','parked_reason','verify run')
         WHERE id = v_id;
        RAISE EXCEPTION 'VERIFY_SENTINEL';
    EXCEPTION WHEN OTHERS THEN
        IF SQLERRM <> 'VERIFY_SENTINEL' THEN
            RAISE EXCEPTION '690 VERIFY FAILED: a park WITH provenance in spec was REFUSED — this guard would break migration 389''s shape. Error: %', SQLERRM;
        END IF;
    END;

    -- ── 4. THE SHELF CLASS (deferred + EMPTY handler, no provenance) must be UNTOUCHED. ──
    -- ⚠ INSERTED straight at 'deferred', not updated into it: CHECK constraint
    -- swi_no_handlerless_promotable forbids an empty handler_agent in
    -- triaged/approved/claimed, so a shelf row cannot pass through a promotable status.
    -- Born deferred is also what write_audit_findings_action actually does.
    BEGIN
        INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                     handler_agent, status, item_key)
        VALUES (v_site, 'migration_690_verify', 'migration_690_verify',
                'synthetic shelf row for migration 690 VERIFY', 'migration_690_verify',
                '', 'deferred', v_key2)
        RETURNING id INTO v_id2;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION '690 VERIFY FAILED: the SHELF class (deferred + empty handler) was REFUSED on INSERT. That is 2,656 live rows and five producers. Error: %', SQLERRM;
    END;

    -- ── 5. An ALREADY-deferred row must stay writable, or the retraction drain breaks. ──
    --
    -- ⚠⚠ CORRECTED 2026-09-02 (migration 700). THIS ASSERTION USED TO BE THE EXPLOIT.
    -- It previously set `handler_agent = 'some-named-handler'` on the shelf row and required
    -- the write to be ACCEPTED, calling that "the sharpest form" of proving an already-deferred
    -- row stays writable. It was sharp in the wrong direction: that write produces
    -- deferred + NAMED handler + no provenance — the exact shape this guard exists to prevent —
    -- and the test demanded it succeed. The council's `editquality` seat found it on the round
    -- that APPROVED 690 (corr `dcd2b3c9`), and it was then induced against the live trigger.
    -- Migration 700 closes it. **A test can assert a vulnerability is a feature; this one did.**
    --
    -- 5a. What the exemption is actually FOR: bookkeeping on a properly-parked row that leaves
    --     handler_agent alone must still be ACCEPTED.
    INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                 handler_agent, status, item_key, result)
    VALUES (v_site, 'migration_690_verify', 'migration_690_verify',
            'synthetic properly-parked row', 'migration_690_verify',
            'legacy-handler', 'deferred', v_key3,
            jsonb_build_object('parked_by','migration_690_verify','parked_reason','verify run'))
    RETURNING id INTO v_id3;
    BEGIN
        UPDATE site_work_items SET summary = summary || ' (bookkeeping)' WHERE id = v_id3;
        RAISE EXCEPTION 'VERIFY_SENTINEL';
    EXCEPTION WHEN OTHERS THEN
        IF SQLERRM <> 'VERIFY_SENTINEL' THEN
            RAISE EXCEPTION '690 VERIFY FAILED: bookkeeping on a properly-parked row was refused — this strands the 170 legacy rows work_item_retraction.go must drain. Error: %', SQLERRM;
        END IF;
    END;

    -- 5b. THE CORRECTED ARM (migration 700): re-pointing an already-deferred row's handler to a
    --     named value with NO provenance must be REFUSED. Under 690 alone this was ACCEPTED.
    v_fired := false;
    BEGIN
        UPDATE site_work_items SET handler_agent = 'some-named-handler' WHERE id = v_id2;
    EXCEPTION WHEN OTHERS THEN
        v_fired := true;
        IF SQLERRM NOT LIKE '%WORK_ITEM_PARK_PROVENANCE_REFUSED%' THEN RAISE; END IF;
    END;
    IF NOT v_fired THEN
        RAISE EXCEPTION '690/700 VERIFY FAILED: a shelf row was re-pointed to a NAMED handler with no provenance and ACCEPTED — migration 700 is not applied, or has regressed.';
    END IF;

    -- ── 6. THE INSERT ARM: a row BORN deferred with a named handler and no provenance
    --      must be refused, or the guard is bypassable by inserting rather than updating.
    v_fired := false;
    BEGIN
        INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                     handler_agent, status, item_key)
        VALUES (v_site, 'migration_690_verify', 'migration_690_verify',
                'synthetic born-parked row', 'migration_690_verify',
                'some-named-handler', 'deferred', v_key4);
    EXCEPTION WHEN OTHERS THEN
        v_fired := true;
        IF SQLERRM NOT LIKE '%WORK_ITEM_PARK_PROVENANCE_REFUSED%' THEN
            RAISE EXCEPTION '690 VERIFY: the INSERT was refused, but NOT by this guard (%).', SQLERRM;
        END IF;
    END;
    IF NOT v_fired THEN
        RAISE EXCEPTION '690 VERIFY FAILED: a row BORN deferred with a named handler and no provenance was ACCEPTED — the guard is bypassable by INSERT.';
    END IF;

    RAISE NOTICE '690 VERIFY PASSED — the guard is LIVE and DISCRIMINATING:';
    RAISE NOTICE '  1. untraceable park on UPDATE (named handler, no provenance) .. REFUSED';
    RAISE NOTICE '  2. parked_by without parked_reason ........................... REFUSED';
    RAISE NOTICE '  3. park with provenance, in result AND in spec ............... ACCEPTED';
    RAISE NOTICE '  4. shelf class (deferred + empty handler), born deferred ..... ACCEPTED';
    RAISE NOTICE '  5a. bookkeeping on a properly-parked row (handler unchanged) .. ACCEPTED';
    RAISE NOTICE '  5b. handler RE-POINT on an already-deferred row, no provenance . REFUSED';
    RAISE NOTICE '  6. untraceable park on INSERT (born parked) .................. REFUSED';
    RAISE NOTICE 'All synthetic rows are discarded by the ROLLBACK below.';
END
$v$;

ROLLBACK;

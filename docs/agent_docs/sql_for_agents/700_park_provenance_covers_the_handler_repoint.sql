-- 700_park_provenance_covers_the_handler_repoint.sql
--
-- Closes a REAL hole in migration `690` (WII-037), found by the council's `editquality` seat on
-- the round that APPROVED it (corr `dcd2b3c9`, medium-severity advisory objection). `bugs_open/396`.
--
-- ── THE HOLE, AND IT WAS CONFIRMED AGAINST THE LIVE TRIGGER BEFORE THIS FILE WAS WRITTEN ──
--
-- `690`'s third early exit reads:
--
--     IF TG_OP = 'UPDATE' AND OLD.status IS NOT DISTINCT FROM 'deferred' THEN RETURN NEW; END IF;
--
-- It exempts EVERY update to a row that was already `deferred` — including one that changes
-- `handler_agent`. So this reaches the forbidden state without ever going near `status`:
--
--     -- a legitimate shelf row: born deferred, empty handler, no provenance (2,656 of these)
--     INSERT INTO site_work_items (..., handler_agent, status) VALUES (..., '', 'deferred');
--     -- then simply re-point it. status never changes, so 690 never looks.
--     UPDATE site_work_items SET handler_agent = 'some-named-handler' WHERE id = ...;
--
-- The row is now `deferred` + NAMED handler + no provenance — **exactly the shape `690` exists to
-- make unrepresentable**, reached by a different entry path. Induced against the LIVE trigger
-- 2026-09-02 inside a rolled-back transaction: **ACCEPTED**. The seat was right.
--
-- ⚠⚠ **AND `690`'s OWN `_VERIFY` ASSERTED THIS WRITE AS CORRECT BEHAVIOUR.** Its assertion 5 takes
-- a shelf row, sets `handler_agent = 'some-named-handler'`, and requires the write to be ACCEPTED —
-- described in the file as "the sharpest form" of proving already-deferred rows stay writable. It
-- was simultaneously the exploit. A test can assert a vulnerability is a feature, and this one did.
-- `690_..._VERIFY.sql` is corrected in the same commit as this file.
--
-- ── THE FIX: ONE ADDED CONJUNCT ──
--
-- The already-deferred exemption now applies only when `handler_agent` is UNCHANGED:
--
--     AND NEW.handler_agent IS NOT DISTINCT FROM OLD.handler_agent
--
-- ── WHY THIS DOES NOT BREAK THE DRAIN, WHICH IS THE WHOLE REASON THE EXEMPTION EXISTS ──
--
-- The exemption was added so `work_item_retraction.go` could still close the 170 legacy rows that
-- carry no provenance, and so ordinary bookkeeping on a parked row keeps working. Both survive:
--
--   * `resolveWorkItems` CLOSES a row — it moves `status` OUT of `deferred`, so the FIRST exit
--     (`NEW.status IS DISTINCT FROM 'deferred'`) returns before any of this is reached, whatever
--     `handler_agent` does.
--   * any update that leaves `handler_agent` alone (summary, result, attempt_count, error,
--     updated_at) still passes, because the new conjunct is satisfied.
--   * DEMOTING a named row to the shelf (`handler_agent = ''`) still passes, on the SECOND exit —
--     an empty handler is the `bugs_closed/077` shape and is deliberately provenance-free.
--
-- So the only newly-refused write is: status stays `deferred`, AND `handler_agent` changes, AND the
-- new value is non-empty, AND the write carries no provenance. That is the hole and nothing else.
--
-- ── ANSWERING THE OTHER THREE ADVISORIES ON THE SAME ROUND, since they are cheap and checkable ──
--
--   * `guardian` / `prior_art_librarian` (medium) — "does `platform/livespec` need a companion
--     entry for this trigger?" **NO, and it is checked, not assumed.** livespec's only
--     `pg_trigger` probe (`livespec.go:258`, key `trigger_bindings.page_component_artefact_archive`)
--     is scoped `p.proname = 'page_component_artefact_archive'`. This trigger's function is
--     `refuse_untraceable_park`, so it cannot move that count. No other livespec entry enumerates
--     `site_work_items` triggers, and neither named lockstep test greps `pg_trigger` at all.
--   * `debug_historian` (medium) — "the induced refusal needs a SAVEPOINT or it poisons the
--     transaction." **It has one.** plpgsql `BEGIN ... EXCEPTION WHEN ... END` IS an implicit
--     subtransaction; that is the construct used, and the sketch elided it. The empirical proof is
--     that `690` APPLIED and COMMITTED on 2026-09-02 with its post-check NOTICE in the output — a
--     poisoned transaction could not have committed.
--   * `guardian` (low) — "a future producer writing this shape now hard-fails." True and accepted:
--     that is what a guard is. Mitigated by the one-statement withdrawal, which the refusal message
--     carries as its own `HINT`.

BEGIN;

-- ── PRECONDITIONS ─────────────────────────────────────────────────────────────────────
DO $pre$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger
                    WHERE tgrelid = 'site_work_items'::regclass
                      AND tgname  = 'trg_site_work_items_park_provenance'
                      AND NOT tgisinternal) THEN
        RAISE EXCEPTION '700: migration 690''s trigger is not attached — apply 690 first; this file only tightens its function';
    END IF;
    -- Refuse if 690's function is not the one we think we are replacing.
    IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'refuse_untraceable_park') THEN
        RAISE EXCEPTION '700: refuse_untraceable_park() does not exist';
    END IF;
END
$pre$;

-- ── THE TIGHTENED GUARD (only the third exit changes) ─────────────────────────────────
CREATE OR REPLACE FUNCTION refuse_untraceable_park() RETURNS trigger
LANGUAGE plpgsql AS $fn$
DECLARE
    v_by   text;
    v_why  text;
BEGIN
    -- Not a park at all. Also the arm that keeps the retraction drain working: closing a row
    -- moves status OUT of 'deferred', so it returns here whatever handler_agent does.
    IF NEW.status IS DISTINCT FROM 'deferred' THEN
        RETURN NEW;
    END IF;

    -- The shelf class (deferred + empty handler_agent) is a different mechanism and is
    -- deliberately provenance-free. bugs_closed/077. This also permits DEMOTING a named row
    -- back to the shelf.
    IF COALESCE(btrim(NEW.handler_agent), '') = '' THEN
        RETURN NEW;
    END IF;

    -- An already-parked row stays writable for ordinary bookkeeping — BUT ONLY IF THE HANDLER
    -- IS NOT BEING RE-POINTED. Without the second conjunct, a shelf row could be handed a named
    -- handler without ever touching `status`, reaching the forbidden shape unguarded (migration
    -- 700; council `dcd2b3c9` editquality objection, induced against the live trigger).
    IF TG_OP = 'UPDATE'
       AND OLD.status IS NOT DISTINCT FROM 'deferred'
       AND NEW.handler_agent IS NOT DISTINCT FROM OLD.handler_agent THEN
        RETURN NEW;
    END IF;

    -- Either sanctioned location. 389 writes spec, park_work_items() writes result.
    v_by  := COALESCE(NEW.spec ->> 'parked_by',     NEW.result ->> 'parked_by');
    v_why := COALESCE(NEW.spec ->> 'parked_reason', NEW.result ->> 'parked_reason');

    IF COALESCE(btrim(v_by), '') = '' OR COALESCE(btrim(v_why), '') = '' THEN
        RAISE EXCEPTION
          'WORK_ITEM_PARK_PROVENANCE_REFUSED: refusing to leave work item % (item_type=%, handler_agent=%) at status=''deferred'' with a named handler and no provenance. A deferred row with a named handler is selected by NOTHING and still holds its idx_swi_dedup slot, so an unattributable park is invisible and unre-filable (bugs_open/396). This fires on the transition INTO deferred AND on re-pointing handler_agent while already deferred (migration 700). Use SELECT park_work_items(...) — migration 621, which requires p_parked_by and p_parked_reason — or set parked_by AND parked_reason in spec or result in the same write. To shelf a finding instead, leave handler_agent empty (bugs_closed/077).',
          NEW.id, NEW.item_type, NEW.handler_agent
          USING HINT = 'This guard is trg_site_work_items_park_provenance (migrations 690 + 700). Withdraw fleet-wide with: DROP TRIGGER trg_site_work_items_park_provenance ON site_work_items;';
    END IF;

    RETURN NEW;
END
$fn$;

COMMENT ON FUNCTION refuse_untraceable_park() IS
  'bugs_open/396: refuses to leave a work item at status=''deferred'' with a NAMED handler_agent unless the write carries parked_by AND parked_reason in spec or result. Fires on the transition INTO deferred (migration 690) and on a handler_agent re-point while already deferred (migration 700, council dcd2b3c9). Does not touch deferred+empty-handler (the bugs_closed/077 shelf) or updates that leave handler_agent unchanged.';

-- ── POST-CHECK: the NEW arm, the arm it must NOT break, and 690's arms still standing ──
-- No outer EXCEPTION handler: a swallowed failure here would let the synthetic rows COMMIT as
-- litter. Every assertion is its own sub-block; anything unexpected aborts the whole migration.
DO $post$
DECLARE
    v_site  uuid;
    v_shelf uuid;
    v_named uuid;
    v_fired boolean;
    v_k1 text := 'MIGRATION_700_SELFTEST_' || gen_random_uuid()::text;
    v_k2 text := 'MIGRATION_700_SELFTEST_' || gen_random_uuid()::text;
BEGIN
    SELECT id INTO v_site FROM sites ORDER BY created_at LIMIT 1;
    IF v_site IS NULL THEN RAISE EXCEPTION '700 POST-CHECK: no sites row'; END IF;

    INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                 handler_agent, status, item_key)
    VALUES (v_site,'migration_700_selftest','migration_700_selftest','synthetic shelf row',
            'migration_700','', 'deferred', v_k1)
    RETURNING id INTO v_shelf;

    -- 1. THE NEW ARM — the hole this file closes. Re-pointing a shelf row's handler with no
    --    provenance must be REFUSED. Under migration 690 alone this was ACCEPTED.
    v_fired := false;
    BEGIN
        UPDATE site_work_items SET handler_agent = 'some-named-handler' WHERE id = v_shelf;
    EXCEPTION WHEN OTHERS THEN
        v_fired := true;
        IF SQLERRM NOT LIKE '%WORK_ITEM_PARK_PROVENANCE_REFUSED%' THEN
            RAISE EXCEPTION '700 POST-CHECK: refused, but not by this guard (%)', SQLERRM;
        END IF;
    END;
    IF NOT v_fired THEN
        RAISE EXCEPTION '700 POST-CHECK FAILED: the handler re-point was ACCEPTED — the hole this file exists to close is still open.';
    END IF;

    -- 2. The same re-point WITH provenance must be ACCEPTED (the discrimination control: an
    --    arm that refused both would pass assertion 1 while breaking every legitimate park).
    BEGIN
        UPDATE site_work_items
           SET handler_agent = 'some-named-handler',
               result = COALESCE(result,'{}'::jsonb) || jsonb_build_object(
                          'parked_by','migration_700_selftest','parked_reason','self-test')
         WHERE id = v_shelf;
        RAISE EXCEPTION 'SELFTEST_ROLLBACK_SENTINEL';
    EXCEPTION WHEN OTHERS THEN
        IF SQLERRM <> 'SELFTEST_ROLLBACK_SENTINEL' THEN
            RAISE EXCEPTION '700 POST-CHECK FAILED: a re-point WITH provenance was refused. Error: %', SQLERRM;
        END IF;
    END;

    -- 3. THE ARM THIS MUST NOT BREAK. A properly-parked row (deferred + NAMED + provenance) must
    --    stay writable for ordinary bookkeeping that leaves handler_agent alone — that is what the
    --    already-deferred exemption exists for, and narrowing it is the risk this file carries.
    INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                 handler_agent, status, item_key, result)
    VALUES (v_site,'migration_700_selftest','migration_700_selftest','synthetic parked row',
            'migration_700','legacy-handler','deferred', v_k2,
            jsonb_build_object('parked_by','migration_700_selftest','parked_reason','self-test'))
    RETURNING id INTO v_named;
    BEGIN
        UPDATE site_work_items SET summary = summary || ' (bookkeeping)' WHERE id = v_named;
        RAISE EXCEPTION 'SELFTEST_ROLLBACK_SENTINEL';
    EXCEPTION WHEN OTHERS THEN
        IF SQLERRM <> 'SELFTEST_ROLLBACK_SENTINEL' THEN
            RAISE EXCEPTION '700 POST-CHECK FAILED: bookkeeping on a properly-parked row was refused — this breaks the retraction drain and strands the legacy rows. Error: %', SQLERRM;
        END IF;
    END;

    -- 4. 690's INSERT arm must still refuse a row BORN deferred+named with no provenance.
    v_fired := false;
    BEGIN
        INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                     handler_agent, status, item_key)
        VALUES (v_site,'migration_700_selftest','migration_700_selftest','synthetic born-parked',
                'migration_700','some-named-handler','deferred',
                'MIGRATION_700_SELFTEST_' || gen_random_uuid()::text);
    EXCEPTION WHEN OTHERS THEN
        v_fired := true;
        IF SQLERRM NOT LIKE '%WORK_ITEM_PARK_PROVENANCE_REFUSED%' THEN RAISE; END IF;
    END;
    IF NOT v_fired THEN
        RAISE EXCEPTION '700 POST-CHECK FAILED: 690''s INSERT arm has regressed — a born-parked row with no provenance was accepted.';
    END IF;

    DELETE FROM site_work_items WHERE id IN (v_shelf, v_named);
    IF EXISTS (SELECT 1 FROM site_work_items WHERE item_key LIKE 'MIGRATION_700_SELFTEST_%') THEN
        RAISE EXCEPTION '700 POST-CHECK: synthetic rows survived cleanup — refusing to commit litter';
    END IF;

    RAISE NOTICE '700 POST-CHECK PASSED (4 assertions): handler re-point REFUSED without provenance (1) and ACCEPTED with it (2); bookkeeping on a properly-parked row still ACCEPTED (3); 690''s INSERT arm still REFUSES a born-parked row (4); synthetic rows removed.';
END
$post$;

COMMIT;

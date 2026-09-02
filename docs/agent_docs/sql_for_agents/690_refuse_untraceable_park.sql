-- 690_refuse_untraceable_park.sql
--
-- bugs_open/396: work items parked at status='deferred' WITH a named handler_agent are
-- selected by nothing (claim_work_item takes triaged/approved; the promoter takes detected)
-- and still hold their idx_swi_dedup slot — so the detector cannot re-file and any other
-- session dispatching that page hits 23505, a failure that reads as "already queued" and
-- means "queued and abandoned". One such row blocked bugs_open/328 for 22 days and
-- completed 2 minutes after being re-armed.
--
-- The bug's §6 candidate 1, ordered first because it makes the bad state UNREPRESENTABLE:
-- "A park writes parked_by + parked_reason or it does not happen — enforced where the write
-- happens, not by convention." Until now nothing enforced it: park_work_items() (migration
-- 621, WII-034) requires both, but NOTHING STOPS A RAW `UPDATE … SET status='deferred'`.
-- That is the standing residual on 396 and this file closes it.
--
-- ── SCOPE: NARROW, AND THE NARROWNESS IS MEASURED [MEASURED 2026-09-02] ──
--
-- This fires ONLY on `deferred` WITH A NAMED handler_agent. That is 396's shape and only
-- 396's shape. `deferred` + EMPTY handler_agent is a DIFFERENT, LEGITIMATE mechanism — the
-- "shelf" convention (bugs_closed/077): capability_gap rows, write_audit_findings_action's
-- filing_mode='record', discovery_checks/remit.go, check_palette_contrast,
-- check_content_duplication, check_missing_tools, cmd/verifier-remit-check. Those rows are
-- deliberately undispatchable and carry NO provenance, correctly.
--
--   deferred + EMPTY handler (shelf, legitimate) : 2,656 rows, 0 with provenance
--   deferred + NAMED handler (the 396 shape)     :   257 rows, 87 with, 170 WITHOUT
--
-- A trigger requiring provenance on every `deferred` write would refuse all 2,656 and break
-- five live producers. The census is what set this threshold, not the bug file.
--
-- ── WHY IT CANNOT BREAK LIVE GO CODE [MEASURED 2026-09-02] ──
--
-- Every Go writer of `deferred` writes it with an EMPTY handler_agent, and
-- write_build_items_routing_test.go ASSERTS that ("a deferred row naming an agent that does
-- not exist is bugs_closed/078's shape"). The two files that name `deferred` alongside a
-- handler are READERS, not writers: work_item_failure_ladder.go has it in a guard list of
-- statuses a failure/completion write must NOT overwrite, and work_item_retraction.go reads
-- it to count parks being drained. So no live Go path can trip this.
--
-- ── PROVENANCE IS ACCEPTED IN EITHER `spec` OR `result`, AND THAT IS NOT A CONVENIENCE ──
--
-- The two sanctioned writers disagree on location and BOTH are correct:
--   migration 389          writes spec.parked_by   / parked_reason / parked_from_status
--   park_work_items() (621) writes result.parked_by / parked_reason / parked_from_status
-- Reading only `spec` is the exact misstep recorded in 396's handoff §8.1 — 62 fully-stamped
-- rows were reported as "no trace of any kind". A guard that repeated it would refuse the
-- sanctioned verb.
--
-- ── ONLY THE TRANSITION INTO `deferred` ──
--
-- Gated on OLD.status IS DISTINCT FROM 'deferred'. An already-parked row stays updatable:
-- 170 legacy rows carry no provenance and work_item_retraction.go must still be able to
-- drain them. A trigger firing on every write to a deferred row would break the drain and
-- strand exactly the rows this bug is about.
--
-- ── WITHDRAWAL: ONE STATEMENT, NO ROLL, LIVE IMMEDIATELY ──
--
--   DROP TRIGGER trg_site_work_items_park_provenance ON site_work_items;
--
-- (or run 690_refuse_untraceable_park_ROLLBACK.sql, which also drops the function).
--
-- ── THIS FILE PROVES ITSELF BEFORE IT COMMITS ──
--
-- The post-check INDUCES the refusal on a synthetic row and aborts the migration if the
-- guard does not fire. A verify block made of SELECTs cannot stop a COMMIT — ON_ERROR_STOP
-- ignores a non-empty result set — so every assertion here is DO/RAISE (LANDMINES).
-- The synthetic rows are created and deleted inside this transaction and touch no real row.

BEGIN;

-- ── PRECONDITIONS ─────────────────────────────────────────────────────────────────────
DO $pre$
BEGIN
    IF to_regclass('public.site_work_items') IS NULL THEN
        RAISE EXCEPTION '690: site_work_items does not exist — wrong database';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_trigger
                WHERE tgrelid = 'site_work_items'::regclass
                  AND tgname  = 'trg_site_work_items_park_provenance'
                  AND NOT tgisinternal) THEN
        RAISE EXCEPTION '690: trg_site_work_items_park_provenance already exists — another session has applied this. STOP and read git log before re-applying.';
    END IF;
    -- The shelf class must exist and must be large. If it is ZERO, either the census is
    -- wrong or this is not the production database, and the scope reasoning above does
    -- not apply. Refuse rather than install a guard whose blast radius was never measured.
    IF (SELECT count(*) FROM site_work_items
         WHERE status = 'deferred' AND COALESCE(btrim(handler_agent),'') = '') = 0 THEN
        RAISE EXCEPTION '690: ZERO deferred+empty-handler rows found. The scope of this trigger was set by a census that measured 2,656. Re-run the census before applying.';
    END IF;
END
$pre$;

-- ── THE GUARD ─────────────────────────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION refuse_untraceable_park() RETURNS trigger
LANGUAGE plpgsql AS $fn$
DECLARE
    v_by   text;
    v_why  text;
BEGIN
    -- Not a park at all.
    IF NEW.status IS DISTINCT FROM 'deferred' THEN
        RETURN NEW;
    END IF;

    -- The shelf class (deferred + empty handler_agent) is a different mechanism and is
    -- deliberately provenance-free. bugs_closed/077. Leave it alone.
    IF COALESCE(btrim(NEW.handler_agent), '') = '' THEN
        RETURN NEW;
    END IF;

    -- Only the TRANSITION into deferred. Already-parked rows stay writable so that
    -- work_item_retraction.go can drain the 170 legacy rows that predate this guard.
    IF TG_OP = 'UPDATE' AND OLD.status IS NOT DISTINCT FROM 'deferred' THEN
        RETURN NEW;
    END IF;

    -- Either sanctioned location. 389 writes spec, park_work_items() writes result.
    v_by  := COALESCE(NEW.spec ->> 'parked_by',     NEW.result ->> 'parked_by');
    v_why := COALESCE(NEW.spec ->> 'parked_reason', NEW.result ->> 'parked_reason');

    IF COALESCE(btrim(v_by), '') = '' OR COALESCE(btrim(v_why), '') = '' THEN
        RAISE EXCEPTION
          'WORK_ITEM_PARK_PROVENANCE_REFUSED: refusing to park work item % (item_type=%, handler_agent=%) at status=''deferred'' with no provenance. A deferred row with a named handler is selected by NOTHING and still holds its idx_swi_dedup slot, so an unattributable park is invisible and unre-filable (bugs_open/396: 170 such rows exist and nobody can say who made them). Use SELECT park_work_items(...) — migration 621, which requires p_parked_by and p_parked_reason — or set parked_by AND parked_reason in spec or result in the same write. To shelf a finding instead, leave handler_agent empty (bugs_closed/077).',
          NEW.id, NEW.item_type, NEW.handler_agent
          USING HINT = 'This guard is trg_site_work_items_park_provenance (migration 690). Withdraw fleet-wide with: DROP TRIGGER trg_site_work_items_park_provenance ON site_work_items;';
    END IF;

    RETURN NEW;
END
$fn$;

COMMENT ON FUNCTION refuse_untraceable_park() IS
  'bugs_open/396: refuses a transition to status=''deferred'' with a NAMED handler_agent unless the same write carries parked_by AND parked_reason in spec or result. Does not touch deferred+empty-handler (the bugs_closed/077 shelf class, 2,656 rows as of 2026-09-02) and does not touch already-deferred rows. Migration 690.';

DROP TRIGGER IF EXISTS trg_site_work_items_park_provenance ON site_work_items;
CREATE TRIGGER trg_site_work_items_park_provenance
    BEFORE INSERT OR UPDATE ON site_work_items
    FOR EACH ROW EXECUTE FUNCTION refuse_untraceable_park();

-- ── POST-CHECK: INDUCE THE REFUSAL, AND PROVE THE ALLOWED SHAPES STILL PASS ───────────
DO $post$
DECLARE
    v_site   uuid;
    v_id     uuid;
    v_id2    uuid;
    v_fired  boolean := false;
    v_key    text := 'MIGRATION_690_SELFTEST_' || gen_random_uuid()::text;
    v_key2   text := 'MIGRATION_690_SELFTEST_' || gen_random_uuid()::text;
    v_key3   text := 'MIGRATION_690_SELFTEST_' || gen_random_uuid()::text;
BEGIN
    SELECT id INTO v_site FROM sites ORDER BY created_at LIMIT 1;
    IF v_site IS NULL THEN
        RAISE EXCEPTION '690 POST-CHECK: no sites row to hang a synthetic work item on';
    END IF;

    INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                 handler_agent, status, item_key)
    VALUES (v_site, 'migration_690_selftest', 'migration_690_selftest',
            'synthetic row for migration 690 self-test — deleted before COMMIT',
            'migration_690', 'some-named-handler', 'triaged', v_key)
    RETURNING id INTO v_id;

    -- ASSERTION 1 (the induced refusal — the whole point of the file).
    BEGIN
        UPDATE site_work_items SET status = 'deferred' WHERE id = v_id;
    EXCEPTION WHEN OTHERS THEN
        v_fired := true;
        IF SQLERRM NOT LIKE '%WORK_ITEM_PARK_PROVENANCE_REFUSED%' THEN
            RAISE EXCEPTION '690 POST-CHECK: a park was refused, but by something OTHER than this guard (%). A mutation that passes may have hit a guard in series — do not record this as proof.', SQLERRM;
        END IF;
    END;
    IF NOT v_fired THEN
        RAISE EXCEPTION '690 POST-CHECK FAILED: an untraceable park was ACCEPTED. The trigger is installed but inert — do not commit.';
    END IF;

    -- ASSERTION 2: the sanctioned shape passes, with provenance in RESULT (the 621 verb).
    BEGIN
        UPDATE site_work_items
           SET status = 'deferred',
               result = COALESCE(result,'{}'::jsonb) || jsonb_build_object(
                          'parked_by','migration_690_selftest','parked_reason','self-test')
         WHERE id = v_id;
        RAISE EXCEPTION 'SELFTEST_ROLLBACK_SENTINEL';
    EXCEPTION WHEN OTHERS THEN
        IF SQLERRM <> 'SELFTEST_ROLLBACK_SENTINEL' THEN
            RAISE EXCEPTION '690 POST-CHECK FAILED: a park WITH provenance in result was refused — the guard would break park_work_items(). Error: %', SQLERRM;
        END IF;
    END;

    -- ASSERTION 3: provenance in SPEC (migration 389's location) also passes.
    BEGIN
        UPDATE site_work_items
           SET status = 'deferred',
               spec = COALESCE(spec,'{}'::jsonb) || jsonb_build_object(
                        'parked_by','migration_690_selftest','parked_reason','self-test')
         WHERE id = v_id;
        RAISE EXCEPTION 'SELFTEST_ROLLBACK_SENTINEL';
    EXCEPTION WHEN OTHERS THEN
        IF SQLERRM <> 'SELFTEST_ROLLBACK_SENTINEL' THEN
            RAISE EXCEPTION '690 POST-CHECK FAILED: a park WITH provenance in spec was refused — the guard would break migration 389''s shape. Error: %', SQLERRM;
        END IF;
    END;

    -- ASSERTION 4: the shelf class (deferred + EMPTY handler, no provenance) still passes.
    --
    -- ⚠ It is INSERTED straight at 'deferred', not updated into it, and that is not a
    -- shortcut — CHECK constraint swi_no_handlerless_promotable forbids an empty
    -- handler_agent in triaged/approved/claimed, so a shelf row CANNOT pass through a
    -- promotable status. This assertion originally staged the row at 'triaged' and the
    -- dry run rejected it, which is how the constraint was found. Born deferred is also
    -- exactly what write_audit_findings_action and the discovery checks actually do.
    BEGIN
        INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                     handler_agent, status, item_key)
        VALUES (v_site, 'migration_690_selftest', 'migration_690_selftest',
                'synthetic shelf row for migration 690 self-test', 'migration_690',
                '', 'deferred', v_key2)
        RETURNING id INTO v_id2;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION '690 POST-CHECK FAILED: the SHELF class (deferred + empty handler) was REFUSED on INSERT. That is 2,656 live rows and five producers. Error: %', SQLERRM;
    END;

    -- ASSERTION 5: an ALREADY-deferred row stays updatable without provenance, so
    -- work_item_retraction.go can still drain the 170 legacy rows. Giving it a NAMED
    -- handler is the sharpest form: the row now matches the refused shape exactly, and
    -- must still be writable because it did not TRANSITION into deferred in this write.
    BEGIN
        UPDATE site_work_items
           SET handler_agent = 'some-named-handler', summary = summary || ' (touched)'
         WHERE id = v_id2;
        RAISE EXCEPTION 'SELFTEST_ROLLBACK_SENTINEL';
    EXCEPTION WHEN OTHERS THEN
        IF SQLERRM <> 'SELFTEST_ROLLBACK_SENTINEL' THEN
            RAISE EXCEPTION '690 POST-CHECK FAILED: an ALREADY-deferred row could not be updated. This breaks the retraction drain and strands the legacy rows. Error: %', SQLERRM;
        END IF;
    END;

    -- ASSERTION 6: the INSERT arm. A row BORN deferred with a named handler and no
    -- provenance must be refused too, or the guard is trivially bypassed by inserting
    -- rather than updating.
    v_fired := false;
    BEGIN
        INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                     handler_agent, status, item_key)
        VALUES (v_site, 'migration_690_selftest', 'migration_690_selftest',
                'synthetic born-parked row', 'migration_690',
                'some-named-handler', 'deferred', v_key3);
    EXCEPTION WHEN OTHERS THEN
        v_fired := true;
        IF SQLERRM NOT LIKE '%WORK_ITEM_PARK_PROVENANCE_REFUSED%' THEN
            RAISE EXCEPTION '690 POST-CHECK: the INSERT was refused, but NOT by this guard (%).', SQLERRM;
        END IF;
    END;
    IF NOT v_fired THEN
        RAISE EXCEPTION '690 POST-CHECK FAILED: a row BORN deferred with a named handler and no provenance was ACCEPTED — the guard is bypassable by INSERT.';
    END IF;

    DELETE FROM site_work_items WHERE id IN (v_id, v_id2);
    -- v_key3 is included deliberately: assertion 6 was SUPPOSED to be refused, so a row
    -- carrying that key would mean the guard let a born-parked row through AND that the
    -- assertion misread its own outcome.
    IF EXISTS (SELECT 1 FROM site_work_items WHERE item_key IN (v_key, v_key2, v_key3)) THEN
        RAISE EXCEPTION '690 POST-CHECK: synthetic self-test rows survived cleanup — refusing to commit litter';
    END IF;

    -- Accurate list of what was actually asserted above — the VERIFY sidecar adds the
    -- "parked_by without parked_reason" case, which is NOT tested here.
    RAISE NOTICE '690 POST-CHECK PASSED (6 assertions): untraceable park REFUSED on UPDATE (1) and on INSERT (6); provenance ACCEPTED in result (2) and in spec (3); shelf class ACCEPTED (4); already-deferred row still writable (5); synthetic rows removed.';
END
$post$;

COMMIT;

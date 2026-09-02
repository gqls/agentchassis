-- 700_park_provenance_covers_the_handler_repoint_ROLLBACK.sql
--
-- Reverts migration 700, restoring migration 690's looser function body. `bugs_open/396`.
--
-- ⚠ READ THIS BEFORE RUNNING IT. Rolling 700 back does NOT return you to a safe state — it
-- returns you to a state with a KNOWN, CONFIRMED hole. Under 690 alone, this succeeds:
--
--     UPDATE site_work_items SET handler_agent = 'some-named-handler'
--      WHERE id = <any row already at status='deferred'>;
--
-- leaving `deferred` + NAMED handler + no provenance — the exact shape the guard exists to
-- prevent, reached without touching `status`. Induced against the live trigger 2026-09-02:
-- ACCEPTED. Found by the council's `editquality` seat (corr `dcd2b3c9`).
--
-- **If what you actually want is the guard OFF, drop the trigger instead** — that is honest and
-- total, rather than leaving a guard that looks complete and is not:
--
--     DROP TRIGGER trg_site_work_items_park_provenance ON site_work_items;
--
-- Use THIS file only if migration 700 itself is causing a problem (i.e. a legitimate producer
-- re-points handler_agent on an already-deferred row and you need it working again TODAY) and you
-- want 690's protection to remain in place meanwhile. Record that you did, and why.
--
-- ⚠ Also revert `690_refuse_untraceable_park_VERIFY.sql` if you run this: its assertion 5b requires
-- 700 and will fail against 690 alone — correctly, and that failure is the signal that this
-- rollback is in force, not a broken test.

BEGIN;

DO $pre$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'refuse_untraceable_park') THEN
        RAISE EXCEPTION '700 ROLLBACK: refuse_untraceable_park() does not exist — nothing to revert';
    END IF;
    RAISE NOTICE '700 ROLLBACK: restoring migration 690 body. The handler-repoint hole will be OPEN again.';
END
$pre$;

CREATE OR REPLACE FUNCTION refuse_untraceable_park() RETURNS trigger
LANGUAGE plpgsql AS $fn$
DECLARE
    v_by   text;
    v_why  text;
BEGIN
    IF NEW.status IS DISTINCT FROM 'deferred' THEN
        RETURN NEW;
    END IF;

    IF COALESCE(btrim(NEW.handler_agent), '') = '' THEN
        RETURN NEW;
    END IF;

    -- 690's original third exit — no handler_agent conjunct. THIS IS THE HOLE.
    IF TG_OP = 'UPDATE' AND OLD.status IS NOT DISTINCT FROM 'deferred' THEN
        RETURN NEW;
    END IF;

    v_by  := COALESCE(NEW.spec ->> 'parked_by',     NEW.result ->> 'parked_by');
    v_why := COALESCE(NEW.spec ->> 'parked_reason', NEW.result ->> 'parked_reason');

    IF COALESCE(btrim(v_by), '') = '' OR COALESCE(btrim(v_why), '') = '' THEN
        RAISE EXCEPTION
          'WORK_ITEM_PARK_PROVENANCE_REFUSED: refusing to park work item % (item_type=%, handler_agent=%) at status=''deferred'' with no provenance (bugs_open/396). Use SELECT park_work_items(...) — migration 621 — or set parked_by AND parked_reason in spec or result in the same write. To shelf a finding instead, leave handler_agent empty (bugs_closed/077).',
          NEW.id, NEW.item_type, NEW.handler_agent
          USING HINT = 'This guard is trg_site_work_items_park_provenance (migration 690, with 700 ROLLED BACK — the handler-repoint path is unguarded). Withdraw entirely with: DROP TRIGGER trg_site_work_items_park_provenance ON site_work_items;';
    END IF;

    RETURN NEW;
END
$fn$;

-- POST-CHECK: prove the revert actually took, by INDUCING the write 700 refused and requiring it
-- to be ACCEPTED again. DO/RAISE, because a SELECT cannot stop a COMMIT.
DO $post$
DECLARE
    v_site uuid; v_id uuid; v_ok boolean := false;
    v_k text := 'MIGRATION_700_ROLLBACK_' || gen_random_uuid()::text;
BEGIN
    SELECT id INTO v_site FROM sites ORDER BY created_at LIMIT 1;
    INSERT INTO site_work_items (site_id, source, item_type, summary, created_by,
                                 handler_agent, status, item_key)
    VALUES (v_site,'migration_700_rollback','migration_700_rollback','synthetic revert probe',
            'migration_700_rollback','', 'deferred', v_k)
    RETURNING id INTO v_id;
    BEGIN
        UPDATE site_work_items SET handler_agent = 'some-named-handler' WHERE id = v_id;
        v_ok := true;
    EXCEPTION WHEN OTHERS THEN
        v_ok := false;
    END;
    DELETE FROM site_work_items WHERE id = v_id;
    IF NOT v_ok THEN
        RAISE EXCEPTION '700 ROLLBACK FAILED: the handler re-point is still refused — the 690 body did not take';
    END IF;
    RAISE NOTICE '700 ROLLBACK COMPLETE: 690 body restored. ⚠ The handler-repoint hole is OPEN again.';
END
$post$;

COMMIT;

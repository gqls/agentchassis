-- 621_park_work_items_verb_ROLLBACK.sql — reverses 621_park_work_items_verb.sql
--
-- Drops both functions. It deliberately does NOT un-park anything:
--
--   * rows parked THROUGH the verb are stamped (`result.parked_by`,
--     `parked_reason`, `parked_from_status`, `release_condition`), so they can be
--     released deliberately, by their holder, with a considered decision — which
--     is the entire point of the verb;
--   * blanket-releasing them here would fire real dispatches onto live customer
--     sites, which is the failure mode `bugs_open/396` §6a warns about and the
--     reason `unpark_work_items` is scoped to a single `p_parked_by`.
--
-- If you are rolling back because the verb misbehaved, the parked rows are still
-- fully attributable and can be released by hand from their own stamps:
--
--   SELECT id, result->>'parked_from_status', result->>'release_condition'
--     FROM site_work_items
--    WHERE status='deferred' AND result->>'parked_by_verb' = 'park_work_items';

BEGIN;

DROP FUNCTION IF EXISTS public.park_work_items(uuid,text,text,text,text[],uuid[],integer,boolean);
DROP FUNCTION IF EXISTS public.unpark_work_items(uuid,text,integer,boolean);

DO $guard$
BEGIN
    IF to_regprocedure('public.park_work_items(uuid,text,text,text,text[],uuid[],integer,boolean)') IS NOT NULL THEN
        RAISE EXCEPTION '621 ROLLBACK: park_work_items still present';
    END IF;
    IF to_regprocedure('public.unpark_work_items(uuid,text,integer,boolean)') IS NOT NULL THEN
        RAISE EXCEPTION '621 ROLLBACK: unpark_work_items still present';
    END IF;
    RAISE NOTICE '621 ROLLBACK OK: both functions dropped; parked rows left stamped and intact.';
END;
$guard$;

COMMIT;

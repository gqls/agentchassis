-- 321_review_queue_sweep_reaches_every_parked_row.sql
--
-- The daily review-queue sweep cannot reach 38% of the work it is able to judge.
-- This raises its per-pass cap so it can. It is a STOPGAP, deliberately, and the
-- structural fix is named at the bottom.
--
-- MEASURED 2026-08-06, live, before this file was written:
--
--   parked rows (status IN needs_human_review, unresolved) ......... 772
--   of which in a type reviewRevalidators can judge ................ 168
--   COVERED ROWS OUTSIDE THE OLDEST-500 HEAD — UNREACHABLE .......... 64
--     required_fields_missing 48 · needs_page 8 · needs_section_data 7
--     · unresolved_cta 1 · oldest filed 2026-07-24
--
-- WHY THEY ARE UNREACHABLE, AND WHY IT DOES NOT SELF-CORRECT.
-- loadParkedReviewItems selects `ORDER BY created_at ASC LIMIT max_items`, with
-- no type filter (revalidate_review_queue_action.go:301). 396 of the oldest 500
-- parked rows are in types with no revalidator; they return `unknown`, which is
-- deliberately NON-TERMINAL, so they stay parked, stay oldest, and are re-selected
-- every single run. Only ~104 head slots ever turn over. A covered row ranked
-- beyond 500 by created_at therefore never gets judged at all — not "slowly", never.
-- The first scheduled run measured the same thing from the other side:
--   scanned 500 · resolved 0 · still_holds 31 · unknown 469
-- i.e. 94% of each pass is work that cannot possibly resolve.
--
-- The starved rows are the NEWEST covered ones, which inverts the selection's own
-- stated rationale ("the oldest items are the ones most likely to be describing a
-- page state that no longer exists"): fresh findings are precisely the ones a
-- re-render is most likely to have already fixed.
--
-- WHY THE CAP AND NOT THE SCHEDULE. `max_items` is read from the STEP config
-- (GetIntField(config,"max_items",50)) and the sweep step has NO input_mapping, so
-- `scheduled_tasks.input_data` is inert — a previous session set 1000 there and the
-- run still reported capped_at 500. For the same reason "one scheduled row per
-- covered type" (the fix the 168-lane handoff recommended) CANNOT work as written:
-- every scheduled row would read this same step config and behave identically. The
-- `item_type` filter the action supports is also step config, not input_data.
--
-- BLAST RADIUS. One key on one step of one agent. dry_run is untouched (false, as
-- owner-approved on 2026-08-06). The revalidators are unchanged, so the JUDGEMENT
-- applied to the newly-reachable rows is the same one already applied daily to
-- their siblings — this changes WHICH rows are judged, not HOW. Nothing outside
-- the action reads result.revalidation or uncovered_types (grepped over
-- platform/, internal/, pkg/ 2026-08-06: no consumer).
--
-- COST. One UPDATE per scanned row per pass (gate 2 stamps every judged item), so
-- ~772/day instead of ~500/day, and a correspondingly larger `items` array in the
-- run's collected_data. Both are trivial; the waste is not the reason for the
-- structural fix, the starvation is.
--
-- NO ORDERING CONSTRAINT IS CLAIMED (2026-07-29 owner ruling retired that
-- condition). This is config: it is live the moment it commits, against the
-- running binary, which already reads max_items. Nothing needs a roll.
--
-- 1500, NOT 800. Parked rows accumulated 205 in the three days to 2026-08-06
-- against 55-81 in each of the prior weeks, so a cap sized to today's 772 could be
-- back in starvation within a week — and silent re-starvation is the exact failure
-- this file exists to end. The guard below asserts the cap still exceeds the live
-- parked count, so if this is ever applied into a queue that has outgrown it, it
-- FAILS LOUDLY instead of quietly under-reaching.
--
-- THE DURABLE FIX, which this does not do: filter the selection to the item_types
-- reviewRevalidators actually covers, and report the coverage gap with one
-- aggregate query instead of dragging ~600 unjudgeable rows through a per-row
-- UPDATE. Then the cap bounds judgeable work and cannot be consumed by a backlog
-- that is growing for unrelated reasons. That is a Go change to a shared action:
-- council gate, then a roll. Tracked in the bugfix_168_deployed_asset_path lane.

BEGIN;

SELECT snapshot_agent('diagnosis-review-queue-revalidator',
    '321: pre-update — raise sweep max_items 500 -> 1500; 64 covered rows were unreachable');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,sweep,config,max_items}',
        '1500'::jsonb,
        true),
    updated_at = NOW()
WHERE type = 'diagnosis-review-queue-revalidator'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    cfg        jsonb;
    parked     int;
    covered    int;
    starved    int;
BEGIN
    SELECT default_config #> '{workflow,steps,sweep}'
    INTO cfg FROM agent_definitions
    WHERE type = 'diagnosis-review-queue-revalidator'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '321: diagnosis-review-queue-revalidator has no sweep step';
    END IF;

    -- The cap only means anything if this is still the action that reads it.
    IF cfg #>> '{action}' IS DISTINCT FROM 'revalidate_review_queue' THEN
        RAISE EXCEPTION '321: sweep step action is %, expected revalidate_review_queue',
            COALESCE(cfg #>> '{action}', '<NULL>');
    END IF;

    -- IS DISTINCT FROM, not <>: a missing jsonb path is NULL and `NULL <> 'x'` is
    -- NULL, so a plain <> against an absent key can never fire.
    IF cfg #>> '{config,max_items}' IS DISTINCT FROM '1500' THEN
        RAISE EXCEPTION '321: max_items is %, expected 1500',
            COALESCE(cfg #>> '{config,max_items}', '<NULL>');
    END IF;

    -- dry_run must survive untouched. A sweep silently flipped to dry_run would
    -- look exactly like a sweep that had nothing to close.
    IF cfg #>> '{config,dry_run}' IS DISTINCT FROM 'false' THEN
        RAISE EXCEPTION '321: dry_run is %, expected false — this file must not disturb it',
            COALESCE(cfg #>> '{config,dry_run}', '<NULL>');
    END IF;

    SELECT count(*) INTO parked
    FROM site_work_items WHERE status IN ('needs_human_review', 'unresolved');

    SELECT count(*) INTO covered
    FROM site_work_items
    WHERE status IN ('needs_human_review', 'unresolved')
      AND item_type IN ('required_fields_missing', 'needs_section_data',
                        'unresolved_cta', 'needs_page');

    -- The point of the whole file: after this, zero covered rows sit beyond the cap.
    SELECT count(*) INTO starved FROM (
        SELECT row_number() OVER (ORDER BY created_at ASC) AS rn, item_type
        FROM site_work_items
        WHERE status IN ('needs_human_review', 'unresolved')
    ) q
    WHERE q.rn > 1500
      AND q.item_type IN ('required_fields_missing', 'needs_section_data',
                          'unresolved_cta', 'needs_page');

    IF parked >= 1500 THEN
        RAISE EXCEPTION '321: % parked rows already meets or exceeds the 1500 cap — '
            'this stopgap has been outgrown; do the selection filter instead '
            '(see the header, and the bugfix_168_deployed_asset_path lane)', parked;
    END IF;

    IF starved > 0 THEN
        RAISE EXCEPTION '321: % covered rows would STILL be beyond the cap — the '
            'premise of this file does not hold; re-measure before applying', starved;
    END IF;

    RAISE NOTICE '321 OK — cap 500 -> 1500; % parked, % judgeable, 0 now beyond the cap '
        '(was 64 at 2026-08-06); dry_run false, action intact', parked, covered;
END $$;

COMMIT;

-- ROLLBACK if needed:
--   UPDATE agent_definitions
--     SET default_config = jsonb_set(default_config,
--         '{workflow,steps,sweep,config,max_items}', '500'::jsonb, true)
--   WHERE type='diagnosis-review-queue-revalidator' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--
-- VERIFY AFTER THE NEXT SCHEDULED RUN (~daily; last_triggered_at is NOT proof it
-- ran — take the chain in the 168-lane handoff §0c):
--   SELECT collected_data #>> '{revalidation_result,scanned}'   AS scanned,
--          collected_data #>> '{revalidation_result,capped_at}' AS capped_at,
--          collected_data #>> '{revalidation_result,resolved}'  AS resolved,
--          collected_data #>> '{revalidation_result,unknown}'   AS unknown
--   FROM orchestration_states
--   WHERE orchestration_name ILIKE '%reval%' ORDER BY created_at DESC LIMIT 1;
-- scanned should now be the full parked count, not 500. If scanned = capped_at
-- again, the cap is binding once more and the selection filter is overdue.

-- 442_repair_flag_only_rows_blocked_by_the_old_promoter.sql
--
-- bugs_open/284, repair step. The pre-`7027a2801` promoter promoted EVERY
-- `detected` row on a site without looking at `handler_agent`; claim then found
-- them handler-less and stamped `blocked` with
-- "No handler_agent set — item cannot be routed to any agent". 60 rows across
-- four item_types are sitting in that state, describing a routing failure that
-- never was: these are flag-only findings whose producers deliberately file no
-- handler.
--
-- ORDERING — THIS FILE REQUIRES THE ROLLED BINARY, AND THAT IS NOW SATISFIED.
-- Repairing before the guard shipped would have re-blocked every row on the next
-- triage. Verified at the artefact 2026-08-16 before writing this file: chassis
-- AND core-manager both on v1.0.1305, image digests matching the local images,
-- OCI label `revision=6a782274b`, and `git merge-base --is-ancestor 7027a2801
-- 6a782274b` exits 0. The tag alone would not have been evidence.
--
-- TARGET STATES — each restores what the row's OWN producer files today, read
-- from the source, not chosen:
--   capability_gap  -> 'deferred'  (check_palette_contrast.go:138,
--                                   check_content_duplication.go:251)
--   image_url_404   -> 'detected'  (check_image_url_404.go:265, :306, :344)
-- ⚠ Both of those producers were changed to `deferred` BY 284's own fix commit
-- `7027a2801`. Reading them at HEAD and concluding "a deferred row cannot have
-- been promoted, so the diagnosis is wrong" is the trap — the rows were born
-- `detected` under the PRE-fix code (all 18 were filed 07-28..08-10, all blocked
-- 08-02..08-11, and no capability_gap row has been blocked since). The diagnosis
-- is sound; only the producers moved underneath it.
--
-- THE TWO HAND-INSERTED ROWS (no `spec.original_pipeline`, so not the promoter's)
-- are judged individually, per the bug file, and deliberately NOT lumped in:
--   * `verify_189_tool_loan_vs_savings_20260806` (page_rerender, 2026-08-06,
--     created_by bugfix-189-verify) -> 'cancelled'. A one-off verification
--     rerender for another lane's bug, ten days stale. Firing someone else's
--     lane's rerender now is not repair; if 189 still needs it, 189 re-files it.
--   * `needs_experience_plan:tools-are-unreachable-from-the-writing`
--     (fundamentallyai.com, 2026-08-12) -> 'deferred'. This one is OWNER-RAISED
--     ("raised_by": "owner, reading the live site 2026-08-12") with measured
--     evidence at the served page, and it has been silently unreachable for four
--     days. It is NOT cancelled. `[MEASURED]` no handler for
--     `needs_experience_plan` has ever existed — all 7 rows ever filed carry an
--     empty handler_agent (3 cancelled, 3 complete, this one) and no agent
--     definition or Go literal names the type — so it is a human-read item, and
--     `deferred` is the state `diagnose_triage`'s roadmap view surfaces
--     (`item_type='capability_gap' OR status='deferred'`). Flagged to the owner
--     in the lane docs rather than silently parked.
--
-- The stale `error` text is cleared (it describes a failure that never applied)
-- and the repair is recorded in `result.repair_284` so each row carries its own
-- provenance instead of looking spontaneously fixed.
--
-- ROLLBACK: 442_..._ROLLBACK.sql restores status='blocked' + the error text for
-- exactly the rows this file stamped (keyed on result.repair_284, not on a
-- re-derived predicate).

BEGIN;

DO $$
DECLARE n_cg int; n_404 int; n_hand int; n_total int; n_violators int;
BEGIN
    SELECT count(*) FILTER (WHERE item_type='capability_gap'),
           count(*) FILTER (WHERE item_type='image_url_404'),
           count(*) FILTER (WHERE NOT (spec ? 'original_pipeline')),
           count(*)
      INTO n_cg, n_404, n_hand, n_total
      FROM site_work_items
     WHERE status='blocked' AND COALESCE(handler_agent,'')='';

    IF n_total = 0 THEN
        RAISE EXCEPTION 'MIGRATION 442: no blocked flag-only rows — already applied (or already repaired by another hand)';
    END IF;

    -- COUNTED NEEDLES. Exact counts, because a repair that silently matches a
    -- different population than the one measured is the failure mode here.
    IF n_cg <> 18 OR n_404 <> 40 OR n_hand <> 2 OR n_total <> 60 THEN
        RAISE EXCEPTION 'MIGRATION 442: population moved since measurement (capability_gap=% want 18, image_url_404=% want 40, hand-inserted=% want 2, total=% want 60). Re-measure and re-derive this file rather than widening it.', n_cg, n_404, n_hand, n_total;
    END IF;

    -- The guard must already be live, or these rows re-block on the next triage.
    -- The DB cannot see the binary, so this asserts the observable consequence:
    -- nothing new has been blocked this way since the roll.
    SELECT count(*) INTO n_violators FROM site_work_items
     WHERE COALESCE(handler_agent,'')='' AND status IN ('triaged','approved','claimed');
    IF n_violators > 0 THEN
        RAISE EXCEPTION 'MIGRATION 442: % handler-less row(s) are in a promotable/claimed state right now — the old promoter is still live somewhere; do NOT repair yet', n_violators;
    END IF;
END $$;

WITH repaired AS (
    UPDATE site_work_items
       SET status = CASE
                      WHEN item_type = 'capability_gap'        THEN 'deferred'
                      WHEN item_type = 'image_url_404'         THEN 'detected'
                      WHEN item_type = 'needs_experience_plan' THEN 'deferred'
                      WHEN item_type = 'page_rerender'         THEN 'cancelled'
                    END,
           error = NULL,
           result = COALESCE(result,'{}'::jsonb) || jsonb_build_object(
                      'repair_284', jsonb_build_object(
                        'repaired_at', now()::text,
                        'from_status', 'blocked',
                        'from_error',  'No handler_agent set — item cannot be routed to any agent',
                        'why', 'the pre-7027a2801 promoter promoted this flag-only row and claim stamped it blocked; restored to the state its producer files (bugs_open/284)')),
           updated_at = now()
     WHERE status='blocked' AND COALESCE(handler_agent,'')=''
       AND item_type IN ('capability_gap','image_url_404','needs_experience_plan','page_rerender')
    RETURNING item_type, status
)
SELECT item_type, status, count(*) AS repaired FROM repaired GROUP BY 1,2 ORDER BY 3 DESC;

DO $$
DECLARE n_left int; n_cg_def int; n_404_det int; n_marked int;
BEGIN
    SELECT count(*) INTO n_left FROM site_work_items
     WHERE status='blocked' AND COALESCE(handler_agent,'')='';
    SELECT count(*) INTO n_cg_def FROM site_work_items
     WHERE item_type='capability_gap' AND status='deferred' AND result ? 'repair_284';
    SELECT count(*) INTO n_404_det FROM site_work_items
     WHERE item_type='image_url_404' AND status='detected' AND result ? 'repair_284';
    SELECT count(*) INTO n_marked FROM site_work_items WHERE result ? 'repair_284';

    IF n_left <> 0 OR n_cg_def <> 18 OR n_404_det <> 40 OR n_marked <> 60 THEN
        RAISE EXCEPTION 'MIGRATION 442: post-condition failed (still blocked=%, capability_gap deferred=%, image_url_404 detected=%, marked=%)', n_left, n_cg_def, n_404_det, n_marked;
    END IF;
    RAISE NOTICE 'migration 442 OK: 60 rows repaired and stamped result.repair_284 (18 deferred, 40 detected, 1 owner item deferred, 1 stale verification cancelled)';
END $$;

COMMIT;

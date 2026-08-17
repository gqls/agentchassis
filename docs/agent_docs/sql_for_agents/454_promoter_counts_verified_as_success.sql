-- 454 — the promoter's success tests must count `verified`, not only `complete`
--
-- A CORRECTION TO MY OWN MIGRATION 444 (bugfix_277/083 lane, same day),
-- found ~4 hours after applying it, by watching a pair change under me.
--
-- ============================================================================
-- THE DEFECT: A METRIC THAT DEGRADES AS THE PLATFORM SUCCEEDS
-- ============================================================================
-- `site_work_items.status` has TWO terminal success states, not one.
-- `complete_work_item_verification.go:218` writes `status='verified'` onto a row
-- that completed AND then passed verification — it is a completion with MORE
-- evidence behind it, not less. `idx_swi_completed`'s own predicate agrees:
-- `WHERE status IN ('complete','verified','rejected','wont_fix')`.
--
-- Both of the promoter's success tests count only `status='complete'`:
--   * 430's known-good rule  — EXISTS (… done.status = 'complete')
--   * 444's 25% success floor — count(*) FILTER (WHERE h.status = 'complete')
--
-- So a pair's apparent success rate FALLS as its successes get verified. The
-- better verification gets, the worse every handler looks to the gate that
-- decides whether to trust it. That is backwards.
--
-- [MEASURED 2026-08-17 16:30Z] caught on `empty_section -> page-build-handler`:
--     11:00Z census:  11 complete / 13 failed  = 46%  (passes the floor)
--     16:30Z census:   3 complete / 12 failed  = 20%  (HELD by the floor)
--   Nothing regressed. A verification sweep moved 9 of those completes to
--   `verified` in between. Counting both: 12 successes / 12 failed = 50%.
--   The pair is fine and 444 was about to hold 2 live rows for no reason.
--
-- FLEET SCOPE TODAY, so the size of this is not overstated: `verified` is rare —
-- 9 rows across exactly 1 pair. Exactly ONE pair changes verdict under this fix
-- (`empty_section`), and NO pair currently has zero completes with some verified,
-- so 430's known-good rule is not misfiring yet. It is LATENT there: the moment a
-- pair's only completes are verified, that pair reads as "never completed" and is
-- held for ever — the exact stranding bugs_open/083 exists to cure. Fixing both
-- predicates now, while the population is 9 rows, rather than after it grows.
--
-- ============================================================================
-- WHY THIS WAS MISSED, RECORDED SO THE NEXT AUTHOR DOES NOT REPEAT IT
-- ============================================================================
-- 444's census read the whole fleet's pair health and its numbers were correct
-- at the instant it ran. What it never asked was whether the QUANTITY IT WAS
-- COUNTING WAS THE WHOLE OF WHAT IT MEANT — every check I ran compared
-- `complete` against `failed` and never enumerated `status`'s domain. One
-- `GROUP BY status` over the table would have shown `verified` immediately.
-- Related and already recorded: `failed` rows carry no `completed_at`
-- (LANDMINES, 2026-08-17). Same table, same day, same family — a status column
-- whose values are assumed rather than enumerated.

BEGIN;

UPDATE scheduled_tasks
SET pre_query = $Q$
    WITH candidates AS (
        SELECT wi.id
        FROM site_work_items wi
        WHERE wi.status = 'detected'
          AND COALESCE(wi.handler_agent, '') <> ''
          -- DOOR-CLOSER 1 (444): only pipelines this transition has ever produced.
          AND wi.pipeline IN ('build', 'content', 'design')
          AND EXISTS (
            SELECT 1 FROM agent_definitions ad
            WHERE ad.type = wi.handler_agent
              AND ad.is_active
              AND COALESCE(ad.is_snapshot, false) = false
              AND ad.deleted_at IS NULL
          )
          -- KNOWN-GOOD (430, corrected by 454): `verified` is a completion that
          -- also passed verification. Counting only 'complete' would hold a pair
          -- whose every success has been verified.
          AND EXISTS (
            SELECT 1 FROM site_work_items done
            WHERE done.item_type = wi.item_type
              AND done.handler_agent = wi.handler_agent
              AND done.status IN ('complete', 'verified')
          )
          -- DOOR-CLOSER 2 (444, corrected by 454): >=5 terminal outcomes => must
          -- still be >=25% good. NB `failed` rows have no completed_at, so this
          -- keys on status only.
          AND (
            SELECT (c + f) < 5 OR c >= 0.25 * (c + f)
            FROM (
              SELECT count(*) FILTER (WHERE h.status IN ('complete', 'verified')) AS c,
                     count(*) FILTER (WHERE h.status = 'failed')                  AS f
              FROM site_work_items h
              WHERE h.item_type = wi.item_type
                AND h.handler_agent = wi.handler_agent
            ) hist
          )
        ORDER BY wi.created_at ASC
        LIMIT 20
    ),
    promoted AS (
        UPDATE site_work_items wi
        SET status = 'triaged',
            triaged_at = now(),
            spec = jsonb_set(COALESCE(wi.spec, '{}'::jsonb), '{original_pipeline}', to_jsonb(wi.pipeline)),
            pipeline = 'build',
            updated_at = now()
        FROM candidates c
        WHERE wi.id = c.id
          AND wi.status = 'detected'
        RETURNING wi.id, wi.item_type, wi.handler_agent
    )
    SELECT COUNT(*)::text AS promoted,
           string_agg(DISTINCT item_type || '->' || handler_agent, ', ') AS pairs
    FROM promoted
    WHERE (SELECT COUNT(*) FROM promoted) > 0
$Q$,
    updated_at = now()
WHERE name = 'detected-item-promoter';

-- ============================================================================
-- Verification. RAISE, not SELECT. The two controls must come out OPPOSITE
-- ways in the same run, or this file has either changed nothing or disabled
-- the floor entirely — the two ways it could be wrong.
-- ============================================================================
DO $$
DECLARE
    q            text;
    empty_passes boolean;
    lit_passes   boolean;
    n_held       int;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'detected-item-promoter';

    IF q NOT LIKE '%done.status IN (''complete'', ''verified'')%' THEN
        RAISE EXCEPTION '454: the known-good test still counts only complete';
    END IF;
    IF q NOT LIKE '%h.status IN (''complete'', ''verified'')%' THEN
        RAISE EXCEPTION '454: the success floor still counts only complete';
    END IF;
    -- 444's and 430's other predicates must have SURVIVED the rewrite
    IF q NOT LIKE '%wi.pipeline IN (''build'', ''content'', ''design'')%'
       OR q NOT LIKE '%0.25 * (c + f)%' OR q NOT LIKE '%(c + f) < 5%'
       OR q NOT LIKE '%LIMIT 20%' OR q NOT LIKE '%original_pipeline%'
       OR q NOT LIKE '%ad.is_active%' THEN
        RAISE EXCEPTION '454: the rewrite LOST a predicate from 430 or 444';
    END IF;

    -- POSITIVE CONTROL — the pair this fix is FOR must now pass the floor.
    SELECT (c + f) < 5 OR c >= 0.25 * (c + f) INTO empty_passes
      FROM (SELECT count(*) FILTER (WHERE status IN ('complete','verified')) AS c,
                   count(*) FILTER (WHERE status='failed') AS f
              FROM site_work_items
             WHERE item_type='empty_section' AND handler_agent='page-build-handler') x;
    IF NOT empty_passes THEN
        RAISE EXCEPTION '454: POSITIVE CONTROL FAILED — empty_section->page-build-handler still fails the floor counting verified, so this fix does not do what it claims';
    END IF;

    -- NEGATIVE CONTROL — the pair 444 was written to hold must STILL be held.
    -- If this passes, 454 has quietly disabled the floor rather than corrected it.
    SELECT (c + f) < 5 OR c >= 0.25 * (c + f) INTO lit_passes
      FROM (SELECT count(*) FILTER (WHERE status IN ('complete','verified')) AS c,
                   count(*) FILTER (WHERE status='failed') AS f
              FROM site_work_items
             WHERE item_type='literal_markdown' AND handler_agent='page-build-handler') x;
    IF lit_passes THEN
        RAISE EXCEPTION '454: NEGATIVE CONTROL FAILED — literal_markdown->page-build-handler now PASSES the floor, so counting verified has disabled the door 444 installed';
    END IF;

    SELECT count(*) INTO n_held
      FROM site_work_items wi
      WHERE wi.status='detected' AND COALESCE(wi.handler_agent,'')<>''
        AND wi.pipeline IN ('build','content','design')
        AND EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type=wi.item_type
                      AND d.handler_agent=wi.handler_agent AND d.status IN ('complete','verified'))
        AND NOT (SELECT (c+f)<5 OR c >= 0.25*(c+f) FROM (
              SELECT count(*) FILTER (WHERE h.status IN ('complete','verified')) AS c,
                     count(*) FILTER (WHERE h.status='failed') AS f
                FROM site_work_items h
               WHERE h.item_type=wi.item_type AND h.handler_agent=wi.handler_agent) hist);

    RAISE NOTICE '454: promoter now counts `verified` as success in BOTH tests. Controls opposite as required: empty_section->page-build-handler PASSES (was held at 20%%, is 50%% counting verified), literal_markdown->page-build-handler STILL HELD. Detected rows now held by the floor: %.', n_held;
END $$;

COMMIT;

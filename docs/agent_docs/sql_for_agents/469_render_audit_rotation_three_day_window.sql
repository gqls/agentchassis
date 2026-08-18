-- ============================================================================
-- 469 — cut the render-audit rotation's re-visit window from 7 days to 3
--       (owner instruction, 2026-08-18)
-- ============================================================================
-- WHY. The render audit is the only thing that GRADES a contrast repair: it
-- re-measures the page in a browser and withdraws the work item if the defect
-- has gone (bugs_open/296 §9 — 40 of the 226 parked rows drained this way on
-- the first pass). Its eligibility window is therefore the confirmation
-- latency of the whole repair loop: a fix that shipped today could wait up to
-- SEVEN DAYS to be graded. That is the bottleneck, NOT the hourly tick.
--
-- WHY 3 AND NOT 2 — the arithmetic, because "faster is better" has a floor.
--   25 eligible sites (status IN ('active','deployed'), measured 2026-08-18).
--     window 7d -> ~3.6 audits/day
--     window 3d -> ~8.3 audits/day
--     window 2d -> ~12.5 audits/day
--   NOMINAL capacity is 24/day (fires hourly, LIMIT 1 site).
--   EFFECTIVE capacity is much lower: the selector also requires
--   NOT EXISTS (a 'claimed' build work item for the site), which is sampled
--   ONCE PER HOUR against a value that flips on a ~20-SECOND timescale
--   (undeployed_asset items hold a claim ~25s of every ~26s during a build
--   burst). Polled read-only every 20s on 2026-08-17, NINE OF FOURTEEN samples
--   had EVERY due site blocked -> realistic throughput ~9 audits/day.
--   So a 2-day window sits AT or PAST real throughput and sites would begin
--   slipping their turn silently. 3 days is ~2.3x faster with headroom.
--   [SAMPLE: 14 x 20s during an active build burst — not a steady-state rate.]
--
-- COST. Audit volume rises ~2.3x. Each run is a headless browser render of the
-- site's pages, so this is a real spend increase and the owner has accepted it.
--
-- BLAST RADIUS. One literal in one scheduled_tasks row. No code, no schema.
-- Live immediately (DB config). Fully reversible: 469_..._ROLLBACK.sql.
-- Does NOT touch which sites are eligible, the LIMIT, or the claimed-build
-- guard — only how soon a site becomes due again.
-- ============================================================================

BEGIN;

-- Guard: refuse if the expected pre-state is absent. The message says
-- "already" so run-migrations.sh's probe reports LIKELY ALREADY APPLIED and
-- skips rather than halting the run (see its header).
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM scheduled_tasks
     WHERE name = 'site-render-audit-rotation' AND pre_query LIKE '%interval ''7 days''%';
    IF n <> 1 THEN
        RAISE EXCEPTION '469: site-render-audit-rotation does not carry the 7-day literal '
                        '(matched % rows) — already applied, or the pre_query has been '
                        'rewritten and this migration must be re-read before use', n;
    END IF;
END $$;

UPDATE scheduled_tasks
   SET pre_query  = replace(pre_query, 'interval ''7 days''', 'interval ''3 days'''),
       updated_at = now()
 WHERE name = 'site-render-audit-rotation';

-- Verify INSIDE the transaction, as a DO that RAISES. A verify block made of
-- bare SELECTs cannot stop the COMMIT — ON_ERROR_STOP ignores a non-empty
-- result set — which is a live trap recorded in LANDMINES.md.
DO $$
DECLARE bad int;
BEGIN
    SELECT count(*) INTO bad FROM scheduled_tasks
     WHERE name = 'site-render-audit-rotation'
       AND (pre_query LIKE '%interval ''7 days''%' OR pre_query NOT LIKE '%interval ''3 days''%');
    IF bad <> 0 THEN
        RAISE EXCEPTION '469 VERIFY FAILED: the rotation still carries a 7-day window or is '
                        'missing the 3-day one (% offending row(s)) — refusing to commit', bad;
    END IF;
END $$;

COMMIT;

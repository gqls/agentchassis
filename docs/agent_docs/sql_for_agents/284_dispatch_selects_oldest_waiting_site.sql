-- 284 — find_dispatchable_site picks the OLDEST-WAITING site, not the lowest UUID
--
-- WDS-002's named open lever ("fairness ORDER BY", recorded 2026-06-04), pulled
-- on the owner's direct instruction 2026-08-02 after the starvation was measured
-- live by the bugfix-154 lane. Delivered as config only: query_database runs
-- whatever string is here, so THERE IS NO IMAGE DEPENDENCY and no apply-order
-- constraint (contrast 278). Live the moment it commits.
--
-- WHY. The selector was:
--
--   SELECT DISTINCT ON (wi.site_id) wi.site_id::text, s.domain ...
--   ORDER BY wi.site_id, wi.priority ASC LIMIT 1
--
-- DISTINCT ON (site_id) FORCES the ORDER BY to lead with site_id, so the sort is
-- UUID-first and LIMIT 1 always takes the lowest-UUID eligible site. Not
-- "effectively arbitrary" (the register's old words) — DETERMINISTIC, which is
-- worse: arbitrary would eventually serve everyone. priority never influences
-- WHICH site wins; it only picks each site's representative row and is then
-- projected away, so cross-site priority was structurally unreachable. The only
-- thing that ever advanced the pointer down the UUID list was the NOT EXISTS
-- busy-exclusion. Measured 2026-08-02: 17 eligible sites; gamesdesign.co.uk
-- (uuid e33263f4…, 14th of 17) held the fleet's oldest eligible item — waiting
-- 3 days 10 hours — while robot-hands.com (uuid 00ff3af5…, 1st) won every idle
-- tick on a priority-110 item, ahead of a priority-5 item at mortgagecalculator.
-- For gamesdesign to be picked, all 13 lower-UUID sites had to be simultaneously
-- busy-or-drained. 33 claims went to 5 sites in 45 minutes; gamesdesign got none.
--
-- THE FIX. Order eligible ITEMS by age and take the top one's site:
--
--   ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1
--
-- - DISTINCT ON is dropped entirely: under LIMIT 1, the site owning the
--   globally-oldest eligible item IS the site whose oldest item is oldest.
--   Output shape unchanged (site_id, domain — one row), so check_has_site's
--   `dispatchable.count > 0` and call_dispatch's input_mapping are untouched.
-- - Starvation-free BY CONSTRUCTION: an item's key is fixed at creation and
--   every future item sorts after it (created_at = now() >= any waiter's), so
--   every site's oldest item eventually becomes the fleet minimum. Bounded
--   wait, not improved odds.
-- - Tiebreaks are load-bearing: batch-inserted items share one transaction
--   timestamp (created_at DEFAULT now()), so created_at alone is not a total
--   order. priority ASC then id ASC makes selection deterministic.
--
-- WHAT THIS DELIBERATELY DOES NOT DO. Cross-site priority stays unimplemented —
-- as it always was; no working semantics are removed. Any priority-major
-- cross-site ORDER BY recreates starvation keyed on priority instead of UUID,
-- and an aging scheme (e.g. created_at + priority * interval '1 minute') needs
-- an owner-agreed scale constant. If cross-site priority is ever wanted, that
-- is the shape — a policy decision, not a bug fix. Within-site, priority keeps
-- governing claim order (LoadWorkItemsAction: priority ASC, created_at ASC).
--
-- KNOWN BOUNDED BIAS, stated rather than discovered later: within-site claiming
-- is priority-major, so a site's oldest item may not be among the <=5 claimed on
-- its first pick; the site keeps its old key and can be re-picked whenever idle
-- until that item drains — at most ceil(backlog/5) batches. That biases service
-- toward the most-starved site, which is the correct direction, and it is
-- bounded, which the UUID order was not.
--
-- CONSUMERS TOLD (owner ruling 2026-07-29 #3): the only mechanical consumer is
-- build-pipeline-trigger's own downstream steps (shape unchanged). The
-- behavioural consumers are all 17 sites' queues: service order changes from
-- lowest-UUID-first to oldest-waiting-first. Immediately after apply, expect
-- gamesdesign.co.uk (3d10h), then robot-hands.com (1d20h), relojistas,
-- vetcomparison, webdesign — measured order in the bugfix_154 NOTES. Sites that
-- happened to sort early by UUID (robot-hands, loancalculator, finetuning) lose
-- an implicit precedence they were never promised.
--
-- COUNCIL: not submitted — the gate's scope is platform/, internal/, pkg/ and
-- refuses docs/config-only submissions client-side; this change has no Go half.
-- Owner authorised directly, 2026-08-02. Register updated in the same commit
-- (WDS-002).

BEGIN;

-- ── STEP 1 — PRE-FLIGHT ASSERTION ─────────────────────────────────────────
-- Derived at apply time, not carried in. Expect exactly 1: the live row whose
-- stored query is byte-identical (md5) to the one this file was written
-- against. Anything else means the row drifted — stop and re-read it.
SELECT count(*) AS rows_to_change_expect_1
FROM agent_definitions
WHERE type = 'build-pipeline-trigger'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND md5(default_config #>> '{workflow,steps,find_dispatchable_site,config,query}')
      = '92c8a59edcf826540de3d8e35ddd7ebb';

-- ── STEP 2 — SNAPSHOT ─────────────────────────────────────────────────────
SELECT snapshot_agent('build-pipeline-trigger',
                      'WDS-002 fairness ORDER BY — oldest-waiting site replaces lowest-UUID (284)');

-- ── STEP 3 — THE CHANGE ───────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,find_dispatchable_site,config,query}',
      $q$"SELECT wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.status IN ('triaged', 'approved') AND wi.attempt_count < wi.max_attempts AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = 'claimed') ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1"$q$::jsonb,
      false   -- create_missing=false: if the path is absent this is not the row
              -- the migration was written against; fail loud.
    ),
    updated_at = now()
WHERE type = 'build-pipeline-trigger'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND md5(default_config #>> '{workflow,steps,find_dispatchable_site,config,query}')
      = '92c8a59edcf826540de3d8e35ddd7ebb';   -- idempotent + refuses a drifted row

-- ── STEP 4 — VERIFY BEFORE COMMIT ─────────────────────────────────────────
-- Expect the query to contain 'ORDER BY wi.created_at' and no 'DISTINCT ON'.
SELECT type,
       default_config #>> '{workflow,steps,find_dispatchable_site,config,query}' AS query_after
FROM agent_definitions
WHERE type = 'build-pipeline-trigger'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

-- ── ROLLBACK ──
-- Preferred: restore the step-2 snapshot. Direct inverse, if unavailable:
-- UPDATE agent_definitions
-- SET default_config = jsonb_set(default_config,
--       '{workflow,steps,find_dispatchable_site,config,query}',
--       $q$"SELECT DISTINCT ON (wi.site_id) wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.status IN ('triaged', 'approved') AND wi.attempt_count < wi.max_attempts AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = 'claimed') ORDER BY wi.site_id, wi.priority ASC LIMIT 1"$q$::jsonb, false)
-- WHERE type='build-pipeline-trigger'
--   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── VERIFY THE FIX AT THE ARTEFACT, not at this row ──
-- Within a few ticks, claims should start following AGE order, not UUID order:
--
--   SELECT s.domain, wi.claimed_at
--   FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
--   WHERE wi.status = 'claimed' ORDER BY wi.claimed_at DESC LIMIT 10;
--
-- First expected pick (measured 2026-08-02): gamesdesign.co.uk — its oldest
-- eligible item had waited 3d10h, double the runner-up. If claims keep landing
-- on lowest-UUID sites while gamesdesign sits eligible and unclaimed, the row
-- reverted or the guard refused the UPDATE.

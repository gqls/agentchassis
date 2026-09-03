-- 744_evidence_register_absence_widen_status_predicate.sql
--
-- Widens `evidence-register-absence`'s (migration 742, CLM-033) site predicate from
-- `s.status = 'deployed'` to `s.status IN ('active','deployed')`.
--
-- WHY. 742's council round returned **REVISE** (corr `0d730d51-a923-4b44-a58f-ab8c898d7e22`)
-- and this was its sharpest objection: *"the pre_query scopes its population with
-- s.status='deployed', but this platform's documented lesson is that sites.status is
-- INFORMATIONAL ... blast-radius/coverage queries must never scope by it without first
-- enumerating GROUP BY status to confirm the value set means what you assume."*
--
-- Checked, and the objection is right in a way worth spelling out. **The estate already has a
-- convention for "a live site" and it is not `deployed` alone.** Every
-- `site-discovery-rotation-*` task — the machinery that dispatches per-site discovery checks —
-- scopes with `WHERE s.status IN ('active', 'deployed')`. 742 used the narrower half.
--
-- `[MEASURED 2026-09-03]` the census is `deployed` 39, `pool` 17, `test` 3, `system` 1 — there
-- are **no `active` sites today**, so widening changes the population by **ZERO** (12 before,
-- 12 after, verified in this migration's own verify block). **That is precisely why it is worth
-- doing now rather than later:** the change is free today and cannot be free once an `active`
-- site exists. And the failure it prevents is the exact class 742 was built to end — a site
-- that no mechanism can see. A detector with a narrower liveness predicate than the dispatcher
-- that feeds every other check would have re-created its own blind spot one status value over.
--
-- ⚠ IT DOES NOT WIDEN TO `pool`, `test` OR `system`, deliberately. `pool-*` sites carry no
-- content, `test` sites serve nothing (0 pages — and the `EXISTS (SELECT 1 FROM pages ...)`
-- clause already excludes them independently), and `system` is not a client site. RFC_060 D4's
-- "every site" means every site that SERVES something.
--
-- THE OTHER FIVE OBJECTIONS, dispositioned rather than waved through — full text in the lane
-- NOTES:
--   * [HIGH] the submission's edit entry named TWO files (the register entry and its index row).
--     A submission-format fault, not a defect in the migration; the resubmission splits them.
--     Worth noting it did NOT fire: admission passed, so the server-side refusal the landmine
--     warns about was not triggered — but the convention stands.
--   * [MED] "no evidence a search was run against discovery_checks/ for an existing absence
--     check." **ACCEPTED — I did not run that search, and I implied I had.** Run now: discovery
--     checks receive a `dctx.SiteID` and are dispatched per-site by the rotations above, so one
--     COULD have hosted this. The design choice still stands, but on narrower grounds than I
--     gave: the rotation takes **ONE site per tick with a 7-day per-site window**, so a
--     register-less site would be seen roughly weekly, against daily whole-fleet here; and a Go
--     check is inert until an image rolls. **What was wrong was implying no machinery existed.**
--   * [MED] review-queue crowding — `revalidate_review_queue` selects the oldest N across the
--     WHOLE parked queue, not per type. `[MEASURED 2026-09-03]` that queue is **1,436** deep;
--     these 12 items are **0.8%** of it and are the NEWEST rows, so an oldest-first sweep will
--     not reach them for a long time. Real interaction, quantified, no action taken.
--   * [MED] the measured-absence claim — re-derivable; the query is in CLM-033 and HANDOFF §8.
--   * [MED] "check whether `missing_evidence_register` already exists (dormant machinery risk)."
--     **Run: zero hits** across `*.go`, `*.sql` and `*.py` outside 742's own files. Nothing rebuilt.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/loancash_couk_fca_validation/
-- Register: CLM-033 (amended in the same commit)
-- Rollback: 744_..._ROLLBACK.sql

BEGIN;

-- GUARD 1: the task must exist and still carry the narrow predicate, or this is a no-op that
-- reads as applied.
DO $$
DECLARE pq text;
BEGIN
  SELECT pre_query INTO pq FROM scheduled_tasks WHERE name = 'evidence-register-absence';
  IF pq IS NULL THEN
    RAISE EXCEPTION '744 ABORT: scheduled task evidence-register-absence not found - migration 742 has not been applied here';
  END IF;
  IF pq NOT LIKE '%s.status = ''deployed''%' THEN
    RAISE EXCEPTION '744 ABORT: the narrow predicate is not present - already widened, or the pre_query has been edited by another hand; read it before writing';
  END IF;
END $$;

-- GUARD 2: the convention this migration adopts must still BE the convention. If the rotations
-- no longer scope on ('active','deployed'), copying that literal is cargo-culting a dead rule.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name LIKE 'site-discovery-rotation%'
     AND pre_query LIKE '%status IN (''active'', ''deployed'')%';
  IF n = 0 THEN
    RAISE EXCEPTION '744 ABORT: no site-discovery-rotation task scopes on status IN (active, deployed) any more - the convention being adopted here no longer exists; re-derive it before widening';
  END IF;
  RAISE NOTICE '744: % discovery rotation(s) confirm the (active, deployed) liveness convention', n;
END $$;

UPDATE scheduled_tasks
   SET pre_query = replace(pre_query, 's.status = ''deployed''', 's.status IN (''active'', ''deployed'')'),
       description = description ||
         ' WIDENED by migration 744 on the 742 council REVISE: liveness is status IN (active, deployed) - the '
         'convention every site-discovery-rotation uses - not deployed alone. Zero population change when applied '
         '(no active sites existed); the point is that it cannot be free later.',
       updated_at = now()
 WHERE name = 'evidence-register-absence';

-- VERIFY as DO/RAISE, and assert the POPULATION is unchanged - a widening that silently moved
-- the count would mean the two predicates disagree about more than I measured.
DO $$
DECLARE pq text; nnarrow int; nwide int;
BEGIN
  SELECT pre_query INTO pq FROM scheduled_tasks WHERE name = 'evidence-register-absence';
  IF pq NOT LIKE '%s.status IN (''active'', ''deployed'')%' THEN
    RAISE EXCEPTION '744 VERIFY: the widened predicate is not present after the update';
  END IF;
  IF pq LIKE '%s.status = ''deployed''%' THEN
    RAISE EXCEPTION '744 VERIFY: the narrow predicate is STILL present - the replace() matched nothing or matched partially';
  END IF;

  SELECT count(*) INTO nnarrow FROM sites s
    LEFT JOIN site_specs eb ON eb.site_id = s.id AND eb.aspect = 'evidence_base' AND eb.is_current
   WHERE s.status = 'deployed' AND eb.id IS NULL
     AND EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id);

  SELECT count(*) INTO nwide FROM sites s
    LEFT JOIN site_specs eb ON eb.site_id = s.id AND eb.aspect = 'evidence_base' AND eb.is_current
   WHERE s.status IN ('active','deployed') AND eb.id IS NULL
     AND EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id);

  IF nwide <> nnarrow THEN
    RAISE EXCEPTION '744 VERIFY: widening changed the population from % to % - that was NOT expected today (no active sites were measured); stop and read why before trusting either number', nnarrow, nwide;
  END IF;

  RAISE NOTICE '744 OK: evidence-register-absence now scopes on status IN (active, deployed); population unchanged at % - the fix is free today, which is the reason to take it now', nwide;
END $$;

COMMIT;

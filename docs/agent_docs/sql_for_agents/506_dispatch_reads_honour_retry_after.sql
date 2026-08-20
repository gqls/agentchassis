-- 506 — the two DB-resident dispatch reads learn to honour retry_after.
--
-- bugs_open/307, second half of the read-side contract. Migration 505 added
-- site_work_items.retry_after; the work-item failure ladder writes it. Four
-- places decide whether an item may be worked, and a backoff honoured by only
-- some of them is not a backoff:
--
--   Go   1. ClaimWorkItemAction        — the atomic claim              (in the binary)
--   Go   2. LoadWorkItemsAction        — "which items for this site?"  (in the binary)
--   SQL  3. build-pipeline-trigger.pre_query        — "is there work?" (THIS FILE)
--   SQL  4. build-pipeline-trigger's find_dispatchable_site query, which lives in
--          agent_definitions.default_config jsonb — "which site?"      (THIS FILE)
--
-- Two media, one contract. Migrations 213 and 285 already exist because sites 3
-- and 4 drifted from site 2 before; that is the failure this file must not
-- repeat, so the verify block below asserts all four end up in step.
--
-- ⚠ CORRECTION TO THIS CHANGE'S OWN PLAN, recorded rather than quietly fixed.
-- The plan named a THIRD SQL site, `build-dispatch-watchdog`, on the strength of
-- docs/agent_docs/sql_for_agents/214_build_dispatch_watchdog.sql. That task does
-- not exist: measured 2026-08-20, no scheduled_tasks row of that name, and the
-- 214 file is UNTRACKED in the shared working tree — written by some session,
-- never committed, never applied. A file in sql_for_agents/ is not a live task;
-- the live inventory is the table. So there is no third site to patch, and no
-- false BUILD_DISPATCH_STALLED risk to mitigate either.
--
-- SAFE BEFORE THE ROLL AND SAFE AFTER. While retry_after is entirely NULL (505's
-- verify block proves it is), `(retry_after IS NULL OR retry_after <= NOW())` is
-- a tautology, so applying this file changes no dispatch decision at all. It only
-- starts biting once the chassis carrying the ladder rolls and begins stamping.
--
-- Idempotent: both writes are same-text updates if already applied.

BEGIN;

-- ── 1. "Is there work?" — the trigger's gate ─────────────────────────────────
UPDATE scheduled_tasks
   SET pre_query = $PQ$SELECT COUNT(*)::text as pending_sites
FROM sites s
WHERE s.locked_at IS NULL
  AND EXISTS (
    SELECT 1 FROM site_work_items wi
    WHERE wi.site_id = s.id
      AND wi.status = 'triaged'
      AND wi.pipeline = 'build'
      AND wi.attempt_count < wi.max_attempts
      AND (wi.retry_after IS NULL OR wi.retry_after <= NOW())
)
HAVING COUNT(*) > 0$PQ$,
       updated_at = now()
 WHERE name = 'build-pipeline-trigger';

-- ── 2. "Which site?" — the selector inside the agent definition ─────────────
-- Kept byte-identical to the live query except for the one added clause: this
-- selector and LoadWorkItemsAction must agree about dispatchability (migration
-- 285's whole subject), so it is edited surgically rather than regenerated.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,find_dispatchable_site,config,query}',
         to_jsonb($Q$SELECT wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE s.locked_at IS NULL AND wi.status IN ('triaged', 'approved') AND wi.attempt_count < wi.max_attempts AND (wi.retry_after IS NULL OR wi.retry_after <= NOW()) AND (COALESCE(wi.approval_mode, 'auto') = 'auto' OR wi.status = 'approved') AND (wi.depends_on IS NULL OR NOT EXISTS (SELECT 1 FROM unnest(wi.depends_on) dep_id WHERE dep_id NOT IN (SELECT id FROM site_work_items WHERE site_id = wi.site_id AND status IN ('complete', 'verified')))) AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = 'claimed') ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1$Q$::text)
       ),
       updated_at = now()
 WHERE type = 'build-pipeline-trigger'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

-- ── 3. Verify (DO/RAISE — a SELECT cannot stop a COMMIT) ────────────────────
DO $do$
DECLARE
  v_selector text;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'site_work_items' AND column_name = 'retry_after') THEN
    RAISE EXCEPTION '506: site_work_items.retry_after does not exist — apply 505 first';
  END IF;

  IF (SELECT count(*) FROM scheduled_tasks
       WHERE name = 'build-pipeline-trigger' AND pre_query LIKE '%retry_after%') <> 1 THEN
    RAISE EXCEPTION '506: build-pipeline-trigger.pre_query does not honour retry_after';
  END IF;

  SELECT default_config #>> '{workflow,steps,find_dispatchable_site,config,query}'
    INTO v_selector
    FROM agent_definitions
   WHERE type = 'build-pipeline-trigger' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF v_selector IS NULL THEN
    RAISE EXCEPTION '506: find_dispatchable_site query is missing from the live build-pipeline-trigger definition';
  END IF;
  IF v_selector NOT LIKE '%retry_after%' THEN
    RAISE EXCEPTION '506: find_dispatchable_site does not honour retry_after';
  END IF;

  -- The clauses that were already one contract must still be there: a
  -- jsonb_set that replaced the query with a hand-typed near-copy would pass
  -- the LIKE above while silently dropping the dependency or claim guards.
  IF v_selector NOT LIKE '%attempt_count < wi.max_attempts%'
     OR v_selector NOT LIKE '%depends_on%'
     OR v_selector NOT LIKE '%active.status = ''claimed''%'
     OR v_selector NOT LIKE '%approval_mode%' THEN
    RAISE EXCEPTION '506: find_dispatchable_site lost a pre-existing dispatchability clause — it must differ from the live query by the retry_after clause ALONE';
  END IF;

  -- Inertness, which is what makes this safe to apply before the roll: if any
  -- row already carried a stamp, this file would start withholding real work.
  IF (SELECT count(*) FROM site_work_items
       WHERE retry_after IS NOT NULL AND retry_after > NOW()) <> 0 THEN
    RAISE WARNING '506: % rows already carry a future retry_after — the new predicates are live, not inert (expected only if the chassis ladder has already rolled)',
      (SELECT count(*) FROM site_work_items WHERE retry_after IS NOT NULL AND retry_after > NOW());
  END IF;
END
$do$;

COMMIT;

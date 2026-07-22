-- 193_load_existing_pages_site_has_no_current_plan.sql
--
-- bugs_open/051 — the "adoption_locked" flag is a MISLEADING NAME. There is no
-- per-page or 90-day lock, and there never was: the live load_existing_pages
-- query derives the flag per SITE as "this site has no current plan", so it is
-- true only on a site's FIRST plan after adoption and false on every re-plan.
-- The premise corrections landed first (comments in v3_site_actions.go, the 053
-- design doc, and bugs_closed/001). This migration is the DB half of retiring the
-- name (owner-approved 2026-07-21, "candidate 3"): rename the wire key the query
-- emits so nobody reasons about a "lock" again.
--
-- ORDER-FREE / TRANSITION. This step emits BOTH the new name
-- `site_has_no_current_plan` (primary) AND the old `adoption_locked` (a deprecated
-- transition alias). Both are the SAME boolean (same CASE). Emitting both makes
-- this safe to apply against the CURRENT chassis, which still reads
-- `adoption_locked` in reconcilePlanWithRealised — that read keeps working
-- unchanged, and the extra column is ignored by every consumer that does not read
-- it (exactly as migration 173's build_status column was). The Go half — reading
-- `site_has_no_current_plan` with an `adoption_locked` fallback — is HELD until
-- reconcilePlanWithRealised is out of the /bugs_open/037 rework; the ready plan is
-- in the 051 bug file.
--
-- FOLLOW-UP (do NOT do here): once a chassis carrying the renamed reader is
-- fleet-live (verify `strings /app/agent-chassis | grep site_has_no_current_plan`),
-- a final migration DROPS the `adoption_locked` transition alias and a Go commit
-- drops the fallback. Only then is the concept fully retired.
--
-- NOTE the query below is the LIVE query as read from agent_definitions on
-- 2026-07-22 (single-branch adoption_locked + build_status, i.e. the 173 shape),
-- with the new alias column added. It is NOT rebuilt from the 053 §054 block.

BEGIN;

SELECT snapshot_agent('build-site-planner',
                      '193_load_existing_pages_site_has_no_current_plan: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_pages,config,query}',
        to_jsonb($Q$
SELECT p.name, p.page_type, p.url, p.title, p.nav_label, p.in_header, p.in_footer,
       p.sections, p.meta_description, p.nav_order, p.build_status,
       CASE
         WHEN NOT EXISTS (
             SELECT 1 FROM site_plans sp
             WHERE sp.site_id = p.site_id AND sp.is_current = true
         ) THEN true
         ELSE false
       END AS site_has_no_current_plan,
       CASE
         WHEN NOT EXISTS (
             SELECT 1 FROM site_plans sp
             WHERE sp.site_id = p.site_id AND sp.is_current = true
         ) THEN true
         ELSE false
       END AS adoption_locked
FROM pages p
WHERE p.site_id = $1 AND p.status = 'active'
ORDER BY p.name
$Q$::text),
        true
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- Guard: exactly one active build-site-planner row must carry the new alias now.
DO $$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n
  FROM agent_definitions
  WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true
    AND (default_config->'workflow'->'steps'->'load_existing_pages'->'config'->>'query')
        LIKE '%site_has_no_current_plan%'
    AND (default_config->'workflow'->'steps'->'load_existing_pages'->'config'->>'query')
        LIKE '%adoption_locked%';
  IF n <> 1 THEN
    RAISE EXCEPTION '193 guard: expected exactly 1 active build-site-planner emitting both aliases, found %', n;
  END IF;
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- Verification
-- ---------------------------------------------------------------------------
-- 1. The step query now carries BOTH aliases:
--    SELECT (q LIKE '%site_has_no_current_plan%') AS has_new,
--           (q LIKE '%adoption_locked%')          AS has_old
--    FROM (SELECT default_config->'workflow'->'steps'->'load_existing_pages'->'config'->>'query' AS q
--          FROM agent_definitions WHERE type='build-site-planner' AND is_active=true) x;
--    -- expect: has_new = t, has_old = t
--
-- 2. The query still runs and both columns agree on a real site (they are the
--    same CASE, so they must be identical row-for-row):
--    SELECT p.name,
--           (NOT EXISTS (SELECT 1 FROM site_plans sp WHERE sp.site_id=p.site_id AND sp.is_current)) AS flag
--    FROM pages p WHERE p.site_id = '<site>' AND p.status='active' ORDER BY p.name;
--
-- 3. Behaviour is unchanged: the running chassis reads adoption_locked, which is
--    still emitted, so preservation on a first plan is exactly as before.

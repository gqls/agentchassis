-- 194_load_existing_pages_drop_adoption_locked_alias.sql
--
-- bugs_open/051 — FINAL step of retiring the misleading "adoption_locked" name.
-- Migration 193 renamed the wire key by emitting BOTH site_has_no_current_plan
-- (new) and adoption_locked (a transition alias) while the chassis still read the
-- old name. The renamed chassis is now fleet-live:
--   * commit ec7ade491 (reconcilePlanWithRealised reads site_has_no_current_plan
--     via noCurrentPlanFlag) shipped in agent-chassis v1.0.1151;
--   * verified in the running pod 2026-07-24: `strings /app/agent-chassis`
--     contains "site_has_no_current_plan" and the new "preserved pages exceed
--     max_pages" log, and NOT the old "adoption-locked pages exceed max_pages".
--
-- So the transition is over. This migration DROPS the adoption_locked alias, so
-- the query emits only site_has_no_current_plan. The misleading name is now gone
-- from the live grep surface (agent_definitions) entirely.
--
-- SAFETY. The only runtime reader of adoption_locked fleet-wide is
-- noCurrentPlanFlag's fallback, which reads site_has_no_current_plan FIRST — that
-- key is still emitted, so the running chassis is unaffected. Confirmed 2026-07-24:
-- no other Go consumer and no other agent_definition references adoption_locked.
-- The Go fallback is now dead (never reached) and is removed in a follow-up commit;
-- until that ships it is harmless.
--
-- The query below is the LIVE query as read from agent_definitions on 2026-07-24
-- (the 193 both-alias shape), with the adoption_locked column removed.

BEGIN;

SELECT snapshot_agent('build-site-planner',
                      '194_load_existing_pages_drop_adoption_locked_alias: pre-update');

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
       END AS site_has_no_current_plan
FROM pages p
WHERE p.site_id = $1 AND p.status = 'active'
ORDER BY p.name
$Q$::text),
        true
    ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- Guard: exactly one active build-site-planner row now emits the new name and
-- NO LONGER emits adoption_locked.
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
        NOT LIKE '%adoption_locked%';
  IF n <> 1 THEN
    RAISE EXCEPTION '194 guard: expected exactly 1 active build-site-planner emitting site_has_no_current_plan and NOT adoption_locked, found %', n;
  END IF;
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- Verification
-- ---------------------------------------------------------------------------
-- 1. The step query now carries ONLY the new name:
--    SELECT (q LIKE '%site_has_no_current_plan%') AS has_new,
--           (q LIKE '%adoption_locked%')          AS has_old
--    FROM (SELECT default_config->'workflow'->'steps'->'load_existing_pages'->'config'->>'query' AS q
--          FROM agent_definitions WHERE type='build-site-planner' AND is_active=true) x;
--    -- expect: has_new = t, has_old = f
--
-- 2. The running chassis (v1.0.1151+, reads site_has_no_current_plan) is
--    unaffected: a re-plan on a first-plan site still force-preserves its pages.
--
-- Rollback: restore the pre-update snapshot captured above (snapshot_agent), which
-- re-adds the adoption_locked alias. Only needed if a pre-ec7ade491 chassis is
-- rolled back into service.

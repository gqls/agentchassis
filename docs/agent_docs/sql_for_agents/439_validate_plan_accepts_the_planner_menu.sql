-- 439 — build-site-planner: validate_plan ACCEPTS what load_components OFFERED
-- (bugs_open/282, the config half; the Go half is component_name_resolver_menu.go).
--
-- WHY. A planner step is SHOWN a component menu (load_components -> the
-- `available_components` array the prompt renders) and validate_site_plan then
-- RESOLVES each proposed section name, DROPPING what it cannot resolve. Those
-- are two halves of one contract, and they were maintained in two places, in
-- two languages: the offer as this SQL string, the acceptance as a Go query
-- with `component_level IN ('section','element')` hardcoded
-- (loadComponentNameResolver). Migration 407 widened the OFFER (tool-level
-- components, per-site opt-in via plan_includes_tools) and nothing widened the
-- ACCEPTANCE.
--
-- MEASURED (live DB, run corr 2f74a975-1a87-40a8-af88-a9bd2ecc1510, plan
-- dcbae4df, loancalculator.co.uk, 2026-08-15 14:25):
--   * collected_data.available_components = 151 rows including 11 finance tool
--     functions — the planner WAS offered them;
--   * collected_data.llm_plan.result names a tool function on 12 pages
--     (index: ["hero","tool-loan-repayment",...]);
--   * collected_data.validate_plan — validate's OWN output — carries none of
--     them (index: 5 sections). The drop is at validate, not at the planner.
--
-- SHAPE. One new key on the validate_plan step: `menu_field`, naming the
-- collected-data path holding the menu the planner actually saw. The Go arm
-- adds those rows to the resolver's valid set (UNION with the section/element
-- base — it never removes, so nothing any site accepts today stops being
-- accepted). Sixth opt-in key on this agent's step configs, unsafe default OFF:
-- absent key = today's behaviour exactly.
--
-- WHY NOT MIRROR 407's PREDICATE INTO Go. It would be a THIRD hand-maintained
-- copy of a string that has already drifted: migration 419 added a
-- requires-backend gate to this same query and guards its apply by asserting
-- 407's exact text. Consuming the offer's OUTPUT single-sources it — any future
-- gate on the menu flows through with nothing to keep in step (016b §9: "single
-- sourcing is a guarantee, a lockstep test is a backstop").
--
-- ORDER-SAFE IN BOTH DIRECTIONS, which is required — config is live on apply,
-- Go rides the next image roll:
--   * this migration applied, Go not yet rolled -> the key is unread; today's
--     behaviour;
--   * Go rolled, this migration not applied -> no menu_field anywhere; today's
--     behaviour.
--
-- CONSUMERS OF validate_site_plan, ENUMERATED (not asserted):
--   SELECT type, k FROM agent_definitions, jsonb_object_keys(default_config#>'{workflow,steps}') AS k
--    WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
--      AND default_config#>'{workflow,steps}'->k->>'action'='validate_site_plan';
--   -> build-site-planner.validate_plan (this migration) and
--      site-planner.validate_plan (0 runs in orchestration_states; its own menu
--      is section/element-only, so opting it in would add nothing — DELIBERATELY
--      NOT TOUCHED).
--   content-gap-planner reaches the same Go resolver through apply_gap_plan
--   (3 call sites) but has NO validate_site_plan step and therefore no
--   menu_field: its menu stays section/element-only, per 407's ruling that
--   "gap-planning a NEW tool page is a different authority" (PLAN-049 landmine).
--
-- ROLLBACK: 439_validate_plan_accepts_the_planner_menu_ROLLBACK.sql, or restore
-- the snapshot this file takes (snapshot_agent note
-- '439_validate_plan_accepts_the_planner_menu: pre-update').

BEGIN;

SELECT snapshot_agent('build-site-planner',
  '439_validate_plan_accepts_the_planner_menu: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,validate_plan,config,menu_field}',
         to_jsonb('available_components'::text)
       ),
       updated_at = now()
 WHERE type = 'build-site-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  n int;
  menu_field text;
  menu_source text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '439: expected exactly 1 live build-site-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,validate_plan,config,menu_field}',
         default_config#>>'{workflow,steps,load_components,output_field}'
    INTO menu_field, menu_source
    FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  -- The two halves of the contract, checked TOGETHER. A menu_field naming a
  -- path no step writes is the one way this change can be inert while looking
  -- applied — exactly the failure mode 282 is, one level up.
  IF menu_field IS DISTINCT FROM 'available_components' THEN
    RAISE EXCEPTION '439: validate_plan.config.menu_field is %, expected available_components', menu_field;
  END IF;
  IF menu_source IS DISTINCT FROM menu_field THEN
    RAISE EXCEPTION '439: menu_field (%) does not name load_components.output_field (%) — the acceptance surface would read nothing', menu_field, menu_source;
  END IF;
END $$;

COMMIT;

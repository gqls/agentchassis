-- 328 — page-build-handler: wire plan-time fact assignments into plan_sections.
--
-- bugs_open/151 candidate 1, config half 1 of 3 (328 wiring, 329 planner prompt,
-- 330 writer prompt). Adds ONE key to the plan_sections step config:
--
--   "section_facts": "spec_sections.section_facts"
--
-- load_page_sections_from_spec emits section_facts (aligned with sections) ONLY
-- when its authoritative tier (site_plan_sections) served the list; plan_sections
-- consumes it ONLY when this config key names it — the feature is opt-in at the
-- step config, per the owner ruling of 2026-08-02 (new behaviour on a shared seam
-- ships as an opt-in field). Exactly ONE live agent wires spec_sections into
-- plan_sections (measured 2026-08-06), so this file touches one row, one step.
--
-- *** DO NOT APPLY until a chassis image containing the Go half is LIVE ***
-- Pod check (both strings in ONE exec; the control is a pre-existing literal
-- from a DIFFERENT file, invariant under this change):
--   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
--     'grep -ac "assigned fact id matches no current evidence_base fact" /app/agent-chassis; \
--      grep -ac "site_plan_sections lookup failed" /app/agent-chassis'
--   -> first count >= 1 (new), second >= 1 (control). `strings` is absent from
--   these images — use grep -ac (LANDMINES:503).
-- Applying early is NOT fatal (PlanSectionsInputSpec is CheckConfig without
-- StrictConfig, so an old binary warns on the unknown key and continues), but
-- the key would be dead config until the roll — image first, config second.

SELECT snapshot_agent('page-build-handler', '328_page_build_handler_wires_section_facts.sql: pre-update');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_328 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'page-build-handler' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_sections,config,section_facts}',
        '"spec_sections.section_facts"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Guard: exactly one row updated and the key reads back.
DO $$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type = 'page-build-handler' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
      AND default_config->'workflow'->'steps'->'plan_sections'->'config'->>'section_facts'
          = 'spec_sections.section_facts';
    IF n <> 1 THEN
        RAISE EXCEPTION '328: expected exactly 1 page-build-handler row carrying the section_facts wiring, found %', n;
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run):
--   UPDATE agent_definitions ad
--   SET default_config = b.default_config
--   FROM agent_definitions_bak_328 b
--   WHERE ad.id = b.id;
-- or surgically:
--   UPDATE agent_definitions
--   SET default_config = default_config #- '{workflow,steps,plan_sections,config,section_facts}'
--   WHERE type='page-build-handler' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

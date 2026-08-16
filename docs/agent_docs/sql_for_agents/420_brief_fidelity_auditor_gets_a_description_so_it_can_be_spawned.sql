-- 420_brief_fidelity_auditor_gets_a_description_so_it_can_be_spawned.sql
--
-- FOUND 2026-08-16 by the first attempt to spawn brief-fidelity-auditor after
-- migration 419 wired it into the improvement loop: the spawn FAILED before the
-- auditor ran —
--   failed to get agent definition: sql: Scan error on column index 3,
--   name "description": converting NULL to string is unsupported
--
-- The auditor's seed (brochure_component_library/agents/brief-fidelity-auditor
-- .seed.sql) never set `description`; the column is nullable with no default; and
-- every agent-definition loader in Go (spawn_actions.go getAgentDefinition,
-- ai_actions.go loadAgentDefinitionForAction, generate_image_actions.go
-- loadAgentDefinitionForImageAction) scans it into a plain string. So this agent
-- has NEVER been spawnable via spawn_agent — its 2026-08-13 findings must have come
-- from a path that did not go through that loader. [MEASURED] it is the ONLY live
-- definition with a NULL description (1 of 193; the other 3 NULLs fleet-wide are
-- inactive scratch probes). The improvement loop's error_step would have carried
-- on past the failure every sweep, so without this the wiring in 419 would have
-- reported a completed sweep with the auditor silently skipped — bugs_open/287.
--
-- THIS FILE IS THE DATA HALF (immediate). The code half — COALESCE(description,'')
-- in the three loaders, the idiom three other loaders in the same package already
-- use — ships with the same bug and needs a roll; it makes the class
-- unrepresentable rather than fixing one row.

SELECT snapshot_agent('brief-fidelity-auditor',
                      '420_brief_fidelity_auditor_gets_a_description_so_it_can_be_spawned.sql: pre-update');

BEGIN;

DO $$
DECLARE n integer; d text;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='brief-fidelity-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n <> 1 THEN RAISE EXCEPTION 'MIGRATION 420: expected 1 live brief-fidelity-auditor, found %', n; END IF;
    SELECT description INTO d FROM agent_definitions
    WHERE type='brief-fidelity-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF d IS NOT NULL THEN RAISE EXCEPTION 'MIGRATION 420: description already set (%) — already applied', left(d,40); END IF;
END $$;

UPDATE agent_definitions
SET description = 'Audits whether a built site is faithful to its own brief: grades the page/component inventory against mission_brief, design_intent and content_direction, and files broken promises as work items in the router''s category vocabulary (audit_source brief-fidelity-audit). Runs inside the improvement loop (mig 419).',
    version = version + 1, updated_at = now()
WHERE type='brief-fidelity-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE d text;
BEGIN
    SELECT description INTO d FROM agent_definitions
    WHERE type='brief-fidelity-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF d IS NULL OR length(d) = 0 THEN RAISE EXCEPTION 'MIGRATION 420: description still empty after update'; END IF;
    RAISE NOTICE 'migration 420 OK: brief-fidelity-auditor is now spawnable (description set)';
END $$;

COMMIT;

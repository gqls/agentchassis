-- 530 — arm the instance-scope guard on tool-deployer's FORK path
--       (council 6acf8e4e round 1: bug_historian + reuse_agent objected that
--       deploy_tool_to_site — the sibling door to create_tool_component —
--       stayed unguarded; measured live 2026-08-21: 13 fork-births in 30 days,
--       latest 08-19, and an UNPLACED library source is invisible to the
--       daily sweep, so this door is the only preventive control that sees it).
--
-- Census before writing (both queries, RUNBOOK §6): exactly ONE live executor,
-- tool-deployer step deploy_tool; the text census agrees (1 row).
-- Same safety argument as 520: inert until a binary carrying the guarded
-- DeployToolToSiteAction rolls; pre-roll binaries neither read nor refuse the
-- key (spec declared no config contract before this change; the new
-- declaration is reporting-only ConfigKeys shipped WITH the reader).
BEGIN;
DO $$
DECLARE armed text;
BEGIN
  SELECT default_config#>>'{workflow,steps,deploy_tool,config,enforce_instance_scope}'
    INTO armed FROM agent_definitions
   WHERE type='tool-deployer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF armed = 'true' THEN
    RAISE EXCEPTION '530: already applied — nothing to do';
  END IF;
END $$;

SELECT snapshot_agent('tool-deployer', '530 pre-image: fork-path instance-scope guard arming');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,deploy_tool,config,enforce_instance_scope}',
      'true'::jsonb, true),
    updated_at = now()
WHERE type='tool-deployer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config#>>'{workflow,steps,deploy_tool,action}' = 'deploy_tool_to_site';

DO $$
DECLARE armed text;
BEGIN
  SELECT default_config#>>'{workflow,steps,deploy_tool,config,enforce_instance_scope}'
    INTO armed FROM agent_definitions
   WHERE type='tool-deployer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF armed IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '530 verify: enforce_instance_scope not armed on tool-deployer.deploy_tool (got %)', COALESCE(armed,'NULL');
  END IF;
END $$;
COMMIT;

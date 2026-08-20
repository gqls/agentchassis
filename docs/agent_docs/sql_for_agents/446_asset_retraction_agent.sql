-- 446: asset-retraction agent — the dispatchable wrapper for retract_asset_files (DGH-010).
--
-- ⚠ NAMED _HOLD BECAUSE OF THE IMAGE-FIRST RULE, NOT BECAUSE IT IS UNFINISHED.
-- The action handler ships in commit 37303a6dd and is INERT until the next
-- chassis roll; a seed naming an unregistered action fails at runtime
-- (CLAUDE.md: image first, then seeds). AFTER the roll that carries
-- retract_asset_files: verify the action is in the running binary
-- (grep -aq "retract_asset_files requires a database handle" /proc/1/exe on a
-- chassis pod, with a negative control), then rename this file to drop _HOLD
-- and apply it via the migration runner, scoped to this directory.
--
-- Mirrors sql_for_agents/215 (page-retraction) — same shape, same rationale:
-- operator-initiated by construction, no cron, safety lives in the ACTION not
-- this config (dry_run defaults TRUE inside the action; a workflow cannot
-- switch the guards off — there is no config key for them).
--
-- DISPATCH (both required):
--   input_data: {"site_id": "<uuid>", "paths": ["/assets/images/<file>", ...]}
-- The deleting run needs "dry_run": false as a LITERAL in the step config —
-- which for this seed means a config override on dispatch, or run the dry
-- audit first (the default) and read the refusals before arming anything.
INSERT INTO agent_definitions (type, display_name, category, description, default_config, is_active)
VALUES (
  'asset-retraction',
  'Asset Retraction Agent',
  'specialist',
  'Asset-level UNPUBLISH: deletes caller-named orphan asset files via retract_asset_files (five-guard chain, dry-run by default). Refuses owned, brand-head, referenced and out-of-scope paths, and records every refusal.',
  jsonb_build_object(
    'workflow', jsonb_build_object(
      'start_step', 'retract',
      'processing_mode', 'orchestrator',
      'steps', jsonb_build_object(
        'retract', jsonb_build_object(
          'action', 'retract_asset_files',
          'description', 'Guard-checked asset file deletion; dry run unless the step config carries dry_run:false',
          'config', jsonb_build_object(
            'site_id_field', 'input_data.site_id',
            'paths_field',   'input_data.paths'
          ),
          'output_field', 'asset_retraction',
          'next_step', 'complete'
        ),
        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Asset retraction complete',
          'config', jsonb_build_object('output_fields', jsonb_build_array('asset_retraction'))
        )
      )
    )
  ),
  true
)
ON CONFLICT DO NOTHING;

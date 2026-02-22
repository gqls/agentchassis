-- ============================================================================
-- 023_asset_deploy_agent.sql
-- Asset Deploy Agent for the unified build/maintenance system
-- ============================================================================
-- Specialist agent that deploys generated images from storage to git repos.
-- Queries undeployed assets for a site, loops through each calling the
-- existing deploy_image_asset action.
--
-- Can be called by:
--   - site-work-orchestrator (handler for undeployed_asset work items)
--   - maintenance-batch-scheduler (standalone per-site scan)
--   - CLI trigger (manual)
--
-- Uses:
--   - ensure_site_record (existing) — resolve site_id/domain
--   - load_undeployed_assets (new) — thin wrapper around findUndeployedAssets
--   - deploy_image_asset (existing) — download from S3, optimize, commit to git
--
-- No new core logic — just orchestration of existing actions.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active,
    capabilities, domain_tags, agent_category, status,
    input_contract, output_contract
) VALUES (
             'asset-deploy-agent',
             'Asset Deploy Agent',
             'Deploys generated assets (logos, hero images) from object storage to site git repositories. Queries for assets that exist in the assets table but are not referenced in any deployed page HTML, then downloads, optimizes, and commits each to the site repo via the git adapter.',
             'specialist',
             jsonb_build_object(
                     'processing_mode', 'task',
                     'timeout_seconds', 600,
                     'workflow', jsonb_build_object(
                             'start_step', 'ensure_site_record',
                             'steps', jsonb_build_object(

                                     'ensure_site_record', jsonb_build_object(
                                     'action', 'ensure_site_record',
                                     'config', jsonb_build_object(
                                             'input_fields', jsonb_build_array('site_id', 'domain')
                                               ),
                                     'next_step', 'load_undeployed_assets',
                                     'description', 'Load site record from domain or site_id',
                                     'output_field', 'site_record'
                                                           ),

                                     'load_undeployed_assets', jsonb_build_object(
                                             'action', 'load_undeployed_assets',
                                             'config', jsonb_build_object(
                                                     'site_id', 'site_record.site_id'
                                                       ),
                                             'next_step', 'check_has_assets',
                                             'description', 'Query assets table for images not referenced in any deployed page',
                                             'output_field', 'undeployed'
                                                               ),

                                     'check_has_assets', jsonb_build_object(
                                             'action', 'conditional',
                                             'config', jsonb_build_object(
                                                     'condition', 'undeployed.has_assets == true',
                                                     'then_step', 'deploy_loop',
                                                     'else_step', 'complete_no_work'
                                                       ),
                                             'description', 'Any assets to deploy?'
                                                         ),

                                     'deploy_loop', jsonb_build_object(
                                             'action', 'loop',
                                             'config', jsonb_build_object(
                                                     'items_field', 'undeployed.assets',
                                                     'item_variable', 'current_asset',
                                                     'mode', 'sequential',
                                                     'max_iterations', 20,
                                                     'sub_workflow', jsonb_build_object(
                                                             'start_step', 'deploy_asset',
                                                             'steps', jsonb_build_object(
                                                                     'deploy_asset', jsonb_build_object(
                                                                     'action', 'deploy_image_asset',
                                                                     'config', jsonb_build_object(
                                                                             'purpose', 'current_asset.purpose',
                                                                             's3_uri', 'current_asset.s3_uri',
                                                                             'domain', 'site_record.domain'
                                                                               ),
                                                                     'next_step', 'asset_done',
                                                                     'description', 'Download from S3, optimize, commit to git',
                                                                     'output_field', 'deploy_result'
                                                                                     ),
                                                                     'asset_done', jsonb_build_object(
                                                                             'action', 'loop_complete',
                                                                             'description', 'Asset deployed'
                                                                                   )
                                                                      )
                                                                     )
                                                       ),
                                             'next_step', 'complete',
                                             'description', 'Deploy each undeployed asset',
                                             'output_field', 'deploy_results'
                                                    ),

                                     'complete', jsonb_build_object(
                                             'action', 'complete_workflow',
                                             'config', jsonb_build_object(
                                                     'output_fields', jsonb_build_array('undeployed', 'deploy_results')
                                                       ),
                                             'description', 'All assets deployed'
                                                 ),

                                     'complete_no_work', jsonb_build_object(
                                             'action', 'complete_workflow',
                                             'config', jsonb_build_object(
                                                     'output_fields', jsonb_build_array('undeployed'),
                                                     'success_message', 'No undeployed assets found'
                                                       ),
                                             'description', 'Nothing to deploy'
                                                         )
                                      )
                                 )
             ),
             true,
             '["asset_management", "image_deploy", "git_operations"]'::jsonb,
             '["maintenance", "assets", "deployment"]'::jsonb,
             'executor',
             'active',
             jsonb_build_object(
                     'required', jsonb_build_array(),
                     'optional', jsonb_build_array('site_id', 'domain'),
                     'description', 'Pass site_id or domain. Agent queries for undeployed assets itself.'
             ),
             jsonb_build_object(
                     'produces', jsonb_build_object(
                     'undeployed', 'count and list of assets found',
                     'deploy_results', 'per-asset deploy outcomes'
                                 )
             )
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              default_config = EXCLUDED.default_config,
                              capabilities = EXCLUDED.capabilities,
                              domain_tags = EXCLUDED.domain_tags,
                              agent_category = EXCLUDED.agent_category,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              status = EXCLUDED.status,
                              updated_at = NOW();


-- ============================================================================
-- Registry entry (add to registry.go):
--
--   "load_undeployed_assets": {
--       Handler:     LoadUndeployedAssetsAction,
--       Category:    "site",
--       Description: "Query undeployed assets for a site (in assets table but not in page HTML)",
--       IsLocal:     true,
--   },
--
-- ============================================================================
-- Data flow trace:
--
-- 1. ensure_site_record
--    Input: input_data.site_id OR input_data.domain
--    Output: site_record → {site_id, domain, name, ...}
--
-- 2. load_undeployed_assets
--    Config: site_id = "site_record.site_id" (path, resolved by ExtractActionInputs)
--    Output: undeployed → {assets: [{asset_id, purpose, asset_type, s3_uri}], count, has_assets}
--
-- 3. deploy_loop / deploy_asset (per asset)
--    Config:
--      purpose = "current_asset.purpose"  (path → "logo" or "hero")
--      s3_uri  = "current_asset.s3_uri"   (path → "s3://bucket/images/...")
--      domain  = "site_record.domain"     (path → "finetuning.uk")
--    All three resolve via ExtractActionInputs in deploy_image_asset.
--    deploy_image_asset handles: download → optimize → base64 → git adapter request
--    Output: deploy_result → {deployed: true, image_url, output_path, size_bytes}
-- ============================================================================
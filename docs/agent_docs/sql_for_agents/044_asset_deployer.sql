INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    status,
    agent_category,
    domain_tags,
    input_contract,
    output_contract
) VALUES (
             'asset-deployer',
             'Asset Deployer',
             'Deploys a single image asset: downloads from S3, optimizes by purpose, commits to git. Reusable for any image deploy task.',
             'specialist',
             '{
                 "workflow": {
                     "start_step": "deploy_asset",
                     "steps": {
                         "deploy_asset": {
                             "action": "deploy_image_asset",
                             "config": {
                                 "input_fields": ["s3_uri", "purpose", "domain", "asset_key"]
                             },
                             "next_step": "complete",
                             "description": "Download from S3, optimize by purpose, commit to git",
                             "output_field": "deploy_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["deploy_result"]
                             },
                             "description": "Asset deploy complete"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 120
             }'::jsonb,
             true,
             '["image-deploy", "asset-management"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.791',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "128Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             'experimental',
             'specialist',
             '["image", "deploy", "asset", "git"]'::jsonb,
             '{"required": ["domain", "s3_uri"], "optional": ["purpose", "asset_key"], "description": "Provide domain + s3_uri. Optional purpose (default: hero) controls resize dimensions; optional asset_key fixes the filename when it differs from purpose. The deploy PATH is derived from (asset_key, purpose) and cannot be chosen — an explicit deploy_path is refused (bugs_open/179 finding A)."}'::jsonb,
             '{"produces": {"deploy_result": "Git commit result with image_url, output_path, size_bytes"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       image_tag = EXCLUDED.image_tag,
                                       description = EXCLUDED.description,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       status = EXCLUDED.status,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       updated_at = NOW();
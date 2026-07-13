-- ============================================================
-- component-quality-auditor agent + backfill
-- ============================================================
-- Agent: scans components periodically, scores them, and creates
--        needs_component_regeneration items for low-quality ones.
-- Backfill: one-shot work item that scores every existing component.

BEGIN;

-- ============================================================
-- 1. component-quality-auditor agent definition
-- ============================================================
BEGIN;

-- Delete any existing row for this type so the insert always wins
DELETE FROM agent_definitions WHERE type = 'component-quality-auditor';

-- Then the plain INSERT (same body as before, just without ON CONFLICT clause)
INSERT INTO agent_definitions (
    type, display_name, description,
    category, agent_category, status,
    default_config,
    input_contract, output_contract,
    image_repository, image_tag,
    resources, topics, health_config
)
VALUES (
           'component-quality-auditor',
           'Component Quality Auditor',
           'Scores existing content_components and creates regeneration work items for low-quality ones. Runs periodically to keep the component library healthy.',
           'maintenance',
           'analyst',
           'active',
           jsonb_build_object(
                   'workflow', jsonb_build_object(
                   'start_step', 'score_components',
                   'processing_mode', 'task',
                   'timeout_seconds', 600,
                   'steps', jsonb_build_object(
                           'score_components', jsonb_build_object(
                                   'action', 'compute_component_quality',
                                   'config', jsonb_build_object(
                                           'scan_all', true,
                                           'stale_days', 7
                                             ),
                                   'next_step', 'create_regen_items',
                                   'output_field', 'quality_scan',
                                   'description', 'Score components not checked in 7+ days'
                                               ),
                           'create_regen_items', jsonb_build_object(
                                   'action', 'loop',
                                   'config', jsonb_build_object(
                                           'items_field', 'quality_scan.results',
                                           'item_variable', 'current_component',
                                           'max_iterations', 100,
                                           'continue_on_error', true,
                                           'sub_workflow', jsonb_build_object(
                                                   'start_step', 'check_quality',
                                                   'steps', jsonb_build_object(
                                                           'check_quality', jsonb_build_object(
                                                                   'action', 'conditional',
                                                                   'config', jsonb_build_object(
                                                                           'condition', 'current_component.quality_score < 50',
                                                                           'then_step', 'create_work_item',
                                                                           'else_step', 'done'
                                                                             )
                                                                            ),
                                                           'create_work_item', jsonb_build_object(
                                                                   'action', 'create_work_item',
                                                                   'config', jsonb_build_object(
                                                                           'item_type',     'needs_component_regeneration',
                                                                           'item_domain',   'build',
                                                                           'handler_agent', 'component-creator',
                                                                           'source',        'component-quality-auditor',
                                                                           'severity',      'medium',
                                                                           'priority',      50,
                                                                           'summary',       'current_component.function',
                                                                           'spec_data', jsonb_build_object(
                                                                                   'component_id',   'current_component.component_id',
                                                                                   'function',       'current_component.function',
                                                                                   'quality_score',  'current_component.quality_score',
                                                                                   'quality_issues', 'current_component.quality_issues'
                                                                                        ),
                                                                           'item_key_prefix', 'quality_regen'
                                                                             ),
                                                                   'next_step', 'done'
                                                                               ),
                                                           'done', jsonb_build_object('action', 'loop_complete')
                                                            )
                                                           )
                                             ),
                                   'next_step', 'complete',
                                   'output_field', 'items_created',
                                   'description', 'Create regeneration items for components scoring below 50'
                                                 ),
                           'complete', jsonb_build_object(
                                   'action', 'complete_workflow',
                                   'config', jsonb_build_object(
                                           'output_fields', jsonb_build_array('quality_scan', 'items_created')
                                             )
                                       )
                            )
                               )
           ),
           jsonb_build_object('description', 'Scheduled agent - no required input'),
           jsonb_build_object('produces', jsonb_build_object(
                   'quality_scan',  'Component quality scores',
                   'items_created', 'Regeneration work items for low-quality components'
                                          )),
           'docker.io/aqls/agent-chassis',
           'v1.0.954',
           jsonb_build_object(
                   'requests', jsonb_build_object('cpu', '100m', 'memory', '128Mi'),
                   'limits',   jsonb_build_object('cpu', '500m', 'memory', '512Mi')
           ),
           jsonb_build_object(
                   'process',  'system.agent.{type}.process',
                   'response', 'system.responses.{type}',
                   'error',    'system.errors.{type}'
           ),
           jsonb_build_object(
                   'port', 8080,
                   'liveness_path',  '/health',
                   'readiness_path', '/ready',
                   'initial_delay_seconds', 30
           )
       );

COMMIT;

-- Verify
SELECT type, agent_category, status,
       default_config->'workflow'->>'start_step' as start_step
FROM agent_definitions
WHERE type = 'component-quality-auditor';

-- ============================================================
-- 2. Backfill: score all existing active components once
-- ============================================================
-- Creates a one-off work item. When the auditor picks it up, it will
-- score everything (no stale filter for the backfill).
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
)
VALUES (
           NULL,
           'backfill', 'maintenance', 'component_quality_scan', 'low',
           'Score all existing components (backfill)',
           jsonb_build_object('scan_all', true)::jsonb,
           10, 'component-quality-auditor', 'triaged', 'manual',
           'backfill_component_quality_' || extract(epoch from now())::int
       )
    ON CONFLICT (item_key) DO NOTHING;

COMMIT;

-- Verify
SELECT type, status FROM agent_definitions WHERE type = 'component-quality-auditor';

SELECT item_key, item_type, status, summary
FROM site_work_items
WHERE item_key LIKE 'backfill_component_quality_%'
ORDER BY created_at DESC
    LIMIT 1;

-- ============================================================================
-- spec-updater agent
--
-- Handler for needs_spec_update work items. Receives a work item with
-- {aspect, field, suggested_value} in its spec and applies the update
-- to site_specs using the same versioning pattern as WriteSiteSpecAction.
--
-- No LLM needed — this is a mechanical merge operation.
-- Items without field/value (description-only) complete with
-- "needs human review" and the item stays for manual triage.
--
-- Workflow is deliberately simple: load site record → apply update → complete.
-- The complexity is in the Go action, not in the workflow.
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'spec-updater',
             'Spec Updater',
             'Applies spec updates from audit findings. Reads aspect/field/value from work item spec, merges into site_specs with versioning. No LLM — mechanical merge only.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.842', 'specialist',
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "processing_mode": "orchestrator",
                     "timeout_seconds": 60,
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "apply_update",
                             "output_field": "site_record"
                         },
                         "apply_update": {
                             "action": "update_site_spec_from_item",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "spec": "input_data.spec",
                                 "work_item_id": "input_data.work_item_id"
                             },
                             "next_step": "complete",
                             "error_step": "complete",
                             "description": "Read work item spec, merge field into site_specs",
                             "output_field": "update_result"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": { "output_fields": ["update_result"] }
                         }
                     }
                 }
             }'::jsonb,
             '{"required": ["site_id", "domain"], "optional": ["work_item_id", "spec"]}'::jsonb,
             '{"produces": {"update_result": "spec update result — updated/skipped/needs_review"}}'::jsonb,
             '["spec", "metadata", "config"]'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();

-- Verify
SELECT type, display_name, status
FROM agent_definitions
WHERE type = 'spec-updater' AND deleted_at IS NULL;
-- 059_improvement_loop.sql
--
-- improvement-loop agent definition
--
-- Runs after initial build completes. Spawns discovery agents to find issues,
-- triages findings into the dispatch queue, then dispatches fixes.
--
-- Flow:
--   1. ensure_site_record
--   2. spawn + call quality-discovery-agent     (broken_nav_links, placeholder_contact, generic_theme)
--   3. spawn + call design-discovery-agent       (undeployed_assets, missing_css, duplicate_palette)
--   4. spawn + call completeness-discovery-agent (empty_sections)
--   5. triage_detected_items                     (promote detected → triaged, domain → build)
--   6. check if any items were promoted
--   7. if yes: insert needs_rerender (priority 99, runs last)
--              → spawn + call build-dispatch-loop (processes all fixes, then rerender)
--   8. complete
--
-- Triggered by:
--   a) Side-effect after last build item completes (future: add to dispatch loop)
--   b) Manual trigger with site_id + domain
--   c) Scheduled maintenance job
--
-- Input: site_id, domain

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics,
    health_config, env_vars, version,
    delegation_preferences, agent_category, status,
    domain_tags, briefing_questionnaire,
    usage_count, is_snapshot, input_contract, output_contract
) VALUES (
             'improvement-loop',
             'Improvement Loop',
             'Post-build quality improvement cycle. Runs discovery agents to find issues, triages findings, dispatches fixes via build-dispatch-loop, and triggers rerender when fixes complete.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "steps": {

                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "spawn_quality_discovery",
                             "description": "Load site record from database",
                             "output_field": "site_record"
                         },

                         "spawn_quality_discovery": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "quality_checker",
                                 "agent_type": "quality-discovery-agent"
                             },
                             "next_step": "call_quality_discovery",
                             "description": "Spawn quality discovery agent",
                             "output_field": "quality_checker_spawned"
                         },

                         "call_quality_discovery": {
                             "action": "call_agent",
                             "config": {
                                 "target_role": "quality_checker",
                                 "agent_type": "quality-discovery-agent",
                                 "input_mapping": {
                                     "site_id": "site_record.site_id",
                                     "domain": "site_record.domain"
                                 },
                                 "timeout_seconds": 120
                             },
                             "next_step": "spawn_design_discovery",
                             "error_step": "spawn_design_discovery",
                             "description": "Run quality checks (broken nav, placeholder contact, generic theme)",
                             "output_field": "quality_result"
                         },

                         "spawn_design_discovery": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "design_checker",
                                 "agent_type": "design-discovery-agent"
                             },
                             "next_step": "call_design_discovery",
                             "description": "Spawn design discovery agent",
                             "output_field": "design_checker_spawned"
                         },

                         "call_design_discovery": {
                             "action": "call_agent",
                             "config": {
                                 "target_role": "design_checker",
                                 "agent_type": "design-discovery-agent",
                                 "input_mapping": {
                                     "site_id": "site_record.site_id",
                                     "domain": "site_record.domain"
                                 },
                                 "timeout_seconds": 120
                             },
                             "next_step": "spawn_completeness_discovery",
                             "error_step": "spawn_completeness_discovery",
                             "description": "Run design checks (undeployed assets, missing CSS, duplicate palette)",
                             "output_field": "design_result"
                         },

                         "spawn_completeness_discovery": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "completeness_checker",
                                 "agent_type": "completeness-discovery-agent"
                             },
                             "next_step": "call_completeness_discovery",
                             "description": "Spawn completeness discovery agent",
                             "output_field": "completeness_checker_spawned"
                         },

                         "call_completeness_discovery": {
                             "action": "call_agent",
                             "config": {
                                 "target_role": "completeness_checker",
                                 "agent_type": "completeness-discovery-agent",
                                 "input_mapping": {
                                     "site_id": "site_record.site_id",
                                     "domain": "site_record.domain"
                                 },
                                 "timeout_seconds": 120
                             },
                             "next_step": "triage_findings",
                             "error_step": "triage_findings",
                             "description": "Run completeness checks (empty sections)",
                             "output_field": "completeness_result"
                         },

                         "triage_findings": {
                             "action": "triage_detected_items",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "target_domain": "build"
                             },
                             "next_step": "check_has_findings",
                             "description": "Promote detected items to triaged with domain=build for dispatch",
                             "output_field": "triage_result"
                         },

                         "check_has_findings": {
                             "action": "conditional",
                             "config": {
                                 "condition": "triage_result.has_items == true",
                                 "then_step": "insert_rerender_item",
                                 "else_step": "complete_clean"
                             },
                             "description": "Check if any issues were found and promoted"
                         },

                         "insert_rerender_item": {
                             "action": "create_work_item",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "source": "improvement-loop",
                                 "item_domain": "build",
                                 "item_type": "needs_rerender",
                                 "severity": "medium",
                                 "summary": "Re-assemble and deploy pages after improvement fixes",
                                 "spec": {"refresh_site_components": true},
                                 "priority": 99,
                                 "handler_agent": "rerender-pages",
                                 "item_key_prefix": "improvement_rerender"
                             },
                             "next_step": "spawn_dispatch",
                             "description": "Insert rerender work item (priority 99 = runs after all fixes)",
                             "output_field": "rerender_item_created"
                         },

                         "spawn_dispatch": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "dispatcher",
                                 "agent_type": "build-dispatch-loop"
                             },
                             "next_step": "call_dispatch",
                             "description": "Spawn dispatch loop to process fix items",
                             "output_field": "dispatch_spawned"
                         },

                         "call_dispatch": {
                             "action": "call_agent",
                             "config": {
                                 "target_role": "dispatcher",
                                 "agent_type": "build-dispatch-loop",
                                 "input_mapping": {
                                     "site_id": "site_record.site_id",
                                     "domain": "site_record.domain"
                                 },
                                 "timeout_seconds": 900
                             },
                             "next_step": "complete",
                             "error_step": "complete",
                             "description": "Dispatch fixes then rerender. On error, items remain in queue for heartbeat.",
                             "output_field": "dispatch_result"
                         },

                         "complete_clean": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["quality_result", "design_result", "completeness_result", "triage_result"],
                                 "success_message": "No issues found — site is clean"
                             },
                             "description": "No findings — site passed all checks"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["quality_result", "design_result", "completeness_result", "triage_result", "dispatch_result"]
                             },
                             "description": "Improvement loop complete — fixes dispatched and deployed"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 1800
             }'::jsonb,
             true,
             '["improvement", "discovery", "dispatch", "quality"]'::jsonb,
             'docker.io/aqls/agent-chassis', 'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb, 1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator', 'experimental',
             '["improvement", "quality", "maintenance"]'::jsonb, '{}'::jsonb,
             0, false,
             '{"required": ["site_id"], "optional": ["domain"], "description": "Receives site_id (and optionally domain) from post-build trigger or manual invocation."}'::jsonb,
             '{"produces": {"quality_result": "quality discovery findings", "design_result": "design discovery findings", "completeness_result": "completeness findings", "triage_result": "items promoted count", "dispatch_result": "dispatch loop result"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();


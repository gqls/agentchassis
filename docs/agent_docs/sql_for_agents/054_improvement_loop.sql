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

---

-- Add design-audit-agent and site-review-agent to the improvement loop
-- Insert after completeness discovery, before triage_findings
-- (design-audit-agent does its own triage internally, but the
-- improvement loop's triage catches anything the audit agents
-- created with status 'detected' rather than auto-triaging)

-- 1. Change completeness discovery next_step to spawn audit
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_completeness_discovery,next_step}',
        '"spawn_design_audit"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

-- 2. Add spawn + call design-audit-agent
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_design_audit}',
        '{
            "action": "spawn_agent",
            "config": { "role": "design_auditor", "agent_type": "design-audit-agent" },
            "next_step": "call_design_audit",
            "description": "Spawn LLM-based design audit agent",
            "output_field": "design_auditor_spawned"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_design_audit}',
        '{
            "action": "call_agent",
            "config": {
                "target_role": "design_auditor",
                "input_mapping": {
                    "site_id": "site_record.site_id",
                    "domain": "site_record.domain"
                },
                "timeout_seconds": 600
            },
            "next_step": "spawn_site_review",
            "error_step": "spawn_site_review",
            "description": "Run visual + content quality audit (LLM-based)",
            "output_field": "design_audit_result"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

-- 3. Add spawn + call site-review-agent
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_site_review}',
        '{
            "action": "spawn_agent",
            "config": { "role": "site_reviewer", "agent_type": "site-review-agent" },
            "next_step": "call_site_review",
            "description": "Spawn strategic site review agent",
            "output_field": "site_reviewer_spawned"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_site_review}',
        '{
            "action": "call_agent",
            "config": {
                "target_role": "site_reviewer",
                "input_mapping": {
                    "site_id": "site_record.site_id",
                    "domain": "site_record.domain"
                },
                "timeout_seconds": 600
            },
            "next_step": "triage_findings",
            "error_step": "triage_findings",
            "description": "Run strategic alignment review (LLM-based)",
            "output_field": "site_review_result"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

-- 4. Update complete step to include new outputs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,complete,config,output_fields}',
        '["quality_result", "design_result", "completeness_result", "design_audit_result", "site_review_result", "triage_result", "dispatch_result"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

-- Verify the chain
SELECT
    default_config->'workflow'->'steps'->'call_completeness_discovery'->>'next_step' as after_completeness,
    default_config->'workflow'->'steps'->'call_design_audit'->>'next_step' as after_audit,
    default_config->'workflow'->'steps'->'call_site_review'->>'next_step' as after_review
FROM agent_definitions
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

-- Expected: after_completeness=spawn_design_audit, after_audit=spawn_site_review, after_review=triage_findings
```

The full improvement loop flow is now:
```
ensure_site_record
→ quality-discovery-agent      (structural: broken nav, placeholder contact)
→ design-discovery-agent       (structural: undeployed assets, missing CSS, hardcoded colours)
→ completeness-discovery-agent (structural: empty sections, empty blog*)
→ design-audit-agent           (LLM: visual design + content quality)
→ site-review-agent            (LLM: strategic alignment)
→ triage_detected_items
→ insert needs_rerender
→ build-dispatch-loop          (processes all fixes)


-- ============================================================================
-- Fix dispatch loop pod accumulation
--
-- 1. Reduce timeout_seconds from 900 to 300 (safety net)
-- 2. Add last_completed_at update to build-dispatch-loop workflow
--    so the scheduler knows the previous execution finished
-- ============================================================================

-- 1. Shorter timeout
UPDATE scheduled_tasks
SET timeout_seconds = 300
WHERE name IN ('build-pipeline-trigger', 'improvement-sweep');

-- 2. Add scheduler callback step to build-dispatch-loop workflow
-- Insert "notify_scheduler" step between the last step and "complete"
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,notify_scheduler}',
                '{
                    "action": "query_database",
                    "config": {
                        "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''build-pipeline-trigger''",
                        "output_format": "object"
                    },
                    "next_step": "complete",
                    "description": "Tell scheduler this execution finished so it can fire again",
                    "output_field": "scheduler_notified"
                }'::jsonb
        ),
    -- Update the step that currently points to "complete" to point to "notify_scheduler"
        '{workflow,steps,check_has_items,config,else_step}',
        '"notify_scheduler"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-dispatch-loop' AND deleted_at IS NULL;

-- Also update the process_items next_step to go through notify_scheduler
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_items,next_step}',
        '"notify_scheduler"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-dispatch-loop' AND deleted_at IS NULL;

-- Verify the workflow flow: load_items → check_has_items → process_items → notify_scheduler → complete
--                                                    ↘ (no items) → notify_scheduler → complete
SELECT
    default_config->'workflow'->'steps'->'check_has_items'->'config'->>'else_step' as no_items_path,
    default_config->'workflow'->'steps'->'process_items'->>'next_step' as after_process,
    default_config->'workflow'->'steps'->'notify_scheduler'->>'next_step' as after_notify
FROM agent_definitions
WHERE type = 'build-dispatch-loop' AND deleted_at IS NULL;

-- Expected:
-- no_items_path: notify_scheduler
-- after_process: notify_scheduler
-- after_notify:  complete

-- 3. Same for improvement-loop — add scheduler callback
-- The improvement-loop has two completion paths: complete_clean and complete
-- Both need to go through notify_scheduler

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,notify_scheduler}',
        '{
            "action": "query_database",
            "config": {
                "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''improvement-sweep''",
                "output_format": "object"
            },
            "next_step": "complete",
            "description": "Tell scheduler this execution finished",
            "output_field": "scheduler_notified"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

-- Redirect complete_clean → notify_scheduler_clean → complete_clean
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,notify_scheduler_clean}',
        '{
            "action": "query_database",
            "config": {
                "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''improvement-sweep''",
                "output_format": "object"
            },
            "next_step": "complete_clean",
            "description": "Tell scheduler this execution finished (clean path)",
            "output_field": "scheduler_notified"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

-- Update check_has_findings else_step: complete_clean → notify_scheduler_clean
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_has_findings,config,else_step}',
        '"notify_scheduler_clean"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

-- Update call_dispatch next_step: complete → notify_scheduler
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_dispatch,next_step}',
        '"notify_scheduler"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

-- Also update call_dispatch error_step to still notify
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_dispatch,error_step}',
        '"notify_scheduler"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND deleted_at IS NULL;

--

-- ============================================================================
-- Migration 087: Improvement Loop Audit Pass Guard
-- ============================================================================
-- Inserts two steps into the improvement-loop workflow:
--
-- 1. check_audit_pass_limit — after ensure_site_record, before discovery.
--    Queries get_audit_pass_count(). If >= 3, skips to notify_scheduler_clean.
--
-- 2. increment_audit_pass — after triage, before check_has_findings.
--    Increments the pass count so the next run knows.
--
-- Current flow:
--   ensure_site_record → spawn_quality_discovery → ... → triage_findings → check_has_findings → ...
--
-- New flow:
--   ensure_site_record → check_audit_pass_limit → spawn_quality_discovery → ... → triage_findings → increment_audit_pass → check_has_findings → ...
--
-- After 3 passes: ensure_site_record → check_audit_pass_limit → notify_scheduler_clean → complete_clean
-- ============================================================================


-- Step 1: Change ensure_site_record to point to the new guard step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,ensure_site_record,next_step}',
        '"check_audit_pass_limit"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);


-- Step 2: Add the guard step (query pass count + conditional)
-- Uses query_database to read the count, then conditional to branch.
-- Two steps: load_pass_count and check_audit_pass_limit.

-- 2a: Add load_pass_count step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_pass_count}',
        '{
            "action": "query_database",
            "config": {
                "query": "SELECT get_audit_pass_count($1) as pass_count, CASE WHEN get_audit_pass_count($1) >= 3 THEN true ELSE false END as limit_reached",
                "params": ["site_record.site_id"],
                "output_format": "object"
            },
            "next_step": "check_audit_pass_limit",
            "description": "Load audit pass count for this site",
            "output_field": "pass_count_data"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- 2b: Fix ensure_site_record to point to load_pass_count (not check_audit_pass_limit)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,ensure_site_record,next_step}',
        '"load_pass_count"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- 2c: Add check_audit_pass_limit conditional step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_audit_pass_limit}',
        '{
            "action": "conditional",
            "config": {
                "condition": "pass_count_data.limit_reached == true",
                "then_step": "notify_scheduler_clean",
                "else_step": "spawn_quality_discovery"
            },
            "description": "Skip audits if site has reached 3 audit passes"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);


-- Step 3: Add increment_audit_pass step between triage_findings and check_has_findings
-- 3a: Insert the new step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,increment_audit_pass}',
        '{
            "action": "query_database",
            "config": {
                "query": "SELECT increment_audit_pass($1) as new_pass_count",
                "params": ["site_record.site_id"],
                "output_format": "object"
            },
            "next_step": "check_has_findings",
            "description": "Increment audit pass count for this site",
            "output_field": "pass_incremented"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- 3b: Point triage_findings to increment_audit_pass instead of check_has_findings
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,triage_findings,next_step}',
        '"increment_audit_pass"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);


-- ============================================================================
-- Verification
-- ============================================================================

-- Check the new flow: ensure_site_record → load_pass_count → check_audit_pass_limit → spawn_quality_discovery
SELECT
    default_config->'workflow'->'steps'->'ensure_site_record'->>'next_step' as "1_ensure→",
    default_config->'workflow'->'steps'->'load_pass_count'->>'next_step' as "2_load→",
    default_config->'workflow'->'steps'->'check_audit_pass_limit'->'config'->>'else_step' as "3_check_else→",
    default_config->'workflow'->'steps'->'check_audit_pass_limit'->'config'->>'then_step' as "3_check_then→",
    default_config->'workflow'->'steps'->'triage_findings'->>'next_step' as "triage→",
    default_config->'workflow'->'steps'->'increment_audit_pass'->>'next_step' as "increment→"
FROM agent_definitions
WHERE type = 'improvement-loop' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- Expected:
-- 1_ensure→      | 2_load→                  | 3_check_else→            | 3_check_then→          | triage→               | increment→
-- load_pass_count | check_audit_pass_limit   | spawn_quality_discovery  | notify_scheduler_clean | increment_audit_pass  | check_has_findings

To reset a site (e.g. after a major redesign): SELECT reset_audit_passes('site-uuid-here').

---

-- ============================================================================
-- 031_improvement_loop_news_enrichment.sql
--
-- Adds evaluate_news_feed as an enrichment step in the improvement-loop.
-- Runs after ensure_site_record, before load_pass_count.
--
-- Why: evaluate_news_feed only ran during initial classification. When we
-- add new fields (like separate_page) or new verticals, existing sites
-- never get updated. Running it every improvement cycle keeps the news
-- recommendation current for all sites. It's deterministic (no LLM),
-- idempotent (deep merge of identical values is a no-op), and fast
-- (one DB read + conditional write).
--
-- Chain change:
--   BEFORE: ensure_site_record → load_pass_count
--   AFTER:  ensure_site_record → enrich_news_feed → load_pass_count
--
-- The evaluate_news_feed action is already registered as a local action
-- (IsLocal: true) — no agent spawn needed.
-- ============================================================================

-- Step 1: Add the new step to the workflow
                                                          UPDATE agent_definitions
                                                   SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,enrich_news_feed}',
    '{
        "action": "evaluate_news_feed",
        "config": {
            "site_id": "site_record.site_id",
            "error_step": "load_pass_count"
        },
        "next_step": "load_pass_count",
        "description": "Refresh news feed recommendation from classification spec (deterministic, no LLM)",
        "output_field": "news_feed_enrichment"
    }'::jsonb
),
updated_at = NOW()
                                               WHERE type = 'improvement-loop'
                                                 AND deleted_at IS NULL;

-- Step 2: Repoint ensure_site_record to the new step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,ensure_site_record,next_step}',
        '"enrich_news_feed"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop'
  AND deleted_at IS NULL;

-- Step 3: Add news_feed_enrichment to the complete step's output_fields
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,complete,config,output_fields}',
        (
            SELECT jsonb_agg(DISTINCT val)
            FROM (
                     SELECT val FROM jsonb_array_elements_text(
                                             default_config->'workflow'->'steps'->'complete'->'config'->'output_fields'
                                     ) AS val
                     UNION
                     SELECT 'news_feed_enrichment'
                 ) sub
        )
                     ),
    updated_at = NOW()
WHERE type = 'improvement-loop'
  AND deleted_at IS NULL;

-- ============================================================================
-- Verify
-- ============================================================================
-- Check the chain is correct:
-- SELECT
--     step_name,
--     step_def->>'action' as action,
--     step_def->>'next_step' as next_step,
--     step_def->>'error_step' as error_step
-- FROM agent_definitions,
--      jsonb_each(default_config->'workflow'->'steps') AS s(step_name, step_def)
-- WHERE type = 'improvement-loop' AND deleted_at IS NULL
-- ORDER BY step_name;
--
-- Specifically verify the chain:
-- ensure_site_record → enrich_news_feed → load_pass_count
--
-- SELECT
--     default_config->'workflow'->'steps'->'ensure_site_record'->>'next_step' as ensure_next,
--     default_config->'workflow'->'steps'->'enrich_news_feed'->>'next_step' as enrich_next,
--     default_config->'workflow'->'steps'->'enrich_news_feed'->>'action' as enrich_action
-- FROM agent_definitions
-- WHERE type = 'improvement-loop' AND deleted_at IS NULL;
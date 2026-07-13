-- ============================================================================
-- Add Input Contracts to Existing Agents
-- ============================================================================
-- These contracts define what each agent expects to receive.
-- Contract validation will fail fast with clear error messages when required
-- fields are missing.
-- ============================================================================

-- backup before adding rerender step
175ec7ca-3e4d-43d3-b90e-0324b7614321 | intake-orchestrator | Intake Orchestrator | Entry point for site creation: classifies project type, runs briefing, spawns appropriate builder agent | orchestrator | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Intake complete - builder has been spawned"}, "call_briefer": {"action": "call_agent", "config": {"agent_type": "briefing-agent", "target_role": "briefer", "input_mapping": {"input_data": "input_data", "questionnaire": "questionnaire", "classification": "classification", "confirmed_type": "confirmed_type"}, "timeout_seconds": 120}, "next_step": "hitl_review_brief", "description": "Run the briefing questionnaire", "output_field": "brief_data"}, "call_builder": {"action": "call_agent", "config": {"max_retries": 10, "target_role": "builder", "input_mapping": {"brief_data": "brief_data", "input_data": "input_data", "questionnaire": "questionnaire", "classification": "classification", "confirmed_type": "confirmed_type", "reviewed_brief": "reviewed_brief"}, "timeout_seconds": 7200}, "next_step": "complete", "description": "Call the spawned builder to execute the site build", "output_field": "build_result"}, "spawn_briefer": {"action": "spawn_agent", "config": {"role": "briefer", "agent_type": "briefing-agent"}, "next_step": "call_classifier", "description": "Spawn briefing agent", "output_field": "briefer_agent"}, "spawn_builder": {"action": "spawn_agent", "config": {"role": "builder", "input_fields": ["input_data", "classification", "brief_data", "reviewed_brief"], "agent_type_field": "confirmed_type.recommended_builder"}, "next_step": "call_builder", "description": "Spawn the appropriate builder agent with all collected data", "output_field": "spawned_builder"}, "call_classifier": {"action": "call_agent", "config": {"agent_type": "site-classifier", "target_role": "classifier", "input_mapping": {"input_data": "input_data", "available_builders": "available_builders"}, "timeout_seconds": 30}, "next_step": "hitl_confirm_type", "description": "Classify the site type from domain and objective", "output_field": "classification"}, "spawn_classifier": {"action": "spawn_agent", "config": {"role": "classifier", "agent_type": "site-classifier"}, "next_step": "spawn_briefer", "description": "Spawn site classifier agent", "output_field": "classifier_agent"}, "hitl_confirm_type": {"action": "request_human_input", "config": {"title": "Confirm Site Type", "fields": [{"name": "site_type", "type": "select", "label": "Site Type", "options": ["landing", "content", "portfolio", "brochure"], "default_from": "classification.response.result.site_type"}, {"name": "recommended_builder", "type": "dynamic_select", "label": "Builder", "default_from": "classification.response.result.recommended_builder", "options_from": "available_builders.agents", "option_label_field": "display_name", "option_value_field": "type"}], "message": "Please confirm or adjust the site classification", "skip_if": "input_data.hitl_mode == auto", "request_type": "confirmation", "timeout_seconds": 86400}, "next_step": "fetch_questionnaire", "description": "Human confirms or adjusts the site type classification", "output_field": "confirmed_type"}, "hitl_review_brief": {"action": "request_human_input", "config": {"title": "Review Brief", "message": "Please review and adjust the briefing answers if needed", "skip_if": "input_data.hitl_mode == auto", "editable": true, "data_field": "brief_data", "request_type": "review", "timeout_seconds": 86400}, "next_step": "spawn_builder", "description": "Human reviews the completed brief", "output_field": "reviewed_brief"}, "fetch_questionnaire": {"action": "fetch_agent_questionnaire", "config": {"agent_type_field": "confirmed_type.recommended_builder"}, "next_step": "call_briefer", "description": "Fetch the briefing questionnaire for the target builder", "output_field": "questionnaire"}, "fetch_available_builders": {"action": "query_agent_definitions", "config": {"fields": ["type", "display_name", "description"], "filter": {"type_pattern": "%-builder"}}, "next_step": "spawn_classifier", "description": "Discover what builder agents are available", "output_field": "available_builders"}}, "start_step": "fetch_available_builders"}, "processing_mode": "orchestration", "timeout_seconds": 600} | t         | 2025-12-04 08:42:32.269013+00 | 2026-01-31 17:51:01.881339+00 |            | ["orchestration", "intake", "classification"] | docker.io/aqls/agent-chassis | v1.0.737  |         | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} |                | experimental | []          | {}                     |           0 | f           |                |
(1 row)


-- Add rerender step to intake-orchestrator workflow
-- After call_builder completes, rerender all pages to ensure consistent nav/contact info
--
-- Flow change:
--   Before: call_builder → complete
--   After:  call_builder → spawn_rerender → call_rerender → complete

-- ============================================================
-- 1. Add spawn_rerender step
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_rerender}',
        '{
            "action": "spawn_agent",
            "config": {
                "role": "rerenderer",
                "agent_type": "rerender-pages"
            },
            "description": "Spawn rerender agent to fix navigation consistency",
            "output_field": "rerender_agent",
            "next_step": "call_rerender"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'intake-orchestrator';

-- ============================================================
-- 2. Add call_rerender step
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_rerender}',
        '{
            "action": "call_agent",
            "config": {
                "agent_type": "rerender-pages",
                "target_role": "rerenderer",
                "input_mapping": {
                    "site_id": "build_result.site_record.site_id",
                    "domain": "build_result.site_record.domain",
                    "input_data": "input_data"
                },
                "timeout_seconds": 900
            },
            "description": "Rerender all pages with consistent nav and contact info",
            "output_field": "rerender_result",
            "next_step": "complete"
        }'::jsonb
                     )
WHERE type = 'intake-orchestrator';

-- ============================================================
-- 3. Update call_builder to go to spawn_rerender instead of complete
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_builder,next_step}',
        '"spawn_rerender"'
                     )
WHERE type = 'intake-orchestrator';

-- ============================================================
-- 4. Verify the update
-- ============================================================
SELECT
    type,
    display_name,
    default_config->'workflow'->'steps'->'call_builder'->>'next_step' as call_builder_next,
    default_config->'workflow'->'steps'->'spawn_rerender'->>'action' as spawn_rerender_action,
    default_config->'workflow'->'steps'->'call_rerender'->>'next_step' as call_rerender_next,
    version
FROM agent_definitions
WHERE type = 'intake-orchestrator';


-- input fields

-- Add rerender step to intake-orchestrator workflow
-- After call_builder completes, rerender all pages to ensure consistent nav/contact info
--
-- Flow change:
--   Before: call_builder → complete
--   After:  call_builder → spawn_rerender → call_rerender → complete

-- ============================================================
-- 1. Add spawn_rerender step
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_rerender}',
        '{
            "action": "spawn_agent",
            "config": {
                "role": "rerenderer",
                "agent_type": "rerender-pages"
            },
            "description": "Spawn rerender agent to fix navigation consistency",
            "output_field": "rerender_agent",
            "next_step": "call_rerender"
        }'::jsonb
                     ),
    version = version + 1,
    updated_at = NOW()
WHERE type = 'intake-orchestrator';

-- ============================================================
-- 2. Add call_rerender step
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_rerender}',
        '{
            "action": "call_agent",
            "config": {
                "agent_type": "rerender-pages",
                "target_role": "rerenderer",
                "input_mapping": {
                    "site_id": "build_result.response.site_record.site_id",
                    "domain": "build_result.response.site_record.domain",
                    "input_data": "input_data"
                },
                "timeout_seconds": 900
            },
            "description": "Rerender all pages with consistent nav and contact info",
            "output_field": "rerender_result",
            "next_step": "complete"
        }'::jsonb
                     )
WHERE type = 'intake-orchestrator';

-- ============================================================
-- 3. Update call_builder to go to spawn_rerender instead of complete
-- ============================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_builder,next_step}',
        '"spawn_rerender"'
                     )
WHERE type = 'intake-orchestrator';

-- ============================================================
-- 4. Verify the update
-- ============================================================
SELECT
    type,
    display_name,
    default_config->'workflow'->'steps'->'call_builder'->>'next_step' as call_builder_next,
    default_config->'workflow'->'steps'->'spawn_rerender'->>'action' as spawn_rerender_action,
    default_config->'workflow'->'steps'->'call_rerender'->>'next_step' as call_rerender_next,
    version
FROM agent_definitions
WHERE type = 'intake-orchestrator';

--------------

-- temporarily disable the rerender stage to fix original build
-- =============================================================
-- Fix intake-orchestrator: remove post-build rerender
--
-- Problem: intake-orchestrator calls rerender-pages after
-- pageflow-builder finishes, overwriting correctly-built pages
-- with broken output from the incomplete site_components path.
--
-- The pageflow-builder build_pages_loop uses:
--   assemble_page → InjectHeader/InjectFooter → correct templates
--
-- The rerender-pages agent uses:
--   rerender_single_page → site_components → wrong templates
--
-- Solution: change call_builder.next_step from spawn_rerender
-- to complete, skipping the rerender entirely.
-- The spawn_rerender and call_rerender steps remain in the
-- definition (harmless, just unreachable) - can be cleaned up
-- later when the rerender workflow is ready.
-- =============================================================

-- First, verify current flow: call_builder → spawn_rerender → call_rerender → complete
SELECT
    default_config->'workflow'->'steps'->'call_builder'->>'next_step' as call_builder_next,
    default_config->'workflow'->'steps'->'spawn_rerender'->>'next_step' as spawn_rerender_next,
    default_config->'workflow'->'steps'->'call_rerender'->>'next_step' as call_rerender_next
FROM agent_definitions
WHERE type = 'intake-orchestrator';

-- Update: call_builder now goes straight to complete
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_builder,next_step}',
        '"complete"'
                     ),
    updated_at = now()
WHERE type = 'intake-orchestrator';

-- Verify the change
SELECT
    default_config->'workflow'->'steps'->'call_builder'->>'next_step' as call_builder_next_step
FROM agent_definitions
WHERE type = 'intake-orchestrator';
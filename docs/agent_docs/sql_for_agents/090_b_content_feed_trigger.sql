--- ============================================================================
-- 027j — Content Feed Scheduling: trigger agent + scheduled task
-- ============================================================================
--
-- Creates:
--   1. content-feed-trigger agent definition
--   2. scheduled_tasks entry for periodic execution
--
-- NOTE: Run \d scheduled_tasks first to verify column names if the INSERT
-- at the bottom fails. The scheduled_tasks INSERT uses only 'name' which
-- we know exists from the UPDATE patterns used by other agents.
-- ============================================================================

-- Step 1: Insert content-feed-trigger agent definition
-- Column mapping checked against \d agent_definitions:
--   category (varchar 50)   = 'orchestrator' (not 'processing_type')
--   capabilities (jsonb)    = tags array
--   domain_tags (jsonb)     = semantic tags
--   agent_category (text)   = 'coordinator'
INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active, agent_category, status,
    capabilities, domain_tags,
    input_contract, output_contract, idle_timeout_seconds
)
VALUES (
           'content-feed-trigger',
           'Content Feed Trigger',
           'Heartbeat agent: finds sites recommended for news feeds and dispatches content-feed-orchestrator per site. Runs on a 6-hour schedule.',
           'orchestrator',
           '{
             "processing_mode": "orchestrator",
             "timeout_seconds": 900,
             "workflow": {
               "start_step": "find_news_sites",
               "steps": {

                 "find_news_sites": {
                   "action": "query_database",
                   "config": {
                     "query": "SELECT DISTINCT s.id::text as site_id, s.domain FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = ''classification'' AND ss.is_current = true AND (ss.data->''content_features''->''news_feed''->>''recommended'')::boolean = true WHERE s.deleted_at IS NULL AND EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = ''deployed'') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW()))) ORDER BY s.domain LIMIT 5",
                     "output_format": "array"
                   },
                   "next_step": "check_has_sites",
                   "description": "Find sites recommended for news that need a feed refresh",
                   "output_field": "news_sites"
                 },

                 "check_has_sites": {
                   "action": "evaluate_condition",
                   "config": {
                     "condition_field": "news_sites.count",
                     "conditions": {"0": "notify_scheduler_idle"},
                     "default": "process_sites"
                   },
                   "description": "Skip if no sites need news refresh"
                 },

                 "process_sites": {
                   "action": "loop",
                   "config": {
                     "items_field": "news_sites",
                     "item_variable": "current_site",
                     "max_iterations": 5,
                     "continue_on_error": true,
                     "sub_workflow": {
                       "start_step": "spawn_orchestrator",
                       "steps": {
                         "spawn_orchestrator": {
                           "action": "spawn_agent",
                           "config": {
                             "role": "feed_orchestrator",
                             "agent_type": "content-feed-orchestrator"
                           },
                           "next_step": "call_orchestrator",
                           "description": "Spawn content-feed-orchestrator for this site",
                           "output_field": "orch_spawned"
                         },
                         "call_orchestrator": {
                           "action": "call_agent",
                           "config": {
                             "target_role": "feed_orchestrator",
                             "input_mapping": {
                               "site_id": "current_site.site_id",
                               "domain": "current_site.domain"
                             },
                             "timeout_seconds": 600
                           },
                           "next_step": "done",
                           "error_step": "done",
                           "description": "Call content-feed-orchestrator for this site",
                           "output_field": "orch_result"
                         },
                         "done": {
                           "action": "loop_complete",
                           "description": "Move to next site"
                         }
                       }
                     }
                   },
                   "next_step": "notify_scheduler",
                   "description": "Process each site: spawn + call content-feed-orchestrator",
                   "output_field": "feed_results"
                 },

                 "notify_scheduler": {
                   "action": "query_database",
                   "config": {
                     "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''content-feed-refresh''",
                     "output_format": "object"
                   },
                   "next_step": "complete",
                   "description": "Tell scheduler this execution finished",
                   "output_field": "scheduler_notified"
                 },

                 "notify_scheduler_idle": {
                   "action": "query_database",
                   "config": {
                     "query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''content-feed-refresh''",
                     "output_format": "object"
                   },
                   "next_step": "complete_idle",
                   "description": "Tell scheduler this execution finished (no sites)",
                   "output_field": "scheduler_notified"
                 },

                 "complete": {
                   "action": "complete_workflow",
                   "config": {
                     "output_fields": ["news_sites", "feed_results"]
                   },
                   "description": "Feed trigger complete"
                 },

                 "complete_idle": {
                   "action": "complete_workflow",
                   "config": {
                     "output_fields": ["news_sites"],
                     "success_message": "No sites need news refresh"
                   },
                   "description": "No sites needed feed refresh this cycle"
                 }
               }
             }
           }'::jsonb,
           true,                -- is_active
           'coordinator',       -- agent_category
           'experimental',      -- status
           '["feed", "trigger", "scheduling"]'::jsonb,   -- capabilities
           '["news", "content-feed", "scheduling"]'::jsonb, -- domain_tags
           '{"required": [], "optional": [], "description": "No input required. Queries sites from database."}'::jsonb,
           '{"produces": {"news_sites": "sites found needing feed refresh", "feed_results": "results from per-site orchestrator calls"}}'::jsonb,
           0                    -- idle_timeout_seconds
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       updated_at = NOW();

-- Step 2: Insert scheduled_tasks entry
-- Using only 'name' column which we know exists. Run \d scheduled_tasks
-- to see full schema if you need to add schedule_interval or other fields.
INSERT INTO scheduled_tasks (name)
VALUES ('content-feed-refresh')
    ON CONFLICT (name) DO NOTHING;

-- Verify
SELECT type, status, agent_category
FROM agent_definitions
WHERE type = 'content-feed-trigger' AND deleted_at IS NULL;

SELECT * FROM scheduled_tasks WHERE name = 'content-feed-refresh';

-- ============================================================================
-- 026b: Agent definitions for news feed pipeline (corrected)
--
-- Fixes applied per dev guide (001e) and contracts (003e):
--   - model alias: claude-sonnet-4-6 (not full version string)
--   - api_key_env_var in every ai_service config
--   - agent_category from allowed set (check_ad_category constraint)
--   - status set to 'experimental'
--   - input_contract / output_contract defined
--   - ON CONFLICT (type, version) matches actual unique constraint
--   - {{if}} guards on {{range}} in templates (lesson #7)
--   - parameterised queries with $1 (lesson #1)
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1. feed-ingester
-- ---------------------------------------------------------------------------
-- Receives: site_id, source_id, source_type, source_config
-- Routes by source_type to the appropriate fetch action, then writes items.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-ingester',
             'Feed Ingester',
             'Fetches content from a single source (RSS, news search, LLM news, scrape) and writes to content_feed_items',
             'code-driven',
             'executor',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id", "source_id", "source_type", "source_config"]}'::jsonb,
             '{"produces": ["write_result", "timestamp_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "route_by_type",
                     "steps": {
                         "route_by_type": {
                             "action": "conditional_route",
                             "config": {
                                 "condition_field": "source_type",
                                 "routes": {
                                     "rss": "fetch_rss",
                                     "api_news": "fetch_llm_news",
                                     "news_search": "search_news",
                                     "scrape": "scrape_source",
                                     "default": "fetch_rss"
                                 }
                             },
                             "description": "Route to fetch action based on source type"
                         },

                         "fetch_rss": {
                             "action": "fetch_rss",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch and parse RSS/Atom feed"
                         },

                         "fetch_llm_news": {
                             "action": "fetch_llm_news",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch news via Grok/xAI API"
                         },

                         "search_news": {
                             "action": "web_search",
                             "config": {
                                 "query_field": "input_data.source_config.queries.0",
                                 "search_type": "news",
                                 "num_results": 10
                             },
                             "output_field": "search_results",
                             "next_step": "normalize_search",
                             "description": "Search for news via web search adapter"
                         },

                         "normalize_search": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "search",
                                 "results_field": "search_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize search results to feed items"
                         },

                         "scrape_source": {
                             "action": "firecrawl_scrape",
                             "config": {
                                 "url_field": "input_data.source_config.urls.0"
                             },
                             "output_field": "scrape_results",
                             "next_step": "normalize_scrape",
                             "description": "Scrape target news page"
                         },

                         "normalize_scrape": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "scrape",
                                 "results_field": "scrape_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize scrape results to feed items"
                         },

                         "write_items": {
                             "action": "write_feed_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "source_id": "input_data.source_id",
                                 "items_field": "fetched_items.items",
                                 "source_type": "input_data.source_type"
                             },
                             "output_field": "write_result",
                             "next_step": "update_timestamps",
                             "description": "Write normalised items to content_feed_items"
                         },

                         "update_timestamps": {
                             "action": "update_source_timestamps",
                             "config": {
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "timestamp_result",
                             "next_step": "complete",
                             "description": "Update source fetch timestamps"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 2. content-feed-orchestrator
-- ---------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'content-feed-orchestrator',
             'Content Feed Orchestrator',
             'Per-site orchestrator that checks for due content sources and spawns feed-ingester agents',
             'code-driven',
             'coordinator',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"]}'::jsonb,
             '{"produces": ["dispatch_result"]}'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "dispatch_sources",
                     "steps": {
                         "dispatch_sources": {
                             "action": "dispatch_feed_sources",
                             "config": {
                                 "site_id": "input_data.site_id"
                             },
                             "output_field": "dispatch_result",
                             "next_step": "complete",
                             "description": "Load due sources and spawn feed-ingester per source"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 3. feed-triage (stub — next phase)
-- ---------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-triage',
             'Feed Triage',
             'Scores ingested feed items for relevance to site vertical, deduplicates, filters noise',
             'code-driven',
             'analyst',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"], "optional": ["vertical"]}'::jsonb,
             '{"produces": ["triage_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "ai_service": {
                     "model": "claude-sonnet-4-6",
                     "provider": "anthropic",
                     "max_tokens": 4000,
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "workflow": {
                     "start_step": "load_items",
                     "steps": {
                         "load_items": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT id::text, source_title, source_summary, source_url, source_published_at::text FROM content_feed_items WHERE site_id = $1 AND status = ''ingested'' ORDER BY created_at DESC LIMIT 50",
                                 "params": ["input_data.site_id"],
                                 "output_format": "array"
                             },
                             "output_field": "pending_items",
                             "next_step": "check_has_items",
                             "description": "Load unscored ingested items"
                         },

                         "check_has_items": {
                             "action": "evaluate_condition",
                             "config": {
                                 "condition": "pending_items != null && len(pending_items) > 0",
                                 "true_step": "score_relevance",
                                 "false_step": "complete"
                             },
                             "description": "Skip if no items to triage"
                         },

                         "score_relevance": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-6",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "prompt_template": "You are a content relevance scorer for a {{.input_data.vertical}} website.\n\nScore each news item 0-100 for relevance to this vertical. Return ONLY a JSON array.\nEach object: {\"id\": \"uuid\", \"score\": 0-100, \"reason\": \"brief\", \"topics\": [\"tag1\"]}\n\n{{if .pending_items}}Items:\n{{range .pending_items}}ID: {{.id}}\nTitle: {{.source_title}}\nSummary: {{.source_summary}}\n---\n{{end}}{{end}}",
                                 "response_format": "json"
                             },
                             "output_field": "scores",
                             "next_step": "apply_scores",
                             "description": "LLM scores items for relevance"
                         },

                         "apply_scores": {
                             "action": "apply_feed_scores",
                             "config": {
                                 "scores_field": "scores",
                                 "relevance_threshold": 40
                             },
                             "output_field": "triage_result",
                             "next_step": "complete",
                             "description": "Update items with scores, mark relevant/rejected"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


--

-- single url at a time
-- more than one url means more than one content_sources row

UPDATE content_sources
SET config = '{"url": "https://oilprice.com/Latest-Energy-News/World-News", "scrape_config": {"only_main_content": true}, "max_items": 5}'::jsonb,
    next_fetch_at = now()
WHERE name = 'OilPrice Latest News Scrape'
  AND site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
    RETURNING id, name, config;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,scrape_source,config}',
        '{"url_field": "input_data.source_config.url", "scrape_config": {"only_main_content": true}}'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'feed-ingester' AND deleted_at IS NULL
    RETURNING type, default_config->'workflow'->'steps'->'scrape_source' as scrape_step;

---

-- ============================================================================
-- 026b: Agent definitions for news feed pipeline (corrected)
--
-- Fixes applied per dev guide (001e) and contracts (003e):
--   - model alias: claude-sonnet-4-6 (not full version string)
--   - api_key_env_var in every ai_service config
--   - agent_category from allowed set (check_ad_category constraint)
--   - status set to 'experimental'
--   - input_contract / output_contract defined
--   - ON CONFLICT (type, version) matches actual unique constraint
--   - {{if}} guards on {{range}} in templates (lesson #7)
--   - parameterised queries with $1 (lesson #1)
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1. feed-ingester
-- ---------------------------------------------------------------------------
-- Receives: site_id, source_id, source_type, source_config
-- Routes by source_type to the appropriate fetch action, then writes items.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-ingester',
             'Feed Ingester',
             'Fetches content from a single source (RSS, news search, LLM news, scrape) and writes to content_feed_items',
             'code-driven',
             'executor',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id", "source_id", "source_type", "source_config"]}'::jsonb,
             '{"produces": ["write_result", "timestamp_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "route_by_type",
                     "steps": {
                         "route_by_type": {
                             "action": "conditional_route",
                             "config": {
                                 "condition_field": "source_type",
                                 "routes": {
                                     "rss": "fetch_rss",
                                     "api_news": "fetch_llm_news",
                                     "news_search": "search_news",
                                     "scrape": "scrape_source",
                                     "default": "fetch_rss"
                                 }
                             },
                             "description": "Route to fetch action based on source type"
                         },

                         "fetch_rss": {
                             "action": "fetch_rss",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch and parse RSS/Atom feed"
                         },

                         "fetch_llm_news": {
                             "action": "fetch_llm_news",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch news via Grok/xAI API"
                         },

                         "search_news": {
                             "action": "web_search",
                             "config": {
                                 "query_field": "input_data.source_config.queries.0",
                                 "search_type": "news",
                                 "num_results": 10
                             },
                             "output_field": "search_results",
                             "next_step": "normalize_search",
                             "description": "Search for news via web search adapter"
                         },

                         "normalize_search": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "search",
                                 "results_field": "search_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize search results to feed items"
                         },

                         "scrape_source": {
                             "action": "firecrawl_scrape",
                             "config": {
                                 "url_field": "input_data.source_config.url",
                                 "scrape_config": {"only_main_content": true}
                             },
                             "output_field": "scrape_results",
                             "next_step": "normalize_scrape",
                             "description": "Scrape target news page"
                         },

                         "normalize_scrape": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "scrape",
                                 "results_field": "scrape_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize scrape results to feed items"
                         },

                         "write_items": {
                             "action": "write_feed_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "source_id": "input_data.source_id",
                                 "items_field": "fetched_items.items",
                                 "source_type": "input_data.source_type"
                             },
                             "output_field": "write_result",
                             "next_step": "update_timestamps",
                             "description": "Write normalised items to content_feed_items"
                         },

                         "update_timestamps": {
                             "action": "update_source_timestamps",
                             "config": {
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "timestamp_result",
                             "next_step": "complete",
                             "description": "Update source fetch timestamps"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 2. content-feed-orchestrator
-- ---------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'content-feed-orchestrator',
             'Content Feed Orchestrator',
             'Per-site orchestrator that checks for due content sources and spawns feed-ingester agents',
             'code-driven',
             'coordinator',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"]}'::jsonb,
             '{"produces": ["dispatch_result"]}'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "dispatch_sources",
                     "steps": {
                         "dispatch_sources": {
                             "action": "dispatch_feed_sources",
                             "config": {
                                 "site_id": "input_data.site_id"
                             },
                             "output_field": "dispatch_result",
                             "next_step": "complete",
                             "description": "Load due sources and spawn feed-ingester per source"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 3. feed-triage (stub — next phase)
-- ---------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-triage',
             'Feed Triage',
             'Scores ingested feed items for relevance to site vertical, deduplicates, filters noise',
             'code-driven',
             'analyst',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"], "optional": ["vertical"]}'::jsonb,
             '{"produces": ["triage_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "ai_service": {
                     "model": "claude-sonnet-4-6",
                     "provider": "anthropic",
                     "max_tokens": 4000,
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "workflow": {
                     "start_step": "load_items",
                     "steps": {
                         "load_items": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT id::text, source_title, source_summary, source_url, source_published_at::text FROM content_feed_items WHERE site_id = $1 AND status = ''ingested'' ORDER BY created_at DESC LIMIT 50",
                                 "params": ["input_data.site_id"],
                                 "output_format": "array"
                             },
                             "output_field": "pending_items",
                             "next_step": "check_has_items",
                             "description": "Load unscored ingested items"
                         },

                         "check_has_items": {
                             "action": "evaluate_condition",
                             "config": {
                                 "condition": "pending_items != null && len(pending_items) > 0",
                                 "true_step": "score_relevance",
                                 "false_step": "complete"
                             },
                             "description": "Skip if no items to triage"
                         },

                         "score_relevance": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-6",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "prompt_template": "You are a content relevance scorer for a {{.input_data.vertical}} website.\n\nScore each news item 0-100 for relevance to this vertical. Return ONLY a JSON array.\nEach object: {\"id\": \"uuid\", \"score\": 0-100, \"reason\": \"brief\", \"topics\": [\"tag1\"]}\n\n{{if .pending_items}}Items:\n{{range .pending_items}}ID: {{.id}}\nTitle: {{.source_title}}\nSummary: {{.source_summary}}\n---\n{{end}}{{end}}",
                                 "response_format": "json"
                             },
                             "output_field": "scores",
                             "next_step": "apply_scores",
                             "description": "LLM scores items for relevance"
                         },

                         "apply_scores": {
                             "action": "apply_feed_scores",
                             "config": {
                                 "scores_field": "scores",
                                 "relevance_threshold": 40
                             },
                             "output_field": "triage_result",
                             "next_step": "complete",
                             "description": "Update items with scores, mark relevant/rejected"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();

---
--

-- ============================================================================
-- 026b: Agent definitions for news feed pipeline (corrected)
--
-- Fixes applied per dev guide (001e) and contracts (003e):
--   - model alias: claude-sonnet-4-6 (not full version string)
--   - api_key_env_var in every ai_service config
--   - agent_category from allowed set (check_ad_category constraint)
--   - status set to 'experimental'
--   - input_contract / output_contract defined
--   - ON CONFLICT (type, version) matches actual unique constraint
--   - {{if}} guards on {{range}} in templates (lesson #7)
--   - parameterised queries with $1 (lesson #1)
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1. feed-ingester
-- ---------------------------------------------------------------------------
-- Receives: site_id, source_id, source_type, source_config
-- Routes by source_type to the appropriate fetch action, then writes items.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-ingester',
             'Feed Ingester',
             'Fetches content from a single source (RSS, news search, LLM news, scrape) and writes to content_feed_items',
             'code-driven',
             'executor',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id", "source_id", "source_type", "source_config"]}'::jsonb,
             '{"produces": ["write_result", "timestamp_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "route_by_type",
                     "steps": {
                         "route_by_type": {
                             "action": "conditional_route",
                             "config": {
                                 "condition_field": "source_type",
                                 "routes": {
                                     "rss": "fetch_rss",
                                     "api_news": "fetch_llm_news",
                                     "news_search": "search_news",
                                     "scrape": "scrape_source",
                                     "default": "fetch_rss"
                                 }
                             },
                             "description": "Route to fetch action based on source type"
                         },

                         "fetch_rss": {
                             "action": "fetch_rss",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch and parse RSS/Atom feed"
                         },

                         "fetch_llm_news": {
                             "action": "fetch_llm_news",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch news via Grok/xAI API"
                         },

                         "search_news": {
                             "action": "web_search",
                             "config": {
                                 "query_field": "input_data.source_config.queries.0",
                                 "search_type": "news",
                                 "num_results": 10
                             },
                             "output_field": "search_results",
                             "next_step": "normalize_search",
                             "description": "Search for news via web search adapter"
                         },

                         "normalize_search": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "search",
                                 "results_field": "search_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize search results to feed items"
                         },

                         "scrape_source": {
                             "action": "firecrawl_scrape",
                             "config": {
                                 "url_field": "input_data.source_config.url",
                                 "scrape_config": {"only_main_content": true}
                             },
                             "output_field": "scrape_results",
                             "next_step": "normalize_scrape",
                             "description": "Scrape target news page"
                         },

                         "normalize_scrape": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "scrape",
                                 "results_field": "scrape_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize scrape results to feed items"
                         },

                         "write_items": {
                             "action": "write_feed_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "source_id": "input_data.source_id",
                                 "items_field": "fetched_items.items",
                                 "source_type": "input_data.source_type"
                             },
                             "output_field": "write_result",
                             "next_step": "update_timestamps",
                             "description": "Write normalised items to content_feed_items"
                         },

                         "update_timestamps": {
                             "action": "update_source_timestamps",
                             "config": {
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "timestamp_result",
                             "next_step": "complete",
                             "description": "Update source fetch timestamps"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 2. content-feed-orchestrator
-- ---------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'content-feed-orchestrator',
             'Content Feed Orchestrator',
             'Per-site orchestrator that checks for due content sources and spawns feed-ingester agents',
             'code-driven',
             'coordinator',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"]}'::jsonb,
             '{"produces": ["dispatch_result"]}'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "dispatch_sources",
                     "steps": {
                         "dispatch_sources": {
                             "action": "dispatch_feed_sources",
                             "config": {
                                 "site_id": "input_data.site_id"
                             },
                             "output_field": "dispatch_result",
                             "next_step": "complete",
                             "description": "Load due sources and spawn feed-ingester per source"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 3. feed-triage (stub — next phase)
-- ---------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-triage',
             'Feed Triage',
             'Scores ingested feed items for relevance to site vertical, deduplicates, filters noise',
             'code-driven',
             'analyst',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"], "optional": ["vertical"]}'::jsonb,
             '{"produces": ["triage_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "ai_service": {
                     "model": "claude-sonnet-4-6",
                     "provider": "anthropic",
                     "max_tokens": 4000,
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "workflow": {
                     "start_step": "load_items",
                     "steps": {
                         "load_items": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT id::text, source_title, source_summary, source_url, source_published_at::text FROM content_feed_items WHERE site_id = $1 AND status = ''ingested'' ORDER BY created_at DESC LIMIT 50",
                                 "params": ["input_data.site_id"],
                                 "output_format": "array"
                             },
                             "output_field": "pending_items",
                             "next_step": "check_has_items",
                             "description": "Load unscored ingested items"
                         },

                         "check_has_items": {
                             "action": "evaluate_condition",
                             "config": {
                                 "condition": "pending_items != null && len(pending_items) > 0",
                                 "true_step": "score_relevance",
                                 "false_step": "complete"
                             },
                             "description": "Skip if no items to triage"
                         },

                         "score_relevance": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-6",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "prompt_template": "You are a content relevance scorer for a {{.input_data.vertical}} website.\n\nScore each news item 0-100 for relevance to this vertical. Return ONLY a JSON array.\nEach object: {\"id\": \"uuid\", \"score\": 0-100, \"reason\": \"brief\", \"topics\": [\"tag1\"]}\n\n{{if .pending_items}}Items:\n{{range .pending_items}}ID: {{.id}}\nTitle: {{.source_title}}\nSummary: {{.source_summary}}\n---\n{{end}}{{end}}",
                                 "response_format": "json"
                             },
                             "output_field": "scores",
                             "next_step": "apply_scores",
                             "description": "LLM scores items for relevance"
                         },

                         "apply_scores": {
                             "action": "apply_feed_scores",
                             "config": {
                                 "scores_field": "scores",
                                 "relevance_threshold": 40
                             },
                             "output_field": "triage_result",
                             "next_step": "complete",
                             "description": "Update items with scores, mark relevant/rejected"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();

---
-- path problems

-- Fix the news_search source to use a single query string
UPDATE content_sources
SET config = '{"query": "UK wholesale gas prices news", "num_results": 5}'::jsonb,
    next_fetch_at = now()
WHERE name = 'Gas wholesale energy news search'
  AND site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
    RETURNING id, name, config;

-- Fix the agent definition for the search_news step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,search_news,config}',
        '{"query_field": "input_data.source_config.query", "search_type": "news", "num_results": 10}'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'feed-ingester' AND deleted_at IS NULL
    RETURNING type, default_config->'workflow'->'steps'->'search_news' as search_step;

--

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,search_news}',
                '{"action": "fetch_news_search", "config": {"source_config": "input_data.source_config"}, "output_field": "search_results", "next_step": "normalize_search", "description": "Search for news via web search adapter"}'::jsonb
        ),
        '{workflow,steps,scrape_source}',
        '{"action": "fetch_scrape", "config": {"source_config": "input_data.source_config"}, "output_field": "scrape_results", "next_step": "normalize_scrape", "description": "Scrape target news page"}'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'feed-ingester' AND deleted_at IS NULL
    RETURNING type,
  default_config->'workflow'->'steps'->'search_news'->'action' as search_action,
  default_config->'workflow'->'steps'->'scrape_source'->'action' as scrape_action;

--
-- ============================================================================
-- 026b: Agent definitions for news feed pipeline (corrected)
--
-- Fixes applied per dev guide (001e) and contracts (003e):
--   - model alias: claude-sonnet-4-6 (not full version string)
--   - api_key_env_var in every ai_service config
--   - agent_category from allowed set (check_ad_category constraint)
--   - status set to 'experimental'
--   - input_contract / output_contract defined
--   - ON CONFLICT (type, version) matches actual unique constraint
--   - {{if}} guards on {{range}} in templates (lesson #7)
--   - parameterised queries with $1 (lesson #1)
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1. feed-ingester
-- ---------------------------------------------------------------------------
-- Receives: site_id, source_id, source_type, source_config
-- Routes by source_type to the appropriate fetch action, then writes items.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-ingester',
             'Feed Ingester',
             'Fetches content from a single source (RSS, news search, LLM news, scrape) and writes to content_feed_items',
             'code-driven',
             'executor',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id", "source_id", "source_type", "source_config"]}'::jsonb,
             '{"produces": ["write_result", "timestamp_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "route_by_type",
                     "steps": {
                         "route_by_type": {
                             "action": "conditional_route",
                             "config": {
                                 "condition_field": "source_type",
                                 "routes": {
                                     "rss": "fetch_rss",
                                     "api_news": "fetch_llm_news",
                                     "news_search": "search_news",
                                     "scrape": "scrape_source",
                                     "default": "fetch_rss"
                                 }
                             },
                             "description": "Route to fetch action based on source type"
                         },

                         "fetch_rss": {
                             "action": "fetch_rss",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch and parse RSS/Atom feed"
                         },

                         "fetch_llm_news": {
                             "action": "fetch_llm_news",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch news via Grok/xAI API"
                         },

                         "search_news": {
                             "action": "fetch_news_search",
                             "config": {
                                 "source_config": "input_data.source_config"
                             },
                             "output_field": "search_results",
                             "next_step": "normalize_search",
                             "description": "Search for news via web search adapter"
                         },

                         "normalize_search": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "search",
                                 "results_field": "search_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize search results to feed items"
                         },

                         "scrape_source": {
                             "action": "fetch_scrape",
                             "config": {
                                 "source_config": "input_data.source_config"
                             },
                             "output_field": "scrape_results",
                             "next_step": "normalize_scrape",
                             "description": "Scrape target news page"
                         },

                         "normalize_scrape": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "scrape",
                                 "results_field": "scrape_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize scrape results to feed items"
                         },

                         "write_items": {
                             "action": "write_feed_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "source_id": "input_data.source_id",
                                 "items_field": "fetched_items.items",
                                 "source_type": "input_data.source_type"
                             },
                             "output_field": "write_result",
                             "next_step": "update_timestamps",
                             "description": "Write normalised items to content_feed_items"
                         },

                         "update_timestamps": {
                             "action": "update_source_timestamps",
                             "config": {
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "timestamp_result",
                             "next_step": "complete",
                             "description": "Update source fetch timestamps"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 2. content-feed-orchestrator
-- ---------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'content-feed-orchestrator',
             'Content Feed Orchestrator',
             'Per-site orchestrator that checks for due content sources and spawns feed-ingester agents',
             'code-driven',
             'coordinator',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"]}'::jsonb,
             '{"produces": ["dispatch_result"]}'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "dispatch_sources",
                     "steps": {
                         "dispatch_sources": {
                             "action": "dispatch_feed_sources",
                             "config": {
                                 "site_id": "input_data.site_id"
                             },
                             "output_field": "dispatch_result",
                             "next_step": "complete",
                             "description": "Load due sources and spawn feed-ingester per source"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 3. feed-triage (stub — next phase)
-- ---------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-triage',
             'Feed Triage',
             'Scores ingested feed items for relevance to site vertical, deduplicates, filters noise',
             'code-driven',
             'analyst',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"], "optional": ["vertical"]}'::jsonb,
             '{"produces": ["triage_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "ai_service": {
                     "model": "claude-sonnet-4-6",
                     "provider": "anthropic",
                     "max_tokens": 4000,
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "workflow": {
                     "start_step": "load_items",
                     "steps": {
                         "load_items": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT id::text, source_title, source_summary, source_url, source_published_at::text FROM content_feed_items WHERE site_id = $1 AND status = ''ingested'' ORDER BY created_at DESC LIMIT 50",
                                 "params": ["input_data.site_id"],
                                 "output_format": "array"
                             },
                             "output_field": "pending_items",
                             "next_step": "check_has_items",
                             "description": "Load unscored ingested items"
                         },

                         "check_has_items": {
                             "action": "evaluate_condition",
                             "config": {
                                 "condition": "pending_items != null && len(pending_items) > 0",
                                 "true_step": "score_relevance",
                                 "false_step": "complete"
                             },
                             "description": "Skip if no items to triage"
                         },

                         "score_relevance": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-6",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "prompt_template": "You are a content relevance scorer for a {{.input_data.vertical}} website.\n\nScore each news item 0-100 for relevance to this vertical. Return ONLY a JSON array.\nEach object: {\"id\": \"uuid\", \"score\": 0-100, \"reason\": \"brief\", \"topics\": [\"tag1\"]}\n\n{{if .pending_items}}Items:\n{{range .pending_items}}ID: {{.id}}\nTitle: {{.source_title}}\nSummary: {{.source_summary}}\n---\n{{end}}{{end}}",
                                 "response_format": "json"
                             },
                             "output_field": "scores",
                             "next_step": "apply_scores",
                             "description": "LLM scores items for relevance"
                         },

                         "apply_scores": {
                             "action": "apply_feed_scores",
                             "config": {
                                 "scores_field": "scores",
                                 "relevance_threshold": 40
                             },
                             "output_field": "triage_result",
                             "next_step": "complete",
                             "description": "Update items with scores, mark relevant/rejected"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ============================================================================
-- 026b: Agent definitions for news feed pipeline (corrected)
--
-- Fixes applied per dev guide (001e) and contracts (003e):
--   - model alias: claude-sonnet-4-6 (not full version string)
--   - api_key_env_var in every ai_service config
--   - agent_category from allowed set (check_ad_category constraint)
--   - status set to 'experimental'
--   - input_contract / output_contract defined
--   - ON CONFLICT (type, version) matches actual unique constraint
--   - {{if}} guards on {{range}} in templates (lesson #7)
--   - parameterised queries with $1 (lesson #1)
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1. feed-ingester
-- ---------------------------------------------------------------------------
-- Receives: site_id, source_id, source_type, source_config
-- Routes by source_type to the appropriate fetch action, then writes items.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-ingester',
             'Feed Ingester',
             'Fetches content from a single source (RSS, news search, LLM news, scrape) and writes to content_feed_items',
             'code-driven',
             'executor',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id", "source_id", "source_type", "source_config"]}'::jsonb,
             '{"produces": ["write_result", "timestamp_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "route_by_type",
                     "steps": {
                         "route_by_type": {
                             "action": "conditional_route",
                             "config": {
                                 "condition_field": "source_type",
                                 "routes": {
                                     "rss": "fetch_rss",
                                     "api_news": "fetch_llm_news",
                                     "news_search": "search_news",
                                     "scrape": "scrape_source",
                                     "default": "fetch_rss"
                                 }
                             },
                             "description": "Route to fetch action based on source type"
                         },

                         "fetch_rss": {
                             "action": "fetch_rss",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch and parse RSS/Atom feed"
                         },

                         "fetch_llm_news": {
                             "action": "fetch_llm_news",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch news via Grok/xAI API"
                         },

                         "search_news": {
                             "action": "fetch_news_search",
                             "config": {
                                 "source_config": "input_data.source_config"
                             },
                             "output_field": "search_results",
                             "next_step": "normalize_search",
                             "description": "Search for news via web search adapter"
                         },

                         "normalize_search": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "search",
                                 "results_field": "search_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize search results to feed items"
                         },

                         "scrape_source": {
                             "action": "fetch_scrape",
                             "config": {
                                 "source_config": "input_data.source_config"
                             },
                             "output_field": "scrape_results",
                             "next_step": "normalize_scrape",
                             "description": "Scrape target news page"
                         },

                         "normalize_scrape": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "scrape",
                                 "results_field": "scrape_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize scrape results to feed items"
                         },

                         "write_items": {
                             "action": "write_feed_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "source_id": "input_data.source_id",
                                 "items_field": "fetched_items.items",
                                 "source_type": "input_data.source_type"
                             },
                             "output_field": "write_result",
                             "next_step": "update_timestamps",
                             "description": "Write normalised items to content_feed_items"
                         },

                         "update_timestamps": {
                             "action": "update_source_timestamps",
                             "config": {
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "timestamp_result",
                             "next_step": "complete",
                             "description": "Update source fetch timestamps"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 2. content-feed-orchestrator
-- ---------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'content-feed-orchestrator',
             'Content Feed Orchestrator',
             'Per-site orchestrator that checks for due content sources and spawns feed-ingester agents',
             'code-driven',
             'coordinator',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"]}'::jsonb,
             '{"produces": ["dispatch_result"]}'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "dispatch_sources",
                     "steps": {
                         "dispatch_sources": {
                             "action": "dispatch_feed_sources",
                             "config": {
                                 "site_id": "input_data.site_id"
                             },
                             "output_field": "dispatch_result",
                             "next_step": "complete",
                             "description": "Load due sources and spawn feed-ingester per source"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 3. feed-triage — scores ingested items against site spec
-- ---------------------------------------------------------------------------
-- Workflow: load_items → check_has_items → read_site_spec → score_relevance → apply_scores → complete
-- Reads site spec before scoring so the LLM has identity, classification,
-- content_direction, and legal_rules context for relevance judgement.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-triage',
             'Feed Triage',
             'Scores ingested feed items for relevance to site spec — identity, vertical, values, legal rules. Pre-display gate.',
             'code-driven',
             'analyst',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"], "optional": ["relevance_threshold"]}'::jsonb,
             '{"produces": ["triage_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "ai_service": {
                     "model": "claude-sonnet-4-6",
                     "provider": "anthropic",
                     "max_tokens": 4000,
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "workflow": {
                     "start_step": "load_items",
                     "steps": {
                         "load_items": {
                             "action": "load_feed_items_for_triage",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "max_items": 50
                             },
                             "output_field": "pending_items",
                             "next_step": "check_has_items",
                             "description": "Load unscored ingested items with source metadata"
                         },

                         "check_has_items": {
                             "action": "evaluate_condition",
                             "config": {
                                 "condition_field": "pending_items.count",
                                 "conditions": {
                                     "0": "complete"
                                 },
                                 "default": "read_site_spec"
                             },
                             "description": "Skip if no items to triage"
                         },

                         "read_site_spec": {
                             "action": "read_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id"
                             },
                             "output_field": "site_spec",
                             "next_step": "score_relevance",
                             "description": "Load all site spec aspects for relevance context"
                         },

                         "score_relevance": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-6",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "prompt_template": "You are a content relevance filter for the website {{.input_data.site_id}}.\n\nYour job: score each news item for how relevant and appropriate it is for this site and its users.\n\n## Site Context\n\n{{if .site_spec.data}}{{if .site_spec.data.identity}}### Identity\n{{range $k, $v := .site_spec.data.identity}}{{$k}}: {{$v}}\n{{end}}{{end}}\n\n{{if .site_spec.data.classification}}### Classification\n{{range $k, $v := .site_spec.data.classification}}{{$k}}: {{$v}}\n{{end}}{{end}}\n\n{{if .site_spec.data.content_direction}}### Content Direction\n{{range $k, $v := .site_spec.data.content_direction}}{{$k}}: {{$v}}\n{{end}}{{end}}\n\n{{if .site_spec.data.legal_rules}}### Legal Rules\nForbidden phrases: {{.site_spec.data.legal_rules.forbidden_phrases}}\n{{end}}{{end}}\n\n## Items to Score\n\n{{if .pending_items.items}}{{range .pending_items.items}}ID: {{.id}}\nTitle: {{.source_title}}\nSummary: {{.source_summary}}\nSource: {{.source_name}} ({{.source_type}})\n---\n{{end}}{{end}}\n\n## Scoring Guide\n\n- 80-100: Directly relevant — covers this site''s industry, audience would want to read this\n- 50-79: Tangentially relevant — adjacent topic, broader market context, general industry news\n- 20-49: Weak relevance — same sector but wrong geography, audience, or focus\n- 0-19: Not relevant — wrong industry, spam, clickbait\n\nAlso check each item against the site''s values and legal rules. If an item conflicts with these (inappropriate tone, forbidden topics, legal risk), set \"flagged\": true regardless of relevance score.\n\nReturn ONLY a JSON array, no other text:\n[{\"id\": \"uuid\", \"score\": 0-100, \"reason\": \"one sentence\", \"topics\": [\"tag1\", \"tag2\"], \"flagged\": false}]",
                                 "response_format": "json"
                             },
                             "output_field": "scores",
                             "next_step": "apply_scores",
                             "description": "LLM scores items for relevance against site spec"
                         },

                         "apply_scores": {
                             "action": "apply_feed_scores",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "scores_field": "scores"
                             },
                             "output_field": "triage_result",
                             "next_step": "complete",
                             "description": "Update items with scores and status"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


--

-- ============================================================================
-- 026b: Agent definitions for news feed pipeline (corrected)
--
-- Fixes applied per dev guide (001e) and contracts (003e):
--   - model alias: claude-sonnet-4-6 (not full version string)
--   - api_key_env_var in every ai_service config
--   - agent_category from allowed set (check_ad_category constraint)
--   - status set to 'experimental'
--   - input_contract / output_contract defined
--   - ON CONFLICT (type, version) matches actual unique constraint
--   - {{if}} guards on {{range}} in templates (lesson #7)
--   - parameterised queries with $1 (lesson #1)
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1. feed-ingester
-- ---------------------------------------------------------------------------
-- Receives: site_id, source_id, source_type, source_config
-- Routes by source_type to the appropriate fetch action, then writes items.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-ingester',
             'Feed Ingester',
             'Fetches content from a single source (RSS, news search, LLM news, scrape) and writes to content_feed_items',
             'code-driven',
             'executor',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id", "source_id", "source_type", "source_config"]}'::jsonb,
             '{"produces": ["write_result", "timestamp_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "route_by_type",
                     "steps": {
                         "route_by_type": {
                             "action": "conditional_route",
                             "config": {
                                 "condition_field": "source_type",
                                 "routes": {
                                     "rss": "fetch_rss",
                                     "api_news": "fetch_llm_news",
                                     "news_search": "search_news",
                                     "scrape": "scrape_source",
                                     "default": "fetch_rss"
                                 }
                             },
                             "description": "Route to fetch action based on source type"
                         },

                         "fetch_rss": {
                             "action": "fetch_rss",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch and parse RSS/Atom feed"
                         },

                         "fetch_llm_news": {
                             "action": "fetch_llm_news",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch news via Grok/xAI API"
                         },

                         "search_news": {
                             "action": "fetch_news_search",
                             "config": {
                                 "source_config": "input_data.source_config"
                             },
                             "output_field": "search_results",
                             "next_step": "normalize_search",
                             "description": "Search for news via web search adapter"
                         },

                         "normalize_search": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "search",
                                 "results_field": "search_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize search results to feed items"
                         },

                         "scrape_source": {
                             "action": "fetch_scrape",
                             "config": {
                                 "source_config": "input_data.source_config"
                             },
                             "output_field": "scrape_results",
                             "next_step": "normalize_scrape",
                             "description": "Scrape target news page"
                         },

                         "normalize_scrape": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "scrape",
                                 "results_field": "scrape_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize scrape results to feed items"
                         },

                         "write_items": {
                             "action": "write_feed_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "source_id": "input_data.source_id",
                                 "items_field": "fetched_items.items",
                                 "source_type": "input_data.source_type"
                             },
                             "output_field": "write_result",
                             "next_step": "update_timestamps",
                             "description": "Write normalised items to content_feed_items"
                         },

                         "update_timestamps": {
                             "action": "update_source_timestamps",
                             "config": {
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "timestamp_result",
                             "next_step": "complete",
                             "description": "Update source fetch timestamps"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 2. content-feed-orchestrator
-- ---------------------------------------------------------------------------
-- Workflow: dispatch_sources → render_news_json → commit_news → complete
-- dispatch_sources is fire-and-forget (spawns ingesters async).
-- render_news_json produces /data/latest-news.json content.
-- commit_news sends it to the git adapter for commit + deploy.
-- Uses items from PREVIOUS ingestion cycles (current cycle's items arrive async).

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'content-feed-orchestrator',
             'Content Feed Orchestrator',
             'Per-site orchestrator: dispatches feed-ingesters for due sources, produces latest-news JSON, commits to git',
             'code-driven',
             'coordinator',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"]}'::jsonb,
             '{"produces": ["dispatch_result", "news_render_result", "news_commit_result"]}'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "dispatch_sources",
                     "steps": {
                         "dispatch_sources": {
                             "action": "dispatch_feed_sources",
                             "config": {
                                 "site_id": "input_data.site_id"
                             },
                             "output_field": "dispatch_result",
                             "next_step": "render_news_json",
                             "description": "Load due sources and spawn feed-ingester per source"
                         },

                         "render_news_json": {
                             "action": "render_news_section",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "page_name": "index",
                                 "max_items": 6,
                                 "max_age_hours": 72
                             },
                             "output_field": "news_render_result",
                             "next_step": "commit_news",
                             "description": "Produce latest-news JSON from current relevant items"
                         },

                         "commit_news": {
                             "action": "git_commit",
                             "config": {
                                 "files_field": "news_render_result.files",
                                 "domain_field": "news_render_result.domain",
                                 "commit_message": "Update latest news feed"
                             },
                             "output_field": "news_commit_result",
                             "next_step": "complete",
                             "description": "Commit latest-news.json to git repo for S3 deploy"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 3. feed-triage — scores ingested items against site spec
-- ---------------------------------------------------------------------------
-- Workflow: load_items → check_has_items → read_site_spec → score_relevance → apply_scores → complete
-- Reads site spec before scoring so the LLM has identity, classification,
-- content_direction, and legal_rules context for relevance judgement.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-triage',
             'Feed Triage',
             'Scores ingested feed items for relevance to site spec — identity, vertical, values, legal rules. Pre-display gate.',
             'code-driven',
             'analyst',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"], "optional": ["relevance_threshold"]}'::jsonb,
             '{"produces": ["triage_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "ai_service": {
                     "model": "claude-sonnet-4-6",
                     "provider": "anthropic",
                     "max_tokens": 4000,
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "workflow": {
                     "start_step": "load_items",
                     "steps": {
                         "load_items": {
                             "action": "load_feed_items_for_triage",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "max_items": 50
                             },
                             "output_field": "pending_items",
                             "next_step": "check_has_items",
                             "description": "Load unscored ingested items with source metadata"
                         },

                         "check_has_items": {
                             "action": "evaluate_condition",
                             "config": {
                                 "condition_field": "pending_items.count",
                                 "conditions": {
                                     "0": "complete"
                                 },
                                 "default": "read_site_spec"
                             },
                             "description": "Skip if no items to triage"
                         },

                         "read_site_spec": {
                             "action": "read_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id"
                             },
                             "output_field": "site_spec",
                             "next_step": "score_relevance",
                             "description": "Load all site spec aspects for relevance context"
                         },

                         "score_relevance": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-6",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "prompt_template": "You are a content relevance filter for the website {{.input_data.site_id}}.\n\nYour job: score each news item for how relevant and appropriate it is for this site and its users.\n\n## Site Context\n\n{{if .site_spec.data}}{{if .site_spec.data.identity}}### Identity\n{{range $k, $v := .site_spec.data.identity}}{{$k}}: {{$v}}\n{{end}}{{end}}\n\n{{if .site_spec.data.classification}}### Classification\n{{range $k, $v := .site_spec.data.classification}}{{$k}}: {{$v}}\n{{end}}{{end}}\n\n{{if .site_spec.data.content_direction}}### Content Direction\n{{range $k, $v := .site_spec.data.content_direction}}{{$k}}: {{$v}}\n{{end}}{{end}}\n\n{{if .site_spec.data.legal_rules}}### Legal Rules\nForbidden phrases: {{.site_spec.data.legal_rules.forbidden_phrases}}\n{{end}}{{end}}\n\n## Items to Score\n\n{{if .pending_items.items}}{{range .pending_items.items}}ID: {{.id}}\nTitle: {{.source_title}}\nSummary: {{.source_summary}}\nSource: {{.source_name}} ({{.source_type}})\n---\n{{end}}{{end}}\n\n## Scoring Guide\n\n- 80-100: Directly relevant — covers this site''s industry, audience would want to read this\n- 50-79: Tangentially relevant — adjacent topic, broader market context, general industry news\n- 20-49: Weak relevance — same sector but wrong geography, audience, or focus\n- 0-19: Not relevant — wrong industry, spam, clickbait\n\nAlso check each item against the site''s values and legal rules. If an item conflicts with these (inappropriate tone, forbidden topics, legal risk), set \"flagged\": true regardless of relevance score.\n\nReturn ONLY a JSON array, no other text:\n[{\"id\": \"uuid\", \"score\": 0-100, \"reason\": \"one sentence\", \"topics\": [\"tag1\", \"tag2\"], \"flagged\": false}]",
                                 "response_format": "json"
                             },
                             "output_field": "scores",
                             "next_step": "apply_scores",
                             "description": "LLM scores items for relevance against site spec"
                         },

                         "apply_scores": {
                             "action": "apply_feed_scores",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "scores_field": "scores"
                             },
                             "output_field": "triage_result",
                             "next_step": "complete",
                             "description": "Update items with scores and status"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();

---
-- Add input_fields to score_relevance step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,score_relevance,config,input_fields}',
        '["input_data", "pending_items", "site_spec"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'feed-triage' AND deleted_at IS NULL;

-- Verify
SELECT default_config->'workflow'->'steps'->'score_relevance'->'config'->'input_fields' as input_fields
FROM agent_definitions
WHERE type = 'feed-triage' AND deleted_at IS NULL;

---

-- update
-- ============================================================================
-- 026b: Agent definitions for news feed pipeline (corrected)
--
-- Fixes applied per dev guide (001e) and contracts (003e):
--   - model alias: claude-sonnet-4-6 (not full version string)
--   - api_key_env_var in every ai_service config
--   - agent_category from allowed set (check_ad_category constraint)
--   - status set to 'experimental'
--   - input_contract / output_contract defined
--   - ON CONFLICT (type, version) matches actual unique constraint
--   - {{if}} guards on {{range}} in templates (lesson #7)
--   - parameterised queries with $1 (lesson #1)
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1. feed-ingester
-- ---------------------------------------------------------------------------
-- Receives: site_id, source_id, source_type, source_config
-- Routes by source_type to the appropriate fetch action, then writes items.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-ingester',
             'Feed Ingester',
             'Fetches content from a single source (RSS, news search, LLM news, scrape) and writes to content_feed_items',
             'code-driven',
             'executor',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id", "source_id", "source_type", "source_config"]}'::jsonb,
             '{"produces": ["write_result", "timestamp_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "route_by_type",
                     "steps": {
                         "route_by_type": {
                             "action": "conditional_route",
                             "config": {
                                 "condition_field": "source_type",
                                 "routes": {
                                     "rss": "fetch_rss",
                                     "api_news": "fetch_llm_news",
                                     "news_search": "search_news",
                                     "scrape": "scrape_source",
                                     "default": "fetch_rss"
                                 }
                             },
                             "description": "Route to fetch action based on source type"
                         },

                         "fetch_rss": {
                             "action": "fetch_rss",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch and parse RSS/Atom feed"
                         },

                         "fetch_llm_news": {
                             "action": "fetch_llm_news",
                             "config": {
                                 "source_config": "input_data.source_config",
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Fetch news via Grok/xAI API"
                         },

                         "search_news": {
                             "action": "fetch_news_search",
                             "config": {
                                 "source_config": "input_data.source_config"
                             },
                             "output_field": "search_results",
                             "next_step": "normalize_search",
                             "description": "Search for news via web search adapter"
                         },

                         "normalize_search": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "search",
                                 "results_field": "search_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize search results to feed items"
                         },

                         "scrape_source": {
                             "action": "fetch_scrape",
                             "config": {
                                 "source_config": "input_data.source_config"
                             },
                             "output_field": "scrape_results",
                             "next_step": "normalize_scrape",
                             "description": "Scrape target news page"
                         },

                         "normalize_scrape": {
                             "action": "normalize_to_feed_items",
                             "config": {
                                 "source_format": "scrape",
                                 "results_field": "scrape_results"
                             },
                             "output_field": "fetched_items",
                             "next_step": "write_items",
                             "description": "Normalize scrape results to feed items"
                         },

                         "write_items": {
                             "action": "write_feed_items",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "source_id": "input_data.source_id",
                                 "items_field": "fetched_items.items",
                                 "source_type": "input_data.source_type"
                             },
                             "output_field": "write_result",
                             "next_step": "update_timestamps",
                             "description": "Write normalised items to content_feed_items"
                         },

                         "update_timestamps": {
                             "action": "update_source_timestamps",
                             "config": {
                                 "source_id": "input_data.source_id"
                             },
                             "output_field": "timestamp_result",
                             "next_step": "complete",
                             "description": "Update source fetch timestamps"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 2. content-feed-orchestrator
-- ---------------------------------------------------------------------------
-- Workflow: dispatch_sources → render_news_json → commit_news → complete
-- dispatch_sources is fire-and-forget (spawns ingesters async).
-- render_news_json produces /data/latest-news.json content.
-- commit_news sends it to the git adapter for commit + deploy.
-- Uses items from PREVIOUS ingestion cycles (current cycle's items arrive async).

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'content-feed-orchestrator',
             'Content Feed Orchestrator',
             'Per-site orchestrator: dispatches feed-ingesters for due sources, produces latest-news JSON, commits to git',
             'code-driven',
             'coordinator',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"]}'::jsonb,
             '{"produces": ["dispatch_result", "news_render_result", "news_commit_result"]}'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "dispatch_sources",
                     "steps": {
                         "dispatch_sources": {
                             "action": "dispatch_feed_sources",
                             "config": {
                                 "site_id": "input_data.site_id"
                             },
                             "output_field": "dispatch_result",
                             "next_step": "render_news_json",
                             "description": "Load due sources and spawn feed-ingester per source"
                         },

                         "render_news_json": {
                             "action": "render_news_section",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "page_name": "index",
                                 "max_items": 6,
                                 "max_age_hours": 72
                             },
                             "output_field": "news_render_result",
                             "next_step": "commit_news",
                             "description": "Produce latest-news JSON from current relevant items"
                         },

                         "commit_news": {
                             "action": "git_commit",
                             "config": {
                                 "files_field": "news_render_result.files",
                                 "domain_field": "news_render_result.domain",
                                 "commit_message": "Update latest news feed"
                             },
                             "output_field": "news_commit_result",
                             "next_step": "complete",
                             "description": "Commit latest-news.json to git repo for S3 deploy"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();


-- ---------------------------------------------------------------------------
-- 3. feed-triage — scores ingested items against site spec
-- ---------------------------------------------------------------------------
-- Workflow: load_items → check_has_items → read_site_spec → score_relevance → apply_scores → complete
-- Reads site spec before scoring so the LLM has identity, classification,
-- content_direction, and legal_rules context for relevance judgement.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    agent_category, status,
    image_repository, image_tag,
    input_contract, output_contract,
    default_config, is_active
) VALUES (
             'feed-triage',
             'Feed Triage',
             'Scores ingested feed items for relevance to site spec — identity, vertical, values, legal rules. Pre-display gate.',
             'code-driven',
             'analyst',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'v1.0.157',
             '{"required": ["site_id"], "optional": ["relevance_threshold"]}'::jsonb,
             '{"produces": ["triage_result"]}'::jsonb,
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "ai_service": {
                     "model": "claude-sonnet-4-6",
                     "provider": "anthropic",
                     "max_tokens": 4000,
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "workflow": {
                     "start_step": "load_items",
                     "steps": {
                         "load_items": {
                             "action": "load_feed_items_for_triage",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "max_items": 50
                             },
                             "output_field": "pending_items",
                             "next_step": "check_has_items",
                             "description": "Load unscored ingested items with source metadata"
                         },

                         "check_has_items": {
                             "action": "evaluate_condition",
                             "config": {
                                 "condition_field": "pending_items.count",
                                 "conditions": {
                                     "0": "complete"
                                 },
                                 "default": "read_site_spec"
                             },
                             "description": "Skip if no items to triage"
                         },

                         "read_site_spec": {
                             "action": "read_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id"
                             },
                             "output_field": "site_spec",
                             "next_step": "score_relevance",
                             "description": "Load all site spec aspects for relevance context"
                         },

                         "score_relevance": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-6",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["input_data", "pending_items", "site_spec"],
                                 "prompt_template": "You are a content relevance filter for the website {{.input_data.site_id}}.\n\nYour job: score each news item for how relevant and appropriate it is for this site and its users.\n\n## Site Context\n\n{{if .site_spec.data}}{{if .site_spec.data.identity}}### Identity\n{{range $k, $v := .site_spec.data.identity}}{{$k}}: {{$v}}\n{{end}}{{end}}\n\n{{if .site_spec.data.classification}}### Classification\n{{range $k, $v := .site_spec.data.classification}}{{$k}}: {{$v}}\n{{end}}{{end}}\n\n{{if .site_spec.data.content_direction}}### Content Direction\n{{range $k, $v := .site_spec.data.content_direction}}{{$k}}: {{$v}}\n{{end}}{{end}}\n\n{{if .site_spec.data.legal_rules}}### Legal Rules\nForbidden phrases: {{.site_spec.data.legal_rules.forbidden_phrases}}\n{{end}}{{end}}\n\n## Items to Score\n\n{{if .pending_items.items}}{{range .pending_items.items}}ID: {{.id}}\nTitle: {{.source_title}}\nSummary: {{.source_summary}}\nSource: {{.source_name}} ({{.source_type}})\n---\n{{end}}{{end}}\n\n## Scoring Guide\n\n- 80-100: Directly relevant — covers this site''s industry, audience would want to read this\n- 50-79: Tangentially relevant — adjacent topic, broader market context, general industry news\n- 20-49: Weak relevance — same sector but wrong geography, audience, or focus\n- 0-19: Not relevant — wrong industry, spam, clickbait\n\nAlso check each item against the site''s values and legal rules. If an item conflicts with these (inappropriate tone, forbidden topics, legal risk), set \"flagged\": true regardless of relevance score.\n\nReturn ONLY a JSON array, no other text:\n[{\"id\": \"uuid\", \"score\": 0-100, \"reason\": \"one sentence\", \"topics\": [\"tag1\", \"tag2\"], \"flagged\": false}]",
                                 "response_format": "json"
                             },
                             "output_field": "scores",
                             "next_step": "apply_scores",
                             "description": "LLM scores items for relevance against site spec"
                         },

                         "apply_scores": {
                             "action": "apply_feed_scores",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "scores_field": "scores"
                             },
                             "output_field": "triage_result",
                             "next_step": "complete",
                             "description": "Update items with scores and status"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Done"
                         }
                     }
                 }
             }'::jsonb,
             true
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    agent_category = EXCLUDED.agent_category,
    status = EXCLUDED.status,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    updated_at = NOW();

---

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,score_relevance,config,prompt_template}',
        to_jsonb('You are a content relevance and credibility filter for the website {{.input_data.site_id}}.

Your job: score each news item on TWO dimensions:
1. RELEVANCE — how relevant is this to the site and its users?
2. CREDIBILITY — is this verified news from a credible source, or unverified rumour/speculation?

## Site Context

{{if .site_spec.data}}{{if .site_spec.data.identity}}### Identity
{{range $k, $v := .site_spec.data.identity}}{{$k}}: {{$v}}
{{end}}{{end}}

{{if .site_spec.data.classification}}### Classification
{{range $k, $v := .site_spec.data.classification}}{{$k}}: {{$v}}
{{end}}{{end}}

{{if .site_spec.data.content_direction}}### Content Direction
{{range $k, $v := .site_spec.data.content_direction}}{{$k}}: {{$v}}
{{end}}{{end}}

{{if .site_spec.data.legal_rules}}### Legal Rules
Forbidden phrases: {{.site_spec.data.legal_rules.forbidden_phrases}}
{{end}}{{end}}

## Items to Score

{{if .pending_items.items}}{{range .pending_items.items}}ID: {{.id}}
Title: {{.source_title}}
Summary: {{.source_summary}}
Source: {{.source_name}} ({{.source_type}})
URL: {{.source_url}}
---
{{end}}{{end}}

## Scoring Guide

### Relevance (0-100)
- 80-100: Directly relevant — covers this site''s industry, audience would want to read this
- 50-79: Tangentially relevant — adjacent topic, broader market context
- 20-49: Weak relevance — same sector but wrong geography, audience, or focus
- 0-19: Not relevant — wrong industry, spam, clickbait

### Source Attribution
For each item, determine the provenance chain:
- "original_source": The publication or entity that ORIGINATED the information (e.g. "Reuters", "OPEC", "UK Government")
- "found_via": How/where we found it (e.g. "X/@reuters", "OilPrice RSS", "direct article")
- "source_tier": Classify the ORIGINAL source:
  * "tier1_news" — major wire services and newspapers (Reuters, AP, Bloomberg, FT, BBC, WSJ)
  * "tier1_official" — government bodies, regulators, company official statements (OPEC, FCA, SEC)
  * "tier2_trade" — established trade/industry publications (OilPrice, TradeWinds, Rigzone)
  * "tier2_analysis" — known analysts, research firms, consultancies
  * "tier3_social" — social media posts, tweets from non-journalists
  * "tier3_blog" — personal blogs, opinion pieces, unattributed commentary
  * "unknown" — cannot determine original source

### Credibility assessment
Based on the source tier and content:
- "high": tier1 sources with verifiable claims
- "medium": tier2 sources, or tier1 with speculative framing ("sources say", "reportedly")
- "low": tier3 sources, unverifiable claims, no clear original source

### Flagging rules
Set "flagged": true if ANY of these apply:
- Item conflicts with site values or legal rules
- Credibility is "low" AND no tier1/tier2 source corroborates
- Item has no real URL or the URL looks fabricated
- Item contains inflammatory language or unsubstantiated claims
- Item is a navigation link, category page, or non-article URL

Return ONLY a JSON array, no other text:
[{
  "id": "uuid",
  "score": 0-100,
  "credibility": "high|medium|low",
  "credibility_reason": "one sentence explaining credibility assessment",
  "source_attribution": {
    "original_source": "Reuters",
    "found_via": "OilPrice RSS",
    "source_tier": "tier1_news"
  },
  "topics": ["tag1", "tag2"],
  "flagged": false
}]'::text)
                     ),
    updated_at = NOW()
WHERE type = 'feed-triage' AND deleted_at IS NULL;

---
-- feed credibility

-- ============================================================================
-- 027l — Feed credibility + source attribution
-- ============================================================================
--
-- 1. Add credibility columns to content_feed_items
-- 2. Update feed-triage prompt to assess credibility + provenance
-- ============================================================================

-- Step 1: Schema changes
ALTER TABLE content_feed_items
    ADD COLUMN IF NOT EXISTS credibility text,
    ADD COLUMN IF NOT EXISTS credibility_reason text,
    ADD COLUMN IF NOT EXISTS source_attribution jsonb DEFAULT '{}'::jsonb;

COMMENT ON COLUMN content_feed_items.credibility IS 'high/medium/low — set by triage LLM';
COMMENT ON COLUMN content_feed_items.credibility_reason IS 'Why this credibility level was assigned';
COMMENT ON COLUMN content_feed_items.source_attribution IS 'Provenance chain: original_source, found_via, source_tier';

-- Index for filtering by credibility
CREATE INDEX IF NOT EXISTS idx_cfi_credibility
    ON content_feed_items (site_id, credibility)
    WHERE credibility IS NOT NULL;

-- Step 2: Backup triage agent
CREATE TABLE IF NOT EXISTS agent_def_feed_triage_backup_20260403 AS
SELECT * FROM agent_definitions
WHERE type = 'feed-triage' AND deleted_at IS NULL;

-- Step 3: Update triage prompt with credibility + source attribution
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,score_relevance,config,prompt_template}',
        to_jsonb('You are a content relevance and credibility filter for the website {{.input_data.site_id}}.

Your job: score each news item on TWO dimensions:
1. RELEVANCE — how relevant is this to the site and its users?
2. CREDIBILITY — is this verified news from a credible source, or unverified rumour/speculation?

## Site Context

{{if .site_spec.data}}{{if .site_spec.data.identity}}### Identity
{{range $k, $v := .site_spec.data.identity}}{{$k}}: {{$v}}
{{end}}{{end}}

{{if .site_spec.data.classification}}### Classification
{{range $k, $v := .site_spec.data.classification}}{{$k}}: {{$v}}
{{end}}{{end}}

{{if .site_spec.data.content_direction}}### Content Direction
{{range $k, $v := .site_spec.data.content_direction}}{{$k}}: {{$v}}
{{end}}{{end}}

{{if .site_spec.data.legal_rules}}### Legal Rules
Forbidden phrases: {{.site_spec.data.legal_rules.forbidden_phrases}}
{{end}}{{end}}

## Items to Score

{{if .pending_items.items}}{{range .pending_items.items}}ID: {{.id}}
Title: {{.source_title}}
Summary: {{.source_summary}}
Source: {{.source_name}} ({{.source_type}})
URL: {{.source_url}}
---
{{end}}{{end}}

## Scoring Guide

### Relevance (0-100)
- 80-100: Directly relevant — covers this site''s industry, audience would want to read this
- 50-79: Tangentially relevant — adjacent topic, broader market context
- 20-49: Weak relevance — same sector but wrong geography, audience, or focus
- 0-19: Not relevant — wrong industry, spam, clickbait

### Source Attribution
For each item, determine the provenance chain:
- "original_source": The publication or entity that ORIGINATED the information (e.g. "Reuters", "OPEC", "UK Government")
- "found_via": How/where we found it (e.g. "X/@reuters", "OilPrice RSS", "direct article")
- "source_tier": Classify the ORIGINAL source:
  * "tier1_news" — major wire services and newspapers (Reuters, AP, Bloomberg, FT, BBC, WSJ, Guardian)
  * "tier1_official" — government bodies, regulators, company official statements (OPEC, FCA, SEC, Bank of England)
  * "tier2_trade" — established trade/industry publications (OilPrice, TradeWinds, Rigzone, industry journals)
  * "tier2_analysis" — known analysts, research firms, consultancies (McKinsey, Deloitte, named analysts)
  * "tier3_social" — social media posts, tweets from non-journalists, anonymous accounts
  * "tier3_blog" — personal blogs, opinion pieces, unattributed commentary
  * "unknown" — cannot determine original source

### Credibility assessment
Based on the source tier and content:
- "high": tier1 sources with verifiable claims, or tier2 sources reporting established facts
- "medium": tier2 sources with analysis/opinion, or tier1 with speculative framing ("sources say", "reportedly")
- "low": tier3 sources, unverifiable claims, no clear original source, anonymous speculation

### Flagging rules
Set "flagged": true if ANY of these apply:
- Item conflicts with site values or legal rules
- Credibility is "low" AND no tier1/tier2 source corroborates the claim
- Item has no real URL or the URL looks fabricated (e.g. placeholder, category page, navigation link)
- Item contains inflammatory language or unsubstantiated claims
- Item is a navigation link, category page, or non-article URL (e.g. ends in just a category path like /Energy/ or /News/)
- Item is older than 48 hours (stale for news display)

Return ONLY a JSON array, no other text:
[{
  "id": "uuid",
  "score": 0-100,
  "credibility": "high|medium|low",
  "credibility_reason": "one sentence explaining credibility assessment",
  "source_attribution": {
    "original_source": "Reuters",
    "found_via": "OilPrice RSS",
    "source_tier": "tier1_news"
  },
  "reason": "one sentence on relevance",
  "topics": ["tag1", "tag2"],
  "flagged": false
}]'::text)
                     ),
    updated_at = NOW()
WHERE type = 'feed-triage' AND deleted_at IS NULL;

-- Verify
SELECT type,
       length(default_config->'workflow'->'steps'->'score_relevance'->'config'->>'prompt_template') as prompt_length
FROM agent_definitions
WHERE type = 'feed-triage' AND deleted_at IS NULL;

-- Expected: prompt_length around 2500-3000 characters
-- Change scores_field to point directly at scores.result
-- The code appends ".result" making it "scores.result.result" (fails, nil)
-- Then falls back to "scores.result" (finds the string, parses it)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,apply_scores,config,scores_field}',
        '"scores.result"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'feed-triage' AND deleted_at IS NULL;

-- Verify
SELECT default_config->'workflow'->'steps'->'apply_scores'->'config'->>'scores_field' as scores_field
FROM agent_definitions WHERE type = 'feed-triage' AND deleted_at IS NULL;

-- Trigger a feed refresh to test
UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = 'content-feed-refresh';

-- 1. Increase max_tokens (20 items × ~200 tokens each = ~4000, need headroom)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,score_relevance,config,ai_service,max_tokens}',
        '8192'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'feed-triage' AND deleted_at IS NULL;

-- 2. Also limit batch size in load_items to prevent massive prompts
-- Check current config
SELECT default_config->'workflow'->'steps'->'load_items'->'config' as load_config
FROM agent_definitions WHERE type = 'feed-triage' AND deleted_at IS NULL;



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

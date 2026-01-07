-- ===========================================================================
-- RESEARCH AGENT
-- File: 046_research_agent.sql
-- ===========================================================================
-- Researches topics via web search, extracts relevant information,
-- synthesizes findings with source citations, stores in research_results.
-- ===========================================================================

BEGIN;

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    is_active,
    status,
    version,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    input_contract,
    output_contract,
    default_config
) VALUES (
             'research-agent',
             'Research Agent',
             'Researches topics via web search, extracts relevant quotes, synthesizes findings with full source attribution. Stores results in research_results table for citation.',
             'specialist',
             true,
             'active',
             1,
             '["web-search", "research", "citation", "synthesis"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.575',
             '{
                 "requests": {"cpu": "100m", "memory": "256Mi"},
                 "limits": {"cpu": "500m", "memory": "512Mi"}
             }'::jsonb,
             '{
                 "error": "system.errors.{type}",
                 "process": "system.agent.{type}.process",
                 "response": "system.responses.{type}"
             }'::jsonb,
             '{
                 "port": 8080,
                 "liveness_path": "/health",
                 "readiness_path": "/ready",
                 "initial_delay_seconds": 15
             }'::jsonb,
             -- Input contract
             '{
                 "expects": {
                     "current_section": "object with topic or research_query field",
                     "reviewed_brief": "object with industry, company context",
                     "site_record": "object with site_id for storing results"
                 },
                 "required": ["current_section"]
             }'::jsonb,
             -- Output contract
             '{
                 "produces": {
                     "id": "uuid - research_results record ID",
                     "query": "string - the search query used",
                     "summary": "string - synthesized findings with citations",
                     "sources": "array of {url, title, domain, quotes[], accessed_at}",
                     "source_count": "number"
                 }
             }'::jsonb,
             -- Workflow
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 90,
                 "workflow": {
                     "start_step": "extract_topic",
                     "steps": {
                         "extract_topic": {
                             "action": "extract_fields",
                             "description": "Extract topic from section data",
                             "config": {
                                 "fields": {
                                     "topic": ["current_section.topic", "current_section.research_query", "current_section.name"],
                                     "industry": ["reviewed_brief.industry", "reviewed_brief.category"],
                                     "company": ["reviewed_brief.company_name"]
                                 }
                             },
                             "next_step": "build_search_query",
                             "output_field": "extracted"
                         },

                         "build_search_query": {
                             "action": "execute_llm_prompt",
                             "description": "Build effective search query from context",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-haiku-4-5-20251001",
                                     "max_tokens": 150,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["extracted"],
                                 "prompt_template": "Create a focused web search query to research:\n\nTopic: {{extracted.topic}}\nIndustry: {{extracted.industry}}\nCompany: {{extracted.company}}\n\nReturn ONLY the search query string, 3-8 words, no quotes or operators.\nFocus on finding authoritative, recent information about this topic in this industry."
                             },
                             "next_step": "search_web",
                             "output_field": "search_query"
                         },

                         "search_web": {
                             "action": "web_search",
                             "description": "Search the web for relevant sources",
                             "config": {
                                 "query_from": "search_query.result",
                                 "max_results": 8,
                                 "exclude_domains": ["pinterest.com", "facebook.com", "twitter.com", "instagram.com"]
                             },
                             "next_step": "filter_results",
                             "output_field": "search_results"
                         },

                         "filter_results": {
                             "action": "filter_search_results",
                             "description": "Filter to most relevant, authoritative sources",
                             "config": {
                                 "results_from": "search_results",
                                 "max_sources": 5,
                                 "prefer_domains": [".gov", ".edu", ".org", "reuters.com", "bbc.com", "forbes.com", "hbr.org", "mckinsey.com"]
                             },
                             "next_step": "fetch_and_extract",
                             "output_field": "filtered_results"
                         },

                         "fetch_and_extract": {
                             "action": "loop",
                             "description": "Fetch and extract content from each source",
                             "config": {
                                 "loop_var": "source",
                                 "iterate_over": "filtered_results",
                                 "max_iterations": 5,
                                 "substeps": {
                                     "fetch_page": {
                                         "action": "web_fetch",
                                         "description": "Fetch page content",
                                         "config": {
                                             "url_from": "source.url",
                                             "timeout_seconds": 10,
                                             "extract_text": true,
                                             "max_content_length": 10000
                                         },
                                         "next_step": "extract_quotes",
                                         "output_field": "page_content"
                                     },

                                     "extract_quotes": {
                                         "action": "execute_llm_prompt",
                                         "description": "Extract relevant quotes from page",
                                         "config": {
                                             "ai_service": {
                                                 "provider": "anthropic",
                                                 "model": "claude-haiku-4-5-20251001",
                                                 "max_tokens": 500,
                                                 "api_key_env_var": "ANTHROPIC_API_KEY"
                                             },
                                             "input_fields": ["page_content", "extracted"],
                                             "output_format": "json",
                                             "prompt_template": "Extract 2-4 relevant quotes from this page about: {{extracted.topic}}\n\nPage content (truncated):\n{{page_content.text}}\n\nReturn JSON:\n{\n  \"quotes\": [\"exact quote 1\", \"exact quote 2\"],\n  \"relevance\": 0.0-1.0,\n  \"key_facts\": [\"fact 1\", \"fact 2\"]\n}\n\nOnly include quotes directly relevant to the topic. Rate relevance 0-1."
                                         },
                                         "output_field": "extracted_data"
                                     }
                                 }
                             },
                             "next_step": "synthesize",
                             "output_field": "fetched_sources"
                         },

                         "synthesize": {
                             "action": "execute_llm_prompt",
                             "description": "Synthesize findings with citations",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5-20250514",
                                     "max_tokens": 1000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["extracted", "fetched_sources"],
                                 "output_format": "json",
                                 "prompt_template": "Synthesize research findings about: {{extracted.topic}}\n\n## Sources\n{{#each fetched_sources}}\n[{{@index}}] {{this.source.title}} ({{this.source.domain}})\nQuotes: {{this.extracted_data.quotes}}\nKey facts: {{this.extracted_data.key_facts}}\nRelevance: {{this.extracted_data.relevance}}\n{{/each}}\n\n## Task\nWrite a synthesis of the key findings. Cite sources by number [0], [1], etc.\n\nReturn JSON:\n{\n  \"summary\": \"2-3 paragraph synthesis with [citations]\",\n  \"key_points\": [\"point 1 [0]\", \"point 2 [1,2]\"],\n  \"confidence\": 0.0-1.0\n}\n\nRules:\n- Every claim must have a citation\n- Be factual and objective\n- Note any conflicting information between sources"
                             },
                             "next_step": "store_research",
                             "output_field": "synthesis"
                         },

                         "store_research": {
                             "action": "insert_research_result",
                             "description": "Store research in database for future reference",
                             "config": {
                                 "table": "research_results",
                                 "fields": {
                                     "site_id": "site_record.site_id",
                                     "query": "search_query.result",
                                     "topic": "extracted.topic",
                                     "summary": "synthesis.summary",
                                     "findings": "synthesis",
                                     "sources": "fetched_sources"
                                 }
                             },
                             "next_step": "complete",
                             "output_field": "stored_research"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output": {
                                     "id": "stored_research.id",
                                     "query": "search_query.result",
                                     "summary": "synthesis.summary",
                                     "sources": "fetched_sources",
                                     "source_count": "fetched_sources.length",
                                     "key_points": "synthesis.key_points"
                                 }
                             }
                         }
                     }
                 }
             }'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              version = EXCLUDED.version,
                              default_config = EXCLUDED.default_config,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = now();

COMMIT;

-------------

-- FIX: research-agent template syntax
-- ====================================
-- Issues:
-- 1. Missing dot prefixes on variable references
-- 2. Handlebars syntax ({{#each}}, {{this.}}, {{@index}}) needs Go template syntax

-- Fix build_search_query step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_search_query,config,prompt_template}',
        '"Create a focused web search query to research:\n\nTopic: {{.extracted.topic}}\nIndustry: {{.extracted.industry}}\nCompany: {{.extracted.company}}\n\nReturn ONLY the search query string, 3-8 words, no quotes or operators.\nFocus on finding authoritative, recent information about this topic in this industry."'
                     )
WHERE type = 'research-agent';

-- Fix synthesize step (convert Handlebars to Go template)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,synthesize,config,prompt_template}',
        '"Synthesize research findings about: {{.extracted.topic}}\n\n## Sources\n{{range $index, $source := .fetched_sources}}\n[{{$index}}] {{$source.source.title}} ({{$source.source.domain}})\nQuotes: {{$source.extracted_data.quotes}}\nKey facts: {{$source.extracted_data.key_facts}}\nRelevance: {{$source.extracted_data.relevance}}\n{{end}}\n\n## Task\nWrite a synthesis of the key findings. Cite sources by number [0], [1], etc.\n\nReturn JSON:\n{\n  \"summary\": \"2-3 paragraph synthesis with [citations]\",\n  \"key_points\": [\"point 1 [0]\", \"point 2 [1,2]\"],\n  \"confidence\": 0.0-1.0\n}\n\nRules:\n- Every claim must have a citation\n- Be factual and objective\n- Note any conflicting information between sources"'
                     )
WHERE type = 'research-agent';

-- Fix extract_quotes step (inside fetch_and_extract loop substeps)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,fetch_and_extract,config,substeps,extract_quotes,config,prompt_template}',
        '"Extract 2-4 relevant quotes from this page about: {{.extracted.topic}}\n\nPage content (truncated):\n{{.page_content.text}}\n\nReturn JSON:\n{\n  \"quotes\": [\"exact quote 1\", \"exact quote 2\"],\n  \"relevance\": 0.0-1.0,\n  \"key_facts\": [\"fact 1\", \"fact 2\"]\n}\n\nOnly include quotes directly relevant to the topic. Rate relevance 0-1."'
                     )
WHERE type = 'research-agent';


-- ============================================
-- 4. research-agent: synthesize step
-- (build_search_query and extract_quotes use haiku which is fine)
-- ============================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,synthesize,config,ai_service,model}',
        '"claude-sonnet-4-5-20250929"'
                     ),
    updated_at = NOW()
WHERE type = 'research-agent';

--
-- Issue 37: Fix research-agent extract_topic paths to use input_data prefix
--
-- Problem: extract_topic step uses paths like "current_section.name" but the data
-- arrives from call_agent at "input_data.current_section.name"
--
-- This causes extracted topic to be empty, leading to empty search queries

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,extract_topic,config,fields}',
        '{
            "topic": ["input_data.current_section.topic", "input_data.current_section.research_query", "input_data.current_section.name", "input_data.current_section.function"],
            "company": ["input_data.reviewed_brief.company_name", "reviewed_brief.company_name"],
            "industry": ["input_data.reviewed_brief.industry", "input_data.reviewed_brief.category", "reviewed_brief.industry"]
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'research-agent';

-- Verify the update
SELECT type,
       default_config->'workflow'->'steps'->'extract_topic'->'config'->'fields' as extract_fields
FROM agent_definitions
WHERE type = 'research-agent';

--

-- shorten results_field path
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,filter_results,config}',
        '{
          "results_field": "search_results.results",
          "max_sources": 5,
          "prefer_domains": [".gov", ".edu", ".org", "reuters.com", "bbc.com", "bloomberg.com", "forbes.com", "hbr.org", "mckinsey.com", "arxiv.org", "www.afp.com/en"]
        }'::jsonb
                     )
WHERE type = 'research-agent';

change iterate_over  >  filtered_results path:

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,fetch_and_extract,config,iterate_over}',
        '"filtered_results.filtered_results"'
                     )
WHERE type = 'research-agent';

--
-- FILE: research_agent_workflow_update.sql
-- Updated research-agent workflow using batch scraping with verified paths
--
-- DATA FLOW SUMMARY:
-- ==================
--
-- Step 1: extract_topic
--   Output field: "extracted"
--   Stored at: collected_data["extracted"] = {topic, company, industry}
--
-- Step 2: build_search_query
--   Input: {{.extracted.topic}}, {{.extracted.industry}}, {{.extracted.company}}
--   Output field: "search_query"
--   Stored at: collected_data["search_query"] = {result: "query string", type: "text"}
--
-- Step 3: search_web
--   Config: query_from: "search_query.result"
--   Path: collected_data["search_query"]["result"] → "query string"
--   Output field: "search_results"
--   Adapter returns: {success, results, query, total, provider}
--   Stored at: collected_data["search_results"]
--   Results array: collected_data["search_results"]["results"]
--
-- Step 4: prepare_urls
--   Input: Finds results at search_results.results
--   Output field: "prepared_urls"
--   Stored at: collected_data["prepared_urls"] = {urls_to_scrape, scrape_count, snippet_context, ...}
--   URLs array: collected_data["prepared_urls"]["urls_to_scrape"]
--
-- Step 5: scrape_pages
--   Config: urls_field: "prepared_urls.urls_to_scrape"
--   Output field: "scrape_results"
--   Adapter returns: {success, results, success_count, error_count, total_count}
--   Stored at: collected_data["scrape_results"]
--   Results array: collected_data["scrape_results"]["results"]
--
-- Step 6: format_content
--   Input: scrape_results (finds results at scrape_results.results)
--         prepared_urls.snippet_context
--   Output field: "research_content"
--   Stored at: collected_data["research_content"] = {research_text, sources, source_count, content_quality}
--
-- Step 7: synthesize
--   Input: {{.extracted.topic}}, {{.research_content.research_text}}
--   Output field: "synthesis"
--   Returns: {result: {summary, key_points, confidence}, type: "json"}
--   Stored at: collected_data["synthesis"]
--   Access summary: collected_data["synthesis"]["result"]["summary"]
--
-- Step 8: store_research
--   Uses paths like: synthesis.result.summary, research_content.sources
--
-- Step 9: complete

UPDATE agent_definitions
SET default_config = '{
  "workflow": {
    "start_step": "extract_topic",
    "steps": {
      "extract_topic": {
        "action": "extract_fields",
        "description": "Extract research topic from section data",
        "config": {
          "fields": {
            "topic": ["current_section.topic", "current_section.research_query", "current_section.name"],
            "company": ["reviewed_brief.company_name"],
            "industry": ["reviewed_brief.industry", "reviewed_brief.category"]
          }
        },
        "next_step": "build_search_query",
        "output_field": "extracted"
      },

      "build_search_query": {
        "action": "execute_llm_prompt",
        "description": "Build effective search query from context",
        "config": {
          "ai_service": {
            "model": "claude-haiku-4-5-20251001",
            "provider": "anthropic",
            "max_tokens": 150,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "input_fields": ["extracted"],
          "prompt_template": "Create a focused web search query to research:\n\nTopic: {{.extracted.topic}}\nIndustry: {{.extracted.industry}}\nCompany: {{.extracted.company}}\n\nReturn ONLY the search query string, 3-8 words, no quotes or operators.\nFocus on finding authoritative, recent information about this topic."
        },
        "next_step": "search_web",
        "output_field": "search_query"
      },

      "search_web": {
        "action": "web_search",
        "description": "Search the web for relevant sources",
        "config": {
          "query_from": "search_query.result",
          "num_results": 10
        },
        "next_step": "prepare_urls",
        "output_field": "search_results"
      },

      "prepare_urls": {
        "action": "prepare_urls",
        "description": "Extract top URLs for scraping and collect snippets",
        "config": {
          "max_scrapes": 3,
          "max_snippets": 5,
          "exclude_domains": ["pinterest.com", "facebook.com", "twitter.com", "instagram.com", "youtube.com", "reddit.com"],
          "prefer_domains": [".gov", ".edu", ".org", "reuters.com", "bbc.com", "forbes.com", "hbr.org", "mckinsey.com"]
        },
        "next_step": "scrape_pages",
        "output_field": "prepared_urls"
      },

      "scrape_pages": {
        "action": "batch_webscrape",
        "description": "Scrape top result pages for detailed content",
        "config": {
          "urls_field": "prepared_urls.urls_to_scrape",
          "scrape_config": {
            "only_main_content": true,
            "capture_screenshot": false
          }
        },
        "next_step": "format_content",
        "output_field": "scrape_results",
        "timeout_seconds": 60
      },

      "format_content": {
        "action": "format_research_content",
        "description": "Format scraped content and snippets for synthesis",
        "config": {
          "scrape_field": "scrape_results",
          "snippets_field": "prepared_urls.snippet_context",
          "max_content_per_source": 6000
        },
        "next_step": "synthesize",
        "output_field": "research_content"
      },

      "synthesize": {
        "action": "execute_llm_prompt",
        "description": "Synthesize findings into comprehensive summary with citations",
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-5-20250514",
            "provider": "anthropic",
            "max_tokens": 1500,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "input_fields": ["extracted", "research_content"],
          "output_format": "json",
          "prompt_template": "Synthesize research findings about: {{.extracted.topic}}\n\n{{.research_content.research_text}}\n\n## Task\nWrite a comprehensive synthesis of the findings. Cite sources by index [0], [1], [2] where applicable.\n\nReturn JSON:\n{\n  \"summary\": \"3-4 paragraph synthesis with [citations] where relevant\",\n  \"key_points\": [\"key point with [citation]\", \"another point [citation]\"],\n  \"recommendations\": [\"actionable recommendation based on findings\"],\n  \"confidence\": 0.0-1.0\n}\n\nRules:\n- Cite sources when making specific claims\n- Be factual and objective\n- Note any conflicting information between sources\n- Rate confidence based on source quality and agreement"
        },
        "next_step": "store_research",
        "output_field": "synthesis"
      },

      "store_research": {
        "action": "insert_research_result",
        "description": "Store research in database for future reference",
        "config": {
          "table": "research_results",
          "fields": {
            "topic": "extracted.topic",
            "query": "search_query.result",
            "site_id": "site_record.site_id",
            "sources": "research_content.sources",
            "summary": "synthesis.result.summary",
            "findings": "synthesis.result"
          }
        },
        "next_step": "complete",
        "output_field": "stored_research"
      },

      "complete": {
        "action": "complete_workflow",
        "description": "Complete research workflow and return results",
        "config": {
          "output": {
            "id": "stored_research.id",
            "query": "search_query.result",
            "summary": "synthesis.result.summary",
            "key_points": "synthesis.result.key_points",
            "recommendations": "synthesis.result.recommendations",
            "confidence": "synthesis.result.confidence",
            "source_count": "research_content.source_count",
            "content_quality": "research_content.content_quality"
          }
        }
      }
    }
  },
  "processing_mode": "task",
  "timeout_seconds": 180
}'::jsonb
WHERE type = 'research-agent';

-- Verify the update
SELECT type,
       default_config->'workflow'->'start_step' as start_step,
       jsonb_object_keys(default_config->'workflow'->'steps') as step_names
FROM agent_definitions
WHERE type = 'research-agent';
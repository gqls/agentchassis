-- Insert website builder specific agent definitions
INSERT INTO agent_definitions (type, display_name, description, category, default_config, capabilities) VALUES
-- Domain Analyst
('domain-analyst', 'Domain Analyst', 'Analyzes domains and determines appropriate website type', 'data-driven',
 '{
     "model": "claude-3-5-sonnet-20241022",
     "temperature": 0.3,
     "processing_mode": "task",
     "workflow": {
         "start_step": "analyze_domain",
         "steps": {
             "analyze_domain": {
                 "action": "validate_input",
                 "next_step": "research_industry"
             },
             "research_industry": {
                 "action": "call_agent",
                 "agent_type": "web-search",
                 "topic": "system.adapter.web.search",
                 "config": {"provider": "firecrawl", "depth": "quick"},
                 "next_step": "categorize_business"
             },
             "categorize_business": {
                 "action": "execute_llm_prompt",
                 "config": {"prompt_template": "domain_categorization"},
                 "next_step": "complete",
                 "store_memory": true
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }', '["analysis", "categorization", "domain-research"]'::jsonb),

-- Site Architect
('site-architect', 'Site Architect', 'Plans website structure and navigation', 'data-driven',
 '{
     "model": "claude-3-5-sonnet-20241022",
     "temperature": 0.5,
     "processing_mode": "task",
     "workflow": {
         "start_step": "analyze_requirements",
         "steps": {
             "analyze_requirements": {
                 "action": "validate_input",
                 "next_step": "determine_pages"
             },
             "determine_pages": {
                 "action": "execute_llm_prompt",
                 "config": {"template": "site_structure"},
                 "next_step": "pause_for_human"
             },
             "pause_for_human": {
                 "action": "pause_for_human_input",
                 "config": {"message": "Review and approve site structure"},
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }', '["planning", "structure", "navigation"]'::jsonb),

-- Content Researcher (Using Perplexity for deep research)
('content-researcher', 'Content Researcher', 'Researches and gathers comprehensive information for website content', 'data-driven',
 '{
     "model": "claude-3-5-sonnet-20241022",
     "temperature": 0.4,
     "processing_mode": "task",
     "workflow": {
         "start_step": "identify_topics",
         "steps": {
             "identify_topics": {
                 "action": "validate_input",
                 "next_step": "deep_research"
             },
             "deep_research": {
                 "action": "call_agent",
                 "agent_type": "perplexity-research",
                 "topic": "system.adapter.perplexity.research",
                 "config": {
                     "mode": "comprehensive",
                     "include_sources": true,
                     "max_depth": 3
                 },
                 "next_step": "crawl_competitors"
             },
             "crawl_competitors": {
                 "action": "call_agent",
                 "agent_type": "firecrawl-scraper",
                 "topic": "system.adapter.firecrawl.scrape",
                 "config": {
                     "scrape_type": "competitor_analysis",
                     "extract_content": true,
                     "extract_meta": true
                 },
                 "next_step": "analyze_findings"
             },
             "analyze_findings": {
                 "action": "execute_llm_prompt",
                 "config": {"prompt_template": "synthesize_research"},
                 "next_step": "complete",
                 "store_memory": true
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }', '["research", "analysis", "fact-checking", "content-gathering", "competitor-analysis"]'::jsonb),

-- HTML Developer
('html-developer', 'HTML Developer', 'Generates HTML/CSS/JS code for websites', 'code-driven',
 '{
     "model": "claude-3-5-sonnet-20241022",
     "temperature": 0.2,
     "processing_mode": "task",
     "workflow": {
         "start_step": "receive_specs",
         "steps": {
             "receive_specs": {
                 "action": "validate_input",
                 "next_step": "generate_template"
             },
             "generate_template": {
                 "action": "execute_llm_prompt",
                 "config": {
                     "language": "html",
                     "framework": "vanilla",
                     "responsive": true
                 },
                 "next_step": "create_pages"
             },
             "create_pages": {
                 "action": "fan_out",
                 "config": {"per_page": true},
                 "next_step": "bundle_site"
             },
             "bundle_site": {
                 "action": "transform_data",
                 "config": {"transformation": "bundle_files"},
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }', '["html", "css", "javascript", "frontend"]'::jsonb),

-- Visual Designer (Updated to use Firecrawl for image search)
('visual-designer', 'Visual Designer', 'Handles images, logos, and visual assets', 'adapter',
 '{
     "processing_mode": "orchestrator",
     "workflow": {
         "start_step": "analyze_brand",
         "steps": {
             "analyze_brand": {
                 "action": "validate_input",
                 "next_step": "search_images"
             },
             "search_images": {
                 "action": "call_agent",
                 "agent_type": "firecrawl-scraper",
                 "topic": "system.adapter.firecrawl.scrape",
                 "config": {
                     "scrape_type": "image_search",
                     "sources": ["unsplash", "pexels"],
                     "extract_images": true
                 },
                 "next_step": "create_logo"
             },
             "create_logo": {
                 "action": "call_agent",
                 "agent_type": "image-generator",
                 "topic": "system.adapter.image.generate",
                 "next_step": "optimize_images"
             },
             "optimize_images": {
                 "action": "transform_data",
                 "config": {"transformation": "optimize_images"},
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }', '["design", "graphics", "branding", "image-processing"]'::jsonb),

-- Site Publisher
('site-publisher', 'Site Publisher', 'Publishes websites to storage buckets', 'adapter',
 '{
     "processing_mode": "orchestrator",
     "workflow": {
         "start_step": "collect_assets",
         "steps": {
             "collect_assets": {
                 "action": "validate_input",
                 "next_step": "organize_files"
             },
             "organize_files": {
                 "action": "transform_data",
                 "config": {"transformation": "organize_structure"},
                 "next_step": "upload_to_bucket"
             },
             "upload_to_bucket": {
                 "action": "s3_upload",
                 "config": {
                     "bucket": "${SITE_BUCKET}",
                     "public": true
                 },
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }', '["deployment", "hosting", "publishing", "s3"]'::jsonb),

-- Website Builder Orchestrator
('website-builder', 'Website Builder', 'Orchestrates complete website creation', 'code-driven',
 '{
     "processing_mode": "orchestrator",
     "workflow": {
         "start_step": "validate_request",
         "steps": {
             "validate_request": {
                 "action": "validate_input",
                 "next_step": "spawn_agents"
             },
             "spawn_agents": {
                 "action": "spawn_group",
                 "config": {"group_type": "website-builder"},
                 "next_step": "analyze_domain"
             },
             "analyze_domain": {
                 "action": "call_agent",
                 "agent_type": "domain-analyst",
                 "next_step": "architect_site"
             },
             "architect_site": {
                 "action": "call_agent",
                 "agent_type": "site-architect",
                 "next_step": "gather_content"
             },
             "gather_content": {
                 "action": "fan_out",
                 "sub_tasks": [
                     {"agent_type": "content-researcher", "step_name": "research"},
                     {"agent_type": "visual-designer", "step_name": "visuals"}
                 ],
                 "next_step": "develop_site"
             },
             "develop_site": {
                 "action": "call_agent",
                 "agent_type": "html-developer",
                 "next_step": "publish_site"
             },
             "publish_site": {
                 "action": "call_agent",
                 "agent_type": "site-publisher",
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }', '["orchestration", "website-creation", "project-management"]'::jsonb)
    ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              default_config = EXCLUDED.default_config,
                              capabilities = EXCLUDED.capabilities,
                              updated_at = NOW();


-- Insert research and scraping adapter agents
INSERT INTO agent_definitions (type, display_name, description, category, default_config, capabilities) VALUES
-- Perplexity Research Adapter
('perplexity-research', 'Perplexity Research', 'Deep research using Perplexity AI', 'adapter',
 '{
     "processing_mode": "adapter",
     "api_config": {
         "provider": "perplexity",
         "api_key_env_var": "PERPLEXITY_API_KEY",
         "model": "pplx-70b-online",
         "timeout": 30
     },
     "workflow": {
         "start_step": "call_perplexity",
         "steps": {
             "call_perplexity": {
                 "action": "http_request",
                 "config": {
                     "url": "https://api.perplexity.ai/chat/completions",
                     "method": "POST",
                     "headers": {
                         "Authorization": "Bearer ${PERPLEXITY_API_KEY}"
                     }
                 },
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }', '["research", "fact-checking", "real-time-data", "comprehensive-search"]'::jsonb),

-- Firecrawl Scraper Adapter
('firecrawl-scraper', 'Firecrawl Scraper', 'Web scraping and content extraction using Firecrawl', 'adapter',
 '{
     "processing_mode": "adapter",
     "api_config": {
         "provider": "firecrawl",
         "api_key_env_var": "FIRECRAWL_API_KEY",
         "base_url": "https://api.firecrawl.dev/v0",
         "timeout": 60
     },
     "workflow": {
         "start_step": "call_firecrawl",
         "steps": {
             "call_firecrawl": {
                 "action": "http_request",
                 "config": {
                     "url": "${base_url}/scrape",
                     "method": "POST",
                     "headers": {
                         "Authorization": "Bearer ${FIRECRAWL_API_KEY}"
                     }
                 },
                 "next_step": "complete"
             },
             "complete": {
                 "action": "complete_workflow"
             }
         }
     }
 }', '["web-scraping", "content-extraction", "competitor-analysis", "image-search"]'::jsonb)
    ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              default_config = EXCLUDED.default_config,
                              capabilities = EXCLUDED.capabilities,
                              updated_at = NOW();
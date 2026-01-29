-- Site Scraper Agent Definition
--
-- Scrapes live websites using existing webscrape adapter (Firecrawl)
-- Uses firecrawl_scrape action (requires patch_02_webscrape_url_field.go)
-- Transforms results into site_context format for webdesign-agent

/*INSERT INTO agent_definitions (
    type,
    name,
    description,
    role,
    default_config,
    can_delegate,
    tags,
    input_contract,
    output_contract
) VALUES (
             'site-scraper',
             'Site Scraper Agent',
             'Scrapes live websites to extract design context. Uses webscrape adapter with Firecrawl. Outputs site_context for webdesign-agent.',
             'specialist',
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 180,
                 "workflow": {
                     "start_step": "scrape_site",
                     "steps": {
                         "scrape_site": {
                             "action": "firecrawl_scrape",
                             "config": {
                                 "url_field": "input_data.url",
                                 "upload_results": false,
                                 "scrape_config": {
                                     "only_main_content": false,
                                     "capture_screenshot": false
                                 }
                             },
                             "description": "Scrape the website homepage",
                             "next_step": "analyze_design",
                             "output_field": "scrape_result"
                         },
                         "analyze_design": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 3000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["scrape_result", "input_data"],
                                 "output_format": "json",
                                 "prompt_template": "Analyze this scraped website data and extract design information.\n\nTarget URL: {{.input_data.url}}\n\nScrape Result (find html_content, markdown_content, title, description, links in this data):\n{{.scrape_result | tojson | truncate 15000}}\n\nExtract from the scraped content:\n1. Colors - hex/rgb colors from CSS or inline styles\n2. Fonts - font-family declarations\n3. Components - section types (hero, features, testimonials, cta, etc.)\n4. Navigation pages from links\n5. Company name and tagline\n\nReturn ONLY valid JSON:\n{\n  \"domain\": \"domain from url\",\n  \"company_name\": \"extracted name\",\n  \"tagline\": \"tagline or null\",\n  \"industry\": \"inferred industry\",\n  \"color_palette\": {\"primary\": \"#hex\", \"secondary\": \"#hex\", \"accent\": \"#hex\", \"background\": \"#fff\", \"text\": \"#333\"},\n  \"typography\": {\"font_family\": \"fonts\", \"heading_font\": \"heading font\", \"base_size\": \"16px\", \"line_height\": \"1.6\"},\n  \"all_component_functions\": [\"hero\", \"features\", \"cta\"],\n  \"pages\": [{\"title\": \"Home\", \"slug\": \"/\"}],\n  \"source\": \"scrape\",\n  \"design_notes\": \"brief notes\"\n}"
                             },
                             "description": "Analyze scraped content for design elements",
                             "next_step": "complete",
                             "output_field": "site_context"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["site_context", "scrape_result"]
                             }
                         }
                     }
                 }
             }',
             false,
             ARRAY['scraper', 'design', 'analysis', 'specialist'],
             '{"required": ["url"], "optional": ["provider"]}',
             '{"produces": {"site_context": "standardized site context for webdesign-agent"}}'
         )
    ON CONFLICT (type) DO UPDATE SET
    name = EXCLUDED.name,
                              description = EXCLUDED.description,
                              default_config = EXCLUDED.default_config,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();*/

-- fix for above

-- Site Scraper Agent Definition
--
-- Scrapes live websites using existing webscrape adapter (Firecrawl)
-- Uses firecrawl_scrape action (requires patch_02_webscrape_url_field.go)
-- Transforms results into site_context format for webdesign-agent

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    input_contract,
    output_contract
) VALUES (
             'site-scraper',
             'Site Scraper Agent',
             'Scrapes live websites to extract design context. Uses webscrape adapter with Firecrawl. Outputs site_context for webdesign-agent.',
             'specialist',
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 180,
                 "workflow": {
                     "start_step": "scrape_site",
                     "steps": {
                         "scrape_site": {
                             "action": "firecrawl_scrape",
                             "config": {
                                 "url_field": "input_data.url",
                                 "upload_results": false,
                                 "scrape_config": {
                                     "only_main_content": false,
                                     "capture_screenshot": false
                                 }
                             },
                             "description": "Scrape the website homepage",
                             "next_step": "analyze_design",
                             "output_field": "scrape_result"
                         },
                         "analyze_design": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 3000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["scrape_result", "input_data"],
                                 "output_format": "json",
                                 "prompt_template": "Analyze this scraped website data and extract design information.\n\nTarget URL: {{.input_data.url}}\n\nScrape Result (find html_content, markdown_content, title, description, links in this data):\n{{.scrape_result | tojson | truncate 15000}}\n\nExtract from the scraped content:\n1. Colors - hex/rgb colors from CSS or inline styles\n2. Fonts - font-family declarations\n3. Components - section types (hero, features, testimonials, cta, etc.)\n4. Navigation pages from links\n5. Company name and tagline\n\nReturn ONLY valid JSON:\n{\n  \"domain\": \"domain from url\",\n  \"company_name\": \"extracted name\",\n  \"tagline\": \"tagline or null\",\n  \"industry\": \"inferred industry\",\n  \"color_palette\": {\"primary\": \"#hex\", \"secondary\": \"#hex\", \"accent\": \"#hex\", \"background\": \"#fff\", \"text\": \"#333\"},\n  \"typography\": {\"font_family\": \"fonts\", \"heading_font\": \"heading font\", \"base_size\": \"16px\", \"line_height\": \"1.6\"},\n  \"all_component_functions\": [\"hero\", \"features\", \"cta\"],\n  \"pages\": [{\"title\": \"Home\", \"slug\": \"/\"}],\n  \"source\": \"scrape\",\n  \"design_notes\": \"brief notes\"\n}"
                             },
                             "description": "Analyze scraped content for design elements",
                             "next_step": "complete",
                             "output_field": "site_context"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["site_context", "scrape_result"]
                             }
                         }
                     }
                 }
             }',
             true,
             '["scraper", "design", "analysis", "specialist"]',
             '{"required": ["url"], "optional": ["provider"]}',
             '{"produces": {"site_context": "standardized site context for webdesign-agent"}}'
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();
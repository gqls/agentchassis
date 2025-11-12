INSERT INTO agent_definitions (
id,
type,
display_name,
description,
category,
default_config,
is_active,
capabilities,
image_repository,
image_tag
) VALUES (
gen_random_uuid(),
'website-capture-firecrawl',
'Website Capture Agent (Firecrawl)',
'Captures website content using Firecrawl API',
'data-extraction',
'{
"workflow": {
"start_step": "validate_input",
"steps": {
"validate_input": {
"action": "validate_url",
"description": "Validate and normalize URL",
"config": {
"url_field": "target_url",
"add_protocol_if_missing": true
},
"next_step": "scrape_main_page"
},
"scrape_main_page": {
"action": "firecrawl_scrape",
"description": "Scrape main page content",
"config": {
"capture_config": {
"formats": ["markdown", "html", "screenshot"],
"capture_screenshot": true,
"full_page": true,
"only_main_content": false
}
},
"next_step": "extract_structure"
},
"extract_structure": {
"action": "firecrawl_extract",
"description": "Extract page structure and components",
"config": {
"schema": {
"navigation": {"type": "array", "items": {"type": "string"}},
"hero_section": {"type": "object"},
"main_sections": {"type": "array"},
"footer_links": {"type": "array"},
"color_scheme": {"type": "object"},
"fonts_used": {"type": "array"}
},
"system_prompt": "Analyze the webpage and extract its structural components, design elements, and content organization"
},
"next_step": "crawl_subpages"
},
"crawl_subpages": {
"action": "firecrawl_crawl",
"description": "Crawl additional pages for context",
"config": {
"crawl_config": {
"limit": 5,
"max_depth": 1,
"formats": ["markdown", "html"]
}
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return captured data"
}
}
},
"processing_mode": "task",
"adapter_topic": "system.adapter.firecrawl.requests",
"timeout_seconds": 180
}'::jsonb,
true,
ARRAY['capture', 'firecrawl', 'scraping'],
'docker.io/aqls/agent-chassis',
'v1.0.407'
);

-- Simple webscrape agent using the registered "scrape_web" action
INSERT INTO agent_definitions (
id,
type,
display_name,
description,
category,
default_config,
is_active,
capabilities,
image_repository,
image_tag
) VALUES (
gen_random_uuid(),
'webscrape-simple',
'Simple Web Scraper',
'Captures website content using Firecrawl and saves to S3',
'data-extraction',
'{
"workflow": {
"start_step": "scrape_website",
"steps": {
"scrape_website": {
"action": "scrape_web",
"description": "Scrape website content and save to S3",
"config": {
"action": "scrape",
"upload_results": true,
"scrape_config": {
"formats": ["markdown", "html", "screenshot"],
"capture_screenshot": true,
"full_page": true,
"only_main_content": false
}
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return scraped data with S3 links"
}
}
},
"processing_mode": "task",
"timeout_seconds": 180
}'::jsonb,
true,
ARRAY['webscraping', 'firecrawl', 's3-upload'],
'docker.io/aqls/agent-chassis',
'v1.0.424'
);


---


INSERT INTO agent_definitions (
id,
type,
display_name,
description,
category,
default_config,
is_active,
capabilities,
image_repository,
image_tag
) VALUES (
gen_random_uuid(),
'website-capture-firecrawl',
'Website Capture Agent (Firecrawl)',
'Captures website content using Firecrawl API with validation and extraction',
'data-driven',
'{
"workflow": {
"start_step": "validate_input",
"steps": {
"validate_input": {
"action": "validate_url",
"description": "Validate and normalize URL",
"config": {
"url_field": "target_url",
"add_protocol_if_missing": true
},
"next_step": "scrape_main_page"
},
"scrape_main_page": {
"action": "firecrawl_scrape",
"description": "Scrape main page content",
"config": {
"upload_results": true,
"scrape_config": {
"formats": ["markdown", "html", "screenshot"],
"capture_screenshot": true,
"full_page": true,
"only_main_content": false
}
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return captured data"
}
}
},
"processing_mode": "task",
"timeout_seconds": 180
}'::jsonb,
true,
'["capture", "firecrawl", "scraping"]'::jsonb,
'docker.io/aqls/agent-chassis',
'v1.0.424'
);

-- Agent with extraction
INSERT INTO agent_definitions (
id,
type,
display_name,
description,
category,
default_config,
is_active,
capabilities,
image_repository,
image_tag
) VALUES (
gen_random_uuid(),
'website-extract-structured',
'Website Structure Extractor',
'Extracts structured data from websites',
'data-driven',
'{
"workflow": {
"start_step": "validate_input",
"steps": {
"validate_input": {
"action": "validate_url",
"config": {
"url_field": "target_url",
"add_protocol_if_missing": true
},
"next_step": "scrape_page"
},
"scrape_page": {
"action": "firecrawl_scrape",
"config": {
"upload_results": true,
"scrape_config": {
"formats": ["markdown", "html"],
"capture_screenshot": true
}
},
"next_step": "extract_structure"
},
"extract_structure": {
"action": "firecrawl_extract",
"config": {
"upload_results": true,
"scrape_config": {
"schema": {
"events": {"type": "array", "items": {"type": "object"}},
"ticket_prices": {"type": "array"},
"venue_info": {"type": "object"},
"navigation_links": {"type": "array"}
},
"system_prompt": "Extract boxing event information, ticket prices, and venue details"
}
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow"
}
}
},
"processing_mode": "task"
}'::jsonb,
true,
'["extraction", "structured-data"]'::jsonb,
'docker.io/aqls/agent-chassis',
'v1.0.424'
);


---

-- Agent Group Definition for Web Scraping Orchestration
INSERT INTO agent_group_definitions (
id,
name,
group_type,
agent_configs,
orchestration_workflow,
usage_count,
version,
created_at,
updated_at,
description
) VALUES (
gen_random_uuid(),
'Smart Website Analyzer',
'website-analyzer',
'[
{"role": "basic_scraper", "agent_type": "website-capture-firecrawl"},
{"role": "structure_extractor", "agent_type": "website-extract-structured"}
]'::jsonb,
'{
"start_step": "analyze_request",
"steps": {
"analyze_request": {
"action": "evaluate_condition",
"description": "Determine if structured extraction is needed",
"config": {
"condition": "{{.input_data.extract_structured}}",
"default": false
},
"next_step": {
"true": "spawn_extractor",
"false": "spawn_basic_scraper"
}
},
"spawn_basic_scraper": {
"action": "spawn_agent",
"description": "Spawn basic scraping agent",
"config": {
"role": "basic_scraper",
"agent_type": "website-capture-firecrawl"
},
"next_step": "execute_basic_scrape"
},
"execute_basic_scrape": {
"action": "call_agent",
"description": "Execute basic website scraping",
"config": {
"agent_type": "website-capture-firecrawl",
"target_role": "basic_scraper",
"input_data": {
"target_url": "{{.input_data.target_url}}"
},
"timeout_seconds": 180
},
"next_step": "check_crawl_needed"
},
"spawn_extractor": {
"action": "spawn_agent",
"description": "Spawn structured extraction agent",
"config": {
"role": "structure_extractor",
"agent_type": "website-extract-structured"
},
"next_step": "execute_extraction"
},
"execute_extraction": {
"action": "call_agent",
"description": "Execute structured data extraction",
"config": {
"agent_type": "website-extract-structured",
"target_role": "structure_extractor",
"input_data": {
"target_url": "{{.input_data.target_url}}"
},
"timeout_seconds": 240
},
"next_step": "check_crawl_needed"
},
"check_crawl_needed": {
"action": "evaluate_condition",
"description": "Check if multi-page crawling is requested",
"config": {
"condition": "{{.input_data.crawl_pages}}",
"default": false
},
"next_step": {
"true": "execute_crawl",
"false": "aggregate_results"
}
},
"execute_crawl": {
"action": "firecrawl_crawl",
"description": "Crawl additional pages from the website",
"config": {
"upload_results": true,
"scrape_config": {
"limit": "{{.input_data.crawl_limit | default 5}}",
"max_depth": "{{.input_data.crawl_depth | default 2}}",
"formats": ["markdown", "html"]
}
},
"next_step": "aggregate_results"
},
"aggregate_results": {
"action": "aggregate_data",
"description": "Combine all scraping results",
"config": {
"response_fields": ["execute_basic_scrape", "execute_extraction", "execute_crawl"],
"include_metadata": true,
"format_output": {
"summary": true,
"s3_links": true,
"extracted_data": true
}
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return comprehensive scraping results"
}
}
}'::jsonb,
0,
1,
NOW(),
NOW(),
'Intelligent website analyzer that can perform basic scraping, structured extraction, and multi-page crawling based on requirements'
);


-- For parallel scraping of multiple URLs
INSERT INTO agent_group_definitions (
id,
name,
group_type,
agent_configs,
orchestration_workflow,
usage_count,
version,
created_at,
updated_at,
description
) VALUES (
gen_random_uuid(),
'Parallel Website Scraper',
'parallel-scraper',
'[
{"role": "scraper_1", "agent_type": "website-capture-firecrawl"},
{"role": "scraper_2", "agent_type": "website-capture-firecrawl"},
{"role": "scraper_3", "agent_type": "website-capture-firecrawl"}
]'::jsonb,
'{
"start_step": "prepare_batch",
"steps": {
"prepare_batch": {
"action": "split_urls",
"description": "Split URLs for parallel processing",
"config": {
"urls_field": "target_urls",
"max_parallel": 3
},
"next_step": "spawn_scrapers"
},
"spawn_scrapers": {
"action": "spawn_parallel",
"description": "Spawn multiple scraping agents",
"config": {
"agents": [
{"role": "scraper_1", "agent_type": "website-capture-firecrawl"},
{"role": "scraper_2", "agent_type": "website-capture-firecrawl"},
{"role": "scraper_3", "agent_type": "website-capture-firecrawl"}
]
},
"next_step": "execute_parallel_scrape"
},
"execute_parallel_scrape": {
"action": "parallel_call",
"description": "Execute scraping in parallel",
"config": {
"calls": [
{
"agent_type": "website-capture-firecrawl",
"target_role": "scraper_1",
"input_data": {"target_url": "{{.prepare_batch.batch_1}}"}
},
{
"agent_type": "website-capture-firecrawl",
"target_role": "scraper_2",
"input_data": {"target_url": "{{.prepare_batch.batch_2}}"}
},
{
"agent_type": "website-capture-firecrawl",
"target_role": "scraper_3",
"input_data": {"target_url": "{{.prepare_batch.batch_3}}"}
}
],
"timeout_seconds": 180,
"wait_for_all": true
},
"next_step": "combine_results"
},
"combine_results": {
"action": "merge_results",
"description": "Combine all parallel scraping results",
"config": {
"merge_strategy": "append",
"group_by": "url"
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return all scraped data"
}
}
}'::jsonb,
0,
1,
NOW(),
NOW(),
'Parallel scraper for handling multiple URLs simultaneously'
);


INSERT INTO agent_definitions (
id,
type,
display_name,
description,
category,
default_config,
is_active,
capabilities,
image_repository,
image_tag
) VALUES (
gen_random_uuid(),
'website-capture-firecrawl',
'Website Capture Agent (Firecrawl)',
'Captures website content using Firecrawl API',
'data-driven',
'{
"workflow": {
"start_step": "validate_input",
"steps": {
"validate_input": {
"action": "validate_url",
"description": "Validate and normalize URL",
"config": {
"url_field": "target_url",
"add_protocol_if_missing": true
},
"next_step": "scrape_main_page"
},
"scrape_main_page": {
"action": "firecrawl_scrape",
"description": "Scrape main page content",
"config": {
"capture_config": {
"formats": ["markdown", "html", "screenshot"],
"capture_screenshot": true,
"full_page": true,
"only_main_content": false
}
},
"next_step": "extract_structure"
},
"extract_structure": {
"action": "firecrawl_extract",
"description": "Extract page structure and components",
"config": {
"schema": {
"navigation": {"type": "array", "items": {"type": "string"}},
"hero_section": {"type": "object"},
"main_sections": {"type": "array"},
"footer_links": {"type": "array"},
"color_scheme": {"type": "object"},
"fonts_used": {"type": "array"}
},
"system_prompt": "Analyze the webpage and extract its structural components, design elements, and content organization"
},
"next_step": "crawl_subpages"
},
"crawl_subpages": {
"action": "firecrawl_crawl",
"description": "Crawl additional pages for context",
"config": {
"crawl_config": {
"limit": 5,
"max_depth": 1,
"formats": ["markdown", "html"]
}
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return captured data"
}
}
},
"processing_mode": "task",
"adapter_topic": "system.adapter.firecrawl.requests",
"timeout_seconds": 180
}'::jsonb,
true,
'["capture", "firecrawl", "scraping"]'::jsonb,
'docker.io/aqls/agent-chassis',
'v1.0.407'
);


---

without the evaluate step

INSERT INTO agent_group_definitions (
id,
name,
group_type,
agent_configs,
orchestration_workflow,
usage_count,
version,
created_at,
updated_at,
description
) VALUES (
gen_random_uuid(),
'Smart Website Analyzer',
'website-analyzer',
'[
{"role": "basic_scraper", "agent_type": "website-capture-firecrawl"},
{"role": "structure_extractor", "agent_type": "website-extract-structured"}
]'::jsonb,
'{
"start_step": "analyze_request",
"steps": {
"analyze_request": {
"action": "evaluate_condition",
"description": "Determine if structured extraction is needed",
"config": {
"condition": "{{.input_data.extract_structured}}",
"default": false
},
"next_step": {
"true": "spawn_extractor",
"false": "spawn_basic_scraper"
}
},
"spawn_basic_scraper": {
"action": "spawn_agent",
"description": "Spawn basic scraping agent",
"config": {
"role": "basic_scraper",
"agent_type": "website-capture-firecrawl"
},
"next_step": "execute_basic_scrape"
},
"execute_basic_scrape": {
"action": "call_agent",
"description": "Execute basic website scraping",
"config": {
"agent_type": "website-capture-firecrawl",
"target_role": "basic_scraper",
"input_data": {
"target_url": "{{.input_data.target_url}}"
},
"timeout_seconds": 180
},
"next_step": "check_crawl_needed"
},
"spawn_extractor": {
"action": "spawn_agent",
"description": "Spawn structured extraction agent",
"config": {
"role": "structure_extractor",
"agent_type": "website-extract-structured"
},
"next_step": "execute_extraction"
},
"execute_extraction": {
"action": "call_agent",
"description": "Execute structured data extraction",
"config": {
"agent_type": "website-extract-structured",
"target_role": "structure_extractor",
"input_data": {
"target_url": "{{.input_data.target_url}}"
},
"timeout_seconds": 240
},
"next_step": "execute_crawl"
},
"execute_crawl": {
"action": "firecrawl_crawl",
"description": "Crawl additional pages from the website",
"config": {
"upload_results": true,
"scrape_config": {
"limit": "{{.input_data.crawl_limit | default 5}}",
"max_depth": "{{.input_data.crawl_depth | default 2}}",
"formats": ["markdown", "html"]
}
},
"next_step": "aggregate_results"
},
"aggregate_results": {
"action": "aggregate_data",
"description": "Combine all scraping results",
"config": {
"response_fields": ["execute_basic_scrape", "execute_extraction", "execute_crawl"],
"include_metadata": true,
"format_output": {
"summary": true,
"s3_links": true,
"extracted_data": true
}
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return comprehensive scraping results"
}
}
}'::jsonb,
0,
1,
NOW(),
NOW(),
'Intelligent website analyzer that can perform basic scraping, structured extraction, and multi-page crawling based on requirements'
);


---
updated
====

UPDATE agent_group_definitions
SET orchestration_workflow = '{
"start_step": "spawn_basic_scraper",
"steps": {
"spawn_basic_scraper": {
"action": "spawn_agent",
"description": "Spawn basic scraping agent",
"config": {
"role": "basic_scraper",
"agent_type": "website-capture-firecrawl"
},
"next_step": "spawn_extractor"
},
"spawn_extractor": {
"action": "spawn_agent",
"description": "Spawn structured extraction agent",
"config": {
"role": "structure_extractor",
"agent_type": "website-extract-structured"
},
"next_step": "execute_basic_scrape"
},
"execute_basic_scrape": {
"action": "call_agent",
"description": "Execute basic website scraping",
"config": {
"agent_type": "website-capture-firecrawl",
"target_role": "basic_scraper",
"timeout_seconds": 180
},
"next_step": "execute_extraction"
},
"execute_extraction": {
"action": "call_agent",
"description": "Execute structured data extraction",
"config": {
"agent_type": "website-extract-structured",
"target_role": "structure_extractor",
"timeout_seconds": 240
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return all scraping results"
}
}
}'::jsonb
WHERE group_type = 'website-analyzer';
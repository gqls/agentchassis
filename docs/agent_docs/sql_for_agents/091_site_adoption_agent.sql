-- ============================================================================
-- Site Adoption Agent — Definition and Workflow
-- ============================================================================
--
-- Workflow:
--   ensure_site_record → crawl_site → analyze_site (LLM) → apply_plan → complete
--
-- The crawl step uses firecrawl_crawl which sends to the webscrape adapter
-- and awaits the async response. The adapter crawls the domain and returns
-- markdown content for each page discovered.
--
-- The analyze step uses execute_llm_prompt to classify pages, extract
-- identity/design/content structure into a JSON plan.
--
-- The apply step (apply_adoption_plan Go action) takes the structured plan
-- and creates site_specs, page records, and work items.
--
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
    is_active, image_repository, image_tag,
    default_config
) VALUES (
             'site-adoption-agent',
             'Site Adoption Agent',
             'Crawls an existing site, analyses structure and content, creates specs and work items to recreate it',
             'code-driven',
             true,
             'docker.io/aqls/agent-chassis',
             'latest',
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {
                                 "input_fields": ["site_id", "domain"]
                             },
                             "output_field": "site_record",
                             "next_step": "crawl_site",
                             "description": "Create or load site record for the domain"
                         },

                         "crawl_site": {
                             "action": "firecrawl_crawl",
                             "config": {
                                 "url_field": "site_record.domain_url",
                                 "scrape_config": {
                                     "only_main_content": true,
                                     "capture_screenshot": false,
                                     "max_pages": 20
                                 }
                             },
                             "output_field": "crawl_result",
                             "next_step": "check_crawl_success",
                             "description": "Crawl the existing site — all pages up to limit"
                         },

                         "check_crawl_success": {
                             "action": "conditional",
                             "config": {
                                 "condition": "crawl_result.success == true",
                                 "then_step": "analyze_site",
                                 "else_step": "crawl_failed"
                             },
                             "description": "Check if the crawl returned results"
                         },

                         "crawl_failed": {
                             "action": "complete_workflow",
                             "config": {
                                 "status": "failed",
                                 "error": "Site crawl failed or returned no content"
                             },
                             "description": "Fail if crawl produced nothing"
                         },

                         "analyze_site": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "prompt_template": "You are analysing a crawled website to plan its adoption into a new system.\n\nDomain: {{.site_record.domain}}\n\nCrawled content:\n{{.crawl_result.content}}\n\nAnalyse this site and produce a JSON object with the following structure. Respond ONLY with valid JSON, no markdown backticks or explanation.\n\n{\n  \"identity\": {\n    \"company_name\": \"extracted company/brand name\",\n    \"tagline\": \"extracted tagline or slogan if found\",\n    \"industry\": \"detected industry vertical\",\n    \"target_audience\": \"who the site serves\",\n    \"tone\": \"describe the writing tone (e.g. professional, friendly, technical)\",\n    \"services\": [{\"name\": \"...\", \"description\": \"...\"}]\n  },\n  \"design\": {\n    \"palette\": {\n      \"primary\": \"#hex\",\n      \"secondary\": \"#hex\",\n      \"accent\": \"#hex\",\n      \"background\": \"#hex\",\n      \"text\": \"#hex\"\n    },\n    \"typography\": {\n      \"heading_font\": \"font name or generic family\",\n      \"body_font\": \"font name or generic family\",\n      \"base_size\": \"16px\"\n    },\n    \"visual_tone\": \"describe the visual style (e.g. minimal, bold, corporate, playful)\"\n  },\n  \"pages\": [\n    {\n      \"name\": \"kebab-case page name\",\n      \"title\": \"Page Title\",\n      \"url\": \"/page.html\",\n      \"page_type\": \"content|blog-index|blog-post|tool|landing\",\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"nav_label\": \"Nav Label\",\n      \"meta_description\": \"extracted or inferred meta description\",\n      \"sections\": [\"hero\", \"features\", \"call-to-action\"],\n      \"existing_content\": {\n        \"hero\": \"The actual text content found in the hero section\",\n        \"features\": \"The actual text content found in the features section\"\n      }\n    }\n  ],\n  \"interactive_features\": [\n    {\n      \"name\": \"feature name\",\n      \"type\": \"calculator|search|form|filter|tool\",\n      \"description\": \"what it does\",\n      \"self_contained\": true,\n      \"page\": \"which page it appears on\"\n    }\n  ]\n}\n\nRules:\n- Extract ACTUAL content from the crawled pages, not placeholder descriptions\n- Page names must be kebab-case\n- The index/home page should be named \"index\"\n- Sections should use standard component names where possible: hero, features, call-to-action, generic-text-block, testimonials, pricing, faq, contact-form\n- For existing_content, extract the real text per section. Keep headings and paragraph text. Strip HTML tags.\n- Identify any interactive features (calculators, search, forms) separately — these need special handling\n- If you cannot determine a colour, omit it rather than guessing",
                                 "model": "claude-sonnet-4-20250514",
                                 "temperature": 0.2,
                                 "max_tokens": 8000,
                                 "response_format": "json"
                             },
                             "output_field": "adoption_analysis",
                             "next_step": "apply_plan",
                             "description": "LLM analyses crawled content — extracts identity, design, pages, sections"
                         },

                         "apply_plan": {
                             "action": "apply_adoption_plan",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "domain": "site_record.domain",
                                 "adoption_plan": "adoption_analysis"
                             },
                             "output_field": "adoption_result",
                             "next_step": "complete",
                             "description": "Write specs, create pages, create work items from analysis"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["site_record", "crawl_result", "adoption_result"]
                             },
                             "description": "Adoption plan applied — dispatch loop will process work items"
                         }
                     }
                 }
             }'::jsonb
         )
    ON CONFLICT (type,version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              updated_at = NOW();

-- ============================================================================
-- Registry entry (add to registry.go):
-- ============================================================================
--
-- "apply_adoption_plan": {
--     Handler:     ApplyAdoptionPlanAction,
--     Category:    "site",
--     Description: "Create specs, pages, and work items from site adoption analysis",
--     IsLocal:     true,
-- },
--
-- ============================================================================
-- Trigger example (CLI):
-- ============================================================================
--
-- To adopt mortgagecalculator.co.uk:
--
-- 1. Ensure the site record exists (or let ensure_site_record create it):
--    INSERT INTO sites (domain, company_name, status)
--    VALUES ('mortgagecalculator.co.uk', 'Mortgage Calculator', 'active')
--    ON CONFLICT (domain) DO NOTHING;
--
-- 2. Send the trigger message:
--    {
--      "headers": {
--        "correlation_id": "...",
--        "message_type": "request",
--        "action": "process"
--      },
--      "config": {
--        "workflow": { ... }  // uses default from agent_definitions
--      },
--      "input_data": {
--        "domain": "mortgagecalculator.co.uk"
--      }
--    }
--
-- The agent will: crawl → analyse → write specs → write pages → write items
-- Then the dispatch loop picks up the items and builds the site.

---
-- path correction

-- ============================================================================
-- Site Adoption Agent — Definition and Workflow (v2)
-- ============================================================================
--
-- Workflow:
--   ensure_site_record → crawl_site → format_crawl → analyze_site (LLM) → apply_plan → complete
--
-- Trigger input_data must include:
--   { "domain": "mortgagecalculator.co.uk", "url": "https://mortgagecalculator.co.uk" }
--
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
    is_active, image_repository, image_tag,
    default_config
) VALUES (
             'site-adoption-agent',
             'Site Adoption Agent',
             'Crawls an existing site, analyses structure and content, creates specs and work items to recreate it',
             'code-driven',
             true,
             'docker.io/aqls/agent-chassis',
             'latest',
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {
                                 "input_fields": ["site_id", "domain"]
                             },
                             "output_field": "site_record",
                             "next_step": "crawl_site",
                             "description": "Create or load site record for the domain"
                         },

                         "crawl_site": {
                             "action": "firecrawl_crawl",
                             "config": {
                                 "url_field": "input_data.url",
                                 "scrape_config": {
                                     "only_main_content": true,
                                     "capture_screenshot": false
                                 }
                             },
                             "output_field": "crawl_result",
                             "next_step": "format_crawl",
                             "description": "Crawl the existing site via webscrape adapter"
                         },

                         "format_crawl": {
                             "action": "format_research_content",
                             "config": {
                                 "scrape_field": "crawl_result",
                                 "max_content_per_source": 8000
                             },
                             "output_field": "formatted_crawl",
                             "next_step": "check_crawl_content",
                             "description": "Format crawl results into readable text for LLM analysis"
                         },

                         "check_crawl_content": {
                             "action": "conditional",
                             "config": {
                                 "condition": "formatted_crawl.content_quality != none",
                                 "then_step": "analyze_site",
                                 "else_step": "crawl_failed"
                             },
                             "description": "Check if the crawl returned usable content"
                         },

                         "crawl_failed": {
                             "action": "complete_workflow",
                             "config": {
                                 "status": "failed",
                                 "error": "Site crawl failed or returned no usable content"
                             },
                             "description": "Fail if crawl produced nothing"
                         },

                         "analyze_site": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "prompt_template": "You are analysing a crawled website to plan its adoption into our website building system.\n\nDomain: {{.site_record.domain}}\n\nCrawled content from the site:\n{{.formatted_crawl.research_text}}\n\nAnalyse this site and produce a JSON object with the following structure. Respond ONLY with valid JSON, no markdown backticks or preamble.\n\n{\n  \"identity\": {\n    \"company_name\": \"extracted company/brand name\",\n    \"tagline\": \"extracted tagline or slogan if found\",\n    \"industry\": \"detected industry vertical\",\n    \"target_audience\": \"who the site serves\",\n    \"tone\": \"describe the writing tone (e.g. professional, friendly, technical)\",\n    \"services\": [{\"name\": \"...\", \"description\": \"...\"}]\n  },\n  \"design\": {\n    \"palette\": {\n      \"primary\": \"#hex or description\",\n      \"secondary\": \"#hex or description\",\n      \"accent\": \"#hex or description\",\n      \"background\": \"#hex or description\",\n      \"text\": \"#hex or description\"\n    },\n    \"typography\": {\n      \"heading_font\": \"font name or generic family\",\n      \"body_font\": \"font name or generic family\"\n    },\n    \"visual_tone\": \"describe the visual style\"\n  },\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title\",\n      \"url\": \"/index.html\",\n      \"page_type\": \"content\",\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"nav_label\": \"Home\",\n      \"meta_description\": \"extracted or inferred meta description\",\n      \"sections\": [\"hero\", \"features\", \"call-to-action\"],\n      \"existing_content\": {\n        \"hero\": {\n          \"heading\": \"The actual heading text\",\n          \"subheading\": \"The actual subheading text\",\n          \"cta_text\": \"Button text if any\"\n        },\n        \"features\": {\n          \"heading\": \"Section heading\",\n          \"content\": \"The actual paragraph text from this section\"\n        }\n      }\n    }\n  ],\n  \"interactive_features\": [\n    {\n      \"name\": \"feature name\",\n      \"type\": \"calculator|search|form|filter|tool\",\n      \"description\": \"what it does\",\n      \"self_contained\": true,\n      \"page\": \"which page it appears on\"\n    }\n  ]\n}\n\nRules:\n- Extract ACTUAL content text from the crawled pages, not placeholder descriptions\n- Page names must be kebab-case (index for homepage)\n- Use standard section names where possible: hero, features, call-to-action, generic-text-block, testimonials, pricing, faq, contact-form\n- For existing_content, extract real headings and paragraph text per section\n- Identify interactive features (calculators, search, forms) separately\n- If you cannot determine colours from the content, use descriptive terms\n- Only include pages you have actual content for\n- Omit the interactive_features array if there are none",
                                 "model": "claude-sonnet-4-20250514",
                                 "temperature": 0.2,
                                 "max_tokens": 8000
                             },
                             "output_field": "adoption_analysis",
                             "next_step": "apply_plan",
                             "description": "LLM analyses crawled content — extracts identity, design, pages, sections"
                         },

                         "apply_plan": {
                             "action": "apply_adoption_plan",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "domain": "site_record.domain",
                                 "adoption_plan": "adoption_analysis"
                             },
                             "output_field": "adoption_result",
                             "next_step": "complete",
                             "description": "Write specs, create pages, create work items from analysis"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["site_record", "formatted_crawl", "adoption_analysis", "adoption_result"]
                             },
                             "description": "Adoption plan applied — dispatch loop will process work items"
                         }
                     }
                 }
             }'::jsonb
         )
    ON CONFLICT (type) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              updated_at = NOW();

---
--

-- ============================================================================
-- Site Adoption Agent — Definition and Workflow (v2)
-- ============================================================================
--
-- Workflow:
--   ensure_site_record → crawl_site → format_crawl → analyze_site (LLM) → apply_plan → complete
--
-- Trigger input_data must include:
--   { "domain": "mortgagecalculator.co.uk", "url": "https://mortgagecalculator.co.uk" }
--
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category,
    is_active, image_repository, image_tag,
    default_config
) VALUES (
             'site-adoption-agent',
             'Site Adoption Agent',
             'Crawls an existing site, analyses structure and content, creates specs and work items to recreate it',
             'code-driven',
             true,
             'docker.io/aqls/agent-chassis',
             'latest',
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {
                                 "input_fields": ["site_id", "domain"]
                             },
                             "output_field": "site_record",
                             "next_step": "crawl_site",
                             "description": "Create or load site record for the domain"
                         },

                         "crawl_site": {
                             "action": "firecrawl_crawl",
                             "config": {
                                 "url_field": "input_data.url",
                                 "scrape_config": {
                                     "only_main_content": true,
                                     "capture_screenshot": false
                                 }
                             },
                             "output_field": "crawl_result",
                             "next_step": "format_crawl",
                             "description": "Crawl the existing site via webscrape adapter"
                         },

                         "format_crawl": {
                             "action": "format_research_content",
                             "config": {
                                 "scrape_field": "crawl_result",
                                 "max_content_per_source": 8000
                             },
                             "output_field": "formatted_crawl",
                             "next_step": "check_crawl_content",
                             "description": "Format crawl results into readable text for LLM analysis"
                         },

                         "check_crawl_content": {
                             "action": "conditional",
                             "config": {
                                 "condition": "formatted_crawl.content_quality != none",
                                 "then_step": "analyze_site",
                                 "else_step": "crawl_failed"
                             },
                             "description": "Check if the crawl returned usable content"
                         },

                         "crawl_failed": {
                             "action": "complete_workflow",
                             "config": {
                                 "status": "failed",
                                 "error": "Site crawl failed or returned no usable content"
                             },
                             "description": "Fail if crawl produced nothing"
                         },

                         "analyze_site": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "prompt_template": "You are analysing a crawled website to plan its adoption into our website building system.\n\nDomain: {{.site_record.domain}}\n\nCrawled content from the site:\n{{.formatted_crawl.research_text}}\n\nAnalyse this site and produce a JSON object with the following structure. Respond ONLY with valid JSON, no markdown backticks or preamble.\n\n{\n  \"identity\": {\n    \"company_name\": \"extracted company/brand name\",\n    \"tagline\": \"extracted tagline or slogan if found\",\n    \"industry\": \"detected industry vertical\",\n    \"target_audience\": \"who the site serves\",\n    \"tone\": \"describe the writing tone (e.g. professional, friendly, technical)\",\n    \"services\": [{\"name\": \"...\", \"description\": \"...\"}]\n  },\n  \"design\": {\n    \"palette\": {\n      \"primary\": \"#hex or description\",\n      \"secondary\": \"#hex or description\",\n      \"accent\": \"#hex or description\",\n      \"background\": \"#hex or description\",\n      \"text\": \"#hex or description\"\n    },\n    \"typography\": {\n      \"heading_font\": \"font name or generic family\",\n      \"body_font\": \"font name or generic family\"\n    },\n    \"visual_tone\": \"describe the visual style\"\n  },\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title\",\n      \"url\": \"/index.html\",\n      \"page_type\": \"content\",\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"nav_label\": \"Home\",\n      \"meta_description\": \"extracted or inferred meta description\",\n      \"sections\": [\"hero\", \"features\", \"call-to-action\"],\n      \"existing_content\": {\n        \"hero\": {\n          \"heading\": \"The actual heading text\",\n          \"subheading\": \"The actual subheading text\",\n          \"cta_text\": \"Button text if any\"\n        },\n        \"features\": {\n          \"heading\": \"Section heading\",\n          \"content\": \"The actual paragraph text from this section\"\n        }\n      }\n    }\n  ],\n  \"interactive_features\": [\n    {\n      \"name\": \"feature name\",\n      \"type\": \"calculator|search|form|filter|tool\",\n      \"description\": \"what it does\",\n      \"self_contained\": true,\n      \"page\": \"which page it appears on\"\n    }\n  ]\n}\n\nRules:\n- Extract ACTUAL content text from the crawled pages, not placeholder descriptions\n- Page names must be kebab-case (index for homepage)\n- Use standard section names where possible: hero, features, call-to-action, generic-text-block, testimonials, pricing, faq, contact-form\n- For existing_content, extract real headings and paragraph text per section\n- Identify interactive features (calculators, search, forms) separately\n- If you cannot determine colours from the content, use descriptive terms\n- Only include pages you have actual content for\n- Omit the interactive_features array if there are none",
                                 "model": "claude-sonnet-4-20250514",
                                 "temperature": 0.2,
                                 "max_tokens": 8000
                             },
                             "output_field": "adoption_analysis",
                             "next_step": "apply_plan",
                             "description": "LLM analyses crawled content — extracts identity, design, pages, sections"
                         },

                         "apply_plan": {
                             "action": "apply_adoption_plan",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "domain": "site_record.domain",
                                 "adoption_plan": "adoption_analysis"
                             },
                             "output_field": "adoption_result",
                             "next_step": "complete",
                             "description": "Write specs, create pages, create work items from analysis"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["site_record", "formatted_crawl", "adoption_analysis", "adoption_result"]
                             },
                             "description": "Adoption plan applied — dispatch loop will process work items"
                         }
                     }
                 }
             }'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              updated_at = NOW();


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

---
-- fix firescrape api call

-- 1. Fix the adoption agent workflow — remove scrape_config
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,crawl_site,config}',
        '{
            "url_field": "input_data.url"
        }'
                     )
WHERE type = 'site-adoption-agent';

-- 2. Fail the stuck orchestration so we can retry cleanly
UPDATE orchestration_states
SET status = 'FAILED',
    error = 'Crawl failed: Firecrawl v2 API rejected scrape_config parameters',
    updated_at = NOW()
WHERE correlation_id = 'ecdef20d-fc92-4e75-aa5d-c4808506d057';

---
-- fix
CREATE OR REPLACE FUNCTION take_site_snapshot(
    p_site_id       UUID,
    p_trigger       TEXT,
    p_git_sha       TEXT DEFAULT NULL,
    p_label         TEXT DEFAULT NULL,
    p_created_by    TEXT DEFAULT 'system'
) RETURNS UUID AS $$
DECLARE
v_snapshot_id   UUID;
    v_site_record   JSONB;
    v_spec_snapshot JSONB;
    v_spec_ids      UUID[];
    v_pages         JSONB;
    v_nav           JSONB;
    v_components    JSONB;
BEGIN
SELECT jsonb_build_object(
               'id', s.id,
               'domain', s.domain,
               'status', s.status,
               'company_name', s.company_name,
               'tagline', s.tagline,
               'schema_mode', s.schema_mode,
               'style_collection_id', s.style_collection_id,
               'default_components', s.default_components,
               'content_data', s.content_data,
               'brand_assets', s.brand_assets,
               'deploy_config', s.deploy_config,
               'last_built_at', s.last_built_at,
               'last_deployed_at', s.last_deployed_at
       ) INTO v_site_record
FROM sites s
WHERE s.id = p_site_id;

IF v_site_record IS NULL THEN
        RAISE EXCEPTION 'Site % not found', p_site_id;
END IF;

SELECT
    COALESCE(jsonb_agg(
                     jsonb_build_object(
                             'id', ss.id,
                             'aspect', ss.aspect,
                             'data', ss.data,
                             'source', ss.source,
                             'source_agent', ss.source_agent,
                             'created_by', ss.created_by,
                             'pinned', ss.pinned,
                             'created_at', ss.created_at
                     ) ORDER BY ss.aspect
             ), '[]'::jsonb),
    COALESCE(array_agg(ss.id), ARRAY[]::uuid[])
INTO v_spec_snapshot, v_spec_ids
FROM site_specs ss
WHERE ss.site_id = p_site_id AND ss.is_current = true;

SELECT COALESCE(jsonb_agg(page_row ORDER BY page_row->>'nav_order'), '[]'::jsonb)
INTO v_pages
FROM (
         SELECT jsonb_build_object(
                        'id', p.id,
                        'name', p.name,
                        'url', p.url,
                        'title', p.title,
                        'page_type', p.page_type,
                        'status', p.status,
                        'meta_description', p.meta_description,
                        'topics', to_jsonb(p.topics),
                        'nav_label', p.nav_label,
                        'nav_order', p.nav_order,
                        'in_header', p.in_header,
                        'in_footer', p.in_footer,
                        'build_status', p.build_status,
                        'version', p.version,
                        'sections', p.sections,
                        'rendered_header', p.rendered_header,
                        'rendered_footer', p.rendered_footer,
                        'rendered_head', p.rendered_head,
                        'page_spec', p.page_spec,
                        'content_direction', p.content_direction,
                        'site_area_id', p.site_area_id,
                        'components', COALESCE(
                                (SELECT jsonb_agg(
                                                jsonb_build_object(
                                                        'id', pc.id,
                                                        'component_id', pc.component_id,
                                                        'position', pc.position,
                                                        'slot_name', pc.slot_name,
                                                        'rendered_html', pc.rendered_html,
                                                        'content_data', pc.content_data,
                                                        'build_status', pc.build_status
                                                ) ORDER BY pc.position
                                        )
                                 FROM page_components pc
                                 WHERE pc.page_id = p.id
                                ), '[]'::jsonb
                                      )
                ) AS page_row
         FROM pages p
         WHERE p.site_id = p_site_id
     ) sub;

SELECT jsonb_build_object(
               'groups', COALESCE(
                (SELECT jsonb_agg(
                                jsonb_build_object(
                                        'id', g.id,
                                        'group_key', g.group_key,
                                        'group_label', g.group_label,
                                        'group_type', g.group_type,
                                        'parent_group_id', g.parent_group_id,
                                        'position', g.position
                                ) ORDER BY g.position
                        )
                 FROM site_nav_groups g
                 WHERE g.site_id = p_site_id
                ), '[]'::jsonb
                         ),
               'items', COALESCE(
                       (SELECT jsonb_agg(
                                       jsonb_build_object(
                                               'id', ni.id,
                                               'group_id', ni.group_id,
                                               'parent_item_id', ni.parent_item_id,
                                               'label', ni.label,
                                               'url', ni.url,
                                               'page_id', ni.page_id,
                                               'item_type', ni.item_type,
                                               'position', ni.position,
                                               'status', ni.status,
                                               'metadata', ni.metadata
                                       ) ORDER BY ni.position
                               )
                        FROM site_nav_items ni
                        WHERE ni.site_id = p_site_id
                       ), '[]'::jsonb
                        )
       ) INTO v_nav;

SELECT COALESCE(jsonb_agg(
                        jsonb_build_object(
                                'id', sc.id,
                                'slot_name', sc.slot_name,
                                'component_id', sc.component_id,
                                'rendered_html', sc.rendered_html,
                                'content_data', sc.content_data,
                                'build_status', sc.build_status,
                                'locked_at', sc.locked_at,
                                'locked_by', sc.locked_by
                        )
                ), '[]'::jsonb)
INTO v_components
FROM site_components sc
WHERE sc.site_id = p_site_id;

INSERT INTO site_snapshots (
    site_id, trigger, git_commit_sha, label,
    site_record, spec_snapshot, pages_snapshot, nav_snapshot,
    components_snapshot, spec_ids, created_by
) VALUES (
             p_site_id, p_trigger, p_git_sha, p_label,
             v_site_record, v_spec_snapshot, v_pages, v_nav,
             v_components, v_spec_ids, p_created_by
         ) RETURNING id INTO v_snapshot_id;

RAISE NOTICE 'Snapshot % created for site % (trigger: %, specs: %, pages: %)',
        v_snapshot_id, p_site_id, p_trigger,
        jsonb_array_length(v_spec_snapshot),
        jsonb_array_length(v_pages);

RETURN v_snapshot_id;
END;
$$ LANGUAGE plpgsql;

   -- fix for revert site too
   CREATE OR REPLACE FUNCTION revert_site_to_snapshot(
    p_snapshot_id   UUID,
    p_reverted_by   TEXT DEFAULT 'admin'
) RETURNS JSONB AS $$
DECLARE
v_snap          RECORD;
    v_safety_id     UUID;
    v_spec          JSONB;
    v_page          JSONB;
    v_page_id       UUID;
    v_comp          JSONB;
    v_group         JSONB;
    v_item          JSONB;
    v_sc            JSONB;
    v_pages_restored INT := 0;
    v_specs_restored INT := 0;
    v_comps_restored INT := 0;
BEGIN
SELECT * INTO v_snap FROM site_snapshots WHERE id = p_snapshot_id;
IF NOT FOUND THEN
        RAISE EXCEPTION 'Snapshot % not found', p_snapshot_id;
END IF;

    v_safety_id := take_site_snapshot(
        v_snap.site_id, 'pre_revert', NULL,
        'Auto-snapshot before revert to ' || p_snapshot_id::text,
        p_reverted_by
    );

    -- ── 1. Restore site_specs ──────────────────────────────────────────

UPDATE site_specs
SET is_current = false, superseded_at = NOW()
WHERE site_id = v_snap.site_id AND is_current = true;

FOR v_spec IN SELECT * FROM jsonb_array_elements(v_snap.spec_snapshot)
                                LOOP
    INSERT INTO site_specs (
    site_id, aspect, data, source, source_agent,
    created_by, pinned, is_current
) VALUES (
                  v_snap.site_id,
                  v_spec->>'aspect',
                  v_spec->'data',
                  'snapshot_revert',
                  v_spec->>'source_agent',
                  p_reverted_by,
                  COALESCE((v_spec->>'pinned')::boolean, false),
                  true
                  );
v_specs_restored := v_specs_restored + 1;
END LOOP;

    -- ── 2. Restore pages and page_components ───────────────────────────

DELETE FROM page_components
WHERE page_id IN (SELECT id FROM pages WHERE site_id = v_snap.site_id);

DELETE FROM pages WHERE site_id = v_snap.site_id;

FOR v_page IN SELECT * FROM jsonb_array_elements(v_snap.pages_snapshot)
                                LOOP
    INSERT INTO pages (
    id, site_id, name, url, title, page_type, status,
    meta_description, topics, nav_label, nav_order,
    in_header, in_footer, build_status, version,
    sections, rendered_header, rendered_footer, rendered_head,
    page_spec, content_direction, site_area_id
) VALUES (
                  (v_page->>'id')::uuid,
                  v_snap.site_id,
                  v_page->>'name',
                  v_page->>'url',
                  v_page->>'title',
                  v_page->>'page_type',
                  v_page->>'status',
                  v_page->>'meta_description',
                  CASE WHEN v_page->'topics' IS NOT NULL AND v_page->'topics' != 'null'::jsonb
                  THEN ARRAY(SELECT jsonb_array_elements_text(v_page->'topics'))
                  ELSE NULL
                  END,
                  v_page->>'nav_label',
                  COALESCE((v_page->>'nav_order')::int, 100),
                  COALESCE((v_page->>'in_header')::boolean, true),
                  COALESCE((v_page->>'in_footer')::boolean, true),
                  COALESCE(v_page->>'build_status', 'deployed'),
                  COALESCE((v_page->>'version')::int, 1),
                  COALESCE(v_page->'sections', '[]'::jsonb),
                  v_page->>'rendered_header',
                  v_page->>'rendered_footer',
                  v_page->>'rendered_head',
                  v_page->'page_spec',
                  v_page->'content_direction',
                  CASE WHEN v_page->>'site_area_id' IS NOT NULL
                  THEN (v_page->>'site_area_id')::uuid
                  ELSE NULL
                  END
                  );

v_page_id := (v_page->>'id')::uuid;

FOR v_comp IN SELECT * FROM jsonb_array_elements(v_page->'components')
                                LOOP
    INSERT INTO page_components (
    page_id, component_id, position, slot_name,
    rendered_html, content_data, build_status
) VALUES (
                  v_page_id,
                  CASE WHEN v_comp->>'component_id' IS NOT NULL
                  THEN (v_comp->>'component_id')::uuid
                  ELSE NULL
                  END,
                  COALESCE((v_comp->>'position')::int, 0),
                  v_comp->>'slot_name',
                  v_comp->>'rendered_html',
                  COALESCE(v_comp->'content_data', '{}'::jsonb),
                  COALESCE(v_comp->>'build_status', 'deployed')
                  );
v_comps_restored := v_comps_restored + 1;
END LOOP;

        v_pages_restored := v_pages_restored + 1;
END LOOP;

    -- ── 3. Restore navigation ──────────────────────────────────────────

DELETE FROM site_nav_items WHERE site_id = v_snap.site_id;
DELETE FROM site_nav_groups WHERE site_id = v_snap.site_id;

FOR v_group IN SELECT * FROM jsonb_array_elements(v_snap.nav_snapshot->'groups')
                                 LOOP
    INSERT INTO site_nav_groups (
    id, site_id, group_key, group_label, group_type,
    parent_group_id, position
) VALUES (
                   (v_group->>'id')::uuid,
                   v_snap.site_id,
                   v_group->>'group_key',
                   COALESCE(v_group->>'group_label', ''),
                   v_group->>'group_type',
                   CASE WHEN v_group->>'parent_group_id' IS NOT NULL
                   THEN (v_group->>'parent_group_id')::uuid
                   ELSE NULL
                   END,
                   COALESCE((v_group->>'position')::int, 0)
                   );
END LOOP;

FOR v_item IN SELECT * FROM jsonb_array_elements(v_snap.nav_snapshot->'items')
                                LOOP
    INSERT INTO site_nav_items (
    id, site_id, group_id, parent_item_id,
    label, url, page_id, item_type,
    position, status, metadata
) VALUES (
                  (v_item->>'id')::uuid,
                  v_snap.site_id,
                  (v_item->>'group_id')::uuid,
                  CASE WHEN v_item->>'parent_item_id' IS NOT NULL
                  THEN (v_item->>'parent_item_id')::uuid
                  ELSE NULL
                  END,
                  v_item->>'label',
                  v_item->>'url',
                  CASE WHEN v_item->>'page_id' IS NOT NULL
                  THEN (v_item->>'page_id')::uuid
                  ELSE NULL
                  END,
                  COALESCE(v_item->>'item_type', 'page_link'),
                  COALESCE((v_item->>'position')::int, 0),
                  COALESCE(v_item->>'status', 'active'),
                  COALESCE(v_item->'metadata', '{}'::jsonb)
                  );
END LOOP;

    -- ── 4. Restore site_components ─────────────────────────────────────

DELETE FROM site_components WHERE site_id = v_snap.site_id;

FOR v_sc IN SELECT * FROM jsonb_array_elements(v_snap.components_snapshot)
                              LOOP
    INSERT INTO site_components (
    id, site_id, slot_name, component_id,
    rendered_html, content_data, build_status
) VALUES (
                (v_sc->>'id')::uuid,
                v_snap.site_id,
                v_sc->>'slot_name',
                CASE WHEN v_sc->>'component_id' IS NOT NULL
                THEN (v_sc->>'component_id')::uuid
                ELSE NULL
                END,
                v_sc->>'rendered_html',
                COALESCE(v_sc->'content_data', '{}'::jsonb),
                COALESCE(v_sc->>'build_status', 'pending')
                );
END LOOP;

    -- ── 5. Restore site record fields ──────────────────────────────────

UPDATE sites SET
                 status = COALESCE(v_snap.site_record->>'status', status),
                 schema_mode = COALESCE(v_snap.site_record->>'schema_mode', schema_mode),
                 default_components = COALESCE(v_snap.site_record->'default_components', default_components),
                 updated_at = NOW()
WHERE id = v_snap.site_id;

RETURN jsonb_build_object(
        'reverted', true,
        'snapshot_id', p_snapshot_id,
        'safety_snapshot_id', v_safety_id,
        'site_id', v_snap.site_id,
        'specs_restored', v_specs_restored,
        'pages_restored', v_pages_restored,
        'components_restored', v_comps_restored,
        'snapshot_trigger', v_snap.trigger,
        'snapshot_created_at', v_snap.created_at
       );
END;
$$ LANGUAGE plpgsql;

   --

   -- fix firecrawl - added action to parse it

   -- Update the format_crawl step to use the new action
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,format_crawl}',
        '{
            "action": "format_crawl_for_analysis",
            "config": {
                "crawl_field": "crawl_result",
                "max_content_per_page": 8000
            },
            "output_field": "formatted_crawl",
            "next_step": "check_crawl_content",
            "description": "Format crawl pages into readable text for LLM analysis"
        }'
                     )
WHERE type = 'site-adoption-agent';

-- Update format step to use the new action
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,format_crawl}',
        '{
            "action": "format_crawl_for_analysis",
            "config": {
                "crawl_field": "crawl_result",
                "max_content_per_page": 8000
            },
            "output_field": "formatted_crawl",
            "next_step": "check_crawl_content",
            "description": "Format crawl pages into readable text for LLM analysis"
        }'
                     )
WHERE type = 'site-adoption-agent';

-- Fail stuck orchestration
UPDATE orchestration_states
SET status = 'FAILED', error = 'format_research_content cannot parse crawl output — replaced with format_crawl_for_analysis'
WHERE correlation_id = '0d3a3da5-eb2a-47e2-9f78-ffbf6698ff22';

--
-- Fix the model name in ai_service
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{ai_service}',
        '{
            "provider": "anthropic",
            "model": "claude-sonnet-4-5",
            "api_key_env_var": "ANTHROPIC_API_KEY"
        }'
                     )
WHERE type = 'site-adoption-agent';

-- Also fix it in the prompt step config
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_site,config,model}',
        '"claude-sonnet-4-5"'
                     )
WHERE type = 'site-adoption-agent';

--

-- This is the correct pattern per the architecture
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_site,config,ai_service}',
        '{
            "provider": "anthropic",
            "model": "claude-sonnet-4-5",
            "api_key_env_var": "ANTHROPIC_API_KEY"
        }'
                     )
WHERE type = 'site-adoption-agent';


---

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_site,config,input_fields}',
        '["site_record", "formatted_crawl", "input_data"]'
                     )
WHERE type = 'site-adoption-agent';

--

-- Increase max_tokens for the analysis step (10 pages needs more than 8000)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_site,config,max_tokens}',
        '16000'
                     )
WHERE type = 'site-adoption-agent';

---
--

-- Lighter prompt: structure only, no existing_content (Go extracts content from crawl)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_site,config,prompt_template}',
        '"You are analysing a crawled website to plan its adoption into our website building system.\n\nDomain: {{.site_record.domain}}\n\nCrawled page summaries:\n{{.formatted_crawl.research_text}}\n\nAnalyse this site and produce a JSON object. Respond ONLY with valid JSON, no markdown backticks.\n\n{\n  \"identity\": {\n    \"company_name\": \"extracted company/brand name\",\n    \"tagline\": \"extracted tagline or slogan\",\n    \"industry\": \"detected industry vertical\",\n    \"target_audience\": \"who the site serves\",\n    \"tone\": \"writing tone description\",\n    \"services\": [{\"name\": \"...\", \"description\": \"...\"}]\n  },\n  \"design\": {\n    \"palette\": {\"primary\": \"#hex or description\", \"secondary\": \"#hex or description\", \"accent\": \"#hex or description\", \"background\": \"#hex or description\", \"text\": \"#hex or description\"},\n    \"typography\": {\"heading_font\": \"font name\", \"body_font\": \"font name\"},\n    \"visual_tone\": \"visual style description\"\n  },\n  \"pages\": [\n    {\n      \"name\": \"kebab-case-name\",\n      \"title\": \"Page Title\",\n      \"url\": \"/path/to/page.html\",\n      \"page_type\": \"content|tool|blog-index|blog-post|landing\",\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"nav_label\": \"Nav Label\",\n      \"meta_description\": \"page description\",\n      \"sections\": [\"hero\", \"features\", \"call-to-action\"]\n    }\n  ],\n  \"interactive_features\": [\n    {\"name\": \"feature\", \"type\": \"calculator|search|form|tool\", \"description\": \"what it does\", \"self_contained\": true, \"page\": \"page-name\"}\n  ]\n}\n\nRules:\n- Page names must be kebab-case (index for homepage)\n- Use standard section names: hero, features, call-to-action, generic-text-block, testimonials, pricing, faq, contact-form, guide-list, tool-list, game-list\n- Do NOT include existing_content — content extraction is handled separately\n- Identify interactive features (calculators, search, games, simulations) separately\n- Only include pages that have actual content (skip 404s)\n- Omit interactive_features if none exist"'
                     )
WHERE type = 'site-adoption-agent';

-- Also add crawl_result to input_fields so Go can access raw pages
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,apply_plan,config,input_fields}',
        '["site_record", "crawl_result", "adoption_analysis"]'
                     )
WHERE type = 'site-adoption-agent';

---

-- 1. Format step: summary mode (500 chars per page for LLM)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,format_crawl,config}',
        '{
            "crawl_field": "crawl_result",
            "summary_chars_per_page": 500
        }'
                     )
WHERE type = 'site-adoption-agent';

-- 2. Lighter prompt (structure only, no existing_content)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_site,config,prompt_template}',
        '"You are analysing a crawled website to plan its adoption into our website building system.\n\nDomain: {{.site_record.domain}}\n\nCrawled page summaries:\n{{.formatted_crawl.research_text}}\n\nAnalyse this site and produce a JSON object. Respond ONLY with valid JSON, no markdown backticks.\n\n{\n  \"identity\": {\n    \"company_name\": \"extracted company/brand name\",\n    \"tagline\": \"extracted tagline or slogan\",\n    \"industry\": \"detected industry vertical\",\n    \"target_audience\": \"who the site serves\",\n    \"tone\": \"writing tone description\",\n    \"services\": [{\"name\": \"...\", \"description\": \"...\"}]\n  },\n  \"design\": {\n    \"palette\": {\"primary\": \"#hex or description\", \"secondary\": \"#hex or description\", \"accent\": \"#hex or description\", \"background\": \"#hex or description\", \"text\": \"#hex or description\"},\n    \"typography\": {\"heading_font\": \"font name\", \"body_font\": \"font name\"},\n    \"visual_tone\": \"visual style description\"\n  },\n  \"pages\": [\n    {\n      \"name\": \"kebab-case-name\",\n      \"title\": \"Page Title\",\n      \"url\": \"/path/to/page.html\",\n      \"page_type\": \"content|tool|blog-index|blog-post|landing\",\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"nav_label\": \"Nav Label\",\n      \"meta_description\": \"page description\",\n      \"sections\": [\"hero\", \"features\", \"call-to-action\"]\n    }\n  ],\n  \"interactive_features\": [\n    {\"name\": \"feature\", \"type\": \"calculator|search|form|tool\", \"description\": \"what it does\", \"self_contained\": true, \"page\": \"page-name\"}\n  ]\n}\n\nRules:\n- Page names must be kebab-case (index for homepage)\n- Use standard section names: hero, features, call-to-action, generic-text-block, testimonials, pricing, faq, contact-form, guide-list, tool-list, game-list\n- Do NOT include existing_content — content extraction is handled separately\n- Identify interactive features (calculators, search, games, simulations) separately\n- Only include pages that have actual content (skip 404s)\n- Omit interactive_features if none exist"'
                     )
WHERE type = 'site-adoption-agent';

-- 3. apply_plan needs crawl_result for Go-side content extraction
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,apply_plan,config,input_fields}',
        '["site_record", "crawl_result", "adoption_analysis"]'
                     )
WHERE type = 'site-adoption-agent';

-- 4. Clean up gamedesign.uk test data
DELETE FROM site_work_items WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk');
DELETE FROM site_specs WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk');
DELETE FROM research_results WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk');
DELETE FROM pages WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk');
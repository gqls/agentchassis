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

---
-- start of 2 step adopt to move toward bigger sites

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

---
-- save raw html as well as markdown
-- Add rawHtml to crawl scrapeOptions
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,crawl_site,config,scrape_config}',
        '{
            "only_main_content": false,
            "formats": ["markdown", "rawHtml"]
        }'
                     )
WHERE type = 'site-adoption-agent';

-- Update adoption agent analyze_site to claude-sonnet-4-6
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_site,config,ai_service,model}',
        '"claude-sonnet-4-6"'
                     )
WHERE type = 'site-adoption-agent';

--

-- ============================================================================
-- Add content direction analysis to the site-adoption-agent workflow
-- ============================================================================
--
-- Adds two steps between analyze_site and apply_plan:
--   1. select_content — picks 2-3 prose-heavy pages from crawl
--   2. derive_content_direction — LLM extracts detailed writing guidelines
--
-- New flow:
--   ... → analyze_site → select_content → derive_content_direction → apply_plan
--
-- ============================================================================

-- ── 1. Redirect analyze_site → select_content (was → apply_plan) ────────

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_site,next_step}',
        '"select_content"'
                     )
WHERE type = 'site-adoption-agent';

-- ── 2. Add select_content step ──────────────────────────────────────────
-- Note: no input_fields — Go actions read from params.CollectedData directly.
-- input_fields is only consumed by execute_llm_prompt / extractDataForAiAgent.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,select_content}',
        '{
            "action": "select_representative_content",
            "config": {
                "max_pages": 3,
                "max_total_chars": 15000
            },
            "output_field": "representative_content",
            "next_step": "derive_content_direction",
            "error_step": "apply_plan",
            "description": "Select 2-3 prose-heavy pages from crawl for writing style analysis"
        }'
                     )
WHERE type = 'site-adoption-agent';

-- ── 3. Add derive_content_direction step ────────────────────────────────
-- Uses {{if}} guards on template variables per dev guide rule 7.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,derive_content_direction}',
        '{
            "action": "execute_llm_prompt",
            "config": {
                "ai_service": {
                    "model": "claude-sonnet-4-6",
                    "provider": "anthropic",
                    "api_key_env_var": "ANTHROPIC_API_KEY"
                },
                "max_tokens": 6000,
                "temperature": 0.2,
                "input_fields": ["site_record", "representative_content", "adoption_analysis"],
                "prompt_template": "You are a senior copywriter and brand strategist. You are analysing the actual content from a website to produce a detailed writing style guide that another writer could follow to recreate content in exactly the same voice, tone, and style.\n\nDomain: {{if .site_record}}{{.site_record.domain}}{{end}}\n\nHere is the actual content from representative pages on this site:\n\n{{if .representative_content}}{{.representative_content.selected_text}}{{end}}\n\nStudy this content carefully and deeply. Pay attention to:\n- How headings are phrased (questions? statements? commands? how long?)\n- Sentence length, rhythm, and construction patterns\n- Level of formality vs conversational tone\n- How technical terms are handled (defined? assumed known? avoided? explained inline?)\n- Use of first/second/third person and when each is used\n- How confidence and authority are expressed without overstepping\n- What the content deliberately avoids saying (no promises? no superlatives? no jargon without explanation?)\n- How calls-to-action are worded (hard sell? soft? informational? what verbs?)\n- Any legal, regulatory, or compliance patterns (disclaimers, caveats, qualifications)\n- Use of examples, analogies, worked calculations, or concrete illustrations\n- How lists, steps, and structured content are formatted\n- The emotional register (reassuring? challenging? neutral? playful? dry?)\n- How the site builds trust (through expertise? through transparency? through relatability?)\n- What assumptions the site makes about what the reader already knows\n- How deeply the content explores each topic (surface overview? deep dive? practical application?)\n- How bold, italics, links, and other formatting are used\n- How sections transition from one idea to the next\n\nProduce a JSON writing style guide. Be as specific and detailed as possible — a writer reading this guide should be able to perfectly match the voice without ever seeing the original site. Respond ONLY with valid JSON, no markdown backticks.\n\n{\n  \"voice\": {\n    \"register\": \"Describe the overall register in one detailed sentence\",\n    \"person\": \"Which grammatical person is used and when\",\n    \"authority_level\": \"How the site establishes authority and credibility\",\n    \"emotional_tone\": \"The emotional quality and how it makes the reader feel\",\n    \"formality\": \"Where it sits on formal-casual spectrum with specific markers\"\n  },\n  \"sentence_style\": {\n    \"average_length\": \"Short/medium/long and typical word count per sentence\",\n    \"structure_patterns\": \"How sentences are typically constructed — e.g. 'Opens with the conclusion, follows with evidence. Avoids subordinate clauses. Uses fragments for emphasis.'\",\n    \"rhythm\": \"The cadence of the writing — e.g. 'Alternates between short punchy statements and longer explanatory sentences. Never more than two long sentences in a row.'\",\n    \"connectives\": \"How ideas are linked — e.g. 'Rarely uses transitional phrases. Paragraphs stand alone. When connecting, uses simple connectives (but, so, because) not formal ones (furthermore, consequently).'\"\n  },\n  \"persuasion_approach\": {\n    \"method\": \"How the site convinces without selling — e.g. 'Leads with utility. Shows rather than tells. Lets the tool/data speak for itself. Never claims to be the best — demonstrates value through the quality of the resource.'\",\n    \"trust_building\": \"How trust is established — e.g. 'Through visible expertise and practical demonstrations. Shows the working. Acknowledges limitations openly.'\",\n    \"social_proof_style\": \"How social proof is handled — e.g. 'No testimonials. No client logos. Authority comes from the depth of the content itself. Implicit proof through usage statistics if present.'\"\n  },\n  \"content_depth\": {\n    \"thoroughness\": \"How deeply topics are explored — e.g. 'Extremely thorough. Every claim has a worked example. Mathematical formulas are shown step-by-step. Nothing is left to assumption.'\",\n    \"assumed_knowledge\": \"What the reader is expected to already know — e.g. 'Assumes intermediate game design knowledge. Explains probability theory from scratch. Does not explain what an RPG is.'\",\n    \"explanation_pattern\": \"How complex ideas are introduced — e.g. 'Concept first, then formula, then worked example, then interactive demo. Always follows abstraction with a concrete case.'\"\n  },\n  \"writing_rules\": [\n    \"Each rule should be specific and actionable, not vague.\",\n    \"Extract at least 10-15 rules from actual patterns in the content.\",\n    \"Example: 'Technical terms must be explained in plain English on first use'\",\n    \"Example: 'Never use superlatives (best, leading, guaranteed) unless backed by a cited source'\",\n    \"Example: 'Paragraphs are 2-3 sentences maximum. One idea per paragraph.'\",\n    \"Example: 'All numerical claims must include the calculation or source'\",\n    \"Example: 'Use imperative mood for instructional content (Calculate, Compare, Choose) not passive (can be calculated)'\"\n  ],\n  \"compliance_rules\": [\n    \"Any legal, regulatory, financial, medical, or professional compliance patterns observed\",\n    \"Example: 'Calculations are illustrative only — always state they do not constitute financial advice'\",\n    \"Example: 'Never recommend a specific product, provider, or course of action'\",\n    \"Return empty array [] if no compliance patterns are evident\"\n  ],\n  \"heading_style\": {\n    \"format\": \"How headings are typically phrased\",\n    \"hierarchy\": \"How heading levels are used — e.g. 'H1 is the page title. H2s are major topic breaks. H3s are sub-points within a topic. Never skip levels.'\",\n    \"examples_from_site\": [\"Copy 3-4 actual headings from the content\"]\n  },\n  \"paragraph_style\": {\n    \"typical_length\": \"How long paragraphs typically are\",\n    \"structure\": \"How paragraphs are internally organised\",\n    \"opening_patterns\": \"How paragraphs typically begin — e.g. 'Often opens with a bold claim or question, then immediately supports it.'\"\n  },\n  \"cta_style\": {\n    \"approach\": \"How calls-to-action are handled\",\n    \"verb_choices\": \"What action verbs are used — e.g. 'Launch, Calculate, Simulate, Read — never Get, Buy, Claim, Unlock'\",\n    \"examples_from_site\": [\"Copy 3-4 actual CTAs from the content\"]\n  },\n  \"terminology\": {\n    \"approach\": \"How domain-specific terms are handled\",\n    \"definition_pattern\": \"How terms are defined when introduced — e.g. 'Bold on first use, followed by a parenthetical plain-English explanation'\",\n    \"key_terms\": [\"List 8-12 domain-specific terms the site uses regularly\"]\n  },\n  \"formatting_conventions\": {\n    \"bold_usage\": \"When and how bold text is used — e.g. 'Key concepts on first mention. Never for emphasis in running text.'\",\n    \"italic_usage\": \"When and how italics are used\",\n    \"link_style\": \"How links are presented — e.g. 'Descriptive anchor text. Never raw URLs. Never click here.'\",\n    \"list_usage\": \"When lists are used vs prose — e.g. 'Lists for requirements, steps, and checklists. Never for narrative content. Always prefaced by an introductory sentence.'\"\n  },\n  \"things_to_avoid\": [\n    \"Specific patterns the site clearly avoids — extract from what is NOT in the content\",\n    \"Include at least 8-10 items, each specific and observable\",\n    \"Example: 'Urgency language (limited time, act now, don''t miss)'\",\n    \"Example: 'Exclamation marks in body copy'\",\n    \"Example: 'Vague claims without supporting data'\",\n    \"Example: 'Passive voice when describing user actions'\"\n  ],\n  \"things_to_emulate\": [\n    \"Specific patterns the site does well that should be preserved\",\n    \"Include at least 8-10 items, each specific and observable\",\n    \"Example: 'Uses the reader''s likely question as a section heading, then answers it directly'\",\n    \"Example: 'Follows every technical explanation with a concrete worked example'\",\n    \"Example: 'Ends articles with a practical checklist summarising key points'\",\n    \"Example: 'Opening sentences that challenge a common misconception'\"\n  ],\n  \"example_phrases\": {\n    \"characteristic\": [\n      \"Copy 5-8 short phrases from the content that are highly characteristic of the site voice\",\n      \"A new writer should be able to read these and immediately grasp the tone\"\n    ],\n    \"would_never_say\": [\n      \"Write 5-8 phrases the site would NEVER use based on its established tone\",\n      \"Example: 'Unlock your potential!' (too salesy)\",\n      \"Example: 'As the leading provider...' (unsubstantiated claim)\",\n      \"Example: 'Don''t miss out on...' (urgency tactic)\"\n    ]\n  }\n}\n\nRules:\n- Extract patterns from the ACTUAL content provided — do not invent or assume\n- Be specific, not generic. 'Write clearly' is useless. 'Paragraphs are 2-3 sentences, leading with the conclusion' is actionable\n- The example_phrases.characteristic must be ACTUAL text copied from the provided content\n- The example_phrases.would_never_say should be invented counter-examples that clearly violate the observed style\n- If compliance patterns exist (financial disclaimers, medical caveats, legal language), capture them in detail in compliance_rules\n- If no compliance patterns exist, return an empty compliance_rules array\n- writing_rules should have at least 10 entries\n- things_to_avoid and things_to_emulate should each have at least 8 entries\n- example_phrases.characteristic should have at least 5 entries\n- Every field in voice, sentence_style, persuasion_approach, and content_depth must be filled with specific observations, not generic writing advice"
        },
        "output_field": "content_direction_analysis",
        "next_step": "apply_plan",
        "error_step": "apply_plan",
        "description": "LLM extracts detailed writing style guide from representative content"
    }'
)
WHERE type = 'site-adoption-agent';

-- ── 4. Add content_direction_analysis to apply_plan input_fields ────────

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,apply_plan,config,input_fields}',
        '["site_record", "crawl_result", "adoption_analysis", "content_direction_analysis"]'
                     )
WHERE type = 'site-adoption-agent';

---
-- improve voice prompt
-- ============================================================================
-- Add content direction analysis to the site-adoption-agent workflow
-- ============================================================================
--
-- Adds two steps between analyze_site and apply_plan:
--   1. select_content — picks 2-3 prose-heavy pages from crawl
--   2. derive_content_direction — LLM extracts detailed writing guidelines
--
-- Uses DO $$ blocks to avoid single-quote escaping issues in the prompt.
-- ============================================================================

-- ── 1. Redirect analyze_site → select_content (was → apply_plan) ────────

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_site,next_step}',
        '"select_content"'
                     )
WHERE type = 'site-adoption-agent';

-- ── 2. Add select_content step ──────────────────────────────────────────

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,select_content}',
        '{
            "action": "select_representative_content",
            "config": {
                "max_pages": 3,
                "max_total_chars": 15000
            },
            "output_field": "representative_content",
            "next_step": "derive_content_direction",
            "error_step": "apply_plan",
            "description": "Select 2-3 prose-heavy pages from crawl for writing style analysis"
        }'
                     )
WHERE type = 'site-adoption-agent';

-- ── 3. Add derive_content_direction step ────────────────────────────────
-- Uses DO $$ block so the prompt can contain single quotes freely.

DO $outer$
DECLARE
prompt_text text;
    step_json jsonb;
BEGIN
    prompt_text := $prompt$You are a senior copywriter and brand strategist. You are analysing the actual content from a website to produce a detailed writing style guide that another writer could follow to recreate content in exactly the same voice, tone, and style.

Domain: {{if .site_record}}{{.site_record.domain}}{{end}}

Here is the actual content from representative pages on this site:

{{if .representative_content}}{{.representative_content.selected_text}}{{end}}

Study this content carefully and deeply. Pay attention to:
- How headings are phrased (questions? statements? commands? how long?)
- Sentence length, rhythm, and construction patterns
- Level of formality vs conversational tone
- How technical terms are handled (defined? assumed known? avoided? explained inline?)
- Use of first/second/third person and when each is used
- How confidence and authority are expressed without overstepping
- What the content deliberately avoids saying (no promises? no superlatives? no jargon without explanation?)
- How calls-to-action are worded (hard sell? soft? informational? what verbs?)
- Any legal, regulatory, or compliance patterns (disclaimers, caveats, qualifications)
- Use of examples, analogies, worked calculations, or concrete illustrations
- How lists, steps, and structured content are formatted
- The emotional register (reassuring? challenging? neutral? playful? dry?)
- How the site builds trust (through expertise? through transparency? through relatability?)
- What assumptions the site makes about what the reader already knows
- How deeply the content explores each topic (surface overview? deep dive? practical application?)
- How bold, italics, links, and other formatting are used
- How sections transition from one idea to the next

Produce a JSON writing style guide. Be as specific and detailed as possible — a writer reading this guide should be able to perfectly match the voice without ever seeing the original site. Respond ONLY with valid JSON, no markdown backticks.

{
  "voice": {
    "register": "Describe the overall register in one detailed sentence",
    "person": "Which grammatical person is used and when",
    "authority_level": "How the site establishes authority and credibility",
    "emotional_tone": "The emotional quality and how it makes the reader feel",
    "formality": "Where it sits on formal-casual spectrum with specific markers"
  },
  "sentence_style": {
    "average_length": "Short/medium/long and typical word count per sentence",
    "structure_patterns": "How sentences are typically constructed",
    "rhythm": "The cadence of the writing",
    "connectives": "How ideas are linked between sentences and paragraphs"
  },
  "persuasion_approach": {
    "method": "How the site convinces without selling",
    "trust_building": "How trust is established",
    "social_proof_style": "How social proof is handled"
  },
  "content_depth": {
    "thoroughness": "How deeply topics are explored",
    "assumed_knowledge": "What the reader is expected to already know",
    "explanation_pattern": "How complex ideas are introduced and explained"
  },
  "writing_rules": [
    "Each rule should be specific and actionable, not vague.",
    "Extract at least 10-15 rules from actual patterns in the content.",
    "Example: 'Technical terms must be explained in plain English on first use'",
    "Example: 'Paragraphs are 2-3 sentences maximum. One idea per paragraph.'"
  ],
  "compliance_rules": [
    "Any legal, regulatory, financial, medical, or professional compliance patterns observed",
    "Return empty array [] if no compliance patterns are evident"
  ],
  "heading_style": {
    "format": "How headings are typically phrased",
    "hierarchy": "How heading levels are used",
    "examples_from_site": ["Copy 3-4 actual headings from the content"]
  },
  "paragraph_style": {
    "typical_length": "How long paragraphs typically are",
    "structure": "How paragraphs are internally organised",
    "opening_patterns": "How paragraphs typically begin"
  },
  "cta_style": {
    "approach": "How calls-to-action are handled",
    "verb_choices": "What action verbs are used",
    "examples_from_site": ["Copy 3-4 actual CTAs from the content"]
  },
  "terminology": {
    "approach": "How domain-specific terms are handled",
    "definition_pattern": "How terms are defined when introduced",
    "key_terms": ["List 8-12 domain-specific terms the site uses regularly"]
  },
  "formatting_conventions": {
    "bold_usage": "When and how bold text is used",
    "italic_usage": "When and how italics are used",
    "link_style": "How links are presented",
    "list_usage": "When lists are used vs prose"
  },
  "things_to_avoid": [
    "Specific patterns the site clearly avoids — extract from what is NOT in the content",
    "Include at least 8-10 items, each specific and observable"
  ],
  "things_to_emulate": [
    "Specific patterns the site does well that should be preserved",
    "Include at least 8-10 items, each specific and observable"
  ],
  "example_phrases": {
    "characteristic": [
      "Copy 5-8 short phrases from the content that are highly characteristic of the site's voice",
      "A new writer should be able to read these and immediately grasp the tone"
    ],
    "would_never_say": [
      "Write 5-8 phrases the site would NEVER use based on its established tone"
    ]
  }
}

Rules:
- Extract patterns from the ACTUAL content provided — do not invent or assume
- Be specific, not generic. 'Write clearly' is useless. 'Paragraphs are 2-3 sentences, leading with the conclusion' is actionable
- The example_phrases.characteristic must be ACTUAL text copied from the provided content
- The example_phrases.would_never_say should be invented counter-examples that clearly violate the observed style
- If compliance patterns exist (financial disclaimers, medical caveats, legal language), capture them in detail
- If no compliance patterns exist, return an empty compliance_rules array
- writing_rules should have at least 10 entries
- things_to_avoid and things_to_emulate should each have at least 8 entries
- example_phrases.characteristic should have at least 5 entries
- Every field in voice, sentence_style, persuasion_approach, and content_depth must be filled with specific observations, not generic writing advice$prompt$;

    step_json := jsonb_build_object(
        'action', 'execute_llm_prompt',
        'config', jsonb_build_object(
            'ai_service', jsonb_build_object(
                'model', 'claude-sonnet-4-6',
                'provider', 'anthropic',
                'api_key_env_var', 'ANTHROPIC_API_KEY'
            ),
            'max_tokens', 6000,
            'temperature', 0.2,
            'input_fields', '["site_record", "representative_content", "adoption_analysis"]'::jsonb,
            'prompt_template', prompt_text
        ),
        'output_field', 'content_direction_analysis',
        'next_step', 'apply_plan',
        'error_step', 'apply_plan',
        'description', 'LLM extracts detailed writing style guide from representative content'
    );

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,derive_content_direction}',
        step_json
                     )
WHERE type = 'site-adoption-agent';

RAISE NOTICE 'derive_content_direction step added successfully';
END $outer$;

-- ── 4. Add content_direction_analysis to apply_plan input_fields ────────

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,apply_plan,config,input_fields}',
        '["site_record", "crawl_result", "adoption_analysis", "content_direction_analysis"]'
                     )
WHERE type = 'site-adoption-agent';

---

-- Add classify_archetype step to site-adoption-agent workflow
--
-- This adds a new LLM step between analyze_site and select_content.
-- analyze_site produces structural classification (pages, sections, identity).
-- classify_archetype produces the character/design/purpose classification
-- that the improvement loop uses to avoid regressing the site.
--
-- The output is stored in collected_data.site_archetype_analysis,
-- which apply_adoption_plan reads and writes as a site_spec (aspect: site_archetype).

DO $do$
DECLARE
v_config jsonb;
  v_steps jsonb;
BEGIN
SELECT default_config INTO v_config
FROM agent_definitions
WHERE type = 'site-adoption-agent' AND deleted_at IS NULL;

v_steps := v_config->'workflow'->'steps';

  -- 1. Update analyze_site to point to classify_archetype instead of select_content
  v_steps := jsonb_set(
    v_steps,
    '{analyze_site,next_step}',
    '"classify_archetype"'
  );

  -- 2. Add the classify_archetype step
  v_steps := jsonb_set(
    v_steps,
    '{classify_archetype}',
    '{
      "action": "execute_llm_prompt",
      "config": {
        "model": "claude-sonnet-4-6",
        "ai_service": {
          "model": "claude-sonnet-4-6",
          "provider": "anthropic",
          "api_key_env_var": "ANTHROPIC_API_KEY"
        },
        "max_tokens": 4000,
        "temperature": 0.2,
        "input_fields": ["site_record", "formatted_crawl", "adoption_analysis"],
        "system_prompt": "You are analyzing a website to produce a site archetype classification. Your job is to describe what this site IS — not what it should become. This is a snapshot, not an aspiration. Think about what you would tell someone who asked \"what kind of site is this?\" after you spent 5 minutes looking at it.",
        "prompt_template": "Domain: {{.site_record.domain}}\n\n== CRAWL DATA ==\n\n{{.formatted_crawl.research_text}}\n\n== STRUCTURAL ANALYSIS (from previous step) ==\n\nIdentity: {{.adoption_analysis.identity}}\nDesign: {{.adoption_analysis.design}}\nInteractive features: {{.adoption_analysis.interactive_features}}\nPages: {{.adoption_analysis.pages}}\n\n== INSTRUCTIONS ==\n\nProduce a JSON object classifying this site. Every field is required. Be specific and honest — \"none\" is a valid answer. Do not invent features that are not in the crawl.\n\n{\n  \"label\": \"A 2-4 word human-readable classification (e.g. developer utility platform, industrial product showcase, vertical listing aggregator, content marketing hub)\",\n  \"industry\": \"the industry or vertical this site serves\",\n\n  \"character\": {\n    \"feel\": \"3-5 descriptive words for the overall impression\",\n    \"polish\": \"how finished does it look — rough/moderate/polished, with a brief note on why\",\n    \"budget_impression\": \"what budget level does it suggest — indie/small business/mid-market/enterprise\",\n    \"age_impression\": \"does it feel new or established — and what gives that impression\",\n    \"commercial_intent\": \"how is it trying to make money, or is it not — describe what you see\",\n    \"density\": \"how much content per screen — sparse/low/medium/high/dense\"\n  },\n\n  \"design\": {\n    \"palette_mood\": \"describe the colour approach — dark/light, accent colours, warmth/coolness\",\n    \"layout_approach\": \"what layout patterns dominate — grids, cards, sidebars, full-width, split panels\",\n    \"typography_feel\": \"what do the fonts suggest — corporate, technical, playful, editorial\",\n    \"imagery\": \"what visual assets are used — photos, illustrations, icons, none, screenshots, charts\",\n    \"animation\": \"how much motion — none, hover states, transitions, full animation\",\n    \"responsive\": \"does the crawl suggest mobile support — and how\"\n  },\n\n  \"content\": {\n    \"primary_type\": \"what is the main content — prose, products, tools, media, listings, data\",\n    \"secondary_type\": \"what else is there, if anything\",\n    \"voice\": \"how does the site talk — formal, casual, technical, sales-oriented, educational\",\n    \"media\": \"what media types are present — text only, images, video, audio, interactive elements\"\n  },\n\n  \"purpose\": [\"array of what the site exists to do\"],\n  \"content_model\": [\"array of content types present\"],\n  \"interaction_patterns\": [\"array of what users DO on this site\"],\n  \"revenue_model\": \"none | advertising | affiliate | e-commerce | subscription | lead-generation | freemium | mixed (describe)\",\n  \"visual_character\": [\"array of style tags\"],\n  \"audience\": [\"array of who this is for\"],\n\n  \"structure\": {\n    \"index_layout\": \"what the homepage looks like\",\n    \"listing_style\": \"how collections of items are presented\",\n    \"navigation\": \"how users move around\",\n    \"page_depth\": \"how many clicks to real content — shallow (1-2), medium (2-3), deep (3+)\"\n  },\n\n  \"constraints\": [\"array of things the improvement loop should NEVER do to this site — inferred from what the site is\"]\n}\n\nRespond with ONLY the JSON object. No explanation, no markdown backticks."
      },
      "next_step": "select_content",
      "description": "LLM classifies site archetype — character, design patterns, purpose, constraints",
      "output_field": "site_archetype_analysis"
    }'::jsonb
  );

  -- 3. Write back
  v_config := jsonb_set(v_config, '{workflow,steps}', v_steps);

UPDATE agent_definitions
SET default_config = v_config,
    updated_at = NOW()
WHERE type = 'site-adoption-agent' AND deleted_at IS NULL;

RAISE NOTICE 'site-adoption-agent: added classify_archetype step (analyze_site → classify_archetype → select_content)';
END $do$;

-- Verify the step chain
SELECT
    step_key,
    step_value->>'action' as action,
    step_value->>'next_step' as next_step,
    step_value->>'output_field' as output_field
FROM agent_definitions,
    jsonb_each(default_config->'workflow'->'steps') as s(step_key, step_value)
WHERE type = 'site-adoption-agent' AND deleted_at IS NULL
ORDER BY step_key;
--
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,crawl_site,config}',
        '{
          "url_field": "input_data.url",
          "limit": 50,
          "maxDiscoveryDepth": 4,
          "scrape_config": {
            "formats": ["markdown", "rawHtml"],
            "only_main_content": false
          }
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'site-adoption-agent'
  AND is_active = true;

--

UPDATE agent_definitions SET default_config = jsonb_set(
        default_config, '{workflow,steps,crawl_site,config}',
        '{"url_field":"input_data.url","scrape_config":{"formats":["markdown","rawHtml"],"only_main_content":false,"limit":50,"max_discovery_depth":4}}'::jsonb
                                              ), updated_at = NOW()
WHERE type = 'site-adoption-agent' AND is_active = true;
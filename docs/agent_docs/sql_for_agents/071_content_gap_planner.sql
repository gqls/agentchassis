-- a bit like blog content planner but more generic

-- ============================================================================
-- content-gap-planner agent
--
-- Receives: needs_content_planning work items from write_audit_findings
-- These are gaps the audit identified but couldn't classify as fixes to
-- existing pages — new pages needed, site-wide content issues, structural gaps.
--
-- This agent:
--   1. Loads site specs (identity, content_direction)
--   2. Loads existing pages (to avoid duplicating)
--   3. Loads the batch of needs_content_planning items for this site
--   4. Asks LLM: given these gaps and what already exists, what should we build?
--   5. For each recommendation:
--      - If "add section to existing page" → create content_rewrite item for page-build-handler
--      - If "create new page" → create page record + needs_content_page item for page-build-handler
--      - If "update spec" → create needs_spec_update item
--      - If "not actionable" → mark original item as wont_fix
--
-- The LLM here is the PLANNER, not the auditor. It makes architectural
-- decisions about what pages to create and what sections they need.
-- The auditor only observed that something was missing.
--
-- No new Go code beyond what exists. Uses:
--   ensure_site_record, read_site_spec, query_database,
--   execute_llm_prompt, create_blog_posts (reusable for any page creation),
--   create_work_item, complete_workflow
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, agent_category, status,
    image_repository, image_tag, category,
    default_config, input_contract, output_contract, domain_tags
) VALUES (
             'content-gap-planner',
             'Content Gap Planner',
             'Plans how to fill content gaps identified by audits. Reads gap descriptions and site context, decides whether to create new pages, add sections to existing pages, or update specs. Creates actionable work items for page-build-handler.',
             'specialist', 'active',
             'docker.io/aqls/agent-chassis', 'v1.0.842', 'specialist',
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "processing_mode": "orchestrator",
                     "timeout_seconds": 300,
                     "steps": {
                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "load_specs",
                             "output_field": "site_record"
                         },
                         "load_specs": {
                             "action": "read_site_spec",
                             "config": { "site_id": "site_record.site_id" },
                             "next_step": "load_existing_pages",
                             "error_step": "load_existing_pages",
                             "output_field": "site_specs"
                         },
                         "load_existing_pages": {
                             "action": "load_page_record",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "load_all": true
                             },
                             "next_step": "load_available_components",
                             "error_step": "load_available_components",
                             "description": "Load all existing pages for context",
                             "output_field": "existing_pages"
                         },
                         "load_available_components": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT name, display_name, \"function\", category FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true ORDER BY category, name",
                                 "output_format": "array"
                             },
                             "next_step": "plan_gaps",
                             "error_step": "plan_gaps",
                             "output_field": "available_components"
                         },
                         "plan_gaps": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-6",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_record", "site_specs", "existing_pages", "input_data", "available_components"],
                                 "output_format": "json",
                                 "prompt_template": "You are a site architect planning how to fill content gaps.\n\n## Site\nDomain: {{.site_record.domain}}\n{{if .site_specs.specs.identity.company_name}}Company: {{.site_specs.specs.identity.company_name}}{{end}}\n{{if .site_specs.specs.identity.industry}}Industry: {{.site_specs.specs.identity.industry}}{{end}}\n{{if .site_specs.specs.identity.target_audience}}Audience: {{.site_specs.specs.identity.target_audience}}{{end}}\n\n{{if .site_specs.specs.content_direction}}## Content Direction\nVoice: {{.site_specs.specs.content_direction.voice}}\nEmphasis: {{.site_specs.specs.content_direction.emphasis}}{{end}}\n\n## Existing Pages\n{{if .existing_pages}}{{range .existing_pages}}- {{.name}} ({{.page_type}}): {{.title}}\n{{end}}{{else}}No pages loaded.{{end}}\n\n## Available Section Components\n{{if .available_components}}{{range .available_components}}- {{.name}} ({{.function}})\n{{end}}{{end}}\n\n## Content Gap to Address\nDescription: {{.input_data.spec.description}}\n{{if .input_data.spec.suggestion}}Audit suggestion: {{.input_data.spec.suggestion}}{{end}}\nOriginal category: {{.input_data.spec.category}}\n\n## Your Task\nDecide how to address this content gap. Choose ONE of these approaches:\n\n### A) Add to existing page\nIf the gap can be filled by adding a section to an existing page (e.g. add FAQ section to the services page), recommend this. Only use section components from the Available list above.\n\n### B) Create new page\nIf the gap needs its own page (e.g. a dedicated FAQ page, a pricing page), recommend creating one. Specify the page name, title, purpose, and sections from the Available list.\n\n### C) Update site spec\nIf the gap is about missing metadata (target audience, tone definition), recommend a spec update.\n\n### D) Not actionable\nIf the gap is too vague, already covered, or not worth addressing, say so.\n\nReturn ONLY valid JSON:\n{\n  \"approach\": \"add_to_page\" | \"new_page\" | \"update_spec\" | \"not_actionable\",\n  \"reasoning\": \"Why this approach\",\n  \"add_to_page\": {\n    \"page_name\": \"existing page name\",\n    \"add_sections\": [\"faq-section\", \"call-to-action\"],\n    \"content_guidance\": \"What the new sections should cover\"\n  },\n  \"new_page\": {\n    \"name\": \"kebab-case-name\",\n    \"title\": \"Page Title | Company Name\",\n    \"page_type\": \"content\",\n    \"purpose\": \"What this page covers and why\",\n    \"sections\": [\"hero\", \"generic-text-block\", \"call-to-action\"],\n    \"nav_label\": \"Nav Label\",\n    \"in_header\": true,\n    \"in_footer\": true\n  },\n  \"update_spec\": {\n    \"aspect\": \"identity\",\n    \"field\": \"target_audience\",\n    \"suggested_value\": \"UK SMEs looking for practical AI solutions\"\n  },\n  \"not_actionable\": {\n    \"reason\": \"Why this gap doesnt need addressing\"\n  }\n}"
                    },
             "next_step": "apply_plan",
             "error_step": "complete",
             "output_field": "gap_plan"
                },
             "apply_plan": {
                    "action": "apply_gap_plan",
             "config": {
                        "site_id": "site_record.site_id",
             "plan": "gap_plan.result",
             "work_item_id": "input_data.work_item_id",
             "domain": "site_record.domain"
                    },
             "next_step": "complete",
             "error_step": "complete",
             "description": "Execute the plan — create pages, work items, or spec updates",
             "output_field": "plan_applied"
                },
             "complete": {
                    "action": "complete_workflow",
             "config": { "output_fields": ["gap_plan", "plan_applied"] }
                }
            }
        }
    }'::jsonb,
    '{"required": ["site_id", "domain"], "optional": ["work_item_id", "spec"]}'::jsonb,
    '{"produces": {"gap_plan": "LLM plan for addressing the gap", "plan_applied": "result of executing the plan"}}'::jsonb,
    '["content", "planning", "audit"]'::jsonb
) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    updated_at = NOW();

-- Verify
SELECT type, display_name, status
FROM agent_definitions
WHERE type = 'content-gap-planner' AND deleted_at IS NULL;

--

-- one page at a time

-- Fix content-gap-planner: use load_site_pages instead of load_page_record
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_pages}',
        '{
            "action": "load_site_pages",
            "config": { "site_id": "site_record.site_id" },
            "next_step": "load_available_components",
            "error_step": "load_available_components",
            "description": "Load all existing pages for context",
            "output_field": "existing_pages"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'content-gap-planner' AND deleted_at IS NULL;

-- Verify
SELECT default_config->'workflow'->'steps'->'load_existing_pages'->>'action' as action
FROM agent_definitions WHERE type = 'content-gap-planner' AND deleted_at IS NULL;

-- Fix the existing_pages range in content-gap-planner template
UPDATE agent_definitions
SET default_config = REPLACE(
        default_config::text,
        '{{if .existing_pages}}{{range .existing_pages}}- {{.name}} ({{.page_type}}): {{.title}}\n{{end}}{{else}}No pages loaded.{{end}}',
        '{{if .existing_pages.pages}}{{range .existing_pages.pages}}- {{.name}} ({{.page_type}}): {{.title}}\n{{end}}{{else}}No pages loaded.{{end}}'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'content-gap-planner' AND deleted_at IS NULL;

-- Verify the fix took
SELECT default_config::text LIKE '%existing_pages.pages%' as fixed
FROM agent_definitions
WHERE type = 'content-gap-planner' AND deleted_at IS NULL;


-- Fix existing_pages → existing_pages.pages in the range template
UPDATE agent_definitions
SET default_config = REPLACE(
        default_config::text,
        '{{if .existing_pages}}{{range .existing_pages}}-',
        '{{if .existing_pages.pages}}{{range .existing_pages.pages}}-'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'content-gap-planner' AND deleted_at IS NULL;
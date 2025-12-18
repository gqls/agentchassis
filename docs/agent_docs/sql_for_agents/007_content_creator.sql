-- ============================================================================
-- VERSION 2 AGENTS - Unified Site Builder Architecture
-- ============================================================================
-- These are v2 agents that work alongside existing v1 agents.
-- v1 agents continue to work as before.
-- v2 agents use the new pages/components structure.
--
-- To use v2: reference agent_type with version, or update workflows to use v2
-- ============================================================================



-- ============================================================================
-- 2. CONTENT-CREATOR-V2 - Understands component-based pages
-- ============================================================================
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
    image_tag,
    resources,
    topics,
    health_config,
    version,
    input_contract,
    output_contract
)
SELECT
    gen_random_uuid(),
    'content-creator',
    'Content Creator V2',
    'Creates content for component-based pages (v2 - receives current_page object)',
    category,
    jsonb_build_object(
            'workflow', jsonb_build_object(
            'start_step', 'create_content',
            'steps', jsonb_build_object(
                    'create_content', jsonb_build_object(
                            'action', 'execute_llm_prompt',
                            'config', jsonb_build_object(
                                    'ai_service', jsonb_build_object(
                                            'model', 'claude-sonnet-4-5-20250514',
                                            'provider', 'anthropic',
                                            'api_key_env_var', 'ANTHROPIC_API_KEY'
                                                  ),
                                    'output_type', 'json',
                                    'input_fields', jsonb_build_array('input_data', 'current_page', 'page_plan'),
                                    'prompt_template', 'You are a professional copywriter creating content for a website page.

CONTEXT:
Domain: {{.input_data.domain}}
Objective: {{.input_data.objective}}
Marketing Model: {{.input_data.model}}

PAGE TO CREATE:
Name: {{.current_page.name}}
Title: {{.current_page.title}}
Purpose: {{.current_page.purpose}}
Components: {{.current_page.components}}

Create content for EACH component listed above.

Component guidelines:
- hero-*: headline (5-8 words), subheadline (15-25 words), cta_text
- services-*: items array with name, description, icon_suggestion
- features-*: items array with name, benefit, detail
- testimonials-*: items array with quote, name, title, company
- team-*: items array with name, role, bio
- pricing-*: tiers array with name, price, features, cta
- faq-*: items array with question, answer
- cta-*: headline, supporting_text, button_text
- contact-*: intro, email, phone, address
- about-*: paragraphs array, key_points
- footer-*: company_name, tagline, link_groups, copyright

OUTPUT FORMAT (valid JSON):
{
  "page_name": "index",
  "hero": {"headline": "...", "subheadline": "...", "cta_text": "...", "cta_url": "#"},
  "sections": [
    {"type": "component-type", "content": {...}}
  ],
  "meta": {"page_title": "...", "meta_description": "..."},
  "footer": {"company_name": "...", "tagline": "...", "copyright": "..."}
}'
                                      ),
                            'output_field', 'content_result',
                            'next_step', 'complete',
                            'description', 'Create content for page components'
                                      ),
                    'complete', jsonb_build_object(
                            'action', 'complete_workflow',
                            'description', 'Return content'
                                )
                     )
                        ),
            'processing_mode', 'task',
            'timeout_seconds', 180
    ),
    true,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    2,  -- VERSION 2
    '{"required": ["input_data", "current_page"], "expects": {"current_page": "object with name, components"}}'::jsonb,
    '{"produces": "page_content", "format": {"type": "object", "properties": {"hero": "object", "sections": "array"}}}'::jsonb
FROM agent_definitions
WHERE type = 'content-creator' AND version = 1
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       display_name = EXCLUDED.display_name,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();



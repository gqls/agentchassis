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
-- 1. CHIEF-STRATEGIST-V2 - Outputs pages with components
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
    'chief-strategist',  -- Same type, different version
    'Chief Strategist V2',
    'Site planner that outputs pages with component types (v2 - unified architecture)',
    category,
    jsonb_build_object(
            'workflow', jsonb_build_object(
            'start_step', 'generate_build_plan',
            'steps', jsonb_build_object(
                    'generate_build_plan', jsonb_build_object(
                            'action', 'execute_llm_prompt',
                            'config', jsonb_build_object(
                                    'ai_service', jsonb_build_object(
                                            'model', 'claude-haiku-4-5-20251001',
                                            'provider', 'anthropic',
                                            'api_key_env_var', 'ANTHROPIC_API_KEY'
                                                  ),
                                    'output_type', 'json',
                                    'input_fields', jsonb_build_array('input_data'),
                                    'prompt_template', 'You are a Site Planner designing the structure for {{.input_data.domain}}.

OBJECTIVE: {{.input_data.objective}}
MARKETING MODEL: {{.input_data.model}}

STEP 1: Determine the best site structure for this objective.

Site Type Guidelines:
- LANDING (1 page, 5-8 components): Product launches, lead gen, focused campaigns
- CORPORATE (4-6 pages): Professional services, consulting, established businesses
- PORTFOLIO (3-5 pages): Creatives, agencies, case study focused
- ECOMMERCE (2-4 pages): Product sales, shopping focused

STEP 2: Plan each page with specific components.

Available component types:
- hero-centered, hero-split, hero-video
- services-grid, services-list
- features-cards, features-comparison
- testimonials-carousel, testimonials-grid
- team-grid, pricing-tiers, faq-accordion
- cta-banner, cta-split
- contact-form, contact-simple
- about-story, about-values
- footer-standard

OUTPUT FORMAT (valid JSON only):
{
  "site_type": "landing|corporate|portfolio|ecommerce",
  "reasoning": "Why this structure fits the objective",
  "theme_suggestion": "professional|bold|minimal|creative",
  "pages": [
    {
      "name": "index",
      "title": "Page Title | Brand",
      "purpose": "What this page achieves",
      "components": [
        {"type": "hero-centered", "priority": "high"},
        {"type": "services-grid", "priority": "high"}
      ],
      "meta_description": "SEO description"
    }
  ],
  "global": {
    "navigation": ["Home", "About", "Services", "Contact"],
    "brand_tone": "professional|friendly|bold|technical"
  }
}'
                                      ),
                            'output_field', 'build_plan_raw',
                            'next_step', 'parse_plan',
                            'description', 'Generate site plan with pages and components'
                                           ),
                    'parse_plan', jsonb_build_object(
                            'action', 'parse_json_field',
                            'config', jsonb_build_object(
                                    'source_field', 'build_plan_raw'
                                      ),
                            'output_field', 'plan_data',
                            'next_step', 'complete',
                            'description', 'Parse JSON plan'
                                  ),
                    'complete', jsonb_build_object(
                            'action', 'complete_workflow',
                            'description', 'Return parsed plan'
                                )
                     )
                        ),
            'processing_mode', 'task',
            'timeout_seconds', 120
    ),
    true,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    2,  -- VERSION 2
    '{"required": ["input_data"], "expects": {"input_data.domain": "string", "input_data.objective": "string"}}'::jsonb,
    '{"produces": "plan_data", "format": {"type": "object", "properties": {"site_type": "string", "pages": "array"}}}'::jsonb
FROM agent_definitions
WHERE type = 'chief-strategist' AND version = 1
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       display_name = EXCLUDED.display_name,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();



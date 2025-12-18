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


=============

-- ============================================================================
-- CHIEF-STRATEGIST V2 - Expanded Site Types with Pages/Components Structure
-- ============================================================================
-- Site types supported:
-- - LANDING: Product launches, lead gen, focused campaigns (1 page)
-- - CORPORATE: Professional services, consulting (4-6 pages)
-- - PORTFOLIO: Creatives, agencies, case studies (3-5 pages)
-- - ECOMMERCE: Product sales, shopping (2-4 pages)
-- - CONTENT: News, blogs, recipes, gossip - content/traffic driven (4-8 pages)
-- - TOOLS: Calculators, utilities - feature/tool driven (2-5 pages)
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_build_plan,config,prompt_template}',
        '"You are a Site Planner designing the structure for {{.input_data.domain}}.

OBJECTIVE: {{.input_data.objective}}
MARKETING MODEL: {{.input_data.model}}

STEP 1: Determine the best site type for this objective.

Site Type Guidelines:

LANDING (1 page, 5-8 components)
- Product launches, lead generation, focused campaigns
- Single conversion goal, minimal navigation
- Revenue: Direct sales, lead capture

CORPORATE (4-6 pages)
- Professional services, consulting, established businesses
- Trust-building, multiple service areas
- Revenue: Service contracts, B2B relationships

PORTFOLIO (3-5 pages)
- Creatives, agencies, freelancers
- Case study focused, visual showcase
- Revenue: Project work, client acquisition

ECOMMERCE (2-4 pages + product structure)
- Product sales, shopping focused
- Category browsing, cart functionality
- Revenue: Direct product sales

CONTENT (4-8 pages + article structure)
- News sites, blogs, recipes, celebrity gossip, lifestyle
- Content-driven traffic, regular publishing
- SEO focused, high page count potential
- Revenue: Advertising, affiliate links, sponsored content

TOOLS (2-5 pages + tool interfaces)
- Calculators (mortgage, tiles, BMI, etc.), converters, utilities
- Feature/functionality driven, practical value
- User retention through bookmarking
- Revenue: Advertising, affiliate referrals, premium features

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
- blog-grid, blog-featured, article-layout
- recipe-card, recipe-grid, recipe-detail
- tool-calculator, tool-converter, tool-interface
- ad-banner, ad-sidebar, affiliate-showcase
- category-grid, content-feed, search-bar
- social-share, comments-section, newsletter-signup

OUTPUT FORMAT (valid JSON only):
{
  "site_type": "landing|corporate|portfolio|ecommerce|content|tools",
  "reasoning": "Why this structure fits the objective",
  "theme_suggestion": "professional|bold|minimal|creative|editorial|functional",
  "revenue_model": "direct_sales|services|advertising|affiliate|freemium",
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
    "brand_tone": "professional|friendly|bold|technical|editorial|practical"
  }
}

EXAMPLES BY SITE TYPE:

CONTENT site (news/gossip/recipes):
{
  "site_type": "content",
  "pages": [
    {"name": "index", "components": [{"type": "hero-split"}, {"type": "blog-featured"}, {"type": "content-feed"}, {"type": "ad-sidebar"}, {"type": "newsletter-signup"}]},
    {"name": "categories", "components": [{"type": "category-grid"}, {"type": "search-bar"}]},
    {"name": "article", "components": [{"type": "article-layout"}, {"type": "social-share"}, {"type": "ad-banner"}, {"type": "comments-section"}]},
    {"name": "about", "components": [{"type": "about-story"}, {"type": "team-grid"}]},
    {"name": "contact", "components": [{"type": "contact-form"}]}
  ]
}

TOOLS site (calculators/utilities):
{
  "site_type": "tools",
  "pages": [
    {"name": "index", "components": [{"type": "hero-centered"}, {"type": "tool-calculator"}, {"type": "features-cards"}, {"type": "ad-banner"}]},
    {"name": "how-it-works", "components": [{"type": "features-comparison"}, {"type": "faq-accordion"}]},
    {"name": "related-tools", "components": [{"type": "services-grid"}, {"type": "affiliate-showcase"}]},
    {"name": "contact", "components": [{"type": "contact-simple"}]}
  ]
}"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'chief-strategist'
  AND is_active = true;


-- ============================================================================
-- VERIFICATION
-- ============================================================================
SELECT
    type,
    version,
    CASE
        WHEN type = 'chief-strategist' THEN
            'Updated prompt with CONTENT and TOOLS site types'
        WHEN type = 'multipage-website-builder' THEN
            default_config->'workflow'->'steps'->'generate_pages_loop'->'config'->>'iterate_over'
        END as status
FROM agent_definitions
WHERE type IN ('chief-strategist', 'multipage-website-builder')
  AND is_active = true;


-- ============================================================================
-- CHIEF-STRATEGIST V2 - Expanded Site Types with Pages/Components Structure
-- ============================================================================
-- Newlines escaped as \n for valid JSON

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_build_plan,config,prompt_template}',
        '"You are a Site Planner designing the structure for {{.input_data.domain}}.\n\nOBJECTIVE: {{.input_data.objective}}\nMARKETING MODEL: {{.input_data.model}}\n\nSTEP 1: Determine the best site type for this objective.\n\nSite Type Guidelines:\n\nLANDING (1 page, 5-8 components)\n- Product launches, lead generation, focused campaigns\n- Single conversion goal, minimal navigation\n- Revenue: Direct sales, lead capture\n\nCORPORATE (4-6 pages)\n- Professional services, consulting, established businesses\n- Trust-building, multiple service areas\n- Revenue: Service contracts, B2B relationships\n\nPORTFOLIO (3-5 pages)\n- Creatives, agencies, freelancers\n- Case study focused, visual showcase\n- Revenue: Project work, client acquisition\n\nECOMMERCE (2-4 pages + product structure)\n- Product sales, shopping focused\n- Category browsing, cart functionality\n- Revenue: Direct product sales\n\nCONTENT (4-8 pages + article structure)\n- News sites, blogs, recipes, celebrity gossip, lifestyle\n- Content-driven traffic, regular publishing\n- SEO focused, high page count potential\n- Revenue: Advertising, affiliate links, sponsored content\n\nTOOLS (2-5 pages + tool interfaces)\n- Calculators (mortgage, tiles, BMI, etc.), converters, utilities\n- Feature/functionality driven, practical value\n- User retention through bookmarking\n- Revenue: Advertising, affiliate referrals, premium features\n\nSTEP 2: Plan each page with specific components.\n\nAvailable component types:\n- hero-centered, hero-split, hero-video\n- services-grid, services-list\n- features-cards, features-comparison\n- testimonials-carousel, testimonials-grid\n- team-grid, pricing-tiers, faq-accordion\n- cta-banner, cta-split\n- contact-form, contact-simple\n- about-story, about-values\n- footer-standard\n- blog-grid, blog-featured, article-layout\n- recipe-card, recipe-grid, recipe-detail\n- tool-calculator, tool-converter, tool-interface\n- ad-banner, ad-sidebar, affiliate-showcase\n- category-grid, content-feed, search-bar\n- social-share, comments-section, newsletter-signup\n\nOUTPUT FORMAT (valid JSON only):\n{\n  \"site_type\": \"landing|corporate|portfolio|ecommerce|content|tools\",\n  \"reasoning\": \"Why this structure fits the objective\",\n  \"theme_suggestion\": \"professional|bold|minimal|creative|editorial|functional\",\n  \"revenue_model\": \"direct_sales|services|advertising|affiliate|freemium\",\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Brand\",\n      \"purpose\": \"What this page achieves\",\n      \"components\": [\n        {\"type\": \"hero-centered\", \"priority\": \"high\"},\n        {\"type\": \"services-grid\", \"priority\": \"high\"}\n      ],\n      \"meta_description\": \"SEO description\"\n    }\n  ],\n  \"global\": {\n    \"navigation\": [\"Home\", \"About\", \"Services\", \"Contact\"],\n    \"brand_tone\": \"professional|friendly|bold|technical|editorial|practical\"\n  }\n}\n\nEXAMPLES BY SITE TYPE:\n\nCONTENT site (news/gossip/recipes):\n{\n  \"site_type\": \"content\",\n  \"pages\": [\n    {\"name\": \"index\", \"components\": [{\"type\": \"hero-split\"}, {\"type\": \"blog-featured\"}, {\"type\": \"content-feed\"}, {\"type\": \"ad-sidebar\"}, {\"type\": \"newsletter-signup\"}]},\n    {\"name\": \"categories\", \"components\": [{\"type\": \"category-grid\"}, {\"type\": \"search-bar\"}]},\n    {\"name\": \"article\", \"components\": [{\"type\": \"article-layout\"}, {\"type\": \"social-share\"}, {\"type\": \"ad-banner\"}, {\"type\": \"comments-section\"}]},\n    {\"name\": \"about\", \"components\": [{\"type\": \"about-story\"}, {\"type\": \"team-grid\"}]},\n    {\"name\": \"contact\", \"components\": [{\"type\": \"contact-form\"}]}\n  ]\n}\n\nTOOLS site (calculators/utilities):\n{\n  \"site_type\": \"tools\",\n  \"pages\": [\n    {\"name\": \"index\", \"components\": [{\"type\": \"hero-centered\"}, {\"type\": \"tool-calculator\"}, {\"type\": \"features-cards\"}, {\"type\": \"ad-banner\"}]},\n    {\"name\": \"how-it-works\", \"components\": [{\"type\": \"features-comparison\"}, {\"type\": \"faq-accordion\"}]},\n    {\"name\": \"related-tools\", \"components\": [{\"type\": \"services-grid\"}, {\"type\": \"affiliate-showcase\"}]},\n    {\"name\": \"contact\", \"components\": [{\"type\": \"contact-simple\"}]}\n  ]\n}"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'chief-strategist'
  AND is_active = true;
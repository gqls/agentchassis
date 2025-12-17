

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,classify_site,config,prompt_template}',
            '"Classify this website project and recommend the appropriate builder.\n\nInput:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n\nAvailable Builders:\n{{range .available_builders.agents}}- {{.type}}: {{.description}}\n{{end}}\n\nClassify the site into ONE of these types based on the objective:\n\n**landing** - Conversion-focused single-purpose sites:\n- Product/service sales pages, SaaS landing pages\n- Lead generation, signups, app downloads\n- Event registration, clear single CTA goal\n\n**content** - Publishing/content sites:\n- News, blogs, magazines, articles\n- Content aggregation, SEO/traffic focused\n- Category navigation, archives\n\n**portfolio** - Showcase/portfolio sites:\n- Creative portfolios, agencies, case studies\n- Visual/image heavy, project galleries\n\n**brochure** - Multi-page business sites:\n- Company websites with About, Services, Team, Contact\n- Informational focus\n\nAnalyze the domain name and stated objective to determine the best fit.\n\nReturn ONLY valid JSON:\n{\n  \"site_type\": \"landing|content|portfolio|brochure\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"Brief explanation of classification\",\n  \"recommended_builder\": \"<exact type from Available Builders list>\",\n  \"detected_industry\": \"Industry/niche if detectable\",\n  \"detected_signals\": [\"Signal 1\", \"Signal 2\"]\n}"'::jsonb
                     )
WHERE type = 'site-classifier';

{
  "workflow": {
    "steps": {
      "complete": {
        "action": "complete_workflow",
        "description": "Return classification result"
      },
      "classify_site": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {
            "model": "claude-haiku-4-5-20251001",
            "provider": "anthropic",
            "max_tokens": 1500,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "input_fields": [
            "input_data",
            "available_builders"
          ],
          "output_field": "classification_result",
          "prompt_template": "Classify this website project and recommend the appropriate builder.\n\nInput:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n\nAvailable Builders:\n{{range .available_builders.agents}}- {{.type}}: {{.description}}\n{{end}}\n\nClassify the site into ONE of these types based on the objective:\n\n**landing** - Conversion-focused single-purpose sites:\n- Product/service sales pages, SaaS landing pages\n- Lead generation, signups, app downloads\n- Event registration, clear single CTA goal\n\n**content** - Publishing/content sites:\n- News, blogs, magazines, articles\n- Content aggregation, SEO/traffic focused\n- Category navigation, archives\n\n**portfolio** - Showcase/portfolio sites:\n- Creative portfolios, agencies, case studies\n- Visual/image heavy, project galleries\n\n**brochure** - Multi-page business sites:\n- Company websites with About, Services, Team, Contact\n- Informational focus\n\nAnalyze the domain name and stated objective to determine the best fit.\n\nReturn ONLY valid JSON:\n{\n  \"site_type\": \"landing|content|portfolio|brochure\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"Brief explanation of classification\",\n  \"recommended_builder\": \"<exact type from Available Builders list>\",\n  \"detected_industry\": \"Industry/niche if detectable\",\n  \"detected_signals\": [\"Signal 1\", \"Signal 2\"]\n}"
        },
        "next_step": "complete"
      }
    },
    "start_step": "classify_site"
  },
  "processing_mode": "task",
  "timeout_seconds": 30
}


==

-- Fix site-classifier prompt template to use flattened field names
-- The extractDataForAiAgent function flattens input_data to root level

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_site,config,prompt_template}',
        to_jsonb('Classify this website project and recommend the appropriate builder.

Input:
- Domain: {{.domain}}
- Objective: {{.objective}}

Available Builders:
{{range .available_builders.agents}}- {{.type}}: {{.description}}
{{end}}

Classify the site into ONE of these types based on the objective:

**landing** - Conversion-focused single-purpose sites:
- Product/service sales pages, SaaS landing pages
- Lead generation, signups, app downloads
- Event registration, clear single CTA goal

**content** - Publishing/content sites:
- News, blogs, magazines, articles
- Content aggregation, SEO/traffic focused
- Category navigation, archives

**portfolio** - Showcase/portfolio sites:
- Creative portfolios, agencies, case studies
- Visual/image heavy, project galleries

**brochure** - Multi-page business sites:
- Corporate sites, general business presence
- Service providers, consultants, professional services
- About/Services/Contact structure

Return ONLY valid JSON with this structure:
{
  "site_type": "landing|content|portfolio|brochure",
  "confidence": 0.0-1.0,
  "reasoning": "brief explanation",
  "recommended_builder": "builder-type",
  "detected_industry": "industry name",
  "detected_signals": ["signal1", "signal2", ...]
}')
                     )
WHERE type = 'site-classifier'
  AND is_active = true;

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps'->'classify_site'->'config'->>'prompt_template' as prompt_template
FROM agent_definitions
WHERE type = 'site-classifier'
  AND is_active = true;

==
fix jsonb
-- Fix site-classifier prompt template to use flattened field names
-- The extractDataForAiAgent function flattens input_data to root level

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_site,config,prompt_template}',
        '"Classify this website project and recommend the appropriate builder.\n\nInput:\n- Domain: {{.domain}}\n- Objective: {{.objective}}\n\nAvailable Builders:\n{{range .available_builders.agents}}- {{.type}}: {{.description}}\n{{end}}\n\nClassify the site into ONE of these types based on the objective:\n\n**landing** - Conversion-focused single-purpose sites:\n- Product/service sales pages, SaaS landing pages\n- Lead generation, signups, app downloads\n- Event registration, clear single CTA goal\n\n**content** - Publishing/content sites:\n- News, blogs, magazines, articles\n- Content aggregation, SEO/traffic focused\n- Category navigation, archives\n\n**portfolio** - Showcase/portfolio sites:\n- Creative portfolios, agencies, case studies\n- Visual/image heavy, project galleries\n\n**brochure** - Multi-page business sites:\n- Corporate sites, general business presence\n- Service providers, consultants, professional services\n- About/Services/Contact structure\n\nReturn ONLY valid JSON with this structure:\n{\n  \"site_type\": \"landing|content|portfolio|brochure\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"brief explanation\",\n  \"recommended_builder\": \"builder-type\",\n  \"detected_industry\": \"industry name\",\n  \"detected_signals\": [\"signal1\", \"signal2\", ...]\n}"'::jsonb
                     )
WHERE type = 'site-classifier'
  AND is_active = true;

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps'->'classify_site'->'config'->>'prompt_template' as prompt_template
FROM agent_definitions
WHERE type = 'site-classifier'
  AND is_active = true;

==
-- Fix site-classifier prompt template to use flattened field names
-- The extractDataForAiAgent function flattens input_data to root level

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_site,config,prompt_template}',
        '"Classify this website project and recommend the appropriate builder.\n\nInput:\n- Domain: {{.domain}}\n- Objective: {{.objective}}\n\nAvailable Builders:\n{{range .available_builders.agents}}- {{.type}}: {{.description}}\n{{end}}\n\nClassify the site into ONE of these types based on the objective:\n\n**landing** - Conversion-focused single-purpose sites:\n- Product/service sales pages, SaaS landing pages\n- Lead generation, signups, app downloads\n- Event registration, clear single CTA goal\n\n**content** - Publishing/content sites:\n- News, blogs, magazines, articles\n- Content aggregation, SEO/traffic focused\n- Category navigation, archives\n\n**portfolio** - Showcase/portfolio sites:\n- Creative portfolios, agencies, case studies\n- Visual/image heavy, project galleries\n\n**brochure** - Multi-page business sites:\n- Corporate sites, general business presence\n- Service providers, consultants, professional services\n- About/Services/Contact structure\n\nReturn ONLY valid JSON with this structure:\n{\n  \"site_type\": \"landing|content|portfolio|brochure\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"brief explanation\",\n  \"recommended_builder\": \"builder-type\",\n  \"detected_industry\": \"industry name\",\n  \"detected_signals\": [\"signal1\", \"signal2\", ...]\n}"'::jsonb
                     )
WHERE type = 'site-classifier'
  AND is_active = true;

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps'->'classify_site'->'config'->>'prompt_template' as prompt_template
FROM agent_definitions
WHERE type = 'site-classifier'
  AND is_active = true;


--
-- ============================================================================
-- Add output_type to Agent Configs
-- This tells ai_actions.go whether to append JSON output instructions
-- ========

-- 1. SITE CLASSIFIER - Outputs JSON classification
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_site,config,output_type}',
        '"json"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'site-classifier'
  AND is_active = true;


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

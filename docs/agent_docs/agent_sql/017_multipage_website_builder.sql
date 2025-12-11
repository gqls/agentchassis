id                     | d8c6adcc-7ce7-4b18-af1b-2f34db616c06
type                   | multipage-website-builder
display_name           | Multi-Page Website Builder
description            | Builds large websites (20+ pages) using batched generation to avoid token limits
category               | orchestrator
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Site assembly complete"}, "generate_batch_1": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture", "input_data"], "prompt_template": "Generate complete HTML for pages 1-4 of {{.input_data.domain}}:\n\n1. index.html (Home)\n2. about.html (About Us)\n3. services.html (Services)\n4. team.html (Team)\n\nArchitecture: {{.site_architecture.result}}\n\nFor each page, return complete HTML from <!DOCTYPE> to </html>.\n\nRETURN AS JSON:\n{\n  \"index.html\": \"<!DOCTYPE html>...\",\n  \"about.html\": \"<!DOCTYPE html>...\",\n  \"services.html\": \"<!DOCTYPE html>...\",\n  \"team.html\": \"<!DOCTYPE html>...\"\n}\n\nIMPORTANT: DO NOT include shared CSS - it will be injected automatically."}, "next_step": "generate_batch_2", "description": "Generate pages 1-4", "output_field": "batch_1_pages"}, "generate_batch_2": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture", "input_data"], "prompt_template": "Generate complete HTML for pages 5-8:\n\n5. products.html (Products)\n6. pricing.html (Pricing)\n7. features.html (Features)\n8. testimonials.html (Testimonials)\n\nReturn as JSON map of filename to HTML."}, "next_step": "generate_batch_3", "description": "Generate pages 5-8", "output_field": "batch_2_pages"}, "generate_batch_3": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture", "input_data"], "prompt_template": "Generate complete HTML for pages 9-12:\n\n9. case-studies.html (Case Studies)\n10. blog.html (Blog)\n11. resources.html (Resources)\n12. faq.html (FAQ)\n\nReturn as JSON map."}, "next_step": "generate_batch_4", "description": "Generate pages 9-12", "output_field": "batch_3_pages"}, "generate_batch_4": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture", "input_data"], "prompt_template": "Generate complete HTML for pages 13-16:\n\n13. support.html (Support)\n14. contact.html (Contact)\n15. careers.html (Careers)\n16. partners.html (Partners)\n\nReturn as JSON map."}, "next_step": "generate_batch_5", "description": "Generate pages 13-16", "output_field": "batch_4_pages"}, "generate_batch_5": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture", "input_data"], "prompt_template": "Generate complete HTML for pages 17-20:\n\n17. press.html (Press)\n18. legal.html (Legal)\n19. privacy.html (Privacy Policy)\n20. sitemap.html (Sitemap)\n\nReturn as JSON map."}, "next_step": "assemble_multipage_site", "description": "Generate pages 17-20", "output_field": "batch_5_pages"}, "analyze_requirements": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 4000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["input_data.domain", "input_data.page_list"], "prompt_template": "Analyze requirements for a {{.input_data.domain}} website with these pages: {{.input_data.page_list}}.\n\nCreate a site architecture including:\n- Navigation structure\n- Shared design elements\n- Color scheme and typography\n- Content themes for each page\n\nReturn as JSON."}, "next_step": "generate_shared_styles", "description": "Analyze site requirements and create architecture", "output_field": "site_architecture"}, "generate_shared_styles": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5-20251001", "provider": "anthropic", "max_tokens": 8000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture"], "prompt_template": "Generate shared CSS for entire site based on:\n{{.site_architecture.result}}\n\nInclude:\n- CSS reset and base styles\n- Typography system\n- Color variables\n- Responsive layout utilities\n- Navigation styles\n- Footer styles\n\nReturn ONLY CSS (no HTML, no markdown)."}, "next_step": "generate_batch_1", "description": "Generate shared CSS for all pages", "output_field": "shared_styles"}, "assemble_multipage_site": {"action": "assemble_multipage_site", "config": {"batch_fields": ["batch_1_pages", "batch_2_pages", "batch_3_pages", "batch_4_pages", "batch_5_pages"], "stream_to_s3": true, "index_html_field": "batch_1_pages.index.html", "navigation_field": "site_architecture.navigation", "shared_styles_field": "shared_styles.result"}, "next_step": "complete", "description": "Assemble all pages with shared styles and navigation", "output_field": "assembled_site"}}, "start_step": "analyze_requirements"}, "processing_mode": "orchestrator", "timeout_seconds": 600}
is_active              | t
created_at             | 2025-12-06 19:02:26.358753+00
updated_at             | 2025-12-11 11:15:23.041575+00
deleted_at             |
capabilities           | ["orchestration", "website-builder", "multi-page", "batched-generation"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.528
command                |
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    |
task_workflow          |
orchestrator_workflow  |
orchestration_workflow |
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         |
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         |
output_contract        |


Add briefing_questionnaire to multipage-website-builder

UPDATE agent_definitions
SET briefing_questionnaire = '{
  "sections": [
    {
      "name": "company_info",
      "title": "Company Information",
      "questions": [
        {"type": "text", "field": "company_name", "label": "Company Name", "required": true},
        {"type": "textarea", "field": "about_us", "label": "About Us", "required": true},
        {"type": "text", "field": "tagline", "label": "Tagline", "required": false}
      ]
    },
    {
      "name": "services",
      "title": "Services & Offerings",
      "questions": [
        {"type": "json_array", "field": "services", "label": "Services (list of {name, description})", "required": true}
      ]
    },
    {
      "name": "team",
      "title": "Team & Leadership",
      "questions": [
        {"type": "json_array", "field": "leadership_team", "label": "Leadership Team (list of {name, title, bio})", "required": false}
      ]
    },
    {
      "name": "portfolio",
      "title": "Case Studies & Portfolio",
      "questions": [
        {"type": "json_array", "field": "case_studies", "label": "Case Studies", "required": false}
      ]
    },
    {
      "name": "contact",
      "title": "Contact Information",
      "questions": [
        {"type": "text", "field": "contact_email", "label": "Email", "required": true},
        {"type": "text", "field": "contact_phone", "label": "Phone", "required": false},
        {"type": "text", "field": "headquarters", "label": "Location", "required": false}
      ]
    },
    {
      "name": "features",
      "title": "Site Features",
      "questions": [
        {"type": "boolean", "field": "has_blog", "label": "Include Blog?", "default": false},
        {"type": "boolean", "field": "has_careers", "label": "Include Careers?", "default": false}
      ]
    }
  ]
}'::jsonb
WHERE type = 'multipage-website-builder';
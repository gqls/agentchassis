-- ===========================================================================
-- SITE-ARCHITECT UPDATE: Use component library action
-- ===========================================================================

UPDATE agent_definitions
SET default_config = '{
  "workflow": {
    "start_step": "load_components",
    "steps": {
      "load_components": {
        "action": "load_component_library",
        "config": {
          "component_level": "section",
          "format_for_prompt": true
        },
        "output_field": "component_library",
        "next_step": "design",
        "description": "Load available section components from database"
      },
      "design": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-5-20250929",
            "provider": "anthropic",
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "output_format": "json",
          "input_fields": ["input_data", "domain_analysis", "reviewed_brief", "component_library"],
          "prompt_template": "You are a site architect designing a professional website.\n\n## Business Information\nDomain: {{.input_data.domain}}\nCompany: {{if .reviewed_brief.company_name}}{{.reviewed_brief.company_name}}{{else}}{{.input_data.domain}}{{end}}\nIndustry: {{.reviewed_brief.industry}}\nTone: {{.reviewed_brief.tone}}\nTagline: {{.reviewed_brief.tagline}}\n\n## Services\n{{range .reviewed_brief.services}}- {{.name}}: {{.description}}\n{{end}}\n\n## Available Components\nYou MUST select sections from this list. Use the function name exactly as shown:\n\n{{.component_library.for_prompt}}\n\n## Design Task\nCreate a site architecture with pages appropriate for this business.\n\n## Rules\n1. Every page MUST have at least 2 sections\n2. Use ONLY function names from the available components list above\n3. The index page should have 3-5 sections (hero, then supporting sections)\n4. Secondary pages should have 2-4 sections\n5. Common patterns:\n   - Landing pages: hero → features → social_proof → call_to_action\n   - About pages: hero → generic-text-block → call_to_action\n   - Services pages: hero → features → call_to_action\n   - Contact pages: hero → generic-text-block → call_to_action\n\n## Output JSON Format\n```json\n{\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Company Name\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"page_type\": \"landing\",\n      \"sections\": [\"hero\", \"features\", \"social_proof\", \"call_to_action\"]\n    },\n    {\n      \"name\": \"about\",\n      \"title\": \"About Us | Company Name\",\n      \"nav_label\": \"About\",\n      \"nav_order\": 2,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"page_type\": \"content\",\n      \"sections\": [\"hero\", \"generic-text-block\", \"call_to_action\"]\n    }\n  ],\n  \"style_collection\": \"default\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"hero\": \"Professional hero image for [describe appropriate imagery]\",\n    \"logo\": \"Modern logo for [company name]\"\n  }\n}\n```\n\nReturn ONLY valid JSON, no markdown backticks or explanation."
        },
        "output_field": "site_architecture",
        "next_step": "complete",
        "description": "Design site structure using available components"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {
          "output": {
            "pages": "site_architecture.pages",
            "style_collection": "site_architecture.style_collection",
            "needs_logo": "site_architecture.needs_logo",
            "needs_images": "site_architecture.needs_images",
            "image_prompts": "site_architecture.image_prompts"
          }
        },
        "description": "Return site architecture"
      }
    }
  }
}'::jsonb
WHERE type = 'site-architect';

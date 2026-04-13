-- Agent definition: component-creator
--
-- Handler agent that processes needs_new_component work items.
-- Generates HTML component templates via LLM following the component contracts,
-- then stores them in content_components with selection metadata.
--
-- Called by the dispatch loop when plan_sections creates needs_new_component items.
-- Also callable standalone for manual component creation.
--
-- Workflow: load context from work item spec → generate template via LLM → store component → complete
--
-- Prerequisites:
--   - migration_component_selection_metadata.sql (adds section_type etc. columns)
--   - store_generated_component registered in action registry
--   - component_selector.go deployed (for the selector that creates work items)

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    image_repository, image_tag, command,
    resources, capabilities, domain_tags,
    default_config,
    input_contract, output_contract
) VALUES (
             'component-creator',
             'Component Creator',
             'Generates new HTML component templates from section type descriptions. Processes needs_new_component work items. Stores results in content_components with selection metadata for future reuse.',
             'builder',
             'specialist',
             'experimental',
             'docker.io/aqls/agent-chassis',
             'latest',
             ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
             '{"requests": {"cpu": "100m", "memory": "256Mi"}, "limits": {"cpu": "500m", "memory": "512Mi"}}'::jsonb,
             '["component-generation", "html-template", "css"]'::jsonb,
             '["component-library", "design", "build"]'::jsonb,
             jsonb_build_object(
                     'model', 'claude-sonnet-4-5',
                     'temperature', 0.4,
                     'processing_mode', 'task',
                     'timeout_seconds', 120,
                     'prompt_template', '
You are generating a reusable HTML component template for a website builder system.
The template will be stored in a component library and reused across multiple sites.

SECTION TYPE: {{.input_data.section_type}}
DESCRIPTION: {{.input_data.description}}
SITE TYPE: {{.input_data.site_type}}
PAGE CONTEXT: {{.input_data.page_context}}
{{if .input_data.design_direction}}DESIGN DIRECTION: {{.input_data.design_direction}}{{end}}
{{if .input_data.reference_content}}REFERENCE CONTENT: {{.input_data.reference_content}}{{end}}

== COMPONENT CONTRACT — YOU MUST FOLLOW ALL OF THESE ==

1. STRUCTURE:
   <style> scoped CSS </style>
   <section class="{function}-section" data-component="{function}">
     HTML using {{.variable}} template placeholders
   </section>
   If interactive: <script> self-contained JS in IIFE </script>

2. NAMING:
   - Choose a function name in kebab-case (lowercase, digits, hyphens only)
   - Root element: data-component="{function}" matching exactly
   - Root class: {function}-section

3. TEMPLATE VARIABLES:
   - Use {{.field_name}} for ALL content that varies per instance
   - Do not hardcode any text content — everything the content writer fills is a variable
   - Generate an input_schema declaring each variable field

4. CSS RULES:
   - ALL colours via CSS variables with fallbacks
   - Light sections: color: var(--color-text); headings: var(--color-heading)
   - Dark sections: color: var(--section-text, rgba(255,255,255,0.9))
   - NEVER hardcode hex colours on text elements
   - Scope ALL CSS to .{function}-section — no global element rules
   - Include @media (max-width: 768px) responsive rules
   - Mobile-first: touch targets >= 44px

5. DARK SECTIONS (if the section has a dark background):
   Set on the root container:
     --section-text: rgba(255,255,255,0.9);
     --section-text-muted: rgba(255,255,255,0.7);
     --section-heading: #ffffff;
     --section-surface: rgba(255,255,255,0.05);
     --section-border: rgba(255,255,255,0.2);

6. CSS VARIABLES AVAILABLE:
   --color-primary, --color-primary-hover, --color-primary-text
   --color-secondary, --color-accent
   --color-text, --color-text-muted, --color-heading
   --color-background, --color-surface, --color-card-bg, --color-border
   --color-header-bg, --color-header-text
   --color-footer-bg, --color-footer-text, --color-white
   --container-max-width (1200px), --spacing-section (5rem 2rem)
   --border-radius, --shadow

7. INTERACTIVE ELEMENTS (if section has JS):
   - Client-side only, no external API calls
   - Wrap in IIFE: (function() { ... })();
   - No global variable pollution
   - Progressive enhancement — works without JS where possible
   - No external CDN imports

8. QUALITY:
   - No placeholder text (Lorem ipsum, TODO, [INSERT])
   - No unrendered template variables in output
   - Semantic HTML (section, article, nav — not div soup)
   - Accessible: labels on inputs, ARIA where needed, focus states
   - No fabricated content

== END CONTRACT ==

Respond with ONLY a JSON object (no markdown fences, no preamble) containing:
{
  "function": "the-kebab-case-function-name",
  "html_template": "the full <style>...<section>...</section> template",
  "input_schema": {
    "fields": {
      "field_name": {
        "type": "text|array|image|url|boolean",
        "source": "llm|site_specs.path|site_assets.type|renderer|static",
        "required": true,
        "llm_guidance": "hint for the content writer about what to generate"
      }
    }
  },
  "is_dark_section": true
}
',
                     'workflow', jsonb_build_object(
                             'start_step', 'generate_template',
                             'steps', jsonb_build_object(
                                     'generate_template', jsonb_build_object(
                                     'action', 'execute_llm_prompt',
                                     'description', 'Generate component HTML template from section type description',
                                     'config', jsonb_build_object(
                                             'api_key_env_var', 'ANTHROPIC_API_KEY',
                                             'input_fields', ARRAY['input_data']
                                               ),
                                     'next_step', 'store_component',
                                     'output_field', 'generate_template'
                                                          ),
                                     'store_component', jsonb_build_object(
                                             'action', 'store_generated_component',
                                             'description', 'Store generated template in content_components library',
                                             'config', jsonb_build_object(
                                                     'section_type', 'input_data.section_type',
                                                     'site_type', 'input_data.site_type',
                                                     'page_context', 'input_data.page_context',
                                                     'description', 'input_data.description',
                                                     'design_direction', 'input_data.design_direction',
                                                     'generated_template', 'generate_template'
                                                       ),
                                             'next_step', 'complete',
                                             'output_field', 'stored_component'
                                                        ),
                                     'complete', jsonb_build_object(
                                             'action', 'complete_workflow',
                                             'description', 'Component creation complete',
                                             'config', jsonb_build_object(
                                                     'output_fields', ARRAY['stored_component']
                                                       )
                                                 )
                                      )
                                 )
             ),
             -- Input contract: what the dispatch loop passes
             jsonb_build_object(
                     'required', jsonb_build_array('section_type'),
                     'optional', jsonb_build_array('site_type', 'page_context', 'description', 'design_direction', 'reference_content')
             ),
             -- Output contract: what the agent returns
             jsonb_build_object(
                     'fields', jsonb_build_object(
                     'component_id', 'UUID of the created/existing component',
                     'function', 'The component function name',
                     'section_type', 'The section type it implements',
                     'status', 'created or already_exists'
                               )
             )
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();

-- Register in action registry (add to registry.go):
-- "store_generated_component": {
--     Handler:     StoreGeneratedComponentAction,
--     Category:    "site",
--     Description: "Store a generated component template in the component library",
--     IsLocal:     true,
-- },
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_template,config}',
        '{
            "ai_service": {
                "provider": "anthropic",
                "model": "claude-sonnet-4-5",
                "api_key_env_var": "ANTHROPIC_API_KEY",
                "max_tokens": 4000
            },
            "input_fields": ["input_data"]
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'component-creator';

--

-- Fix prompt: input_data.X → input_data.spec.X
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{prompt_template}',
        to_jsonb(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
                                                                 default_config->>'prompt_template',
                                                                 '{{.input_data.section_type}}', '{{.input_data.spec.section_type}}'),
                                                         '{{.input_data.site_type}}', '{{.input_data.spec.site_type}}'),
                                                 '{{.input_data.description}}', '{{.input_data.spec.description}}'),
                                         '{{.input_data.design_direction}}', '{{.input_data.spec.design_direction}}'),
                                 '{{.input_data.reference_content}}', '{{.input_data.spec.reference_content}}'),
                         '{{.input_data.page_context}}', '{{.input_data.spec.page_context}}'
                 )::text)
                     ),
    updated_at = NOW()
WHERE type = 'component-creator' AND is_active = true;

-- Verify
SELECT
    default_config->>'prompt_template' LIKE '%input_data.spec.section_type%' as uses_spec_path,
    default_config->>'prompt_template' NOT LIKE '%input_data.section_type}}%' as no_bare_path
FROM agent_definitions
WHERE type = 'component-creator' AND is_active = true;
-- Expected: t, t

-- Clean up junk components
DELETE FROM content_components WHERE function IN ('general-hero', 'general-section') AND created_from = 'generated';

-- Reset completed needs_new_component items
UPDATE site_work_items
SET status = 'triaged', error = NULL, result = '{}'::jsonb,
    attempt_count = 0, claimed_by = NULL, claimed_at = NULL, completed_at = NULL
WHERE site_id = (SELECT id FROM sites WHERE domain = 'vonc.com')
  AND item_type = 'needs_new_component'
  AND status = 'complete';
--

-- ============================================================
-- COMPONENT-CREATOR: Add site context loading + improved prompt
-- ============================================================
--
-- Two separate updates to avoid deeply nested jsonb_set escaping issues.
-- Run them in order within the same transaction.
--
-- Changes:
--   1. Workflow: added ensure_site_record + read_site_spec before generate_template
--   2. Prompt: now includes mission_brief, design_intent, content_direction
--   3. max_tokens: 4000 → 8000 (interactive components need room)
--
-- The workflow error_step on read_site_spec falls through to generate_template,
-- so sites without specs still work (the prompt's {{if}} guards handle missing data).

BEGIN;

-- --------------------------------------------------------
-- PART A: Update the workflow to add site context loading
-- --------------------------------------------------------
UPDATE agent_definitions
SET default_config = default_config || jsonb_build_object(
        'workflow', jsonb_build_object(
                'start_step', 'ensure_site_record',
                'steps', jsonb_build_object(
                        'ensure_site_record', jsonb_build_object(
                                'action', 'ensure_site_record',
                                'config', '{}'::jsonb,
                                'next_step', 'read_site_spec',
                                'error_step', 'generate_template',
                                'description', 'Load site record for domain and identity context',
                                'output_field', 'site_record'
                                              ),
                        'read_site_spec', jsonb_build_object(
                                'action', 'read_site_spec',
                                'config', jsonb_build_object('site_id', 'site_record.site_id'),
                                'next_step', 'generate_template',
                                'error_step', 'generate_template',
                                'description', 'Load site specs for mission, design intent, content direction',
                                'output_field', 'site_specs'
                                          ),
                        'generate_template', jsonb_build_object(
                                'action', 'execute_llm_prompt',
                                'config', jsonb_build_object(
                                        'ai_service', jsonb_build_object(
                                                'model', 'claude-sonnet-4-6',
                                                'provider', 'anthropic',
                                                'max_tokens', 8000,
                                                'api_key_env_var', 'ANTHROPIC_API_KEY'
                                                      ),
                                        'input_fields', '["input_data", "site_record", "site_specs"]'::jsonb
                                          ),
                                'next_step', 'store_component',
                                'description', 'Generate component HTML template with full site context',
                                'output_field', 'generate_template'
                                             ),
                        'store_component', jsonb_build_object(
                                'action', 'store_generated_component',
                                'config', jsonb_build_object(
                                        'section_type', 'input_data.section_type',
                                        'site_type', 'input_data.site_type',
                                        'description', 'input_data.description',
                                        'page_context', 'input_data.page_context',
                                        'design_direction', 'input_data.design_direction',
                                        'generated_template', 'generate_template'
                                          ),
                                'next_step', 'complete',
                                'description', 'Store generated template in content_components library',
                                'output_field', 'stored_component'
                                           ),
                        'complete', jsonb_build_object(
                                'action', 'complete_workflow',
                                'config', jsonb_build_object('output_fields', '["stored_component"]'::jsonb),
                                'description', 'Component creation complete'
                                    )
                         )
                    )
                                       ),
    updated_at = now()
WHERE type = 'component-creator';

-- --------------------------------------------------------
-- PART B: Update the prompt template
-- --------------------------------------------------------
-- Using a dollar-quoted string to avoid escaping nightmares.
-- The prompt uses Go template syntax: {{.field}} and {{if .field}}

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{prompt_template}',
        to_jsonb($PROMPT$
            You are generating a reusable HTML component template for a website builder system.
The template will be stored in a component library and reused across multiple sites.

== WHAT YOU ARE BUILDING ==

SECTION TYPE: {{.input_data.spec.section_type}}
DESCRIPTION: {{.input_data.spec.description}}
SITE TYPE: {{.input_data.spec.site_type}}
PAGE CONTEXT: {{.input_data.spec.page_context}}
{{if .input_data.spec.design_direction}}DESIGN DIRECTION: {{.input_data.spec.design_direction}}{{end}}

== SITE CONTEXT ==

{{if .site_record}}DOMAIN: {{.site_record.domain}}
COMPANY: {{.site_record.company_name}}{{end}}

{{if .site_specs}}{{if .site_specs.specs}}{{if .site_specs.specs.mission_brief}}== MISSION ==
{{.site_specs.specs.mission_brief.text}}
{{end}}{{end}}{{end}}

{{if .site_specs}}{{if .site_specs.specs}}{{if .site_specs.specs.design_intent}}== DESIGN INTENT ==
{{if .site_specs.specs.design_intent.style_direction}}Style: {{.site_specs.specs.design_intent.style_direction}}{{end}}
{{if .site_specs.specs.design_intent.color_mood}}Mood: {{.site_specs.specs.design_intent.color_mood}}{{end}}
{{if .site_specs.specs.design_intent.typography_direction}}Typography: {{.site_specs.specs.design_intent.typography_direction}}{{end}}
{{if .site_specs.specs.design_intent.imagery_direction}}Imagery: {{.site_specs.specs.design_intent.imagery_direction}}{{end}}
{{if .site_specs.specs.design_intent.layout_direction}}Layout: {{.site_specs.specs.design_intent.layout_direction}}{{end}}
{{end}}{{end}}{{end}}

{{if .site_specs}}{{if .site_specs.specs}}{{if .site_specs.specs.content_direction}}== CONTENT DIRECTION ==
{{if .site_specs.specs.content_direction.voice}}Voice: {{.site_specs.specs.content_direction.voice}}{{end}}
{{if .site_specs.specs.content_direction.tone}}Tone: {{.site_specs.specs.content_direction.tone}}{{end}}
{{if .site_specs.specs.content_direction.emphasis}}Emphasis: {{.site_specs.specs.content_direction.emphasis}}{{end}}
{{end}}{{end}}{{end}}

{{if .site_specs}}{{if .site_specs.specs}}{{if .site_specs.specs.classification}}== CLASSIFICATION ==
Site type: {{.site_specs.specs.classification.site_type}}
{{if .site_specs.specs.classification.tone_suggestion}}Tone: {{.site_specs.specs.classification.tone_suggestion}}{{end}}
{{if .site_specs.specs.classification.industry}}Industry: {{.site_specs.specs.classification.industry}}{{end}}
{{end}}{{end}}{{end}}

== COMPONENT CONTRACT — YOU MUST FOLLOW ALL OF THESE ==

1. STRUCTURE:
   <style> scoped CSS </style>
   <section class="{function}-section" data-component="{function}">
     HTML using {{.variable}} template placeholders
   </section>
   If interactive: <script> self-contained JS in IIFE </script>

2. NAMING:
   - Choose a function name in kebab-case (lowercase, digits, hyphens only)
   - Root element: data-component="{function}" matching exactly
   - Root class: {function}-section

3. TEMPLATE VARIABLES:
   - Use {{.field_name}} for ALL content that varies per instance
   - Do not hardcode any text content — everything the content writer fills is a variable
   - Generate an input_schema declaring each variable field
   - The input_schema llm_guidance field is important: it tells the content writer
     WHAT to write for this site. Be specific using the mission and site context above.

4. CSS RULES:
   - ALL colours via CSS variables with fallbacks
   - Light sections: color: var(--color-text); headings: var(--color-heading)
   - Dark sections: color: var(--section-text, rgba(255,255,255,0.9))
   - NEVER hardcode hex colours on text elements
   - Scope ALL CSS to .{function}-section — no global element rules
   - Include @media (max-width: 768px) responsive rules
   - Mobile-first: touch targets >= 44px
   - Use the design intent (mood, typography) to inform spacing, border-radius, shadows

5. DARK SECTIONS (if the section has a dark background):
   Set on the root container:
     --section-text: rgba(255,255,255,0.9);
     --section-text-muted: rgba(255,255,255,0.7);
     --section-heading: #ffffff;
     --section-surface: rgba(255,255,255,0.05);
     --section-border: rgba(255,255,255,0.2);

6. CSS VARIABLES AVAILABLE:
   --color-primary, --color-primary-hover, --color-primary-text
   --color-secondary, --color-accent
   --color-text, --color-text-muted, --color-heading
   --color-background, --color-surface, --color-card-bg, --color-border
   --color-header-bg, --color-header-text
   --color-footer-bg, --color-footer-text, --color-white
   --container-max-width (1200px), --spacing-section (5rem 2rem)
   --border-radius, --shadow

7. INTERACTIVE ELEMENTS (if section has JS):
   - Client-side only, no external API calls
   - Wrap in IIFE: (function() { ... })();
- No global variable pollution
   - Progressive enhancement — works without JS where possible
   - No external CDN imports

8. QUALITY:
   - No placeholder text (Lorem ipsum, TODO, [INSERT])
   - No unrendered template variables in output
   - Semantic HTML (section, article, nav — not div soup)
   - Accessible: labels on inputs, ARIA where needed, focus states
   - No fabricated content
   - The component should feel like it belongs on this specific site.
     Use the mission and design intent to shape the visual language.

== END CONTRACT ==

Respond with ONLY a JSON object (no markdown fences, no preamble) containing:
{
  "function": "the-kebab-case-function-name",
  "html_template": "the full <style>...<section>...</section> template",
  "input_schema": {
    "fields": {
      "field_name": {
        "type": "text|array|image|url|boolean",
        "source": "llm|site_specs.path|site_assets.type|renderer|static",
        "required": true,
        "llm_guidance": "specific hint for the content writer about what to generate for this site"
      }
    }
  },
  "is_dark_section": true
}
$PROMPT$::text)
),
updated_at = now()
WHERE type = 'component-creator';

COMMIT;

-- --------------------------------------------------------
-- VERIFY
-- --------------------------------------------------------
SELECT type,
       default_config->'workflow'->>'start_step' as start_step,
    (SELECT array_agg(key) FROM jsonb_object_keys(default_config->'workflow'->'steps') key) as step_names,
    LENGTH(default_config->>'prompt_template') as prompt_len,
    default_config->'workflow'->'steps'->'generate_template'->'config'->'ai_service'->>'max_tokens' as max_tokens
FROM agent_definitions
WHERE type = 'component-creator';

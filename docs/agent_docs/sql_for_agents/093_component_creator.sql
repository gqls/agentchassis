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

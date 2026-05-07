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


---

-- ============================================================
-- Revised component-creator prompt with field classification
-- ============================================================
-- Replaces section 3 (TEMPLATE VARIABLES) and the response schema
-- to teach the LLM how to classify fields into three tiers:
--   - Voice-defining content (source: "llm", required: true)
--   - Tunable labels (source: "static", required: false, with fallback)
--   - Site data (source: "site_specs.*" or similar)
--
-- Also adds a critical invariant: the html_template variables and
-- the input_schema fields MUST be in sync. If the template has
-- {{.foo}}, the schema MUST have "foo". If the schema has "bar",
-- the template MUST use {{.bar}}.

-- View current prompt first to see what we're working with
SELECT LENGTH(default_config->>'prompt_template') as prompt_len,
    LEFT(default_config->>'prompt_template', 500) as prompt_start
FROM agent_definitions
WHERE type = 'component-creator';


-- ============================================================
-- The new prompt (apply via UPDATE once reviewed)
-- ============================================================

-- Use jsonb_set to update prompt_template under default_config
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{prompt_template}',
        to_jsonb('
You are generating a reusable HTML component template for a website builder system.
The template will be stored in a component library and reused across multiple sites.

SECTION TYPE: {{.input_data.spec.section_type}}
DESCRIPTION: {{.input_data.spec.description}}
SITE TYPE: {{.input_data.spec.site_type}}
PAGE CONTEXT: {{.input_data.spec.page_context}}
{{if .input_data.spec.design_direction}}DESIGN DIRECTION: {{.input_data.spec.design_direction}}{{end}}
{{if .input_data.spec.mission_brief}}MISSION BRIEF: {{.input_data.spec.mission_brief}}{{end}}
{{if .input_data.spec.design_intent}}DESIGN INTENT: {{.input_data.spec.design_intent}}{{end}}
{{if .input_data.spec.content_direction}}CONTENT DIRECTION: {{.input_data.spec.content_direction}}{{end}}
{{if .input_data.spec.classification}}CLASSIFICATION: {{.input_data.spec.classification}}{{end}}
{{if .input_data.spec.reference_content}}REFERENCE CONTENT: {{.input_data.spec.reference_content}}{{end}}

== COMPONENT CONTRACT — YOU MUST FOLLOW ALL OF THESE ==

1. STRUCTURE:
   <style> scoped CSS </style>
   <section class="{function}-section" data-component="{function}">
     HTML using {{.variable}} template placeholders
   </section>
   If interactive: <script> self-contained JS in IIFE </script>
   (The pipeline will automatically extract inline <script> blocks into a
   separate JS asset — include the <script> tag naturally, do not attempt
   to reference external files yourself.)

2. NAMING:
   - Choose a function name in kebab-case (lowercase, digits, hyphens only)
   - Root element: data-component="{function}" matching exactly
   - Root class: {function}-section

3. TEMPLATE VARIABLES — FIELD CLASSIFICATION:

   Classify every piece of non-structural text into ONE of these three tiers.
   Use {{.field_name}} placeholders for tier A and tier B; fully hardcode nothing.

   TIER A — VOICE CONTENT (source: "llm", required: true)
     Fields that define the site''s voice, positioning, or message.
     The content writer MUST produce these or the section has no reason to exist.
     Examples: hero headline, section intro paragraph, CTA heading, testimonial quote.
     Typically 3-10 fields per component.

   TIER B — TUNABLE LABELS (source: "static", required: false, with fallback)
     UI labels that MIGHT vary per site tone but have a safe default.
     The template renders with the fallback if the LLM skips them.
     Examples: button labels ("Submit", "Learn more"), stat labels ("Users", "Active"),
       placeholder text, timer labels, badge text, tab labels.
     Provide a sensible English fallback in the schema.
     Typically 5-20 fields per component.

   TIER C — SITE DATA (source: "site_specs.{path}" or "site_assets.{type}")
     Fields that come from the site''s stored data, not from the LLM.
     Examples: company_name, contact_email, hero_image_url.
     Mark required: true only if the section genuinely cannot render without them.

   Also: RENDERER fields (source: "renderer", required: false)
     Values filled at render time by JS or the renderer, not by the content writer.
     Provide a safe initial fallback (e.g. "00:00" for a timer).

   RULE: every {{.variable}} in the html_template MUST appear in input_schema.fields.
         Every key in input_schema.fields MUST appear as {{.variable}} in the template.
         Template and schema are two views of the same contract.

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
  "html_template": "the full <style>...<section>...</section> template with {{.variables}}",
  "input_schema": {
    "fields": {
      "voice_field_example": {
        "type": "text",
        "source": "llm",
        "required": true,
        "llm_guidance": "Detailed guidance for the content writer"
      },
      "tunable_label_example": {
        "type": "text",
        "source": "static",
        "required": false,
        "fallback": "Learn more",
        "llm_guidance": "Optional: override default if site tone differs"
      },
      "site_data_example": {
        "type": "text",
        "source": "site_specs.identity.company_name",
        "required": true
      },
      "renderer_field_example": {
        "type": "text",
        "source": "renderer",
        "required": false,
        "fallback": "00:00",
        "llm_guidance": "Initial display before JS takes over"
      }
    }
  },
  "is_dark_section": true
}
'::text),
        false  -- preserve existing structure; only replace the prompt_template key
                     ),
    updated_at = now()
WHERE type = 'component-creator';

-- Verify
SELECT LENGTH(default_config->>'prompt_template') as new_prompt_len,
       default_config->'prompt_template' IS NOT NULL as has_prompt
FROM agent_definitions
WHERE type = 'component-creator';

---
-- improve prompt

-- ============================================================================
-- Update component-creator prompt_template to use {{placeholder "x"}} syntax
-- ============================================================================
--
-- Why: the previous prompt contained literal {{.variable}} and {{.variables}}
-- in illustrative examples. RenderPromptTemplate interprets these as Go
-- text/template actions and resolves them to <no value> at render time, so
-- the LLM never actually sees {{.x}} and instead imitates whatever surrounding
-- pattern is visible in the rendered prompt. Result: components with 51-field
-- schemas and 0 template placeholders (provocation-feed, archetype-combinations).
--
-- This UPDATE REPLACES the prompt with one that uses {{placeholder "field_name"}}
-- for every illustrative placeholder. The new `placeholder` template function
-- (added to datahelpers.RenderPromptTemplate's funcMap) renders those calls as
-- the literal string {{.field_name}}, so the LLM sees the correct target syntax.
--
-- REQUIRED DEPLOYMENTS BEFORE APPLYING THIS SQL:
--   1. Chassis binary with the new `placeholder` funcMap entry in
--      platform/orchestration/datahelpers/data_helpers.go. Without it, the
--      Parse call inside RenderPromptTemplate will fail with
--        "function \"placeholder\" not defined"
--      and the component-creator agent will error on every work item.
--   2. Chassis binary with Layer 1 pre-store validation in
--      platform/orchestration/actions/store_generated_component_action.go
--      (optional but recommended — protects against LLM non-compliance).
--
-- Dollar-quoted ($PROMPT$...$PROMPT$) so single quotes inside the prompt
-- (site's, hasn't, site's) and the single $ in ${field_name} (an INVALID
-- syntax example) don't need escaping. The tag $PROMPT$ does not appear in
-- the prompt body (verified).
--
-- Agent ID: 23720180-7a39-4e3d-92e1-ebdbf95b57f4 (component-creator).
-- ============================================================================

BEGIN;

-- Preview what we're replacing (for sanity)
SELECT id, type,
       length(default_config->>'prompt_template') AS old_prompt_bytes,
    left(default_config->>'prompt_template', 80) || '…' AS old_prompt_start
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';

-- Apply the update
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{prompt_template}',
        to_jsonb($PROMPT$

            You are generating a reusable HTML component template for a website builder system.
The template will be stored in a component library and reused across multiple sites.

SECTION TYPE: {{.input_data.spec.section_type}}
DESCRIPTION: {{.input_data.spec.description}}
SITE TYPE: {{.input_data.spec.site_type}}
PAGE CONTEXT: {{.input_data.spec.page_context}}
{{if .input_data.spec.design_direction}}DESIGN DIRECTION: {{.input_data.spec.design_direction}}{{end}}
{{if .input_data.spec.mission_brief}}MISSION BRIEF: {{.input_data.spec.mission_brief}}{{end}}
{{if .input_data.spec.design_intent}}DESIGN INTENT: {{.input_data.spec.design_intent}}{{end}}
{{if .input_data.spec.content_direction}}CONTENT DIRECTION: {{.input_data.spec.content_direction}}{{end}}
{{if .input_data.spec.classification}}CLASSIFICATION: {{.input_data.spec.classification}}{{end}}
{{if .input_data.spec.reference_content}}REFERENCE CONTENT: {{.input_data.spec.reference_content}}{{end}}

== COMPONENT CONTRACT — YOU MUST FOLLOW ALL OF THESE ==

1. STRUCTURE:
   <style> scoped CSS </style>
   <section class="{function}-section" data-component="{function}">
     HTML using Go-template placeholders of the form {{placeholder "variable_name"}}
   </section>
   If interactive: <script> self-contained JS in IIFE </script>
   (The pipeline will automatically extract inline <script> blocks into a
   separate JS asset — include the <script> tag naturally, do not attempt
   to reference external files yourself.)

2. PLACEHOLDER SYNTAX — ONE FORMAT ONLY:
   Every variable in the HTML MUST use Go text/template action syntax exactly:
     {{placeholder "field_name"}}
   This is the ONLY accepted placeholder format. The following are INVALID
   and will cause the component to be rejected by validation:
     - <no value>field_name</no>
     - <field_name>
     - {field_name}
     - [[field_name]]
     - ${field_name}
     - {% field_name %}
     - Hardcoded literal text in place of what should be a placeholder

3. NAMING:
   - Choose a function name in kebab-case (lowercase, digits, hyphens only)
   - Root element: data-component="{function}" matching exactly
   - Root class: {function}-section

4. TEMPLATE VARIABLES — FIELD CLASSIFICATION:

   Classify every piece of non-structural text into ONE of these three tiers.
   Use {{placeholder "field_name"}} for tier A and tier B; fully hardcode nothing.

   TIER A — VOICE CONTENT (source: "llm", required: true)
     Fields that define the site's voice, positioning, or message.
     The content writer MUST produce these or the section has no reason to exist.
     Examples: hero headline, section intro paragraph, CTA heading, testimonial quote.
     Typically 3-10 fields per component.

   TIER B — TUNABLE LABELS (source: "static", required: false, with fallback)
     UI labels that MIGHT vary per site tone but have a safe default.
     Put {{placeholder "field_name"}} in the template just like any other field.
     The RENDERER substitutes the fallback value from the schema at render time
     when the content writer hasn't provided an override.
     DO NOT hardcode the fallback text directly into the template HTML. The
     placeholder MUST be present, or the field is not reachable and the
     "fallback" in the schema has no effect.
     Examples: button labels ("Submit", "Learn more"), stat labels ("Users",
       "Active"), placeholder text, timer labels, badge text, tab labels.
     Provide a sensible English fallback in the schema.
     Typically 5-20 fields per component.

   TIER C — SITE DATA (source: "site_specs.{path}" or "site_assets.{type}")
     Fields that come from the site's stored data, not from the LLM.
     Examples: company_name, contact_email, hero_image_url.
     Mark required: true only if the section genuinely cannot render without them.

   Also: RENDERER fields (source: "renderer", required: false)
     Values filled at render time by JS or the renderer, not by the content writer.
     Provide a safe initial fallback (e.g. "00:00" for a timer).

   RULE: every {{placeholder "field_name"}} in the html_template MUST have a matching
   entry in input_schema.fields. Every key in input_schema.fields MUST appear
   as {{placeholder "that_key"}} somewhere in the template. Template and
   schema are two views of the same contract. Count both before returning;
   if the counts differ, revise the template to add the missing placeholders.

   Worked example — CORRECT:
     html_template fragment:
       <button class="cta">{{placeholder "cta_primary_label"}}</button>
     input_schema.fields.cta_primary_label:
       { "type": "text", "source": "static", "fallback": "Get started",
         "required": false, "llm_guidance": "..." }

   Worked example — WRONG (fallback hardcoded, no placeholder):
     html_template fragment:
       <button class="cta">Get started</button>
     input_schema.fields.cta_primary_label:
       { "type": "text", "source": "static", "fallback": "Get started", ... }
     The wrong version produces a 0-placeholder template and a schema that
     describes fields the template can never consume.

5. CSS RULES:
   - ALL colours via CSS variables with fallbacks
   - Light sections: color: var(--color-text); headings: var(--color-heading)
   - Dark sections: color: var(--section-text, rgba(255,255,255,0.9))
   - NEVER hardcode hex colours on text elements
   - Scope ALL CSS to .{function}-section — no global element rules
   - Include @media (max-width: 768px) responsive rules
   - Mobile-first: touch targets >= 44px

6. DARK SECTIONS (if the section has a dark background):
   Set on the root container:
     --section-text: rgba(255,255,255,0.9);
     --section-text-muted: rgba(255,255,255,0.7);
     --section-heading: #ffffff;
     --section-surface: rgba(255,255,255,0.05);
     --section-border: rgba(255,255,255,0.2);

7. CSS VARIABLES AVAILABLE:
   --color-primary, --color-primary-hover, --color-primary-text
   --color-secondary, --color-accent
   --color-text, --color-text-muted, --color-heading
   --color-background, --color-surface, --color-card-bg, --color-border
   --color-header-bg, --color-header-text
   --color-footer-bg, --color-footer-text, --color-white
   --container-max-width (1200px), --spacing-section (5rem 2rem)
   --border-radius, --shadow

8. INTERACTIVE ELEMENTS (if section has JS):
   - Client-side only, no external API calls
   - Wrap in IIFE: (function() { ... })();
   - No global variable pollution
   - Progressive enhancement — works without JS where possible
   - No external CDN imports

9. QUALITY:
   - No placeholder text (Lorem ipsum, TODO, [INSERT])
   - No unrendered template variables in output
   - Semantic HTML (section, article, nav — not div soup)
   - Accessible: labels on inputs, ARIA where needed, focus states
   - No fabricated content

== PRE-OUTPUT SELF-CHECK ==

Before returning your JSON, verify these two conditions:

(a) Count the occurrences of {{placeholder "field_name"}} tokens in your html_template.
    Count the keys in your input_schema.fields object.
    These two counts MUST be equal. If they differ, revise the template to
    add the missing placeholders, or revise the schema to remove the
    orphaned fields. Do not return a mismatched pair.

(b) Every schema field (Tier A, Tier B, Tier C, Renderer) must be reachable
    through its {{placeholder "field_name"}} token somewhere in the template. Tier B
    fallbacks are not a licence to omit the placeholder — the fallback only
    takes effect when the placeholder is present.

== END CONTRACT ==

Respond with ONLY a JSON object (no markdown fences, no preamble) containing:
{
  "function": "the-kebab-case-function-name",
  "html_template": "the full <style>...<section>...</section> template with Go-template placeholders as specified in the CONTRACT above",
  "input_schema": {
    "fields": {
      "voice_field_example": {
        "type": "text",
        "source": "llm",
        "required": true,
        "llm_guidance": "Detailed guidance for the content writer"
      },
      "tunable_label_example": {
        "type": "text",
        "source": "static",
        "required": false,
        "fallback": "Learn more",
        "llm_guidance": "Optional: override default if site tone differs"
      },
      "site_data_example": {
        "type": "text",
        "source": "site_specs.identity.company_name",
        "required": true
      },
      "renderer_field_example": {
        "type": "text",
        "source": "renderer",
        "required": false,
        "fallback": "00:00",
        "llm_guidance": "Initial display before JS takes over"
      }
    }
  },
  "is_dark_section": true
}
$PROMPT$::text)
    ),
    updated_at = NOW()
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
RETURNING id, type,
          length(default_config->>'prompt_template') AS new_prompt_bytes,
          left(default_config->>'prompt_template', 80) || '…' AS new_prompt_start;

-- Confirm one row updated before committing
DO $verify$
DECLARE
    cnt int;
    has_placeholder boolean;
    has_old_literal boolean;
BEGIN
    SELECT count(*) INTO cnt
    FROM agent_definitions
    WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';
    IF cnt <> 1 THEN
        RAISE EXCEPTION 'expected 1 component-creator row, found %', cnt;
    END IF;

    -- New prompt MUST contain the placeholder function calls
    SELECT (default_config->>'prompt_template') ~ '\{\{placeholder '
    INTO has_placeholder
    FROM agent_definitions
    WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';
    IF NOT has_placeholder THEN
        RAISE EXCEPTION 'new prompt_template does not contain {{placeholder ...}} calls — update failed or prompt body missing';
    END IF;

    -- Old illustrative {{.variable}} / {{.variables}} tokens should be GONE.
    -- Note: the prompt still has legitimate {{.input_data.spec.x}} actions at
    -- the top — those are the real template data references, not illustrations.
    -- So we check specifically for the old illustration tokens that caused the bug.
    SELECT (default_config->>'prompt_template') ~ '\{\{\.variable\}\}|\{\{\.variables\}\}'
    INTO has_old_literal
    FROM agent_definitions
    WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';
    IF has_old_literal THEN
        RAISE EXCEPTION 'new prompt_template still contains old illustrative {{.variable}}/{{.variables}} tokens — migration incomplete';
    END IF;
END;
$verify$;

COMMIT;

-- ============================================================================
-- Post-apply: regenerate the two known-broken components
-- ============================================================================
-- Run these AFTER the chassis binary with `placeholder` funcMap is deployed,
-- AFTER this UPDATE is applied, AND AFTER confirming the verification block
-- in this file committed successfully.
--
-- These create work items that the component-creator agent will pick up and
-- process. Expected outcome: templates with 30+ {{.x}} placeholders, schema
-- and template in sync, quality_score > 70.
--
-- Uncomment to run:
--
-- INSERT INTO work_items (
--     id, summary, item_type, item_domain, handler_agent, source,
--     priority, severity, spec_data, status, created_at, updated_at
-- ) VALUES
-- (
--     gen_random_uuid(),
--     'regenerate provocation-feed with corrected prompt',
--     'needs_component_regeneration',
--     'build',
--     'component-creator',
--     'manual_regen_after_prompt_fix',
--     80,
--     'high',
--     jsonb_build_object(
--         'function', 'provocation-feed',
--         'section_type', 'provocation-feed',
--         'quality_issues', jsonb_build_array('0 placeholders, 51 schema fields')
--     ),
--     'pending',
--     NOW(), NOW()
-- ),
-- (
--     gen_random_uuid(),
--     'regenerate archetype-combinations with corrected prompt',
--     'needs_component_regeneration',
--     'build',
--     'component-creator',
--     'manual_regen_after_prompt_fix',
--     80,
--     'high',
--     jsonb_build_object(
--         'function', 'archetype-combinations',
--         'section_type', 'archetype-combinations',
--         'quality_issues', jsonb_build_array('0 placeholders, 35 schema fields')
--     ),
--     'pending',
--     NOW(), NOW()
-- );
--
-- After component-creator processes these, verify:
--
-- SELECT function, template_variable_count, schema_field_count,
--        schema_template_synced, quality_score, quality_issues
-- FROM content_components
-- WHERE function IN ('provocation-feed', 'archetype-combinations')
--   AND is_active = true AND forked_from IS NULL
-- ORDER BY created_at DESC;
--
-- Success criteria:
--   template_variable_count > 0
--   schema_template_synced = true
--   quality_score >= 70

--

                                          -- ============================================================================
-- Migration 036: Component registry hygiene
-- ============================================================================
-- Two cleanup operations on content_components, motivated by issues
-- surfaced during the gamesdesign Phase 1 build:
--
-- (a) Component name vs function divergence.
--
-- The naming contract (doc 003) requires `name` to equal `function` in
-- kebab-case. Several rows have snake_case names but kebab-case
-- functions:
--
--     name              | function
--     ------------------+------------------
--     category_section  | category-listing
--     article_grid      | content-listing
--     call_to_action    | call-to-action
--
-- Verified via earlier audit query: zero page_components rows reference
-- these by name, so renaming is safe with no orphan risk.
--
-- (b) Truncated component templates.
--
-- tool-list, guide-list, game-list have ~10K-character html_template
-- values that lack a closing </section> tag — the LLM that generated
-- them ran out of token budget mid-output. plan_sections's truncation
-- guard correctly skips them, but with no alternative they fall through
-- to "selector unavailable — passing to content writer as-is", which
-- yields free-form content the validator then rejects as containing
-- "placeholder" substrings.
--
-- Deactivating these rows lets component-creator regenerate them
-- cleanly. Once the regenerated rows land (each as a fresh row with
-- the same function value, is_active=true), the planner picks them up
-- on the next plan run.
--
-- All operations are idempotent and self-checking. Re-running the
-- migration is a no-op if the changes are already in place.
-- ============================================================================

BEGIN;

-- ── (a) Rename contract violations ───────────────────────────────────────

-- Check first that these rows aren't being referenced by name in
-- page_components. The reference path is by component_id (UUID), so this
-- shouldn't be possible, but worth confirming explicitly.
DO $$
DECLARE
ref_count INT;
BEGIN
SELECT COUNT(*) INTO ref_count
FROM page_components pc
         JOIN content_components cc ON pc.component_id = cc.id
WHERE cc.name IN ('category_section', 'article_grid', 'call_to_action')
  AND cc.name <> cc.function;
IF ref_count > 0 THEN
        RAISE NOTICE 'Found % page_components rows referencing snake_case names; renaming is still safe via UUID, but flagged for awareness.', ref_count;
END IF;
END$$;

UPDATE content_components
SET name = function,
    updated_at = NOW()
WHERE name IN ('category_section', 'article_grid', 'call_to_action')
  AND name <> function;

-- ── (b) Deactivate truncated templates ───────────────────────────────────

-- Identify by the truncation signature: long template, missing </section>.
-- Wrap in CTE so the same selection is used for both deactivation and
-- the regeneration work item insertion.
WITH truncated AS (
    SELECT id, name, function
    FROM content_components
    WHERE is_active = true
      AND html_template IS NOT NULL
      AND LENGTH(html_template) > 500
      AND html_template NOT LIKE '%</section>%'
)
UPDATE content_components cc
SET is_active = false,
    updated_at = NOW()
    FROM truncated t
WHERE cc.id = t.id;

-- Emit a needs_component_regeneration work item for each function that
-- was just deactivated, so the component-creator agent picks them up.
-- Pipeline 'build' so they're visible to the build dispatch loop;
-- handler_agent 'component-creator' so dispatch routes correctly.
--
-- spec carries the function name and a regen reason for the creator
-- agent's prompt context. item_key uses the function name so the dedup
-- index prevents duplicates if the migration is run twice.
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by,
    item_key
)
SELECT
    -- A NULL site_id signals a global / cross-site item. component_creator
    -- handles either — many components are site-agnostic. If the schema
    -- requires non-null site_id, this query needs adjusting to pick a
    -- canonical "system" site UUID (eac60db8-... per the sites table).
    'eac60db8-b032-432b-b36d-76f37632045d'::uuid,
    '036_component_hygiene',
    'build',
    'needs_component_regeneration',
    'high',
    'Regenerate truncated template for component: ' || cc.function,
    jsonb_build_object(
            'function',  cc.function,
            'name',      cc.name,
            'reason',    'template_truncated',
            'old_id',    cc.id::text,
            'old_length', LENGTH(cc.html_template)
    ),
    20, -- mid priority; not blocking but should run promptly
    'component-creator',
    'triaged',
    '036_phase1_hygiene',
    'needs_component:' || cc.function
FROM content_components cc
WHERE cc.is_active = false
  AND cc.html_template IS NOT NULL
  AND LENGTH(cc.html_template) > 500
  AND cc.html_template NOT LIKE '%</section>%'
  AND cc.updated_at > NOW() - INTERVAL '1 minute' -- only the rows we just deactivated
ON CONFLICT DO NOTHING;

-- ── Verification ─────────────────────────────────────────────────────────

-- (a) After: no remaining name<>function divergence among active rows
SELECT name, function, is_active
FROM content_components
WHERE name <> function AND is_active = true;

-- (b) After: deactivated truncated rows
SELECT name, function, is_active, LENGTH(html_template) AS len
FROM content_components
WHERE function IN ('tool-list', 'guide-list', 'game-list');

-- (c) After: regeneration work items emitted
SELECT item_key, status, summary, spec
FROM site_work_items
WHERE source = '036_component_hygiene'
ORDER BY item_key;

COMMIT;

----- backup
23720180-7a39-4e3d-92e1-ebdbf95b57f4 | component-creator | Component Creator | Generates new HTML component templates from section type descriptions. Processes needs_new_component work items. Stores results in content_components with selection metadata for future reuse. | builder  | {"model": "claude-sonnet-4-6", "workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["stored_component"]}, "description": "Component creation complete"}, "read_site_spec": {"action": "read_site_spec", "config": {"site_id": "site_record.site_id"}, "next_step": "generate_template", "error_step": "generate_template", "description": "Load site specs for mission, design intent, content direction", "output_field": "site_specs"}, "store_component": {"action": "store_generated_component", "config": {"site_type": "input_data.site_type", "description": "input_data.description", "page_context": "input_data.page_context", "section_type": "input_data.section_type", "design_direction": "input_data.design_direction", "generated_template": "generate_template"}, "next_step": "complete", "description": "Store generated template in content_components library", "output_field": "stored_component"}, "generate_template": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["input_data", "site_record", "site_specs"]}, "next_step": "store_component", "description": "Generate component HTML template with full site context", "output_field": "generate_template"}, "ensure_site_record": {"action": "ensure_site_record", "config": {}, "next_step": "read_site_spec", "error_step": "generate_template", "description": "Load site record for domain and identity context", "output_field": "site_record"}}, "start_step": "ensure_site_record"}, "temperature": 0.4, "processing_mode": "task", "prompt_template": "\n\nYou are generating a reusable HTML component template for a website builder system.\nThe template will be stored in a component library and reused across multiple sites.\n\nSECTION TYPE: {{.input_data.spec.section_type}}\nDESCRIPTION: {{.input_data.spec.description}}\nSITE TYPE: {{.input_data.spec.site_type}}\nPAGE CONTEXT: {{.input_data.spec.page_context}}\n{{if .input_data.spec.design_direction}}DESIGN DIRECTION: {{.input_data.spec.design_direction}}{{end}}\n{{if .input_data.spec.mission_brief}}MISSION BRIEF: {{.input_data.spec.mission_brief}}{{end}}\n{{if .input_data.spec.design_intent}}DESIGN INTENT: {{.input_data.spec.design_intent}}{{end}}\n{{if .input_data.spec.content_direction}}CONTENT DIRECTION: {{.input_data.spec.content_direction}}{{end}}\n{{if .input_data.spec.classification}}CLASSIFICATION: {{.input_data.spec.classification}}{{end}}\n{{if .input_data.spec.reference_content}}REFERENCE CONTENT: {{.input_data.spec.reference_content}}{{end}}\n\n== COMPONENT CONTRACT — YOU MUST FOLLOW ALL OF THESE ==\n\n1. STRUCTURE:\n   <style> scoped CSS </style>\n   <section class=\"{function}-section\" data-component=\"{function}\">\n     HTML using Go-template placeholders of the form {{placeholder \"variable_name\"}}\n   </section>\n   If interactive: <script> self-contained JS in IIFE </script>\n   (The pipeline will automatically extract inline <script> blocks into a\n   separate JS asset — include the <script> tag naturally, do not attempt\n   to reference external files yourself.)\n\n2. PLACEHOLDER SYNTAX — ONE FORMAT ONLY:\n   Every variable in the HTML MUST use Go text/template action syntax exactly:\n     {{placeholder \"field_name\"}}\n   This is the ONLY accepted placeholder format. The following are INVALID\n   and will cause the component to be rejected by validation:\n     - <no value>field_name</no>\n     - <field_name>\n     - {field_name}\n     - [[field_name]]\n     - ${field_name}\n     - {% field_name %}\n     - Hardcoded literal text in place of what should be a placeholder\n\n3. NAMING:\n   - Choose a function name in kebab-case (lowercase, digits, hyphens only)\n   - Root element: data-component=\"{function}\" matching exactly\n   - Root class: {function}-section\n\n4. TEMPLATE VARIABLES — FIELD CLASSIFICATION:\n\n   Classify every piece of non-structural text into ONE of these three tiers.\n   Use {{placeholder \"field_name\"}} for tier A and tier B; fully hardcode nothing.\n\n   TIER A — VOICE CONTENT (source: \"llm\", required: true)\n     Fields that define the site's voice, positioning, or message.\n     The content writer MUST produce these or the section has no reason to exist.\n     Examples: hero headline, section intro paragraph, CTA heading, testimonial quote.\n     Typically 3-10 fields per component.\n\n   TIER B — TUNABLE LABELS (source: \"static\", required: false, with fallback)\n     UI labels that MIGHT vary per site tone but have a safe default.\n     Put {{placeholder \"field_name\"}} in the template just like any other field.\n     The RENDERER substitutes the fallback value from the schema at render time\n     when the content writer hasn't provided an override.\n     DO NOT hardcode the fallback text directly into the template HTML. The\n     placeholder MUST be present, or the field is not reachable and the\n     \"fallback\" in the schema has no effect.\n     Examples: button labels (\"Submit\", \"Learn more\"), stat labels (\"Users\",\n       \"Active\"), placeholder text, timer labels, badge text, tab labels.\n     Provide a sensible English fallback in the schema.\n     Typically 5-20 fields per component.\n\n   TIER C — SITE DATA (source: \"site_specs.{path}\" or \"site_assets.{type}\")\n     Fields that come from the site's stored data, not from the LLM.\n     Examples: company_name, contact_email, hero_image_url.\n     Mark required: true only if the section genuinely cannot render without them.\n\n   Also: RENDERER fields (source: \"renderer\", required: false)\n     Values filled at render time by JS or the renderer, not by the content writer.\n     Provide a safe initial fallback (e.g. \"00:00\" for a timer).\n\n   RULE: every {{placeholder \"field_name\"}} in the html_template MUST have a matching\n   entry in input_schema.fields. Every key in input_schema.fields MUST appear\n   as {{placeholder \"that_key\"}} somewhere in the template. Template and\n   schema are two views of the same contract. Count both before returning;\n   if the counts differ, revise the template to add the missing placeholders.\n\n   Worked example — CORRECT:\n     html_template fragment:\n       <button class=\"cta\">{{placeholder \"cta_primary_label\"}}</button>\n     input_schema.fields.cta_primary_label:\n       { \"type\": \"text\", \"source\": \"static\", \"fallback\": \"Get started\",\n         \"required\": false, \"llm_guidance\": \"...\" }\n\n   Worked example — WRONG (fallback hardcoded, no placeholder):\n     html_template fragment:\n       <button class=\"cta\">Get started</button>\n     input_schema.fields.cta_primary_label:\n       { \"type\": \"text\", \"source\": \"static\", \"fallback\": \"Get started\", ... }\n     The wrong version produces a 0-placeholder template and a schema that\n     describes fields the template can never consume.\n\n5. CSS RULES:\n   - ALL colours via CSS variables with fallbacks\n   - Light sections: color: var(--color-text); headings: var(--color-heading)\n   - Dark sections: color: var(--section-text, rgba(255,255,255,0.9))\n   - NEVER hardcode hex colours on text elements\n   - Scope ALL CSS to .{function}-section — no global element rules\n   - Include @media (max-width: 768px) responsive rules\n   - Mobile-first: touch targets >= 44px\n\n6. DARK SECTIONS (if the section has a dark background):\n   Set on the root container:\n     --section-text: rgba(255,255,255,0.9);\n     --section-text-muted: rgba(255,255,255,0.7);\n     --section-heading: #ffffff;\n     --section-surface: rgba(255,255,255,0.05);\n     --section-border: rgba(255,255,255,0.2);\n\n7. CSS VARIABLES AVAILABLE:\n   --color-primary, --color-primary-hover, --color-primary-text\n   --color-secondary, --color-accent\n   --color-text, --color-text-muted, --color-heading\n   --color-background, --color-surface, --color-card-bg, --color-border\n   --color-header-bg, --color-header-text\n   --color-footer-bg, --color-footer-text, --color-white\n   --container-max-width (1200px), --spacing-section (5rem 2rem)\n   --border-radius, --shadow\n\n8. INTERACTIVE ELEMENTS (if section has JS):\n   - Client-side only, no external API calls\n   - Wrap in IIFE: (function() { ... })();\n   - No global variable pollution\n   - Progressive enhancement — works without JS where possible\n   - No external CDN imports\n\n9. QUALITY:\n   - No placeholder text (Lorem ipsum, TODO, [INSERT])\n   - No unrendered template variables in output\n   - Semantic HTML (section, article, nav — not div soup)\n   - Accessible: labels on inputs, ARIA where needed, focus states\n   - No fabricated content\n\n== PRE-OUTPUT SELF-CHECK ==\n\nBefore returning your JSON, verify these two conditions:\n\n(a) Count the occurrences of {{placeholder \"field_name\"}} tokens in your html_template.\n    Count the keys in your input_schema.fields object.\n    These two counts MUST be equal. If they differ, revise the template to\n    add the missing placeholders, or revise the schema to remove the\n    orphaned fields. Do not return a mismatched pair.\n\n(b) Every schema field (Tier A, Tier B, Tier C, Renderer) must be reachable\n    through its {{placeholder \"field_name\"}} token somewhere in the template. Tier B\n    fallbacks are not a licence to omit the placeholder — the fallback only\n    takes effect when the placeholder is present.\n\n== END CONTRACT ==\n\nRespond with ONLY a JSON object (no markdown fences, no preamble) containing:\n{\n  \"function\": \"the-kebab-case-function-name\",\n  \"html_template\": \"the full <style>...<section>...</section> template with Go-template placeholders as specified in the CONTRACT above\",\n  \"input_schema\": {\n    \"fields\": {\n      \"voice_field_example\": {\n        \"type\": \"text\",\n        \"source\": \"llm\",\n        \"required\": true,\n        \"llm_guidance\": \"Detailed guidance for the content writer\"\n      },\n      \"tunable_label_example\": {\n        \"type\": \"text\",\n        \"source\": \"static\",\n        \"required\": false,\n        \"fallback\": \"Learn more\",\n        \"llm_guidance\": \"Optional: override default if site tone differs\"\n      },\n      \"site_data_example\": {\n        \"type\": \"text\",\n        \"source\": \"site_specs.identity.company_name\",\n        \"required\": true\n      },\n      \"renderer_field_example\": {\n        \"type\": \"text\",\n        \"source\": \"renderer\",\n        \"required\": false,\n        \"fallback\": \"00:00\",\n        \"llm_guidance\": \"Initial display before JS takes over\"\n      }\n    }\n  },\n  \"is_dark_section\": true\n}\n", "timeout_seconds": 120} | t         | 2026-03-31 13:21:03.832413+00 | 2026-05-05 14:32:23.017088+00 |            | ["component-generation", "html-template", "css"] | docker.io/aqls/agent-chassis | v1.0.990  | {./agent-chassis,-config,configs/agent-chassis.yaml} | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} | specialist     | experimental | ["component-library", "design", "build"] | {}                     |           0 | f           | {"optional": ["site_type", "page_context", "description", "design_direction", "reference_content"], "required": ["section_type"]} | {"fields": {"status": "created or already_exists", "function": "The component function name", "component_id": "UUID of the created/existing component", "section_type": "The section type it implements"}} |                    0
(1 row)


--- shorter templates
-- ============================================================================
-- Migration 037: Tighten component-creator prompt — structural completeness
-- ============================================================================
-- Adds a third self-check condition (c) to the pre-output verification
-- block of the component-creator prompt. The new condition asks the
-- LLM to verify structural element closure before returning, and to
-- prefer truncation of elaboration over leaving structure incomplete.
--
-- This is one of three layers addressing the truncated-template issue
-- surfaced during the gamesdesign Phase 1 build:
--
--   (1) Prompt: ask for structural completeness explicitly (this migration).
--   (2) Action-side: store_generated_component pre-store validation
--       extended to reject substantive-but-no-placeholders templates
--       and templates containing "<no value>" artifacts (separate code
--       change in store_generated_component_action.go).
--   (3) Read-time: loadComponentSchemas in plan_sections already skips
--       templates lacking </section> (existing safety net).
--
-- Layer (2) is the durable defence — it stops bad templates entering
-- the registry. Layer (1) reduces how often (2) needs to fire. Layer
-- (3) was the only line of defence before this migration.
--
-- Idempotent: uses replace() against the prompt text. Re-running finds
-- no matching old string and is a no-op.
-- ============================================================================

BEGIN;

CREATE TABLE IF NOT EXISTS agent_def_component_creator_backup_20260506 AS
SELECT * FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
  AND type = 'component-creator';

DO $$
DECLARE
snap_count INT;
BEGIN
SELECT COUNT(*) INTO snap_count
FROM agent_def_component_creator_backup_20260506;
IF snap_count = 0 THEN
        RAISE EXCEPTION 'Snapshot empty — no row found for component-creator. Aborting.';
END IF;
END$$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{prompt_template}',
        to_jsonb(
                replace(
                        default_config #>> '{prompt_template}',
                        E'(b) Every schema field (Tier A, Tier B, Tier C, Renderer) must be reachable\n    through its {{placeholder "field_name"}} token somewhere in the template. Tier B\n    fallbacks are not a licence to omit the placeholder — the fallback only\n    takes effect when the placeholder is present.',
                        E'(b) Every schema field (Tier A, Tier B, Tier C, Renderer) must be reachable\n    through its {{placeholder "field_name"}} token somewhere in the template. Tier B\n    fallbacks are not a licence to omit the placeholder — the fallback only\n    takes effect when the placeholder is present.\n\n(c) Structural completeness. Your html_template MUST close every element\n    it opens. Verify before returning:\n      - The outer <section ...> has a matching </section> at the end.\n      - The <style> block (if present) has a matching </style>.\n      - The <script> block (if present) has a matching </script>.\n      - All inner <div>, <a>, <span>, <button> tags are closed.\n    If you are approaching the response size limit, prefer to TRIM\n    elaboration (CSS comments, decorative styles, repeated card variants)\n    rather than leave structure incomplete. A complete shorter template\n    is correct; an incomplete longer template is broken and will be\n    rejected. Never emit "<no value>" anywhere — that is the rendered\n    output of a broken template, not source.'
                )
        )
                     ),
    updated_at = NOW()
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';

-- Verification — confirm the new check (c) is present and the old block is intact.
SELECT
    'has_structural_self_check' AS check,
    (default_config #>> '{prompt_template}') LIKE '%(c) Structural completeness%' AS pass
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
UNION ALL
SELECT
    'has_no_value_warning',
    (default_config #>> '{prompt_template}') LIKE '%Never emit "<no value>"%'
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
UNION ALL
SELECT
    'still_has_check_a',
    (default_config #>> '{prompt_template}') LIKE '%(a) Count the occurrences%'
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';

COMMIT;

--

-- Migration 038: fix component-creator workflow config paths
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                jsonb_set(
                        jsonb_set(
                                jsonb_set(
                                        default_config,
                                        '{workflow,steps,store_component,config,section_type}',
                                        '"input_data.spec.section_type"'::jsonb
                                ),
                                '{workflow,steps,store_component,config,description}',
                                '"input_data.spec.description"'::jsonb
                        ),
                        '{workflow,steps,store_component,config,site_type}',
                        '"input_data.spec.site_type"'::jsonb
                ),
                '{workflow,steps,store_component,config,page_context}',
                '"input_data.spec.page_context"'::jsonb
        ),
        '{workflow,steps,store_component,config,design_direction}',
        '"input_data.spec.design_direction"'::jsonb
                     )
WHERE type = 'component-creator';

---
-- change prompt
    -- backup:
-- UPDATE agent_definitions
-- SET default_config = (
--     SELECT default_config
--     FROM agent_def_component_creator_backup_20260507_pre_directory_build
--              LIMIT 1
--     ), updated_at = NOW()
-- WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';

-- ============================================================================
-- Migration 039: Component-creator prompt — Tier D (derived lists / query.*)
-- ============================================================================
-- Adds knowledge of the `query.*` source type to the component-creator
-- prompt, plus the array + items schema shape used by list/directory/grid
-- components. Without this, the LLM falls back to the per-item-field
-- pattern (post1_title, post2_title, ..., post6_title) which forces
-- fabrication: there is no real list of posts, just six independent
-- string fields the LLM has to invent.
--
-- This migration teaches a fourth field-classification tier:
--
--   TIER D — DERIVED LIST (source: "query.{name}")
--     Used for "show items from somewhere" — a list of tool pages, blog
--     posts, entities, case studies, etc. The schema declares ONE field
--     of type=array with an items sub-schema; the resolver runs a real
--     DB query at plan time and produces the items list. The LLM does
--     NOT fabricate items.
--
-- The migration also extends pre-output self-check (a) to handle array
-- fields, since the placeholder-count-equals-schema-field-count rule
-- breaks for arrays (one schema field maps to N placeholders inside a
-- {{range}}).
--
-- The known query vocabulary starts small (one query: pages_where_type)
-- and is extensible. The prompt enumerates what's available so the LLM
-- doesn't invent new query names. Adding queries later means adding to
-- both the queryresolve Go package and this prompt.
--
-- Idempotent: uses replace() against the prompt text. Re-running finds
-- no matching old strings and is a no-op.
-- ============================================================================

BEGIN;

-- Snapshot is already in place from migration 038. Verify it exists before
-- proceeding so we never apply this prompt change without a rollback target.
DO $$
DECLARE
backup_count INT;
BEGIN
SELECT COUNT(*) INTO backup_count
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name = 'agent_def_component_creator_backup_20260507_pre_directory_build';
IF backup_count = 0 THEN
        RAISE EXCEPTION
            'Backup table from migration 038 not found. '
            'Run migration 038 (backups) first. '
            'Aborting before any prompt change.';
END IF;
END$$;

-- ----------------------------------------------------------------------------
-- (1) Insert a new TIER D block in the field-classification section.
-- ----------------------------------------------------------------------------
-- Anchored on the trailing line of the existing RENDERER block, so the new
-- text lands BEFORE the "RULE:" line that closes the section.
-- ----------------------------------------------------------------------------

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{prompt_template}',
        to_jsonb(
                replace(
                        default_config #>> '{prompt_template}',
                        E'    Also: RENDERER fields (source: "renderer", required: false)\n      Values filled at render time by JS or the renderer, not by the content writer.\n      Provide a safe initial fallback (e.g. "00:00" for a timer).',
                        E'    Also: RENDERER fields (source: "renderer", required: false)\n      Values filled at render time by JS or the renderer, not by the content writer.\n      Provide a safe initial fallback (e.g. "00:00" for a timer).\n\n    TIER D — DERIVED LISTS (source: "query.{name}")\n      Used when the section displays a LIST of real items drawn from the\n      site''s own data: tool pages, blog posts, entities, case studies,\n      news articles, etc. The list is RESOLVED AT PLAN TIME by a database\n      query. The LLM does NOT fabricate items, does NOT invent names\n      or URLs, does NOT pad to a fixed count.\n\n      Schema shape — declare ONE field of type "array":\n        "items": {\n          "type": "array",\n          "source": "query.pages_where_type:tool",\n          "min_items": 1,\n          "limit": 6,\n          "required": true,\n          "items": {\n            "title":            { "type": "text" },\n            "url":              { "type": "url"  },\n            "meta_description": { "type": "text" },\n            "nav_label":        { "type": "text" }\n          }\n        }\n\n      Template shape — iterate with {{range}}, reference each row''s\n      fields with {{.field_name}} (NOT {{placeholder "..."}}):\n        <div class="grid">\n          {{range .items}}\n          <article class="card">\n            <h3>{{.title}}</h3>\n            <p>{{.meta_description}}</p>\n            <a href="{{.url}}">Open</a>\n          </article>\n          {{end}}\n        </div>\n\n      Known query names (use these EXACTLY — do not invent new ones):\n        query.pages_where_type:tool        — all tool pages on this site\n        query.pages_where_type:blog_post   — all blog posts on this site\n        query.pages_where_type:entity_page — all entity pages\n        query.pages_where_type:guide       — all guide pages\n        query.pages_where_type:game        — all game pages\n\n      Item shape returned by the resolver (you can reference any of these\n      in the template via {{.field_name}}):\n        title             — page title\n        url               — page URL (already canonical)\n        meta_description  — short description\n        nav_label         — navigation label\n        name              — internal page name (slug-form)\n\n      When to use Tier D: any time the section''s purpose is "show a list\n      of these items." If you find yourself writing post1_title, post2_title,\n      post3_title, ..., STOP — that is the wrong shape. Use Tier D instead.\n\n      LLM-source fields STILL apply to the section''s heading/eyebrow/intro/CTA\n      around the list. Tier D is for the items themselves; Tier A handles\n      the surrounding voice content.'
                )
        )
                     ),
    updated_at = NOW()
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';

-- ----------------------------------------------------------------------------
-- (2) Extend pre-output self-check (a) to cover array fields.
-- ----------------------------------------------------------------------------
-- The existing (a) rule says "placeholder count = schema field count".
-- That rule breaks for array fields: one schema field maps to N {{.X}}
-- references inside a {{range}}. We add a clarifying paragraph rather
-- than rewriting (a) so the rule still applies for non-array fields.
-- ----------------------------------------------------------------------------

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{prompt_template}',
        to_jsonb(
                replace(
                        default_config #>> '{prompt_template}',
                        E'(a) Count the occurrences of {{placeholder "field_name"}} tokens in your html_template.\n    Count the keys in your input_schema.fields object.\n    These two counts MUST be equal. If they differ, revise the template to\n    add the missing placeholders, or revise the schema to remove the\n    orphaned fields. Do not return a mismatched pair.',
                        E'(a) Count the occurrences of {{placeholder "field_name"}} tokens in your html_template.\n    Count the keys in your input_schema.fields object.\n    These two counts MUST be equal. If they differ, revise the template to\n    add the missing placeholders, or revise the schema to remove the\n    orphaned fields. Do not return a mismatched pair.\n\n    Exception for Tier D array fields: an array field with type "array"\n    counts as ONE schema field even though the template references its\n    items multiple times via {{range .items}} ... {{.field}} ... {{end}}.\n    The {{.title}}, {{.url}}, etc. references INSIDE a {{range}} block\n    are NOT {{placeholder "..."}} tokens and do NOT count toward the\n    placeholder total. They are template field accesses against the\n    array''s items, declared in the array''s items sub-schema, not in\n    input_schema.fields.'
                )
        )
                     ),
    updated_at = NOW()
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';

-- ----------------------------------------------------------------------------
-- Verification — confirm both edits landed and the existing structure is intact.
-- ----------------------------------------------------------------------------

SELECT
    'has_tier_d' AS check,
    (default_config #>> '{prompt_template}') LIKE '%TIER D — DERIVED LISTS%' AS pass
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
UNION ALL
SELECT
    'has_query_vocabulary',
    (default_config #>> '{prompt_template}') LIKE '%query.pages_where_type:tool%'
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
UNION ALL
SELECT
    'has_array_exception_in_check_a',
    (default_config #>> '{prompt_template}') LIKE '%Exception for Tier D array fields%'
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
UNION ALL
SELECT
    'still_has_tier_a',
    (default_config #>> '{prompt_template}') LIKE '%TIER A — VOICE CONTENT%'
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
UNION ALL
SELECT
    'still_has_tier_b',
    (default_config #>> '{prompt_template}') LIKE '%TIER B — TUNABLE LABELS%'
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
UNION ALL
SELECT
    'still_has_tier_c',
    (default_config #>> '{prompt_template}') LIKE '%TIER C — SITE DATA%'
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
UNION ALL
SELECT
    'still_has_check_b',
    (default_config #>> '{prompt_template}') LIKE '%(b) Every schema field%'
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4'
UNION ALL
SELECT
    'still_has_check_c',
    (default_config #>> '{prompt_template}') LIKE '%(c) Structural completeness%'
FROM agent_definitions
WHERE id = '23720180-7a39-4e3d-92e1-ebdbf95b57f4';

COMMIT;
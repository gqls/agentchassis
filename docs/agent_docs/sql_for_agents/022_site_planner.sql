-- ============================================================================
-- Add Input Contracts to Existing Agents
-- ============================================================================
-- These contracts define what each agent expects to receive.
-- Contract validation will fail fast with clear error messages when required
-- fields are missing.
-- ============================================================================

-- site-planner
-- Plans site structure, pages, and components
UPDATE agent_definitions
SET input_contract = '{
    "required": ["site_record"],
    "optional": ["input_data", "reviewed_brief"]
}'::jsonb,
    output_contract = '{
    "produces": ["validated_plan", "pages", "style_collection", "needs_logo", "needs_images"]
}'::jsonb
WHERE type = 'site-planner';


-- switch to haiku

-- =============================================================
-- 1. Switch site-planner model from sonnet to haiku
--
-- The site-planner does structured planning (component selection,
-- page layout) which haiku handles well. This reduces cost and
-- is less likely to hit rate limits during overload.
-- =============================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,ai_service,model}',
        '"claude-haiku-4-5"'
                     ),
    updated_at = now()
WHERE type = 'site-planner';

-- Verify
SELECT
    type,
    default_config->'workflow'->'steps'->'plan_site'->'config'->'ai_service'->>'model' as planner_model
FROM agent_definitions
WHERE type = 'site-planner';

---

-- ============================================================================
-- Fix 2: site-planner prompt edits (Defect 1 + Defect 2)
-- ============================================================================
-- Applies:
--   2a. add faq / pricing to the standard mappings
--   2b. component list shows [function] first; instruct to use only that
--       (kills the "FAQ Section" display-name leak at source)
--   2c. add no-pairing rule (rule 7); fix call_to_action -> call-to-action
--
-- Writes the full corrected prompt_template back via jsonb_set, dollar-quoted.
-- Takes effect on the next site-planner run (no cache, no chassis rebuild).
-- ============================================================================

-- 0. Backup current prompt
INSERT INTO agent_definition_prompt_backups (type, step, prompt)
SELECT 'site-planner', 'plan_site',
       default_config #>> '{workflow,steps,plan_site,config,prompt_template}'
FROM agent_definitions
WHERE type = 'site-planner';
-- (relies on the backup table created by fix1; if running fix2 standalone,
--  create it first — see fix1_content_gap_planner_prompt.sql section 0)

-- 1. Apply the corrected prompt
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb($PROMPT$Plan a website for {{.input_data.domain}}.

## Site Brief
{{.reviewed_brief}}

## Available Section Components
The following components are available. In the "sections" arrays you MUST use ONLY the function name — shown in [brackets] below — never the display name or title.

{{range .available_components}}
- [{{.function}}] {{.display_name}}: {{.description}}
{{end}}

## Available Style Collections
{{.available_styles}}

## Task
Create a comprehensive site plan using ONLY the components listed above.

Return JSON in this format:
```json
{
  "pages": [
    {
      "name": "index",
      "title": "Page Title | Site Name",
      "nav_label": "Home",
      "nav_order": 1,
      "in_header": true,
      "in_footer": true,
      "sections": ["hero", "features", "testimonials", "call-to-action"]
    }
  ],
  "style_collection": "style-name",
  "needs_logo": true,
  "needs_images": true,
  "image_prompts": {
    "logo": "Description for logo generation",
    "hero_home": "Description for home hero image"
  }
}
```

STRICT RULES:
1. ONLY use component function names from the "Available Section Components" list above (the value in [brackets])
2. DO NOT invent new component names - if unsure, use "hero" for hero sections, "features" for feature lists, "call-to-action" for CTAs
3. Use these standard mappings:
   - For any hero/banner at page top: use "hero" or page-specific variants like "contact-hero", "services-hero", "about-hero"
   - For feature lists: use "features"
   - For service listings: use "services-grid"
   - For testimonials/quotes: use "testimonials" or "social_proof"
   - For calls to action: use "call-to-action"
   - For contact forms: use "contact-form"
   - For contact details: use "contact-info"
   - For team sections: use "leadership-team"
   - For case studies: use "case-studies-list"
   - For about content: use "about-content"
   - For differentiators/why-us: use "differentiators-section"
   - For FAQ / question-and-answer content: use "faq"
   - For pricing tables/tiers: use "pricing"

4. Choose style_collection based on industry and tone from the brief
5. Keep header navigation to 5-8 items maximum
6. Always include: index (home) and contact pages
7. Structured components (faq, pricing, data-driven features) are complete content surfaces. Do not pair them with a generic-text-block covering the same content on the same page.

IMAGE GENERATION (REQUIRED - DO NOT SKIP):
You MUST include needs_logo, needs_images, and image_prompts in your response.

- Set needs_logo: true (always)
- Set needs_images: true (always)
- Provide image_prompts object with BOTH of these keys:
  - "logo": A detailed 2-3 sentence prompt for logo generation. Describe the style (modern, classic, minimal), colors that match the brand, and any relevant imagery for the industry.
  - "hero_home": A detailed 2-3 sentence prompt for the homepage hero background. Describe the mood (professional, energetic, calm), imagery type (abstract, photographic, geometric), and colors/atmosphere that fit the brand.

Example image_prompts:
{
  "logo": "A modern, minimal logo for a tech consulting company. Use clean geometric shapes with a navy blue and teal color palette. The design should convey innovation and trustworthiness.",
  "hero_home": "A professional, abstract background with flowing gradients in deep navy and teal. Include subtle geometric patterns that suggest technology and connectivity. The mood should be confident and forward-thinking."
}$PROMPT$::text),
        false
                     ),
    updated_at = NOW()
WHERE type = 'site-planner'
    RETURNING type, version, updated_at,
          left(default_config #>> '{workflow,steps,plan_site,config,prompt_template}', 80) AS prompt_head;

